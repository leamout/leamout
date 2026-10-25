package calls

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/server/internal/database/sqlc"
	"github.com/leamout/leamout/server/internal/runtime/calling"
	"github.com/leamout/leamout/server/internal/telephony/routing"
	"github.com/leamout/leamout/server/pkg/apperror"
)

func (s *Service) Create(ctx context.Context, organizationID uuid.UUID, req CreateRequest) (sqlc.Call, error) {
	if err := validateOrganizationID(organizationID); err != nil {
		return sqlc.Call{}, err
	}
	req, err := normalizeCreateRequest(req)
	if err != nil {
		return sqlc.Call{}, err
	}
	if req.VoiceAgentID != nil {
		if err := s.requireRuntimeAllowance(ctx, organizationID); err != nil {
			return sqlc.Call{}, err
		}
	}
	if req.VoiceAgentID != nil {
		if err := s.requireVoiceAgentReady(
			ctx,
			organizationID,
			*req.VoiceAgentID,
		); err != nil {
			return sqlc.Call{}, err
		}
	}

	call, err := s.repo.Create(ctx, organizationID, req)
	if err != nil {
		return sqlc.Call{}, apperror.NewInternal("create call", err)
	}
	if err := s.acquireOrganizationAdmission(ctx, organizationID, call.ID.String()); err != nil {
		reason := "plan_concurrent_limit"
		_, _ = s.repo.MarkFailed(ctx, organizationID, call.ID, &reason)
		return sqlc.Call{}, err
	}

	decision, err := s.router.ResolveOutbound(ctx, routing.OutboundRequest{
		OrganizationID: organizationID,
		TrunkID:        req.TrunkID,
		Destination:    req.ToURI,
	})
	if err != nil {
		reason := "route_resolution_failed"
		_, _ = s.repo.MarkFailed(ctx, organizationID, call.ID, &reason)
		_ = s.admission.ReleaseOrganization(ctx, organizationID, call.ID.String())
		return sqlc.Call{}, err
	}
	_, ok := decision.Primary()
	if !ok {
		reason := "route_resolution_failed"
		_, _ = s.repo.MarkFailed(ctx, organizationID, call.ID, &reason)
		_ = s.admission.ReleaseOrganization(ctx, organizationID, call.ID.String())
		return sqlc.Call{}, apperror.NewNotFound("no eligible outbound route")
	}

	result, selected, err := executeRoutePlan(ctx, decision.Routes, func(
		attemptCtx context.Context,
		route routing.OutboundRoute,
	) (calling.OriginateResult, error) {
		trunkID := route.TrunkID
		endpointID := route.TrunkEndpointID
		if _, attributionErr := s.repo.SetRouteAttribution(
			attemptCtx,
			organizationID,
			call.ID,
			RouteAttribution{
				TrunkID:         &trunkID,
				TrunkEndpointID: &endpointID,
			},
		); attributionErr != nil {
			return calling.OriginateResult{}, &calling.OriginateError{
				Class: calling.OriginateFailureInternal, Err: attributionErr,
			}
		}
		if admissionErr := s.admission.Acquire(
			attemptCtx, route.TrunkID, call.ID.String(), route.Limits,
		); admissionErr != nil {
			return calling.OriginateResult{}, &calling.OriginateError{
				Class: calling.OriginateFailureCapacity, Err: admissionErr,
			}
		}
		addresses, resolveErr := s.egressPolicy.ResolveHost(attemptCtx, route.Host)
		if resolveErr != nil {
			_ = s.admission.Release(attemptCtx, route.TrunkID, call.ID.String())
			return calling.OriginateResult{}, &calling.OriginateError{
				Class: calling.OriginateFailureValidation,
				Err:   fmt.Errorf("resolve public SIP trunk endpoint: %w", resolveErr),
			}
		}
		channelID := uuid.NewString()
		if err := s.repo.PrepareOrigination(attemptCtx, OriginationAttempt{
			ChannelID: channelID, CallID: call.ID, OrganizationID: organizationID,
			TrunkID: route.TrunkID, TrunkEndpointID: route.TrunkEndpointID,
		}); err != nil {
			_ = s.admission.Release(attemptCtx, route.TrunkID, call.ID.String())
			return calling.OriginateResult{}, &calling.OriginateError{Class: calling.OriginateFailureInternal, Err: err}
		}
		result, originateErr := s.controller.Originate(attemptCtx, calling.OriginateRequest{
			ChannelID: channelID,
			CallID:    call.ID, Destination: req.ToURI, CallerID: req.FromURI,
			PinnedIP: addresses[0].String(),
			TrunkID:  route.TrunkID,
			Host:     route.Host, Port: route.Port, Transport: route.Transport,
			Privacy: req.Privacy, DTMFMode: req.DTMFMode, MediaEncryption: req.MediaEncryption,
		})
		if originateErr == nil {
			if err := s.repo.MarkOriginationSubmitted(attemptCtx, channelID); err != nil {
				return result, &calling.OriginateError{Class: calling.OriginateFailureUncertain, Err: err}
			}
		} else {
			_ = s.repo.MarkOriginationUncertain(context.WithoutCancel(attemptCtx), channelID)
		}
		if originateErr != nil {
			var classified *calling.OriginateError
			if !errors.As(originateErr, &classified) ||
				classified.Class != calling.OriginateFailureUncertain {
				_ = s.admission.Release(attemptCtx, route.TrunkID, call.ID.String())
			}
		}
		return result, originateErr
	}, func(attemptCtx context.Context, attempt routeAttemptOutcome) {
		if s.metrics != nil {
			s.metrics.RouteAttempt(
				attemptCtx,
				attempt.Route.TrunkID,
				attempt.Route.TrunkEndpointID,
				attempt.Outcome,
				attempt.FailureClass,
				attempt.Attempt,
			)
			if attempt.Outcome == "succeeded" {
				s.metrics.EndpointSelection(
					attemptCtx,
					attempt.Route.TrunkID,
					attempt.Route.TrunkEndpointID,
					attempt.Attempt > 1,
				)
			}
		}
	})
	if err != nil {
		var originateErr *calling.OriginateError
		if errors.As(err, &originateErr) &&
			originateErr.Class == calling.OriginateFailureUncertain {
			// The bgapi request may already be running on FreeSWITCH. Keep
			// the call nonterminal and admission leases intact. Its stable
			// UUID allows lifecycle events and reconciliation to recover it.
			return call, apperror.NewServiceUnavailable(
				"call origination outcome is uncertain; do not retry", err,
			)
		}
		reason := "originate_failed"
		_, _ = s.repo.MarkFailed(ctx, organizationID, call.ID, &reason)
		_ = s.admission.ReleaseOrganization(ctx, organizationID, call.ID.String())
		return sqlc.Call{}, apperror.NewInternal("originate call", err)
	}
	if err := s.channels.Bind(ctx, call.ID, result.ChannelID); err != nil {
		_ = s.controller.Hangup(ctx, result.ChannelID)
		_ = s.admission.Release(ctx, selected.TrunkID, call.ID.String())
		_ = s.admission.ReleaseOrganization(ctx, organizationID, call.ID.String())
		reason := "channel_binding_failed"
		_, _ = s.repo.MarkFailed(ctx, organizationID, call.ID, &reason)
		return sqlc.Call{}, apperror.NewInternal("bind call channel", err)
	}

	return s.repo.Get(ctx, organizationID, call.ID)
}

const maxOutboundAttempts = 3

type routeAttemptOutcome struct {
	Route        routing.OutboundRoute
	Attempt      int
	Outcome      string
	FailureClass string
	SIPStatus    int
	Duration     time.Duration
	Err          error
}

type routeAttemptFunc func(context.Context, routing.OutboundRoute) (calling.OriginateResult, error)
type routeAttemptObserver func(context.Context, routeAttemptOutcome)

func executeRoutePlan(
	ctx context.Context,
	routes []routing.OutboundRoute,
	attempt routeAttemptFunc,
	observe routeAttemptObserver,
) (calling.OriginateResult, routing.OutboundRoute, error) {
	limit := min(len(routes), maxOutboundAttempts)
	var lastErr error
	var lastResult calling.OriginateResult
	var lastRoute routing.OutboundRoute
	for index := 0; index < limit; index++ {
		if err := ctx.Err(); err != nil {
			return calling.OriginateResult{}, routing.OutboundRoute{}, err
		}
		route := routes[index]
		started := time.Now()
		result, err := attempt(ctx, route)
		outcome := routeAttemptOutcome{
			Route:    route,
			Attempt:  index + 1,
			Duration: time.Since(started),
			Err:      err,
		}
		if err == nil {
			outcome.Outcome = "succeeded"
			if observe != nil {
				observe(context.WithoutCancel(ctx), outcome)
			}
			return result, route, nil
		}

		outcome.FailureClass, outcome.SIPStatus, outcome.Outcome = classifyOriginateFailure(err)
		if result.ChannelID != "" {
			// A channel may already be active even if origination reported an error.
			// Retrying on another carrier could connect and charge two calls.
			outcome.Outcome = "terminal_failure"
		}
		if observe != nil {
			observe(context.WithoutCancel(ctx), outcome)
		}
		lastErr = err
		lastResult = result
		lastRoute = route
		if outcome.Outcome != "retryable_failure" {
			break
		}
	}
	if lastErr == nil {
		lastErr = errors.New("outbound route plan is empty")
	}
	return lastResult, lastRoute, lastErr
}

func classifyOriginateFailure(err error) (failureClass string, sipStatus int, outcome string) {
	outcome = "terminal_failure"
	var originateError *calling.OriginateError
	if !errors.As(err, &originateError) {
		return string(calling.OriginateFailureInternal), 0, outcome
	}
	failureClass = string(originateError.Class)
	sipStatus = originateError.SIPStatus
	// An originate timeout does not prove that the remote carrier never
	// established a call; do not send another INVITE on an ambiguous outcome.
	if originateError.Class == calling.OriginateFailureCapacity ||
		(originateError.Class == calling.OriginateFailureSIP && sipStatus >= 500 && sipStatus <= 599) {
		outcome = "retryable_failure"
	}
	return failureClass, sipStatus, outcome
}
