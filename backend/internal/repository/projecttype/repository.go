package projecttype

import (
	"context"
	"database/sql"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
)

type Repository interface {
	Create(ctx context.Context, item *model.ProjectTypeModel) (int64, error)
	List(ctx context.Context) ([]model.ProjectTypeModel, error)
	GetByCodeOrName(ctx context.Context, code string, name string) (*model.ProjectTypeModel, error)
	GetByID(ctx context.Context, projectTypeID int64) (*model.ProjectTypeModel, error)
	Update(ctx context.Context, item *model.ProjectTypeModel) error
	Delete(ctx context.Context, projectTypeID int64) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, item *model.ProjectTypeModel) (int64, error) {
	const query = `
		INSERT INTO project_types (code, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRowContext(ctx, query, item.Code, item.Name, item.Description, item.CreatedAt, item.UpdatedAt).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *repository) List(ctx context.Context) ([]model.ProjectTypeModel, error) {
	const query = `
		SELECT id, code, name, description, created_at, updated_at
		FROM project_types
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.ProjectTypeModel, 0)
	for rows.Next() {
		var item model.ProjectTypeModel
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

func (r *repository) GetByCodeOrName(ctx context.Context, code string, name string) (*model.ProjectTypeModel, error) {
	const query = `
		SELECT id, code, name, description, created_at, updated_at
		FROM project_types
		WHERE code = $1 OR name = $2
	`

	row := r.db.QueryRowContext(ctx, query, code, name)
	var item model.ProjectTypeModel
	err := row.Scan(&item.ID, &item.Code, &item.Name, &item.Description, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &item, nil
}

func (r *repository) GetByID(ctx context.Context, projectTypeID int64) (*model.ProjectTypeModel, error) {
	const query = `
		SELECT id, code, name, description, created_at, updated_at
		FROM project_types
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, projectTypeID)
	var item model.ProjectTypeModel
	err := row.Scan(&item.ID, &item.Code, &item.Name, &item.Description, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &item, nil
}

func (r *repository) Update(ctx context.Context, item *model.ProjectTypeModel) error {
	const query = `
		UPDATE project_types
		SET code = $2, name = $3, description = $4, updated_at = $5
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, item.ID, item.Code, item.Name, item.Description, item.UpdatedAt)
	return err
}

func (r *repository) Delete(ctx context.Context, projectTypeID int64) error {
	const query = `
		DELETE FROM project_types
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, projectTypeID)
	return err
}
