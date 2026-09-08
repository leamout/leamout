package provisioning

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/leamout/leamout/internal/platform/config"
)

const (
	defaultProviderOperationDiagnosticLimit = 50
	maxProviderOperationDiagnosticLimit     = 500
)

type ProviderOperationDiagnostic struct {
	ID                     uuid.UUID  `json:"id"`
	OrganizationID         uuid.UUID  `json:"organization_id"`
	CarrierProviderID      uuid.UUID  `json:"carrier_provider_id"`
	Provider               string     `json:"provider"`
	PhoneNumberID          uuid.UUID  `json:"phone_number_id"`
	Number                 string     `json:"number"`
	OperationType          string     `json:"operation_type"`
	State                  string     `json:"state"`
	ProviderOperationID    *string    `json:"provider_operation_id,omitempty"`
	ProviderResourceID     *string    `json:"provider_resource_id,omitempty"`
	Attempts               int32      `json:"attempts"`
	LastError              *string    `json:"last_error,omitempty"`
	NextAttemptAt          *time.Time `json:"next_attempt_at,omitempty"`
	CompletedAt            *time.Time `json:"completed_at,omitempty"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
	RequiresOperatorAction bool       `json:"requires_operator_action"`
}

const providerOperationDiagnosticColumns = `
	po.id,
	po.organization_id,
	po.carrier_provider_id,
	cp.slug,
	po.phone_number_id,
	pn.number,
	po.operation_type,
	po.state,
	po.provider_operation_id,
	po.provider_resource_id,
	po.attempts,
	po.last_error,
	po.next_attempt_at,
	po.completed_at,
	po.created_at,
	po.updated_at`

func ListProviderOperationDiagnostics(ctx context.Context, cfg config.Config, state string, limit int) ([]ProviderOperationDiagnostic, error) {
	state, err := normalizeProviderOperationState(state, true)
	if err != nil {
		return nil, err
	}
	limit, err = normalizeProviderOperationLimit(limit)
	if err != nil {
		return nil, err
	}

	db, err := openProviderOperationDB(ctx, cfg)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(ctx, `SELECT `+providerOperationDiagnosticColumns+`
FROM provider_operations AS po
JOIN carrier_providers AS cp ON cp.id = po.carrier_provider_id
JOIN phone_numbers AS pn ON pn.id = po.phone_number_id
WHERE ($1 = '' OR po.state = $1)
ORDER BY po.updated_at DESC, po.created_at DESC
LIMIT $2`, state, int32(limit))
	if err != nil {
		return nil, fmt.Errorf("list provider operation diagnostics: %w", err)
	}
	defer rows.Close()

	result := make([]ProviderOperationDiagnostic, 0)
	for rows.Next() {
		diagnostic, err := scanProviderOperationDiagnostic(rows)
		if err != nil {
			return nil, fmt.Errorf("scan provider operation diagnostic: %w", err)
		}
		result = append(result, diagnostic)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate provider operation diagnostics: %w", err)
	}
	return result, nil
}

func GetProviderOperationDiagnostic(ctx context.Context, cfg config.Config, id uuid.UUID) (ProviderOperationDiagnostic, error) {
	if id == uuid.Nil {
		return ProviderOperationDiagnostic{}, fmt.Errorf("provider operation id is required")
	}

	db, err := openProviderOperationDB(ctx, cfg)
	if err != nil {
		return ProviderOperationDiagnostic{}, err
	}
	defer db.Close()

	diagnostic, err := scanProviderOperationDiagnostic(db.QueryRow(ctx, `SELECT `+providerOperationDiagnosticColumns+`
FROM provider_operations AS po
JOIN carrier_providers AS cp ON cp.id = po.carrier_provider_id
JOIN phone_numbers AS pn ON pn.id = po.phone_number_id
WHERE po.id = $1`, id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return ProviderOperationDiagnostic{}, fmt.Errorf("provider operation %s not found", id)
		}
		return ProviderOperationDiagnostic{}, fmt.Errorf("get provider operation diagnostic: %w", err)
	}
	return diagnostic, nil
}

// ScheduleProviderOperationReconciliation makes an unfinished operation eligible
// for the existing worker reconciler immediately. It never changes terminal
// state and does not execute provider side effects in the operator process.
func ScheduleProviderOperationReconciliation(ctx context.Context, cfg config.Config, id uuid.UUID) (ProviderOperationDiagnostic, error) {
	if id == uuid.Nil {
		return ProviderOperationDiagnostic{}, fmt.Errorf("provider operation id is required")
	}

	db, err := openProviderOperationDB(ctx, cfg)
	if err != nil {
		return ProviderOperationDiagnostic{}, err
	}
	defer db.Close()

	diagnostic, err := scanProviderOperationDiagnostic(db.QueryRow(ctx, `WITH scheduled AS (
	UPDATE provider_operations
	SET next_attempt_at = now()
	WHERE id = $1
	  AND state IN ('pending', 'provider_accepted')
	RETURNING id
)
SELECT `+providerOperationDiagnosticColumns+`
FROM provider_operations AS po
JOIN scheduled AS s ON s.id = po.id
JOIN carrier_providers AS cp ON cp.id = po.carrier_provider_id
JOIN phone_numbers AS pn ON pn.id = po.phone_number_id`, id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return ProviderOperationDiagnostic{}, fmt.Errorf("provider operation %s is not pending reconciliation", id)
		}
		return ProviderOperationDiagnostic{}, fmt.Errorf("schedule provider operation reconciliation: %w", err)
	}
	return diagnostic, nil
}

func openProviderOperationDB(ctx context.Context, cfg config.Config) (*pgxpool.Pool, error) {
	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect provider operation diagnostics database: %w", err)
	}
	if err := db.Ping(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping provider operation diagnostics database: %w", err)
	}
	return db, nil
}

type providerOperationScanner interface {
	Scan(dest ...any) error
}

func scanProviderOperationDiagnostic(row providerOperationScanner) (ProviderOperationDiagnostic, error) {
	var diagnostic ProviderOperationDiagnostic
	var nextAttemptAt pgtype.Timestamptz
	var completedAt pgtype.Timestamptz
	if err := row.Scan(
		&diagnostic.ID,
		&diagnostic.OrganizationID,
		&diagnostic.CarrierProviderID,
		&diagnostic.Provider,
		&diagnostic.PhoneNumberID,
		&diagnostic.Number,
		&diagnostic.OperationType,
		&diagnostic.State,
		&diagnostic.ProviderOperationID,
		&diagnostic.ProviderResourceID,
		&diagnostic.Attempts,
		&diagnostic.LastError,
		&nextAttemptAt,
		&completedAt,
		&diagnostic.CreatedAt,
		&diagnostic.UpdatedAt,
	); err != nil {
		return ProviderOperationDiagnostic{}, err
	}
	if nextAttemptAt.Valid {
		value := nextAttemptAt.Time
		diagnostic.NextAttemptAt = &value
	}
	if completedAt.Valid {
		value := completedAt.Time
		diagnostic.CompletedAt = &value
	}
	diagnostic.RequiresOperatorAction = diagnostic.State == "failed"
	return diagnostic, nil
}

func normalizeProviderOperationState(state string, allowEmpty bool) (string, error) {
	state = strings.ToLower(strings.TrimSpace(state))
	if state == "" && allowEmpty {
		return "", nil
	}
	switch state {
	case "pending", "provider_accepted", "succeeded", "failed":
		return state, nil
	default:
		return "", fmt.Errorf("provider operation state must be pending, provider_accepted, succeeded, or failed")
	}
}

func normalizeProviderOperationLimit(limit int) (int, error) {
	if limit == 0 {
		return defaultProviderOperationDiagnosticLimit, nil
	}
	if limit < 0 || limit > maxProviderOperationDiagnosticLimit {
		return 0, fmt.Errorf("provider operation diagnostic limit must be between 1 and %d", maxProviderOperationDiagnosticLimit)
	}
	return limit, nil
}
