package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/google/uuid"
	"github.com/leamout/leamout/server/internal/database/pgconv"
	"github.com/leamout/leamout/server/internal/database/sqlc"
	"github.com/leamout/leamout/server/internal/integrations/ses"
	"github.com/leamout/leamout/server/internal/platform/email"
	"github.com/leamout/leamout/server/internal/security/password"
	"github.com/leamout/leamout/server/internal/security/token"
	"github.com/leamout/leamout/server/internal/tenancy/invitations"
)

func TestNotificationDeliveryFlow(t *testing.T) {
	pool, cipher := testEmailDatabase(t)
	ctx := context.Background()
	queries := sqlc.New(pool)
	messages := make(chan sesv2.SendEmailInput, 8)
	transportErrors := make(chan error, 8)
	endpoint := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request sesv2.SendEmailInput
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			transportErrors <- err
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		messages <- request
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{"MessageId": "delivery-id"}); err != nil {
			transportErrors <- err
		}
	}))
	defer endpoint.Close()
	client := sesv2.New(sesv2.Options{Region: "us-east-1", BaseEndpoint: aws.String(endpoint.URL), Credentials: credentials.NewStaticCredentialsProvider("test", "test", ""), RetryMaxAttempts: 1})
	sender := ses.NewSender(client, ses.Config{From: "noreply@example.com", ConfigurationSet: "transactional"})
	worker := email.NewWorker(email.NewRepository(pool), sender, cipher)
	checkDelivery := func(expectedSubject, expectedRecipient string) {
		t.Helper()
		if err := worker.ProcessOne(ctx); err != nil {
			t.Fatal(err)
		}
		select {
		case err := <-transportErrors:
			t.Fatal(err)
		default:
		}
		select {
		case message := <-messages:
			if len(message.Destination.ToAddresses) != 1 || message.Destination.ToAddresses[0] != expectedRecipient || aws.ToString(message.Content.Simple.Subject.Data) != expectedSubject || aws.ToString(message.Content.Simple.Body.Html.Data) == "" || aws.ToString(message.Content.Simple.Body.Text.Data) == "" || aws.ToString(message.ConfigurationSetName) != "transactional" {
				t.Fatal("incorrect SES delivery")
			}
		default:
			t.Fatal("worker did not send an email")
		}
	}
	user, err := queries.CreateUser(ctx, sqlc.CreateUserParams{Email: "recover@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	originalHash, err := password.Hash("OldPassword123!")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := queries.SetUserPassword(ctx, sqlc.SetUserPasswordParams{ID: user.ID, PasswordHash: &originalHash}); err != nil {
		t.Fatal(err)
	}
	existingSession, err := queries.CreateSession(ctx, sqlc.CreateSessionParams{UserID: user.ID, TokenHash: token.Hash("session"), Assurance: "password", ExpiresAt: pgconv.TimeToTimestamptz(time.Now().Add(time.Hour))})
	if err != nil {
		t.Fatal(err)
	}
	service := NewService(NewRepository(queries))
	service.ConfigureEmail(pool, email.NewService(cipher))
	service.ConfigureRecoveryURL("https://console.example.com/forgot-password")
	unknown, err := service.StartRecovery(ctx, "unknown@example.com")
	if err != nil || unknown.ID == uuid.Nil {
		t.Fatalf("recovery must not enumerate accounts: %v", err)
	}
	if _, err := queries.GetActiveAuthChallenge(ctx, sqlc.GetActiveAuthChallengeParams{AuthTransactionID: &unknown.ID, Purpose: "password_reset"}); err == nil {
		t.Fatal("unknown account received a recovery challenge")
	}
	recovery, err := service.StartRecovery(ctx, user.Email)
	if err != nil {
		t.Fatal(err)
	}
	code := queuedCode(t, pool, cipher, recovery.ID)
	if _, err := service.VerifyOTP(ctx, recovery.ID, code); err == nil {
		t.Fatal("recovery code allowed a login session")
	}
	if err := service.ResetPassword(ctx, recovery.ID, "9999999", "NewPassword123!"); err == nil {
		t.Fatal("accepted wrong recovery code")
	}
	challenge, err := queries.GetActiveAuthChallenge(ctx, sqlc.GetActiveAuthChallengeParams{AuthTransactionID: &recovery.ID, Purpose: "password_reset"})
	if err != nil || challenge.Attempts != 1 {
		t.Fatalf("failed attempt was not recorded: %v", err)
	}
	checkDelivery("Your Leamout verification code", user.Email)
	if err := service.ResetPassword(ctx, recovery.ID, code, "NewPassword123!"); err != nil {
		t.Fatal(err)
	}
	if err := service.ResetPassword(ctx, recovery.ID, code, "AnotherPassword123!"); err == nil {
		t.Fatal("replayed recovery code")
	}
	updated, err := queries.GetUserByID(ctx, user.ID)
	if err != nil || updated.PasswordHash == nil || !password.Verify("NewPassword123!", *updated.PasswordHash) {
		t.Fatal("password was not replaced")
	}
	if _, err := queries.GetSessionByID(ctx, existingSession.ID); err == nil {
		t.Fatal("old session survived password recovery")
	}
	checkDelivery("Your Leamout password was changed", user.Email)
	organization, err := queries.CreateOrganizationWithOwner(ctx, sqlc.CreateOrganizationWithOwnerParams{Name: "Acme", UserID: user.ID})
	if err != nil {
		t.Fatal(err)
	}
	invitee, err := queries.CreateUser(ctx, sqlc.CreateUserParams{Email: "invitee@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	inviteService := invitations.NewService(invitations.NewRepository(pool), email.NewService(cipher), "https://console.example.com/invitations/accept")
	if _, err := inviteService.Create(ctx, invitee.ID, organization.ID, invitations.CreateRequest{Email: "unauthorized@example.com"}); err == nil {
		t.Fatal("nonmember sent an invitation")
	}
	invitation, err := inviteService.Create(ctx, user.ID, organization.ID, invitations.CreateRequest{Email: invitee.Email, Role: "member"})
	if err != nil {
		t.Fatal(err)
	}
	var deliveryID uuid.UUID
	var payload string
	if err := pool.QueryRow(ctx, "SELECT id, encrypted_data FROM email_deliveries WHERE cancellation_key=$1", "invitation:"+invitation.ID.String()).Scan(&deliveryID, &payload); err != nil {
		t.Fatal(err)
	}
	decoded, err := cipher.DecryptForScope("email:"+deliveryID.String(), payload)
	if err != nil {
		t.Fatal(err)
	}
	var data email.Data
	if err := json.Unmarshal([]byte(decoded), &data); err != nil {
		t.Fatal(err)
	}
	acceptURL, err := url.Parse(data.AcceptURL)
	if err != nil {
		t.Fatal(err)
	}
	invitationToken := acceptURL.Query().Get("token")
	if strings.Contains(payload, invitationToken) {
		t.Fatal("invitation token stored in plaintext")
	}
	checkDelivery("You’re invited to join Acme on Leamout", invitee.Email)
	if err := inviteService.Accept(ctx, user.ID, invitationToken); err == nil {
		t.Fatal("wrong user accepted invitation")
	}
	if err := inviteService.Accept(ctx, invitee.ID, invitationToken); err == nil {
		t.Fatal("unverified email accepted invitation")
	}
	if _, err := queries.MarkUserEmailVerified(ctx, invitee.ID); err != nil {
		t.Fatal(err)
	}
	if err := inviteService.Accept(ctx, invitee.ID, invitationToken); err != nil {
		t.Fatal(err)
	}
	if _, err := queries.GetOrganizationMember(ctx, sqlc.GetOrganizationMemberParams{OrganizationID: organization.ID, UserID: invitee.ID}); err != nil {
		t.Fatal(err)
	}
	if err := inviteService.Accept(ctx, invitee.ID, invitationToken); err == nil {
		t.Fatal("replayed invitation")
	}
	revoked, err := inviteService.Create(ctx, user.ID, organization.ID, invitations.CreateRequest{Email: "revoked@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if err := inviteService.Revoke(ctx, user.ID, organization.ID, revoked.ID); err != nil {
		t.Fatal(err)
	}
	var state string
	if err := pool.QueryRow(ctx, "SELECT status FROM email_deliveries WHERE cancellation_key=$1", "invitation:"+revoked.ID.String()).Scan(&state); err != nil || state != "cancelled" {
		t.Fatal("revoked invitation email was not cancelled")
	}
}
