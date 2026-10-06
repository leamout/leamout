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

	"github.com/coffeyvidzro/monogo/internal/database/sqlc"
	"github.com/coffeyvidzro/monogo/internal/platform/email"
	"github.com/coffeyvidzro/monogo/internal/security/encryption"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
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
	if err := pool.QueryRow(context.Background(), `SELECT d.id,d.encrypted_data FROM email_deliveries d JOIN auth_challenges c ON c.id=d.challenge_id WHERE c.auth_transaction_id=$1 AND d.status='pending' ORDER BY d.created_at DESC LIMIT 1`, transactionID).Scan(&id, &payload); err != nil {
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
	if err := pool.QueryRow(ctx, `SELECT encrypted_data FROM email_deliveries WHERE challenge_id IN (SELECT id FROM auth_challenges WHERE auth_transaction_id=$1)`, transaction.ID).Scan(&payload); err != nil || payload != nil {
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
	if _, err := pool.Exec(ctx, `UPDATE email_deliveries SET created_at=created_at-interval '61 seconds' WHERE challenge_id IN (SELECT id FROM auth_challenges WHERE auth_transaction_id=$1)`, transaction.ID); err != nil {
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
	_, err = service.QueueTx(ctx, tx, email.Request{To: "member@example.com", Template: "invitation", Data: email.Data{Organization: "Acme", AcceptURL: "https://example.com/invite", ExpiresAt: time.Now().Add(time.Hour)}})
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
	id, err := service.QueueTx(ctx, tx, email.Request{To: "member@example.com", Template: "invitation", Data: email.Data{Organization: "Acme", AcceptURL: "https://example.com/invite", ExpiresAt: time.Now().Add(time.Hour)}})
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
	if err := repo.Complete(ctx, id, uuid.New(), email.Result{MessageID: "wrong"}); err != nil {
		t.Fatal(err)
	}
	ready, err := repo.Ready(ctx, id, lease)
	if err != nil || !ready {
		t.Fatal("wrong lease changed delivery")
	}
	if err := repo.Complete(ctx, id, lease, email.Result{MessageID: "ses-id"}); err != nil {
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
	id, err := email.NewService(cipher).QueueTx(ctx, tx, email.Request{To: "retry@example.com", Template: "invitation", Data: email.Data{Organization: "Acme", AcceptURL: "https://example.com/invite", ExpiresAt: time.Now().Add(time.Hour)}})
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
