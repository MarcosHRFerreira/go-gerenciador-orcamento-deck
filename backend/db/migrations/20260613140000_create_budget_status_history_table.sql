-- migrate:up
CREATE TABLE IF NOT EXISTS budget_status_history (
    id BIGSERIAL PRIMARY KEY,
    budget_id BIGINT NOT NULL,
    from_status_id BIGINT NULL,
    to_status_id BIGINT NOT NULL,
    changed_by_user_id BIGINT NOT NULL,
    notes TEXT NOT NULL DEFAULT '',
    changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_budget_status_history_budget_id
        FOREIGN KEY (budget_id) REFERENCES budgets(id) ON DELETE CASCADE,
    CONSTRAINT fk_budget_status_history_from_status_id
        FOREIGN KEY (from_status_id) REFERENCES budget_statuses(id) ON DELETE SET NULL,
    CONSTRAINT fk_budget_status_history_to_status_id
        FOREIGN KEY (to_status_id) REFERENCES budget_statuses(id) ON DELETE RESTRICT,
    CONSTRAINT fk_budget_status_history_changed_by_user_id
        FOREIGN KEY (changed_by_user_id) REFERENCES users(id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_budget_status_history_budget_id ON budget_status_history(budget_id);
CREATE INDEX IF NOT EXISTS idx_budget_status_history_changed_at ON budget_status_history(changed_at DESC);

-- migrate:down
DROP TABLE IF EXISTS budget_status_history;
