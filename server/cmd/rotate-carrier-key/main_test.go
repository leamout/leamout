package main

import (
	"strings"
	"testing"

	"github.com/leamout/leamout/internal/security/secrets"
)

func testKey(seed byte) string {
	// 32-byte raw-url-safe base64 without padding.
	alphabet := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	return strings.Repeat(string(alphabet[int(seed)%len(alphabet)]), 43)
}

func TestRotateValueReencryptsOldCiphertext(t *testing.T) {
	oldCipher, err := secrets.New(testKey(0))
	if err != nil {
		t.Fatal(err)
	}
	newCipher, err := secrets.New(testKey(1))
	if err != nil {
		t.Fatal(err)
	}
	ciphertext, err := oldCipher.Encrypt("carrier-secret")
	if err != nil {
		t.Fatal(err)
	}

	rotated, changed, err := rotateValue(&ciphertext, oldCipher, newCipher)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("expected ciphertext to be re-encrypted")
	}
	plaintext, err := newCipher.Decrypt(*rotated)
	if err != nil {
		t.Fatalf("decrypt with new key: %v", err)
	}
	if plaintext != "carrier-secret" {
		t.Fatalf("plaintext = %q", plaintext)
	}
}

func TestRotateValueIsRerunnable(t *testing.T) {
	oldCipher, err := secrets.New(testKey(0))
	if err != nil {
		t.Fatal(err)
	}
	newCipher, err := secrets.New(testKey(1))
	if err != nil {
		t.Fatal(err)
	}
	ciphertext, err := newCipher.Encrypt("carrier-secret")
	if err != nil {
		t.Fatal(err)
	}

	rotated, changed, err := rotateValue(&ciphertext, oldCipher, newCipher)
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Fatal("already-rotated ciphertext should be left unchanged")
	}
	if *rotated != ciphertext {
		t.Fatal("already-rotated ciphertext changed")
	}
}

func TestRotateValueRejectsUnknownCiphertext(t *testing.T) {
	oldCipher, err := secrets.New(testKey(0))
	if err != nil {
		t.Fatal(err)
	}
	newCipher, err := secrets.New(testKey(1))
	if err != nil {
		t.Fatal(err)
	}
	otherCipher, err := secrets.New(testKey(2))
	if err != nil {
		t.Fatal(err)
	}
	ciphertext, err := otherCipher.Encrypt("carrier-secret")
	if err != nil {
		t.Fatal(err)
	}

	if _, _, err := rotateValue(&ciphertext, oldCipher, newCipher); err == nil {
		t.Fatal("expected ciphertext encrypted with an unknown key to fail")
	}
}
