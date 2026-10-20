package webhooks

import (
	"context"
	"net/http"
	"testing"

	"github.com/leamout/leamout/server/internal/database/sqlc"
)

func TestWebhookSenderDoesNotUseProxy(t *testing.T) {
	t.Setenv("HTTPS_PROXY", "http://127.0.0.1:3128")
	client := NewHTTPSender().client
	transport, ok := client.Transport.(*http.Transport)
	if !ok || transport.Proxy != nil {
		t.Fatal("webhook deliveries must use a guarded, proxy-free transport")
	}
}

func TestWebhookURLAndTestSendRejectPrivateDestination(t *testing.T) {
	endpoint := "https://169.254.169.254/latest/meta-data/"
	if _, err := normalizeURL(endpoint); err == nil {
		t.Fatal("private webhook URL accepted")
	}
	if _, err := sendTest(context.Background(), sqlc.WebhookEndpoint{Url: endpoint}); err == nil {
		t.Fatal("webhook test send accepted private destination")
	}
}
