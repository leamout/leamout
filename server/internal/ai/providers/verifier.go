package providers

import (
	"context"
	"fmt"
	"strings"

	aicatalog "github.com/leamout/leamout/server/internal/ai/catalog"
)

type Verifier struct {
	catalog *aicatalog.Catalog
}

type verificationResult struct {
	State       string
	FailureCode string
}

func NewVerifier(catalog *aicatalog.Catalog) *Verifier {
	return &Verifier{catalog: catalog}
}

func (v *Verifier) Verify(
	ctx context.Context,
	provider string,
	secret string,
) (verificationResult, error) {
	if v == nil || v.catalog == nil {
		return verificationResult{}, fmt.Errorf("AI provider catalog is unavailable")
	}
	verifier, ok := v.catalog.CredentialVerifier(provider)
	if !ok {
		return verificationResult{}, fmt.Errorf("AI provider %q does not support credential verification", provider)
	}
	if err := verifier.VerifyCredential(ctx, secret); err != nil {
		if ctx != nil && ctx.Err() != nil {
			return verificationResult{}, ctx.Err()
		}
		message := strings.ToLower(err.Error())
		switch {
		case strings.Contains(message, "http 401"), strings.Contains(message, "http 403"):
			return verificationResult{
				State:       ConnectionInvalid,
				FailureCode: "authentication_failed",
			}, nil
		case strings.Contains(message, "http 429"):
			return verificationResult{
				State:       ConnectionUnavailable,
				FailureCode: "provider_rate_limited",
			}, nil
		default:
			return verificationResult{
				State:       ConnectionUnavailable,
				FailureCode: "provider_unavailable",
			}, nil
		}
	}
	return verificationResult{State: ConnectionReady}, nil
}
