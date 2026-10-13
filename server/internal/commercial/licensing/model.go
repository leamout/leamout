package licensing

import "time"

const (
	Cloud      = "cloud"
	SelfHosted = "self_hosted"
)

// Config is installation configuration, never request-controlled.
// PublicKeys is a programmatic trust override for tests and embedding; it has no environment tag.
// Distributed binaries pin the issuer keyring via BuildPublicKeys.
type Config struct {
	Mode         string `env:"DEPLOYMENT_MODE" envDefault:"cloud"`
	DeploymentID string `env:"LICENSE_DEPLOYMENT_ID"`
	File         string `env:"LICENSE_FILE"`
	PublicKeys   map[string]string
}

type Claims struct {
	Version            int       `json:"version"`
	LicenseID          string    `json:"license_id"`
	CustomerID         string    `json:"customer_id"`
	DeploymentID       string    `json:"deployment_id"`
	ValidFrom          time.Time `json:"valid_from"`
	ExpiresAt          time.Time `json:"expires_at"`
	GraceUntil         time.Time `json:"grace_until"`
	MaxConcurrentCalls int64     `json:"max_concurrent_calls"`
	MaxNodes           int64     `json:"max_nodes"`
	Features           []string  `json:"features"`
}

// Payload is base64 of the exact JSON bytes covered by the signature.
type Document struct {
	KeyID     string `json:"key_id"`
	Payload   string `json:"payload"`
	Signature string `json:"signature"`
}

type Status struct {
	Mode   string  `json:"deployment_mode"`
	State  string  `json:"state"`
	Claims *Claims `json:"license,omitempty"`
}

// BuildPublicKeys is a comma-separated key-id:base64-public-key keyring set by
// the issuer at build time with -ldflags. It must never contain private keys.
var BuildPublicKeys string
