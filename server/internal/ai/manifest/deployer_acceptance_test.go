package manifest

import (
	"context"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leamout/leamout/server/internal/ai/agents"
	"github.com/leamout/leamout/server/internal/ai/providers"
	"github.com/leamout/leamout/server/internal/ai/tools"
	"github.com/leamout/leamout/server/internal/database/sqlc"
	"github.com/leamout/leamout/server/internal/security/encryption"
)

type portableDeploymentAcceptance struct {
	t         *testing.T
	ctx       context.Context
	pool      *pgxpool.Pool
	agents    *agents.Service
	providers *providers.Service
	tools     *tools.Service
	deployer  *Deployer
}

func TestPortableDeploymentWithOrganizationIntegrations(t *testing.T) {
	h := newPortableDeploymentAcceptance(t)
	organizationID := h.createOrganization("Portable deployment")
	otherOrganizationID := h.createOrganization("Other organization")

	credential, err := h.providers.CreateCredential(
		h.ctx,
		organizationID,
		providers.CreateCredentialRequest{
			Provider: "gemini",
			Name:     "Gemini integration",
			Secret:   "offline-acceptance-credential",
		},
	)
	if err != nil {
		t.Fatalf("create organization integration: %v", err)
	}
	if credential.ConnectionState != providers.ConnectionUnchecked {
		t.Fatalf("credential state = %q, want unchecked", credential.ConnectionState)
	}

	otherCredential, err := h.providers.CreateCredential(
		h.ctx,
		otherOrganizationID,
		providers.CreateCredentialRequest{
			Provider: "gemini",
			Name:     "Other Gemini integration",
			Secret:   "different-offline-credential",
		},
	)
	if err != nil {
		t.Fatalf("create other organization integration: %v", err)
	}
	if otherCredential.ID == credential.ID {
		t.Fatal("organizations were assigned the same integration")
	}

	pkg := samplePortablePackage()
	result, err := h.deployer.DeployPortable(
		h.ctx,
		organizationID,
		pkg,
		DeploymentInput{
			Integrations: map[string]uuid.UUID{"gemini": credential.ID},
			Variables: map[string]string{
				"DIRECTORY_URL": "https://example.com/directory",
			},
		},
	)
	if err != nil {
		t.Fatalf("deploy portable package: %v", err)
	}
	if len(result.Agents) != 2 || len(result.RoutingPlan.AgentIDs) != 2 {
		t.Fatalf("deployed agents or routing ids missing: %#v", result)
	}
	if result.RoutingPlan.Manifest.DefaultAgent != "receptionist" {
		t.Fatalf("default routing alias = %q", result.RoutingPlan.Manifest.DefaultAgent)
	}
	if len(result.RoutingPlan.Manifest.Routes) != 1 ||
		result.RoutingPlan.Manifest.Routes[0].Agent != "after-hours" {
		t.Fatalf("schedule routing was not preserved: %#v", result.RoutingPlan.Manifest.Routes)
	}

	for alias, wantName := range map[string]string{
		"receptionist": "Receptionist",
		"after-hours":  "After hours",
	} {
		deployed, ok := result.Agents[alias]
		if !ok || deployed.AgentID == uuid.Nil {
			t.Fatalf("missing deployed agent %q: %#v", alias, deployed)
		}
		if result.RoutingPlan.AgentIDs[alias] != deployed.AgentID {
			t.Fatalf("routing id for %q does not match deployed agent", alias)
		}
		agent, getErr := h.agents.Get(h.ctx, organizationID, deployed.AgentID)
		if getErr != nil {
			t.Fatalf("get %q: %v", alias, getErr)
		}
		if agent.OrganizationID != organizationID || agent.Name != wantName || agent.Engine != "realtime" {
			t.Fatalf("unexpected deployed %q: %#v", alias, agent)
		}

		bindings, listErr := h.providers.ListBindings(h.ctx, organizationID, deployed.AgentID)
		if listErr != nil {
			t.Fatalf("list %q provider bindings: %v", alias, listErr)
		}
		if len(bindings) != 1 || bindings[0].Role != "realtime" ||
			bindings[0].Provider != "gemini" || bindings[0].CredentialID == nil ||
			*bindings[0].CredentialID != credential.ID {
			t.Fatalf("unexpected %q provider bindings: %#v", alias, bindings)
		}
		resolved, resolveErr := h.providers.Resolve(h.ctx, organizationID, deployed.AgentID)
		if resolveErr != nil {
			t.Fatalf("resolve %q provider credentials: %v", alias, resolveErr)
		}
		if len(resolved) != 1 || resolved[0].APIKey != "offline-acceptance-credential" {
			t.Fatalf("%q provider integration did not resolve its scoped credential", alias)
		}

		if _, getErr := h.agents.Get(h.ctx, otherOrganizationID, deployed.AgentID); getErr == nil {
			t.Fatalf("other organization can read %q", alias)
		}
		otherBindings, listErr := h.providers.ListBindings(
			h.ctx,
			otherOrganizationID,
			deployed.AgentID,
		)
		if listErr != nil || len(otherBindings) != 0 {
			t.Fatalf("other organization bindings for %q = %#v, error %v", alias, otherBindings, listErr)
		}
		otherTools, listErr := h.tools.List(h.ctx, otherOrganizationID, deployed.AgentID)
		if listErr != nil || len(otherTools) != 0 {
			t.Fatalf("other organization tools for %q = %#v, error %v", alias, otherTools, listErr)
		}
	}

	if result.Agents["receptionist"].AgentID == result.Agents["after-hours"].AgentID {
		t.Fatal("both aliases deployed to one agent")
	}

	receptionist := result.Agents["receptionist"]
	receptionistTools, err := h.tools.List(h.ctx, organizationID, receptionist.AgentID)
	if err != nil {
		t.Fatalf("list receptionist tools: %v", err)
	}
	if len(receptionistTools) != 2 {
		t.Fatalf("receptionist tools = %d, want 2", len(receptionistTools))
	}
	for _, tool := range receptionistTools {
		switch tool.Name {
		case "hangup_call":
			if tool.Type != tools.TypeBuiltin || tool.EndpointUrl != nil {
				t.Fatalf("invalid builtin tool: %#v", tool)
			}
		case "directory":
			if tool.Type != tools.TypeWebhook || tool.EndpointUrl == nil ||
				*tool.EndpointUrl != "https://example.com/directory" {
				t.Fatalf("invalid resolved webhook tool: %#v", tool)
			}
			secret := receptionist.SigningSecrets["directory"]
			if secret == "" {
				t.Fatal("deployment response omitted webhook signing secret")
			}
			storedSecret, readErr := h.tools.SigningSecret(
				h.ctx,
				organizationID,
				receptionist.AgentID,
				tool.ID,
			)
			if readErr != nil || storedSecret != secret {
				t.Fatalf("webhook secret did not round-trip: %v", readErr)
			}
			if _, readErr := h.tools.SigningSecret(
				h.ctx,
				otherOrganizationID,
				receptionist.AgentID,
				tool.ID,
			); readErr == nil {
				t.Fatal("other organization can read webhook signing secret")
			}
		default:
			t.Fatalf("unexpected receptionist tool %q", tool.Name)
		}
	}
	if len(receptionist.SigningSecrets) != 1 {
		t.Fatalf("receptionist signing secrets = %#v", receptionist.SigningSecrets)
	}

	afterHours := result.Agents["after-hours"]
	afterHoursTools, err := h.tools.List(h.ctx, organizationID, afterHours.AgentID)
	if err != nil {
		t.Fatalf("list after-hours tools: %v", err)
	}
	if len(afterHoursTools) != 1 || afterHoursTools[0].Name != "hangup_call" ||
		len(afterHours.SigningSecrets) != 0 {
		t.Fatalf("after-hours tools or secrets are unexpected: %#v, %#v", afterHoursTools, afterHours.SigningSecrets)
	}

	if got := h.countAgents(organizationID); got != 2 {
		t.Fatalf("stored voice agent count = %d, want 2", got)
	}
}

func TestPortableDeploymentRejectsInvalidPreflightWithoutWrites(t *testing.T) {
	h := newPortableDeploymentAcceptance(t)
	organizationID := h.createOrganization("Invalid package preflight")
	cases := []struct {
		name      string
		change    func(*PortablePackage)
		variables map[string]string
	}{
		{
			name: "missing webhook variable",
		},
		{
			name: "insecure webhook variable",
			variables: map[string]string{
				"DIRECTORY_URL": "http://insecure.example.com/directory",
			},
		},
		{
			name: "unknown default agent",
			change: func(pkg *PortablePackage) {
				pkg.Routing.DefaultAgent = "not-installed"
			},
			variables: map[string]string{
				"DIRECTORY_URL": "https://example.com/directory",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pkg := samplePortablePackage()
			if tc.change != nil {
				tc.change(&pkg)
			}
			if _, err := h.deployer.DeployPortable(
				h.ctx,
				organizationID,
				pkg,
				DeploymentInput{Variables: tc.variables},
			); err == nil {
				t.Fatal("invalid deployment was accepted")
			}
			if got := h.countAgents(organizationID); got != 0 {
				t.Fatalf("failed preflight left %d voice agents", got)
			}
		})
	}
}

func newPortableDeploymentAcceptance(t *testing.T) *portableDeploymentAcceptance {
	t.Helper()
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set TEST_DATABASE_URL to run portable deployment acceptance")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	t.Cleanup(cancel)

	adminPool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	schema := "manifest_acceptance_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := adminPool.Exec(ctx, "CREATE SCHEMA "+pgx.Identifier{schema}.Sanitize()); err != nil {
		adminPool.Close()
		t.Fatal(err)
	}
	adminPool.Close()

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	config.ConnConfig.RuntimeParams["search_path"] = schema + ",public"
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		pool.Close()
		cleanupPool, cleanupErr := pgxpool.New(context.Background(), databaseURL)
		if cleanupErr == nil {
			_, _ = cleanupPool.Exec(
				context.Background(),
				"DROP SCHEMA "+pgx.Identifier{schema}.Sanitize()+" CASCADE",
			)
			cleanupPool.Close()
		}
	})

	migrations, err := filepath.Glob("../../../migrations/*.sql")
	if err != nil || len(migrations) == 0 {
		t.Fatalf("find migrations: %v", err)
	}
	for _, path := range migrations {
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Fatal(readErr)
		}
		if _, execErr := pool.Exec(ctx, string(content)); execErr != nil {
			t.Fatalf("migration %s: %v", filepath.Base(path), execErr)
		}
	}

	key := base64.RawURLEncoding.EncodeToString([]byte("portable-deployment-test-key-32!"))
	cipher, err := encryption.New(key)
	if err != nil {
		t.Fatal(err)
	}
	queries := sqlc.New(pool)
	providerService := providers.NewService(providers.NewRepository(queries), cipher)
	agentRepository := agents.NewRepository(queries)
	agentRepository.SetDatabase(pool)
	agentService := agents.NewService(agentRepository, providerService)
	agentService.SetPlanAgentLimit(func(context.Context, uuid.UUID) (int64, bool, error) {
		return 5, true, nil
	})
	toolService := tools.NewService(tools.NewRepository(queries), cipher)

	return &portableDeploymentAcceptance{
		t:         t,
		ctx:       ctx,
		pool:      pool,
		agents:    agentService,
		providers: providerService,
		tools:     toolService,
		deployer:  NewDeployer(agentService, toolService, NewValidator()),
	}
}

func (h *portableDeploymentAcceptance) createOrganization(name string) uuid.UUID {
	h.t.Helper()
	id := uuid.New()
	if _, err := h.pool.Exec(
		h.ctx,
		"INSERT INTO organizations (id, name) VALUES ($1, $2)",
		id,
		name,
	); err != nil {
		h.t.Fatal(err)
	}
	return id
}

func (h *portableDeploymentAcceptance) countAgents(organizationID uuid.UUID) int {
	h.t.Helper()
	var count int
	if err := h.pool.QueryRow(
		h.ctx,
		"SELECT COUNT(*) FROM voice_agents WHERE organization_id = $1",
		organizationID,
	).Scan(&count); err != nil {
		h.t.Fatal(err)
	}
	return count
}
