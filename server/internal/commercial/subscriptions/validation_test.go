package subscriptions

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestValidateSync(t *testing.T) {
	start := time.Now().UTC()
	end := start.Add(time.Hour)
	err := validateSync(SyncRequest{
		PlanID:             uuid.New(),
		Status:             StatusActive,
		CurrentPeriodStart: &start,
		CurrentPeriodEnd:   &end,
	})
	if err != nil {
		t.Fatalf("validateSync() error = %v", err)
	}
}

func TestValidateSyncRejectsInvalidPeriod(t *testing.T) {
	start := time.Now().UTC()
	end := start.Add(-time.Hour)
	err := validateSync(SyncRequest{
		PlanID:             uuid.New(),
		Status:             StatusActive,
		CurrentPeriodStart: &start,
		CurrentPeriodEnd:   &end,
	})
	if err == nil {
		t.Fatal("expected invalid subscription period to be rejected")
	}
}
