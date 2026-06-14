package budgetfollowup

import (
	"context"
	"database/sql"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
)

type Repository interface {
	Create(ctx context.Context, item *model.BudgetFollowUpModel) (int64, error)
	ListByBudgetID(ctx context.Context, budgetID int64) ([]model.BudgetFollowUpModel, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, item *model.BudgetFollowUpModel) (int64, error) {
	const query = `
		INSERT INTO budget_follow_ups (
			budget_id,
			created_by_user_id,
			notes,
			follow_up_at,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRowContext(
		ctx,
		query,
		item.BudgetID,
		item.CreatedByUserID,
		item.Notes,
		item.FollowUpAt,
		item.CreatedAt,
		item.UpdatedAt,
	).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *repository) ListByBudgetID(ctx context.Context, budgetID int64) ([]model.BudgetFollowUpModel, error) {
	const query = `
		SELECT id, budget_id, created_by_user_id, notes, follow_up_at, created_at, updated_at
		FROM budget_follow_ups
		WHERE budget_id = $1
		ORDER BY follow_up_at DESC, id DESC
	`

	rows, err := r.db.QueryContext(ctx, query, budgetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.BudgetFollowUpModel, 0)
	for rows.Next() {
		var item model.BudgetFollowUpModel
		if err := rows.Scan(
			&item.ID,
			&item.BudgetID,
			&item.CreatedByUserID,
			&item.Notes,
			&item.FollowUpAt,
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
