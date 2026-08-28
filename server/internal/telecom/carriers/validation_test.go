package carriers

import "testing"

func TestNormalizeCIDR(t *testing.T) {
	got, err := normalizeCIDR("192.0.2.17/24")
	if err != nil || got.String() != "192.0.2.0/24" {
		t.Fatalf("normalizeCIDR() = %v, %v", got, err)
	}
	if _, err := normalizeCIDR("192.0.2.1"); err == nil {
		t.Fatal("accepted address without prefix")
	}
}

func TestNormalizeCodecs(t *testing.T) {
	got, err := normalizeCodecs([]string{"pcmu", " OPUS ", "PCMU"})
	if err != nil || len(got) != 2 || got[0] != "PCMU" || got[1] != "OPUS" {
		t.Fatalf("normalizeCodecs() = %v, %v", got, err)
	}
	if _, err := normalizeCodecs([]string{"invalid"}); err == nil {
		t.Fatal("accepted unsupported codec")
	}
}

func TestValidateLimits(t *testing.T) {
	zero32 := int32(0)
	zero64 := int64(0)
	if validateLimits(&zero32, nil, nil) == nil || validateLimits(nil, &zero32, nil) == nil || validateLimits(nil, nil, &zero64) == nil {
		t.Fatal("accepted non-positive limit")
	}
}
