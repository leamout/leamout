package keycloak

import (
	"net/http"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type Config struct {
	Issuer       string `env:"ISSUER"`
	ClientID     string `env:"CLIENT_ID" envDefault:"leamout-console"`
	ClientSecret string `env:"CLIENT_SECRET"`
}

type Client struct {
	config         oauth2.Config
	verifier       *oidc.IDTokenVerifier
	http           *http.Client
	issuer         string
	logoutEndpoint string
}

type Identity struct {
	Subject       string
	SessionID     string `json:"sid"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
}

type Tokens struct {
	OAuth      oauth2.Token `json:"oauth"`
	IDToken    string       `json:"id_token"`
	Nonce      string       `json:"nonce"`
	ValidUntil time.Time    `json:"valid_until"`
}

type providerMetadata struct {
	LogoutEndpoint string `json:"end_session_endpoint"`
}
