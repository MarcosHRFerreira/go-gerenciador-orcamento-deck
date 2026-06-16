-- migrate:up
ALTER TABLE budget_import_batches
    ADD COLUMN IF NOT EXISTS rows_expected INTEGER NOT NULL DEFAULT 0;

-- migrate:down
ALTER TABLE budget_import_batches
    DROP COLUMN IF EXISTS rows_expected;
