package packages

import (
	"strings"
	"testing"
)

func TestParseComposablePackage(t *testing.T) {
	manifest, err := Parse([]byte(`{
		"schema_version":"1",
		"name":"receptionist",
		"agents":[{
			"alias":"receptionist",
			"name":"Receptionist",
			"engine":"composable",
			"instructions":"Answer the company phone.",
			"providers":[
				{"role":"stt","provider":"deepgram"},
				{"role":"llm","provider":"groq","config":{"model":"example-model"}},
				{"role":"tts","provider":"cartesia"}
			],
			"tools":["take_message"]
		}],
		"tools":[{
			"alias":"take_message",
			"type":"webhook",
			"name":"take_message",
			"parameters":{"type":"object"}
		}],
		"routing":{"default_agent":"receptionist"}
	}`))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if manifest.Agents[0].Engine != EngineComposable {
		t.Fatalf("engine = %q, want %q", manifest.Agents[0].Engine, EngineComposable)
	}
}

func TestParseRealtimePackage(t *testing.T) {
	_, err := Parse([]byte(`{
		"schema_version":"1",
		"name":"realtime-agent",
		"agents":[{
			"alias":"primary",
			"name":"Primary",
			"engine":"realtime",
			"instructions":"Help the caller.",
			"providers":[{"role":"realtime","provider":"openai"}]
		}]
	}`))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
}

func TestValidateRejectsEngineRoleMismatch(t *testing.T) {
	_, err := Parse([]byte(`{
		"schema_version":"1",
		"name":"bad-package",
		"agents":[{
			"alias":"primary",
			"name":"Primary",
			"engine":"realtime",
			"instructions":"Help the caller.",
			"providers":[{"role":"llm","provider":"openai"}]
		}]
	}`))
	if err == nil || !strings.Contains(err.Error(), `requires provider role "realtime"`) {
		t.Fatalf("Parse() error = %v", err)
	}
}

func TestValidateRejectsUnknownRoutingAgent(t *testing.T) {
	_, err := Parse([]byte(`{
		"schema_version":"1",
		"name":"bad-routing",
		"agents":[{
			"alias":"primary",
			"name":"Primary",
			"engine":"realtime",
			"instructions":"Help the caller.",
			"providers":[{"role":"realtime","provider":"openai"}]
		}],
		"routing":{"default_agent":"missing"}
	}`))
	if err == nil || !strings.Contains(err.Error(), `unknown agent "missing"`) {
		t.Fatalf("Parse() error = %v", err)
	}
}

func TestParseRejectsTenantCredentialIDs(t *testing.T) {
	_, err := Parse([]byte(`{
		"schema_version":"1",
		"name":"credential-leak",
		"agents":[{
			"alias":"primary",
			"name":"Primary",
			"engine":"realtime",
			"instructions":"Help the caller.",
			"providers":[{
				"role":"realtime",
				"provider":"openai",
				"credential_id":"00000000-0000-0000-0000-000000000001"
			}]
		}]
	}`))
	if err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("Parse() error = %v", err)
	}
}
