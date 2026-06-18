-- migrate:up
ALTER TABLE projects
    ADD COLUMN IF NOT EXISTS code VARCHAR(50);

UPDATE projects
SET code = 'PROJECT_' || id
WHERE COALESCE(BTRIM(code), '') = '';

ALTER TABLE projects
    ALTER COLUMN code SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_projects_code ON projects(code);

-- migrate:down
DROP INDEX IF EXISTS uq_projects_code;

ALTER TABLE projects
    DROP COLUMN IF EXISTS code;
