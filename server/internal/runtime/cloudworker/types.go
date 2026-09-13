package worker

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/leamout/leamout/internal/integrations/freeswitch"
	natsintegration "github.com/leamout/leamout/internal/integrations/nats"
	redisintegration "github.com/leamout/leamout/internal/integrations/redis"
	"github.com/leamout/leamout/internal/modules/idempotency"
	"github.com/leamout/leamout/internal/modules/outbox"
	"github.com/leamout/leamout/internal/modules/webhooks"
	"github.com/leamout/leamout/internal/platform/logging"
	"github.com/leamout/leamout/internal/telecom/calls"
	"github.com/leamout/leamout/internal/telecom/numbers"
	"github.com/leamout/leamout/internal/telecom/recordings"
	"github.com/leamout/leamout/internal/telecom/routing"
	"github.com/leamout/leamout/internal/telecom/wholesale"
)

type Worker struct {
	db                      *pgxpool.Pool
	freeSwitch              *freeswitch.Client
	nats                    *natsintegration.Client
	redis                   *redisintegration.Client
	calls                   *calls.Consumer
	recordings              *recordings.Consumer
	callReconciliation      *calls.ReconciliationJob
	endpointHealth          *routing.EndpointHealthJob
	recordingReconciliation *recordings.ReconciliationJob
	providerOperations      *numbers.ProviderOperationJob
	commpeakCDRPolling      *wholesale.CDRPollJob
	outbox                  *outbox.PublisherJob
	webhookConsumer         *webhooks.Consumer
	webhookDelivery         *webhooks.DeliveryJob
	idempotencyCleanup      *idempotency.CleanupJob
	health                  *healthState
	logger                  *logging.Logger
	componentNames          []string
}

const workerHealthAddress = ":8081"

var componentNames = []string{
	"freeswitch-events",
	"call-reconciliation",
	"carrier-endpoint-health",
	"recording-reconciliation",
	"provider-operations",
	"commpeak-cdr-polling",
	"outbox-publisher",
	"webhook-consumer",
	"webhook-delivery",
	"idempotency-cleanup",
}
