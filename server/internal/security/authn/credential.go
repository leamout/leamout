package authn

import (
	"errors"

	"github.com/google/uuid"
)

type CredentialType string

const (
	CredentialSession CredentialType = "session"
	CredentialAPIKey  CredentialType = "api_key"
)

type Credential struct {
	ID   uuid.UUID
	Type CredentialType
}

type CredentialInput struct {
	Type  CredentialType
	Value string
}

func (c CredentialInput) Validate() error {
	if c.Type == "" || c.Value == "" {
		return errors.New("invalid credential")
	}

	switch c.Type {
	case CredentialSession, CredentialAPIKey:
		return nil
	default:
		return errors.New("invalid credential")
	}
}
