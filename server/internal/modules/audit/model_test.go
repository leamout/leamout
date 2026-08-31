package audit

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/security/authn"
)

func TestActorFromContextSupportsUserAndOrganizationToken(t *testing.T) {
	for _, subjectType := range []authn.SubjectType{authn.SubjectUser, authn.SubjectOrganizationToken} {
		t.Run(string(subjectType), func(t *testing.T) {
			id := uuid.New()
			ctx := authn.WithPrincipal(context.Background(), authn.Principal{
				Subject:    authn.Subject{ID: id, Type: subjectType},
				Credential: authn.Credential{Type: authn.CredentialSession},
			})
			actor, err := ActorFromContext(ctx)
			if err != nil || actor.ID != id || actor.Type != string(subjectType) {
				t.Fatalf("actor = %+v, err = %v", actor, err)
			}
		})
	}
}

func TestNewEventContainsOnlyExplicitRedactedMetadata(t *testing.T) {
	secret := "do-not-persist-this-secret"
	event, err := NewEvent(uuid.New(), Actor{Type: "user", ID: uuid.New()}, "carrier.credential_rotated", "carrier_connection", uuid.New(), map[string]any{
		"direction":  "outbound",
		"credential": "[REDACTED]",
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(event.Metadata), secret) || !strings.Contains(string(event.Metadata), "[REDACTED]") {
		t.Fatalf("unsafe metadata: %s", event.Metadata)
	}
}
