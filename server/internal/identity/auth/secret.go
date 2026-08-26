package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/leamout/leamout/pkg/apperror"
	"golang.org/x/crypto/argon2"
)

func hashPassword(password string) (string, error) {
	salt := make([]byte, 16)

	if _, err := rand.Read(salt); err != nil {
		return "", apperror.NewInternal(
			"failed to generate password salt",
			err,
		)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		3,
		64*1024,
		4,
		32,
	)

	return fmt.Sprintf(
		"argon2id$v=19$m=65536,t=3,p=4$%s$%s",
		hex.EncodeToString(salt),
		hex.EncodeToString(hash),
	), nil
}

func verifyPassword(password string, encoded string) bool {
	parts := strings.Split(encoded, "$")

	if len(parts) != 5 || parts[0] != "argon2id" {
		return false
	}

	salt, err := hex.DecodeString(parts[3])
	if err != nil {
		return false
	}

	expected, err := hex.DecodeString(parts[4])
	if err != nil {
		return false
	}

	actual := argon2.IDKey(
		[]byte(password),
		salt,
		3,
		64*1024,
		4,
		32,
	)

	if len(actual) != len(expected) {
		return false
	}

	return subtle.ConstantTimeCompare(actual, expected) == 1
}

func generateOTP() (string, error) {
	var value uint32

	buffer := make([]byte, 4)

	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}

	value =
		uint32(buffer[0])<<24 |
			uint32(buffer[1])<<16 |
			uint32(buffer[2])<<8 |
			uint32(buffer[3])

	value %= 1_000_000

	return fmt.Sprintf("%06d", value), nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func subtleCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}

	return subtle.ConstantTimeCompare(
		[]byte(a),
		[]byte(b),
	) == 1
}
