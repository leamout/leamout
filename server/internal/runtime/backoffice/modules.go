package backoffice

import (
	backofficecalls "github.com/leamout/leamout/internal/backoffice/calls"
	"github.com/leamout/leamout/internal/backoffice/carrierconnections"
	"github.com/leamout/leamout/internal/backoffice/commercial"
	backofficeconferences "github.com/leamout/leamout/internal/backoffice/conferences"
	"github.com/leamout/leamout/internal/backoffice/dashboard"
	"github.com/leamout/leamout/internal/backoffice/numbers"
	backofficeorganizations "github.com/leamout/leamout/internal/backoffice/organizations"
	"github.com/leamout/leamout/internal/backoffice/providercdrs"
	"github.com/leamout/leamout/internal/backoffice/providers"
	backofficerecordings "github.com/leamout/leamout/internal/backoffice/recordings"
	"github.com/leamout/leamout/internal/backoffice/sipdomains"
	"github.com/leamout/leamout/internal/backoffice/subscribers"
	"github.com/leamout/leamout/internal/backoffice/trunks"
	backofficeusers "github.com/leamout/leamout/internal/backoffice/users"
	"github.com/leamout/leamout/internal/backoffice/voiceapplications"
	"github.com/leamout/leamout/internal/database/sqlc"
)

// Modules contains the Backoffice feature handlers mounted by the HTTP runtime.
type Modules struct {
	Dashboard          *dashboard.Handler
	Users              *backofficeusers.Handler
	Organizations      *backofficeorganizations.Handler
	Calls              *backofficecalls.Handler
	Numbers            *numbers.Handler
	Trunks             *trunks.Handler
	CarrierConnections *carrierconnections.Handler
	Providers          *providers.Handler
	Commercial         *commercial.Handler
	Subscribers        *subscribers.Handler
	SIPDomains         *sipdomains.Handler
	VoiceApplications  *voiceapplications.Handler
	Conferences        *backofficeconferences.Handler
	Recordings         *backofficerecordings.Handler
	ProviderCDRs       *providercdrs.Handler
}

func newModules(queries *sqlc.Queries) Modules {
	usersRepository := backofficeusers.NewRepository(queries)
	organizationsRepository := backofficeorganizations.NewRepository(queries)
	callsRepository := backofficecalls.NewRepository(queries)
	numbersRepository := numbers.NewRepository(queries)
	trunksRepository := trunks.NewRepository(queries)
	carrierConnectionsRepository := carrierconnections.NewRepository(queries)
	providersRepository := providers.NewRepository(queries)
	commercialRepository := commercial.NewRepository(queries)
	subscribersRepository := subscribers.NewRepository(queries)
	sipDomainsRepository := sipdomains.NewRepository(queries)
	voiceApplicationsRepository := voiceapplications.NewRepository(queries)
	conferencesRepository := backofficeconferences.NewRepository(queries)
	recordingsRepository := backofficerecordings.NewRepository(queries)
	providerCDRsRepository := providercdrs.NewRepository(queries)

	return Modules{
		Dashboard:          dashboard.NewHandler(),
		Users:              backofficeusers.NewHandler(usersRepository),
		Organizations:      backofficeorganizations.NewHandler(organizationsRepository),
		Calls:              backofficecalls.NewHandler(callsRepository),
		Numbers:            numbers.NewHandler(numbersRepository),
		Trunks:             trunks.NewHandler(trunksRepository),
		CarrierConnections: carrierconnections.NewHandler(carrierConnectionsRepository),
		Providers:          providers.NewHandler(providersRepository),
		Commercial:         commercial.NewHandler(commercialRepository),
		Subscribers:        subscribers.NewHandler(subscribersRepository),
		SIPDomains:         sipdomains.NewHandler(sipDomainsRepository),
		VoiceApplications:  voiceapplications.NewHandler(voiceApplicationsRepository),
		Conferences:        backofficeconferences.NewHandler(conferencesRepository),
		Recordings:         backofficerecordings.NewHandler(recordingsRepository),
		ProviderCDRs:       providercdrs.NewHandler(providerCDRsRepository),
	}
}
