-- migrate:up
WITH canonical_salespeople AS (
    SELECT
        id AS canonical_id,
        LOWER(BTRIM(name)) AS canonical_name
    FROM salespeople
    WHERE POSITION(' ' IN BTRIM(name)) = 0
),
duplicate_salespeople AS (
    SELECT
        duplicate.id AS duplicate_id,
        canonical.canonical_id
    FROM salespeople duplicate
    JOIN canonical_salespeople canonical
      ON LOWER(SPLIT_PART(BTRIM(duplicate.name), ' ', 1)) = canonical.canonical_name
    WHERE duplicate.id <> canonical.canonical_id
)
UPDATE budgets AS budget
SET
    salesperson_id = duplicate.canonical_id,
    updated_at = NOW()
FROM duplicate_salespeople AS duplicate
WHERE budget.salesperson_id = duplicate.duplicate_id;

-- migrate:down
SELECT 1;
