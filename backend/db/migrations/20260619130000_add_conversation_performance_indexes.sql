-- migrate:up
CREATE INDEX IF NOT EXISTS idx_conversations_type_project_id_id
    ON conversations(type, project_id, id);

CREATE INDEX IF NOT EXISTS idx_conversation_participants_user_conversation_last_read
    ON conversation_participants(user_id, conversation_id, last_read_message_id);

-- migrate:down
DROP INDEX IF EXISTS idx_conversation_participants_user_conversation_last_read;
DROP INDEX IF EXISTS idx_conversations_type_project_id_id;
