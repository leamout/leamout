package ses

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/aws/smithy-go"
	"github.com/coffeyvidzro/monogo/internal/platform/email"
)

type API interface {
	SendEmail(context.Context, *sesv2.SendEmailInput, ...func(*sesv2.Options)) (*sesv2.SendEmailOutput, error)
}
type Sender struct {
	client API
	config Config
}

func NewSender(client API, cfg Config) *Sender { return &Sender{client: client, config: cfg} }
func (s *Sender) Send(ctx context.Context, m email.Message) (email.Result, error) {
	content := func(value string) *types.Content {
		return &types.Content{Data: aws.String(value), Charset: aws.String("UTF-8")}
	}
	input := &sesv2.SendEmailInput{
		FromEmailAddress: aws.String(s.config.From),
		Destination:      &types.Destination{ToAddresses: []string{m.To}},
		Content:          &types.EmailContent{Simple: &types.Message{Subject: content(m.Subject), Body: &types.Body{Html: content(m.HTML), Text: content(m.Text)}}},
	}
	if s.config.ReplyTo != "" {
		input.ReplyToAddresses = []string{s.config.ReplyTo}
	}
	if s.config.ConfigurationSet != "" {
		input.ConfigurationSetName = aws.String(s.config.ConfigurationSet)
	}
	out, err := s.client.SendEmail(ctx, input)
	if err != nil {
		code := "transport_error"
		permanent := false
		var api smithy.APIError
		if errors.As(err, &api) {
			// Store only known codes, never arbitrary provider error text.
			switch api.ErrorCode() {
			case "MessageRejected", "MailFromDomainNotVerifiedException", "BadRequestException", "NotFoundException", "AccountSuspendedException", "SendingPausedException", "AccessDeniedException":
				code = api.ErrorCode()
				permanent = true
			case "TooManyRequestsException", "LimitExceededException", "InternalServiceErrorException":
				code = api.ErrorCode()
			default:
				code = "provider_error"
			}
		}
		return email.Result{}, &email.SendError{Code: code, Permanent: permanent}
	}
	if out == nil || aws.ToString(out.MessageId) == "" {
		return email.Result{}, &email.SendError{Code: "missing_message_id"}
	}
	return email.Result{MessageID: aws.ToString(out.MessageId)}, nil
}
