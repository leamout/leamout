package kindesync

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/leamout/leamout/server/internal/database/sqlc"
	"github.com/leamout/leamout/server/internal/integrations/kinde"
	"github.com/leamout/leamout/server/pkg/apperror"
	"github.com/leamout/leamout/server/pkg/httputil"
)

const maxWebhookBytes = 65536

type Handler struct {
	queries *sqlc.Queries
	client  *kinde.Client
}

func NewHandler(queries *sqlc.Queries, client *kinde.Client) *Handler {
	return &Handler{queries: queries, client: client}
}

// Receive applies a signed Kinde identity event exactly once. Never create
// users or grant access based on a webhook, only change an existing identity.
func (h *Handler) Receive(w http.ResponseWriter, r *http.Request) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/jwt" {
		httputil.Error(w, apperror.NewBadRequest("Kinde webhook requires application/jwt"))
		return
	}
	webhookID := strings.TrimSpace(r.Header.Get("webhook-id"))
	if webhookID == "" || len(webhookID) > 256 {
		httputil.Error(w, apperror.NewBadRequest("invalid Kinde webhook delivery ID"))
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxWebhookBytes))
	if err != nil {
		httputil.Error(w, apperror.NewPayloadTooLarge("Kinde webhook payload is too large"))
		return
	}
	event, err := h.client.VerifyWebhook(r.Context(), string(body))
	if err != nil {
		if errors.Is(err, kinde.ErrWebhookUnavailable) {
			httputil.Error(w, apperror.NewServiceUnavailable("identity webhook verification is unavailable", err))
			return
		}
		httputil.Error(w, apperror.NewUnauthorized("invalid Kinde webhook"))
		return
	}
	if event.Type != "user.updated" && event.Type != "user.deleted" {
		w.WriteHeader(http.StatusOK)
		return
	}

	disabled := event.Type == "user.deleted" ||
		(event.Data.User.IsSuspended != nil && *event.Data.User.IsSuspended)
	name := ""
	email := ""
	if event.Data.User.FirstName != nil && event.Data.User.LastName != nil {
		name = strings.TrimSpace(*event.Data.User.FirstName + " " + *event.Data.User.LastName)
	}
	if event.Data.User.Email != nil {
		email = strings.TrimSpace(*event.Data.User.Email)
	}
	_, err = h.queries.ApplyKindeIdentityEvent(r.Context(), sqlc.ApplyKindeIdentityEventParams{
		Issuer: h.client.Issuer(),
		EventID: event.EventID,
		Subject: event.Subject(),
		EventType: event.Type,
		OccurredAt: event.Timestamp,
		DisableUser: disabled,
		ProfileName: name,
		ProfileEmail: email,
	})
	if err != nil {
		httputil.Error(w, apperror.NewServiceUnavailable("identity synchronization unavailable", err))
		return
	}
	w.WriteHeader(http.StatusOK)
}
