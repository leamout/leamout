package trunks

import (
	"net/http"

	"github.com/leamout/leamout/server/pkg/listquery"
)

type ListRequest struct {
	Status         *string
	Direction      *string
	InboundEnabled *bool
}

func parseFilters(r *http.Request) (ListRequest, error) {
	p := listquery.Parser{Values: r.URL.Query()}
	req := ListRequest{
		Status:         p.Text("status"),
		Direction:      p.Text("direction"),
		InboundEnabled: p.Bool("inbound_enabled"),
	}
	return req, p.Err
}

func validateFilters(req ListRequest) error {
	if err := listquery.Enum("status", req.Status, "active", "disabled"); err != nil {
		return err
	}
	if err := listquery.Enum("direction", req.Direction, "inbound", "outbound", "bidirectional"); err != nil {
		return err
	}
	return nil
}
