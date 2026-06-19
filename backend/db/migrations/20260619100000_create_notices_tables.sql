-- migrate:up
CREATE TABLE IF NOT EXISTS notices (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(180) NOT NULL,
    body TEXT NOT NULL,
    scope_type VARCHAR(30) NOT NULL,
    priority VARCHAR(30) NOT NULL,
    pinned BOOLEAN NOT NULL DEFAULT FALSE,
    expires_at TIMESTAMPTZ NULL,
    created_by_user_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_notices_scope_type
        CHECK (scope_type IN ('all', 'users')),
    CONSTRAINT chk_notices_priority
        CHECK (priority IN ('info', 'warning', 'critical')),
    CONSTRAINT fk_notices_created_by_user_id
        FOREIGN KEY (created_by_user_id) REFERENCES users(id) ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS notice_recipients (
    id BIGSERIAL PRIMARY KEY,
    notice_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    read_at TIMESTAMPTZ NULL,
    hidden_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_notice_recipients_notice_id
        FOREIGN KEY (notice_id) REFERENCES notices(id) ON DELETE CASCADE,
    CONSTRAINT fk_notice_recipients_user_id
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT uq_notice_recipients_notice_user
        UNIQUE (notice_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_notices_created_by_user_id
    ON notices(created_by_user_id);

CREATE INDEX IF NOT EXISTS idx_notices_created_at
    ON notices(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_notice_recipients_user_id
    ON notice_recipients(user_id);

CREATE INDEX IF NOT EXISTS idx_notice_recipients_user_read_at
    ON notice_recipients(user_id, read_at);

-- migrate:down
DROP INDEX IF EXISTS idx_notice_recipients_user_read_at;
DROP INDEX IF EXISTS idx_notice_recipients_user_id;
DROP INDEX IF EXISTS idx_notices_created_at;
DROP INDEX IF EXISTS idx_notices_created_by_user_id;

DROP TABLE IF EXISTS notice_recipients;
DROP TABLE IF EXISTS notices;
