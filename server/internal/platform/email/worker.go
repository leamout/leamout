package email

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/leamout/leamout/server/internal/integrations/ses"
	"github.com/leamout/leamout/server/internal/security/encryption"
)

type Worker struct {
	repo     *Repository
	sender   *ses.Sender
	cipher   *encryption.Cipher
	renderer *Renderer
}

func NewWorker(repo *Repository, sender *ses.Sender, cipher *encryption.Cipher) *Worker {
	return &Worker{repo: repo, sender: sender, cipher: cipher, renderer: NewRenderer()}
}
func (w *Worker) Run(ctx context.Context) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		if err := w.ProcessOne(ctx); err != nil && !errors.Is(err, pgx.ErrNoRows) {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}
func (w *Worker) ProcessOne(ctx context.Context) error {
	lease := uuid.New()
	d, err := w.repo.Claim(ctx, lease)
	if err != nil {
		return err
	}
	fail := func(code string, permanent bool) error { return w.repo.Fail(ctx, d, lease, code, permanent) }
	data, err := w.cipher.DecryptForScope("email:"+d.ID.String(), d.Payload)
	if err != nil {
		return fail("invalid_payload", true)
	}
	var values Data
	if err := json.Unmarshal([]byte(data), &values); err != nil {
		return fail("invalid_payload", true)
	}
	if !d.ExpiresAt.After(time.Now()) {
		return fail("expired", true)
	}
	message, err := w.renderer.Render(d.Template, values)
	if err != nil {
		return fail("invalid_template", true)
	}
	ready, err := w.repo.Ready(ctx, d.ID, lease)
	if err != nil {
		return err
	}
	if !ready {
		return fail("superseded_or_expired", true)
	}
	message.To = d.To
	sendCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	result, err := w.sender.Send(sendCtx, message)
	if err != nil {
		var provider *ses.SendError
		if errors.As(err, &provider) {
			return fail(provider.Code, provider.Permanent)
		}
		// Never persist raw SDK errors: they may contain addresses or request bodies.
		return fail("send_failed", false)
	}
	if result.MessageID == "" {
		return fail("missing_message_id", false)
	}
	if err := w.repo.Complete(ctx, d.ID, lease, result); err != nil {
		return fmt.Errorf("complete email delivery: %w", err)
	}
	return nil
}
