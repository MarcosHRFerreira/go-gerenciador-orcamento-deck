package budgetstatushistory

import (
	"context"
	"database/sql"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
)

type Repository interface {
	Create(ctx context.Context, item *model.BudgetStatusHistoryModel) (int64, error)
	ListByBudgetID(ctx context.Context, budgetID int64) ([]model.BudgetStatusHistoryModel, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, item *model.BudgetStatusHistoryModel) (int64, error) {
	const query = `
		INSERT INTO budget_status_history (
			budget_id,
			from_status_id,
			to_status_id,
			changed_by_user_id,
			notes,
			changed_at,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRowContext(
		ctx,
		query,
		item.BudgetID,
		nullableStatusID(item.FromStatusID),
		item.ToStatusID,
		item.ChangedByUserID,
		item.Notes,
		item.ChangedAt,
		item.CreatedAt,
		item.UpdatedAt,
	).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *repository) ListByBudgetID(ctx context.Context, budgetID int64) ([]model.BudgetStatusHistoryModel, error) {
	const query = `
		SELECT id, budget_id, from_status_id, to_status_id, changed_by_user_id, notes, changed_at, created_at, updated_at
		FROM budget_status_history
		WHERE budget_id = $1
		ORDER BY changed_at DESC, id DESC
	`

	rows, err := r.db.QueryContext(ctx, query, budgetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.BudgetStatusHistoryModel, 0)
	for rows.Next() {
		var item model.BudgetStatusHistoryModel
		if err := rows.Scan(
			&item.ID,
			&item.BudgetID,
			&item.FromStatusID,
			&item.ToStatusID,
			&item.ChangedByUserID,
			&item.Notes,
			&item.ChangedAt,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func nullableStatusID(value sql.NullInt64) interface{} {
	if !value.Valid {
		return nil
	}

	return value.Int64
}
