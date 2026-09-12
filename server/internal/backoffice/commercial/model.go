package commercial

type Account struct {
	OrganizationID string
	Organization   string
	BillingModel   string
	WalletCount    string
	Currencies     string
}

type Detail struct {
	Account
	OrganizationCreatedAt string
}
