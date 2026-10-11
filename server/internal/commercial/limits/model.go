package limits

type Effective struct {
	MaxConcurrentCalls *int64
	RetentionDays      *int64
}
