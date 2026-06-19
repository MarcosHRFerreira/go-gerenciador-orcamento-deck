package budgetstatus

import (
	"context"
	"database/sql"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
)

type Repository interface {
	Create(ctx context.Context, status *model.BudgetStatusModel) (int64, error)
	List(ctx context.Context) ([]model.BudgetStatusModel, error)
	GetByCodeOrName(ctx context.Context, code string, name string) (*model.BudgetStatusModel, error)
	GetByID(ctx context.Context, statusID int64) (*model.BudgetStatusModel, error)
	Update(ctx context.Context, status *model.BudgetStatusModel) error
	Delete(ctx context.Context, statusID int64) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, status *model.BudgetStatusModel) (int64, error) {
	const query = `
		INSERT INTO budget_statuses (code, name, description, is_final, sort_order, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRowContext(
		ctx,
		query,
		status.Code,
		status.Name,
		status.Description,
		status.IsFinal,
		status.SortOrder,
		status.CreatedAt,
		status.UpdatedAt,
	).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *repository) List(ctx context.Context) ([]model.BudgetStatusModel, error) {
	const query = `
		SELECT id, code, name, description, is_final, sort_order, created_at, updated_at
		FROM budget_statuses
		ORDER BY sort_order ASC, id ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.BudgetStatusModel, 0)
	for rows.Next() {
		var item model.BudgetStatusModel
		if err := rows.Scan(
			&item.ID,
			&item.Code,
			&item.Name,
			&item.Description,
			&item.IsFinal,
			&item.SortOrder,
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

func (r *repository) GetByCodeOrName(ctx context.Context, code string, name string) (*model.BudgetStatusModel, error) {
	const query = `
		SELECT id, code, name, description, is_final, sort_order, created_at, updated_at
		FROM budget_statuses
		WHERE UPPER(TRIM(code)) = UPPER(TRIM($1))
			OR UPPER(TRIM(name)) = UPPER(TRIM($2))
		ORDER BY id ASC
		LIMIT 1
	`

	row := r.db.QueryRowContext(ctx, query, code, name)

	var item model.BudgetStatusModel
	err := row.Scan(
		&item.ID,
		&item.Code,
		&item.Name,
		&item.Description,
		&item.IsFinal,
		&item.SortOrder,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &item, nil
}

func (r *repository) GetByID(ctx context.Context, statusID int64) (*model.BudgetStatusModel, error) {
	const query = `
		SELECT id, code, name, description, is_final, sort_order, created_at, updated_at
		FROM budget_statuses
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, statusID)

	var item model.BudgetStatusModel
	err := row.Scan(
		&item.ID,
		&item.Code,
		&item.Name,
		&item.Description,
		&item.IsFinal,
		&item.SortOrder,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &item, nil
}

func (r *repository) Update(ctx context.Context, status *model.BudgetStatusModel) error {
	const query = `
		UPDATE budget_statuses
		SET code = $2, name = $3, description = $4, is_final = $5, sort_order = $6, updated_at = $7
		WHERE id = $1
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		status.ID,
		status.Code,
		status.Name,
		status.Description,
		status.IsFinal,
		status.SortOrder,
		status.UpdatedAt,
	)

	return err
}

func (r *repository) Delete(ctx context.Context, statusID int64) error {
	const query = `
		DELETE FROM budget_statuses
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, statusID)
	return err
}
