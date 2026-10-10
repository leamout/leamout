package billing

type CheckoutRequest struct {
	PlanCode   string `json:"plan_code"`
	SuccessURL string `json:"success_url"`
	CancelURL  string `json:"cancel_url"`
}

type PortalRequest struct {
	ReturnURL string `json:"return_url"`
}

type SessionResponse struct {
	URL string `json:"url"`
}
