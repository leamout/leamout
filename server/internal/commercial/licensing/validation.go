package licensing

import (
	"crypto/ed25519"
	"encoding/base64"
	"strings"
	"time"

	"github.com/google/uuid"
)

func validateID(id uuid.UUID, missing error) error {
	if id == uuid.Nil {
		return missing
	}
	return nil
}

func normalizeStatus(status Status) (Status, error) {
	value := Status(strings.ToLower(strings.TrimSpace(string(status))))
	switch value {
	case StatusPending, StatusActive, StatusSuspended, StatusExpired, StatusRevoked:
		return value, nil
	default:
		return "", ErrInvalidStatus
	}
}

func validateTransition(from, to Status) error {
	if from == to {
		return nil
	}
	switch from {
	case StatusPending:
		if to == StatusActive || to == StatusRevoked {
			return nil
		}
	case StatusActive:
		if to == StatusSuspended || to == StatusExpired || to == StatusRevoked {
			return nil
		}
	case StatusSuspended:
		if to == StatusActive || to == StatusExpired || to == StatusRevoked {
			return nil
		}
	case StatusExpired, StatusRevoked:
		return ErrInvalidTransition
	}
	return ErrInvalidTransition
}

func normalizeCreate(input CreateInput, issuedAt time.Time) (CreateInput, time.Time, error) {
	issuedAt = issuedAt.UTC()
	if issuedAt.IsZero() {
		return CreateInput{}, time.Time{}, ErrInvalidExpiration
	}
	if input.SigningKeyID != nil {
		keyID := strings.TrimSpace(*input.SigningKeyID)
		if keyID == "" {
			return CreateInput{}, time.Time{}, ErrSigningKeyRequired
		}
		if strings.IndexFunc(keyID, func(r rune) bool { return r == ' ' || r == '\t' || r == '\n' || r == '\r' }) >= 0 {
			return CreateInput{}, time.Time{}, ErrInvalidSigningKey
		}
		input.SigningKeyID = &keyID
	}
	if input.ExpiresAt != nil {
		expiresAt := input.ExpiresAt.UTC()
		if !expiresAt.After(issuedAt) {
			return CreateInput{}, time.Time{}, ErrInvalidExpiration
		}
		input.ExpiresAt = &expiresAt
	}
	return input, issuedAt, nil
}

func normalizeDeployment(input ActivateDeploymentInput) (ActivateDeploymentInput, error) {
	deploymentID, err := normalizeDeploymentID(input.DeploymentID)
	if err != nil {
		return ActivateDeploymentInput{}, err
	}
	input.DeploymentID = deploymentID
	publicKey, err := normalizeDeploymentPublicKey(input.PublicKey)
	if err != nil {
		return ActivateDeploymentInput{}, err
	}
	input.PublicKey = publicKey
	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return ActivateDeploymentInput{}, ErrInvalidDeploymentName
		}
		input.Name = &name
	}
	return input, nil
}

func normalizeDeploymentID(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", ErrDeploymentIDRequired
	}
	if strings.IndexFunc(value, func(r rune) bool { return r == ' ' || r == '\t' || r == '\n' || r == '\r' }) >= 0 {
		return "", ErrInvalidDeploymentID
	}
	return value, nil
}

func normalizeDeploymentPublicKey(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", ErrDeploymentPublicKeyRequired
	}
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil || len(decoded) != ed25519.PublicKeySize {
		return "", ErrInvalidDeploymentPublicKey
	}
	return base64.RawURLEncoding.EncodeToString(decoded), nil
}
