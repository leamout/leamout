package server

import (
	"strings"

	"github.com/leamout/leamout/internal/integrations/carriers/didww"
	redisintegration "github.com/leamout/leamout/internal/integrations/redis"
	"github.com/leamout/leamout/internal/platform/config"
	"github.com/leamout/leamout/internal/telecom/numbers"
)

func configureManagedNumberAcquisition(
	cfg config.Config,
	service *numbers.Service,
	redisClient *redisintegration.Client,
) error {
	if strings.TrimSpace(cfg.DIDWWAPIKey) == "" {
		return nil
	}

	client, err := didww.NewClient(didww.Config{
		BaseURL: cfg.DIDWWBaseURL,
		APIKey:  cfg.DIDWWAPIKey,
	})
	if err != nil {
		return err
	}

	service.SetManagedAcquisition(client, numbers.NewRedisSelectionStore(redisClient))
	return nil
}
