package ses

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
)

func TestSenderSendsHTMLOnly(t *testing.T) {
	var received sesv2.SendEmailInput
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"MessageId":"html-message"}`)); err != nil {
			t.Error(err)
		}
	}))
	defer server.Close()

	client := sesv2.NewFromConfig(aws.Config{
		Region:      "us-east-1",
		Credentials: credentials.NewStaticCredentialsProvider("test", "test", ""),
	}, func(options *sesv2.Options) {
		options.BaseEndpoint = aws.String(server.URL)
	})

	sender := NewSender(client, Config{From: "noreply@example.com"})
	result, err := sender.Send(t.Context(), Message{
		To:      "recipient@example.com",
		Subject: "Your verification code",
		HTML:    "<p>012345</p>",
	})
	if err != nil {
		t.Fatal(err)
	}

	if result.MessageID != "html-message" {
		t.Fatal("missing SES message ID")
	}

	if received.Content == nil || received.Content.Simple == nil || received.Content.Simple.Body == nil {
		t.Fatal("missing email content")
	}

	body := received.Content.Simple.Body
	if body.Text != nil || body.Html == nil || aws.ToString(body.Html.Data) != "<p>012345</p>" {
		t.Fatal("expected an HTML-only SES request")
	}
}
