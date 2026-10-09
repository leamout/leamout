package numbers

import (
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/leamout/leamout/server/pkg/apperror"
	"github.com/leamout/leamout/server/pkg/listquery"
)

var e164 = regexp.MustCompile(`^\+[1-9][0-9]{6,14}$`)

func validateOrganization(id uuid.UUID) error {
	if id == uuid.Nil {
		return apperror.NewBadRequest("organization context required")
	}
	return nil
}

func validateIDs(organizationID, id uuid.UUID) error {
	if err := validateOrganization(organizationID); err != nil {
		return err
	}
	if id == uuid.Nil {
		return apperror.NewBadRequest("number id is required")
	}
	return nil
}

func normalizeCreate(req *CreateRequest) error {
	req.Number = strings.TrimSpace(req.Number)
	req.CountryCode = strings.ToUpper(strings.TrimSpace(req.CountryCode))
	if !e164.MatchString(req.Number) {
		return apperror.NewBadRequest("number must be in E.164 format")
	}
	if len(req.CountryCode) != 2 ||
		req.CountryCode[0] < 'A' || req.CountryCode[0] > 'Z' ||
		req.CountryCode[1] < 'A' || req.CountryCode[1] > 'Z' {
		return apperror.NewBadRequest("country_code must be a two-letter ISO country code")
	}
	if req.TrunkID != nil && *req.TrunkID == uuid.Nil {
		return apperror.NewBadRequest("trunk_id must be a valid UUID")
	}
	return nil
}

func validateUpdate(req UpdateRequest) error {
	if req.VoiceEnabled == nil {
		return apperror.NewBadRequest("at least one capability is required")
	}
	return nil
}

func validateFilters(req ListRequest) error {
	if req.CountryCode != nil {
		value := *req.CountryCode
		if len(value) != 2 || value[0] < 'A' || value[0] > 'Z' || value[1] < 'A' || value[1] > 'Z' {
			return apperror.NewBadRequest("country_code must be a two-letter ISO country code")
		}
	}
	if err := listquery.Enum("status", req.Status, "active", "disabled", "porting"); err != nil {
		return err
	}
	if err := listquery.ID("trunk_id", req.TrunkID); err != nil {
		return err
	}
	return nil
}
