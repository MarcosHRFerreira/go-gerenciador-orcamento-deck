package deliveryalert

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
)

type Repository interface {
	ListDueInTwoDaysCandidates(ctx context.Context, referenceDate time.Time) ([]model.DeliveryAlertCandidateModel, error)
	CreateEvent(ctx context.Context, item *model.DeliveryAlertEventModel) (int64, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) ListDueInTwoDaysCandidates(ctx context.Context, referenceDate time.Time) ([]model.DeliveryAlertCandidateModel, error) {
	const query = `
		SELECT
			b.id,
			b.budget_number,
			p.id,
			p.code,
			p.name,
			s.id,
			s.name,
			s.email,
			u.id,
			u.name,
			u.username,
			b.delivery_date::timestamp,
			(b.delivery_date - $1::date)::int,
			b.construction_company
		FROM budgets b
		INNER JOIN budget_statuses bs ON bs.id = b.status_id
		INNER JOIN salespeople s ON s.id = b.salesperson_id
		INNER JOIN users u
			ON u.active = TRUE
			AND u.user_kind = 'salesperson'
			AND (
				LOWER(u.email) = LOWER(s.email)
				OR LOWER(u.username) = LOWER(split_part(s.email, '@', 1))
			)
		LEFT JOIN projects p ON p.id = b.project_id
		WHERE b.delivery_date = ($1::date + 2)
			AND (
				UPPER(COALESCE(bs.code, '')) = 'PEDIDO'
				OR LOWER(COALESCE(bs.name, '')) IN ('pedido', 'fechado')
			)
			AND b.salesperson_id IS NOT NULL
			AND NOT EXISTS (
				SELECT 1
				FROM delivery_alert_events dae
				WHERE dae.budget_id = b.id
					AND dae.alert_type = $2
					AND dae.delivery_date = b.delivery_date
			)
		ORDER BY b.delivery_date ASC, b.id ASC
	`

	rows, err := r.db.QueryContext(ctx, query, referenceDate, model.DeliveryAlertTypeDueInTwoDays)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.DeliveryAlertCandidateModel, 0)
	for rows.Next() {
		var item model.DeliveryAlertCandidateModel
		var projectID sql.NullInt64
		var projectCode sql.NullString
		var projectName sql.NullString
		if err := rows.Scan(
			&item.BudgetID,
			&item.BudgetNumber,
			&projectID,
			&projectCode,
			&projectName,
			&item.SalespersonID,
			&item.SalespersonName,
			&item.SalespersonEmail,
			&item.RecipientUserID,
			&item.RecipientUserName,
			&item.RecipientUsername,
			&item.DeliveryDate,
			&item.DaysUntilDelivery,
			&item.ConstructionCompany,
		); err != nil {
			return nil, err
		}

		if projectID.Valid {
			projectIDValue := projectID.Int64
			item.ProjectID = &projectIDValue
		}
		if projectCode.Valid {
			projectCodeValue := projectCode.String
			item.ProjectCode = &projectCodeValue
		}
		if projectName.Valid {
			projectNameValue := projectName.String
			item.ProjectName = &projectNameValue
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *repository) CreateEvent(ctx context.Context, item *model.DeliveryAlertEventModel) (int64, error) {
	const query = `
		INSERT INTO delivery_alert_events (
			budget_id,
			recipient_user_id,
			conversation_id,
			message_id,
			alert_type,
			delivery_date,
			sent_at,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	var eventID int64
	if err := r.db.QueryRowContext(
		ctx,
		query,
		item.BudgetID,
		item.RecipientUserID,
		item.ConversationID,
		item.MessageID,
		item.AlertType,
		item.DeliveryDate,
		item.SentAt,
		item.CreatedAt,
		item.UpdatedAt,
	).Scan(&eventID); err != nil {
		return 0, err
	}

	return eventID, nil
}

func IsDuplicateEventError(err error) bool {
	if err == nil {
		return false
	}

	var pqError interface{ SQLState() string }
	if errors.As(err, &pqError) {
		return pqError.SQLState() == "23505"
	}

	return false
}
