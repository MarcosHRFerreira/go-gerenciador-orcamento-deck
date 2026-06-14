-- migrate:up
CREATE TABLE IF NOT EXISTS budgets (
    id BIGSERIAL PRIMARY KEY,
    budget_number VARCHAR(50) NOT NULL,
    year_budget INTEGER NOT NULL,
    revision INTEGER NOT NULL DEFAULT 0,
    sent_at TIMESTAMPTZ NOT NULL,
    gross_value NUMERIC(14,2) NOT NULL DEFAULT 0,
    commission_value NUMERIC(14,2) NOT NULL DEFAULT 0,
    area_m2 NUMERIC(14,2) NOT NULL DEFAULT 0,
    status_id BIGINT NOT NULL,
    priority_id BIGINT NULL,
    installer_id BIGINT NULL,
    project_id BIGINT NULL,
    salesperson_id BIGINT NULL,
    contact_id BIGINT NULL,
    loss_reason_id BIGINT NULL,
    competitor_name VARCHAR(150) NOT NULL DEFAULT '',
    competitor_price NUMERIC(14,2) NULL,
    designer_name VARCHAR(150) NOT NULL DEFAULT '',
    specification_details TEXT NOT NULL DEFAULT '',
    current_follow_up TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_budgets_budget_number_year UNIQUE (budget_number, year_budget),
    CONSTRAINT fk_budgets_status_id
        FOREIGN KEY (status_id) REFERENCES budget_statuses(id) ON DELETE RESTRICT,
    CONSTRAINT fk_budgets_priority_id
        FOREIGN KEY (priority_id) REFERENCES priorities(id) ON DELETE SET NULL,
    CONSTRAINT fk_budgets_installer_id
        FOREIGN KEY (installer_id) REFERENCES installers(id) ON DELETE SET NULL,
    CONSTRAINT fk_budgets_project_id
        FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE SET NULL,
    CONSTRAINT fk_budgets_salesperson_id
        FOREIGN KEY (salesperson_id) REFERENCES salespeople(id) ON DELETE SET NULL,
    CONSTRAINT fk_budgets_contact_id
        FOREIGN KEY (contact_id) REFERENCES contacts(id) ON DELETE SET NULL,
    CONSTRAINT fk_budgets_loss_reason_id
        FOREIGN KEY (loss_reason_id) REFERENCES loss_reasons(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_budgets_year_budget ON budgets(year_budget);
CREATE INDEX IF NOT EXISTS idx_budgets_status_id ON budgets(status_id);
CREATE INDEX IF NOT EXISTS idx_budgets_sent_at ON budgets(sent_at DESC);

-- migrate:down
DROP TABLE IF EXISTS budgets;
