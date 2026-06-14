package priority

import (
	"context"
	"database/sql"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
)

type Repository interface {
	Create(ctx context.Context, priority *model.PriorityModel) (int64, error)
	List(ctx context.Context) ([]model.PriorityModel, error)
	GetByCodeOrName(ctx context.Context, code string, name string) (*model.PriorityModel, error)
	GetByID(ctx context.Context, priorityID int64) (*model.PriorityModel, error)
	Update(ctx context.Context, priority *model.PriorityModel) error
	Delete(ctx context.Context, priorityID int64) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, priority *model.PriorityModel) (int64, error) {
	const query = `
		INSERT INTO priorities (code, name, weight, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRowContext(
		ctx,
		query,
		priority.Code,
		priority.Name,
		priority.Weight,
		priority.CreatedAt,
		priority.UpdatedAt,
	).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *repository) List(ctx context.Context) ([]model.PriorityModel, error) {
	const query = `
		SELECT id, code, name, weight, created_at, updated_at
		FROM priorities
		ORDER BY weight DESC, id ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.PriorityModel, 0)
	for rows.Next() {
		var item model.PriorityModel
		if err := rows.Scan(
			&item.ID,
			&item.Code,
			&item.Name,
			&item.Weight,
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

func (r *repository) GetByCodeOrName(ctx context.Context, code string, name string) (*model.PriorityModel, error) {
	const query = `
		SELECT id, code, name, weight, created_at, updated_at
		FROM priorities
		WHERE code = $1 OR name = $2
	`

	row := r.db.QueryRowContext(ctx, query, code, name)

	var item model.PriorityModel
	err := row.Scan(
		&item.ID,
		&item.Code,
		&item.Name,
		&item.Weight,
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

func (r *repository) GetByID(ctx context.Context, priorityID int64) (*model.PriorityModel, error) {
	const query = `
		SELECT id, code, name, weight, created_at, updated_at
		FROM priorities
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, priorityID)

	var item model.PriorityModel
	err := row.Scan(
		&item.ID,
		&item.Code,
		&item.Name,
		&item.Weight,
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

func (r *repository) Update(ctx context.Context, priority *model.PriorityModel) error {
	const query = `
		UPDATE priorities
		SET code = $2, name = $3, weight = $4, updated_at = $5
		WHERE id = $1
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		priority.ID,
		priority.Code,
		priority.Name,
		priority.Weight,
		priority.UpdatedAt,
	)

	return err
}

func (r *repository) Delete(ctx context.Context, priorityID int64) error {
	const query = `
		DELETE FROM priorities
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, priorityID)
	return err
}
