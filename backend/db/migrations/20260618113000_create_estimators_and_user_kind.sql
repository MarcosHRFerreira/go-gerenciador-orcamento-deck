-- migrate:up
CREATE TABLE IF NOT EXISTS estimators (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(30) NOT NULL,
    name VARCHAR(150) NOT NULL,
    email VARCHAR(250) NOT NULL DEFAULT '',
    phone VARCHAR(50) NOT NULL DEFAULT '',
    active BOOLEAN NOT NULL DEFAULT TRUE,
    notes TEXT NOT NULL DEFAULT '',
    user_id BIGINT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_estimators_code UNIQUE (code),
    CONSTRAINT uq_estimators_user_id UNIQUE (user_id),
    CONSTRAINT fk_estimators_user_id
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS user_kind VARCHAR(30) NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.table_constraints
        WHERE table_name = 'users'
          AND constraint_name = 'chk_users_user_kind'
    ) THEN
        ALTER TABLE users
            ADD CONSTRAINT chk_users_user_kind
            CHECK (user_kind IS NULL OR user_kind IN ('salesperson', 'estimator'));
    END IF;
END $$;

UPDATE users
SET user_kind = 'salesperson'
WHERE role = 'user'
  AND user_kind IS NULL;

ALTER TABLE budgets
    ADD COLUMN IF NOT EXISTS estimator_id BIGINT NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.table_constraints
        WHERE table_name = 'budgets'
          AND constraint_name = 'fk_budgets_estimator_id'
    ) THEN
        ALTER TABLE budgets
            ADD CONSTRAINT fk_budgets_estimator_id
            FOREIGN KEY (estimator_id) REFERENCES estimators(id) ON DELETE SET NULL;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_estimators_active ON estimators(active);
CREATE INDEX IF NOT EXISTS idx_budgets_estimator_id ON budgets(estimator_id);

-- migrate:down
DROP INDEX IF EXISTS idx_budgets_estimator_id;
DROP INDEX IF EXISTS idx_estimators_active;

ALTER TABLE budgets
    DROP CONSTRAINT IF EXISTS fk_budgets_estimator_id;

ALTER TABLE budgets
    DROP COLUMN IF EXISTS estimator_id;

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS chk_users_user_kind;

ALTER TABLE users
    DROP COLUMN IF EXISTS user_kind;

DROP TABLE IF EXISTS estimators;
