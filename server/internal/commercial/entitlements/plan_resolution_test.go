package entitlements

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestEffectiveForOrganizationPlanAtUsesSelectedPlanWithoutSubscriptionRead(t *testing.T) {
	t.Parallel()

	organizationID := uuid.New()
	planID := uuid.New()
	at := time.Date(2026, 8, 31, 10, 0, 0, 0, time.UTC)
	store := &fakeStore{
		plan: []Entitlement{{Key: "voice.enabled", Kind: KindFeature, Enabled: boolPtr(true)}},
		organization: []Entitlement{{
			Key:     "voice.enabled",
			Kind:    KindFeature,
			Enabled: boolPtr(false),
		}},
	}
	service := NewService(store, nil)

	set, err := service.EffectiveForOrganizationPlanAt(context.Background(), organizationID, planID, at)
	if err != nil {
		t.Fatalf("EffectiveForOrganizationPlanAt() error = %v", err)
	}
	if set.Enabled("voice.enabled") {
		t.Fatal("expected organization override to apply")
	}
}

func TestEffectiveForOrganizationPlanAtRequiresPlanID(t *testing.T) {
	t.Parallel()

	service := NewService(&fakeStore{}, nil)
	_, err := service.EffectiveForOrganizationPlanAt(context.Background(), uuid.New(), uuid.Nil, time.Now())
	if !errors.Is(err, ErrPlanIDRequired) {
		t.Fatalf("EffectiveForOrganizationPlanAt() error = %v, want %v", err, ErrPlanIDRequired)
	}
}
