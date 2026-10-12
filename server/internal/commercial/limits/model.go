package limits

type Effective struct {
	MaxConcurrentCalls    *int64
	RetentionDays         *int64
	MonthlyRuntimeMinutes *int64
	MaxVoiceAgents        *int64
	MaxSIPTrunks          *int64
}
