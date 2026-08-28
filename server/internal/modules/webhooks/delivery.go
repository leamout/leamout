package webhooks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/leamout/leamout/internal/database/sqlc"
	"io"
	"net/http"
	"time"
)

const maxResponseBody = 4096

func sendTest(c context.Context, endpoint sqlc.WebhookEndpoint) (int, error) {
	body, err := json.Marshal(map[string]any{"id": "test", "type": "webhook.test", "occurred_at": time.Now().UTC(), "data": map[string]bool{"test": true}})
	if err != nil {
		return 0, err
	}
	req, err := http.NewRequestWithContext(c, http.MethodPost, endpoint.Url, bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	now := time.Now().UTC()
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Leamout-Webhooks/1.0")
	req.Header.Set("X-Leamout-Event", "webhook.test")
	req.Header.Set("X-Leamout-Timestamp", fmt.Sprintf("%d", now.Unix()))
	req.Header.Set(signatureHeader, sign(endpoint.SigningSecret, body, now))
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, maxResponseBody))
	return resp.StatusCode, nil
}
