package calls

import (
	"net/http"

	"github.com/leamout/leamout/server/pkg/listquery"
)

func parseFilters(r *http.Request) (ListRequest, error) {
	p := listquery.Parser{Values: r.URL.Query()}
	req := ListRequest{
		Direction:     p.Text("direction"),
		TrunkID:       p.UUID("trunk_id"),
		VoiceAgentID:  p.UUID("voice_agent_id"),
		CreatedFrom:   p.Time("created_from"),
		CreatedBefore: p.Time("created_before"),
	}
	return req, p.Err
}

func validateFilters(req ListRequest) error {
	if err := listquery.Enum("direction", req.Direction, "inbound", "outbound"); err != nil {
		return err
	}
	if err := listquery.ID("trunk_id", req.TrunkID); err != nil {
		return err
	}
	if err := listquery.ID("voice_agent_id", req.VoiceAgentID); err != nil {
		return err
	}
	return listquery.Range(req.CreatedFrom, req.CreatedBefore)
}
