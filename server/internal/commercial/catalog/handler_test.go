package catalog

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func TestCatalogRoutesRejectInvalidIDs(t *testing.T) {
	router := chi.NewRouter()
	RegisterRoutes(router, NewHandler(&Service{}), func(next http.Handler) http.Handler { return next })

	for _, path := range []string{
		"/products/not-a-uuid",
		"/products/not-a-uuid/plans",
		"/plans/not-a-uuid",
		"/plans/not-a-uuid/prices",
	} {
		t.Run(path, func(t *testing.T) {
			request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != http.StatusBadRequest {
				t.Fatalf("expected status 400, got %d", response.Code)
			}
		})
	}
}

func TestWriteCatalogErrorMapsNotFoundErrors(t *testing.T) {
	for _, test := range []struct {
		name string
		err  error
	}{
		{name: "product", err: ErrProductNotFound},
		{name: "plan", err: ErrPlanNotFound},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			writeCatalogError(response, test.err)

			if response.Code != http.StatusNotFound {
				t.Fatalf("expected status 404, got %d", response.Code)
			}
		})
	}
}

func TestCatalogResponses(t *testing.T) {
	productID := uuid.New()
	planID := uuid.New()
	priceID := uuid.New()
	description := "Commercial self-hosted edition"
	effectiveFrom := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)

	values := []any{
		newProductResponse(Product{
			ID: productID, Code: "self-hosted", Name: "Self-hosted", Description: &description,
		}),
		newPlanResponse(Plan{
			ID: planID, ProductID: productID, Code: "enterprise", Name: "Enterprise", Description: &description,
		}),
		newPriceResponse(Price{
			ID: priceID, PlanID: planID, Currency: "USD", AmountMinor: 29900,
			BillingInterval: BillingIntervalMonth, EffectiveFrom: effectiveFrom,
		}),
	}

	for _, value := range values {
		encoded, err := json.Marshal(value)
		if err != nil {
			t.Fatalf("marshal response: %v", err)
		}
		var response map[string]any
		if err := json.Unmarshal(encoded, &response); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if response["id"] == "" {
			t.Fatalf("expected response ID, got %s", encoded)
		}
		if _, exposed := response["active"]; exposed {
			t.Fatalf("response must not expose internal availability: %s", encoded)
		}
	}
}
