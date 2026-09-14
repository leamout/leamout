package worker

const workerHealthAddress = ":8081"

var componentNames = []string{
	"freeswitch-events",
	"call-reconciliation",
	"carrier-endpoint-health",
	"recording-reconciliation",
	"outbox-publisher",
	"webhook-consumer",
	"webhook-delivery",
	"idempotency-cleanup",
}
