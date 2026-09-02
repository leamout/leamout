package catalog

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type stubCatalogReader struct {
	products []Product
	product  Product
	plans    []Plan
	plan     Plan
	prices   []Price
	err      error
}

func (s *stubCatalogReader) GetProduct(context.Context, uuid.UUID) (Product, error) {
	return s.product, s.err
}

func (s *stubCatalogReader) ListProducts(context.Context, bool) ([]Product, error) {
	return s.products, s.err
}

func (s *stubCatalogReader) GetPlan(context.Context, uuid.UUID) (Plan, error) {
	return s.plan, s.err
}

func (s *stubCatalogReader) ListPlans(context.Context, uuid.UUID, bool) ([]Plan, error) {
	return s.plans, s.err
}

func (s *stubCatalogReader) ListPrices(context.Context, uuid.UUID, bool) ([]Price, error) {
	return s.prices, s.err
}

func TestCatalogRoutes(t *testing.T) {
	productID := uuid.New()
	planID := uuid.New()
	priceID := uuid.New()
	now := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)
	service := &stubCatalogReader{
		products: []Product{{ID: productID, Code: "self-hosted", Name: "Self-hosted", Active: true}},
		product:  Product{ID: productID, Code: "self-hosted", Name: "Self-hosted", Active: true},
		plans:    []Plan{{ID: planID, ProductID: productID, Code: "enterprise", Name: "Enterprise", Active: true}},
		plan:     Plan{ID: planID, ProductID: productID, Code: "enterprise", Name: "Enterprise", Active: true},
		prices: []Price{{
			ID: priceID, PlanID: planID, Currency: "USD", AmountMinor: 29900,
			BillingInterval: BillingIntervalMonth, Active: true, EffectiveFrom: now,
		}},
	}
	router := chi.NewRouter()
	RegisterRoutes(router, NewHandler(service), func(next http.Handler) http.Handler { return next })

	tests := []struct {
		name       string
		path       string
		dataKey    string
		collection bool
	}{
		{name: "list products", path: "/products", dataKey: "products", collection: true},
		{name: "get product", path: "/products/" + productID.String()},
		{name: "list plans", path: "/products/" + productID.String() + "/plans", dataKey: "plans", collection: true},
		{name: "get plan", path: "/plans/" + planID.String()},
		{name: "list prices", path: "/plans/" + planID.String() + "/prices", dataKey: "prices", collection: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, test.path, nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d: %s", response.Code, response.Body.String())
			}
			var body struct {
				Success bool           `json:"success"`
				Data    map[string]any `json:"data"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if !body.Success {
				t.Fatal("expected successful response")
			}
			if test.collection {
				if _, ok := body.Data[test.dataKey].([]any); !ok {
					t.Fatalf("expected %q collection, got %#v", test.dataKey, body.Data[test.dataKey])
				}
			}
		})
	}
}

func TestCatalogRoutesRejectInvalidIDs(t *testing.T) {
	router := chi.NewRouter()
	RegisterRoutes(router, NewHandler(&stubCatalogReader{}), func(next http.Handler) http.Handler { return next })

	for _, path := range []string{"/products/not-a-uuid", "/products/not-a-uuid/plans", "/plans/not-a-uuid", "/plans/not-a-uuid/prices"} {
		request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusBadRequest {
			t.Errorf("%s: expected status 400, got %d", path, response.Code)
		}
	}
}

func TestCatalogRoutesMapNotFoundErrors(t *testing.T) {
	productID := uuid.New()
	service := &stubCatalogReader{err: ErrProductNotFound}
	router := chi.NewRouter()
	RegisterRoutes(router, NewHandler(service), func(next http.Handler) http.Handler { return next })

	request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/products/"+productID.String(), nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", response.Code)
	}
}

func TestGetPlanHidesInactiveCatalog(t *testing.T) {
	productID := uuid.New()
	planID := uuid.New()
	service := &stubCatalogReader{
		product: Product{ID: productID, Active: true},
		plan:    Plan{ID: planID, ProductID: productID, Active: false},
	}
	router := chi.NewRouter()
	RegisterRoutes(router, NewHandler(service), func(next http.Handler) http.Handler { return next })

	request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/plans/"+planID.String(), nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", response.Code)
	}
}

func TestCatalogRoutesReturnInternalErrors(t *testing.T) {
	service := &stubCatalogReader{err: errors.New("database unavailable")}
	router := chi.NewRouter()
	RegisterRoutes(router, NewHandler(service), func(next http.Handler) http.Handler { return next })

	request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/products", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", response.Code)
	}
}
