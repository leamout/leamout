package carrierconnections

type CarrierConnection struct {
	ID                 string
	Organization       string
	Name               string
	Provider           string
	Scope              string
	Status             string
	InboundEnabled     bool
	MaxCPS             int32
	MaxConcurrentCalls int32
	Trunks             int64
}
