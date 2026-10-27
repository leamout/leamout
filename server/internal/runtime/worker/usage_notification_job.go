package worker

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leamout/leamout/server/internal/commercial/limits"
	"github.com/leamout/leamout/server/internal/database/pgconv"
	"github.com/leamout/leamout/server/internal/database/sqlc"
	"github.com/leamout/leamout/server/internal/observability/logging"
	"github.com/leamout/leamout/server/internal/platform/email"
)

const usageNotificationScanInterval = 5 * time.Minute

type usageNotificationJob struct {
	pool        *pgxpool.Pool
	queries     *sqlc.Queries
	limits      *limits.Service
	email       *email.Service
	billingURL  string
	logger      *logging.Logger
	interval    time.Duration
}

func newUsageNotificationJob(
	pool *pgxpool.Pool,
	queries *sqlc.Queries,
	limitService *limits.Service,
	emailService *email.Service,
	billingURL string,
	logger *logging.Logger,
) *usageNotificationJob {
	return &usageNotificationJob{
		pool: pool, queries: queries, limits: limitService, email: emailService,
		billingURL: billingURL, logger: logger, interval: usageNotificationScanInterval,
	}
}

func (j *usageNotificationJob) Run(ctx context.Context) error {
	if j == nil || j.pool == nil || j.queries == nil || j.limits == nil || j.email == nil {
		return fmt.Errorf("usage notification dependencies are not configured")
	}
	if err := j.scan(ctx); err != nil {
		j.logger.Error(ctx, "scan usage notifications", "error", err)
	}
	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := j.scan(ctx); err != nil {
				j.logger.Error(ctx, "scan usage notifications", "error", err)
			}
		}
	}
}

func (j *usageNotificationJob) scan(ctx context.Context) error {
	recipients, err := j.queries.ListUsageNotificationRecipients(ctx)
	if err != nil {
		return fmt.Errorf("list usage notification recipients: %w", err)
	}

	type owner struct { email string }
	type organization struct {
		name string
		owners []owner
	}
	organizations := make(map[uuid.UUID]*organization)
	for _, recipient := range recipients {
		entry := organizations[recipient.OrganizationID]
		if entry == nil {
			entry = &organization{name: recipient.OrganizationName}
			organizations[recipient.OrganizationID] = entry
		}
		entry.owners = append(entry.owners, owner{email: recipient.Recipient})
	}

	for organizationID, org := range organizations {
		if err := ctx.Err(); err != nil {
			return err
		}
		usage, err := j.limits.RuntimeUsage(ctx, organizationID)
		if err != nil {
			j.logger.Error(ctx, "read organization runtime usage for notifications", "organization_id", organizationID, "error", err)
			continue
		}
		if usage.Unlimited || usage.AllowanceSeconds <= 0 {
			continue
		}

		used := usage.UsedSeconds
		allowance := usage.AllowanceSeconds
		percent := int64(math.Floor(float64(used) * 100 / float64(allowance)))
		switch {
		case used >= allowance:
			if err := j.notify(ctx, organizationID, org.name, org.owners, usage, 100, true); err != nil {
				j.logger.Error(ctx, "queue quota exhaustion notifications", "organization_id", organizationID, "error", err)
			}
		case used*100 >= allowance*80:
			if err := j.notify(ctx, organizationID, org.name, org.owners, usage, 80, false); err != nil {
				j.logger.Error(ctx, "queue usage warning notifications", "organization_id", organizationID, "error", err)
			}
		default:
			_ = percent
		}
	}
	return nil
}

func (j *usageNotificationJob) notify(
	ctx context.Context,
	organizationID uuid.UUID,
	organizationName string,
	owners []struct{ email string },
	usage limits.RuntimeUsage,
	threshold int32,
	exhausted bool,
) error {
	tx, err := j.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil { return fmt.Errorf("begin usage notification transaction: %w", err) }
	defer func() { _ = tx.Rollback(ctx) }()

	inserted, err := sqlc.New(tx).InsertUsageNotificationEvent(ctx, sqlc.InsertUsageNotificationEventParams{
		ID: uuid.New(),
		OrganizationID: organizationID,
		PeriodStart: pgconv.TimeToTimestamptz(usage.PeriodStart),
		ThresholdPercent: threshold,
	})
	if err != nil { return fmt.Errorf("insert usage notification deduplication key: %w", err) }
	if inserted == 0 {
		return tx.Commit(ctx)
	}

	usedMinutes := int64(math.Ceil(float64(usage.UsedSeconds) / 60))
	allowanceMinutes := int64(math.Ceil(float64(usage.AllowanceSeconds) / 60))
	percent := int64(math.Floor(float64(usage.UsedSeconds) * 100 / float64(usage.AllowanceSeconds)))
	if exhausted { percent = 100 } else if percent >= 100 { percent = 99 }
	template := "usage-allowance-warning"
	if exhausted { template = "usage-allowance-exhausted" }
	for _, owner := range owners {
		_, err := j.email.QueueTx(ctx, tx, email.Request{
			To: owner.email,
			Template: template,
			Data: email.Data{
				ExpiresAt: time.Now().UTC().Add(30 * 24 * time.Hour),
				Organization: organizationName,
				BillingURL: j.billingURL,
				UsedMinutes: usedMinutes,
				AllowanceMinutes: allowanceMinutes,
				UsagePercent: percent,
				PeriodEnd: usage.PeriodEnd,
			},
		})
		if err != nil { return fmt.Errorf("queue %s for owner: %w", template, err) }
	}
	return tx.Commit(ctx)
}
