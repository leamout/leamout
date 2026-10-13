package media

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/leamout/leamout/server/internal/commercial/licensing"

	redisintegration "github.com/leamout/leamout/server/internal/integrations/redis"
	"github.com/leamout/leamout/server/internal/media/session"
	"github.com/leamout/leamout/server/internal/platform/logging"
	"github.com/leamout/leamout/server/internal/runtime/medianodes"
)

const mediaNodeHeartbeatInterval = 5 * time.Second

type nodeRegistration struct {
	license  *licensing.Service
	redis    *redisintegration.Client
	registry *medianodes.Registry
	cancel   context.CancelFunc
	done     chan struct{}
	nodeID   string
	draining atomic.Bool
}

func startNodeRegistration(
	ctx context.Context,
	cfg Config,
	manager *session.Manager,
	logger *logging.Logger,
) (*nodeRegistration, error) {
	if cfg.RedisURL == "" {
		return nil, nil
	}

	client, err := redisintegration.New(
		ctx,
		redisintegration.DefaultConfig(cfg.RedisURL),
	)
	if err != nil {
		return nil, err
	}

	licenseService, err := licensing.NewService(cfg.Commercial)
	if err != nil {
		_ = client.Close()
		return nil, err
	}
	registry := medianodes.NewRegistry(client)
	registrationCtx, cancel := context.WithCancel(context.Background())
	registration := &nodeRegistration{
		license:  licenseService,
		redis:    client,
		registry: registry,
		cancel:   cancel,
		done:     make(chan struct{}),
		nodeID:   cfg.NodeID,
	}

	heartbeat := func() {
		draining := registration.draining.Load()
		if registration.license.SelfHosted() && registration.requireLicenseNode(registrationCtx) != nil {
			draining = true
		}
		err := registry.Heartbeat(registrationCtx, medianodes.Node{
			ID:         cfg.NodeID,
			ControlURL: cfg.ControlURL,
			AudioURL:   cfg.PublicWebSocket,
			Capacity:   cfg.MaxSessions,
			Active:     manager.Active(),
			Draining:   draining,
		})
		if err != nil && logger != nil {
			logger.Warn(
				context.Background(),
				"heartbeat media node",
				"node_id", cfg.NodeID,
				"error", err,
			)
		}
		for _, sessionID := range manager.SessionIDs() {
			_ = registry.Refresh(registrationCtx, cfg.NodeID, sessionID)
		}
	}

	heartbeat()
	go func() {
		defer close(registration.done)
		ticker := time.NewTicker(mediaNodeHeartbeatInterval)
		defer ticker.Stop()
		for {
			select {
			case <-registrationCtx.Done():
				return
			case <-ticker.C:
				heartbeat()
			}
		}
	}()

	return registration, nil
}

func (r *nodeRegistration) beginDrain(
	ctx context.Context,
	cfg Config,
	manager *session.Manager,
) {
	if r == nil {
		return
	}
	r.draining.Store(true)
	_ = r.registry.Heartbeat(ctx, medianodes.Node{
		ID:         cfg.NodeID,
		ControlURL: cfg.ControlURL,
		AudioURL:   cfg.PublicWebSocket,
		Capacity:   cfg.MaxSessions,
		Active:     manager.Active(),
		Draining:   true,
	})
}

func (r *nodeRegistration) close(ctx context.Context) {
	if r == nil {
		return
	}
	r.cancel()
	<-r.done
	if r.license.SelfHosted() {
		_ = r.redis.ReleaseCallLease(ctx, r.licenseNodePrefix(), r.nodeID)
	}
	_ = r.registry.RemoveNode(ctx, r.nodeID)
	_ = r.redis.Close()
}

// A node is a media process with a unique stable MEDIA_NODE_ID.
// Short Redis leases reclaim capacity after crashes. Redis is shared by the deployment.
func (r *nodeRegistration) licenseNodePrefix() string {
	return "commercial:license:nodes:" + r.license.DeploymentID()
}

func (r *nodeRegistration) requireLicenseNode(ctx context.Context) error {
	claims, err := r.license.RequireRuntime()
	if err != nil {
		return err
	}
	allowed, err := r.redis.AcquireConcurrentLease(ctx, r.licenseNodePrefix(), r.nodeID, claims.MaxNodes, 15*time.Second)
	if err != nil {
		return err
	}
	if !allowed {
		return fmt.Errorf("licensed media node capacity exhausted")
	}
	return nil
}
