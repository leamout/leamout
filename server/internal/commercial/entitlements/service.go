package entitlements

import "context"

// Resolver computes the effective entitlements for a commercial customer.
type Resolver interface {
	Resolve(context.Context, string) (EntitlementSet, error)
}

// Service provides access to effective customer entitlements.
type Service struct {
	resolver Resolver
}

func NewService(resolver Resolver) *Service {
	return &Service{resolver: resolver}
}

func (s *Service) Resolve(ctx context.Context, customerID string) (EntitlementSet, error) {
	return s.resolver.Resolve(ctx, customerID)
}
