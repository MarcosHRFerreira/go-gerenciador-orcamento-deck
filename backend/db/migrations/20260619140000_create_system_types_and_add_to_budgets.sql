-- migrate:up
CREATE TABLE IF NOT EXISTS system_types (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(150) NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_system_types_name ON system_types(name);

ALTER TABLE budgets
    ADD COLUMN IF NOT EXISTS system_type_id BIGINT NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_budgets_system_type_id'
    ) THEN
        ALTER TABLE budgets
            ADD CONSTRAINT fk_budgets_system_type_id
                FOREIGN KEY (system_type_id) REFERENCES system_types(id) ON DELETE SET NULL;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_budgets_system_type_id ON budgets(system_type_id);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM system_types
        WHERE code = 'VRF' OR name = 'VRF'
    ) THEN
        INSERT INTO system_types (code, name, description)
        VALUES ('VRF', 'VRF', '');
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM system_types
        WHERE code = 'AGUA_GELADA' OR name = 'Agua Gelada'
    ) THEN
        INSERT INTO system_types (code, name, description)
        VALUES ('AGUA_GELADA', 'Agua Gelada', '');
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM system_types
        WHERE code = 'EXAUSTAO' OR name = 'Exaustao'
    ) THEN
        INSERT INTO system_types (code, name, description)
        VALUES ('EXAUSTAO', 'Exaustao', '');
    END IF;
END $$;

-- migrate:down
DROP INDEX IF EXISTS idx_budgets_system_type_id;
DROP INDEX IF EXISTS idx_system_types_name;

ALTER TABLE budgets
    DROP CONSTRAINT IF EXISTS fk_budgets_system_type_id;

ALTER TABLE budgets
    DROP COLUMN IF EXISTS system_type_id;

DROP TABLE IF EXISTS system_types;
