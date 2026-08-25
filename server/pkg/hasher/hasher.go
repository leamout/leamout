package hasher

import (
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HA1Hashes contains the pre-computed HA1 hashes required by OpenSIPS
// for SIP Digest Authentication.
// HA1 = HASH("username:domain:password")
type HA1Hashes struct {
	MD5        string
	SHA256     string
	SHA512_256 string
}

// ComputeHA1 generates all three HA1 hash variants that OpenSIPS supports.
// The raw input is: "username:domain:password"
func ComputeHA1(username, domain, password string) HA1Hashes {
	raw := fmt.Sprintf("%s:%s:%s", username, domain, password)

	md5Hash := md5.Sum([]byte(raw))
	sha256Hash := sha256.Sum256([]byte(raw))
	sha512Hash := sha512.Sum512_256([]byte(raw))

	return HA1Hashes{
		MD5:        hex.EncodeToString(md5Hash[:]),
		SHA256:     hex.EncodeToString(sha256Hash[:]),
		SHA512_256: hex.EncodeToString(sha512Hash[:]),
	}
}

// ComputeHA1MD5 generates only the MD5 HA1 hash (most common for SIP).
func ComputeHA1MD5(username, domain, password string) string {
	raw := fmt.Sprintf("%s:%s:%s", username, domain, password)
	hash := md5.Sum([]byte(raw))
	return hex.EncodeToString(hash[:])
}

// HashPassword hashes a plaintext password using bcrypt.
// Returns the hashed password string suitable for storage in the users.password_hash column.
func HashPassword(password string) (string, error) {
	// Use bcrypt with default cost (10) - good balance of security and performance
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedBytes), nil
}

// VerifyPassword compares a plaintext password against a bcrypt hash.
// Returns true if they match, false otherwise.
func VerifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
