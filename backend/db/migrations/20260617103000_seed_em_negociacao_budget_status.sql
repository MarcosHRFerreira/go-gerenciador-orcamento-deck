-- migrate:up
INSERT INTO budget_statuses (
    code,
    name,
    description,
    is_final,
    sort_order,
    created_at,
    updated_at
)
SELECT
    'EM_NEGOCIACAO',
    'Em Negociacao',
    'Status inicial padrao para orcamentos em negociacao.',
    FALSE,
    10,
    NOW(),
    NOW()
WHERE NOT EXISTS (
    SELECT 1
    FROM budget_statuses
    WHERE code = 'EM_NEGOCIACAO'
       OR name = 'Em Negociacao'
);

-- migrate:down
DELETE FROM budget_statuses
WHERE code = 'EM_NEGOCIACAO'
  AND name = 'Em Negociacao';
