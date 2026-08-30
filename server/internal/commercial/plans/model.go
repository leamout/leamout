package plans

// Plan is a versionable commercial offer belonging to a product.
// Entitlements are kept in the entitlements package so plan metadata does not
// depend on the policy evaluation implementation.
type Plan struct {
	ID          string
	ProductID   string
	Code        string
	Name        string
	Description string
	Active      bool
}
