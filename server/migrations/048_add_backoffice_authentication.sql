ALTER TABLE users
ADD COLUMN is_platform_admin BOOLEAN NOT NULL DEFAULT false;

ALTER TABLE sessions
ADD COLUMN audience TEXT NOT NULL DEFAULT 'api';

ALTER TABLE sessions
ADD CONSTRAINT chk_sessions_audience
CHECK (audience IN ('api', 'backoffice'));
