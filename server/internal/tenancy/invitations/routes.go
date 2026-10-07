package invitations

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

func RegisterRoutes(router chi.Router, handler *Handler, requireSession func(http.Handler) http.Handler, organizationAccess func(http.Handler) http.Handler) {
	router.With(organizationAccess).Post("/organizations/{organization_id}/invitations", handler.Create)
	router.With(organizationAccess).Delete("/organizations/{organization_id}/invitations/{invitation_id}", handler.Revoke)
	router.With(requireSession).Post("/invitations/accept", handler.Accept)
}
