package authn

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

type stubSessionResolver struct {
	session Session
}

func (s stubSessionResolver) ResolveSession(context.Context, string) (Session, error) {
	return s.session, nil
}

func TestResolveSessionPreservesAssurance(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()
	resolver := NewResolver(stubSessionResolver{session: Session{
		ID:        sessionID,
		UserID:    userID,
		Assurance: AssurancePassword,
	}}, nil)

	principal, err := resolver.Resolve(t.Context(), CredentialInput{
		Type:  CredentialSession,
		Value: "session-token",
	})
	if err != nil {
		t.Fatalf("resolve session: %v", err)
	}
	if principal.Subject.ID != userID {
		t.Fatalf("subject ID = %s, want %s", principal.Subject.ID, userID)
	}
	if principal.Credential.ID != sessionID {
		t.Fatalf("credential ID = %s, want %s", principal.Credential.ID, sessionID)
	}
	if principal.Assurance != AssurancePassword {
		t.Fatalf("assurance = %v, want %v", principal.Assurance, AssurancePassword)
	}
}
