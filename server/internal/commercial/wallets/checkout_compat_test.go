package wallets

import (
	"context"

	"github.com/google/uuid"
	commercialpayments "github.com/leamout/leamout/internal/commercial/payments"
)

func (s *paymentStub) GetByCheckout(
	context.Context,
	uuid.UUID,
	uuid.UUID,
) (commercialpayments.Payment, error) {
	return s.payment, nil
}
