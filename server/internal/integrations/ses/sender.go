package ses

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/leamout/leamout/server/internal/platform/email"
)

type API interface {
	SendEmail(
		context.Context,
		*sesv2.SendEmailInput,
		...func(*sesv2.Options),
	) (*sesv2.SendEmailOutput, error)
}

type Sender struct {
	client API
	config Config
}

func NewSender(client API, cfg Config) *Sender {
	return &Sender{client: client, config: cfg}
}

func (s *Sender) Send(ctx context.Context, message email.Message) (email.Result, error) {
	input := &sesv2.SendEmailInput{
		FromEmailAddress: aws.String(s.config.From),
		Destination: &types.Destination{
			ToAddresses: []string{message.To},
		},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: content(message.Subject),
				Body: &types.Body{
					Html: content(message.HTML),
					Text: content(message.Text),
				},
			},
		},
	}

	if s.config.ConfigurationSet != "" {
		input.ConfigurationSetName = aws.String(s.config.ConfigurationSet)
	}

	output, err := s.client.SendEmail(ctx, input)
	if err != nil {
		return email.Result{}, classifySendError(err)
	}
	if output == nil || aws.ToString(output.MessageId) == "" {
		return email.Result{}, &email.SendError{Code: "missing_message_id"}
	}

	return email.Result{MessageID: aws.ToString(output.MessageId)}, nil
}

func content(value string) *types.Content {
	return &types.Content{
		Data:    aws.String(value),
		Charset: aws.String("UTF-8"),
	}
}

// Classify concrete SES errors, including errors wrapped by the SDK.
// Persist only known codes, never provider messages containing recipient data.
func classifySendError(err error) *email.SendError {
	var (
		rejected      *types.MessageRejected
		unverified    *types.MailFromDomainNotVerifiedException
		badRequest    *types.BadRequestException
		notFound      *types.NotFoundException
		suspended     *types.AccountSuspendedException
		paused        *types.SendingPausedException
		throttled     *types.TooManyRequestsException
		limitExceeded *types.LimitExceededException
		internal      *types.InternalServiceErrorException
	)

	switch {
	case errors.As(err, &rejected):
		return &email.SendError{Code: "MessageRejected", Permanent: true}
	case errors.As(err, &unverified):
		return &email.SendError{Code: "MailFromDomainNotVerifiedException", Permanent: true}
	case errors.As(err, &badRequest):
		return &email.SendError{Code: "BadRequestException", Permanent: true}
	case errors.As(err, &notFound):
		return &email.SendError{Code: "NotFoundException", Permanent: true}
	case errors.As(err, &suspended):
		return &email.SendError{Code: "AccountSuspendedException", Permanent: true}
	case errors.As(err, &paused):
		return &email.SendError{Code: "SendingPausedException", Permanent: true}
	case errors.As(err, &throttled):
		return &email.SendError{Code: "TooManyRequestsException"}
	case errors.As(err, &limitExceeded):
		return &email.SendError{Code: "LimitExceededException"}
	case errors.As(err, &internal):
		return &email.SendError{Code: "InternalServiceErrorException"}
	default:
		return &email.SendError{Code: "transport_error"}
	}
}
