package systemtype

import (
	"context"
	"database/sql"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
)

type Repository interface {
	Create(ctx context.Context, item *model.SystemTypeModel) (int64, error)
	List(ctx context.Context) ([]model.SystemTypeModel, error)
	GetByCodeOrName(ctx context.Context, code string, name string) (*model.SystemTypeModel, error)
	GetByID(ctx context.Context, systemTypeID int64) (*model.SystemTypeModel, error)
	Update(ctx context.Context, item *model.SystemTypeModel) error
	Delete(ctx context.Context, systemTypeID int64) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, item *model.SystemTypeModel) (int64, error) {
	const query = `
		INSERT INTO system_types (code, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRowContext(
		ctx,
		query,
		item.Code,
		item.Name,
		item.Description,
		item.CreatedAt,
		item.UpdatedAt,
	).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *repository) List(ctx context.Context) ([]model.SystemTypeModel, error) {
	const query = `
		SELECT id, code, name, description, created_at, updated_at
		FROM system_types
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.SystemTypeModel, 0)
	for rows.Next() {
		var item model.SystemTypeModel
		if err := rows.Scan(
			&item.ID,
			&item.Code,
			&item.Name,
			&item.Description,
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

func (r *repository) GetByCodeOrName(ctx context.Context, code string, name string) (*model.SystemTypeModel, error) {
	const query = `
		SELECT id, code, name, description, created_at, updated_at
		FROM system_types
		WHERE code = $1 OR name = $2
	`

	var item model.SystemTypeModel
	err := r.db.QueryRowContext(ctx, query, code, name).Scan(
		&item.ID,
		&item.Code,
		&item.Name,
		&item.Description,
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

func (r *repository) GetByID(ctx context.Context, systemTypeID int64) (*model.SystemTypeModel, error) {
	const query = `
		SELECT id, code, name, description, created_at, updated_at
		FROM system_types
		WHERE id = $1
	`

	var item model.SystemTypeModel
	err := r.db.QueryRowContext(ctx, query, systemTypeID).Scan(
		&item.ID,
		&item.Code,
		&item.Name,
		&item.Description,
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

func (r *repository) Update(ctx context.Context, item *model.SystemTypeModel) error {
	const query = `
		UPDATE system_types
		SET code = $2, name = $3, description = $4, updated_at = $5
		WHERE id = $1
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		item.ID,
		item.Code,
		item.Name,
		item.Description,
		item.UpdatedAt,
	)
	return err
}

func (r *repository) Delete(ctx context.Context, systemTypeID int64) error {
	const query = `
		DELETE FROM system_types
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, systemTypeID)
	return err
}
