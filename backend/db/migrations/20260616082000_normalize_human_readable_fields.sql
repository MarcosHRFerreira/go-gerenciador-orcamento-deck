-- migrate:up
UPDATE salespeople
SET
    name = INITCAP(LOWER(REGEXP_REPLACE(BTRIM(name), '\s+', ' ', 'g'))),
    updated_at = NOW()
WHERE BTRIM(name) <> '';

UPDATE installers
SET
    name = INITCAP(LOWER(REGEXP_REPLACE(BTRIM(name), '\s+', ' ', 'g'))),
    updated_at = NOW()
WHERE BTRIM(name) <> '';

UPDATE contacts
SET
    name = INITCAP(LOWER(REGEXP_REPLACE(BTRIM(name), '\s+', ' ', 'g'))),
    role = INITCAP(LOWER(REGEXP_REPLACE(BTRIM(role), '\s+', ' ', 'g'))),
    updated_at = NOW()
WHERE BTRIM(name) <> ''
   OR BTRIM(role) <> '';

UPDATE project_types
SET
    name = INITCAP(LOWER(REGEXP_REPLACE(BTRIM(name), '\s+', ' ', 'g'))),
    description = INITCAP(LOWER(REGEXP_REPLACE(BTRIM(description), '\s+', ' ', 'g'))),
    updated_at = NOW()
WHERE BTRIM(name) <> ''
   OR BTRIM(description) <> '';

UPDATE projects
SET
    name = INITCAP(LOWER(REGEXP_REPLACE(BTRIM(name), '\s+', ' ', 'g'))),
    updated_at = NOW()
WHERE BTRIM(name) <> '';

UPDATE priorities
SET
    name = INITCAP(LOWER(REGEXP_REPLACE(BTRIM(name), '\s+', ' ', 'g'))),
    updated_at = NOW()
WHERE BTRIM(name) <> '';

UPDATE budget_statuses
SET
    name = INITCAP(LOWER(REGEXP_REPLACE(BTRIM(name), '\s+', ' ', 'g'))),
    description = INITCAP(LOWER(REGEXP_REPLACE(BTRIM(description), '\s+', ' ', 'g'))),
    updated_at = NOW()
WHERE BTRIM(name) <> ''
   OR BTRIM(description) <> '';

UPDATE loss_reasons
SET
    name = INITCAP(LOWER(REGEXP_REPLACE(BTRIM(name), '\s+', ' ', 'g'))),
    description = INITCAP(LOWER(REGEXP_REPLACE(BTRIM(description), '\s+', ' ', 'g'))),
    updated_at = NOW()
WHERE BTRIM(name) <> ''
   OR BTRIM(description) <> '';

UPDATE budgets
SET
    competitor_name = INITCAP(LOWER(REGEXP_REPLACE(BTRIM(competitor_name), '\s+', ' ', 'g'))),
    designer_name = INITCAP(LOWER(REGEXP_REPLACE(BTRIM(designer_name), '\s+', ' ', 'g'))),
    specification_details = INITCAP(LOWER(REGEXP_REPLACE(BTRIM(specification_details), '\s+', ' ', 'g'))),
    current_follow_up = INITCAP(LOWER(REGEXP_REPLACE(BTRIM(current_follow_up), '\s+', ' ', 'g'))),
    updated_at = NOW()
WHERE BTRIM(competitor_name) <> ''
   OR BTRIM(designer_name) <> ''
   OR BTRIM(specification_details) <> ''
   OR BTRIM(current_follow_up) <> '';

-- migrate:down
SELECT 1;
