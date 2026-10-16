-- Short-lived, single-use OIDC authorization attempts. No passwords or OTPs.
CREATE TABLE IF NOT EXISTS oidc_login_attempts (
    state_hash TEXT PRIMARY KEY,
    verifier_ciphertext TEXT NOT NULL,
    nonce TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX oidc_login_attempts_expiry_idx ON oidc_login_attempts (expires_at);

CREATE TABLE IF NOT EXISTS oidc_logout_events (
    token_id TEXT PRIMARY KEY,
    expires_at TIMESTAMPTZ NOT NULL
);
