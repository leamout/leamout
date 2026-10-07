package invitations

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/coffeyvidzro/monogo/internal/database/sqlc"
	"github.com/coffeyvidzro/monogo/internal/platform/email"
	"github.com/coffeyvidzro/monogo/internal/security/encryption"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func invitationDatabase(t *testing.T) (*pgxpool.Pool, *encryption.Cipher) {
	t.Helper()
	databaseURL := os.Getenv("EMAIL_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set EMAIL_TEST_DATABASE_URL for isolated PostgreSQL tests")
	}
	ctx := context.Background()
	admin, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	schema := "invitation_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		admin.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		defer admin.Close()
		if _, err := admin.Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE"); err != nil {
			t.Error(err)
		}
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
	for _, migration := range []string{"001_extensions.sql", "002_create_users.sql", "005_create_organizations.sql", "006_create_organization_members.sql", "007_create_organization_invitations.sql", "032_create_email_deliveries.sql"} {
		body, err := os.ReadFile(filepath.Join("../../../migrations", migration))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, string(body)); err != nil {
			t.Fatalf("%s: %v", migration, err)
		}
	}
	cipher, err := encryption.New(base64.RawURLEncoding.EncodeToString(make([]byte, 32)))
	if err != nil {
		t.Fatal(err)
	}
	return pool, cipher
}

func TestInvitationEmailLifecycle(t *testing.T) {
	pool, cipher := invitationDatabase(t)
	ctx := context.Background()
	q := sqlc.New(pool)
	owner, err := q.CreateUser(ctx, sqlc.CreateUserParams{Email: "owner@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	organization, err := q.CreateOrganizationWithOwner(ctx, sqlc.CreateOrganizationWithOwnerParams{Name: "Acme", UserID: owner.ID})
	if err != nil {
		t.Fatal(err)
	}
	member, err := q.CreateUser(ctx, sqlc.CreateUserParams{Email: "member@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := q.MarkUserEmailVerified(ctx, member.ID); err != nil {
		t.Fatal(err)
	}
	service := NewService(NewRepository(pool), email.NewService(cipher), "https://leamout.com")
	invitation, err := service.Create(ctx, owner.ID, organization.ID, CreateRequest{Email: member.Email})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Create(ctx, owner.ID, organization.ID, CreateRequest{Email: member.Email}); err == nil {
		t.Fatal("duplicate invitation accepted")
	}
	var deliveryID uuid.UUID
	var payload string
	if err := pool.QueryRow(ctx, "SELECT id, encrypted_data FROM email_deliveries WHERE cancellation_key=$1", cancellationKey(invitation.ID)).Scan(&deliveryID, &payload); err != nil {
		t.Fatal(err)
	}
	plaintext, err := cipher.DecryptForScope("email:"+deliveryID.String(), payload)
	if err != nil {
		t.Fatal(err)
	}
	var data email.Data
	if err := json.Unmarshal([]byte(plaintext), &data); err != nil {
		t.Fatal(err)
	}
	link, err := url.Parse(data.AcceptURL)
	if err != nil {
		t.Fatal(err)
	}
	secret := link.Query().Get("token")
	if len(secret) != 64 {
		t.Fatal("missing invitation secret")
	}
	if _, err := service.Accept(ctx, owner.ID, secret); err == nil {
		t.Fatal("wrong recipient accepted invitation")
	}
	accepted, err := service.Accept(ctx, member.ID, secret)
	if err != nil || accepted.Status != "accepted" {
		t.Fatalf("acceptance: %v", err)
	}
	if _, err := service.Accept(ctx, member.ID, secret); err == nil {
		t.Fatal("invitation replay accepted")
	}
	if _, err := q.GetOrganizationMember(ctx, sqlc.GetOrganizationMemberParams{OrganizationID: organization.ID, UserID: member.ID}); err != nil {
		t.Fatal(err)
	}
	var status string
	var erased *string
	if err := pool.QueryRow(ctx, "SELECT status, encrypted_data FROM email_deliveries WHERE id=$1", deliveryID).Scan(&status, &erased); err != nil || status != "cancelled" || erased != nil {
		t.Fatal("invitation email was not cancelled after acceptance")
	}
	var count int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM email_deliveries WHERE template='invitation-accepted' AND recipient=$1", owner.Email).Scan(&count); err != nil || count != 1 {
		t.Fatal("acceptance notification missing or duplicated")
	}

	revoked, err := service.Create(ctx, owner.ID, organization.ID, CreateRequest{Email: "revoked@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if err := service.Revoke(ctx, owner.ID, organization.ID, revoked.ID); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, "SELECT status, encrypted_data FROM email_deliveries WHERE cancellation_key=$1", cancellationKey(revoked.ID)).Scan(&status, &erased); err != nil || status != "cancelled" || erased != nil {
		t.Fatal("revocation did not cancel email")
	}

	broken := NewService(NewRepository(pool), email.NewService(nil), "https://leamout.com")
	if _, err := broken.Create(ctx, owner.ID, organization.ID, CreateRequest{Email: "rollback@example.com"}); err == nil {
		t.Fatal("accepted email persistence failure")
	}
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM organization_invitations WHERE email='rollback@example.com'").Scan(&count); err != nil || count != 0 {
		t.Fatal("invitation committed without email")
	}
}
