package limits

import (
	"testing"

	"github.com/leamout/leamout/server/internal/commercial/plans"
)

func TestFromPlan(t *testing.T) {
	t.Parallel()

	value := fromPlan(plans.Plan{
		Limits: map[string]int64{
			plans.LimitMaxConcurrentCalls: 50,
			plans.LimitRetentionDays:      90,
		},
	})
	if value.MaxConcurrentCalls == nil || *value.MaxConcurrentCalls != 50 {
		t.Fatalf("max concurrent calls = %v", value.MaxConcurrentCalls)
	}
	if value.RetentionDays == nil || *value.RetentionDays != 90 {
		t.Fatalf("retention days = %v", value.RetentionDays)
	}
}

func TestFromPlanLeavesMissingLimitsUncapped(t *testing.T) {
	t.Parallel()

	value := fromPlan(plans.Plan{Limits: map[string]int64{}})
	if value.MaxConcurrentCalls != nil || value.RetentionDays != nil {
		t.Fatalf("expected missing limits to remain uncapped: %+v", value)
	}
}
