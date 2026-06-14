-- migrate:up
ALTER TABLE users
ADD COLUMN IF NOT EXISTS must_change_password BOOLEAN NOT NULL DEFAULT FALSE;

-- migrate:down
ALTER TABLE users
DROP COLUMN IF EXISTS must_change_password;
