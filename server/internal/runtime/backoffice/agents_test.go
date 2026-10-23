package backoffice

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/leamout/leamout/server/internal/database/sqlc"
)

func testAgentInspection(t *testing.T, pool *pgxpool.Pool, cookie string, request func(string, string, string) *httptest.ResponseRecorder) {
	t.Helper()
	ctx := t.Context()
	queries := sqlc.New(pool)
	organization, err := queries.CreateOrganization(ctx, "Agent inspection customer")
	if err != nil {
		t.Fatal(err)
	}
	other, err := queries.CreateOrganization(ctx, "Other agent customer")
	if err != nil {
		t.Fatal(err)
	}
	for i := range 27 {
		if _, err := pool.Exec(ctx, `INSERT INTO voice_agents (organization_id, name, instructions) VALUES ($1, $2, 'Be helpful')`, organization.ID, fmt.Sprintf("Inspection agent %02d", i)); err != nil {
			t.Fatal(err)
		}
	}
	specialID := uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO voice_agents (id, organization_id, name, instructions, engine, status, engine_config)
 VALUES ($1, $2, '<script>alert(1)</script>', '<script>prompt(1)</script>', 'realtime', 'disabled', '{"private_value":"DO_NOT_RENDER_ENGINE_CONFIG"}')`, specialID, other.ID); err != nil {
		t.Fatal(err)
	}
	if err := queries.DeleteOrganization(ctx, other.ID); err != nil {
		t.Fatal(err)
	}
	base := "/agents?organization_id=" + organization.ID.String()
	first := request(http.MethodGet, base, cookie)
	second := request(http.MethodGet, base+"&page=2", cookie)
	if first.Code != http.StatusOK || strings.Count(first.Body.String(), `class="link link-primary font-medium"`) != 25 {
		t.Fatalf("agent first page: %d / %s", first.Code, first.Body.String())
	}
	if second.Code != http.StatusOK || strings.Count(second.Body.String(), `class="link link-primary font-medium"`) != 2 {
		t.Fatalf("agent second page: %d / %s", second.Code, second.Body.String())
	}
	if first.Header().Get("Cache-Control") != "no-store" || strings.Contains(first.Body.String(), specialID.String()) {
		t.Fatal("agent list must be uncached and scoped to requested organization")
	}
	for _, query := range []string{"?q=" + specialID.String(), "?q=Other&status=disabled&engine=realtime"} {
		response := request(http.MethodGet, "/agents"+query, cookie)
		if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), specialID.String()) || strings.Contains(response.Body.String(), "Inspection agent") {
			t.Fatalf("agent filters: %d / %s", response.Code, response.Body.String())
		}
	}
	detail := request(http.MethodGet, "/agents/"+specialID.String(), cookie)
	if detail.Code != http.StatusOK || !strings.Contains(detail.Body.String(), "deleted") || !strings.Contains(detail.Body.String(), "&lt;script&gt;") {
		t.Fatalf("disabled agent detail in deleted organization: %d / %s", detail.Code, detail.Body.String())
	}
	if strings.Contains(detail.Body.String(), "<script>alert") || strings.Contains(detail.Body.String(), "<script>prompt") || strings.Contains(detail.Body.String(), "DO_NOT_RENDER_ENGINE_CONFIG") {
		t.Fatal("detail rendered unsafe markup or raw engine configuration")
	}
	empty := request(http.MethodGet, "/agents?engine=composable&status=disabled", cookie)
	if empty.Code != http.StatusOK || !strings.Contains(empty.Body.String(), "No agents match") {
		t.Fatal("empty filtered results must render an empty state")
	}
	for _, path := range []string{"/agents/not-a-uuid", "/agents/" + uuid.NewString()} {
		if w := request(http.MethodGet, path, cookie); w.Code != http.StatusNotFound {
			t.Fatalf("missing agent %s: %d", path, w.Code)
		}
	}
	if w := request(http.MethodGet, "/agents?organization_id=invalid", cookie); w.Code != http.StatusBadRequest {
		t.Fatalf("invalid organization filter: %d", w.Code)
	}
	if w := request(http.MethodPost, "/agents/"+specialID.String(), cookie); w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("agent detail must be read-only: %d", w.Code)
	}
}
