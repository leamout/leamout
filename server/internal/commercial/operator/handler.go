// Package operator exposes the deliberately privileged HTTP surface for
// administering Leamout-owned commercial state.
package operator

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/commercial/catalog"
	"github.com/leamout/leamout/internal/commercial/entitlements"
	"github.com/leamout/leamout/internal/commercial/licensing"
	commercialstate "github.com/leamout/leamout/internal/commercial/state"
	"github.com/leamout/leamout/internal/commercial/subscriptions"
	"github.com/leamout/leamout/pkg/apperror"
	"github.com/leamout/leamout/pkg/helper"
	"github.com/leamout/leamout/pkg/httputil"
)

type Handler struct {
	catalog       *catalog.Service
	subscriptions *subscriptions.Service
	entitlements  *entitlements.Service
	state         *commercialstate.Service
	licensing     *licensing.Service
}

func NewHandler(c *catalog.Service, s *subscriptions.Service, e *entitlements.Service, state *commercialstate.Service, l *licensing.Service) *Handler {
	return &Handler{catalog: c, subscriptions: s, entitlements: e, state: state, licensing: l}
}

func (h *Handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	items, err := h.catalog.ListProducts(r.Context(), activeOnly(r))
	respond(w, map[string]any{"products": items}, err)
}

func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "product_id")
	if err == nil {
		var item catalog.Product
		item, err = h.catalog.GetProduct(r.Context(), id)
		respond(w, item, err)
		return
	}
	respond(w, nil, err)
}

func (h *Handler) ListPlans(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "product_id")
	if err == nil {
		var items []catalog.Plan
		items, err = h.catalog.ListPlans(r.Context(), id, activeOnly(r))
		respond(w, map[string]any{"plans": items}, err)
		return
	}
	respond(w, nil, err)
}

func (h *Handler) GetPlan(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "plan_id")
	if err == nil {
		var item catalog.Plan
		item, err = h.catalog.GetPlan(r.Context(), id)
		respond(w, item, err)
		return
	}
	respond(w, nil, err)
}

func (h *Handler) ListPrices(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "plan_id")
	if err == nil {
		var items []catalog.Price
		items, err = h.catalog.ListPrices(r.Context(), id, activeOnly(r))
		respond(w, map[string]any{"prices": items}, err)
		return
	}
	respond(w, nil, err)
}

func (h *Handler) GetPrice(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "price_id")
	if err == nil {
		var item catalog.Price
		item, err = h.catalog.GetPrice(r.Context(), id)
		respond(w, item, err)
		return
	}
	respond(w, nil, err)
}

func (h *Handler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	organizationID, err := pathID(r, "organization_id")
	if err != nil {
		respond(w, nil, err)
		return
	}
	input, err := helper.DecodeJSON[subscriptions.CreateInput](r)
	if err != nil {
		respond(w, nil, err)
		return
	}
	item, err := h.subscriptions.Create(r.Context(), organizationID, input)
	respondCreated(w, item, err)
}

func (h *Handler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	organizationID, err := pathID(r, "organization_id")
	if err == nil {
		var items []subscriptions.Subscription
		items, err = h.subscriptions.List(r.Context(), organizationID)
		respond(w, map[string]any{"subscriptions": items}, err)
		return
	}
	respond(w, nil, err)
}

func (h *Handler) CurrentSubscription(w http.ResponseWriter, r *http.Request) {
	organizationID, err := pathID(r, "organization_id")
	if err == nil {
		var item subscriptions.Subscription
		item, err = h.subscriptions.Current(r.Context(), organizationID)
		respond(w, item, err)
		return
	}
	respond(w, nil, err)
}

func (h *Handler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	organizationID, subscriptionID, err := organizationResourceIDs(r, "subscription_id")
	if err == nil {
		var item subscriptions.Subscription
		item, err = h.subscriptions.Get(r.Context(), organizationID, subscriptionID)
		respond(w, item, err)
		return
	}
	respond(w, nil, err)
}

type statusRequest[T ~string] struct {
	Status T `json:"status"`
}
type priceRequest struct {
	PriceID uuid.UUID `json:"price_id"`
}

func (h *Handler) TransitionSubscription(w http.ResponseWriter, r *http.Request) {
	organizationID, subscriptionID, err := organizationResourceIDs(r, "subscription_id")
	if err != nil {
		respond(w, nil, err)
		return
	}
	input, err := helper.DecodeJSON[statusRequest[subscriptions.Status]](r)
	if err != nil {
		respond(w, nil, err)
		return
	}
	item, err := h.subscriptions.Transition(r.Context(), organizationID, subscriptionID, input.Status)
	respond(w, item, err)
}

func (h *Handler) ChangeSubscriptionPrice(w http.ResponseWriter, r *http.Request) {
	organizationID, subscriptionID, err := organizationResourceIDs(r, "subscription_id")
	if err != nil {
		respond(w, nil, err)
		return
	}
	input, err := helper.DecodeJSON[priceRequest](r)
	if err != nil {
		respond(w, nil, err)
		return
	}
	item, err := h.subscriptions.ChangePrice(r.Context(), organizationID, subscriptionID, input.PriceID)
	respond(w, item, err)
}

func (h *Handler) UpdateSubscriptionPeriod(w http.ResponseWriter, r *http.Request) {
	organizationID, subscriptionID, err := organizationResourceIDs(r, "subscription_id")
	if err != nil {
		respond(w, nil, err)
		return
	}
	input, err := helper.DecodeJSON[subscriptions.PeriodUpdate](r)
	if err != nil {
		respond(w, nil, err)
		return
	}
	item, err := h.subscriptions.UpdatePeriod(r.Context(), organizationID, subscriptionID, input)
	respond(w, item, err)
}

func (h *Handler) SetSubscriptionProvider(w http.ResponseWriter, r *http.Request) {
	organizationID, subscriptionID, err := organizationResourceIDs(r, "subscription_id")
	if err != nil {
		respond(w, nil, err)
		return
	}
	input, err := helper.DecodeJSON[subscriptions.ProviderReference](r)
	if err != nil {
		respond(w, nil, err)
		return
	}
	item, err := h.subscriptions.SetProvider(r.Context(), organizationID, subscriptionID, input)
	respond(w, item, err)
}

func (h *Handler) ResolveState(w http.ResponseWriter, r *http.Request) {
	organizationID, err := pathID(r, "organization_id")
	if err == nil {
		var item commercialstate.OrganizationState
		item, err = h.state.Resolve(r.Context(), organizationID)
		respond(w, item, err)
		return
	}
	respond(w, nil, err)
}

func (h *Handler) CreatePlanEntitlement(w http.ResponseWriter, r *http.Request) {
	planID, err := pathID(r, "plan_id")
	if err != nil {
		respond(w, nil, err)
		return
	}
	input, err := helper.DecodeJSON[entitlements.CreateInput](r)
	if err != nil {
		respond(w, nil, err)
		return
	}
	item, err := h.entitlements.CreatePlan(r.Context(), planID, input)
	respondCreated(w, item, err)
}

func (h *Handler) ListPlanEntitlements(w http.ResponseWriter, r *http.Request) {
	planID, err := pathID(r, "plan_id")
	if err == nil {
		var items []entitlements.Entitlement
		items, err = h.entitlements.ListPlan(r.Context(), planID)
		respond(w, map[string]any{"entitlements": items}, err)
		return
	}
	respond(w, nil, err)
}

func (h *Handler) DeletePlanEntitlement(w http.ResponseWriter, r *http.Request) {
	planID, entitlementID, err := resourceIDs(r, "plan_id", "entitlement_id")
	if err == nil {
		err = h.entitlements.DeletePlan(r.Context(), planID, entitlementID)
	}
	respondNoContent(w, err)
}

func (h *Handler) CreateOrganizationEntitlement(w http.ResponseWriter, r *http.Request) {
	organizationID, err := pathID(r, "organization_id")
	if err != nil {
		respond(w, nil, err)
		return
	}
	input, err := helper.DecodeJSON[entitlements.CreateInput](r)
	if err != nil {
		respond(w, nil, err)
		return
	}
	item, err := h.entitlements.CreateOrganization(r.Context(), organizationID, input)
	respondCreated(w, item, err)
}

func (h *Handler) ListOrganizationEntitlements(w http.ResponseWriter, r *http.Request) {
	organizationID, err := pathID(r, "organization_id")
	if err == nil {
		var items []entitlements.Entitlement
		items, err = h.entitlements.ListOrganization(r.Context(), organizationID)
		respond(w, map[string]any{"entitlements": items}, err)
		return
	}
	respond(w, nil, err)
}

func (h *Handler) DeleteOrganizationEntitlement(w http.ResponseWriter, r *http.Request) {
	organizationID, entitlementID, err := resourceIDs(r, "organization_id", "entitlement_id")
	if err == nil {
		err = h.entitlements.DeleteOrganization(r.Context(), organizationID, entitlementID)
	}
	respondNoContent(w, err)
}

func (h *Handler) CreateLicense(w http.ResponseWriter, r *http.Request) {
	organizationID, err := pathID(r, "organization_id")
	if err != nil {
		respond(w, nil, err)
		return
	}
	input, err := helper.DecodeJSON[licensing.CreateInput](r)
	if err != nil {
		respond(w, nil, err)
		return
	}
	item, err := h.licensing.Create(r.Context(), organizationID, input)
	respondCreated(w, item, err)
}

func (h *Handler) ListLicenses(w http.ResponseWriter, r *http.Request) {
	organizationID, err := pathID(r, "organization_id")
	if err == nil {
		var items []licensing.License
		items, err = h.licensing.List(r.Context(), organizationID)
		respond(w, map[string]any{"licenses": items}, err)
		return
	}
	respond(w, nil, err)
}

func (h *Handler) GetLicense(w http.ResponseWriter, r *http.Request) {
	organizationID, licenseID, err := organizationResourceIDs(r, "license_id")
	if err == nil {
		var item licensing.License
		item, err = h.licensing.Get(r.Context(), organizationID, licenseID)
		respond(w, item, err)
		return
	}
	respond(w, nil, err)
}

func (h *Handler) TransitionLicense(w http.ResponseWriter, r *http.Request) {
	organizationID, licenseID, err := organizationResourceIDs(r, "license_id")
	if err != nil {
		respond(w, nil, err)
		return
	}
	input, err := helper.DecodeJSON[statusRequest[licensing.Status]](r)
	if err != nil {
		respond(w, nil, err)
		return
	}
	item, err := h.licensing.Transition(r.Context(), organizationID, licenseID, input.Status)
	respond(w, item, err)
}

func (h *Handler) ListDeployments(w http.ResponseWriter, r *http.Request) {
	organizationID, licenseID, err := organizationResourceIDs(r, "license_id")
	if err == nil {
		var items []licensing.Deployment
		items, err = h.licensing.ListDeployments(r.Context(), organizationID, licenseID)
		respond(w, map[string]any{"deployments": items}, err)
		return
	}
	respond(w, nil, err)
}

func activeOnly(r *http.Request) bool { return r.URL.Query().Get("active_only") != "false" }

func pathID(r *http.Request, name string) (uuid.UUID, error) {
	id, err := uuid.Parse(chi.URLParam(r, name))
	if err != nil {
		return uuid.Nil, apperror.NewBadRequest("invalid " + name)
	}
	return id, nil
}
func resourceIDs(r *http.Request, first, second string) (uuid.UUID, uuid.UUID, error) {
	a, err := pathID(r, first)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	b, err := pathID(r, second)
	return a, b, err
}
func organizationResourceIDs(r *http.Request, resource string) (uuid.UUID, uuid.UUID, error) {
	return resourceIDs(r, "organization_id", resource)
}

func respond(w http.ResponseWriter, data any, err error) {
	if err != nil {
		httputil.Error(w, operatorError(err))
		return
	}
	httputil.OK(w, data)
}
func respondCreated(w http.ResponseWriter, data any, err error) {
	if err != nil {
		httputil.Error(w, operatorError(err))
		return
	}
	httputil.Created(w, data)
}
func respondNoContent(w http.ResponseWriter, err error) {
	if err != nil {
		httputil.Error(w, operatorError(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func operatorError(err error) error {
	switch {
	case errors.Is(err, catalog.ErrProductNotFound), errors.Is(err, catalog.ErrPlanNotFound), errors.Is(err, catalog.ErrPriceNotFound),
		errors.Is(err, subscriptions.ErrSubscriptionNotFound), errors.Is(err, licensing.ErrLicenseNotFound), errors.Is(err, licensing.ErrDeploymentNotFound):
		return apperror.NewNotFound(err.Error())
	case errors.Is(err, subscriptions.ErrCurrentSubscriptionExists), errors.Is(err, subscriptions.ErrProviderConflict),
		errors.Is(err, entitlements.ErrEntitlementConflict), errors.Is(err, licensing.ErrDeploymentLimitReached), errors.Is(err, licensing.ErrActivationConflict):
		return apperror.NewConflict(err.Error())
	case errors.Is(err, subscriptions.ErrOrganizationUnavailable), errors.Is(err, subscriptions.ErrPriceUnavailable),
		errors.Is(err, entitlements.ErrScopeUnavailable), errors.Is(err, entitlements.ErrSubscriptionUnavailable),
		errors.Is(err, licensing.ErrCommercialStateUnavailable), errors.Is(err, licensing.ErrLicenseUnavailable):
		return apperror.NewUnprocessableEntity(err.Error())
	}
	// Domain validation and transition errors are safe and actionable to operators.
	if errors.Is(err, subscriptions.ErrInvalidTransition) || errors.Is(err, subscriptions.ErrInvalidStatus) ||
		errors.Is(err, subscriptions.ErrInvalidPeriod) || errors.Is(err, subscriptions.ErrTerminalSubscription) ||
		errors.Is(err, subscriptions.ErrOrganizationIDRequired) || errors.Is(err, subscriptions.ErrSubscriptionIDRequired) ||
		errors.Is(err, subscriptions.ErrPriceIDRequired) || errors.Is(err, subscriptions.ErrInvalidInitialStatus) ||
		errors.Is(err, subscriptions.ErrProviderRequired) || errors.Is(err, subscriptions.ErrProviderIDRequired) ||
		errors.Is(err, subscriptions.ErrInvalidProvider) || errors.Is(err, subscriptions.ErrNoChanges) ||
		errors.Is(err, entitlements.ErrInvalidEntitlement) || errors.Is(err, entitlements.ErrInvalidKind) ||
		errors.Is(err, entitlements.ErrInvalidPeriod) || errors.Is(err, entitlements.ErrKeyRequired) ||
		errors.Is(err, entitlements.ErrInvalidKey) || errors.Is(err, entitlements.ErrFeatureValueRequired) ||
		errors.Is(err, entitlements.ErrLimitValueRequired) || errors.Is(err, entitlements.ErrInvalidLimit) ||
		errors.Is(err, entitlements.ErrKindMismatch) || errors.Is(err, entitlements.ErrPlanIDRequired) ||
		errors.Is(err, entitlements.ErrOrganizationIDRequired) || errors.Is(err, entitlements.ErrEntitlementIDRequired) ||
		errors.Is(err, licensing.ErrInvalidTransition) || errors.Is(err, licensing.ErrInvalidStatus) ||
		errors.Is(err, licensing.ErrInvalidExpiration) || errors.Is(err, licensing.ErrSigningKeyRequired) ||
		errors.Is(err, licensing.ErrInvalidSigningKey) || errors.Is(err, licensing.ErrInvalidDeploymentLimit) ||
		errors.Is(err, licensing.ErrOrganizationIDRequired) || errors.Is(err, licensing.ErrLicenseIDRequired) {
		return apperror.NewBadRequest(err.Error())
	}
	return err
}
