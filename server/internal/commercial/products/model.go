package products

// Product groups plans that deliver the same commercial offering.
type Product struct {
	ID          string
	Code        string
	Name        string
	Description string
	Active      bool
}
