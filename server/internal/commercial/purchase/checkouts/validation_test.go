package checkouts

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestValidateCreateEnforcesTargetAndProviderPair(t *testing.T) {
	now := time.Now()
	walletID := uuid.New()
	priceID := uuid.New()
	valid := CreateInput{
		WalletID:      &walletID,
		Type:          TypeWalletTopup,
		Provider:      ProviderPaystack,
		PaymentMethod: MethodMobileMoney,
		Reference:     "wallet.123",
		AmountMinor:   100,
		Currency:      "GHS",
		ExpiresAt:     now.Add(time.Hour),
	}

	tests := []struct {
		name  string
		input CreateInput
		want  error
	}{
		{name: "wallet topup", input: valid},
		{name: "subscription", input: CreateInput{
			PriceID:       &priceID,
			Type:          TypeSubscription,
			Provider:      ProviderStripe,
			PaymentMethod: MethodCard,
			Reference:     "sub.123",
			AmountMinor:   100,
			Currency:      "USD",
			ExpiresAt:     now.Add(time.Hour),
		}},
		{name: "mixed target", input: func() CreateInput {
			input := valid
			input.PriceID = &priceID
			return input
		}(), want: ErrInvalidCheckout},
		{name: "wrong provider method", input: func() CreateInput {
			input := valid
			input.PaymentMethod = MethodCard
			return input
		}(), want: ErrInvalidCheckout},
		{name: "browser style invalid reference", input: func() CreateInput {
			input := valid
			input.Reference = "wallet 123"
			return input
		}(), want: ErrInvalidCheckout},
		{name: "expired", input: func() CreateInput {
			input := valid
			input.ExpiresAt = now
			return input
		}(), want: ErrInvalidCheckout},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := validateCreate(test.input, now); !errors.Is(err, test.want) {
				t.Fatalf("validateCreate() error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestValidateTransitionRequiresTerminalClosure(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name       string
		transition Transition
		want       error
	}{
		{name: "processing", transition: Transition{Expected: StatusPending, Status: StatusProcessing, NextAction: ActionWait}},
		{name: "success", transition: Transition{Expected: StatusProcessing, Status: StatusSucceeded, NextAction: ActionNone, CompletedAt: &now}},
		{name: "terminal action", transition: Transition{Expected: StatusProcessing, Status: StatusFailed, NextAction: ActionWait, CompletedAt: &now}, want: ErrInvalidTransition},
		{name: "terminal without timestamp", transition: Transition{Expected: StatusProcessing, Status: StatusFailed, NextAction: ActionNone}, want: ErrInvalidTransition},
		{name: "reopen terminal", transition: Transition{Expected: StatusSucceeded, Status: StatusProcessing, NextAction: ActionWait}, want: ErrInvalidTransition},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := validateTransition(test.transition); !errors.Is(err, test.want) {
				t.Fatalf("validateTransition() error = %v, want %v", err, test.want)
			}
		})
	}
}
