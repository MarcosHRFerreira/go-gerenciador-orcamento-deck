-- migrate:up
CREATE TABLE IF NOT EXISTS budget_import_batches (
    id BIGSERIAL PRIMARY KEY,
    preview_id VARCHAR(100) NOT NULL DEFAULT '',
    file_name VARCHAR(255) NOT NULL,
    source_company VARCHAR(50) NOT NULL,
    source_layout VARCHAR(50) NOT NULL,
    status VARCHAR(30) NOT NULL,
    executed_by_user_id BIGINT NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at TIMESTAMPTZ NULL,
    rows_processed INTEGER NOT NULL DEFAULT 0,
    budgets_created INTEGER NOT NULL DEFAULT 0,
    budgets_updated INTEGER NOT NULL DEFAULT 0,
    budgets_ignored INTEGER NOT NULL DEFAULT 0,
    rows_failed INTEGER NOT NULL DEFAULT 0,
    catalogs_created INTEGER NOT NULL DEFAULT 0,
    result_message TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_budget_import_batches_executed_by_user_id
        FOREIGN KEY (executed_by_user_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_budget_import_batches_started_at
    ON budget_import_batches(started_at DESC);
CREATE INDEX IF NOT EXISTS idx_budget_import_batches_source_layout
    ON budget_import_batches(source_layout);

CREATE TABLE IF NOT EXISTS budget_import_rows_raw (
    id BIGSERIAL PRIMARY KEY,
    import_batch_id BIGINT NOT NULL,
    row_number INTEGER NOT NULL,
    budget_number VARCHAR(50) NOT NULL DEFAULT '',
    status VARCHAR(30) NOT NULL,
    action VARCHAR(30) NOT NULL DEFAULT '',
    raw_row_data JSONB NOT NULL DEFAULT '{}'::jsonb,
    normalized_row_data JSONB NULL,
    messages JSONB NOT NULL DEFAULT '[]'::jsonb,
    budget_id BIGINT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_budget_import_rows_raw_import_batch_id
        FOREIGN KEY (import_batch_id) REFERENCES budget_import_batches(id) ON DELETE CASCADE,
    CONSTRAINT fk_budget_import_rows_raw_budget_id
        FOREIGN KEY (budget_id) REFERENCES budgets(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_budget_import_rows_raw_import_batch_id
    ON budget_import_rows_raw(import_batch_id);
CREATE INDEX IF NOT EXISTS idx_budget_import_rows_raw_row_number
    ON budget_import_rows_raw(import_batch_id, row_number);

ALTER TABLE budgets
    ADD COLUMN IF NOT EXISTS source_company VARCHAR(50) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS source_layout VARCHAR(50) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS import_batch_id BIGINT NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_budgets_import_batch_id'
    ) THEN
        ALTER TABLE budgets
            ADD CONSTRAINT fk_budgets_import_batch_id
            FOREIGN KEY (import_batch_id) REFERENCES budget_import_batches(id) ON DELETE SET NULL;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_budgets_source_layout ON budgets(source_layout);
CREATE INDEX IF NOT EXISTS idx_budgets_import_batch_id ON budgets(import_batch_id);

-- migrate:down
DROP INDEX IF EXISTS idx_budgets_import_batch_id;
DROP INDEX IF EXISTS idx_budgets_source_layout;
ALTER TABLE budgets DROP CONSTRAINT IF EXISTS fk_budgets_import_batch_id;
ALTER TABLE budgets
    DROP COLUMN IF EXISTS import_batch_id,
    DROP COLUMN IF EXISTS source_layout,
    DROP COLUMN IF EXISTS source_company;

DROP INDEX IF EXISTS idx_budget_import_rows_raw_row_number;
DROP INDEX IF EXISTS idx_budget_import_rows_raw_import_batch_id;
DROP TABLE IF EXISTS budget_import_rows_raw;

DROP INDEX IF EXISTS idx_budget_import_batches_source_layout;
DROP INDEX IF EXISTS idx_budget_import_batches_started_at;
DROP TABLE IF EXISTS budget_import_batches;
