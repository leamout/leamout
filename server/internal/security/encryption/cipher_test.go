package encryption

import "testing"

func TestCipherRoundTrip(t *testing.T) {
	cipher, err := New("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	first, err := cipher.Encrypt("carrier-secret")
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}
	second, err := cipher.Encrypt("carrier-secret")
	if err != nil {
		t.Fatalf("Encrypt() second error = %v", err)
	}
	if first == second {
		t.Fatal("Encrypt() reused a nonce")
	}
	plaintext, err := cipher.Decrypt(first)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}
	if plaintext != "carrier-secret" {
		t.Fatalf("Decrypt() = %q", plaintext)
	}
}

func TestCipherRejectsTampering(t *testing.T) {
	cipher, err := New("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	encrypted, err := cipher.Encrypt("carrier-secret")
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	last := encrypted[len(encrypted)-1]
	replacement := byte('A')
	if last == replacement {
		replacement = 'B'
	}
	tampered := encrypted[:len(encrypted)-1] + string(replacement)
	if _, err := cipher.Decrypt(tampered); err == nil {
		t.Fatal("Decrypt() accepted tampered ciphertext")
	}
}

func TestCipherScopedRoundTrip(t *testing.T) {
	cipher, err := New("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	ciphertext, err := cipher.EncryptForScope("organization/org-a/carrier/connection-a/outbound", "carrier-secret")
	if err != nil {
		t.Fatalf("EncryptForScope() error = %v", err)
	}
	plaintext, err := cipher.DecryptForScope("organization/org-a/carrier/connection-a/outbound", ciphertext)
	if err != nil {
		t.Fatalf("DecryptForScope() error = %v", err)
	}
	if plaintext != "carrier-secret" {
		t.Fatalf("DecryptForScope() = %q", plaintext)
	}
}

func TestCipherScopedRejectsCrossTenantAndPurposeReplay(t *testing.T) {
	cipher, err := New("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	ciphertext, err := cipher.EncryptForScope("organization/org-a/carrier/connection-a/outbound", "carrier-secret")
	if err != nil {
		t.Fatalf("EncryptForScope() error = %v", err)
	}
	for _, scope := range []string{
		"organization/org-b/carrier/connection-a/outbound",
		"organization/org-a/carrier/connection-a/inbound",
		"organization/org-a/carrier/connection-b/outbound",
	} {
		if _, err := cipher.DecryptForScope(scope, ciphertext); err == nil {
			t.Fatalf("DecryptForScope() accepted ciphertext in scope %q", scope)
		}
	}
	if _, err := cipher.DecryptForScope("organization/org-a/carrier/connection-a/outbound", "not-versioned"); err == nil {
		t.Fatal("DecryptForScope() accepted an unscoped envelope")
	}
}

func TestCipherScopedRequiresScope(t *testing.T) {
	cipher, err := New("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if _, err := cipher.EncryptForScope(" ", "carrier-secret"); err == nil {
		t.Fatal("EncryptForScope() accepted an empty scope")
	}
}
