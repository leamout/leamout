package licensing

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/leamout/leamout/server/pkg/apperror"
)

const maxDocumentSize = 64 << 10

type Service struct {
	config Config
	now    func() time.Time
}

func NewService(config Config) (*Service, error) {
	if config.Mode == "" {
		config.Mode = Cloud
	}
	if err := config.Validate(); err != nil {
		return nil, err
	}
	// Copy trust configuration so callers cannot mutate the keyring after setup.
	trust, err := config.keyring()
	if err != nil {
		return nil, err
	}
	keys := make(map[string]string, len(trust))
	for id, key := range trust {
		keys[id] = key
	}
	config.PublicKeys = keys
	return &Service{config: config, now: time.Now}, nil
}

func (s *Service) SelfHosted() bool     { return s != nil && s.config.Mode == SelfHosted }
func (s *Service) DeploymentID() string { return s.config.DeploymentID }

// Read on each decision so atomic file replacement renews a running installation.
// No Cloud billing or database dependency is involved.
func (s *Service) Status() Status {
	if !s.SelfHosted() {
		return Status{Mode: Cloud, State: "cloud"}
	}
	status := Status{Mode: SelfHosted, State: "invalid"}
	file, err := os.Open(s.config.File)
	if err != nil {
		status.State = "missing"
		return status
	}
	defer func() { _ = file.Close() }()
	data, err := io.ReadAll(io.LimitReader(file, maxDocumentSize+1))
	if err != nil || len(data) > maxDocumentSize {
		return status
	}
	claims, err := Verify(data, s.config.PublicKeys, s.config.DeploymentID)
	if err != nil {
		return status
	}
	status.Claims = &claims
	now := s.now().UTC()
	switch {
	case now.Before(claims.ValidFrom):
		status.State = "not_yet_valid"
	case now.Before(claims.ExpiresAt):
		status.State = "active"
	case now.Before(claims.GraceUntil):
		status.State = "grace"
	default:
		status.State = "expired"
	}
	return status
}

func (s *Service) RequireRuntime() (Claims, error) {
	status := s.Status()
	if status.Mode != SelfHosted {
		return Claims{}, fmt.Errorf("self-hosted license requested in Cloud mode")
	}
	if status.State != "active" && status.State != "grace" {
		return Claims{}, apperror.NewPaymentRequired("self-hosted license is " + status.State)
	}
	return *status.Claims, nil
}

func (s *Service) Enabled(_ context.Context, feature string) (bool, error) {
	claims, err := s.RequireRuntime()
	if err != nil {
		return false, err
	}
	for _, enabled := range claims.Features {
		if enabled == feature {
			return true, nil
		}
	}
	return false, nil
}

// RequireDataFeature preserves licensed historical data access after expiry.
// It never authorizes new runtime sessions or accepts an invalid signature.
func (s *Service) RequireDataFeature(feature string) error {
	if !s.SelfHosted() {
		return nil
	}
	status := s.Status()
	if status.Claims == nil || status.State == "not_yet_valid" {
		return apperror.NewForbidden("verified self-hosted license required")
	}
	for _, enabled := range status.Claims.Features {
		if enabled == feature {
			return nil
		}
	}
	return apperror.NewForbidden("self-hosted feature is not licensed: " + feature)
}
