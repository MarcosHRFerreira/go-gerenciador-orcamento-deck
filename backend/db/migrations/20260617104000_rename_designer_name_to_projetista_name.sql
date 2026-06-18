-- migrate:up
ALTER TABLE budgets
    RENAME COLUMN designer_name TO projetista_name;

-- migrate:down
ALTER TABLE budgets
    RENAME COLUMN projetista_name TO designer_name;
