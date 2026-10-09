package agents

import (
	"net/http"

	"github.com/leamout/leamout/server/pkg/listquery"
)

type ListRequest struct {
	Engine   *string
	Language *string
}

func parseFilters(r *http.Request) (ListRequest, error) {
	p := listquery.Parser{Values: r.URL.Query()}
	req := ListRequest{
		Engine:   p.Text("engine"),
		Language: p.Text("language"),
	}
	return req, p.Err
}

func validateFilters(req ListRequest) error {
	if err := listquery.Enum("engine", req.Engine, "composable", "realtime"); err != nil {
		return err
	}
	return nil
}
