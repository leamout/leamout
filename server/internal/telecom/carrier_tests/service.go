package carrier_tests

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/modules/audit"
	"github.com/leamout/leamout/internal/telecom/calls"
	"github.com/leamout/leamout/internal/telecom/routing"
	"github.com/leamout/leamout/pkg/apperror"
)

const (
	testRateLimit    = int64(3)
	testRateWindow   = time.Minute
	originateTimeout = 12 * time.Second
	maximumDuration  = 5 * time.Second
)

var e164Pattern = regexp.MustCompile(`^\+[1-9][0-9]{6,14}$`)

type rateLimiter interface {
	AllowFixedWindow(context.Context, string, int64, time.Duration) (bool, error)
}

type routeResolver interface {
	ResolveOutbound(context.Context, routing.OutboundRequest) (routing.OutboundDecision, error)
}

type callController interface {
	Originate(context.Context, calls.OriginateRequest) (string, error)
	Hangup(context.Context, string) error
}

type resultStore interface {
	Create(context.Context, uuid.UUID, uuid.UUID, audit.Actor, Request) (Result, error)
	AttributeRoute(context.Context, uuid.UUID, routing.OutboundDecision) error
	Finish(context.Context, uuid.UUID, string, *string, *string, *string, bool) (Result, error)
	List(context.Context, uuid.UUID, uuid.UUID, int32, int32) ([]Result, error)
}

func (s *Service) List(ctx context.Context, organizationID, connectionID uuid.UUID, limit, offset int32) ([]Result, error) {
	if organizationID == uuid.Nil || connectionID == uuid.Nil {
		return nil, apperror.NewBadRequest("organization and carrier connection are required")
	}
	if limit < 1 || limit > 100 {
		return nil, apperror.NewBadRequest("limit must be between 1 and 100")
	}
	if offset < 0 {
		return nil, apperror.NewBadRequest("offset cannot be negative")
	}
	items, err := s.repo.List(ctx, organizationID, connectionID, limit, offset)
	if err != nil {
		return nil, apperror.NewInternal("list carrier test calls", err)
	}
	return items, nil
}

type Service struct {
	repo             resultStore
	routes           routeResolver
	calls            callController
	limiter          rateLimiter
	allowlist        map[string]struct{}
	originateTimeout time.Duration
	maximumDuration  time.Duration
}

func NewService(repo resultStore, routes routeResolver, controller callController, limiter rateLimiter, destinations []string) (*Service, error) {
	if repo == nil || routes == nil || controller == nil || limiter == nil {
		return nil, fmt.Errorf("carrier test-call dependencies are required")
	}
	allowlist := make(map[string]struct{}, len(destinations))
	for _, destination := range destinations {
		destination = strings.TrimSpace(destination)
		if destination == "" {
			continue
		}
		if !e164Pattern.MatchString(destination) {
			return nil, fmt.Errorf("carrier test-call destination %q must be E.164", destination)
		}
		allowlist[destination] = struct{}{}
	}
	return &Service{repo: repo, routes: routes, calls: controller, limiter: limiter, allowlist: allowlist, originateTimeout: originateTimeout, maximumDuration: maximumDuration}, nil
}

func (s *Service) Run(ctx context.Context, organizationID, connectionID uuid.UUID, req Request) (Result, error) {
	if organizationID == uuid.Nil || connectionID == uuid.Nil || req.TrunkID == uuid.Nil {
		return Result{}, apperror.NewBadRequest("organization, carrier connection, and trunk are required")
	}
	req.From, req.To = strings.TrimSpace(req.From), strings.TrimSpace(req.To)
	if !e164Pattern.MatchString(req.From) || !e164Pattern.MatchString(req.To) {
		return Result{}, apperror.NewBadRequest("from and to must be E.164")
	}
	if _, allowed := s.allowlist[req.To]; !allowed {
		return Result{}, apperror.NewForbidden("destination is not allowlisted for carrier test calls")
	}
	allowed, err := s.limiter.AllowFixedWindow(ctx, "ratelimit:carrier-test:"+organizationID.String()+":"+connectionID.String(), testRateLimit, testRateWindow)
	if err != nil {
		return Result{}, apperror.NewServiceUnavailable("carrier test-call rate limiter unavailable", err)
	}
	if !allowed {
		return Result{}, apperror.NewTooManyRequests("carrier test-call rate limit exceeded")
	}
	actor, err := audit.ActorFromContext(ctx)
	if err != nil {
		return Result{}, apperror.NewInternal("attribute carrier test call", err)
	}
	result, err := s.repo.Create(ctx, organizationID, connectionID, actor, req)
	if err != nil {
		return Result{}, apperror.NewInternal("persist carrier test call", err)
	}

	route, err := s.routes.ResolveOutbound(ctx, routing.OutboundRequest{OrganizationID: organizationID, From: req.From, To: req.To, TrunkID: req.TrunkID})
	if err != nil || route.CarrierConnectionID != connectionID {
		if err == nil {
			err = fmt.Errorf("trunk does not belong to requested carrier connection")
		}
		return s.finishFailure(ctx, result.ID, "failed", nil, "ROUTE_REJECTED", err)
	}
	if err := s.repo.AttributeRoute(ctx, result.ID, route); err != nil {
		return Result{}, apperror.NewInternal("persist carrier test-call route", err)
	}

	originateCtx, cancel := context.WithTimeout(ctx, s.originateTimeout)
	requestedCallID := uuid.NewString()
	callID, originateErr := s.calls.Originate(originateCtx, calls.OriginateRequest{
		CarrierConnectionID: route.CarrierConnectionID, Host: route.Host, Port: route.Port,
		Transport: route.Transport, Destination: route.To, CallerID: route.From,
		Variables: map[string]string{"origination_uuid": requestedCallID, "originate_timeout": "10", "call_timeout": "10"},
	})
	cancel()
	if originateErr != nil {
		status, code := "failed", "ORIGINATE_FAILED"
		if errors.Is(originateErr, context.DeadlineExceeded) {
			status, code = "timed_out", "ORIGINATE_TIMEOUT"
			cleanupCtx, cleanupCancel := context.WithTimeout(context.WithoutCancel(ctx), 3*time.Second)
			_ = s.calls.Hangup(cleanupCtx, requestedCallID)
			cleanupCancel()
		}
		return s.finishFailure(ctx, result.ID, status, nil, code, originateErr)
	}

	timer := time.NewTimer(s.maximumDuration)
	select {
	case <-ctx.Done():
		if !timer.Stop() {
			<-timer.C
		}
	case <-timer.C:
	}
	hangupCtx, hangupCancel := context.WithTimeout(context.WithoutCancel(ctx), 3*time.Second)
	hangupErr := s.calls.Hangup(hangupCtx, callID)
	hangupCancel()
	if hangupErr != nil {
		return s.finishFailure(ctx, result.ID, "failed", &callID, "HANGUP_FAILED", hangupErr)
	}
	code := "ANSWERED"
	completed, err := s.repo.Finish(context.WithoutCancel(ctx), result.ID, "completed", &callID, &code, nil, true)
	if err != nil {
		return Result{}, apperror.NewInternal("complete carrier test call", err)
	}
	return completed, nil
}

func (s *Service) finishFailure(ctx context.Context, id uuid.UUID, status string, callID *string, code string, cause error) (Result, error) {
	message := "unknown failure"
	if cause != nil {
		message = truncate(cause.Error(), 512)
	}
	result, err := s.repo.Finish(context.WithoutCancel(ctx), id, status, callID, &code, &message, false)
	if err != nil {
		return Result{}, apperror.NewInternal("persist carrier test-call failure", errors.Join(cause, err))
	}
	return result, nil
}

func truncate(value string, limit int) string {
	if len(value) > limit {
		return value[:limit]
	}
	return value
}
