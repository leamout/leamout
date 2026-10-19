package manifest

import (
	"encoding/json"
	"strings"
	"testing"
	"testing/fstest"

	agentcontract "github.com/leamout/contracts/agent"
	"github.com/leamout/contracts/ai"
)

func samplePortablePackage() PortablePackage {
	receptionist := agentcontract.Manifest{
		SchemaVersion: agentcontract.SchemaVersion,
		Name:          "Receptionist",
		Engine:        agentcontract.EngineRealtime,
		Instructions:  "Greet callers.",
		Providers: []agentcontract.ProviderBinding{
			{
				Role:     ai.KindRealtime,
				Provider: "gemini",
				Config:   json.RawMessage(`{"model":"gemini-3.8-live"}`),
			},
		},
		Tools: []string{"hangup_call", "directory"},
	}
	afterHours := receptionist
	afterHours.Name = "After hours"
	afterHours.Tools = []string{"hangup_call"}
	return PortablePackage{
		Agents: map[string]agentcontract.Manifest{
			"receptionist": receptionist,
			"after-hours": afterHours,
		},
		Tools: []agentcontract.ToolDefinition{
			{
				Type:        agentcontract.ToolTypeBuiltin,
				Name:        "hangup_call",
				Description: "End call",
				Parameters:  json.RawMessage(`{"type":"object"}`),
			},
			{
				Type:        agentcontract.ToolTypeWebhook,
				Name:        "directory",
				Description: "Search directory",
				Parameters:  json.RawMessage(`{"type":"object"}`),
				EndpointURL: "${DIRECTORY_URL}",
			},
		},
		Routing: agentcontract.RoutingManifest{
			SchemaVersion: agentcontract.SchemaVersion,
			Name:          "Receptionist routing",
			DefaultAgent:  "receptionist",
			Schedules: map[string]agentcontract.Schedule{
				"closed": {
					Timezone: "UTC",
					Hours: []agentcontract.ScheduleHour{
						{
							Days:  []string{"mon"},
							Start: "09:00",
							End:   "17:00",
						},
					},
				},
			},
			Routes: []agentcontract.Route{
				{
					Schedule: "closed",
					Agent:    "after-hours",
				},
			},
		},
	}
}

func TestValidatePortable(t *testing.T) {
	t.Parallel()
	if err := NewValidator().ValidatePortable(samplePortablePackage()); err != nil {
		t.Fatalf("valid package rejected: %v", err)
	}
}

func TestValidatePortableRejectsBrokenReferences(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		change func(*PortablePackage)
	}{
		{
			name: "missing default agent",
			change: func(pkg *PortablePackage) {
				pkg.Routing.DefaultAgent = "absent"
			},
		},
		{
			name: "missing tool",
			change: func(pkg *PortablePackage) {
				agent := pkg.Agents["after-hours"]
				agent.Tools = append(agent.Tools, "unknown")
				pkg.Agents["after-hours"] = agent
			},
		},
		{
			name: "unsafe alias",
			change: func(pkg *PortablePackage) {
				pkg.Agents["../other"] = pkg.Agents["receptionist"]
			},
		},
		{
			name: "duplicate agent names",
			change: func(pkg *PortablePackage) {
				agent := pkg.Agents["after-hours"]
				agent.Name = "RECEPTIONIST"
				pkg.Agents["after-hours"] = agent
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pkg := samplePortablePackage()
			tc.change(&pkg)
			if err := NewValidator().ValidatePortable(pkg); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestPreparePortableToolsOnlyReferencedAndPreflight(t *testing.T) {
	t.Parallel()
	pkg := samplePortablePackage()
	got, err := preparePortableTools(pkg, map[string]string{
		"DIRECTORY_URL": "https://example.com/directory",
	})
	if err != nil {
		t.Fatalf("prepare: %v", err)
	}
	if len(got["after-hours"]) != 1 || got["after-hours"][0].Name != "hangup_call" {
		t.Fatalf("after-hours tools = %#v", got["after-hours"])
	}
	if len(got["receptionist"]) != 2 || got["receptionist"][1].EndpointURL != "https://example.com/directory" {
		t.Fatalf("receptionist tools = %#v", got["receptionist"])
	}
	for _, variables := range []map[string]string{
		nil,
		{"DIRECTORY_URL": "http://insecure.example/directory"},
		{"DIRECTORY_URL": "not-a-url"},
	} {
		if _, err := preparePortableTools(pkg, variables); err == nil {
			t.Fatalf("expected preflight error for variables %#v", variables)
		}
	}
}

func TestReadPortableDirectory(t *testing.T) {
	t.Parallel()
	pkg := samplePortablePackage()
	files := fstest.MapFS{}
	addJSON := func(name string, value any) {
		t.Helper()
		data, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		files[name] = &fstest.MapFile{Data: data}
	}
	addJSON("tools.json", pkg.Tools)
	addJSON("routing.json", pkg.Routing)
	for alias, agent := range pkg.Agents {
		addJSON("agents/"+alias+".json", agent)
	}
	loaded, err := ReadPortable(files)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Agents) != 2 || loaded.Routing.DefaultAgent != "receptionist" {
		t.Fatalf("loaded package = %#v", loaded)
	}
	files["agents/other.txt"] = &fstest.MapFile{Data: []byte("unexpected")}
	if _, err := ReadPortable(files); err == nil || !strings.Contains(err.Error(), "unexpected agents entry") {
		t.Fatalf("expected unexpected entry error, got %v", err)
	}
}

func TestReadPortableRejectsInvalidJSON(t *testing.T) {
	t.Parallel()
	files := fstest.MapFS{
		"tools.json": &fstest.MapFile{Data: []byte("not-json")},
	}
	if _, err := ReadPortable(files); err == nil {
		t.Fatal("expected malformed JSON error")
	}
}
