-- migrate:up
CREATE TABLE IF NOT EXISTS delivery_alert_events (
    id BIGSERIAL PRIMARY KEY,
    budget_id BIGINT NOT NULL,
    recipient_user_id BIGINT NOT NULL,
    conversation_id BIGINT NOT NULL,
    message_id BIGINT NOT NULL,
    alert_type VARCHAR(50) NOT NULL,
    delivery_date DATE NOT NULL,
    sent_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_delivery_alert_events_budget_id
        FOREIGN KEY (budget_id) REFERENCES budgets(id) ON DELETE CASCADE,
    CONSTRAINT fk_delivery_alert_events_recipient_user_id
        FOREIGN KEY (recipient_user_id) REFERENCES users(id) ON DELETE RESTRICT,
    CONSTRAINT fk_delivery_alert_events_conversation_id
        FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE RESTRICT,
    CONSTRAINT fk_delivery_alert_events_message_id
        FOREIGN KEY (message_id) REFERENCES conversation_messages(id) ON DELETE RESTRICT,
    CONSTRAINT uq_delivery_alert_events_budget_type_date
        UNIQUE (budget_id, alert_type, delivery_date)
);

CREATE INDEX IF NOT EXISTS idx_delivery_alert_events_budget_id
    ON delivery_alert_events(budget_id);

CREATE INDEX IF NOT EXISTS idx_delivery_alert_events_recipient_user_id
    ON delivery_alert_events(recipient_user_id);

CREATE INDEX IF NOT EXISTS idx_delivery_alert_events_sent_at
    ON delivery_alert_events(sent_at DESC);

-- migrate:down
DROP INDEX IF EXISTS idx_delivery_alert_events_sent_at;
DROP INDEX IF EXISTS idx_delivery_alert_events_recipient_user_id;
DROP INDEX IF EXISTS idx_delivery_alert_events_budget_id;
DROP TABLE IF EXISTS delivery_alert_events;
