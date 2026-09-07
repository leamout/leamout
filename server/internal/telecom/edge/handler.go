package edge

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/telecom/routing"
	"github.com/leamout/leamout/pkg/helper"
	"github.com/leamout/leamout/pkg/httputil"
)

type Handler struct {
	service        *Service
	inboundRouting interface {
		ResolveManagedInboundDelivery(context.Context, routing.InboundRequest) (routing.ManagedInboundDeliveryDecision, error)
	}
	secret string
}

func NewHandler(service *Service, inboundRouting interface {
	ResolveManagedInboundDelivery(context.Context, routing.InboundRequest) (routing.ManagedInboundDeliveryDecision, error)
}, secret string) *Handler {
	return &Handler{service: service, inboundRouting: inboundRouting, secret: strings.TrimSpace(secret)}
}

func (h *Handler) Admit(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	req, err := helper.DecodeJSON[Request](r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	if req.Username == "" || req.Realm == "" || !e164(req.From) || !e164(req.To) {
		http.Error(w, "invalid admission request", http.StatusBadRequest)
		return
	}
	decision, err := h.service.Admit(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrDenied) {
			w.Header().Set("Cache-Control", "no-store")
			httputil.JSON(w, http.StatusForbidden, map[string]bool{"allowed": false})
			return
		}
		http.Error(w, "admission unavailable", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	httputil.JSON(w, http.StatusOK, decision)
}

type inboundRequest struct {
	SourceIP     string `json:"source_ip"`
	CalledNumber string `json:"called_number"`
	CallerNumber string `json:"caller_number"`
}

type inboundResponse struct {
	Allowed             bool      `json:"allowed"`
	OrganizationID      uuid.UUID `json:"organization_id"`
	CarrierConnectionID uuid.UUID `json:"carrier_connection_id"`
	PhoneNumberID       uuid.UUID `json:"phone_number_id"`
	RuntimeAttachmentID uuid.UUID `json:"runtime_attachment_id"`
	DeploymentID        uuid.UUID `json:"deployment_id"`
	DeploymentIdentity  string    `json:"deployment_identity"`
	RouteURI            string    `json:"route_uri"`
}

// ResolveInbound authorizes the hosted managed edge to forward a carrier
// INVITE to a self-hosted runtime attachment. The destination runtime performs
// its own local number and voice-binding resolution after this hop.
func (h *Handler) ResolveInbound(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if h.inboundRouting == nil {
		http.Error(w, "inbound routing unavailable", http.StatusServiceUnavailable)
		return
	}
	req, err := helper.DecodeJSON[inboundRequest](r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	decision, err := h.inboundRouting.ResolveManagedInboundDelivery(r.Context(), routing.InboundRequest{
		SourceIP: req.SourceIP, CalledNumber: req.CalledNumber, CallerNumber: req.CallerNumber,
	})
	if err != nil {
		if errors.Is(err, routing.ErrNoRoute) || errors.Is(err, routing.ErrTenantMismatch) {
			w.Header().Set("Cache-Control", "no-store")
			httputil.JSON(w, http.StatusNotFound, map[string]bool{"allowed": false})
			return
		}
		http.Error(w, "inbound routing unavailable", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	httputil.JSON(w, http.StatusOK, inboundResponse{
		Allowed: true, OrganizationID: decision.OrganizationID,
		CarrierConnectionID: decision.CarrierConnectionID, PhoneNumberID: decision.PhoneNumberID,
		RuntimeAttachmentID: decision.RuntimeAttachmentID, DeploymentID: decision.DeploymentID,
		DeploymentIdentity: decision.DeploymentIdentity,
		RouteURI:           fmt.Sprintf("sip:%s@%s:%d;transport=%s", decision.CalledNumber, decision.IngressHost, decision.IngressPort, decision.IngressTransport),
	})
}

func (h *Handler) authorized(r *http.Request) bool {
	provided := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	return h.secret != "" && subtle.ConstantTimeCompare([]byte(provided), []byte(h.secret)) == 1
}

func e164(value string) bool {
	if len(value) < 8 || len(value) > 16 || value[0] != '+' || value[1] < '1' || value[1] > '9' {
		return false
	}
	for i := 2; i < len(value); i++ {
		if value[i] < '0' || value[i] > '9' {
			return false
		}
	}
	return true
}
