-- migrate:up
ALTER TABLE budgets
    ADD COLUMN IF NOT EXISTS product_line_id BIGINT NULL,
    ADD COLUMN IF NOT EXISTS construction_company VARCHAR(200) NOT NULL DEFAULT '';

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_budgets_product_line_id'
    ) THEN
        ALTER TABLE budgets
            ADD CONSTRAINT fk_budgets_product_line_id
                FOREIGN KEY (product_line_id) REFERENCES product_lines(id) ON DELETE SET NULL;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_budgets_product_line_id ON budgets(product_line_id);

-- migrate:down
DROP INDEX IF EXISTS idx_budgets_product_line_id;

ALTER TABLE budgets
    DROP CONSTRAINT IF EXISTS fk_budgets_product_line_id;

ALTER TABLE budgets
    DROP COLUMN IF EXISTS construction_company,
    DROP COLUMN IF EXISTS product_line_id;
