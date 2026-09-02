package operator

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/leamout/leamout/internal/commercial/catalog"
	"github.com/leamout/leamout/internal/commercial/entitlements"
	"github.com/leamout/leamout/internal/commercial/subscriptions"
	"github.com/leamout/leamout/pkg/apperror"
)

func TestOperatorError(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "not found", err: catalog.ErrPlanNotFound, want: http.StatusNotFound},
		{name: "conflict", err: entitlements.ErrEntitlementConflict, want: http.StatusConflict},
		{name: "unprocessable", err: subscriptions.ErrPriceUnavailable, want: http.StatusUnprocessableEntity},
		{name: "validation", err: subscriptions.ErrInvalidStatus, want: http.StatusBadRequest},
		{name: "required field", err: subscriptions.ErrPriceIDRequired, want: http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var appErr *apperror.AppError
			if !errors.As(operatorError(tt.err), &appErr) || appErr.Status != tt.want {
				t.Fatalf("operatorError(%v) did not produce status %d", tt.err, tt.want)
			}
		})
	}
}

func TestCommercialModelsUseAPINames(t *testing.T) {
	t.Parallel()
	payload, err := json.Marshal(subscriptions.CreateInput{})
	if err != nil {
		t.Fatal(err)
	}
	if string(payload) != `{"price_id":"00000000-0000-0000-0000-000000000000"}` {
		t.Fatalf("unexpected JSON field names: %s", payload)
	}
}
