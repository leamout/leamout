package password

import (
	"strings"
	"testing"
)

func TestHashAndVerify(t *testing.T) {
	encoded, err := Hash("correct horse battery staple")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}
	if !Verify("correct horse battery staple", encoded) {
		t.Fatal("Verify() rejected the correct password")
	}
	if Verify("wrong password", encoded) {
		t.Fatal("Verify() accepted the wrong password")
	}
	if NeedsRehash(encoded) {
		t.Fatal("NeedsRehash() = true for a current hash")
	}
}

func TestVerifyUsesEncodedParameters(t *testing.T) {
	encoded, err := Hash("parameter-aware")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	parts := strings.Split(encoded, "$")
	if len(parts) != 5 {
		t.Fatalf("unexpected encoded hash: %q", encoded)
	}

	// Changing a persisted parameter must affect verification. This protects
	// against silently ignoring the parameters embedded in the hash string.
	parts[2] = "m=32768,t=3,p=4"
	modified := strings.Join(parts, "$")
	if Verify("parameter-aware", modified) {
		t.Fatal("Verify() ignored the encoded memory parameter")
	}
	if !NeedsRehash(modified) {
		t.Fatal("NeedsRehash() = false for non-default parameters")
	}
}

func TestVerifyRejectsMalformedHashes(t *testing.T) {
	for _, encoded := range []string{
		"",
		"bcrypt$not-argon2",
		"argon2id$v=18$m=65536,t=3,p=4$00$00",
		"argon2id$v=19$m=0,t=3,p=4$00$00",
		"argon2id$v=19$m=65536,t=3,p=4$zz$00",
	} {
		if Verify("password", encoded) {
			t.Fatalf("Verify() accepted malformed hash %q", encoded)
		}
	}
}
