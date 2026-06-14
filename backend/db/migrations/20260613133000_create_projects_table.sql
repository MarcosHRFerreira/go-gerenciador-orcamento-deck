-- migrate:up
CREATE TABLE IF NOT EXISTS projects (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    project_type_id BIGINT NULL,
    city VARCHAR(100) NOT NULL DEFAULT '',
    state VARCHAR(50) NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_projects_project_type_id
        FOREIGN KEY (project_type_id) REFERENCES project_types(id) ON DELETE SET NULL
);

-- migrate:down
DROP TABLE IF EXISTS projects;
