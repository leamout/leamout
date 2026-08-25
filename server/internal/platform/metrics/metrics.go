package metrics

// Metrics is the process-level metrics registry used by the runtime.
//
// Keep this type small and infrastructure-focused. Domain packages should
// record business metrics through explicit methods rather than depending on a
// metrics implementation or exposition format.
type Metrics = Registry

// New creates an empty metrics registry.
func New() *Metrics {
	return &Metrics{}
}
