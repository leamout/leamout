package checkout

import (
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestCheckoutRoutesIncludeUnifiedLifecycle(t *testing.T) {
	router := chi.NewRouter()
	identity := func(next http.Handler) http.Handler { return next }
	RegisterRoutes(router, &Handler{}, identity, identity)

	routes := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/checkouts"},
		{http.MethodGet, "/checkouts/3c8d110b-d675-4d08-a4a4-c5343edbe9cc"},
		{http.MethodPost, "/checkouts/3c8d110b-d675-4d08-a4a4-c5343edbe9cc/confirm"},
		{http.MethodPost, "/checkouts/3c8d110b-d675-4d08-a4a4-c5343edbe9cc/continue"},
	}

	for _, route := range routes {
		routeContext := chi.NewRouteContext()
		if !router.Match(routeContext, route.method, route.path) {
			t.Fatalf("%s %s is not registered", route.method, route.path)
		}
	}
}
