package auth

import (
	"crypto/subtle"
	"net/http"
	"time"

	"github.com/leamout/leamout/server/internal/database/pgconv"
	"github.com/leamout/leamout/server/internal/database/sqlc"
	"github.com/leamout/leamout/server/internal/identity/session"
	"github.com/leamout/leamout/server/internal/integrations/kinde"
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
	client      *kinde.Client
	cipher      *encryption.Cipher
	frontendURL string
	development bool
}

func NewHandler(queries *sqlc.Queries, sessions *session.Service, client *kinde.Client, cipher *encryption.Cipher, frontendURL string, development bool) *Handler {
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
	identitySessionID := identity.SessionID
	if identitySessionID == "" {
		identitySessionID, err = token.Generate(32)
		if err != nil {
			h.failed(w)
			return
		}
	}
	value, created, err := h.sessions.Create(r.Context(), user.ID, identity.Subject, identitySessionID, tokens)
	if err != nil {
		h.failed(w)
		return
	}
	session.SetCookie(w, value, created.ExpiresAt.Time, h.development, "")
	http.Redirect(w, r, h.frontendURL+"/organizations", http.StatusFound)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if err := h.revokeBrowserSession(w, r); err != nil {
		h.failed(w)
		return
	}
	httputil.OK(w, map[string]string{"logout_url": h.client.LogoutURL(h.frontendURL + "/log-in")})
}

// BackofficeLogout signs out the operator locally and at Kinde, using a browser redirect.
func (h *Handler) BackofficeLogout(w http.ResponseWriter, r *http.Request) {
	if err := h.revokeBrowserSession(w, r); err != nil {
		h.failed(w)
		return
	}
	http.Redirect(w, r, h.client.LogoutURL(h.frontendURL+"/log-in"), http.StatusSeeOther)
}

func (h *Handler) revokeBrowserSession(w http.ResponseWriter, r *http.Request) error {
	if cookie, err := r.Cookie(authn.SessionCookieName); err == nil {
		if stored, err := h.sessions.Get(r.Context(), cookie.Value); err == nil {
			if err := h.sessions.Revoke(r.Context(), stored.ID, stored.UserID); err != nil {
				return err
			}
		}
	}
	session.ClearCookie(w, h.development, "")
	return nil
}

func (h *Handler) Account(w http.ResponseWriter, r *http.Request) {
	h.account(w, r, "/settings/profile")
}

func (h *Handler) BackofficeAccount(w http.ResponseWriter, r *http.Request) {
	h.account(w, r, "/organizations")
}

func (h *Handler) account(w http.ResponseWriter, r *http.Request, returnPath string) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	cookie, err := r.Cookie(authn.SessionCookieName)
	if err != nil || cookie.Value == "" {
		httputil.Error(w, apperror.NewUnauthorized("session authentication required"))
		return
	}
	accessToken, err := h.sessions.AccessToken(r.Context(), cookie.Value)
	if err != nil {
		httputil.Error(w, apperror.NewUnauthorized("session authentication required"))
		return
	}
	destination, err := h.client.PortalURL(r.Context(), accessToken, h.frontendURL+returnPath)
	if err != nil {
		httputil.Error(w, apperror.NewServiceUnavailable("account management is unavailable", err))
		return
	}
	http.Redirect(w, r, destination, http.StatusSeeOther)
}

func (h *Handler) setStateCookie(w http.ResponseWriter, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{Name: stateCookie, Value: value, Path: "/", MaxAge: maxAge, Secure: !h.development, HttpOnly: true, SameSite: http.SameSiteLaxMode})
}

func (h *Handler) failed(w http.ResponseWriter) {
	httputil.Error(w, apperror.NewUnauthorized("sign-in could not be completed; return to Leamout and try again"))
}
