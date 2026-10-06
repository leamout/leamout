package orchestration

import "testing"

func TestToolDefinitionsFromSnapshot(t *testing.T) {
	definitions, err := ToolDefinitionsFromSnapshot([]byte(`{
        "engine":"composable",
        "instructions":"Help the caller.",
        "voice":null,
        "language":"en",
        "engine_config":{},
        "interruption_policy":"allow",
        "recording_policy":"none",
        "providers":[],
        "tools":[
            {
                "id":"46b5e4f1-c076-44b0-bf95-aa986e333ad7",
                "name":"lookup_order",
                "description":"Look up an order.",
                "parameters":{"type":"object"}
            }
        ]
    }`))
	if err != nil {
		t.Fatalf("ToolDefinitionsFromSnapshot() error = %v", err)
	}
	if len(definitions) != 1 || definitions[0].Name != "lookup_order" {
		t.Fatalf("definitions = %+v", definitions)
	}
}
