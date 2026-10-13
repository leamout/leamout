package entitlements

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/leamout/leamout/server/internal/commercial/licensing"
	"github.com/leamout/leamout/server/pkg/apperror"
)

type Service struct {
	license *licensing.Service
	repo    *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) SetLicenseService(service *licensing.Service) {
	s.license = service
}

func (s *Service) ListEffective(
	ctx context.Context,
	organizationID uuid.UUID,
) ([]EffectiveEntitlement, error) {
	if organizationID == uuid.Nil {
		return nil, apperror.NewBadRequest("organization_id is required")
	}
	result := make([]EffectiveEntitlement, 0, len(capabilities))
	for _, capability := range capabilities {
		enabled, err := s.Enabled(ctx, organizationID, capability)
		if err != nil {
			return nil, err
		}
		result = append(result, EffectiveEntitlement{
			Capability: capability,
			Enabled:    enabled,
		})
	}
	return result, nil
}

func (s *Service) Enabled(
	ctx context.Context,
	organizationID uuid.UUID,
	capability Capability,
) (bool, error) {
	if organizationID == uuid.Nil {
		return false, apperror.NewBadRequest("organization_id is required")
	}
	capability, err := normalizeCapability(capability)
	if err != nil {
		return false, err
	}
	if s.license.SelfHosted() {
		return s.license.Enabled(ctx, string(capability))
	}
	enabled, err := s.repo.EffectiveEnabled(ctx, organizationID, capability)
	if err != nil {
		return false, apperror.NewInternal("get effective organization entitlement", err)
	}
	return enabled, nil
}

func (s *Service) Require(
	ctx context.Context,
	organizationID uuid.UUID,
	capability Capability,
) error {
	enabled, err := s.Enabled(ctx, organizationID, capability)
	if err != nil {
		return err
	}
	if !enabled {
		return apperror.NewForbidden(
			fmt.Sprintf("organization capability %q is not enabled", capability),
		)
	}
	return nil
}

func (s *Service) Set(
	ctx context.Context,
	organizationID uuid.UUID,
	capability Capability,
	enabled bool,
) (Entitlement, error) {
	if organizationID == uuid.Nil {
		return Entitlement{}, apperror.NewBadRequest("organization_id is required")
	}
	capability, err := normalizeCapability(capability)
	if err != nil {
		return Entitlement{}, err
	}
	if s.license.SelfHosted() {
		return Entitlement{}, apperror.NewForbidden("self-hosted features are managed by the signed license")
	}
	value, err := s.repo.Set(ctx, organizationID, capability, enabled)
	if err != nil {
		return Entitlement{}, apperror.NewInternal("set organization entitlement", err)
	}
	return value, nil
}
