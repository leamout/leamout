package auth

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leamout/leamout/server/internal/database/pgconv"
	"github.com/leamout/leamout/server/internal/database/sqlc"
)

func TestIdentityPersistence(t *testing.T) {
	queries := identityDatabase(t)
	ctx := t.Context()
	expiry := time.Now().Add(time.Minute)
	attempt := sqlc.CreateOIDCLoginAttemptParams{
		StateHash: "state", VerifierCiphertext: "encrypted", Nonce: "nonce", ExpiresAt: pgconv.NullableTimestamptz(&expiry),
	}
	if err := queries.CreateOIDCLoginAttempt(ctx, attempt); err != nil {
		t.Fatal(err)
	}
	if _, err := queries.ConsumeOIDCLoginAttempt(ctx, attempt.StateHash); err != nil {
		t.Fatal(err)
	}
	if _, err := queries.ConsumeOIDCLoginAttempt(ctx, attempt.StateHash); err == nil {
		t.Fatal("login state was replayable")
	}
	expired := time.Now().Add(-time.Minute)
	attempt.StateHash = "expired"
	attempt.ExpiresAt = pgconv.NullableTimestamptz(&expired)
	if err := queries.CreateOIDCLoginAttempt(ctx, attempt); err != nil {
		t.Fatal(err)
	}
	if _, err := queries.ConsumeOIDCLoginAttempt(ctx, attempt.StateHash); err == nil {
		t.Fatal("expired state was accepted")
	}
	params := sqlc.UpsertIdentityUserParams{
		IdentityIssuer:  "https://identity.example.com/realms/leamout",
		IdentitySubject: "one",
		Email:           "user@example.com",
	}
	user, err := queries.UpsertIdentityUser(ctx, params)
	if err != nil {
		t.Fatal(err)
	}
	same, err := queries.UpsertIdentityUser(ctx, params)
	if err != nil || same.ID != user.ID {
		t.Fatalf("unstable identity: %v", err)
	}
	params.IdentitySubject = "attacker"
	if _, err := queries.UpsertIdentityUser(ctx, params); err == nil {
		t.Fatal("email collision linked an unrelated identity")
	}
	params.IdentitySubject = "one"
	create := sqlc.CreateSessionParams{
		UserID: user.ID, TokenHash: "one", IdentitySubject: "one", IdentitySessionID: "sso-one",
		TokensCiphertext: "encrypted", Assurance: "unknown", ExpiresAt: pgconv.NullableTimestamptz(&expiry),
	}
	first, err := queries.CreateSession(ctx, create)
	if err != nil {
		t.Fatal(err)
	}
	create.TokenHash = "two"
	create.IdentitySessionID = "sso-two"
	second, err := queries.CreateSession(ctx, create)
	if err != nil {
		t.Fatal(err)
	}
	event := sqlc.ApplyOIDCLogoutParams{
		Issuer: params.IdentityIssuer, Subject: "one", SessionID: "sso-one", TokenID: "event-one", ExpiresAt: pgconv.NullableTimestamptz(&expiry),
	}
	if err := queries.ApplyOIDCLogout(ctx, event); err != nil {
		t.Fatal(err)
	}
	if _, err := queries.GetSessionByID(ctx, first.ID); err == nil {
		t.Fatal("logout did not revoke matching session")
	}
	if _, err := queries.GetSessionByID(ctx, second.ID); err != nil {
		t.Fatal("logout revoked unrelated SSO session")
	}
	create.TokenHash = "three"
	create.IdentitySessionID = "sso-one"
	third, err := queries.CreateSession(ctx, create)
	if err != nil {
		t.Fatal(err)
	}
	if err := queries.ApplyOIDCLogout(ctx, event); err != nil {
		t.Fatal(err)
	}
	if _, err := queries.GetSessionByID(ctx, third.ID); err != nil {
		t.Fatal("replayed logout revoked a new session")
	}
	if err := queries.DisableUser(ctx, user.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := queries.UpsertIdentityUser(ctx, params); err == nil {
		t.Fatal("disabled user was reactivated")
	}
	if _, err := queries.GetSessionByID(ctx, second.ID); err == nil {
		t.Fatal("disabled user retained access")
	}
}

func identityDatabase(t *testing.T) *sqlc.Queries {
	t.Helper()
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set TEST_DATABASE_URL to run identity database acceptance")
	}
	ctx := t.Context()
	admin, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	schema := "identity_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+pgx.Identifier{schema}.Sanitize()); err != nil {
		admin.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, _ = admin.Exec(cleanup, "DROP SCHEMA "+pgx.Identifier{schema}.Sanitize()+" CASCADE")
		admin.Close()
	})
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	config.ConnConfig.RuntimeParams["search_path"] = schema + ",public"
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	migrations, err := filepath.Glob("../../../migrations/*.sql")
	if err != nil || len(migrations) == 0 {
		t.Fatalf("find migrations: %v", err)
	}
	for _, path := range migrations {
		contents, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, string(contents)); err != nil {
			t.Fatalf("migration %s: %v", path, err)
		}
	}
	return sqlc.New(pool)
}
