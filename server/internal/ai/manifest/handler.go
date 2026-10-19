package manifest

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/leamout/leamout/server/internal/platform/middleware"
	"github.com/leamout/leamout/server/pkg/apperror"
	"github.com/leamout/leamout/server/pkg/helper"
	"github.com/leamout/leamout/server/pkg/httputil"
)

type PortableDeploymentRequest struct {
	Package      PortablePackage      `json:"package"`
	Integrations map[string]uuid.UUID `json:"integrations,omitempty"`
	Variables    map[string]string    `json:"variables,omitempty"`
}

type Handler struct {
	deployer *Deployer
}

func NewHandler(deployer *Deployer) *Handler {
	return &Handler{deployer: deployer}
}

func RegisterRoutes(
	router chi.Router,
	handler *Handler,
	auth func(http.Handler) http.Handler,
) {
	router.Route("/agent-packages", func(r chi.Router) {
		r.Use(auth)
		r.Post("/deploy", handler.Deploy)
	})
}

func (h *Handler) Deploy(w http.ResponseWriter, r *http.Request) {
	organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, apperror.NewBadRequest("organization context required"))
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	request, err := helper.DecodeJSON[PortableDeploymentRequest](r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	if h == nil || h.deployer == nil {
		httputil.Error(w, apperror.NewServiceUnavailable("agent package deployer is unavailable", nil))
		return
	}
	result, err := h.deployer.DeployPortable(
		r.Context(),
		organizationID,
		request.Package,
		DeploymentInput{
			Integrations: request.Integrations,
			Variables:    request.Variables,
		},
	)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	httputil.Created(w, result)
}
