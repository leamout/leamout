package realtime

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1" // #nosec G505 -- coturn's TURN REST authentication protocol requires HMAC-SHA1.
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	config Config
	now    func() time.Time
	random io.Reader
}

func NewService(config Config) (*Service, error) {
	for index := range config.URLs {
		config.URLs[index] = strings.TrimSpace(config.URLs[index])
	}
	if err := validateConfig(config); err != nil {
		return nil, err
	}
	return &Service{config: config, now: time.Now, random: rand.Reader}, nil
}

// Issue creates short-lived credentials compatible with coturn's
// use-auth-secret mechanism. The expiry is encoded as the first username field
// so coturn can reject the credential without calling the Leamout API.
func (s *Service) Issue(organizationID uuid.UUID) (ICECredentials, error) {
	if organizationID == uuid.Nil {
		return ICECredentials{}, fmt.Errorf("organization id is required")
	}

	nonce := make([]byte, 16)
	if _, err := io.ReadFull(s.random, nonce); err != nil {
		return ICECredentials{}, fmt.Errorf("generate TURN credential nonce: %w", err)
	}

	expiresAt := s.now().UTC().Add(s.config.TTL)
	username := fmt.Sprintf("%d:%s:%s", expiresAt.Unix(), organizationID, hex.EncodeToString(nonce))
	digest := hmac.New(sha1.New, []byte(s.config.AuthSecret))
	_, _ = digest.Write([]byte(username))
	credential := base64.StdEncoding.EncodeToString(digest.Sum(nil))

	return ICECredentials{
		ICEServers: []ICEServer{{
			URLs:       append([]string(nil), s.config.URLs...),
			Username:   username,
			Credential: credential,
		}},
		ExpiresAt: expiresAt,
	}, nil
}
