package checkout

import (
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestCheckoutRoutesIncludeCreation(t *testing.T) {
	router := chi.NewRouter()
	identity := func(next http.Handler) http.Handler { return next }
	RegisterRoutes(router, &Handler{}, identity, identity)

	for method, path := range map[string]string{
		http.MethodPost: "/checkouts",
		http.MethodGet:  "/checkouts/3c8d110b-d675-4d08-a4a4-c5343edbe9cc",
	} {
		routeContext := chi.NewRouteContext()
		if !router.Match(routeContext, method, path) {
			t.Fatalf("%s %s is not registered", method, path)
		}
	}
}
