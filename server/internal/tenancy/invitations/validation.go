package invitations

import (
	"github.com/coffeyvidzro/monogo/pkg/apperror"
	"net/mail"
	"strings"
)

func normalizeRequest(req CreateRequest) (CreateRequest, error) {
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Role = strings.TrimSpace(req.Role)
	if req.Role == "" {
		req.Role = "member"
	}
	address, err := mail.ParseAddress(req.Email)
	if err != nil || address.Address != req.Email {
		return CreateRequest{}, apperror.NewBadRequest("invalid invitation email")
	}
	if req.Role != "admin" && req.Role != "member" {
		return CreateRequest{}, apperror.NewBadRequest("invitation role must be admin or member")
	}
	return req, nil
}
