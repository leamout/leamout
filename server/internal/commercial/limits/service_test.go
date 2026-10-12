package limits

import (
	"github.com/leamout/leamout/server/internal/commercial/subscriptions"
	"testing"
	"time"

	"github.com/leamout/leamout/server/internal/commercial/plans"
)

func TestFromPlan(t *testing.T) {
	t.Parallel()

	value := fromPlan(plans.Plan{
		Limits: map[string]int64{
			plans.LimitMaxConcurrentCalls:    50,
			plans.LimitRetentionDays:         90,
			plans.LimitMonthlyRuntimeMinutes: 30000,
			plans.LimitMaxVoiceAgents:        100,
			plans.LimitMaxSIPTrunks:          30,
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

func TestFromPlanRejectsMissingLimits(t *testing.T) {
	t.Parallel()

	value := fromPlan(plans.Plan{Limits: map[string]int64{}})
	if validPlanLimits(value) {
		t.Fatalf("missing cloud limits must fail closed: %+v", value)
	}
}

func TestSubscriptionPeriodActive(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, time.October, 12, 12, 0, 0, 0, time.UTC)
	start, end := now.Add(-time.Hour), now.Add(time.Hour)
	cases := []struct {
		name         string
		subscription *subscriptions.Subscription
		want         bool
	}{
		{"missing", nil, false},
		{"active", &subscriptions.Subscription{
			Status: subscriptions.StatusActive, CurrentPeriodStart: &start, CurrentPeriodEnd: &end,
		}, true},
		{"expired", &subscriptions.Subscription{
			Status: subscriptions.StatusActive, CurrentPeriodStart: &start, CurrentPeriodEnd: &now,
		}, false},
		{"past_due", &subscriptions.Subscription{
			Status: subscriptions.StatusPastDue, CurrentPeriodStart: &start, CurrentPeriodEnd: &end,
		}, false},
		{"trial", &subscriptions.Subscription{
			Status: subscriptions.StatusTrialing, CurrentPeriodStart: &start, CurrentPeriodEnd: &end, TrialEndsAt: &end,
		}, true},
		{"trial_ended", &subscriptions.Subscription{
			Status: subscriptions.StatusTrialing, CurrentPeriodStart: &start, CurrentPeriodEnd: &end, TrialEndsAt: &now,
		}, false},
		{"missing_period", &subscriptions.Subscription{Status: subscriptions.StatusActive}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := subscriptionPeriodActive(tc.subscription, now); got != tc.want {
				t.Fatalf("subscriptionPeriodActive() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestCalendarMonthCrossesYear(t *testing.T) {
	t.Parallel()
	start, end := calendarMonth(time.Date(2026, time.December, 31, 23, 59, 0, 0, time.UTC))
	if !start.Equal(time.Date(2026, time.December, 1, 0, 0, 0, 0, time.UTC)) ||
		!end.Equal(time.Date(2027, time.January, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected calendar month: %s to %s", start, end)
	}
}
