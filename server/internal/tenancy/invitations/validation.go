package invitations

import (
	"net/mail"
	"strings"

	"github.com/leamout/leamout/server/pkg/apperror"
)

func validateCreate(req CreateRequest) (CreateRequest, error) {
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	address, err := mail.ParseAddress(req.Email)
	if err != nil || address.Address != req.Email {
		return req, apperror.NewBadRequest("invalid email")
	}
	req.Role = strings.TrimSpace(req.Role)
	if req.Role == "" {
		req.Role = "member"
	}
	if req.Role != "member" && req.Role != "admin" {
		return req, apperror.NewBadRequest("role must be member or admin")
	}
	return req, nil
}
