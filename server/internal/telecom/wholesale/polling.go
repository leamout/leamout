package wholesale

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CDRPageSource interface {
	PollCDRs(context.Context, string, time.Time, int, int) (json.RawMessage, int, error)
}

type CDRPollCursor struct {
	Provider      string
	Direction     string
	WindowDate    time.Time
	Page          int
	AttemptCount  int
	NextAttemptAt time.Time
}

type CDRPollStore interface {
	Cursor(context.Context, string, string, time.Time) (CDRPollCursor, error)
	StorePageAndAdvance(context.Context, CDRPollCursor, json.RawMessage, int, time.Time, int, time.Time) error
	Fail(context.Context, CDRPollCursor, error, time.Time) error
}

type CDRPollRepository struct{ db *pgxpool.Pool }

func NewCDRPollRepository(db *pgxpool.Pool) *CDRPollRepository { return &CDRPollRepository{db: db} }

func (r *CDRPollRepository) Cursor(ctx context.Context, provider, direction string, startDate time.Time) (CDRPollCursor, error) {
	if r == nil || r.db == nil {
		return CDRPollCursor{}, errors.New("provider CDR polling repository unavailable")
	}
	startDate = utcDate(startDate)
	if _, err := r.db.Exec(ctx, `
		INSERT INTO provider_cdr_poll_cursors (provider, direction, window_date, page, next_attempt_at)
		VALUES ($1, $2, $3, 1, now())
		ON CONFLICT (provider, direction) DO NOTHING
	`, provider, direction, startDate); err != nil {
		return CDRPollCursor{}, err
	}
	var cursor CDRPollCursor
	if err := r.db.QueryRow(ctx, `
		SELECT provider, direction, window_date, page, attempt_count, next_attempt_at
		FROM provider_cdr_poll_cursors
		WHERE provider = $1 AND direction = $2
	`, provider, direction).Scan(
		&cursor.Provider,
		&cursor.Direction,
		&cursor.WindowDate,
		&cursor.Page,
		&cursor.AttemptCount,
		&cursor.NextAttemptAt,
	); err != nil {
		return CDRPollCursor{}, err
	}
	return cursor, nil
}

func (r *CDRPollRepository) StorePageAndAdvance(ctx context.Context, cursor CDRPollCursor, raw json.RawMessage, count int, nextDate time.Time, nextPage int, nextAttemptAt time.Time) error {
	if r == nil || r.db == nil {
		return errors.New("provider CDR polling repository unavailable")
	}
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	hash := sha256.Sum256(raw)
	if _, err := tx.Exec(ctx, `
		INSERT INTO provider_cdr_pages (
			provider, direction, window_date, page, record_count, payload_sha256, raw
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (provider, direction, window_date, page, payload_sha256) DO NOTHING
	`, cursor.Provider, cursor.Direction, utcDate(cursor.WindowDate), cursor.Page, count, hex.EncodeToString(hash[:]), raw); err != nil {
		return err
	}
	command, err := tx.Exec(ctx, `
		UPDATE provider_cdr_poll_cursors
		SET window_date = $3,
			page = $4,
			attempt_count = 0,
			next_attempt_at = $5,
			last_error = NULL,
			last_success_at = now()
		WHERE provider = $1 AND direction = $2
	`, cursor.Provider, cursor.Direction, utcDate(nextDate), nextPage, nextAttemptAt)
	if err != nil {
		return err
	}
	if command.RowsAffected() != 1 {
		return errors.New("provider CDR polling cursor disappeared")
	}
	return tx.Commit(ctx)
}

func (r *CDRPollRepository) Fail(ctx context.Context, cursor CDRPollCursor, pollErr error, nextAttemptAt time.Time) error {
	if r == nil || r.db == nil {
		return errors.New("provider CDR polling repository unavailable")
	}
	message := strings.TrimSpace(pollErr.Error())
	if len(message) > 2048 {
		message = message[:2048]
	}
	command, err := r.db.Exec(ctx, `
		UPDATE provider_cdr_poll_cursors
		SET attempt_count = attempt_count + 1,
			next_attempt_at = $3,
			last_error = $4
		WHERE provider = $1 AND direction = $2
	`, cursor.Provider, cursor.Direction, nextAttemptAt, message)
	if err != nil {
		return err
	}
	if command.RowsAffected() != 1 {
		return errors.New("provider CDR polling cursor disappeared")
	}
	return nil
}

type CDRPollJobConfig struct {
	Provider        string
	Directions      []string
	PerPage         int
	TickInterval    time.Duration
	CurrentDayDelay time.Duration
	RetryBase       time.Duration
	RetryMax        time.Duration
	InitialLookback time.Duration
}

func DefaultCDRPollJobConfig(provider string) CDRPollJobConfig {
	return CDRPollJobConfig{
		Provider:        provider,
		Directions:      []string{"termination", "origination"},
		PerPage:         1000,
		TickInterval:    10 * time.Second,
		CurrentDayDelay: time.Minute,
		RetryBase:       time.Minute,
		RetryMax:        time.Hour,
		InitialLookback: 24 * time.Hour,
	}
}

type CDRPollJob struct {
	store  CDRPollStore
	source CDRPageSource
	config CDRPollJobConfig
	now    func() time.Time
}

func NewCDRPollJob(store CDRPollStore, source CDRPageSource, config CDRPollJobConfig) (*CDRPollJob, error) {
	config.Provider = strings.ToLower(strings.TrimSpace(config.Provider))
	if config.Provider == "" {
		return nil, errors.New("provider CDR polling provider is required")
	}
	if store == nil {
		return nil, errors.New("provider CDR polling store is required")
	}
	if config.PerPage <= 0 || config.PerPage > 1000 || config.TickInterval <= 0 || config.CurrentDayDelay <= 0 ||
		config.RetryBase <= 0 || config.RetryMax < config.RetryBase || config.InitialLookback < 0 {
		return nil, errors.New("invalid provider CDR polling configuration")
	}
	return &CDRPollJob{store: store, source: source, config: config, now: time.Now}, nil
}

func (j *CDRPollJob) Run(ctx context.Context) error {
	if j.source == nil {
		<-ctx.Done()
		return nil
	}
	ticker := time.NewTicker(j.config.TickInterval)
	defer ticker.Stop()
	for {
		_ = j.RunOnce(ctx)
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func (j *CDRPollJob) RunOnce(ctx context.Context) error {
	if j.source == nil {
		return nil
	}
	now := j.now().UTC()
	startDate := utcDate(now.Add(-j.config.InitialLookback))
	var errs []error
	for _, direction := range j.config.Directions {
		direction = strings.ToLower(strings.TrimSpace(direction))
		cursor, err := j.store.Cursor(ctx, j.config.Provider, direction, startDate)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s cursor: %w", direction, err))
			continue
		}
		if cursor.NextAttemptAt.After(now) {
			continue
		}
		raw, count, err := j.source.PollCDRs(ctx, direction, cursor.WindowDate, cursor.Page, j.config.PerPage)
		if err != nil {
			next := now.Add(j.retryDelay(cursor.AttemptCount))
			if storeErr := j.store.Fail(ctx, cursor, err, next); storeErr != nil {
				errs = append(errs, fmt.Errorf("%s record failure: %w", direction, storeErr))
			}
			errs = append(errs, fmt.Errorf("%s poll: %w", direction, err))
			continue
		}
		if !json.Valid(raw) || count < 0 || count > j.config.PerPage {
			err = errors.New("provider CDR poll returned invalid page")
			next := now.Add(j.retryDelay(cursor.AttemptCount))
			_ = j.store.Fail(ctx, cursor, err, next)
			errs = append(errs, fmt.Errorf("%s poll: %w", direction, err))
			continue
		}

		today := utcDate(now)
		nextDate := utcDate(cursor.WindowDate)
		nextPage := cursor.Page
		nextAttemptAt := now
		if count >= j.config.PerPage {
			nextPage++
		} else if nextDate.Before(today) {
			nextDate = nextDate.AddDate(0, 0, 1)
			nextPage = 1
		} else {
			nextPage = 1
			nextAttemptAt = now.Add(j.config.CurrentDayDelay)
		}
		if err := j.store.StorePageAndAdvance(ctx, cursor, raw, count, nextDate, nextPage, nextAttemptAt); err != nil {
			errs = append(errs, fmt.Errorf("%s store page: %w", direction, err))
		}
	}
	return errors.Join(errs...)
}

func (j *CDRPollJob) retryDelay(attemptCount int) time.Duration {
	exponent := math.Min(float64(attemptCount), 10)
	delay := time.Duration(float64(j.config.RetryBase) * math.Pow(2, exponent))
	if delay > j.config.RetryMax {
		return j.config.RetryMax
	}
	return delay
}

func utcDate(value time.Time) time.Time {
	year, month, day := value.UTC().Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
