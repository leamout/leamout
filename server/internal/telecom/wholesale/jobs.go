package wholesale

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

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
