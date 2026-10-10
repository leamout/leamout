package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leamout/leamout/server/internal/database/sqlc"
	"github.com/leamout/leamout/server/internal/integrations/ses"
	"github.com/leamout/leamout/server/internal/platform/email"
	"github.com/leamout/leamout/server/internal/security/encryption"
)

func testEmailDatabase(t *testing.T) (*pgxpool.Pool, *encryption.Cipher) {
	t.Helper()
	url := os.Getenv("EMAIL_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("set EMAIL_TEST_DATABASE_URL to run isolated PostgreSQL integration tests")
	}
	ctx := context.Background()
	admin, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	schema := "email_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		admin.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		defer admin.Close()
		if _, err := admin.Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE"); err != nil {
			t.Errorf("drop test schema: %v", err)
		}
	})
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema + ",public"
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	for _, name := range []string{"001_extensions.sql", "002_create_users.sql", "003_create_authentication.sql", "032_create_email_deliveries.sql"} {
		body, err := os.ReadFile(filepath.Join("../../../migrations", name))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, string(body)); err != nil {
			t.Fatalf("migration %s: %v", name, err)
		}
	}
	cipher, err := encryption.New(base64.RawURLEncoding.EncodeToString(make([]byte, 32)))
	if err != nil {
		t.Fatal(err)
	}
	return pool, cipher
}
func queuedCode(t *testing.T, pool *pgxpool.Pool, cipher *encryption.Cipher, transactionID uuid.UUID) string {
	t.Helper()
	var id uuid.UUID
	var payload string
	if err := pool.QueryRow(context.Background(), `SELECT id,encrypted_data FROM email_deliveries WHERE cancellation_key=$1 AND status='pending' ORDER BY created_at DESC LIMIT 1`, otpEmailCancellationKey(transactionID)).Scan(&id, &payload); err != nil {
		t.Fatal(err)
	}
	plain, err := cipher.DecryptForScope("email:"+id.String(), payload)
	if err != nil {
		t.Fatal(err)
	}
	var data email.Data
	if err := json.Unmarshal([]byte(plain), &data); err != nil {
		t.Fatal(err)
	}
	return data.Code
}
func TestOTPEmailTransactionAndReplay(t *testing.T) {
	pool, cipher := testEmailDatabase(t)
	ctx := context.Background()
	service := NewService(NewRepository(sqlc.New(pool)))
	service.ConfigureEmail(pool, email.NewService(cipher))
	transaction, err := service.Start(ctx, "new@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if transaction.UserID != nil {
		t.Fatal("expected new user")
	}
	secret, err := service.SendOTP(ctx, transaction.ID)
	if err != nil {
		t.Fatal(err)
	}
	if secret != "" {
		t.Fatal("OTP exposed to caller")
	}
	code := queuedCode(t, pool, cipher, transaction.ID)
	if _, err := service.SendOTP(ctx, transaction.ID); err == nil {
		t.Fatal("resend cooldown bypassed")
	}
	if _, err := service.VerifyOTP(ctx, transaction.ID, "wrong-code"); err == nil {
		t.Fatal("accepted invalid code")
	}
	var attempts int
	if err := pool.QueryRow(ctx, `SELECT attempts FROM auth_challenges WHERE auth_transaction_id=$1`, transaction.ID).Scan(&attempts); err != nil {
		t.Fatal(err)
	}
	if attempts != 1 {
		t.Fatal("failed-code attempt not committed")
	}
	user, err := service.VerifyOTP(ctx, transaction.ID, code)
	if err != nil {
		t.Fatal(err)
	}
	var verified bool
	if err := pool.QueryRow(ctx, `SELECT email_verified FROM users WHERE id=$1`, user.ID).Scan(&verified); err != nil || !verified {
		t.Fatal("user not verified")
	}
	if _, err := service.VerifyOTP(ctx, transaction.ID, code); err == nil {
		t.Fatal("code replay accepted")
	}
	// Verified challenges must never be sent by a delayed worker.
	_, err = email.NewRepository(pool).Claim(ctx, uuid.New())
	if err == nil {
		t.Fatal("claimed consumed OTP")
	}
	var payload *string
	if err := pool.QueryRow(ctx, `SELECT encrypted_data FROM email_deliveries WHERE cancellation_key=$1`, otpEmailCancellationKey(transaction.ID)).Scan(&payload); err != nil || payload != nil {
		t.Fatal("expired payload not scrubbed")
	}
}
func TestOTPResendInvalidatesPreviousCode(t *testing.T) {
	pool, cipher := testEmailDatabase(t)
	ctx := context.Background()
	service := NewService(NewRepository(sqlc.New(pool)))
	service.ConfigureEmail(pool, email.NewService(cipher))
	transaction, err := service.Start(ctx, "resend@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.SendOTP(ctx, transaction.ID); err != nil {
		t.Fatal(err)
	}
	oldCode := queuedCode(t, pool, cipher, transaction.ID)
	// Advance cooldown without waiting; challenges remain unexpired.
	if _, err := pool.Exec(ctx, `UPDATE auth_challenges SET created_at=created_at-interval '61 seconds' WHERE auth_transaction_id=$1`, transaction.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `UPDATE email_deliveries SET created_at=created_at-interval '61 seconds' WHERE cancellation_key=$1`, otpEmailCancellationKey(transaction.ID)); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SendOTP(ctx, transaction.ID); err != nil {
		t.Fatal(err)
	}
	newCode := queuedCode(t, pool, cipher, transaction.ID)
	if oldCode != newCode {
		if _, err := service.VerifyOTP(ctx, transaction.ID, oldCode); err == nil {
			t.Fatal("old code accepted")
		}
	}
	if _, err := service.VerifyOTP(ctx, transaction.ID, newCode); err != nil {
		t.Fatal(err)
	}
}
func TestEmailQueueRollbackAndLeaseOwnership(t *testing.T) {
	pool, cipher := testEmailDatabase(t)
	ctx := context.Background()
	service := email.NewService(cipher)
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.QueueTx(ctx, tx, email.Request{To: "member@example.com", Template: "organization-invitation", Data: email.Data{Organization: "Acme", AcceptURL: "https://example.com/invite", ExpiresAt: time.Now().Add(time.Hour)}})
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM email_deliveries`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("rolled-back email persisted")
	}
	tx, err = pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	id, err := service.QueueTx(ctx, tx, email.Request{To: "member@example.com", Template: "organization-invitation", Data: email.Data{Organization: "Acme", AcceptURL: "https://example.com/invite", ExpiresAt: time.Now().Add(time.Hour)}})
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	repo := email.NewRepository(pool)
	lease := uuid.New()
	delivery, err := repo.Claim(ctx, lease)
	if err != nil || delivery.ID != id {
		t.Fatalf("claim: %v", err)
	}
	if _, err := repo.Claim(ctx, uuid.New()); err == nil {
		t.Fatal("leased email claimed twice")
	}
	if err := repo.Complete(ctx, id, uuid.New(), ses.Result{MessageID: "wrong"}); err != nil {
		t.Fatal(err)
	}
	ready, err := repo.Ready(ctx, id, lease)
	if err != nil || !ready {
		t.Fatal("wrong lease changed delivery")
	}
	if err := repo.Complete(ctx, id, lease, ses.Result{MessageID: "ses-id"}); err != nil {
		t.Fatal(err)
	}
	var status string
	var payload *string
	if err := pool.QueryRow(ctx, `SELECT status,encrypted_data FROM email_deliveries WHERE id=$1`, id).Scan(&status, &payload); err != nil || status != "sent" || payload != nil {
		t.Fatal("completion failed to scrub payload")
	}
}

func TestOTPQueueFailureRollsBackChallenge(t *testing.T) {
	pool, cipher := testEmailDatabase(t)
	ctx := context.Background()
	service := NewService(NewRepository(sqlc.New(pool)))
	service.ConfigureEmail(pool, email.NewService(cipher))
	transaction, err := service.Start(ctx, "rollback@example.com")
	if err != nil {
		t.Fatal(err)
	}
	_, err = pool.Exec(ctx, `CREATE FUNCTION reject_email() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'simulated insert failure'; END $$; CREATE TRIGGER reject_email BEFORE INSERT ON email_deliveries FOR EACH ROW EXECUTE FUNCTION reject_email()`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.SendOTP(ctx, transaction.ID); err == nil {
		t.Fatal("queue failure ignored")
	}
	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM auth_challenges WHERE auth_transaction_id=$1`, transaction.ID).Scan(&count); err != nil || count != 0 {
		t.Fatal("orphaned challenge after queue failure")
	}
}
func TestConcurrentOTPVerificationIsSingleUse(t *testing.T) {
	pool, cipher := testEmailDatabase(t)
	ctx := context.Background()
	service := NewService(NewRepository(sqlc.New(pool)))
	service.ConfigureEmail(pool, email.NewService(cipher))
	transaction, err := service.Start(ctx, "concurrent@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.SendOTP(ctx, transaction.ID); err != nil {
		t.Fatal(err)
	}
	code := queuedCode(t, pool, cipher, transaction.ID)
	var successes atomic.Int32
	var group sync.WaitGroup
	for range 4 {
		group.Go(func() {
			if _, err := service.VerifyOTP(ctx, transaction.ID, code); err == nil {
				successes.Add(1)
			}
		})
	}
	group.Wait()
	if successes.Load() != 1 {
		t.Fatalf("expected one successful verification, got %d", successes.Load())
	}
}

func TestEmailRetryAndTerminalFailure(t *testing.T) {
	pool, cipher := testEmailDatabase(t)
	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	id, err := email.NewService(cipher).QueueTx(ctx, tx, email.Request{To: "retry@example.com", Template: "organization-invitation", Data: email.Data{Organization: "Acme", AcceptURL: "https://example.com/invite", ExpiresAt: time.Now().Add(time.Hour)}})
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	repo := email.NewRepository(pool)
	lease := uuid.New()
	delivery, err := repo.Claim(ctx, lease)
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.Fail(ctx, delivery, lease, "TooManyRequestsException", false); err != nil {
		t.Fatal(err)
	}
	var status string
	var payload *string
	if err := pool.QueryRow(ctx, `SELECT status,encrypted_data FROM email_deliveries WHERE id=$1`, id).Scan(&status, &payload); err != nil || status != "pending" || payload == nil {
		t.Fatal("retry lost payload")
	}
	if _, err := repo.Claim(ctx, uuid.New()); err == nil {
		t.Fatal("retry backoff ignored")
	}
	if _, err := pool.Exec(ctx, `UPDATE email_deliveries SET available_at=now()-interval '1 second' WHERE id=$1`, id); err != nil {
		t.Fatal(err)
	}
	lease = uuid.New()
	delivery, err = repo.Claim(ctx, lease)
	if err != nil || delivery.Attempts != 2 {
		t.Fatal("retry not claimed")
	}
	if err := repo.Fail(ctx, delivery, lease, "MessageRejected", true); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `SELECT status,encrypted_data FROM email_deliveries WHERE id=$1`, id).Scan(&status, &payload); err != nil || status != "failed" || payload != nil {
		t.Fatal("terminal failure retained payload")
	}
}

func TestGenericEmailCancellationIsIsolatedAndTransactional(t *testing.T) {
	pool, cipher := testEmailDatabase(t)
	ctx := context.Background()
	service := email.NewService(cipher)
	// The delivery outbox must work without any authentication tables.
	if _, err := pool.Exec(ctx, `DROP TABLE auth_challenges, auth_transactions, users CASCADE`); err != nil {
		t.Fatal(err)
	}
	keyA, keyB := "invoice:one", "invitation:two"
	queue := func(key *string) uuid.UUID {
		tx, err := pool.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = tx.Rollback(ctx) }()
		id, err := service.QueueTx(ctx, tx, email.Request{To: "generic@example.com", Template: "organization-invitation", CancellationKey: key, Data: email.Data{Organization: "Acme", AcceptURL: "https://example.com/invite", ExpiresAt: time.Now().Add(time.Hour)}})
		if err != nil {
			t.Fatal(err)
		}
		if err := tx.Commit(ctx); err != nil {
			t.Fatal(err)
		}
		return id
	}
	idA := queue(&keyA)
	idB := queue(&keyB)
	idC := queue(nil)
	repo := email.NewRepository(pool)
	lease := uuid.New()
	delivery, err := repo.Claim(ctx, lease)
	if err != nil || delivery.ID != idA {
		t.Fatalf("claim: %v", err)
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.CancelTx(ctx, tx, keyA); err != nil {
		t.Fatal(err)
	}
	if err := tx.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	ready, err := repo.Ready(ctx, idA, lease)
	if err != nil || !ready {
		t.Fatal("rolled-back cancellation persisted")
	}
	tx, err = pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.CancelTx(ctx, tx, keyA); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	ready, err = repo.Ready(ctx, idA, lease)
	if err != nil || ready {
		t.Fatal("cancelled delivery remains ready")
	}
	// A provider response arriving after cancellation cannot restore the delivery.
	if err := repo.Complete(ctx, idA, lease, ses.Result{MessageID: "late-response"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.Fail(ctx, delivery, lease, "late-failure", false); err != nil {
		t.Fatal(err)
	}
	var status string
	var payload *string
	if err := pool.QueryRow(ctx, `SELECT status,encrypted_data FROM email_deliveries WHERE id=$1`, idA).Scan(&status, &payload); err != nil || status != "cancelled" || payload != nil {
		t.Fatal("cancelled job revived or retained secrets")
	}
	for _, id := range []uuid.UUID{idB, idC} {
		if err := pool.QueryRow(ctx, `SELECT status FROM email_deliveries WHERE id=$1`, id).Scan(&status); err != nil || status != "pending" {
			t.Fatal("unrelated job cancelled")
		}
	}
}

func TestExhaustedOTPAttemptsCancelQueuedDelivery(t *testing.T) {
	pool, cipher := testEmailDatabase(t)
	ctx := context.Background()
	service := NewService(NewRepository(sqlc.New(pool)))
	service.ConfigureEmail(pool, email.NewService(cipher))
	transaction, err := service.Start(ctx, "exhausted@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.SendOTP(ctx, transaction.ID); err != nil {
		t.Fatal(err)
	}
	for range 5 {
		if _, err := service.VerifyOTP(ctx, transaction.ID, "invalid"); err == nil {
			t.Fatal("invalid code accepted")
		}
	}
	var status string
	var payload *string
	if err := pool.QueryRow(ctx, `SELECT status,encrypted_data FROM email_deliveries WHERE cancellation_key=$1`, otpEmailCancellationKey(transaction.ID)).Scan(&status, &payload); err != nil || status != "cancelled" || payload != nil {
		t.Fatal("exhausted challenge delivery not cancelled")
	}
}

func TestPasswordLoginCancelsQueuedOTP(t *testing.T) {
	pool, cipher := testEmailDatabase(t)
	ctx := context.Background()
	queries := sqlc.New(pool)
	user, err := queries.CreateUser(ctx, sqlc.CreateUserParams{Email: "password@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	service := NewService(NewRepository(queries))
	service.ConfigureEmail(pool, email.NewService(cipher))
	if _, err := service.SetPassword(ctx, user.ID, "strong-test-password-123"); err != nil {
		t.Fatal(err)
	}
	transaction, err := service.Start(ctx, user.Email)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.SendOTP(ctx, transaction.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.LoginWithPassword(ctx, transaction.ID, "strong-test-password-123"); err != nil {
		t.Fatal(err)
	}
	var status string
	var payload *string
	if err := pool.QueryRow(ctx, `SELECT status,encrypted_data FROM email_deliveries WHERE cancellation_key=$1`, otpEmailCancellationKey(transaction.ID)).Scan(&status, &payload); err != nil || status != "cancelled" || payload != nil {
		t.Fatal("password login left OTP queued")
	}
}
