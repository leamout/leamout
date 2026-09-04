package commpeak

import "encoding/json"

type CDRDirection string

const (
	CDRDirectionTermination CDRDirection = "termination"
	CDRDirectionOrigination CDRDirection = "origination"
)

type CDRRequest struct {
	TimeRange    string
	TechPrefix   string
	CLI          string
	Destination  string
	SIPAccountID string
	DIDs         []string
	Page         int
	PerPage      int
}

// CDRPage intentionally preserves provider records as raw JSON until Leamout's
// provider-neutral wholesale reconciliation schema is defined. Provider
// adapters should not invent durable billing semantics ahead of that model.
type CDRPage struct {
	Direction CDRDirection
	Raw       json.RawMessage
}
