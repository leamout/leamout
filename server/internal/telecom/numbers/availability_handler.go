package numbers

import (
	"net/http"

	"github.com/leamout/leamout/pkg/httputil"
)

func (h *Handler) SearchAvailable(w http.ResponseWriter, r *http.Request) {
	organizationID, err := organizationID(r)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	items, err := h.service.SearchAvailable(r.Context(), organizationID, AvailableSearchRequest{
		CountryCode: r.URL.Query().Get("country_code"),
		Contains:    r.URL.Query().Get("contains"),
	})
	if err != nil {
		httputil.Error(w, err)
		return
	}

	httputil.OK(w, map[string]any{"numbers": items})
}
