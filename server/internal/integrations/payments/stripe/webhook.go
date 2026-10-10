package stripe

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const webhookTolerance = 5 * time.Minute

type Event struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Created int64  `json:"created"`
	Data    struct {
		Object json.RawMessage `json:"object"`
	} `json:"data"`
}

type Subscription struct {
	ID                string            `json:"id"`
	Customer          string            `json:"customer"`
	Status            string            `json:"status"`
	Metadata          map[string]string `json:"metadata"`
	CancelAtPeriodEnd bool              `json:"cancel_at_period_end"`
	CanceledAt        *int64            `json:"canceled_at"`
	TrialEnd          *int64            `json:"trial_end"`
	Items             struct {
		Data []SubscriptionItem `json:"data"`
	} `json:"items"`
}

type SubscriptionItem struct {
	CurrentPeriodStart int64 `json:"current_period_start"`
	CurrentPeriodEnd   int64 `json:"current_period_end"`
	Price              struct {
		ID string `json:"id"`
	} `json:"price"`
}

func (s Subscription) CurrentPeriod() (int64, int64) {
	if len(s.Items.Data) == 0 {
		return 0, 0
	}
	item := s.Items.Data[0]
	return item.CurrentPeriodStart, item.CurrentPeriodEnd
}

func (c *Client) ParseWebhook(payload []byte, signatureHeader string) (Event, error) {
	if c == nil || c.webhookSecret == "" {
		return Event{}, fmt.Errorf("stripe webhook secret is not configured")
	}

	timestamp, signatures, err := parseSignatureHeader(signatureHeader)
	if err != nil {
		return Event{}, err
	}
	eventTime := time.Unix(timestamp, 0)
	delta := c.now().Sub(eventTime)
	if delta > webhookTolerance || delta < -webhookTolerance {
		return Event{}, fmt.Errorf("stripe webhook timestamp is outside tolerance")
	}

	message := strconv.FormatInt(timestamp, 10) + "." + string(payload)
	mac := hmac.New(sha256.New, []byte(c.webhookSecret))
	_, _ = mac.Write([]byte(message))
	expected := mac.Sum(nil)

	for _, candidate := range signatures {
		raw, decodeErr := hex.DecodeString(candidate)
		if decodeErr == nil && hmac.Equal(expected, raw) {
			var event Event
			if err := json.Unmarshal(payload, &event); err != nil {
				return Event{}, fmt.Errorf("decode Stripe event: %w", err)
			}
			if event.ID == "" || event.Type == "" {
				return Event{}, fmt.Errorf("invalid Stripe event")
			}
			return event, nil
		}
	}
	return Event{}, fmt.Errorf("stripe webhook signature mismatch")
}

func DecodeEventObject[T any](event Event) (T, error) {
	var value T
	if err := json.Unmarshal(event.Data.Object, &value); err != nil {
		return value, fmt.Errorf("decode Stripe %s object: %w", event.Type, err)
	}
	return value, nil
}

func parseSignatureHeader(value string) (int64, []string, error) {
	var timestamp int64
	signatures := make([]string, 0, 2)
	for _, part := range strings.Split(value, ",") {
		key, raw, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok {
			continue
		}
		switch key {
		case "t":
			parsed, err := strconv.ParseInt(raw, 10, 64)
			if err != nil {
				return 0, nil, fmt.Errorf("invalid Stripe webhook timestamp")
			}
			timestamp = parsed
		case "v1":
			if raw != "" {
				signatures = append(signatures, raw)
			}
		}
	}
	if timestamp == 0 || len(signatures) == 0 {
		return 0, nil, fmt.Errorf("invalid Stripe-Signature header")
	}
	return timestamp, signatures, nil
}
