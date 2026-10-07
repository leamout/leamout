package telephony

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/leamout/leamout/server/internal/database/sqlc"
	"github.com/leamout/leamout/server/internal/platform/metrics"
	"github.com/leamout/leamout/server/internal/runtime/calling"
	"github.com/leamout/leamout/server/internal/security/encryption"
	"github.com/leamout/leamout/server/internal/telephony/calls"
	"github.com/leamout/leamout/server/internal/telephony/numbers"
	"github.com/leamout/leamout/server/internal/telephony/recordings"
	"github.com/leamout/leamout/server/internal/telephony/routing"
	"github.com/leamout/leamout/server/internal/telephony/trunks"
	"github.com/leamout/leamout/server/internal/telephony/webrtc"
)

type Dependencies struct {
	DB                *pgxpool.Pool
	Queries           *sqlc.Queries
	CallsController   *calling.Controller
	CallsChannelStore *calling.ChannelStore
	CallsAdmission    *calling.AdmissionLimiter
	CredentialCipher  *encryption.Cipher
	WebRTCService     *webrtc.Service
	RecordingStorage  recordings.Storage
	Metrics           *metrics.Registry
}

type Module struct {
	Calls      CallsModule
	Numbers    NumbersModule
	WebRTC     WebRTCModule
	Recordings RecordingsModule
	Routing    RoutingModule
	Trunks     TrunksModule
}

type CallsModule struct {
	Repository *calls.Repository
	Service    *calls.Service
	Handler    *calls.Handler
}

type NumbersModule struct {
	Repository *numbers.Repository
	Service    *numbers.Service
	Handler    *numbers.Handler
}

type WebRTCModule struct {
	Service *webrtc.Service
	Handler *webrtc.Handler
}

type RecordingsModule struct {
	Repository *recordings.Repository
	Service    *recordings.Service
	Handler    *recordings.Handler
}

type RoutingModule struct {
	Repository *routing.Repository
	Service    *routing.Service
}

type TrunksModule struct {
	Repository *trunks.Repository
	Service    *trunks.Service
	Handler    *trunks.Handler
}

func New(deps Dependencies) (*Module, error) {
	routingRepository := routing.NewRepository(deps.Queries, deps.DB)
	routingService := routing.NewService(routingRepository, nil)

	callsRepository := calls.NewRepository(deps.Queries, deps.DB)
	callsService := calls.NewService(
		callsRepository,
		routingService,
		deps.CallsController,
		deps.CallsChannelStore,
		deps.CallsAdmission,
		deps.Metrics,
	)

	numbersRepository := numbers.NewRepository(deps.Queries)
	numbersService := numbers.NewService(numbersRepository)

	recordingsRepository := recordings.NewRepository(deps.DB)
	recordingsService := recordings.NewService(
		recordingsRepository,
		deps.RecordingStorage,
	)

	trunksRepository := trunks.NewRepository(deps.Queries)
	trunksService := trunks.NewService(
		trunksRepository,
		deps.DB,
		deps.CredentialCipher,
	)

	return &Module{
		Calls: CallsModule{
			Repository: callsRepository,
			Service:    callsService,
			Handler:    calls.NewHandler(callsService),
		},
		Numbers: NumbersModule{
			Repository: numbersRepository,
			Service:    numbersService,
			Handler:    numbers.NewHandler(numbersService),
		},
		Routing: RoutingModule{
			Repository: routingRepository,
			Service:    routingService,
		},
		Recordings: RecordingsModule{
			Repository: recordingsRepository,
			Service:    recordingsService,
			Handler:    recordings.NewHandler(recordingsService),
		},
		Trunks: TrunksModule{
			Repository: trunksRepository,
			Service:    trunksService,
			Handler:    trunks.NewHandler(trunksService),
		},
		WebRTC: WebRTCModule{
			Service: deps.WebRTCService,
			Handler: webrtc.NewHandler(deps.WebRTCService),
		},
	}, nil
}
