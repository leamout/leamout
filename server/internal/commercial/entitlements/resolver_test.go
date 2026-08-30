package entitlements

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
)

type fakeSource struct {
	plan, organization, license []Entitlement
	err                         error
}

func (f fakeSource) ListPlan(context.Context, uuid.UUID) ([]Entitlement, error) {
	return f.plan, f.err
}

func (f fakeSource) ListOrganization(context.Context, uuid.UUID) ([]Entitlement, error) {
	return f.organization, f.err
}

func (f fakeSource) ListLicense(context.Context, uuid.UUID, uuid.UUID) ([]Entitlement, error) {
	return f.license, f.err
}

func TestResolverAppliesPrecedenceValidityAndStableOrdering(t *testing.T) {
	now := time.Date(2026, time.August, 30, 12, 0, 0, 0, time.UTC)
	expired := now.Add(-time.Nanosecond)
	trueValue, falseValue := true, false
	planLimit, organizationLimit := int64(100), int64(0)
	resolver, err := NewResolver(fakeSource{
		plan: []Entitlement{
			{Key: "voice.recording", Kind: KindFeature, Enabled: &trueValue},
			{Key: "voice.concurrent_calls", Kind: KindLimit, Limit: &planLimit},
			{Key: "expired.plan", Kind: KindFeature, Enabled: &trueValue, ExpiresAt: &expired},
		},
		organization: []Entitlement{
			{Key: "voice.recording", Kind: KindFeature, Enabled: &falseValue},
			{Key: "voice.concurrent_calls", Kind: KindLimit, Limit: &organizationLimit},
		},
		license: []Entitlement{
			{Key: "voice.recording", Kind: KindFeature, Enabled: &trueValue},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	resolved, err := resolver.Resolve(context.Background(), ResolveRequest{
		OrganizationID: uuid.New(), PlanID: uuid.New(), LicenseID: uuid.New(), At: now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := keys(resolved), []string{"voice.concurrent_calls", "voice.recording"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("keys = %v, want %v", got, want)
	}
	if resolved[0].Limit == nil || *resolved[0].Limit != 0 {
		t.Fatalf("zero-valued organization override was not preserved: %+v", resolved[0])
	}
	if resolved[1].Enabled == nil || !*resolved[1].Enabled {
		t.Fatalf("license override did not win: %+v", resolved[1])
	}
}

func TestResolverRejectsKindConflict(t *testing.T) {
	now := time.Now()
	enabled := true
	limit := int64(1)
	resolver, _ := NewResolver(fakeSource{
		plan:         []Entitlement{{Key: "voice.calls", Kind: KindLimit, Limit: &limit}},
		organization: []Entitlement{{Key: "voice.calls", Kind: KindFeature, Enabled: &enabled}},
	})
	_, err := resolver.Resolve(context.Background(), ResolveRequest{
		OrganizationID: uuid.New(), PlanID: uuid.New(), At: now,
	})
	if !errors.Is(err, ErrKindConflict) {
		t.Fatalf("expected kind conflict, got %v", err)
	}
}

func TestResolverWrapsSourceError(t *testing.T) {
	resolver, _ := NewResolver(fakeSource{err: errors.New("load failed")})
	_, err := resolver.Resolve(context.Background(), ResolveRequest{
		OrganizationID: uuid.New(), PlanID: uuid.New(), At: time.Now(),
	})
	if err == nil {
		t.Fatal("expected source error")
	}
}

func keys(entitlements []Entitlement) []string {
	result := make([]string, len(entitlements))
	for index, entitlement := range entitlements {
		result[index] = entitlement.Key
	}
	return result
}
