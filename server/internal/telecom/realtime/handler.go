package realtime

import (
	"net/http"

	"github.com/leamout/leamout/internal/runtime/middleware"
	"github.com/leamout/leamout/pkg/apperror"
	"github.com/leamout/leamout/pkg/httputil"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	if service == nil {
		panic("realtime: service is required")
	}
	return &Handler{service: service}
}

func (h *Handler) IssueICECredentials(w http.ResponseWriter, r *http.Request) {
	organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, apperror.NewBadRequest("organization context required"))
		return
	}

	credentials, err := h.service.Issue(organizationID)
	if err != nil {
		httputil.Error(w, apperror.NewInternal("issue TURN credentials", err))
		return
	}
	httputil.OK(w, credentials)
}
