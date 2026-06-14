-- migrate:up
CREATE TABLE IF NOT EXISTS budget_follow_ups (
    id BIGSERIAL PRIMARY KEY,
    budget_id BIGINT NOT NULL,
    created_by_user_id BIGINT NOT NULL,
    notes TEXT NOT NULL,
    follow_up_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_budget_follow_ups_budget_id
        FOREIGN KEY (budget_id) REFERENCES budgets(id) ON DELETE CASCADE,
    CONSTRAINT fk_budget_follow_ups_created_by_user_id
        FOREIGN KEY (created_by_user_id) REFERENCES users(id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_budget_follow_ups_budget_id ON budget_follow_ups(budget_id);
CREATE INDEX IF NOT EXISTS idx_budget_follow_ups_follow_up_at ON budget_follow_ups(follow_up_at DESC);

-- migrate:down
DROP TABLE IF EXISTS budget_follow_ups;
