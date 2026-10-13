package billing

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leamout/leamout/server/internal/commercial/licensing"
	"github.com/leamout/leamout/server/internal/commercial/limits"
	"github.com/leamout/leamout/server/internal/commercial/plans"
	"github.com/leamout/leamout/server/internal/commercial/subscriptions"
	"github.com/leamout/leamout/server/internal/database/sqlc"
	polarintegration "github.com/leamout/leamout/server/internal/integrations/payments/polar"
)

const (
	acceptanceSecretKey = "billing-acceptance-webhook-key-32"
	developerProductID  = "58ff657a-74d9-4342-89c4-2a3b784c6965"
	proProductID        = "da5b3818-e92e-4b09-afad-ae61895f05db"
	scaleProductID      = "a3023020-9822-4835-9509-1635c7d35f63"
)

type acceptancePolar struct {
	mu               sync.Mutex
	states           map[string]polarintegration.CustomerState
	subscriptions    map[string]polarintegration.Subscription
	failCustomerRead bool
}

func (p *acceptancePolar) RoundTrip(request *http.Request) (*http.Response, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	path := request.URL.EscapedPath()
	if strings.HasPrefix(path, "/v1/customers/external/") && strings.HasSuffix(path, "/state") {
		if p.failCustomerRead {
			p.failCustomerRead = false
			return acceptanceResponse(http.StatusInternalServerError, `{}`), nil
		}
		externalID := strings.TrimSuffix(strings.TrimPrefix(path, "/v1/customers/external/"), "/state")
		state, ok := p.states[externalID]
		if !ok {
			return acceptanceResponse(http.StatusNotFound, `{}`), nil
		}
		return acceptanceJSONResponse(http.StatusOK, state), nil
	}
	if strings.HasPrefix(path, "/v1/subscriptions/") {
		id := strings.TrimPrefix(path, "/v1/subscriptions/")
		value, ok := p.subscriptions[id]
		if !ok {
			return acceptanceResponse(http.StatusNotFound, `{}`), nil
		}
		return acceptanceJSONResponse(http.StatusOK, value), nil
	}
	return acceptanceResponse(http.StatusNotFound, `{}`), nil
}

func acceptanceJSONResponse(status int, value any) *http.Response {
	payload, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return acceptanceResponse(status, string(payload))
}

func acceptanceResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

type billingAcceptance struct {
	t       *testing.T
	ctx     context.Context
	pool    *pgxpool.Pool
	queries *sqlc.Queries
	service *Service
	handler *Handler
	worker  *WebhookJob
	polar   *acceptancePolar
	key     []byte
}

func TestBillingAcceptanceLifecycle(t *testing.T) {
	harness := newBillingAcceptance(t)
	organizationID := harness.createOrganization()
	periodStart := time.Now().UTC().Add(-time.Hour).Truncate(time.Second)
	periodEnd := periodStart.AddDate(0, 1, 0)
	subscriptionID := uuid.NewString()
	customerID := uuid.NewString()

	purchaseEvent := harness.subscriptionEvent(
		"subscription.created",
		organizationID,
		subscriptionID,
		customerID,
		developerProductID,
		"active",
		periodStart,
		periodEnd,
		false,
	)
	harness.deliverAndProcess("purchase", purchaseEvent)
	harness.assertPlan(organizationID, "developer", 3)
	harness.deliver("purchase", purchaseEvent)
	harness.assertReceiptCount("purchase", 1)

	periodEnd = periodEnd.AddDate(0, 1, 0)
	harness.deliverAndProcess("renewal", harness.subscriptionEvent(
		"subscription.cycled",
		organizationID,
		subscriptionID,
		customerID,
		developerProductID,
		"active",
		periodStart,
		periodEnd,
		false,
	))
	harness.assertPlan(organizationID, "developer", 3)

	harness.deliverAndProcess("upgrade", harness.subscriptionEvent(
		"subscription.updated",
		organizationID,
		subscriptionID,
		customerID,
		proProductID,
		"active",
		periodStart,
		periodEnd,
		false,
	))
	harness.assertPlan(organizationID, "pro", 15)

	harness.deliverAndProcess("downgrade", harness.subscriptionEvent(
		"subscription.updated",
		organizationID,
		subscriptionID,
		customerID,
		developerProductID,
		"active",
		periodStart,
		periodEnd,
		false,
	))
	harness.assertPlan(organizationID, "developer", 3)

	harness.deliverAndProcess("scheduled-cancel", harness.subscriptionEvent(
		"subscription.updated",
		organizationID,
		subscriptionID,
		customerID,
		developerProductID,
		"active",
		periodStart,
		periodEnd,
		true,
	))
	status := harness.status(organizationID)
	if status.Subscription == nil || !status.Subscription.CancelAtPeriodEnd {
		t.Fatal("scheduled cancellation was not displayed")
	}

	pastDueAt := time.Now().UTC().Truncate(time.Second)
	pastDueSubscription := harness.subscription(
		organizationID,
		subscriptionID,
		customerID,
		developerProductID,
		"past_due",
		periodStart,
		periodEnd,
		false,
	)
	pastDueSubscription.PastDueAt = &pastDueAt
	pastDueEvent := map[string]any{
		"type":        "subscription.past_due",
		"timestamp":   time.Now().UTC(),
		"api_version": polarintegration.APIVersion,
		"data":        pastDueSubscription,
	}
	harness.deliverAndProcess("payment-failed", pastDueEvent)
	status = harness.status(organizationID)
	if status.Subscription == nil || status.Subscription.GraceEndsAt == nil {
		t.Fatal("failed payment did not persist a grace deadline")
	}
	wantGraceEnd := pastDueAt.AddDate(0, 0, 3)
	if !status.Subscription.GraceEndsAt.Equal(wantGraceEnd) || status.Plan == nil || status.Plan.Code != "developer" {
		t.Fatalf("grace status = %#v, want Developer through %s", status, wantGraceEnd)
	}

	harness.deliverAndProcess("payment-recovered", harness.subscriptionEvent(
		"subscription.active",
		organizationID,
		subscriptionID,
		customerID,
		developerProductID,
		"active",
		periodStart,
		periodEnd,
		false,
	))
	status = harness.status(organizationID)
	if status.Subscription == nil || status.Subscription.PastDueAt != nil || status.Subscription.GraceEndsAt != nil {
		t.Fatal("payment recovery did not clear grace timestamps")
	}

	harness.deliverAndProcess("revocation", harness.subscriptionEvent(
		"subscription.revoked",
		organizationID,
		subscriptionID,
		customerID,
		developerProductID,
		"canceled",
		periodStart,
		periodEnd,
		false,
	))
	harness.assertPlan(organizationID, "free", 1)
}

func TestBillingAcceptanceRetryAndMissingInitialWebhook(t *testing.T) {
	harness := newBillingAcceptance(t)
	organizationID := harness.createOrganization()
	periodStart := time.Now().UTC().Add(-time.Hour).Truncate(time.Second)
	periodEnd := periodStart.AddDate(0, 1, 0)
	subscriptionID := uuid.NewString()
	customerID := uuid.NewString()
	value := harness.subscription(
		organizationID,
		subscriptionID,
		customerID,
		scaleProductID,
		"active",
		periodStart,
		periodEnd,
		false,
	)
	harness.polar.states[organizationID.String()] = polarintegration.CustomerState{
		ID:         customerID,
		ExternalID: organizationID.String(),
		ActiveSubscriptions: []polarintegration.CustomerStateSubscription{
			{
				ID:                 subscriptionID,
				ProductID:          scaleProductID,
				Status:             "active",
				CurrentPeriodStart: periodStart,
				CurrentPeriodEnd:   periodEnd,
			},
		},
	}
	harness.polar.subscriptions[subscriptionID] = value

	if err := harness.service.RecoverOrganization(harness.ctx, organizationID); err != nil {
		t.Fatal(err)
	}
	harness.assertPlan(organizationID, "scale", 40)

	retryOrganizationID := harness.createOrganization()
	harness.polar.states[retryOrganizationID.String()] = polarintegration.CustomerState{
		ID:         uuid.NewString(),
		ExternalID: retryOrganizationID.String(),
	}
	harness.polar.failCustomerRead = true
	harness.deliver("retry", map[string]any{
		"type":        "customer.state_changed",
		"timestamp":   time.Now().UTC(),
		"api_version": polarintegration.APIVersion,
		"data":        harness.polar.states[retryOrganizationID.String()],
	})
	worked, err := harness.worker.RunOnce(harness.ctx)
	if !worked || err == nil {
		t.Fatalf("first retry attempt = (%v, %v), want transient failure", worked, err)
	}
	var receiptStatus string
	var attempts int
	if err := harness.pool.QueryRow(
		harness.ctx,
		"SELECT status, attempts FROM billing_webhook_events WHERE event_id = $1",
		"retry",
	).Scan(&receiptStatus, &attempts); err != nil {
		t.Fatal(err)
	}
	if receiptStatus != "pending" || attempts != 1 {
		t.Fatalf("retry receipt = (%s, %d), want (pending, 1)", receiptStatus, attempts)
	}
	if _, err := harness.pool.Exec(
		harness.ctx,
		"UPDATE billing_webhook_events SET available_at = NOW() WHERE event_id = $1",
		"retry",
	); err != nil {
		t.Fatal(err)
	}
	if worked, err := harness.worker.RunOnce(harness.ctx); !worked || err != nil {
		t.Fatalf("second retry attempt = (%v, %v), want success", worked, err)
	}
	harness.assertReceiptStatus("retry", "processed", 2)
}

func newBillingAcceptance(t *testing.T) *billingAcceptance {
	t.Helper()
	databaseURL := os.Getenv("BILLING_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set BILLING_TEST_DATABASE_URL to run billing acceptance")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	t.Cleanup(cancel)

	adminPool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	schema := "billing_acceptance_" + strings.ReplaceAll(uuid.NewString(), "-", "")
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

	key := []byte(acceptanceSecretKey)
	secret := "whsec_" + base64.StdEncoding.EncodeToString(key)
	polarAPI := &acceptancePolar{
		states:        make(map[string]polarintegration.CustomerState),
		subscriptions: make(map[string]polarintegration.Subscription),
	}
	httpClient := &http.Client{
		Transport: polarAPI,
	}
	polarClient := polarintegration.NewWithHTTPClient(
		"acceptance-token",
		secret,
		"sandbox",
		httpClient,
	)
	licenseService, err := licensing.NewService(licensing.Config{
		Mode: licensing.Cloud,
	})
	if err != nil {
		t.Fatal(err)
	}
	queries := sqlc.New(pool)
	plansService := plans.NewService(plans.NewRepository(queries))
	subscriptionsService := subscriptions.NewService(subscriptions.NewRepository(queries))
	limitsService := limits.NewService(plansService, subscriptionsService, queries)
	limitsService.SetLicenseService(licenseService)
	billingService := NewService(
		plansService,
		subscriptionsService,
		polarClient,
		map[string]string{
			"developer": developerProductID,
			"pro":       proProductID,
			"scale":     scaleProductID,
		},
		"https://console.example.test",
	)
	billingService.SetRepository(NewRepository(queries, pool))
	billingService.SetStatusDependencies(licenseService, limitsService)
	repository := NewRepository(queries, pool)
	worker, err := NewWebhookJob(
		repository,
		billingService,
		func(context.Context, error) {},
	)
	if err != nil {
		t.Fatal(err)
	}
	return &billingAcceptance{
		t:       t,
		ctx:     ctx,
		pool:    pool,
		queries: queries,
		service: billingService,
		handler: NewHandler(billingService),
		worker:  worker,
		polar:   polarAPI,
		key:     key,
	}
}

func (h *billingAcceptance) createOrganization() uuid.UUID {
	h.t.Helper()
	id := uuid.New()
	if _, err := h.pool.Exec(
		h.ctx,
		"INSERT INTO organizations (id, name) VALUES ($1, $2)",
		id,
		"Billing acceptance",
	); err != nil {
		h.t.Fatal(err)
	}
	return id
}

func (h *billingAcceptance) subscriptionEvent(
	eventType string,
	organizationID uuid.UUID,
	subscriptionID string,
	customerID string,
	productID string,
	status string,
	periodStart time.Time,
	periodEnd time.Time,
	cancelAtPeriodEnd bool,
) map[string]any {
	return map[string]any{
		"type":        eventType,
		"timestamp":   time.Now().UTC(),
		"api_version": polarintegration.APIVersion,
		"data": h.subscription(
			organizationID,
			subscriptionID,
			customerID,
			productID,
			status,
			periodStart,
			periodEnd,
			cancelAtPeriodEnd,
		),
	}
}

func (h *billingAcceptance) subscription(
	organizationID uuid.UUID,
	subscriptionID string,
	customerID string,
	productID string,
	status string,
	periodStart time.Time,
	periodEnd time.Time,
	cancelAtPeriodEnd bool,
) polarintegration.Subscription {
	return polarintegration.Subscription{
		ID:         subscriptionID,
		CustomerID: customerID,
		ProductID:  productID,
		Customer: polarintegration.Customer{
			ID:         customerID,
			ExternalID: organizationID.String(),
		},
		Status: status,
		Metadata: map[string]any{
			"organization_id": organizationID.String(),
		},
		CurrentPeriodStart: periodStart,
		CurrentPeriodEnd:   periodEnd,
		CancelAtPeriodEnd:  cancelAtPeriodEnd,
	}
}

func (h *billingAcceptance) deliverAndProcess(eventID string, value map[string]any) {
	h.t.Helper()
	h.deliver(eventID, value)
	worked, err := h.worker.RunOnce(h.ctx)
	if !worked || err != nil {
		h.t.Fatalf("process %s = (%v, %v)", eventID, worked, err)
	}
	h.assertReceiptStatus(eventID, "processed", 1)
}

func (h *billingAcceptance) deliver(eventID string, value map[string]any) {
	h.t.Helper()
	payload, err := json.Marshal(value)
	if err != nil {
		h.t.Fatal(err)
	}
	timestamp := time.Now().Unix()
	message := fmt.Sprintf("%s.%d.%s", eventID, timestamp, payload)
	mac := hmac.New(sha256.New, h.key)
	_, _ = mac.Write([]byte(message))

	request := httptest.NewRequestWithContext(
		h.ctx,
		http.MethodPost,
		"/v1/webhooks/polar",
		bytes.NewReader(payload),
	)
	request.Header.Set("webhook-id", eventID)
	request.Header.Set("webhook-timestamp", fmt.Sprint(timestamp))
	request.Header.Set("webhook-signature", "v1,"+base64.StdEncoding.EncodeToString(mac.Sum(nil)))
	recorder := httptest.NewRecorder()
	h.handler.PolarWebhook(recorder, request)
	if recorder.Code != http.StatusAccepted {
		h.t.Fatalf("deliver %s: status %d: %s", eventID, recorder.Code, recorder.Body.String())
	}
}

func (h *billingAcceptance) status(organizationID uuid.UUID) Status {
	h.t.Helper()
	value, err := h.service.Status(h.ctx, organizationID)
	if err != nil {
		h.t.Fatal(err)
	}
	return value
}

func (h *billingAcceptance) assertPlan(
	organizationID uuid.UUID,
	code string,
	maxConcurrentCalls int64,
) {
	h.t.Helper()
	status := h.status(organizationID)
	if status.Plan == nil || status.Plan.Code != code {
		h.t.Fatalf("plan = %#v, want %s", status.Plan, code)
	}
	if status.Limits.MaxConcurrentCalls == nil || *status.Limits.MaxConcurrentCalls != maxConcurrentCalls {
		h.t.Fatalf("max concurrent calls = %v, want %d", status.Limits.MaxConcurrentCalls, maxConcurrentCalls)
	}
}

func (h *billingAcceptance) assertReceiptCount(eventID string, want int) {
	h.t.Helper()
	var count int
	if err := h.pool.QueryRow(
		h.ctx,
		"SELECT COUNT(*) FROM billing_webhook_events WHERE event_id = $1",
		eventID,
	).Scan(&count); err != nil {
		h.t.Fatal(err)
	}
	if count != want {
		h.t.Fatalf("receipt count for %s = %d, want %d", eventID, count, want)
	}
}

func (h *billingAcceptance) assertReceiptStatus(eventID, wantStatus string, wantAttempts int) {
	h.t.Helper()
	var status string
	var attempts int
	if err := h.pool.QueryRow(
		h.ctx,
		"SELECT status, attempts FROM billing_webhook_events WHERE event_id = $1",
		eventID,
	).Scan(&status, &attempts); err != nil {
		h.t.Fatal(err)
	}
	if status != wantStatus || attempts != wantAttempts {
		h.t.Fatalf("receipt %s = (%s, %d), want (%s, %d)", eventID, status, attempts, wantStatus, wantAttempts)
	}
}

var _ http.RoundTripper = (*acceptancePolar)(nil)
