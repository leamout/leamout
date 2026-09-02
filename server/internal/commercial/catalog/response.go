package catalog

import "time"

type productResponse struct {
	ID          string  `json:"id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

func newProductResponse(product Product) productResponse {
	return productResponse{
		ID:          product.ID.String(),
		Code:        product.Code,
		Name:        product.Name,
		Description: product.Description,
	}
}

type planResponse struct {
	ID          string  `json:"id"`
	ProductID   string  `json:"product_id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

func newPlanResponse(plan Plan) planResponse {
	return planResponse{
		ID:          plan.ID.String(),
		ProductID:   plan.ProductID.String(),
		Code:        plan.Code,
		Name:        plan.Name,
		Description: plan.Description,
	}
}

type priceResponse struct {
	ID              string          `json:"id"`
	PlanID          string          `json:"plan_id"`
	Currency        string          `json:"currency"`
	AmountMinor     int64           `json:"amount_minor"`
	BillingInterval BillingInterval `json:"billing_interval"`
	EffectiveFrom   time.Time       `json:"effective_from"`
	EffectiveUntil  *time.Time      `json:"effective_until,omitempty"`
}

func newPriceResponse(price Price) priceResponse {
	return priceResponse{
		ID:              price.ID.String(),
		PlanID:          price.PlanID.String(),
		Currency:        price.Currency,
		AmountMinor:     price.AmountMinor,
		BillingInterval: price.BillingInterval,
		EffectiveFrom:   price.EffectiveFrom,
		EffectiveUntil:  price.EffectiveUntil,
	}
}
