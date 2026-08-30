package licensing

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/commercial/entitlements"
)

var (
	ErrAuthorityStoreRequired = errors.New("license authority store is required")
	ErrResolverRequired       = errors.New("entitlement resolver is required")
	ErrKeyProviderRequired    = errors.New("signing key provider is required")
	ErrSigningKeyRequired     = errors.New("license signing_key_id is required")
	ErrSubscriptionInactive   = errors.New("license subscription is not active")
	ErrAuthorityMismatch      = errors.New("license authority does not match request")
	ErrSignerKeyMismatch      = errors.New("signer returned a document for a different key")
)

type EntitlementResolver interface {
	Resolve(context.Context, entitlements.ResolveRequest) ([]entitlements.Entitlement, error)
}

type IssueRequest struct {
	OrganizationID uuid.UUID
	LicenseID      uuid.UUID
	At             time.Time
}

// Issuer converts active control-plane authority into a portable signed
// document. It deliberately owns lifecycle checks so handlers cannot sign a
// persisted license merely by obtaining a cryptographic signer.
type Issuer struct {
	store    AuthorityStore
	resolver EntitlementResolver
	keys     SigningKeyProvider
}

func NewIssuer(store AuthorityStore, resolver EntitlementResolver, keys SigningKeyProvider) (*Issuer, error) {
	if store == nil {
		return nil, ErrAuthorityStoreRequired
	}
	if resolver == nil {
		return nil, ErrResolverRequired
	}
	if keys == nil {
		return nil, ErrKeyProviderRequired
	}
	return &Issuer{store: store, resolver: resolver, keys: keys}, nil
}

func (i *Issuer) Issue(ctx context.Context, request IssueRequest) (Document, error) {
	if request.OrganizationID == uuid.Nil {
		return Document{}, ErrOrganizationRequired
	}
	if request.LicenseID == uuid.Nil {
		return Document{}, fmt.Errorf("license_id is required")
	}
	if request.At.IsZero() {
		return Document{}, fmt.Errorf("issuance time is required")
	}

	license, err := i.store.GetLicense(ctx, request.OrganizationID, request.LicenseID)
	if err != nil {
		return Document{}, fmt.Errorf("load license authority: %w", err)
	}
	if license.ID != request.LicenseID.String() || license.OrganizationID != request.OrganizationID.String() {
		return Document{}, ErrAuthorityMismatch
	}
	if err := Validate(license); err != nil {
		return Document{}, fmt.Errorf("validate license authority: %w", err)
	}
	if license.Status != StatusActive || request.At.Before(license.IssuedAt) {
		return Document{}, ErrLicenseNotActive
	}
	if license.ExpiresAt != nil && !request.At.Before(*license.ExpiresAt) {
		return Document{}, ErrLicenseExpired
	}
	if license.SigningKeyID == "" {
		return Document{}, ErrSigningKeyRequired
	}

	var planID uuid.UUID
	if license.SubscriptionID != "" {
		subscriptionID, err := uuid.Parse(license.SubscriptionID)
		if err != nil {
			return Document{}, fmt.Errorf("invalid persisted subscription_id: %w", err)
		}
		subscription, err := i.store.GetSubscription(ctx, request.OrganizationID, subscriptionID)
		if err != nil {
			return Document{}, fmt.Errorf("load subscription authority: %w", err)
		}
		if subscription.ID != subscriptionID || subscription.PlanID == uuid.Nil {
			return Document{}, ErrAuthorityMismatch
		}
		if subscription.Status != "active" || request.At.Before(subscription.StartsAt) ||
			(subscription.EndsAt != nil && !request.At.Before(*subscription.EndsAt)) {
			return Document{}, ErrSubscriptionInactive
		}
		planID = subscription.PlanID
	}

	effective, err := i.resolver.Resolve(ctx, entitlements.ResolveRequest{
		OrganizationID: request.OrganizationID, PlanID: planID, LicenseID: request.LicenseID, At: request.At,
	})
	if err != nil {
		return Document{}, fmt.Errorf("resolve license entitlements: %w", err)
	}
	claims := make([]EntitlementClaim, 0, len(effective))
	for _, entitlement := range effective {
		claims = append(claims, EntitlementClaim{
			Key: entitlement.Key, Kind: entitlement.Kind, Enabled: entitlement.Enabled, Limit: entitlement.Limit,
		})
	}
	sort.Slice(claims, func(left, right int) bool { return claims[left].Key < claims[right].Key })
	signer, err := i.keys.Signer(ctx, license.SigningKeyID)
	if err != nil {
		return Document{}, fmt.Errorf("select license signer: %w", err)
	}
	document, err := signer.Sign(Claims{
		Version: DocumentVersion, LicenseID: license.ID, OrganizationID: license.OrganizationID,
		SubscriptionID: license.SubscriptionID, MaxDeployments: license.MaxDeployments,
		IssuedAt: license.IssuedAt, ExpiresAt: license.ExpiresAt, Entitlements: claims,
	})
	if err != nil {
		return Document{}, fmt.Errorf("sign license document: %w", err)
	}
	if document.KeyID != license.SigningKeyID || document.Algorithm != "Ed25519" {
		return Document{}, ErrSignerKeyMismatch
	}
	return document, nil
}
