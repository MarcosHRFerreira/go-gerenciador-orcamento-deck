-- migrate:up
INSERT INTO priorities (code, name, weight, created_at, updated_at)
VALUES
    ('faixa_0_a_50k', 'Faixa 0 a 50k', 10, NOW(), NOW()),
    ('faixa_50k_a_250k', 'Faixa 50k a 250k', 20, NOW(), NOW()),
    ('faixa_acima_de_250k', 'Faixa acima de 250k', 30, NOW(), NOW())
ON CONFLICT (code) DO UPDATE
SET
    name = EXCLUDED.name,
    weight = EXCLUDED.weight,
    updated_at = NOW();

UPDATE budgets AS b
SET
    priority_id = p.id,
    updated_at = NOW()
FROM priorities AS p
WHERE p.code = CASE
    WHEN b.gross_value <= 50000 THEN 'faixa_0_a_50k'
    WHEN b.gross_value <= 250000 THEN 'faixa_50k_a_250k'
    ELSE 'faixa_acima_de_250k'
END
AND b.gross_value > 0
AND b.priority_id IS DISTINCT FROM p.id;

-- migrate:down
UPDATE budgets
SET
    priority_id = NULL,
    updated_at = NOW()
WHERE priority_id IN (
    SELECT id
    FROM priorities
    WHERE code IN (
        'faixa_0_a_50k',
        'faixa_50k_a_250k',
        'faixa_acima_de_250k'
    )
);

DELETE FROM priorities
WHERE code IN (
    'faixa_0_a_50k',
    'faixa_50k_a_250k',
    'faixa_acima_de_250k'
);
