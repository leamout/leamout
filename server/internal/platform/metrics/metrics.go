package metrics

import "sync/atomic"

// Registry contains process metrics that are safe to update concurrently.
//
// Keep this type small and infrastructure-focused. Domain packages should
// record business metrics through explicit methods rather than depending on a
// metrics implementation or exposition format.
type Registry struct {
	requests atomic.Uint64
	errors   atomic.Uint64
}

// New creates an empty metrics registry.
func New() *Registry {
	return &Registry{}
}

// IncRequests increments the total request counter.
func (r *Registry) IncRequests() {
	r.requests.Add(1)
}

// IncErrors increments the total error counter.
func (r *Registry) IncErrors() {
	r.errors.Add(1)
}

// Requests returns the total number of requests recorded by the registry.
func (r *Registry) Requests() uint64 {
	return r.requests.Load()
}

// Errors returns the total number of errors recorded by the registry.
func (r *Registry) Errors() uint64 {
	return r.errors.Load()
}
