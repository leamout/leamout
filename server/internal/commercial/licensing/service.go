package licensing

import (
	"context"
	"time"

	"github.com/google/uuid"
	commercialstate "github.com/leamout/leamout/internal/commercial/state"
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
	return s.repo.Create(ctx, organizationID, *resolved.SubscriptionID, int32(limit), normalized.SigningKeyID, issuedAt, normalized.ExpiresAt)
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
