package commercial

type Account struct {
	OrganizationID     string
	Organization       string
	Plan               string
	SubscriptionStatus string
	BillingModel       string
	RenewsAt           string
}

type Detail struct {
	Account
	SubscriptionID, PlanID, PriceID, PricingType, Currency, AmountMinor, BillingInterval, StartsAt, EndsAt, OrganizationCreatedAt string
}
