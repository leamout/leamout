package wallets

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func TestReadRoutesAreRegistered(t *testing.T) {
	router := chi.NewRouter()
	identity := func(next http.Handler) http.Handler { return next }
	RegisterRoutes(router, &Handler{}, identity)

	for _, path := range []string{"/wallets", "/wallets/" + uuid.NewString(), "/wallets/" + uuid.NewString() + "/ledger"} {
		routeContext := chi.NewRouteContext()
		if !router.Match(routeContext, http.MethodGet, path) {
			t.Fatalf("GET %s is not registered", path)
		}
	}
	routeContext := chi.NewRouteContext()
	if router.Match(routeContext, http.MethodPost, "/wallets/"+uuid.NewString()+"/topups") {
		t.Fatal("wallet top-up initiation route is still registered")
	}
}

func TestWalletResponseIncludesPrepaidBalance(t *testing.T) {
	wallet := Wallet{
		ID: uuid.New(), OrganizationID: uuid.New(), Currency: "USD", Status: StatusActive,
		CreatedAt: time.Date(2026, 9, 11, 1, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 9, 11, 2, 0, 0, 0, time.UTC),
	}
	encoded, err := json.Marshal(newWalletResponse(wallet, &Balance{PostedMinor: 5000, ReservedMinor: 1250, AvailableMinor: 3750}))
	if err != nil {
		t.Fatalf("marshal wallet response: %v", err)
	}
	var response struct {
		ID      uuid.UUID `json:"id"`
		Balance struct {
			PostedMinor    int64 `json:"posted_minor"`
			ReservedMinor  int64 `json:"reserved_minor"`
			AvailableMinor int64 `json:"available_minor"`
		} `json:"balance"`
	}
	if err = json.Unmarshal(encoded, &response); err != nil {
		t.Fatalf("unmarshal wallet response: %v", err)
	}
	if response.ID != wallet.ID || response.Balance.PostedMinor != 5000 || response.Balance.ReservedMinor != 1250 || response.Balance.AvailableMinor != 3750 {
		t.Fatalf("unexpected wallet response: %+v", response)
	}
}
