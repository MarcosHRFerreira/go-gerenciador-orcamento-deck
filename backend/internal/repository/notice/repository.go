package notice

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
)

type Repository interface {
	Create(ctx context.Context, notice *model.NoticeModel, recipientUserIDs []int64) (int64, error)
	ListByUser(ctx context.Context, userID int64, statusFilter string, now time.Time) ([]model.NoticeListItemModel, error)
	GetByIDForUser(ctx context.Context, noticeID int64, userID int64, now time.Time) (*model.NoticeListItemModel, error)
	CountUnreadByUser(ctx context.Context, userID int64, now time.Time) (int64, error)
	MarkAsRead(ctx context.Context, noticeID int64, userID int64, readAt time.Time) error
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

func (r *repository) Create(ctx context.Context, notice *model.NoticeModel, recipientUserIDs []int64) (int64, error) {
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

	noticeID, err := insertNotice(ctx, tx, notice)
	if err != nil {
		return 0, err
	}

	for _, recipientUserID := range recipientUserIDs {
		if err := insertNoticeRecipient(ctx, tx, noticeID, recipientUserID, notice.CreatedAt); err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	committed = true
	return noticeID, nil
}

func (r *repository) ListByUser(ctx context.Context, userID int64, statusFilter string, now time.Time) ([]model.NoticeListItemModel, error) {
	args := []interface{}{userID, now}
	conditions := []string{
		"nr.user_id = $1",
		"nr.hidden_at IS NULL",
		"(n.expires_at IS NULL OR n.expires_at > $2)",
	}

	normalizedStatusFilter := strings.TrimSpace(strings.ToLower(statusFilter))
	switch normalizedStatusFilter {
	case "unread":
		conditions = append(conditions, "nr.read_at IS NULL")
	case "read":
		conditions = append(conditions, "nr.read_at IS NOT NULL")
	}

	query := fmt.Sprintf(`
		SELECT
			n.id,
			n.title,
			n.body,
			n.scope_type,
			n.priority,
			n.pinned,
			n.expires_at,
			n.created_by_user_id,
			n.created_at,
			n.updated_at,
			nr.id,
			nr.read_at,
			u.name
		FROM notices n
		INNER JOIN notice_recipients nr ON nr.notice_id = n.id
		INNER JOIN users u ON u.id = n.created_by_user_id
		WHERE %s
		ORDER BY n.pinned DESC, n.created_at DESC, n.id DESC
	`, strings.Join(conditions, " AND "))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.NoticeListItemModel, 0)
	for rows.Next() {
		item, err := scanNoticeListItem(rows)
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

func (r *repository) GetByIDForUser(ctx context.Context, noticeID int64, userID int64, now time.Time) (*model.NoticeListItemModel, error) {
	const query = `
		SELECT
			n.id,
			n.title,
			n.body,
			n.scope_type,
			n.priority,
			n.pinned,
			n.expires_at,
			n.created_by_user_id,
			n.created_at,
			n.updated_at,
			nr.id,
			nr.read_at,
			u.name
		FROM notices n
		INNER JOIN notice_recipients nr ON nr.notice_id = n.id
		INNER JOIN users u ON u.id = n.created_by_user_id
		WHERE n.id = $1
			AND nr.user_id = $2
			AND nr.hidden_at IS NULL
			AND (n.expires_at IS NULL OR n.expires_at > $3)
	`

	row := r.db.QueryRowContext(ctx, query, noticeID, userID, now)

	item, err := scanNoticeListItem(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &item, nil
}

func (r *repository) CountUnreadByUser(ctx context.Context, userID int64, now time.Time) (int64, error) {
	const query = `
		SELECT COUNT(*)
		FROM notice_recipients nr
		INNER JOIN notices n ON n.id = nr.notice_id
		WHERE nr.user_id = $1
			AND nr.hidden_at IS NULL
			AND nr.read_at IS NULL
			AND (n.expires_at IS NULL OR n.expires_at > $2)
	`

	var count int64
	if err := r.db.QueryRowContext(ctx, query, userID, now).Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func (r *repository) MarkAsRead(ctx context.Context, noticeID int64, userID int64, readAt time.Time) error {
	const query = `
		UPDATE notice_recipients
		SET read_at = COALESCE(read_at, $3), updated_at = $3
		WHERE notice_id = $1
			AND user_id = $2
			AND hidden_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, query, noticeID, userID, readAt)
	return err
}

func insertNotice(ctx context.Context, db executor, notice *model.NoticeModel) (int64, error) {
	const query = `
		INSERT INTO notices (
			title,
			body,
			scope_type,
			priority,
			pinned,
			expires_at,
			created_by_user_id,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	var noticeID int64
	err := db.QueryRowContext(
		ctx,
		query,
		notice.Title,
		notice.Body,
		notice.ScopeType,
		notice.Priority,
		notice.Pinned,
		nullableTime(notice.ExpiresAt),
		notice.CreatedByUserID,
		notice.CreatedAt,
		notice.UpdatedAt,
	).Scan(&noticeID)
	if err != nil {
		return 0, err
	}

	return noticeID, nil
}

func insertNoticeRecipient(ctx context.Context, db executor, noticeID int64, userID int64, createdAt time.Time) error {
	const query = `
		INSERT INTO notice_recipients (
			notice_id,
			user_id,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4)
	`

	_, err := db.ExecContext(ctx, query, noticeID, userID, createdAt, createdAt)
	return err
}

type noticeScanner interface {
	Scan(dest ...interface{}) error
}

func scanNoticeListItem(scanner noticeScanner) (model.NoticeListItemModel, error) {
	var item model.NoticeListItemModel
	var expiresAt sql.NullTime
	var readAt sql.NullTime

	err := scanner.Scan(
		&item.ID,
		&item.Title,
		&item.Body,
		&item.ScopeType,
		&item.Priority,
		&item.Pinned,
		&expiresAt,
		&item.CreatedByUserID,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.RecipientID,
		&readAt,
		&item.CreatedByUserName,
	)
	if err != nil {
		return model.NoticeListItemModel{}, err
	}

	item.ExpiresAt = nullableTimePointer(expiresAt)
	item.ReadAt = nullableTimePointer(readAt)

	return item, nil
}

func nullableTime(value *time.Time) interface{} {
	if value == nil {
		return nil
	}

	return *value
}

func nullableTimePointer(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}

	result := value.Time
	return &result
}
