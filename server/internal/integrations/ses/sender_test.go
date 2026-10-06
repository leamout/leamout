package ses

import (
	"context"
	"errors"
	"testing"

	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/coffeyvidzro/monogo/internal/platform/email"
)

type fakeAPI struct {
	input *sesv2.SendEmailInput
	err   error
}

func (f *fakeAPI) SendEmail(ctx context.Context, in *sesv2.SendEmailInput, _ ...func(*sesv2.Options)) (*sesv2.SendEmailOutput, error) {
	f.input = in
	return &sesv2.SendEmailOutput{MessageId: aws.String("ses-id")}, f.err
}
func TestSenderMapsMessage(t *testing.T) {
	api := &fakeAPI{}
	sender := NewSender(api, Config{From: "Leamout <sender@example.com>", ConfigurationSet: "transactional"})
	result, err := sender.Send(context.Background(), email.Message{To: "user@example.com", Subject: "Sign in", HTML: "<p>code</p>", Text: "code"})
	if err != nil {
		t.Fatal(err)
	}
	in := api.input
	if result.MessageID != "ses-id" || aws.ToString(in.FromEmailAddress) != "Leamout <sender@example.com>" || in.Destination.ToAddresses[0] != "user@example.com" || aws.ToString(in.Content.Simple.Body.Text.Data) != "code" || aws.ToString(in.Content.Simple.Body.Html.Data) != "<p>code</p>" || aws.ToString(in.Content.Simple.Subject.Charset) != "UTF-8" || aws.ToString(in.ConfigurationSetName) != "transactional" || len(in.ReplyToAddresses) != 0 {
		t.Fatal("incorrect SES request")
	}
}
func TestSenderClassifiesAndRedactsErrors(t *testing.T) {
	secret := aws.String("secret@example.com 012345")
	cases := []struct {
		name      string
		err       error
		code      string
		permanent bool
	}{
		{"rejected", &types.MessageRejected{Message: secret}, "MessageRejected", true},
		{"unverified", &types.MailFromDomainNotVerifiedException{Message: secret}, "MailFromDomainNotVerifiedException", true},
		{"bad request", &types.BadRequestException{Message: secret}, "BadRequestException", true},
		{"not found", &types.NotFoundException{Message: secret}, "NotFoundException", true},
		{"suspended", &types.AccountSuspendedException{Message: secret}, "AccountSuspendedException", true},
		{"paused", &types.SendingPausedException{Message: secret}, "SendingPausedException", true},
		{"throttled", &types.TooManyRequestsException{Message: secret}, "TooManyRequestsException", false},
		{"limit exceeded", &types.LimitExceededException{Message: secret}, "LimitExceededException", false},
		{"internal", &types.InternalServiceErrorException{Message: secret}, "InternalServiceErrorException", false},
		{"transport", errors.New(aws.ToString(secret)), "transport_error", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			api := &fakeAPI{err: fmt.Errorf("send operation: %w", tc.err)}
			_, err := NewSender(api, Config{}).Send(context.Background(), email.Message{})
			var classified *email.SendError
			if !errors.As(err, &classified) || classified.Code != tc.code || classified.Permanent != tc.permanent {
				t.Fatalf("classification failed: %v", err)
			}
			if strings.Contains(err.Error(), "secret@example.com") || strings.Contains(err.Error(), "012345") {
				t.Fatal("provider error leaked sensitive data")
			}
		})
	}
}
