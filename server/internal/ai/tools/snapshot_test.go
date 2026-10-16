package tools

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/leamout/leamout/server/internal/database/sqlc"
)

func snapshotFixture(t *testing.T) ([]byte, sessionTool) {
	t.Helper()
	endpoint := "https://old.example.com/execute"
	tool := sessionTool{
		ID: uuid.New(), Name: "lookup_order", Type: TypeWebhook,
		Description: "Look up an order", Parameters: json.RawMessage(`{"type":"object"}`),
		EndpointURL: &endpoint, TimeoutMS: 2500,
	}
	value, err := json.Marshal(map[string]any{"engine": "realtime", "tools": []sessionTool{tool}})
	if err != nil {
		t.Fatal(err)
	}
	return value, tool
}

func TestResolveSessionToolAuthorizesOnlyFrozenIdentity(t *testing.T) {
	value, expected := snapshotFixture(t)
	actual, err := resolveSessionTool(value, uuid.Nil, expected.Name)
	if err != nil || actual.ID != expected.ID {
		t.Fatalf("frozen tool = %+v, %v", actual, err)
	}
	if _, err := resolveSessionTool(value, expected.ID, expected.Name); err != nil {
		t.Fatalf("matching ID/name rejected: %v", err)
	}
	cases := []struct {
		id   uuid.UUID
		name string
	}{
		{uuid.Nil, "newly_added_tool"},
		{uuid.New(), expected.Name},
		{expected.ID, "renamed_tool"},
		{uuid.Nil, ""},
	}
	for _, candidate := range cases {
		if _, err := resolveSessionTool(value, candidate.id, candidate.name); err == nil {
			t.Fatalf("unauthorized tool accepted: %+v", candidate)
		}
	}
}

func TestBindExecutionToolPreservesFrozenSettings(t *testing.T) {
	value, expected := snapshotFixture(t)
	frozen, err := resolveSessionTool(value, uuid.Nil, expected.Name)
	if err != nil {
		t.Fatal(err)
	}
	replacementURL := "https://new.example.com/tool"
	current := sqlc.VoiceAgentTool{
		ID: expected.ID, Enabled: true, Type: TypeBuiltin, Name: "renamed_tool",
		Description: "changed", Parameters: []byte(`{"changed":true}`),
		EndpointUrl: &replacementURL, TimeoutMs: 30000,
	}
	got, err := bindExecutionTool(current, frozen)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != expected.ID || got.Name != expected.Name || got.Type != expected.Type ||
		got.TimeoutMs != expected.TimeoutMS || got.Description != expected.Description ||
		got.EndpointUrl == nil || *got.EndpointUrl != *expected.EndpointURL ||
		string(got.Parameters) != string(expected.Parameters) {
		t.Fatalf("active session used mutable tool configuration: %+v", got)
	}
}

func TestBindExecutionToolRejectsRevocation(t *testing.T) {
	_, frozen := snapshotFixture(t)
	for _, current := range []sqlc.VoiceAgentTool{
		{ID: frozen.ID, Enabled: false},
		{ID: uuid.New(), Enabled: true},
	} {
		if _, err := bindExecutionTool(current, frozen); err == nil {
			t.Fatalf("revoked tool accepted: %+v", current)
		}
	}
}

func TestResolveSessionToolRejectsLegacyOrMalformedSnapshot(t *testing.T) {
	_, frozen := snapshotFixture(t)
	legacy := `{"tools":[{"id":"` + frozen.ID.String() + `","name":"lookup_order","parameters":{}}]}`
	duplicate := `{"tools":[{"id":"` + frozen.ID.String() + `","name":"lookup_order","type":"webhook","endpoint_url":"https://old.example.com","timeout_ms":2500,"parameters":{}},{"id":"` + uuid.New().String() + `","name":"lookup_order","type":"webhook","endpoint_url":"https://old.example.com","timeout_ms":2500,"parameters":{}}]}`
	for name, raw := range map[string]string{
		"legacy": legacy, "duplicate": duplicate, "empty": `{}`,
		"null": `{"tools":null}`, "invalid": `{oops`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := resolveSessionTool([]byte(raw), uuid.Nil, frozen.Name); err == nil {
				t.Fatal("invalid snapshot was authorized")
			}
		})
	}
}

func TestValidateExecutionSessionRejectsIncorrectScope(t *testing.T) {
	req := ExecuteRequest{
		OrganizationID: uuid.New(), VoiceAgentID: uuid.New(),
		SessionID: uuid.New(), CallID: uuid.New(),
	}
	stored := sqlc.VoiceAgentSession{
		ID: req.SessionID, OrganizationID: req.OrganizationID,
		VoiceAgentID: req.VoiceAgentID, CallID: req.CallID, State: "active",
	}
	if err := validateExecutionSession(stored, req); err != nil {
		t.Fatal(err)
	}
	bad := stored
	bad.CallID = uuid.New()
	if err := validateExecutionSession(bad, req); err == nil {
		t.Fatal("cross-call execution accepted")
	}
	bad = stored
	bad.OrganizationID = uuid.New()
	if err := validateExecutionSession(bad, req); err == nil {
		t.Fatal("cross-organization execution accepted")
	}
	bad = stored
	bad.VoiceAgentID = uuid.New()
	if err := validateExecutionSession(bad, req); err == nil {
		t.Fatal("cross-agent execution accepted")
	}
	bad = stored
	bad.State = "completed"
	if err := validateExecutionSession(bad, req); err == nil {
		t.Fatal("terminal session execution accepted")
	}
}
