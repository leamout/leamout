package kinde

import (
	"net/http"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-jose/go-jose/v4"
	"golang.org/x/oauth2"
)

type Config struct {
	Issuer       string `env:"ISSUER"`
	ClientID     string `env:"CLIENT_ID"`
	ClientSecret string `env:"CLIENT_SECRET"`
}

type Client struct {
	config   oauth2.Config
	http     *http.Client
	issuer   string
	mu       sync.Mutex
	verifier *oidc.IDTokenVerifier

	webhookMu        sync.Mutex
	webhookKeys      jose.JSONWebKeySet
	webhookKeysAt    time.Time
	webhookAttemptAt time.Time
}

type Identity struct {
	Subject       string
	SessionID     string `json:"sid"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
}

type Tokens struct {
	OAuth      oauth2.Token `json:"oauth"`
	IDToken    string       `json:"id_token"`
	Nonce      string       `json:"nonce"`
	ValidUntil time.Time    `json:"valid_until"`
}
