package lossreason

import (
	"context"
	"database/sql"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
)

type Repository interface {
	Create(ctx context.Context, reason *model.LossReasonModel) (int64, error)
	List(ctx context.Context) ([]model.LossReasonModel, error)
	GetByCodeOrName(ctx context.Context, code string, name string) (*model.LossReasonModel, error)
	GetByID(ctx context.Context, reasonID int64) (*model.LossReasonModel, error)
	Update(ctx context.Context, reason *model.LossReasonModel) error
	Delete(ctx context.Context, reasonID int64) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, reason *model.LossReasonModel) (int64, error) {
	const query = `
		INSERT INTO loss_reasons (code, name, description, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRowContext(
		ctx,
		query,
		reason.Code,
		reason.Name,
		reason.Description,
		reason.Active,
		reason.CreatedAt,
		reason.UpdatedAt,
	).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *repository) List(ctx context.Context) ([]model.LossReasonModel, error) {
	const query = `
		SELECT id, code, name, description, active, created_at, updated_at
		FROM loss_reasons
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.LossReasonModel, 0)
	for rows.Next() {
		var item model.LossReasonModel
		if err := rows.Scan(
			&item.ID,
			&item.Code,
			&item.Name,
			&item.Description,
			&item.Active,
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

func (r *repository) GetByCodeOrName(ctx context.Context, code string, name string) (*model.LossReasonModel, error) {
	const query = `
		SELECT id, code, name, description, active, created_at, updated_at
		FROM loss_reasons
		WHERE code = $1 OR name = $2
	`

	row := r.db.QueryRowContext(ctx, query, code, name)

	var item model.LossReasonModel
	err := row.Scan(
		&item.ID,
		&item.Code,
		&item.Name,
		&item.Description,
		&item.Active,
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

func (r *repository) GetByID(ctx context.Context, reasonID int64) (*model.LossReasonModel, error) {
	const query = `
		SELECT id, code, name, description, active, created_at, updated_at
		FROM loss_reasons
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, reasonID)

	var item model.LossReasonModel
	err := row.Scan(
		&item.ID,
		&item.Code,
		&item.Name,
		&item.Description,
		&item.Active,
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

func (r *repository) Update(ctx context.Context, reason *model.LossReasonModel) error {
	const query = `
		UPDATE loss_reasons
		SET code = $2, name = $3, description = $4, active = $5, updated_at = $6
		WHERE id = $1
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		reason.ID,
		reason.Code,
		reason.Name,
		reason.Description,
		reason.Active,
		reason.UpdatedAt,
	)

	return err
}

func (r *repository) Delete(ctx context.Context, reasonID int64) error {
	const query = `
		DELETE FROM loss_reasons
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, reasonID)
	return err
}
