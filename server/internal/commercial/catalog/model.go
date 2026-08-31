package catalog

// Product is a commercial product family that groups reusable plans.
type Product struct {
	ID     string
	Name   string
	Active bool
}

// Plan is a reusable commercial offer within a product.
type Plan struct {
	ID        string
	ProductID string
	Name      string
	Active    bool
}
