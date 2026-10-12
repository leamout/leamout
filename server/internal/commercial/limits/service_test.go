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
			plans.LimitRetentionDays: 90,
			plans.LimitMonthlyRuntimeMinutes: 30000,
			plans.LimitMaxVoiceAgents: 100,
			plans.LimitMaxSIPTrunks: 30,
		},
	})
	if value.MaxConcurrentCalls == nil || *value.MaxConcurrentCalls != 50 {
		t.Fatalf("max concurrent calls = %v", value.MaxConcurrentCalls)
	}
	if value.RetentionDays == nil || *value.RetentionDays != 90 {
		t.Fatalf("retention days = %v", value.RetentionDays)
	}
	if value.MonthlyRuntimeMinutes == nil || *value.MonthlyRuntimeMinutes != 30000 {
		t.Fatalf("monthly runtime minutes = %v", value.MonthlyRuntimeMinutes)
	}
	if value.MaxVoiceAgents == nil || *value.MaxVoiceAgents != 100 {
		t.Fatalf("max voice agents = %v", value.MaxVoiceAgents)
	}
	if value.MaxSIPTrunks == nil || *value.MaxSIPTrunks != 30 {
		t.Fatalf("max SIP trunks = %v", value.MaxSIPTrunks)
	}
}

func TestFromPlanLeavesMissingLimitsUncapped(t *testing.T) {
	t.Parallel()

	value := fromPlan(plans.Plan{Limits: map[string]int64{}})
	if value.MaxConcurrentCalls != nil || value.RetentionDays != nil {
		t.Fatalf("expected missing limits to remain uncapped: %+v", value)
	}
}
