package providers

import (
	"context"
	"errors"
	"testing"

	"github.com/leamout/contracts/ai"
	aicatalog "github.com/leamout/leamout/server/internal/ai/catalog"
)

func TestVerifierUsesCatalogCredentialVerifier(t *testing.T) {
	tests := []struct {
		name            string
		verifyErr       error
		wantState       string
		wantFailureCode string
	}{
		{
			name:      "ready",
			wantState: ConnectionReady,
		},
		{
			name:            "invalid",
			verifyErr:       errors.New("verify credential: HTTP 401"),
			wantState:       ConnectionInvalid,
			wantFailureCode: "authentication_failed",
		},
		{
			name:            "rate limited",
			verifyErr:       errors.New("verify credential: HTTP 429"),
			wantState:       ConnectionUnavailable,
			wantFailureCode: "provider_rate_limited",
		},
		{
			name:            "unavailable",
			verifyErr:       errors.New("verify credential: HTTP 503"),
			wantState:       ConnectionUnavailable,
			wantFailureCode: "provider_unavailable",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			catalog, err := aicatalog.New(testCredentialProvider{verifyErr: test.verifyErr})
			if err != nil {
				t.Fatal(err)
			}
			result, err := NewVerifier(catalog).Verify(
				context.Background(),
				"test",
				"test-secret",
			)
			if err != nil {
				t.Fatalf("Verify() error = %v", err)
			}
			if result.State != test.wantState {
				t.Errorf("state = %q, want %q", result.State, test.wantState)
			}
			if result.FailureCode != test.wantFailureCode {
				t.Errorf("failure code = %q, want %q", result.FailureCode, test.wantFailureCode)
			}
		})
	}
}

func TestVerifierPreservesContextCancellation(t *testing.T) {
	catalog, err := aicatalog.New(testCredentialProvider{})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = NewVerifier(catalog).Verify(ctx, "test", "test-secret")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Verify() error = %v, want context.Canceled", err)
	}
}

type testCredentialProvider struct {
	verifyErr error
}

func (testCredentialProvider) Descriptor() ai.Descriptor {
	return ai.Descriptor{
		ID:           "test",
		Name:         "Test",
		Kind:         ai.KindLLM,
		Capabilities: []ai.Capability{ai.CapabilityStreaming},
	}
}

func (p testCredentialProvider) VerifyCredential(ctx context.Context, _ string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return p.verifyErr
}

func (testCredentialProvider) Generate(
	context.Context,
	ai.LLMRequest,
) (ai.LLMStream, error) {
	return nil, errors.New("unused")
}
