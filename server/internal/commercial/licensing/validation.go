package licensing

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
)

var signingKeyIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$`)

func (c Config) keyring() (map[string]string, error) {
	if len(c.PublicKeys) > 0 {
		return c.PublicKeys, nil
	}
	if BuildPublicKeys == "" {
		return map[string]string{}, nil
	}
	return parsePublicKeyring(BuildPublicKeys)
}

func parsePublicKeyring(value string) (map[string]string, error) {
	keys := make(map[string]string)
	for _, entry := range strings.Split(value, ",") {
		parts := strings.SplitN(entry, ":", 2)
		if len(parts) != 2 || !signingKeyIDPattern.MatchString(parts[0]) || keys[parts[0]] != "" {
			return nil, fmt.Errorf("invalid license public keyring")
		}
		key, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil || len(key) != ed25519.PublicKeySize {
			return nil, fmt.Errorf("invalid license public key %q", parts[0])
		}
		keys[parts[0]] = parts[1]
	}
	return keys, nil
}

// ValidatePublicKeyring checks release inputs before passing them to the linker.
// Only public Ed25519 keys are accepted; private keys and empty keyrings fail.
func ValidatePublicKeyring(value string) error {
	_, err := parsePublicKeyring(value)
	return err
}

func (c Config) Validate() error {
	keys, err := c.keyring()
	if err != nil {
		return err
	}
	if c.Mode != "" && c.Mode != Cloud && c.Mode != SelfHosted {
		return fmt.Errorf("unsupported DEPLOYMENT_MODE")
	}
	if c.Mode == SelfHosted && (strings.TrimSpace(c.DeploymentID) == "" || strings.TrimSpace(c.File) == "" || len(keys) == 0) {
		return fmt.Errorf("self_hosted requires LICENSE_DEPLOYMENT_ID, LICENSE_FILE and a build-time issuer keyring")
	}
	for id, encoded := range keys {
		key, err := base64.StdEncoding.DecodeString(encoded)
		if !signingKeyIDPattern.MatchString(id) || err != nil || len(key) != ed25519.PublicKeySize {
			return fmt.Errorf("invalid license public key %q", id)
		}
	}
	return nil
}

func strictJSON(data []byte, value any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("trailing JSON data")
	}
	return nil
}

func ValidateClaims(c Claims) error {
	if c.Version != 1 || strings.TrimSpace(c.LicenseID) == "" || strings.TrimSpace(c.CustomerID) == "" || strings.TrimSpace(c.DeploymentID) == "" {
		return fmt.Errorf("license identity or version invalid")
	}
	if c.ValidFrom.IsZero() || !c.ExpiresAt.After(c.ValidFrom) || c.GraceUntil.Before(c.ExpiresAt) || c.GraceUntil.After(c.ExpiresAt.AddDate(0, 0, 30)) {
		return fmt.Errorf("license validity or grace period invalid (maximum 30 days)")
	}
	if c.MaxConcurrentCalls < 1 || c.MaxNodes < 1 {
		return fmt.Errorf("licensed capacities must be positive")
	}
	seen := map[string]bool{}
	for _, feature := range c.Features {
		switch feature {
		case "sso", "scim", "advanced_rbac", "retention_policies", "private_networking", "audit_logging", "premium_templates":
		default:
			return fmt.Errorf("unsupported licensed feature %q", feature)
		}
		if seen[feature] {
			return fmt.Errorf("duplicate licensed feature %q", feature)
		}
		seen[feature] = true
	}
	return nil
}

func Verify(data []byte, publicKeys map[string]string, deploymentID string) (Claims, error) {
	var doc Document
	if err := strictJSON(data, &doc); err != nil {
		return Claims{}, fmt.Errorf("invalid license document")
	}
	key, err := base64.StdEncoding.DecodeString(publicKeys[doc.KeyID])
	if err != nil || len(key) != ed25519.PublicKeySize {
		return Claims{}, fmt.Errorf("unknown license signing key")
	}
	payload, err := base64.StdEncoding.DecodeString(doc.Payload)
	if err != nil {
		return Claims{}, fmt.Errorf("invalid license payload")
	}
	signature, err := base64.StdEncoding.DecodeString(doc.Signature)
	if err != nil || !ed25519.Verify(ed25519.PublicKey(key), payload, signature) {
		return Claims{}, fmt.Errorf("invalid license signature")
	}
	var claims Claims
	if err := strictJSON(payload, &claims); err != nil {
		return Claims{}, fmt.Errorf("invalid license claims")
	}
	if err := ValidateClaims(claims); err != nil {
		return Claims{}, err
	}
	if claims.DeploymentID != deploymentID {
		return Claims{}, fmt.Errorf("license belongs to another deployment")
	}
	return claims, nil
}
