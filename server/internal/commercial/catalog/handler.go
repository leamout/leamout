package catalog

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/leamout/leamout/pkg/apperror"
	"github.com/leamout/leamout/pkg/httputil"
)

// Handler exposes the customer-facing, read-only commercial catalog.
type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.service.ListProducts(r.Context(), true)
	if err != nil {
		writeCatalogError(w, err)
		return
	}

	responses := make([]productResponse, 0, len(products))
	for _, product := range products {
		responses = append(responses, newProductResponse(product))
	}
	httputil.OK(w, map[string]any{"products": responses})
}

func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	id, err := catalogID(r, "product_id")
	if err != nil {
		writeCatalogError(w, err)
		return
	}

	product, err := h.service.GetProduct(r.Context(), id)
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	if !product.Active {
		writeCatalogError(w, ErrProductNotFound)
		return
	}

	httputil.OK(w, newProductResponse(product))
}

func (h *Handler) ListPlans(w http.ResponseWriter, r *http.Request) {
	productID, err := catalogID(r, "product_id")
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	product, err := h.service.GetProduct(r.Context(), productID)
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	if !product.Active {
		writeCatalogError(w, ErrProductNotFound)
		return
	}

	plans, err := h.service.ListPlans(r.Context(), productID, true)
	if err != nil {
		writeCatalogError(w, err)
		return
	}

	responses := make([]planResponse, 0, len(plans))
	for _, plan := range plans {
		responses = append(responses, newPlanResponse(plan))
	}
	httputil.OK(w, map[string]any{"plans": responses})
}

func (h *Handler) GetPlan(w http.ResponseWriter, r *http.Request) {
	id, err := catalogID(r, "plan_id")
	if err != nil {
		writeCatalogError(w, err)
		return
	}

	plan, err := h.service.GetPlan(r.Context(), id)
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	if !plan.Active {
		writeCatalogError(w, ErrPlanNotFound)
		return
	}
	product, err := h.service.GetProduct(r.Context(), plan.ProductID)
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	if !product.Active {
		writeCatalogError(w, ErrPlanNotFound)
		return
	}

	httputil.OK(w, newPlanResponse(plan))
}

func (h *Handler) ListPrices(w http.ResponseWriter, r *http.Request) {
	planID, err := catalogID(r, "plan_id")
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	plan, err := h.service.GetPlan(r.Context(), planID)
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	if !plan.Active {
		writeCatalogError(w, ErrPlanNotFound)
		return
	}
	product, err := h.service.GetProduct(r.Context(), plan.ProductID)
	if err != nil {
		writeCatalogError(w, err)
		return
	}
	if !product.Active {
		writeCatalogError(w, ErrPlanNotFound)
		return
	}

	prices, err := h.service.ListPrices(r.Context(), planID, true)
	if err != nil {
		writeCatalogError(w, err)
		return
	}

	responses := make([]priceResponse, 0, len(prices))
	for _, price := range prices {
		responses = append(responses, newPriceResponse(price))
	}
	httputil.OK(w, map[string]any{"prices": responses})
}

func catalogID(r *http.Request, name string) (uuid.UUID, error) {
	id, err := uuid.Parse(chi.URLParam(r, name))
	if err != nil {
		return uuid.Nil, apperror.NewBadRequest("invalid " + name)
	}
	return id, nil
}

func writeCatalogError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrProductNotFound):
		httputil.Error(w, apperror.NewNotFound("catalog product not found"))
	case errors.Is(err, ErrPlanNotFound):
		httputil.Error(w, apperror.NewNotFound("catalog plan not found"))
	default:
		httputil.Error(w, err)
	}
}
