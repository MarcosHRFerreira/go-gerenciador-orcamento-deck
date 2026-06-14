package project

import (
	"context"
	"database/sql"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
)

type Repository interface {
	Create(ctx context.Context, item *model.ProjectModel) (int64, error)
	List(ctx context.Context) ([]model.ProjectModel, error)
	GetByID(ctx context.Context, projectID int64) (*model.ProjectModel, error)
	Update(ctx context.Context, item *model.ProjectModel) error
	Delete(ctx context.Context, projectID int64) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, item *model.ProjectModel) (int64, error) {
	const query = `
		INSERT INTO projects (name, project_type_id, city, state, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRowContext(
		ctx,
		query,
		item.Name,
		nullableProjectTypeID(item.ProjectTypeID),
		item.City,
		item.State,
		item.Notes,
		item.CreatedAt,
		item.UpdatedAt,
	).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *repository) List(ctx context.Context) ([]model.ProjectModel, error) {
	const query = `
		SELECT id, name, project_type_id, city, state, notes, created_at, updated_at
		FROM projects
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.ProjectModel, 0)
	for rows.Next() {
		var item model.ProjectModel
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.ProjectTypeID,
			&item.City,
			&item.State,
			&item.Notes,
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

func (r *repository) GetByID(ctx context.Context, projectID int64) (*model.ProjectModel, error) {
	const query = `
		SELECT id, name, project_type_id, city, state, notes, created_at, updated_at
		FROM projects
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, projectID)

	var item model.ProjectModel
	err := row.Scan(
		&item.ID,
		&item.Name,
		&item.ProjectTypeID,
		&item.City,
		&item.State,
		&item.Notes,
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

func (r *repository) Update(ctx context.Context, item *model.ProjectModel) error {
	const query = `
		UPDATE projects
		SET name = $2, project_type_id = $3, city = $4, state = $5, notes = $6, updated_at = $7
		WHERE id = $1
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		item.ID,
		item.Name,
		nullableProjectTypeID(item.ProjectTypeID),
		item.City,
		item.State,
		item.Notes,
		item.UpdatedAt,
	)

	return err
}

func (r *repository) Delete(ctx context.Context, projectID int64) error {
	const query = `
		DELETE FROM projects
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, projectID)
	return err
}

func nullableProjectTypeID(value sql.NullInt64) interface{} {
	if !value.Valid {
		return nil
	}

	return value.Int64
}
