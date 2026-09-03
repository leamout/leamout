package token

import "testing"

func TestGenerate(t *testing.T) {
	first, err := Generate(32)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	second, err := Generate(32)
	if err != nil {
		t.Fatalf("Generate() second error = %v", err)
	}
	if len(first) != 64 {
		t.Fatalf("Generate() length = %d, want 64", len(first))
	}
	if first == second {
		t.Fatal("Generate() returned duplicate random tokens")
	}
}

func TestHashAndVerify(t *testing.T) {
	digest := Hash("opaque-secret")
	if len(digest) != 64 {
		t.Fatalf("Hash() length = %d, want 64", len(digest))
	}
	if !Verify("opaque-secret", digest) {
		t.Fatal("Verify() rejected the correct token")
	}
	if Verify("different-secret", digest) {
		t.Fatal("Verify() accepted an incorrect token")
	}
}

func TestGenerateRejectsInvalidSize(t *testing.T) {
	if _, err := Generate(0); err == nil {
		t.Fatal("Generate(0) unexpectedly succeeded")
	}
}
