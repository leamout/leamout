package limits

import "time"

type Effective struct {
	MaxConcurrentCalls    *int64    `json:"max_concurrent_calls"`
	RetentionDays         *int64    `json:"retention_days"`
	MonthlyRuntimeMinutes *int64    `json:"monthly_runtime_minutes"`
	MaxVoiceAgents        *int64    `json:"max_voice_agents"`
	MaxSIPTrunks          *int64    `json:"max_sip_trunks"`
	PeriodStart           time.Time `json:"period_start"`
	PeriodEnd             time.Time `json:"period_end"`
}

type RuntimeUsage struct {
	Unlimited        bool      `json:"unlimited"`
	PeriodStart      time.Time `json:"period_start"`
	PeriodEnd        time.Time `json:"period_end"`
	UsedSeconds      int64     `json:"used_seconds"`
	AllowanceSeconds int64     `json:"allowance_seconds"`
}
