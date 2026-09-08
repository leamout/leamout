package provider_diagnostics

import "time"

type OperationSummary struct {
	PendingCount    int64 `json:"pending_count"`
	AcceptedCount   int64 `json:"accepted_count"`
	FailedCount     int64 `json:"failed_count"`
	RetryReadyCount int64 `json:"retry_ready_count"`
}

type OperationDiagnostic struct {
	ID            string     `json:"id"`
	Provider      string     `json:"provider"`
	OperationType string     `json:"operation_type"`
	State         string     `json:"state"`
	Attempts      int32      `json:"attempts"`
	LastError     *string    `json:"last_error,omitempty"`
	NextAttemptAt *time.Time `json:"next_attempt_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type CDRPollDiagnostic struct {
	Provider           string     `json:"provider"`
	Direction          string     `json:"direction"`
	WindowDate         time.Time  `json:"window_date"`
	Page               int32      `json:"page"`
	AttemptCount       int32      `json:"attempt_count"`
	NextAttemptAt      time.Time  `json:"next_attempt_at"`
	LastError          *string    `json:"last_error,omitempty"`
	LastSuccessAt      *time.Time `json:"last_success_at,omitempty"`
	LastPageReceivedAt *time.Time `json:"last_page_received_at,omitempty"`
}

type Snapshot struct {
	Operations  OperationSummary      `json:"operations"`
	Problematic []OperationDiagnostic `json:"problematic_operations"`
	CDRPolling  []CDRPollDiagnostic   `json:"cdr_polling"`
}
