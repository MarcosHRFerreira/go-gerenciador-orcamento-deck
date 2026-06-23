package conversation

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
)

type Repository interface {
	CreateOrAppendDirect(ctx context.Context, actorUserID int64, participantUserID int64, projectID *int64, initialMessage string, now time.Time) (int64, int64, error)
	ListByUser(ctx context.Context, userID int64) ([]model.ConversationListItemModel, error)
	ListMessagesByConversation(ctx context.Context, conversationID int64, userID int64) ([]model.ConversationMessageDetailsModel, error)
	CountUnreadByUser(ctx context.Context, userID int64) (int64, error)
	SendMessage(ctx context.Context, conversationID int64, senderUserID int64, body string, now time.Time) (int64, error)
	MarkAsRead(ctx context.Context, conversationID int64, userID int64, now time.Time) error
}

type repository struct {
	db *sql.DB
}

type executor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateOrAppendDirect(ctx context.Context, actorUserID int64, participantUserID int64, projectID *int64, initialMessage string, now time.Time) (int64, int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, 0, err
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	conversationID, err := findDirectConversationID(ctx, tx, actorUserID, participantUserID, projectID)
	if err != nil {
		return 0, 0, err
	}

	if conversationID == 0 {
		conversationID, err = insertConversation(ctx, tx, actorUserID, projectID, now)
		if err != nil {
			return 0, 0, err
		}

		if err := insertConversationParticipant(ctx, tx, conversationID, actorUserID, now); err != nil {
			return 0, 0, err
		}
		if err := insertConversationParticipant(ctx, tx, conversationID, participantUserID, now); err != nil {
			return 0, 0, err
		}
	}

	messageID, err := insertConversationMessage(ctx, tx, conversationID, actorUserID, initialMessage, now)
	if err != nil {
		return 0, 0, err
	}

	if err := updateConversationParticipantLastRead(ctx, tx, conversationID, actorUserID, &messageID, now); err != nil {
		return 0, 0, err
	}
	if err := updateConversationUpdatedAt(ctx, tx, conversationID, now); err != nil {
		return 0, 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, err
	}

	committed = true
	return conversationID, messageID, nil
}

func (r *repository) ListByUser(ctx context.Context, userID int64) ([]model.ConversationListItemModel, error) {
	const query = `
		SELECT
			c.id,
			c.type,
			c.updated_at,
			p.id,
			p.code,
			p.name,
			MAX(CASE WHEN cp.user_id <> $1 THEN u.id END) AS participant_user_id,
			MAX(CASE WHEN cp.user_id <> $1 THEN u.name END) AS participant_name,
			MAX(CASE WHEN cp.user_id <> $1 THEN u.username END) AS participant_username,
			MAX(CASE WHEN cp.user_id <> $1 THEN u.role END) AS participant_role,
			lm.id,
			lm.body,
			lm.created_at,
			lm.sender_user_id,
			COALESCE(unread.unread_count, 0) AS unread_count
		FROM conversations c
		LEFT JOIN projects p ON p.id = c.project_id
		INNER JOIN conversation_participants self_cp
			ON self_cp.conversation_id = c.id
			AND self_cp.user_id = $1
		INNER JOIN conversation_participants cp ON cp.conversation_id = c.id
		INNER JOIN users u ON u.id = cp.user_id
		LEFT JOIN LATERAL (
			SELECT id, body, created_at, sender_user_id
			FROM conversation_messages
			WHERE conversation_id = c.id
			ORDER BY id DESC
			LIMIT 1
		) lm ON TRUE
		LEFT JOIN LATERAL (
			SELECT COUNT(*) AS unread_count
			FROM conversation_messages cm
			WHERE cm.conversation_id = c.id
				AND cm.sender_user_id <> $1
				AND (
					self_cp.last_read_message_id IS NULL
					OR cm.id > self_cp.last_read_message_id
				)
		) unread ON TRUE
		WHERE c.type = 'direct'
		GROUP BY
			c.id,
			c.type,
			c.updated_at,
			p.id,
			p.code,
			p.name,
			self_cp.last_read_message_id,
			lm.id,
			lm.body,
			lm.created_at,
			lm.sender_user_id,
			unread.unread_count
		ORDER BY COALESCE(lm.created_at, c.updated_at) DESC, c.id DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.ConversationListItemModel, 0)
	for rows.Next() {
		item, err := scanConversationListItem(rows)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *repository) ListMessagesByConversation(ctx context.Context, conversationID int64, userID int64) ([]model.ConversationMessageDetailsModel, error) {
	const query = `
		SELECT
			m.id,
			m.conversation_id,
			m.sender_user_id,
			u.name,
			u.username,
			u.role,
			m.body,
			m.created_at,
			m.updated_at
		FROM conversation_messages m
		INNER JOIN users u ON u.id = m.sender_user_id
		WHERE EXISTS (
			SELECT 1
			FROM conversation_participants cp
			WHERE cp.conversation_id = m.conversation_id
				AND cp.user_id = $2
		)
			AND m.conversation_id = $1
		ORDER BY m.id ASC
	`

	rows, err := r.db.QueryContext(ctx, query, conversationID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.ConversationMessageDetailsModel, 0)
	for rows.Next() {
		item, err := scanConversationMessageDetails(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}
func (r *repository) CountUnreadByUser(ctx context.Context, userID int64) (int64, error) {
	const query = `
		SELECT COUNT(*)
		FROM conversations c
		INNER JOIN conversation_participants cp
			ON cp.conversation_id = c.id
			AND cp.user_id = $1
		WHERE EXISTS (
			SELECT 1
			FROM conversation_messages cm
			WHERE cm.conversation_id = c.id
				AND cm.sender_user_id <> $1
				AND (
					cp.last_read_message_id IS NULL
					OR cm.id > cp.last_read_message_id
				)
		)
	`

	var count int64
	if err := r.db.QueryRowContext(ctx, query, userID).Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func (r *repository) SendMessage(ctx context.Context, conversationID int64, senderUserID int64, body string, now time.Time) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	isParticipant, err := isConversationParticipant(ctx, tx, conversationID, senderUserID)
	if err != nil {
		return 0, err
	}
	if !isParticipant {
		return 0, sql.ErrNoRows
	}

	messageID, err := insertConversationMessage(ctx, tx, conversationID, senderUserID, body, now)
	if err != nil {
		return 0, err
	}

	if err := updateConversationParticipantLastRead(ctx, tx, conversationID, senderUserID, &messageID, now); err != nil {
		return 0, err
	}
	if err := updateConversationUpdatedAt(ctx, tx, conversationID, now); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	committed = true
	return messageID, nil
}

func (r *repository) MarkAsRead(ctx context.Context, conversationID int64, userID int64, now time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	isParticipant, err := isConversationParticipant(ctx, tx, conversationID, userID)
	if err != nil {
		return err
	}
	if !isParticipant {
		return sql.ErrNoRows
	}

	latestMessageID, err := latestConversationMessageID(ctx, tx, conversationID)
	if err != nil {
		return err
	}

	if err := updateConversationParticipantLastRead(ctx, tx, conversationID, userID, latestMessageID, now); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	committed = true
	return nil
}

func findDirectConversationID(ctx context.Context, db executor, userAID int64, userBID int64, projectID *int64) (int64, error) {
	const query = `
		SELECT c.id
		FROM conversations c
		INNER JOIN conversation_participants cp
			ON cp.conversation_id = c.id
		WHERE c.type = 'direct'
			AND c.project_id IS NOT DISTINCT FROM $3
			AND cp.user_id IN ($1, $2)
		GROUP BY c.id
		HAVING COUNT(DISTINCT cp.user_id) = 2
			AND (
				SELECT COUNT(*)
				FROM conversation_participants all_cp
				WHERE all_cp.conversation_id = c.id
			) = 2
		LIMIT 1
	`

	var conversationID int64
	err := db.QueryRowContext(ctx, query, userAID, userBID, projectID).Scan(&conversationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}

	return conversationID, nil
}

func insertConversation(ctx context.Context, db executor, createdByUserID int64, projectID *int64, now time.Time) (int64, error) {
	const query = `
		INSERT INTO conversations (type, created_by_user_id, project_id, created_at, updated_at)
		VALUES ('direct', $1, $2, $3, $4)
		RETURNING id
	`

	var conversationID int64
	err := db.QueryRowContext(ctx, query, createdByUserID, projectID, now, now).Scan(&conversationID)
	if err != nil {
		return 0, err
	}

	return conversationID, nil
}

func insertConversationParticipant(ctx context.Context, db executor, conversationID int64, userID int64, now time.Time) error {
	const query = `
		INSERT INTO conversation_participants (
			conversation_id,
			user_id,
			joined_at,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5)
	`

	_, err := db.ExecContext(ctx, query, conversationID, userID, now, now, now)
	return err
}

func insertConversationMessage(ctx context.Context, db executor, conversationID int64, senderUserID int64, body string, now time.Time) (int64, error) {
	const query = `
		INSERT INTO conversation_messages (
			conversation_id,
			sender_user_id,
			body,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	var messageID int64
	err := db.QueryRowContext(ctx, query, conversationID, senderUserID, body, now, now).Scan(&messageID)
	if err != nil {
		return 0, err
	}

	return messageID, nil
}

func updateConversationParticipantLastRead(ctx context.Context, db executor, conversationID int64, userID int64, messageID *int64, now time.Time) error {
	const query = `
		UPDATE conversation_participants
		SET last_read_message_id = $3,
			updated_at = $4
		WHERE conversation_id = $1
			AND user_id = $2
	`

	_, err := db.ExecContext(ctx, query, conversationID, userID, messageID, now)
	return err
}

func updateConversationUpdatedAt(ctx context.Context, db executor, conversationID int64, now time.Time) error {
	const query = `
		UPDATE conversations
		SET updated_at = $2
		WHERE id = $1
	`

	_, err := db.ExecContext(ctx, query, conversationID, now)
	return err
}

func latestConversationMessageID(ctx context.Context, db executor, conversationID int64) (*int64, error) {
	const query = `
		SELECT id
		FROM conversation_messages
		WHERE conversation_id = $1
		ORDER BY id DESC
		LIMIT 1
	`

	var messageID int64
	err := db.QueryRowContext(ctx, query, conversationID).Scan(&messageID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &messageID, nil
}

func isConversationParticipant(ctx context.Context, db executor, conversationID int64, userID int64) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM conversation_participants
			WHERE conversation_id = $1
				AND user_id = $2
		)
	`

	var exists bool
	if err := db.QueryRowContext(ctx, query, conversationID, userID).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

type conversationListScanner interface {
	Scan(dest ...interface{}) error
}

func scanConversationListItem(scanner conversationListScanner) (model.ConversationListItemModel, error) {
	var item model.ConversationListItemModel
	var projectID sql.NullInt64
	var projectCode sql.NullString
	var projectName sql.NullString
	var lastMessageID sql.NullInt64
	var lastMessageBody sql.NullString
	var lastMessageAt sql.NullTime
	var lastMessageSenderID sql.NullInt64

	err := scanner.Scan(
		&item.ConversationID,
		&item.Type,
		&item.UpdatedAt,
		&projectID,
		&projectCode,
		&projectName,
		&item.ParticipantUserID,
		&item.ParticipantName,
		&item.ParticipantUsername,
		&item.ParticipantRole,
		&lastMessageID,
		&lastMessageBody,
		&lastMessageAt,
		&lastMessageSenderID,
		&item.UnreadCount,
	)
	if err != nil {
		return model.ConversationListItemModel{}, err
	}

	item.ProjectID = nullableInt64Pointer(projectID)
	item.ProjectCode = nullableStringPointer(projectCode)
	item.ProjectName = nullableStringPointer(projectName)
	item.LastMessageID = nullableInt64Pointer(lastMessageID)
	item.LastMessageBody = nullableStringPointer(lastMessageBody)
	item.LastMessageAt = nullableTimePointer(lastMessageAt)
	item.LastMessageSenderID = nullableInt64Pointer(lastMessageSenderID)

	return item, nil
}

type conversationMessageScanner interface {
	Scan(dest ...interface{}) error
}

func scanConversationMessageDetails(scanner conversationMessageScanner) (model.ConversationMessageDetailsModel, error) {
	var item model.ConversationMessageDetailsModel
	err := scanner.Scan(
		&item.MessageID,
		&item.ConversationID,
		&item.SenderUserID,
		&item.SenderName,
		&item.SenderUsername,
		&item.SenderRole,
		&item.Body,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return model.ConversationMessageDetailsModel{}, err
	}

	return item, nil
}

func nullableInt64Pointer(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}

	result := value.Int64
	return &result
}

func nullableStringPointer(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}

	result := value.String
	return &result
}

func nullableTimePointer(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}

	result := value.Time
	return &result
}
