package licensing

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/commercial/entitlements"
)

type fakeAuthorityStore struct {
	license      License
	subscription SubscriptionAuthority
}

func (f fakeAuthorityStore) GetLicense(context.Context, uuid.UUID, uuid.UUID) (License, error) {
	return f.license, nil
}

func (f fakeAuthorityStore) GetSubscription(context.Context, uuid.UUID, uuid.UUID) (SubscriptionAuthority, error) {
	return f.subscription, nil
}

type fakeEntitlementResolver struct {
	values  []entitlements.Entitlement
	request entitlements.ResolveRequest
}

type fixedProvider struct{ signer DocumentSigner }

func (f fixedProvider) Signer(context.Context, string) (DocumentSigner, error) { return f.signer, nil }

type fixedSigner struct{ document Document }

func (f fixedSigner) Sign(Claims) (Document, error) { return f.document, nil }

func (f *fakeEntitlementResolver) Resolve(_ context.Context, request entitlements.ResolveRequest) ([]entitlements.Entitlement, error) {
	f.request = request
	return f.values, nil
}

func TestIssuerBuildsVerifiableDocumentFromActiveAuthority(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	provider, err := NewStaticKeyProvider(map[string]ed25519.PrivateKey{"current": privateKey})
	if err != nil {
		t.Fatal(err)
	}
	organizationID, licenseID, subscriptionID, planID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	now := time.Date(2026, time.August, 30, 12, 0, 0, 0, time.UTC)
	expiresAt := now.Add(24 * time.Hour)
	enabled := true
	limit := int64(10)
	resolver := &fakeEntitlementResolver{values: []entitlements.Entitlement{
		{Key: "voice.recording", Kind: entitlements.KindFeature, Enabled: &enabled},
		{Key: "voice.concurrent_calls", Kind: entitlements.KindLimit, Limit: &limit},
	}}
	issuer, err := NewIssuer(fakeAuthorityStore{
		license: License{
			ID: licenseID.String(), OrganizationID: organizationID.String(), SubscriptionID: subscriptionID.String(),
			Status: StatusActive, MaxDeployments: 2, SigningKeyID: "current", IssuedAt: now, ExpiresAt: &expiresAt,
		},
		subscription: SubscriptionAuthority{ID: subscriptionID, PlanID: planID, Status: "active", StartsAt: now},
	}, resolver, provider)
	if err != nil {
		t.Fatal(err)
	}

	document, err := issuer.Issue(context.Background(), IssueRequest{
		OrganizationID: organizationID, LicenseID: licenseID, At: now.Add(time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}
	keyring, _ := NewKeyring(map[string]ed25519.PublicKey{"current": publicKey})
	claims, err := keyring.Verify(document, now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if claims.LicenseID != licenseID.String() || claims.MaxDeployments != 2 || len(claims.Entitlements) != 2 {
		t.Fatalf("unexpected claims: %+v", claims)
	}
	if claims.Entitlements[0].Key != "voice.concurrent_calls" {
		t.Fatalf("claims were not sorted deterministically: %+v", claims.Entitlements)
	}
	if resolver.request.PlanID != planID || resolver.request.LicenseID != licenseID {
		t.Fatalf("unexpected resolution request: %+v", resolver.request)
	}
}

func TestIssuerRejectsInactiveAuthority(t *testing.T) {
	organizationID, licenseID := uuid.New(), uuid.New()
	now := time.Now().UTC()
	_, privateKey, _ := ed25519.GenerateKey(rand.Reader)
	provider, _ := NewStaticKeyProvider(map[string]ed25519.PrivateKey{"current": privateKey})

	tests := []struct {
		name    string
		license License
		want    error
	}{
		{"pending", validLicense(organizationID, licenseID, now, StatusPending), ErrLicenseNotActive},
		{"missing key", func() License {
			value := validLicense(organizationID, licenseID, now, StatusActive)
			value.SigningKeyID = ""
			return value
		}(), ErrSigningKeyRequired},
		{"expired", func() License {
			value := validLicense(organizationID, licenseID, now.Add(-time.Hour), StatusActive)
			expiry := now
			value.ExpiresAt = &expiry
			return value
		}(), ErrLicenseExpired},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			issuer, _ := NewIssuer(fakeAuthorityStore{license: test.license}, &fakeEntitlementResolver{}, provider)
			_, err := issuer.Issue(context.Background(), IssueRequest{OrganizationID: organizationID, LicenseID: licenseID, At: now})
			if !errors.Is(err, test.want) {
				t.Fatalf("expected %v, got %v", test.want, err)
			}
		})
	}
}

func TestIssuerRejectsInactiveSubscription(t *testing.T) {
	organizationID, licenseID, subscriptionID := uuid.New(), uuid.New(), uuid.New()
	now := time.Now().UTC()
	license := validLicense(organizationID, licenseID, now, StatusActive)
	license.SubscriptionID = subscriptionID.String()
	_, privateKey, _ := ed25519.GenerateKey(rand.Reader)
	provider, _ := NewStaticKeyProvider(map[string]ed25519.PrivateKey{"current": privateKey})
	issuer, _ := NewIssuer(fakeAuthorityStore{
		license: license, subscription: SubscriptionAuthority{ID: subscriptionID, PlanID: uuid.New(), Status: "cancelled", StartsAt: now},
	}, &fakeEntitlementResolver{}, provider)
	_, err := issuer.Issue(context.Background(), IssueRequest{OrganizationID: organizationID, LicenseID: licenseID, At: now})
	if !errors.Is(err, ErrSubscriptionInactive) {
		t.Fatalf("expected inactive subscription, got %v", err)
	}
}

func TestIssuerRejectsProviderReturningWrongKey(t *testing.T) {
	organizationID, licenseID := uuid.New(), uuid.New()
	now := time.Now().UTC()
	issuer, _ := NewIssuer(
		fakeAuthorityStore{license: validLicense(organizationID, licenseID, now, StatusActive)},
		&fakeEntitlementResolver{},
		fixedProvider{signer: fixedSigner{document: Document{Algorithm: "Ed25519", KeyID: "other"}}},
	)
	_, err := issuer.Issue(context.Background(), IssueRequest{OrganizationID: organizationID, LicenseID: licenseID, At: now})
	if !errors.Is(err, ErrSignerKeyMismatch) {
		t.Fatalf("expected signer key mismatch, got %v", err)
	}
}

func validLicense(organizationID, licenseID uuid.UUID, issuedAt time.Time, status Status) License {
	return License{
		ID: licenseID.String(), OrganizationID: organizationID.String(), Status: status,
		MaxDeployments: 1, SigningKeyID: "current", IssuedAt: issuedAt,
	}
}
