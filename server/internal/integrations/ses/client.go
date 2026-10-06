// Package ses implements transactional delivery using Amazon SES v2.
package ses

import (
	"context"
	"fmt"
	"net/mail"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
)

type Config struct{ Region, From, ReplyTo, ConfigurationSet string }

func New(ctx context.Context, cfg Config) (*Sender, error) {
	if strings.TrimSpace(cfg.Region) == "" {
		return nil, fmt.Errorf("AWS_REGION is required for SES")
	}
	if _, err := mail.ParseAddress(cfg.From); err != nil {
		return nil, fmt.Errorf("EMAIL_FROM must be a valid sender")
	}
	if cfg.ReplyTo != "" {
		if _, err := mail.ParseAddress(cfg.ReplyTo); err != nil {
			return nil, fmt.Errorf("EMAIL_REPLY_TO must be a valid address")
		}
	}
	awsConfig, err := config.LoadDefaultConfig(ctx, config.WithRegion(cfg.Region), config.WithRetryMaxAttempts(1))
	if err != nil {
		return nil, fmt.Errorf("load SES AWS configuration: %w", err)
	}
	return NewSender(sesv2.NewFromConfig(awsConfig), cfg), nil
}
