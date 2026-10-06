// Package ses implements transactional delivery using Amazon SES v2.
package ses

import (
	"context"
	"fmt"
	"net/mail"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
)

type Config struct {
	Region           string
	From             string
	AccessKey        string
	SecretKey        string
	ConfigurationSet string
}

func New(ctx context.Context, cfg Config) (*Sender, error) {
	if strings.TrimSpace(cfg.Region) == "" {
		return nil, fmt.Errorf("AWS_REGION is required for SES")
	}
	if _, err := mail.ParseAddress(cfg.From); err != nil {
		return nil, fmt.Errorf("FROM_EMAIL must be a valid sender")
	}
	options := []func(*config.LoadOptions) error{
		config.WithRegion(cfg.Region),
		config.WithRetryMaxAttempts(1),
	}
	if cfg.AccessKey != "" || cfg.SecretKey != "" {
		if cfg.AccessKey == "" || cfg.SecretKey == "" {
			return nil, fmt.Errorf("AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY must both be set")
		}
		options = append(options, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		))
	}
	awsConfig, err := config.LoadDefaultConfig(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("load SES AWS configuration: %w", err)
	}
	return NewSender(sesv2.NewFromConfig(awsConfig), cfg), nil
}
