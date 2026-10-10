package auth

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/leamout/leamout/server/pkg/apperror"
)

func TestSendOTPWithoutDelivery(t *testing.T) {
	code, err := NewService(nil).SendOTP(context.Background(), uuid.New())
	var app *apperror.AppError
	if code != "" || !errors.As(err, &app) || app.Status != http.StatusServiceUnavailable {
		t.Fatalf("expected unavailable response without a code, got code %q and error %v", code, err)
	}
}
