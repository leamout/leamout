package ses

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/smithy-go"
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
	sender := NewSender(api, Config{From: "Leamout <sender@example.com>", ReplyTo: "support@example.com", ConfigurationSet: "transactional"})
	result, err := sender.Send(context.Background(), email.Message{To: "user@example.com", Subject: "Sign in", HTML: "<p>code</p>", Text: "code"})
	if err != nil {
		t.Fatal(err)
	}
	in := api.input
	if result.MessageID != "ses-id" || aws.ToString(in.FromEmailAddress) != "Leamout <sender@example.com>" || in.Destination.ToAddresses[0] != "user@example.com" || aws.ToString(in.Content.Simple.Body.Text.Data) != "code" || aws.ToString(in.Content.Simple.Body.Html.Data) != "<p>code</p>" || aws.ToString(in.Content.Simple.Subject.Charset) != "UTF-8" || aws.ToString(in.ConfigurationSetName) != "transactional" || in.ReplyToAddresses[0] != "support@example.com" {
		t.Fatal("incorrect SES request")
	}
}
func TestSenderClassifiesAndRedactsErrors(t *testing.T) {
	for _, tc := range []struct {
		code      string
		permanent bool
	}{{"MessageRejected", true}, {"TooManyRequestsException", false}, {"InternalServiceErrorException", false}} {
		api := &fakeAPI{err: &smithy.GenericAPIError{Code: tc.code, Message: "secret@example.com 012345"}}
		_, err := NewSender(api, Config{}).Send(context.Background(), email.Message{})
		var e *email.SendError
		if !errors.As(err, &e) || e.Code != tc.code || e.Permanent != tc.permanent || e.Error() == api.err.Error() {
			t.Fatalf("classification failed: %v", err)
		}
	}
}
