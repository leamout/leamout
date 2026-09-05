package numbers

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/leamout/leamout/pkg/apperror"
)

// AvailableSearchRequest is the provider-neutral customer search contract.
// Provider-specific coverage IDs and product identifiers never enter the
// public API.
type AvailableSearchRequest struct {
	CountryCode string
	Contains    string
}

// AvailableNumberResponse is safe to return to customers. SelectionID is an
// opaque, short-lived handle to the provider purchase inputs retained by
// Leamout.
type AvailableNumberResponse struct {
	SelectionID  string `json:"selection_id"`
	Number       string `json:"number"`
	CountryCode  string `json:"country_code"`
	VoiceEnabled bool   `json:"voice_enabled"`
}

// ManagedNumberCandidate is the provider-neutral internal result produced by a
// numbering adapter. Provider identifiers are persisted only behind an opaque
// selection handle and are never serialized by AvailableNumberResponse.
type ManagedNumberCandidate struct {
	Provider              string
	ProviderInventoryID   string
	ProviderProductID     string
	Number                string
	CountryCode           string
	ChannelsIncludedCount int
}

// ManagedNumberInventory is implemented by provider adapters such as DIDWW.
type ManagedNumberInventory interface {
	SearchAvailable(context.Context, AvailableSearchRequest) ([]ManagedNumberCandidate, error)
}

// ManagedNumberSelectionStore retains short-lived provider purchase inputs for
// a customer-visible opaque selection identifier.
type ManagedNumberSelectionStore interface {
	Save(context.Context, uuid.UUID, ManagedNumberCandidate) (string, error)
}

func (s *Service) SetManagedAcquisition(inventory ManagedNumberInventory, selections ManagedNumberSelectionStore) {
	s.managedInventory = inventory
	s.managedSelections = selections
}

func (s *Service) SearchAvailable(
	ctx context.Context,
	organizationID uuid.UUID,
	req AvailableSearchRequest,
) ([]AvailableNumberResponse, error) {
	if err := validateOrganizationID(organizationID); err != nil {
		return nil, err
	}
	if s.managedInventory == nil || s.managedSelections == nil {
		return nil, apperror.NewServiceUnavailable("managed number inventory is not configured", nil)
	}

	countryCode, err := normalizeCountryCode(req.CountryCode)
	if err != nil {
		return nil, err
	}
	contains, err := normalizeNumberContains(req.Contains)
	if err != nil {
		return nil, err
	}
	req.CountryCode = countryCode
	req.Contains = contains

	candidates, err := s.managedInventory.SearchAvailable(ctx, req)
	if err != nil {
		return nil, apperror.NewServiceUnavailable("search managed number inventory", err)
	}

	result := make([]AvailableNumberResponse, 0, len(candidates))
	for _, candidate := range candidates {
		candidate.Provider = strings.TrimSpace(candidate.Provider)
		candidate.ProviderInventoryID = strings.TrimSpace(candidate.ProviderInventoryID)
		candidate.ProviderProductID = strings.TrimSpace(candidate.ProviderProductID)
		if candidate.Provider == "" || candidate.ProviderInventoryID == "" || candidate.ProviderProductID == "" {
			return nil, apperror.NewServiceUnavailable("managed number provider returned incomplete purchase metadata", nil)
		}
		candidate.Number, err = normalizeNumber(candidate.Number)
		if err != nil {
			return nil, apperror.NewServiceUnavailable("managed number provider returned an invalid number", err)
		}
		candidate.CountryCode, err = normalizeCountryCode(candidate.CountryCode)
		if err != nil {
			return nil, apperror.NewServiceUnavailable("managed number provider returned an invalid country", err)
		}
		if candidate.ChannelsIncludedCount <= 0 {
			return nil, apperror.NewServiceUnavailable("managed number provider returned a number without included voice capacity", nil)
		}

		selectionID, err := s.managedSelections.Save(ctx, organizationID, candidate)
		if err != nil {
			return nil, apperror.NewServiceUnavailable("store managed number selection", err)
		}
		result = append(result, AvailableNumberResponse{
			SelectionID:  selectionID,
			Number:       candidate.Number,
			CountryCode:  candidate.CountryCode,
			VoiceEnabled: true,
		})
	}

	return result, nil
}
