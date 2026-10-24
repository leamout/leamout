package calling

import (
	"context"
	"github.com/google/uuid"
	"github.com/leamout/leamout/server/internal/telephony/routing"
	"testing"
)

func TestAdmissionPrefix(t *testing.T) {
	trunkID := uuid.New()
	want := "telecom:admission:trunk:" + trunkID.String()
	if got := admissionPrefix(trunkID); got != want {
		t.Fatalf("admission prefix = %q, want %q", got, want)
	}
}

func TestAdmissionLimiterValidatesArgumentsBeforeRedis(t *testing.T) {
	limiter := &AdmissionLimiter{}

	if err := limiter.Acquire(
		context.Background(),
		uuid.Nil,
		"",
		routing.Limits{
			MaxCPS:             1,
			MaxConcurrentCalls: 1,
		},
	); err == nil {
		t.Fatal("expected invalid admission identity error")
	}

	if err := limiter.Bind(context.Background(), uuid.Nil, "", uuid.Nil); err == nil {
		t.Fatal("expected invalid admission bind error")
	}

	if err := limiter.Release(context.Background(), uuid.Nil, ""); err == nil {
		t.Fatal("expected invalid admission release error")
	}

	if err := limiter.Refresh(context.Background(), uuid.Nil, uuid.Nil); err == nil {
		t.Fatal("expected invalid admission refresh error")
	}
}

func TestOrganizationAdmissionPrefixIsTenantScoped(t *testing.T) {
	first, second := uuid.New(), uuid.New()

	firstKey := organizationAdmissionPrefix(first)
	secondKey := organizationAdmissionPrefix(second)

	if firstKey != "telecom:admission:organization:"+first.String() {
		t.Fatalf("organization admission key = %q", firstKey)
	}
	if firstKey == secondKey {
		t.Fatal("different organizations must never share admission capacity")
	}
	if firstKey != organizationAdmissionPrefix(first) {
		t.Fatal("organization admission keys must remain stable")
	}
	if firstKey == admissionPrefix(first) {
		t.Fatal("organization admission and trunk keys must remain separate")
	}
}
