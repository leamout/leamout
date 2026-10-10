package ses

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
)

type Message struct {
	To      string
	Subject string
	HTML    string
	Text    string
}

type Result struct {
	MessageID string
}

// SendError classifies provider failures without persisting recipient or body data.
type SendError struct {
	Code      string
	Permanent bool
}

func (e *SendError) Error() string { return "email provider: " + e.Code }

type Sender struct {
	client *sesv2.Client
	config Config
}

func NewSender(client *sesv2.Client, cfg Config) *Sender {
	return &Sender{client: client, config: cfg}
}

func (s *Sender) Send(ctx context.Context, message Message) (Result, error) {
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
		return Result{}, classifySendError(err)
	}
	if output == nil || aws.ToString(output.MessageId) == "" {
		return Result{}, &SendError{Code: "missing_message_id"}
	}

	return Result{MessageID: aws.ToString(output.MessageId)}, nil
}

func content(value string) *types.Content {
	return &types.Content{
		Data:    aws.String(value),
		Charset: aws.String("UTF-8"),
	}
}

// Classify concrete SES errors, including errors wrapped by the SDK.
// Persist only known codes, never provider messages containing recipient data.
func classifySendError(err error) *SendError {
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
		return &SendError{Code: "MessageRejected", Permanent: true}
	case errors.As(err, &unverified):
		return &SendError{Code: "MailFromDomainNotVerifiedException", Permanent: true}
	case errors.As(err, &badRequest):
		return &SendError{Code: "BadRequestException", Permanent: true}
	case errors.As(err, &notFound):
		return &SendError{Code: "NotFoundException", Permanent: true}
	case errors.As(err, &suspended):
		return &SendError{Code: "AccountSuspendedException", Permanent: true}
	case errors.As(err, &paused):
		return &SendError{Code: "SendingPausedException", Permanent: true}
	case errors.As(err, &throttled):
		return &SendError{Code: "TooManyRequestsException"}
	case errors.As(err, &limitExceeded):
		return &SendError{Code: "LimitExceededException"}
	case errors.As(err, &internal):
		return &SendError{Code: "InternalServiceErrorException"}
	default:
		return &SendError{Code: "transport_error"}
	}
}
