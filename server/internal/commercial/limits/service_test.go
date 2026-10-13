package limits

import (
	"testing"
	"time"

	"github.com/leamout/leamout/server/internal/commercial/plans"
	"github.com/leamout/leamout/server/internal/commercial/subscriptions"
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
	start := now.Add(-time.Hour)
	end := now.Add(time.Hour)
	pastDue := now.Add(-24 * time.Hour)
	graceEnd := pastDue.AddDate(0, 0, 7)
	expiredPeriod := now.Add(-time.Minute)
	cases := []struct {
		name         string
		subscription *subscriptions.Subscription
		want         bool
	}{
		{
			name: "missing",
		},
		{
			name: "active",
			subscription: &subscriptions.Subscription{
				Status:             subscriptions.StatusActive,
				CurrentPeriodStart: &start,
				CurrentPeriodEnd:   &end,
			},
			want: true,
		},
		{
			name: "expired",
			subscription: &subscriptions.Subscription{
				Status:             subscriptions.StatusActive,
				CurrentPeriodStart: &start,
				CurrentPeriodEnd:   &now,
			},
		},
		{
			name: "past_due_without_grace",
			subscription: &subscriptions.Subscription{
				Status:             subscriptions.StatusPastDue,
				CurrentPeriodStart: &start,
				CurrentPeriodEnd:   &end,
			},
		},
		{
			name: "past_due_during_grace",
			subscription: &subscriptions.Subscription{
				Status:             subscriptions.StatusPastDue,
				CurrentPeriodStart: &start,
				CurrentPeriodEnd:   &end,
				PastDueAt:          &pastDue,
				GraceEndsAt:        &graceEnd,
			},
			want: true,
		},
		{
			name: "grace_continues_after_provider_period_ends",
			subscription: &subscriptions.Subscription{
				Status:             subscriptions.StatusPastDue,
				CurrentPeriodStart: &start,
				CurrentPeriodEnd:   &expiredPeriod,
				PastDueAt:          &pastDue,
				GraceEndsAt:        &graceEnd,
			},
			want: true,
		},
		{
			name: "grace_start_inclusive",
			subscription: &subscriptions.Subscription{
				Status:             subscriptions.StatusPastDue,
				CurrentPeriodStart: &start,
				CurrentPeriodEnd:   &end,
				PastDueAt:          &now,
				GraceEndsAt:        &graceEnd,
			},
			want: true,
		},
		{
			name: "grace_end_exclusive",
			subscription: &subscriptions.Subscription{
				Status:             subscriptions.StatusPastDue,
				CurrentPeriodStart: &start,
				CurrentPeriodEnd:   &end,
				PastDueAt:          &pastDue,
				GraceEndsAt:        &now,
			},
		},
		{
			name: "future_delinquency_does_not_grant_access",
			subscription: &subscriptions.Subscription{
				Status:             subscriptions.StatusPastDue,
				CurrentPeriodStart: &start,
				CurrentPeriodEnd:   &end,
				PastDueAt:          &end,
				GraceEndsAt:        &graceEnd,
			},
		},
		{
			name: "future_period_does_not_grant_grace",
			subscription: &subscriptions.Subscription{
				Status:             subscriptions.StatusPastDue,
				CurrentPeriodStart: &end,
				CurrentPeriodEnd:   &graceEnd,
				PastDueAt:          &pastDue,
				GraceEndsAt:        &graceEnd,
			},
		},
		{
			name: "cancellation_overrides_grace",
			subscription: &subscriptions.Subscription{
				Status:             subscriptions.StatusCanceled,
				CurrentPeriodStart: &start,
				CurrentPeriodEnd:   &end,
				PastDueAt:          &pastDue,
				GraceEndsAt:        &graceEnd,
			},
		},
		{
			name: "trial",
			subscription: &subscriptions.Subscription{
				Status:             subscriptions.StatusTrialing,
				CurrentPeriodStart: &start,
				CurrentPeriodEnd:   &end,
				TrialEndsAt:        &end,
			},
			want: true,
		},
		{
			name: "trial_ended",
			subscription: &subscriptions.Subscription{
				Status:             subscriptions.StatusTrialing,
				CurrentPeriodStart: &start,
				CurrentPeriodEnd:   &end,
				TrialEndsAt:        &now,
			},
		},
		{
			name: "missing_period",
			subscription: &subscriptions.Subscription{
				Status: subscriptions.StatusActive,
			},
		},
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
