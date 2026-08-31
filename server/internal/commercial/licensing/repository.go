package licensing

import "context"

// Repository persists licenses and deployments within an organization boundary.
type Repository interface {
	GetLicense(context.Context, string, string) (License, error)
	CreateLicense(context.Context, License) (License, error)
	UpdateLicense(context.Context, License) (License, error)
	ListDeployments(context.Context, string, string) ([]Deployment, error)
}
