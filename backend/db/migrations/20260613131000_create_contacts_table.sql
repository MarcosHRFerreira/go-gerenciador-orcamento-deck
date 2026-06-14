-- migrate:up
CREATE TABLE IF NOT EXISTS contacts (
    id BIGSERIAL PRIMARY KEY,
    installer_id BIGINT NOT NULL,
    name VARCHAR(150) NOT NULL,
    email VARCHAR(250) NOT NULL,
    phone VARCHAR(30) NOT NULL,
    role VARCHAR(100) NOT NULL DEFAULT '',
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_contacts_installer_id
        FOREIGN KEY (installer_id) REFERENCES installers(id) ON DELETE CASCADE,
    CONSTRAINT uq_contacts_installer_email UNIQUE (installer_id, email),
    CONSTRAINT uq_contacts_installer_phone UNIQUE (installer_id, phone)
);

-- migrate:down
DROP TABLE IF EXISTS contacts;
