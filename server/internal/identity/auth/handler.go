package auth

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"time"

	"github.com/leamout/leamout/server/internal/database/pgconv"
	"github.com/leamout/leamout/server/internal/database/sqlc"
	"github.com/leamout/leamout/server/internal/identity/session"
	"github.com/leamout/leamout/server/internal/integrations/keycloak"
	"github.com/leamout/leamout/server/internal/security/authn"
	"github.com/leamout/leamout/server/internal/security/encryption"
	"github.com/leamout/leamout/server/internal/security/token"
	"github.com/leamout/leamout/server/pkg/apperror"
	"github.com/leamout/leamout/server/pkg/httputil"
	"golang.org/x/oauth2"
)

const stateCookie = "leamout-oidc-state"

type Handler struct {
	queries     *sqlc.Queries
	sessions    *session.Service
	client      *keycloak.Client
	cipher      *encryption.Cipher
	frontendURL string
	development bool
}

func NewHandler(queries *sqlc.Queries, sessions *session.Service, client *keycloak.Client, cipher *encryption.Cipher, frontendURL string, development bool) *Handler {
	return &Handler{
		queries:     queries,
		sessions:    sessions,
		client:      client,
		cipher:      cipher,
		frontendURL: frontendURL,
		development: development,
	}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	state, err := token.Generate(32)
	if err != nil {
		h.failed(w)
		return
	}
	nonce, err := token.Generate(32)
	if err != nil {
		h.failed(w)
		return
	}
	verifier := oauth2.GenerateVerifier()
	encrypted, err := h.cipher.EncryptForScope("oidc-login:"+token.Hash(state), verifier)
	if err != nil {
		h.failed(w)
		return
	}
	expiry := time.Now().Add(5 * time.Minute)
	if err := h.queries.DeleteExpiredOIDCLoginAttempts(r.Context()); err != nil {
		h.failed(w)
		return
	}
	if err := h.queries.CreateOIDCLoginAttempt(r.Context(), sqlc.CreateOIDCLoginAttemptParams{
		StateHash:          token.Hash(state),
		VerifierCiphertext: encrypted,
		Nonce:              nonce,
		ExpiresAt:          pgconv.NullableTimestamptz(&expiry),
	}); err != nil {
		h.failed(w)
		return
	}
	h.setStateCookie(w, state, 300)
	http.Redirect(w, r, h.client.AuthorizationURL(state, nonce, verifier, r.URL.Query().Get("register") == "true"), http.StatusFound)
}

func (h *Handler) Callback(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(stateCookie)
	state := r.URL.Query().Get("state")
	h.setStateCookie(w, "", -1)
	if err != nil || state == "" || subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(state)) != 1 {
		h.failed(w)
		return
	}
	attempt, err := h.queries.ConsumeOIDCLoginAttempt(r.Context(), token.Hash(state))
	if err != nil || r.URL.Query().Get("error") != "" || r.URL.Query().Get("code") == "" {
		h.failed(w)
		return
	}
	verifier, err := h.cipher.DecryptForScope("oidc-login:"+token.Hash(state), attempt.VerifierCiphertext)
	if err != nil {
		h.failed(w)
		return
	}
	identity, tokens, err := h.client.Exchange(r.Context(), r.URL.Query().Get("code"), verifier, attempt.Nonce)
	if err != nil {
		h.failed(w)
		return
	}
	user, err := h.queries.UpsertIdentityUser(r.Context(), sqlc.UpsertIdentityUserParams{
		IdentityIssuer:  h.client.Issuer(),
		IdentitySubject: identity.Subject,
		Email:           identity.Email,
		Name:            &identity.Name,
	})
	if err != nil {
		h.failed(w)
		return
	}
	value, created, err := h.sessions.Create(r.Context(), user.ID, identity.Subject, identity.SessionID, tokens)
	if err != nil {
		h.failed(w)
		return
	}
	session.SetCookie(w, value, created.ExpiresAt.Time, h.development, "")
	http.Redirect(w, r, h.frontendURL+"/organizations", http.StatusFound)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	idToken := ""
	if cookie, err := r.Cookie(authn.SessionCookieName); err == nil {
		stored, err := h.sessions.Get(r.Context(), cookie.Value)
		if err == nil {
			clear, err := h.cipher.DecryptForScope("identity-session:"+stored.UserID.String(), stored.TokensCiphertext)
			if err == nil {
				var tokens keycloak.Tokens
				if json.Unmarshal([]byte(clear), &tokens) == nil {
					idToken = tokens.IDToken
				}
			}
			if err := h.sessions.Revoke(r.Context(), stored.ID, stored.UserID); err != nil {
				h.failed(w)
				return
			}
		}
	}
	session.ClearCookie(w, h.development, "")
	httputil.OK(w, map[string]string{"logout_url": h.client.LogoutURL(idToken, h.frontendURL+"/log-in")})
}

func (h *Handler) Account(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, h.client.Issuer()+"/account/", http.StatusFound)
}

func (h *Handler) setStateCookie(w http.ResponseWriter, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{Name: stateCookie, Value: value, Path: "/", MaxAge: maxAge, Secure: !h.development, HttpOnly: true, SameSite: http.SameSiteLaxMode})
}

func (h *Handler) failed(w http.ResponseWriter) {
	httputil.Error(w, apperror.NewUnauthorized("sign-in could not be completed; return to Leamout and try again"))
}

func (h *Handler) BackchannelLogout(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 32<<10)
	if err := r.ParseForm(); err != nil {
		h.failed(w)
		return
	}
	event, err := h.client.VerifyLogout(r.Context(), r.PostForm.Get("logout_token"))
	if err != nil {
		h.failed(w)
		return
	}
	if err := h.queries.DeleteExpiredOIDCLogoutEvents(r.Context()); err != nil {
		h.failed(w)
		return
	}
	if err := h.queries.ApplyOIDCLogout(r.Context(), sqlc.ApplyOIDCLogoutParams{
		TokenID: event.TokenID, ExpiresAt: pgconv.NullableTimestamptz(&event.Expiry),
		Issuer: event.Issuer, Subject: event.Subject, SessionID: event.SessionID,
	}); err != nil {
		h.failed(w)
		return
	}
	w.WriteHeader(http.StatusOK)
}
