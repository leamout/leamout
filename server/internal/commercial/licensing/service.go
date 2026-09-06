package licensing

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	commercialstate "github.com/leamout/leamout/internal/commercial/state"
)

const (
	deploymentTokenPrefix = "lm_dep_"
	deploymentSecretBytes = 32
)

// Service owns self-hosted license lifecycle and deployment activation policy.
type Service struct {
	repo  *Repository
	state *commercialstate.Service
	now   func() time.Time
}

func NewService(repo *Repository, state *commercialstate.Service) *Service {
	return &Service{repo: repo, state: state, now: time.Now}
}

// Create creates a pending license from current commercial state. It does not
// claim to have produced a signed license artifact; activation is a separate step.
func (s *Service) Create(ctx context.Context, organizationID uuid.UUID, input CreateInput) (License, error) {
	if err := validateID(organizationID, ErrOrganizationIDRequired); err != nil {
		return License{}, err
	}
	normalized, issuedAt, err := normalizeCreate(input, s.now())
	if err != nil {
		return License{}, err
	}
	resolved, err := s.state.ResolveAt(ctx, organizationID, issuedAt)
	if err != nil {
		return License{}, err
	}
	if resolved.Standing != commercialstate.StandingActive || resolved.SubscriptionID == nil {
		return License{}, ErrCommercialStateUnavailable
	}
	limit, ok := resolved.Limit(MaxDeploymentsEntitlement)
	if !ok || limit <= 0 || limit > int64(^uint32(0)>>1) {
		return License{}, ErrInvalidDeploymentLimit
	}
	snapshot := entitlementSnapshot{Features: resolved.Features, Limits: resolved.Limits}
	return s.repo.Create(
		ctx,
		organizationID,
		*resolved.SubscriptionID,
		int32(limit),
		normalized.SigningKeyID,
		issuedAt,
		normalized.ExpiresAt,
		snapshot,
	)
}

func (s *Service) Get(ctx context.Context, organizationID, id uuid.UUID) (License, error) {
	if err := validateID(organizationID, ErrOrganizationIDRequired); err != nil {
		return License{}, err
	}
	if err := validateID(id, ErrLicenseIDRequired); err != nil {
		return License{}, err
	}
	return s.repo.Get(ctx, organizationID, id)
}

func (s *Service) List(ctx context.Context, organizationID uuid.UUID) ([]License, error) {
	if err := validateID(organizationID, ErrOrganizationIDRequired); err != nil {
		return nil, err
	}
	return s.repo.List(ctx, organizationID)
}

func (s *Service) Transition(ctx context.Context, organizationID, id uuid.UUID, to Status) (License, error) {
	target, err := normalizeStatus(to)
	if err != nil {
		return License{}, err
	}
	current, err := s.Get(ctx, organizationID, id)
	if err != nil {
		return License{}, err
	}
	if target == StatusActive {
		if current.SigningKeyID == nil {
			return License{}, ErrSigningKeyRequired
		}
		if current.ExpiresAt != nil && !current.ExpiresAt.After(s.now()) {
			return License{}, ErrLicenseUnavailable
		}
	}
	return s.repo.Transition(ctx, organizationID, id, target)
}

func (s *Service) UpdateExpiration(ctx context.Context, organizationID, id uuid.UUID, expiresAt *time.Time) (License, error) {
	current, err := s.Get(ctx, organizationID, id)
	if err != nil {
		return License{}, err
	}
	if current.Status == StatusExpired || current.Status == StatusRevoked {
		return License{}, ErrInvalidTransition
	}
	if expiresAt != nil {
		value := expiresAt.UTC()
		if !value.After(current.IssuedAt) {
			return License{}, ErrInvalidExpiration
		}
		expiresAt = &value
	}
	return s.repo.UpdateExpiration(ctx, organizationID, id, expiresAt)
}

func (s *Service) ActivateDeployment(ctx context.Context, organizationID, licenseID uuid.UUID, input ActivateDeploymentInput) (Deployment, error) {
	if err := validateID(organizationID, ErrOrganizationIDRequired); err != nil {
		return Deployment{}, err
	}
	if err := validateID(licenseID, ErrLicenseIDRequired); err != nil {
		return Deployment{}, err
	}
	normalized, err := normalizeDeployment(input)
	if err != nil {
		return Deployment{}, err
	}
	return s.repo.ActivateDeployment(ctx, organizationID, licenseID, normalized, s.now())
}

func (s *Service) ListDeployments(ctx context.Context, organizationID, licenseID uuid.UUID) ([]Deployment, error) {
	if err := validateID(organizationID, ErrOrganizationIDRequired); err != nil {
		return nil, err
	}
	if err := validateID(licenseID, ErrLicenseIDRequired); err != nil {
		return nil, err
	}
	return s.repo.ListDeployments(ctx, organizationID, licenseID)
}

func (s *Service) TouchDeployment(ctx context.Context, organizationID, licenseID uuid.UUID, deploymentID string) (Deployment, error) {
	normalized, err := normalizeDeployment(ActivateDeploymentInput{DeploymentID: deploymentID})
	if err != nil {
		return Deployment{}, err
	}
	return s.repo.TouchDeployment(ctx, organizationID, licenseID, normalized.DeploymentID, s.now())
}

func (s *Service) DeactivateDeployment(ctx context.Context, organizationID, licenseID uuid.UUID, deploymentID string) (Deployment, error) {
	normalized, err := normalizeDeployment(ActivateDeploymentInput{DeploymentID: deploymentID})
	if err != nil {
		return Deployment{}, err
	}
	return s.repo.DeactivateDeployment(ctx, organizationID, licenseID, normalized.DeploymentID)
}

// EnrollManagedCarrier rotates the managed-carrier machine credential for an
// already-active self-hosted deployment. The usable token is returned only from
// this call; persistence contains only its cryptographic hash.
func (s *Service) EnrollManagedCarrier(
	ctx context.Context,
	organizationID, licenseID uuid.UUID,
	deploymentID string,
) (ManagedCarrierEnrollment, error) {
	normalized, err := normalizeDeployment(ActivateDeploymentInput{DeploymentID: deploymentID})
	if err != nil {
		return ManagedCarrierEnrollment{}, err
	}

	deployment, err := s.repo.GetDeployment(ctx, organizationID, licenseID, normalized.DeploymentID)
	if err != nil {
		return ManagedCarrierEnrollment{}, err
	}
	if deployment.Status != DeploymentStatusActive {
		return ManagedCarrierEnrollment{}, ErrDeploymentInactive
	}

	now := s.now().UTC()
	license, err := s.Get(ctx, organizationID, licenseID)
	if err != nil {
		return ManagedCarrierEnrollment{}, err
	}
	if license.Status != StatusActive || (license.ExpiresAt != nil && !license.ExpiresAt.After(now)) {
		return ManagedCarrierEnrollment{}, ErrLicenseUnavailable
	}

	commercial, err := s.state.ResolveAt(ctx, organizationID, now)
	if err != nil {
		return ManagedCarrierEnrollment{}, fmt.Errorf("resolve managed carrier commercial state: %w", err)
	}
	if commercial.Standing != commercialstate.StandingActive || !commercial.Enabled(ManagedVoiceEntitlement) {
		return ManagedCarrierEnrollment{}, ErrManagedCarrierUnavailable
	}

	token, prefix, hash, err := generateDeploymentToken()
	if err != nil {
		return ManagedCarrierEnrollment{}, err
	}
	scopes := []string{ManagedCarrierTransitScope}
	credential, err := s.repo.RotateDeploymentCredential(
		ctx,
		deployment.ID,
		ManagedCarrierPurpose,
		hash,
		prefix,
		scopes,
		license.ExpiresAt,
	)
	if err != nil {
		return ManagedCarrierEnrollment{}, err
	}

	return ManagedCarrierEnrollment{
		DeploymentID: deployment.DeploymentID,
		Token:        token,
		TokenPrefix:  credential.TokenPrefix,
		Scopes:       credential.Scopes,
		ExpiresAt:    credential.ExpiresAt,
	}, nil
}

// AuthenticateDeploymentCredential resolves a deployment machine credential at
// the Leamout control-plane boundary. It deliberately re-checks cloud-owned
// deployment, license, organization, and commercial state instead of trusting
// state reported by the self-hosted runtime.
func (s *Service) AuthenticateDeploymentCredential(
	ctx context.Context,
	token string,
	requiredScope string,
) (DeploymentIdentity, error) {
	if _, err := parseDeploymentTokenPrefix(token); err != nil {
		return DeploymentIdentity{}, ErrInvalidDeploymentCredential
	}
	identity, err := s.repo.GetDeploymentCredentialByTokenHash(ctx, hashDeploymentToken(token))
	if err != nil {
		return DeploymentIdentity{}, ErrInvalidDeploymentCredential
	}

	now := s.now().UTC()
	if identity.Purpose != ManagedCarrierPurpose ||
		identity.DeploymentStatus != string(DeploymentStatusActive) ||
		identity.LicenseStatus != string(StatusActive) {
		return DeploymentIdentity{}, ErrInvalidDeploymentCredential
	}
	if identity.LicenseExpiresAt != nil && !identity.LicenseExpiresAt.After(now) {
		return DeploymentIdentity{}, ErrInvalidDeploymentCredential
	}
	if requiredScope != "" && !hasScope(identity.Scopes, requiredScope) {
		return DeploymentIdentity{}, ErrDeploymentCredentialScope
	}

	commercial, err := s.state.ResolveAt(ctx, identity.OrganizationID, now)
	if err != nil {
		return DeploymentIdentity{}, ErrInvalidDeploymentCredential
	}
	if commercial.Standing != commercialstate.StandingActive || !commercial.Enabled(ManagedVoiceEntitlement) {
		return DeploymentIdentity{}, ErrManagedCarrierUnavailable
	}

	if err := s.repo.TouchDeploymentCredential(ctx, identity.CredentialID); err != nil {
		return DeploymentIdentity{}, err
	}
	return DeploymentIdentity{
		CredentialID:   identity.CredentialID,
		OrganizationID: identity.OrganizationID,
		LicenseID:      identity.LicenseID,
		DeploymentID:   identity.DeploymentID,
		Purpose:        identity.Purpose,
		Scopes:         append([]string(nil), identity.Scopes...),
	}, nil
}

func generateDeploymentToken() (token, prefix, hash string, err error) {
	secret := make([]byte, deploymentSecretBytes)
	if _, err := rand.Read(secret); err != nil {
		return "", "", "", fmt.Errorf("generate deployment credential: %w", err)
	}
	encoded := base64.RawURLEncoding.EncodeToString(secret)
	prefix = deploymentTokenPrefix + encoded[:8]
	token = prefix + "_" + encoded
	return token, prefix, hashDeploymentToken(token), nil
}

func hashDeploymentToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func parseDeploymentTokenPrefix(token string) (string, error) {
	if len(token) < len(deploymentTokenPrefix)+9 || !strings.HasPrefix(token, deploymentTokenPrefix) {
		return "", ErrInvalidDeploymentCredential
	}
	separator := len(deploymentTokenPrefix) + 8
	if token[separator] != '_' || len(token) == separator+1 {
		return "", ErrInvalidDeploymentCredential
	}
	return token[:separator], nil
}

func hasScope(scopes []string, required string) bool {
	for _, scope := range scopes {
		if scope == required {
			return true
		}
	}
	return false
}
