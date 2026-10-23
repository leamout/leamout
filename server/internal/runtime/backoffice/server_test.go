package backoffice

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/leamout/leamout/server/internal/backoffice/agents"
	backofficeauth "github.com/leamout/leamout/server/internal/backoffice/auth"
	"github.com/leamout/leamout/server/internal/backoffice/organizations"
	"github.com/leamout/leamout/server/internal/database/sqlc"
	identityauth "github.com/leamout/leamout/server/internal/identity/auth"
	"github.com/leamout/leamout/server/internal/identity/session"
	"github.com/leamout/leamout/server/internal/integrations/keycloak"
	"github.com/leamout/leamout/server/internal/observability/logging"
	"github.com/leamout/leamout/server/internal/platform/middleware"
	"github.com/leamout/leamout/server/internal/security/authn"
	"github.com/leamout/leamout/server/internal/security/encryption"
)

func TestAnonymousRouteBoundary(t *testing.T) {
	logger := logging.New()
	router, err := newRouter(&modules{
		authentication: backofficeauth.NewHandler(&identityauth.Handler{}, nil, true),
		authn:          middleware.NewAuthnMiddleware(authn.NewResolver(nil, nil)),
		organizations:  organizations.NewHandler(nil, logger),
		agents:         agents.NewHandler(nil, logger),
	}, logger, "https://backoffice.example.test")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/organizations", "/organizations/" + uuid.NewString(), "/agents", "/agents/" + uuid.NewString()} {
		r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil)
		r.Header.Set("Authorization", "Bearer organization-token")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("GET %s with no session: %d", path, w.Code)
		}
	}
	for _, path := range []string{"/healthz", "/static/js/htmx.min.js", "/log-in"} {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil))
		if w.Code != http.StatusOK {
			t.Errorf("public endpoint %s: %d", path, w.Code)
		}
	}
}

func TestStandaloneAuthenticationRoutes(t *testing.T) {
	logger := logging.New()
	router, err := newRouter(&modules{
		authentication: backofficeauth.NewHandler(&identityauth.Handler{}, nil, true),
		authn:          middleware.NewAuthnMiddleware(authn.NewResolver(nil, nil)),
		organizations:  organizations.NewHandler(nil, logger),
		agents:         agents.NewHandler(nil, logger),
	}, logger, "https://backoffice.example.test")
	if err != nil {
		t.Fatal(err)
	}
	login := httptest.NewRecorder()
	router.ServeHTTP(login, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/log-in", nil))
	if login.Code != http.StatusOK || !strings.Contains(login.Body.String(), `href="/auth/login"`) {
		t.Fatal("standalone login page must link to its own identity flow")
	}
	callback := httptest.NewRecorder()
	router.ServeHTTP(callback, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/auth/callback?state=invalid", nil))
	if callback.Code != http.StatusUnauthorized {
		t.Fatalf("callback must reject missing state cookie: %d", callback.Code)
	}
	for _, origin := range []string{"https://backoffice.example.test", "https://other.example.test"} {
		r := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "https://backoffice.example.test/auth/logout", nil)
		r.Header.Set("Origin", origin)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, r)
		if origin == "https://other.example.test" {
			if response.Code != http.StatusForbidden {
				t.Fatalf("cross-origin logout status: %d", response.Code)
			}
		} else if response.Code != http.StatusSeeOther || response.Header().Get("Location") != "/log-in" {
			t.Fatalf("browser logout must return to standalone login: %d", response.Code)
		}
	}
}

func TestBackofficeAccessAndOrganizations(t *testing.T) {
	pool := backofficeDatabase(t)
	ctx := t.Context()
	queries := sqlc.New(pool)
	logger := logging.New()
	cipher, err := encryption.New(base64.RawURLEncoding.EncodeToString([]byte(strings.Repeat("a", 32))))
	if err != nil {
		t.Fatal(err)
	}
	// Unexpired encrypted tokens exercise the real session resolver without
	// requiring an external Keycloak server for this database acceptance test.
	sessions := session.NewService(session.NewRepository(queries), pool, nil, cipher)
	modules := &modules{
		authentication: backofficeauth.NewHandler(&identityauth.Handler{}, sessions, true),
		queries:        queries,
		authn:          middleware.NewAuthnMiddleware(authn.NewResolver(sessions, nil)),
		organizations:  organizations.NewHandler(organizations.NewRepository(queries), logger),
		agents:         agents.NewHandler(agents.NewRepository(queries), logger),
	}
	router, err := newRouter(modules, logger, "https://backoffice.example.test")
	if err != nil {
		t.Fatal(err)
	}
	request := func(method, path, cookie string) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequestWithContext(ctx, method, path, nil)
		if cookie != "" {
			r.AddCookie(&http.Cookie{Name: authn.SessionCookieName, Value: cookie})
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		return w
	}
	for _, path := range []string{"/organizations", "/organizations/" + uuid.NewString(), "/agents", "/agents/" + uuid.NewString()} {
		if w := request(http.MethodGet, path, ""); w.Code != http.StatusUnauthorized {
			t.Fatalf("anonymous %s: %d", path, w.Code)
		}
	}
	if w := request(http.MethodGet, "/healthz", ""); w.Code != http.StatusOK {
		t.Fatalf("health status: %d", w.Code)
	}
	if w := request(http.MethodGet, "/static/js/htmx.min.js", ""); w.Code != http.StatusOK {
		t.Fatalf("asset status: %d", w.Code)
	}
	user, err := queries.UpsertIdentityUser(ctx, sqlc.UpsertIdentityUserParams{
		IdentityIssuer:  "https://identity.example.test",
		IdentitySubject: uuid.NewString(),
		Email:           "operator@example.test",
	})
	if err != nil {
		t.Fatal(err)
	}
	cookie, stored, err := sessions.Create(ctx, user.ID, user.IdentitySubject, "test-sso", keycloak.Tokens{ValidUntil: time.Now().Add(time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/organizations", "/agents", "/agents/" + uuid.NewString()} {
		if w := request(http.MethodGet, path, cookie); w.Code != http.StatusForbidden {
			t.Fatalf("non-admin status: %d", w.Code)
		}
	}
	if _, err := pool.Exec(ctx, "UPDATE users SET is_platform_admin = TRUE WHERE id = $1", user.ID); err != nil {
		t.Fatal(err)
	}
	for i := range 27 {
		if _, err := queries.CreateOrganization(ctx, fmt.Sprintf("Customer %02d", i)); err != nil {
			t.Fatal(err)
		}
	}
	first := request(http.MethodGet, "/organizations", cookie)
	second := request(http.MethodGet, "/organizations?page=2", cookie)
	if first.Code != http.StatusOK || strings.Count(first.Body.String(), `class="link link-primary font-medium"`) != 25 {
		t.Fatalf("first page: %d / %s", first.Code, first.Body.String())
	}
	if second.Code != http.StatusOK || strings.Count(second.Body.String(), `class="link link-primary font-medium"`) != 2 {
		t.Fatalf("second page: %d / %s", second.Code, second.Body.String())
	}
	if first.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("operator responses must not be cached")
	}
	special, err := queries.CreateOrganization(ctx, "<script>alert(1)</script> Unique")
	if err != nil {
		t.Fatal(err)
	}
	detail := request(http.MethodGet, "/organizations/"+special.ID.String(), cookie)
	if detail.Code != http.StatusOK || strings.Contains(detail.Body.String(), "<script>alert(1)</script>") || !strings.Contains(detail.Body.String(), "No subscription") {
		t.Fatalf("detail must escape names and tolerate missing subscriptions: %d", detail.Code)
	}
	if err := queries.DeleteOrganization(ctx, special.ID); err != nil {
		t.Fatal(err)
	}
	filtered := request(http.MethodGet, "/organizations?q=Unique&status=deleted", cookie)
	if filtered.Code != http.StatusOK || !strings.Contains(filtered.Body.String(), special.ID.String()) || strings.Contains(filtered.Body.String(), "Customer 00") {
		t.Fatal("deleted organization search did not isolate the matching result")
	}
	if w := request(http.MethodGet, "/organizations/"+special.ID.String(), cookie); w.Code != http.StatusOK {
		t.Fatalf("deleted organization detail: %d", w.Code)
	}
	for _, path := range []string{"/organizations/not-a-uuid", "/organizations/" + uuid.NewString()} {
		if w := request(http.MethodGet, path, cookie); w.Code != http.StatusNotFound {
			t.Fatalf("missing detail status: %d", w.Code)
		}
	}
	if w := request(http.MethodGet, "/organizations?page=-1", cookie); w.Code != http.StatusBadRequest {
		t.Fatalf("invalid filter status: %d", w.Code)
	}
	testAgentInspection(t, pool, cookie, request)
	if _, err := pool.Exec(ctx, "UPDATE users SET is_platform_admin = FALSE WHERE id = $1", user.ID); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/organizations", "/agents", "/agents/" + uuid.NewString()} {
		if w := request(http.MethodGet, path, cookie); w.Code != http.StatusForbidden {
			t.Fatal("operator revocation did not take immediate effect")
		}
	}
	if _, err := pool.Exec(ctx, "UPDATE users SET is_platform_admin = TRUE WHERE id = $1", user.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, "UPDATE users SET disabled_at = NOW() WHERE id = $1", user.ID); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/organizations", "/agents", "/agents/" + uuid.NewString()} {
		if w := request(http.MethodGet, path, cookie); w.Code != http.StatusUnauthorized {
			t.Fatal("disabled operator retained access")
		}
	}
	if _, err := pool.Exec(ctx, "UPDATE users SET disabled_at = NULL WHERE id = $1", user.ID); err != nil {
		t.Fatal(err)
	}
	logout := request(http.MethodPost, "/auth/logout", cookie)
	if logout.Code != http.StatusSeeOther || logout.Header().Get("Location") != "/log-in" {
		t.Fatalf("operator sign-out status: %d", logout.Code)
	}
	if _, err := queries.GetSessionByID(ctx, stored.ID); err == nil {
		t.Fatal("browser sign-out did not revoke the stored session")
	}
	for _, path := range []string{"/organizations", "/agents", "/agents/" + uuid.NewString()} {
		if w := request(http.MethodGet, path, cookie); w.Code != http.StatusUnauthorized {
			t.Fatal("revoked session retained operator access")
		}
	}
}

func backofficeDatabase(t *testing.T) *pgxpool.Pool {
	t.Helper()
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set TEST_DATABASE_URL to run backoffice database acceptance")
	}
	ctx := t.Context()
	admin, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	schema := "backoffice_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+pgx.Identifier{schema}.Sanitize()); err != nil {
		admin.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, _ = admin.Exec(cleanup, "DROP SCHEMA "+pgx.Identifier{schema}.Sanitize()+" CASCADE")
		admin.Close()
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
		t.Fatalf("find migrations: %v", err)
	}
	for _, path := range migrations {
		contents, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, string(contents)); err != nil {
			t.Fatalf("migration %s: %v", path, err)
		}
	}
	return pool
}
