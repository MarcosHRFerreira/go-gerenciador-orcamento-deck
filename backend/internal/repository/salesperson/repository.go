package salesperson

import (
	"context"
	"database/sql"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
)

type Repository interface {
	Create(ctx context.Context, salesperson *model.SalespersonModel) (int64, error)
	List(ctx context.Context) ([]model.SalespersonModel, error)
	GetByEmail(ctx context.Context, email string) (*model.SalespersonModel, error)
	GetByUsername(ctx context.Context, username string) (*model.SalespersonModel, error)
	GetByID(ctx context.Context, salespersonID int64) (*model.SalespersonModel, error)
	Update(ctx context.Context, salesperson *model.SalespersonModel) error
	Delete(ctx context.Context, salespersonID int64) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, salesperson *model.SalespersonModel) (int64, error) {
	const query = `
		INSERT INTO salespeople (name, email, phone, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRowContext(
		ctx,
		query,
		salesperson.Name,
		salesperson.Email,
		salesperson.Phone,
		salesperson.Active,
		salesperson.CreatedAt,
		salesperson.UpdatedAt,
	).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *repository) List(ctx context.Context) ([]model.SalespersonModel, error) {
	const query = `
		SELECT id, name, email, phone, active, created_at, updated_at
		FROM salespeople
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.SalespersonModel, 0)
	for rows.Next() {
		var item model.SalespersonModel
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Email,
			&item.Phone,
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

func (r *repository) GetByEmail(ctx context.Context, email string) (*model.SalespersonModel, error) {
	const query = `
		SELECT id, name, email, phone, active, created_at, updated_at
		FROM salespeople
		WHERE email = $1
	`

	row := r.db.QueryRowContext(ctx, query, email)

	var item model.SalespersonModel
	err := row.Scan(
		&item.ID,
		&item.Name,
		&item.Email,
		&item.Phone,
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

func (r *repository) GetByUsername(ctx context.Context, username string) (*model.SalespersonModel, error) {
	const query = `
		SELECT id, name, email, phone, active, created_at, updated_at
		FROM salespeople
		WHERE LOWER(email) = LOWER($1)
		   OR LOWER(split_part(email, '@', 1)) = LOWER($1)
		   OR LOWER(name) = LOWER($1)
	`

	row := r.db.QueryRowContext(ctx, query, username)

	var item model.SalespersonModel
	err := row.Scan(
		&item.ID,
		&item.Name,
		&item.Email,
		&item.Phone,
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

func (r *repository) GetByID(ctx context.Context, salespersonID int64) (*model.SalespersonModel, error) {
	const query = `
		SELECT id, name, email, phone, active, created_at, updated_at
		FROM salespeople
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, salespersonID)

	var item model.SalespersonModel
	err := row.Scan(
		&item.ID,
		&item.Name,
		&item.Email,
		&item.Phone,
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

func (r *repository) Update(ctx context.Context, salesperson *model.SalespersonModel) error {
	const query = `
		UPDATE salespeople
		SET name = $2, email = $3, phone = $4, active = $5, updated_at = $6
		WHERE id = $1
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		salesperson.ID,
		salesperson.Name,
		salesperson.Email,
		salesperson.Phone,
		salesperson.Active,
		salesperson.UpdatedAt,
	)

	return err
}

func (r *repository) Delete(ctx context.Context, salespersonID int64) error {
	const query = `
		DELETE FROM salespeople
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, salespersonID)
	return err
}
