package nats

const (
	JobsStreamName   = "LEAMOUT_JOBS"
	EventsStreamName = "LEAMOUT_EVENTS"
	DLQStreamName    = "LEAMOUT_DLQ"

	JobsSubject   = "leamout.job.>"
	EventsSubject = "leamout.event.>"
	DLQSubject    = "leamout.dlq.>"
)
