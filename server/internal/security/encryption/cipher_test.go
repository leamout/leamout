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
