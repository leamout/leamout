package ses

import (
	"errors"
	"testing"

	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
)

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
			err := classifySendError(fmt.Errorf("send operation: %w", tc.err))
			var classified *SendError
			if !errors.As(err, &classified) || classified.Code != tc.code || classified.Permanent != tc.permanent {
				t.Fatalf("classification failed: %v", err)
			}
			if strings.Contains(err.Error(), "secret@example.com") || strings.Contains(err.Error(), "012345") {
				t.Fatal("provider error leaked sensitive data")
			}
		})
	}
}
