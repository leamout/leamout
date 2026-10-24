package calling

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	redisintegration "github.com/leamout/leamout/server/internal/integrations/redis"
	"github.com/leamout/leamout/server/internal/telephony/routing"
	"time"
)

const callAdmissionLeaseTTL = 26 * time.Hour

type AdmissionLimiter struct {
	client *redisintegration.Client
}

func NewAdmissionLimiter(client *redisintegration.Client) *AdmissionLimiter {
	if client == nil {
		panic("calling: Redis admission store is required")
	}
	return &AdmissionLimiter{
		client: client,
	}
}

func (l *AdmissionLimiter) Acquire(
	ctx context.Context,
	trunkID uuid.UUID,
	leaseID string,
	limits routing.Limits,
) error {
	if trunkID == uuid.Nil || leaseID == "" {
		return fmt.Errorf("trunk id and lease id are required")
	}
	if limits.MaxCPS < 1 || limits.MaxConcurrentCalls < 1 {
		return fmt.Errorf("trunk admission limits must be positive")
	}

	allowed, reason, err := l.client.AcquireCallLease(
		ctx,
		admissionPrefix(trunkID),
		leaseID,
		int64(limits.MaxCPS),
		int64(limits.MaxConcurrentCalls),
		callAdmissionLeaseTTL,
	)
	if err != nil {
		return fmt.Errorf("acquire trunk call lease: %w", err)
	}
	if allowed {
		return nil
	}

	switch reason {
	case "cps":
		return ErrAdmissionCPS
	case "concurrent":
		return ErrAdmissionConcurrent
	default:
		return fmt.Errorf("trunk admission rejected: %s", reason)
	}
}

func (l *AdmissionLimiter) Bind(
	ctx context.Context,
	trunkID uuid.UUID,
	leaseID string,
	callID uuid.UUID,
) error {
	if trunkID == uuid.Nil || leaseID == "" || callID == uuid.Nil {
		return fmt.Errorf("trunk id, lease id, and call id are required")
	}
	return l.client.BindCallLease(
		ctx,
		admissionPrefix(trunkID),
		leaseID,
		callID.String(),
	)
}

func (l *AdmissionLimiter) Release(
	ctx context.Context,
	trunkID uuid.UUID,
	callOrLeaseID string,
) error {
	if trunkID == uuid.Nil || callOrLeaseID == "" {
		return fmt.Errorf("trunk id and call or lease id are required")
	}
	return l.client.ReleaseCallLease(
		ctx,
		admissionPrefix(trunkID),
		callOrLeaseID,
	)
}

func (l *AdmissionLimiter) Refresh(
	ctx context.Context,
	trunkID, callID uuid.UUID,
) error {
	if trunkID == uuid.Nil || callID == uuid.Nil {
		return fmt.Errorf("trunk id and call id are required")
	}
	return l.client.RefreshCallLease(
		ctx,
		admissionPrefix(trunkID),
		callID.String(),
		callAdmissionLeaseTTL,
	)
}

func (l *AdmissionLimiter) AcquireOrganization(
	ctx context.Context,
	organizationID uuid.UUID,
	leaseID string,
	maxConcurrent int64,
) error {
	if organizationID == uuid.Nil || leaseID == "" {
		return fmt.Errorf("organization id and lease id are required")
	}
	if maxConcurrent < 1 {
		return fmt.Errorf("organization concurrent call limit must be positive")
	}

	allowed, err := l.client.AcquireConcurrentLease(
		ctx,
		organizationAdmissionPrefix(organizationID),
		leaseID,
		maxConcurrent,
		callAdmissionLeaseTTL,
	)
	if err != nil {
		return fmt.Errorf("acquire organization call lease: %w", err)
	}
	if !allowed {
		return ErrPlanConcurrent
	}
	return nil
}

func (l *AdmissionLimiter) BindOrganization(
	ctx context.Context,
	organizationID uuid.UUID,
	leaseID string,
	callID uuid.UUID,
) error {
	if organizationID == uuid.Nil || leaseID == "" || callID == uuid.Nil {
		return fmt.Errorf("organization id, lease id, and call id are required")
	}
	return l.client.BindCallLease(
		ctx,
		organizationAdmissionPrefix(organizationID),
		leaseID,
		callID.String(),
	)
}

func (l *AdmissionLimiter) ReleaseOrganization(
	ctx context.Context,
	organizationID uuid.UUID,
	callOrLeaseID string,
) error {
	if organizationID == uuid.Nil || callOrLeaseID == "" {
		return fmt.Errorf("organization id and call or lease id are required")
	}
	return l.client.ReleaseCallLease(
		ctx,
		organizationAdmissionPrefix(organizationID),
		callOrLeaseID,
	)
}

func (l *AdmissionLimiter) RefreshOrganization(
	ctx context.Context,
	organizationID, callID uuid.UUID,
) error {
	if organizationID == uuid.Nil || callID == uuid.Nil {
		return fmt.Errorf("organization id and call id are required")
	}
	return l.client.RefreshCallLease(
		ctx,
		organizationAdmissionPrefix(organizationID),
		callID.String(),
		callAdmissionLeaseTTL,
	)
}
func admissionPrefix(trunkID uuid.UUID) string {
	return "telecom:admission:trunk:" + trunkID.String()
}

func organizationAdmissionPrefix(organizationID uuid.UUID) string {
	return "telecom:admission:organization:" + organizationID.String()
}
