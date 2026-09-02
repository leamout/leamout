package operator

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, handler *Handler, auth func(http.Handler) http.Handler) {
	router.Route("/commercial", func(r chi.Router) {
		r.Use(auth)
		r.Get("/products", handler.ListProducts)
		r.Get("/products/{product_id}", handler.GetProduct)
		r.Get("/products/{product_id}/plans", handler.ListPlans)
		r.Get("/plans/{plan_id}", handler.GetPlan)
		r.Get("/plans/{plan_id}/prices", handler.ListPrices)
		r.Get("/prices/{price_id}", handler.GetPrice)
		r.Post("/plans/{plan_id}/entitlements", handler.CreatePlanEntitlement)
		r.Get("/plans/{plan_id}/entitlements", handler.ListPlanEntitlements)
		r.Delete("/plans/{plan_id}/entitlements/{entitlement_id}", handler.DeletePlanEntitlement)

		r.Route("/organizations/{organization_id}", func(r chi.Router) {
			r.Post("/subscriptions", handler.CreateSubscription)
			r.Get("/subscriptions", handler.ListSubscriptions)
			r.Get("/subscriptions/current", handler.CurrentSubscription)
			r.Get("/subscriptions/{subscription_id}", handler.GetSubscription)
			r.Put("/subscriptions/{subscription_id}/status", handler.TransitionSubscription)
			r.Put("/subscriptions/{subscription_id}/price", handler.ChangeSubscriptionPrice)
			r.Put("/subscriptions/{subscription_id}/period", handler.UpdateSubscriptionPeriod)
			r.Put("/subscriptions/{subscription_id}/provider", handler.SetSubscriptionProvider)
			r.Get("/state", handler.ResolveState)
			r.Post("/entitlements", handler.CreateOrganizationEntitlement)
			r.Get("/entitlements", handler.ListOrganizationEntitlements)
			r.Delete("/entitlements/{entitlement_id}", handler.DeleteOrganizationEntitlement)
			r.Post("/licenses", handler.CreateLicense)
			r.Get("/licenses", handler.ListLicenses)
			r.Get("/licenses/{license_id}", handler.GetLicense)
			r.Put("/licenses/{license_id}/status", handler.TransitionLicense)
			r.Get("/licenses/{license_id}/deployments", handler.ListDeployments)
		})
	})
}
