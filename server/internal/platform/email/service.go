package email

import (
	"context"
	"encoding/json"
	"fmt"
	"net/mail"
	"time"

	"github.com/coffeyvidzro/monogo/internal/security/encryption"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Service struct {
	cipher   *encryption.Cipher
	renderer *Renderer
}

func NewService(cipher *encryption.Cipher) *Service {
	return &Service{cipher: cipher, renderer: NewRenderer()}
}

// QueueTx must use the same transaction as the business mutation.
func (s *Service) QueueTx(ctx context.Context, tx pgx.Tx, req Request) (uuid.UUID, error) {
	if s == nil || s.cipher == nil || tx == nil {
		return uuid.Nil, fmt.Errorf("email delivery is not configured")
	}
	if req.Template == "otp" && (req.ChallengeID == nil || *req.ChallengeID == uuid.Nil) {
		return uuid.Nil, fmt.Errorf("OTP challenge is required")
	}
	address, err := mail.ParseAddress(req.To)
	if err != nil || address.Address != req.To {
		return uuid.Nil, fmt.Errorf("invalid email recipient")
	}
	if !req.Data.ExpiresAt.After(time.Now()) {
		return uuid.Nil, fmt.Errorf("email has expired")
	}
	if _, err := s.renderer.Render(req.Template, req.Data); err != nil {
		return uuid.Nil, err
	}
	data, err := json.Marshal(req.Data)
	if err != nil {
		return uuid.Nil, err
	}
	id := uuid.New()
	payload, err := s.cipher.EncryptForScope("email:"+id.String(), string(data))
	if err != nil {
		return uuid.Nil, err
	}
	_, err = tx.Exec(ctx, `INSERT INTO email_deliveries (id, recipient, template, encrypted_data, challenge_id, expires_at) VALUES ($1,$2,$3,$4,$5,$6)`, id, req.To, req.Template, payload, req.ChallengeID, req.Data.ExpiresAt)
	return id, err
}
