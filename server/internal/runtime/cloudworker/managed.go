package worker

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/leamout/leamout/internal/commercial"
	"github.com/leamout/leamout/internal/integrations/carriers/commpeak"
	"github.com/leamout/leamout/internal/integrations/carriers/didww"
	redisintegration "github.com/leamout/leamout/internal/integrations/redis"
	"github.com/leamout/leamout/internal/platform/config"
	"github.com/leamout/leamout/internal/telecom/numbers"
	"github.com/leamout/leamout/internal/telecom/wholesale"
)

func newManagedJobs(
	db *pgxpool.Pool,
	redisClient *redisintegration.Client,
	cfg config.Config,
) (*numbers.ProviderOperationJob, *wholesale.CDRPollJob, error) {
	commercialModule := commercial.NewCloud(db)
	numbersRepository := numbers.NewRepository(db, redisClient)
	numbersService := numbers.NewService(numbersRepository)
	numbersService.SetManagedPurchaseAuthority(commercialModule.Wallets.Service)

	if strings.TrimSpace(cfg.DIDWW.APIKey) != "" {
		didwwClient, err := didww.NewClient(didww.Config{
			BaseURL: cfg.DIDWW.APIBaseURL,
			APIKey:  cfg.DIDWW.APIKey,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("initialize DIDWW provider executor: %w", err)
		}
		numbersService.SetManagedProvider("didww", didwwClient)
	}
	providerOperations, err := numbers.NewProviderOperationJob(
		numbersRepository,
		numbersService,
		numbers.DefaultProviderOperationJobConfig(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("initialize provider operation job: %w", err)
	}

	var commpeakSource wholesale.CDRPageSource
	if strings.TrimSpace(cfg.CommPeak.Authorization) != "" {
		commpeakClient, err := commpeak.NewClient(commpeak.Config{
			BaseURL:       cfg.CommPeak.APIBaseURL,
			Authorization: cfg.CommPeak.Authorization,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("initialize CommPeak CDR client: %w", err)
		}
		commpeakSource = commpeakClient
	}
	commpeakCDRPolling, err := wholesale.NewCDRPollJob(
		wholesale.NewRepository(db),
		commpeakSource,
		wholesale.DefaultCDRPollJobConfig("commpeak"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("initialize CommPeak CDR polling job: %w", err)
	}
	return providerOperations, commpeakCDRPolling, nil
}
