package wholesale

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

type pollStoreStub struct {
	cursor       CDRPollCursor
	stored       bool
	storedCount  int
	nextDate     time.Time
	nextPage     int
	nextAttempt  time.Time
	failed       bool
	failureDelay time.Time
}

func (s *pollStoreStub) Cursor(_ context.Context, provider, direction string, startDate time.Time) (CDRPollCursor, error) {
	if s.cursor.Provider == "" {
		s.cursor = CDRPollCursor{Provider: provider, Direction: direction, WindowDate: startDate, Page: 1}
	}
	return s.cursor, nil
}

func (s *pollStoreStub) StorePageAndAdvance(_ context.Context, _ CDRPollCursor, _ json.RawMessage, count int, nextDate time.Time, nextPage int, nextAttempt time.Time) error {
	s.stored = true
	s.storedCount = count
	s.nextDate = nextDate
	s.nextPage = nextPage
	s.nextAttempt = nextAttempt
	return nil
}

func (s *pollStoreStub) Fail(_ context.Context, _ CDRPollCursor, _ error, nextAttempt time.Time) error {
	s.failed = true
	s.failureDelay = nextAttempt
	return nil
}

type pollSourceStub struct {
	count int
	err   error
}

func (s pollSourceStub) PollCDRs(context.Context, string, time.Time, int, int) (json.RawMessage, int, error) {
	if s.err != nil {
		return nil, 0, s.err
	}
	return json.RawMessage(`{"data":[]}`), s.count, nil
}

func testPollConfig() CDRPollJobConfig {
	config := DefaultCDRPollJobConfig("commpeak")
	config.Directions = []string{"termination"}
	config.TickInterval = time.Second
	config.CurrentDayDelay = time.Minute
	config.RetryBase = time.Minute
	config.RetryMax = time.Hour
	config.InitialLookback = 24 * time.Hour
	return config
}

func TestCDRPollJobAdvancesCompletedHistoricalWindow(t *testing.T) {
	now := time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)
	store := &pollStoreStub{cursor: CDRPollCursor{
		Provider: "commpeak", Direction: "termination", WindowDate: now.AddDate(0, 0, -1), Page: 3,
	}}
	job, err := NewCDRPollJob(store, pollSourceStub{count: 12}, testPollConfig())
	if err != nil {
		t.Fatal(err)
	}
	job.now = func() time.Time { return now }
	if err := job.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !store.stored || store.storedCount != 12 || store.nextPage != 1 || !store.nextDate.Equal(utcDate(now)) {
		t.Fatalf("stored=%v count=%d next=%s page=%d", store.stored, store.storedCount, store.nextDate, store.nextPage)
	}
}

func TestCDRPollJobReplaysCurrentDayAfterLastPage(t *testing.T) {
	now := time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)
	store := &pollStoreStub{cursor: CDRPollCursor{
		Provider: "commpeak", Direction: "termination", WindowDate: utcDate(now), Page: 2,
	}}
	job, err := NewCDRPollJob(store, pollSourceStub{count: 4}, testPollConfig())
	if err != nil {
		t.Fatal(err)
	}
	job.now = func() time.Time { return now }
	if err := job.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if store.nextPage != 1 || !store.nextDate.Equal(utcDate(now)) || !store.nextAttempt.Equal(now.Add(time.Minute)) {
		t.Fatalf("next=%s page=%d attempt=%s", store.nextDate, store.nextPage, store.nextAttempt)
	}
}

func TestCDRPollJobRecordsExponentialRetry(t *testing.T) {
	now := time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)
	store := &pollStoreStub{cursor: CDRPollCursor{
		Provider: "commpeak", Direction: "termination", WindowDate: utcDate(now), Page: 1, AttemptCount: 2,
	}}
	job, err := NewCDRPollJob(store, pollSourceStub{err: errors.New("provider unavailable")}, testPollConfig())
	if err != nil {
		t.Fatal(err)
	}
	job.now = func() time.Time { return now }
	if err := job.RunOnce(context.Background()); err == nil {
		t.Fatal("expected polling error")
	}
	if !store.failed || !store.failureDelay.Equal(now.Add(4*time.Minute)) {
		t.Fatalf("failed=%v next=%s", store.failed, store.failureDelay)
	}
}
