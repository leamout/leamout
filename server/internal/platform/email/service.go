package email

import (
	"context"
	"encoding/json"
	"fmt"
	"net/mail"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/leamout/leamout/server/internal/database/pgconv"
	"github.com/leamout/leamout/server/internal/database/sqlc"
	"github.com/leamout/leamout/server/internal/security/encryption"
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
	err = sqlc.New(tx).CreateEmailDelivery(ctx, sqlc.CreateEmailDeliveryParams{ID: id, Recipient: req.To, Template: req.Template, EncryptedData: &payload, CancellationKey: req.CancellationKey, ExpiresAt: pgconv.NullableTimestamptz(&req.Data.ExpiresAt)})
	return id, err
}

// CancelTx cancels jobs belonging to the caller's opaque business reference.
func (s *Service) CancelTx(ctx context.Context, tx pgx.Tx, key string) error {
	return sqlc.New(tx).CancelEmailDeliveries(ctx, &key)
}
