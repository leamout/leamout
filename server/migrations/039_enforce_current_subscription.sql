-- The current commercial model resolves exactly one active/past-due subscription
-- per organization. Keep historical and pending rows, but make the current-state
-- invariant authoritative in PostgreSQL so concurrent creates/transitions cannot
-- leave state resolution choosing arbitrarily between multiple current rows.
CREATE UNIQUE INDEX IF NOT EXISTS uq_subscriptions_current_organization
    ON subscriptions (organization_id)
    WHERE status IN ('active', 'past_due');
