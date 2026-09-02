package licensing

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/runtime/middleware"
	"github.com/leamout/leamout/pkg/apperror"
	"github.com/leamout/leamout/pkg/helper"
	"github.com/leamout/leamout/pkg/httputil"
)

// Handler exposes organization-scoped customer license operations.
type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	organizationID, err := requestOrganizationID(r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	licenses, err := h.service.List(r.Context(), organizationID)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	responses := make([]licenseResponse, 0, len(licenses))
	for _, license := range licenses {
		responses = append(responses, newLicenseResponse(license))
	}
	httputil.OK(w, map[string]any{"licenses": responses})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	organizationID, licenseID, err := requestLicenseIDs(r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	license, err := h.service.Get(r.Context(), organizationID, licenseID)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	httputil.OK(w, newLicenseResponse(license))
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	organizationID, err := requestOrganizationID(r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	license, err := h.service.Create(r.Context(), organizationID, CreateInput{})
	if err != nil {
		httputil.Error(w, err)
		return
	}
	httputil.Created(w, newLicenseResponse(license))
}

func (h *Handler) ListDeployments(w http.ResponseWriter, r *http.Request) {
	organizationID, licenseID, err := requestLicenseIDs(r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	deployments, err := h.service.ListDeployments(r.Context(), organizationID, licenseID)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	responses := make([]deploymentResponse, 0, len(deployments))
	for _, deployment := range deployments {
		responses = append(responses, newDeploymentResponse(deployment))
	}
	httputil.OK(w, map[string]any{"deployments": responses})
}

func (h *Handler) ActivateDeployment(w http.ResponseWriter, r *http.Request) {
	organizationID, licenseID, err := requestLicenseIDs(r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	input, err := helper.DecodeJSON[ActivateDeploymentInput](r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	deployment, err := h.service.ActivateDeployment(r.Context(), organizationID, licenseID, input)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	httputil.OK(w, newDeploymentResponse(deployment))
}

func (h *Handler) HeartbeatDeployment(w http.ResponseWriter, r *http.Request) {
	organizationID, licenseID, deploymentID, err := requestDeploymentIDs(r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	deployment, err := h.service.TouchDeployment(r.Context(), organizationID, licenseID, deploymentID)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	httputil.OK(w, newDeploymentResponse(deployment))
}

func (h *Handler) DeactivateDeployment(w http.ResponseWriter, r *http.Request) {
	organizationID, licenseID, deploymentID, err := requestDeploymentIDs(r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	deployment, err := h.service.DeactivateDeployment(r.Context(), organizationID, licenseID, deploymentID)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	httputil.OK(w, newDeploymentResponse(deployment))
}

func requestOrganizationID(r *http.Request) (uuid.UUID, error) {
	organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
	if !ok {
		return uuid.Nil, apperror.NewBadRequest("organization context required")
	}
	return organizationID, nil
}

func requestLicenseIDs(r *http.Request) (uuid.UUID, uuid.UUID, error) {
	organizationID, err := requestOrganizationID(r)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	licenseID, err := uuid.Parse(chi.URLParam(r, "license_id"))
	if err != nil {
		return uuid.Nil, uuid.Nil, apperror.NewBadRequest("invalid license_id")
	}
	return organizationID, licenseID, nil
}

func requestDeploymentIDs(r *http.Request) (uuid.UUID, uuid.UUID, string, error) {
	organizationID, licenseID, err := requestLicenseIDs(r)
	if err != nil {
		return uuid.Nil, uuid.Nil, "", err
	}
	deploymentID := chi.URLParam(r, "deployment_id")
	if deploymentID == "" {
		return uuid.Nil, uuid.Nil, "", ErrDeploymentIDRequired
	}
	return organizationID, licenseID, deploymentID, nil
}
