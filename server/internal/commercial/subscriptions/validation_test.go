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

func TestValidateSyncPaymentGrace(t *testing.T) {
	t.Parallel()

	pastDue := time.Date(2026, time.October, 12, 12, 0, 0, 0, time.UTC)
	graceEnd := pastDue.AddDate(0, 0, 7)
	beforePastDue := pastDue.Add(-time.Second)
	cases := []struct {
		name      string
		status    string
		pastDueAt *time.Time
		graceEnd  *time.Time
		wantError bool
	}{
		{
			name:      "seven_day_grace",
			status:    StatusPastDue,
			pastDueAt: &pastDue,
			graceEnd:  &graceEnd,
		},
		{
			name:      "immediate_free_policy",
			status:    StatusPastDue,
			pastDueAt: &pastDue,
			graceEnd:  &pastDue,
		},
		{
			name:      "missing_grace_end",
			status:    StatusPastDue,
			pastDueAt: &pastDue,
			wantError: true,
		},
		{
			name:      "missing_delinquency_start",
			status:    StatusPastDue,
			graceEnd:  &graceEnd,
			wantError: true,
		},
		{
			name:      "grace_cannot_precede_delinquency",
			status:    StatusPastDue,
			pastDueAt: &pastDue,
			graceEnd:  &beforePastDue,
			wantError: true,
		},
		{
			name:      "recovery_must_clear_grace",
			status:    StatusActive,
			pastDueAt: &pastDue,
			graceEnd:  &graceEnd,
			wantError: true,
		},
		{
			name:      "revocation_must_clear_grace",
			status:    StatusCanceled,
			pastDueAt: &pastDue,
			graceEnd:  &graceEnd,
			wantError: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateSync(SyncRequest{
				PlanID:      uuid.New(),
				Status:      tc.status,
				PastDueAt:   tc.pastDueAt,
				GraceEndsAt: tc.graceEnd,
			})
			if (err != nil) != tc.wantError {
				t.Fatalf("validateSync() error = %v, want error = %v", err, tc.wantError)
			}
		})
	}
}
