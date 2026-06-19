-- migrate:up
ALTER TABLE conversations
    ADD COLUMN IF NOT EXISTS project_id BIGINT NULL;

ALTER TABLE conversations
    ADD CONSTRAINT fk_conversations_project_id
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_conversations_project_id
    ON conversations(project_id);

-- migrate:down
DROP INDEX IF EXISTS idx_conversations_project_id;

ALTER TABLE conversations
    DROP CONSTRAINT IF EXISTS fk_conversations_project_id;

ALTER TABLE conversations
    DROP COLUMN IF EXISTS project_id;
