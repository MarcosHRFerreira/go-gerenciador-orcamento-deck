-- migrate:up
UPDATE budgets
SET
    source_company = 'Rocktec',
    source_layout = 'rocktec',
    updated_at = NOW()
WHERE btrim(source_company) = ''
  AND btrim(source_layout) = '';

-- migrate:down
SELECT 1;
