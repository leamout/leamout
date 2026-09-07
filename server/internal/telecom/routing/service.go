package routing

import "context"

type Service struct {
	resolver *Resolver
}

func NewService(resolver *Resolver) *Service {
	return &Service{resolver: resolver}
}

func (s *Service) ResolveOutbound(
	ctx context.Context,
	req OutboundRequest,
) (OutboundDecision, error) {
	if err := validateOutboundRequest(req); err != nil {
		return OutboundDecision{}, err
	}
	return s.resolver.resolveOutbound(ctx, req)
}

func (s *Service) ResolveInbound(
	ctx context.Context,
	req InboundRequest,
) (InboundDecision, error) {
	sourceIP, err := validateInboundRequest(req)
	if err != nil {
		return InboundDecision{}, err
	}
	return s.resolver.resolveInbound(ctx, req, sourceIP)
}

// ResolveManagedInboundDelivery resolves the extra hosted-edge hop used only
// by Self-Hosted + Managed connectivity. Cloud + Managed and both BYOC modes
// resolve calls in their local runtime with ResolveInbound.
func (s *Service) ResolveManagedInboundDelivery(
	ctx context.Context,
	req InboundRequest,
) (ManagedInboundDeliveryDecision, error) {
	sourceIP, err := validateInboundRequest(req)
	if err != nil {
		return ManagedInboundDeliveryDecision{}, err
	}
	return s.resolver.resolveManagedInboundDelivery(ctx, req, sourceIP)
}
