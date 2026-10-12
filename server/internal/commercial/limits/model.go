package limits

import "time"

type Effective struct {
	MaxConcurrentCalls    *int64
	RetentionDays         *int64
	MonthlyRuntimeMinutes *int64
	MaxVoiceAgents        *int64
	MaxSIPTrunks          *int64
	PeriodStart           time.Time
	PeriodEnd             time.Time
}

type RuntimeUsage struct {
	PeriodStart      time.Time
	PeriodEnd        time.Time
	UsedSeconds      int64
	AllowanceSeconds int64
}
