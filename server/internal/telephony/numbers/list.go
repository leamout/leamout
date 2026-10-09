package numbers

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/leamout/leamout/server/pkg/apperror"
	"github.com/leamout/leamout/server/pkg/listquery"
)

type ListRequest struct {
	Status       *string
	CountryCode  *string
	TrunkID      *uuid.UUID
	VoiceEnabled *bool
}

func parseFilters(r *http.Request) (ListRequest, error) {
	p := listquery.Parser{Values: r.URL.Query()}
	req := ListRequest{
		Status:       p.Text("status"),
		CountryCode:  p.Text("country_code"),
		TrunkID:      p.UUID("trunk_id"),
		VoiceEnabled: p.Bool("voice_enabled"),
	}
	if req.CountryCode != nil {
		value := strings.ToUpper(*req.CountryCode)
		req.CountryCode = &value
	}
	return req, p.Err
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
