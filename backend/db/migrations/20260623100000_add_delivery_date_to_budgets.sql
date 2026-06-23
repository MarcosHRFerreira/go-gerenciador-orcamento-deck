-- migrate:up
ALTER TABLE budgets
    ADD COLUMN IF NOT EXISTS delivery_date DATE NULL;

CREATE INDEX IF NOT EXISTS idx_budgets_delivery_date
    ON budgets(delivery_date);

-- migrate:down
DROP INDEX IF EXISTS idx_budgets_delivery_date;

ALTER TABLE budgets
    DROP COLUMN IF EXISTS delivery_date;
