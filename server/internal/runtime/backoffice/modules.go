package backoffice

import (
	backofficecalls "github.com/leamout/leamout/internal/backoffice/calls"
	"github.com/leamout/leamout/internal/backoffice/carrierconnections"
	"github.com/leamout/leamout/internal/backoffice/commercial"
	"github.com/leamout/leamout/internal/backoffice/dashboard"
	"github.com/leamout/leamout/internal/backoffice/numbers"
	backofficeorganizations "github.com/leamout/leamout/internal/backoffice/organizations"
	"github.com/leamout/leamout/internal/backoffice/providers"
	"github.com/leamout/leamout/internal/backoffice/trunks"
)

// Modules contains the Backoffice feature handlers mounted by the HTTP runtime.
type Modules struct {
	Dashboard          *dashboard.Handler
	Organizations      *backofficeorganizations.Handler
	Calls              *backofficecalls.Handler
	Numbers            *numbers.Handler
	Trunks             *trunks.Handler
	CarrierConnections *carrierconnections.Handler
	Providers          *providers.Handler
	Commercial         *commercial.Handler
}

func newModules() Modules {
	return Modules{
		Dashboard:          dashboard.NewHandler(),
		Organizations:      backofficeorganizations.NewHandler(),
		Calls:              backofficecalls.NewHandler(),
		Numbers:            numbers.NewHandler(),
		Trunks:             trunks.NewHandler(),
		CarrierConnections: carrierconnections.NewHandler(),
		Providers:          providers.NewHandler(),
		Commercial:         commercial.NewHandler(),
	}
}
