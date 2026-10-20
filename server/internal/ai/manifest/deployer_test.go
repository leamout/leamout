package manifest

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	agentcontract "github.com/leamout/contracts/agent"
	"github.com/leamout/contracts/ai"
	"github.com/leamout/leamout/server/internal/ai/agents"
	"github.com/leamout/leamout/server/internal/ai/providers"
	"github.com/leamout/leamout/server/internal/ai/tools"
	"github.com/leamout/leamout/server/internal/database/sqlc"
)

func TestPortableDeploymentFailureDoesNotConsumeAgentQuota(t *testing.T) {
	ctx, pool := deploymentTestDatabase(t)
	organizationID := deploymentTestOrganization(t, ctx, pool)
	otherOrganizationID := deploymentTestOrganization(t, ctx, pool)
	credentialID := deploymentTestCredential(t, ctx, pool, organizationID, "gemini")
	foreignCredentialID := deploymentTestCredential(t, ctx, pool, otherOrganizationID, "gemini")
	wrongProviderID := deploymentTestCredential(t, ctx, pool, organizationID, "openai")

	queries := sqlc.New(pool)
	providerService := providers.NewService(providers.NewRepository(queries), nil)
	agentRepository := agents.NewRepository(queries)
	agentRepository.SetDatabase(pool)
	agentService := agents.NewService(agentRepository, providerService)
	limit := int64(2)
	agentService.SetPlanAgentLimit(func(context.Context, uuid.UUID) (int64, bool, error) {
		return limit, true, nil
	})
	// No signing cipher: creating the second agent's webhook fails after
	// agent rows and provider bindings have already been written.
	deployer := NewDeployer(
		agentService,
		tools.NewService(tools.NewRepository(queries), nil),
		providerService,
		pool,
		NewValidator(),
	)
	pkg := deploymentTestPackage()
	for name, integrations := range map[string]map[string]uuid.UUID{
		"missing":              nil,
		"foreign organization": {"gemini": foreignCredentialID},
		"wrong provider":       {"gemini": wrongProviderID},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := deployer.DeployPortable(ctx, organizationID, pkg, DeploymentInput{
				Integrations: integrations,
			}); err == nil {
				t.Fatal("expected integration validation to fail")
			}
			assertDeployedAgentCount(t, ctx, pool, organizationID, 0)
		})
	}

	input := DeploymentInput{Integrations: map[string]uuid.UUID{"gemini": credentialID}}
	if _, err := deployer.DeployPortable(ctx, organizationID, pkg, input); err == nil {
		t.Fatal("expected webhook signing failure")
	}
	assertDeployedAgentCount(t, ctx, pool, organizationID, 0)

	// Capacity failure on the second agent also rolls back the first one.
	pkg.Agents["beta"] = deploymentTestAgent("Beta", nil)
	limit = 1
	if _, err := deployer.DeployPortable(ctx, organizationID, pkg, input); err == nil {
		t.Fatal("expected organization limit failure")
	}
	assertDeployedAgentCount(t, ctx, pool, organizationID, 0)

	// A subsequent successful deployment must have the full original quota.
	limit = 2
	result, err := deployer.DeployPortable(ctx, organizationID, pkg, input)
	if err != nil {
		t.Fatalf("retry deployment: %v", err)
	}
	if len(result.Agents) != 2 {
		t.Fatalf("deployed agents = %d, want 2", len(result.Agents))
	}
	assertDeployedAgentCount(t, ctx, pool, organizationID, 2)
}

func deploymentTestPackage() PortablePackage {
	return PortablePackage{
		Agents: map[string]agentcontract.Manifest{
			"alpha": deploymentTestAgent("Alpha", nil),
			"beta":  deploymentTestAgent("Beta", []string{"notify"}),
		},
		Tools: []agentcontract.ToolDefinition{{
			Type:        agentcontract.ToolTypeWebhook,
			Name:        "notify",
			Description: "Notify the operator.",
			EndpointURL: "https://example.com/notify",
			Parameters:  json.RawMessage(`{"type":"object"}`),
		}},
		Routing: agentcontract.RoutingManifest{
			SchemaVersion: agentcontract.SchemaVersion,
			Name:          "Test routing",
			DefaultAgent:  "alpha",
		},
	}
}

func deploymentTestAgent(name string, toolNames []string) agentcontract.Manifest {
	return agentcontract.Manifest{
		SchemaVersion: agentcontract.SchemaVersion,
		Name:          name,
		Engine:        agentcontract.EngineRealtime,
		Instructions:  "Answer the caller.",
		Providers: []agentcontract.ProviderBinding{{
			Role:     ai.KindRealtime,
			Provider: "gemini",
			Config:   json.RawMessage(`{"model":"gemini-3.8-live"}`),
		}},
		Tools: toolNames,
	}
}

func deploymentTestDatabase(t *testing.T) (context.Context, *pgxpool.Pool) {
	t.Helper()
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set TEST_DATABASE_URL to run deployment database tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	t.Cleanup(cancel)

	admin, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	schema := "agent_deploy_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+pgx.Identifier{schema}.Sanitize()); err != nil {
		admin.Close()
		t.Fatal(err)
	}
	admin.Close()
	t.Cleanup(func() {
		cleanup, err := pgxpool.New(context.Background(), databaseURL)
		if err == nil {
			_, _ = cleanup.Exec(context.Background(), "DROP SCHEMA "+pgx.Identifier{schema}.Sanitize()+" CASCADE")
			cleanup.Close()
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

	migrations, err := filepath.Glob("../../../migrations/*.sql")
	if err != nil || len(migrations) == 0 {
		t.Fatalf("list migrations: %v", err)
	}
	for _, path := range migrations {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, string(body)); err != nil {
			t.Fatalf("apply migration %s: %v", filepath.Base(path), err)
		}
	}
	return ctx, pool
}

func deploymentTestOrganization(t *testing.T, ctx context.Context, pool *pgxpool.Pool) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := pool.Exec(ctx, "INSERT INTO organizations (id, name) VALUES ($1, $2)", id, "Agent deployment test"); err != nil {
		t.Fatal(err)
	}
	return id
}

func deploymentTestCredential(t *testing.T, ctx context.Context, pool *pgxpool.Pool, organizationID uuid.UUID, provider string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := pool.Exec(ctx,
		"INSERT INTO ai_provider_credentials (id, organization_id, provider, name, secret_ciphertext) VALUES ($1, $2, $3, $4, $5)",
		id, organizationID, provider, "test", "encrypted-test-secret",
	); err != nil {
		t.Fatal(err)
	}
	return id
}

func assertDeployedAgentCount(t *testing.T, ctx context.Context, pool *pgxpool.Pool, organizationID uuid.UUID, want int) {
	t.Helper()
	for _, table := range []string{"voice_agents", "voice_agent_provider_bindings", "voice_agent_tools"} {
		var count int
		if err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM "+table+" WHERE organization_id = $1", organizationID).Scan(&count); err != nil {
			t.Fatal(err)
		}
		expected := 0
		if table == "voice_agents" || table == "voice_agent_provider_bindings" {
			expected = want
		}
		if count != expected {
			t.Fatalf("%s records = %d, want %d", table, count, expected)
		}
	}
}
