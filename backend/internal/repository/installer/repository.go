package installer

import (
	"context"
	"database/sql"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
)

type Repository interface {
	Create(ctx context.Context, installer *model.InstallerModel) (int64, error)
	List(ctx context.Context) ([]model.InstallerModel, error)
	GetByDocument(ctx context.Context, document string) (*model.InstallerModel, error)
	GetByID(ctx context.Context, installerID int64) (*model.InstallerModel, error)
	Update(ctx context.Context, installer *model.InstallerModel) error
	Delete(ctx context.Context, installerID int64) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, installer *model.InstallerModel) (int64, error) {
	const query = `
		INSERT INTO installers (name, document, email, phone, city, state, notes, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRowContext(
		ctx,
		query,
		installer.Name,
		nullIfEmpty(installer.Document),
		installer.Email,
		installer.Phone,
		installer.City,
		installer.State,
		installer.Notes,
		installer.Active,
		installer.CreatedAt,
		installer.UpdatedAt,
	).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *repository) List(ctx context.Context) ([]model.InstallerModel, error) {
	const query = `
		SELECT id, name, COALESCE(document, ''), email, phone, city, state, notes, active, created_at, updated_at
		FROM installers
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.InstallerModel, 0)
	for rows.Next() {
		var item model.InstallerModel
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Document,
			&item.Email,
			&item.Phone,
			&item.City,
			&item.State,
			&item.Notes,
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

func (r *repository) GetByDocument(ctx context.Context, document string) (*model.InstallerModel, error) {
	if document == "" {
		return nil, nil
	}

	const query = `
		SELECT id, name, COALESCE(document, ''), email, phone, city, state, notes, active, created_at, updated_at
		FROM installers
		WHERE document = $1
	`

	row := r.db.QueryRowContext(ctx, query, document)
	var item model.InstallerModel
	err := row.Scan(
		&item.ID,
		&item.Name,
		&item.Document,
		&item.Email,
		&item.Phone,
		&item.City,
		&item.State,
		&item.Notes,
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

func (r *repository) GetByID(ctx context.Context, installerID int64) (*model.InstallerModel, error) {
	const query = `
		SELECT id, name, COALESCE(document, ''), email, phone, city, state, notes, active, created_at, updated_at
		FROM installers
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, installerID)
	var item model.InstallerModel
	err := row.Scan(
		&item.ID,
		&item.Name,
		&item.Document,
		&item.Email,
		&item.Phone,
		&item.City,
		&item.State,
		&item.Notes,
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

func (r *repository) Update(ctx context.Context, installer *model.InstallerModel) error {
	const query = `
		UPDATE installers
		SET name = $2, document = $3, email = $4, phone = $5, city = $6, state = $7, notes = $8, active = $9, updated_at = $10
		WHERE id = $1
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		installer.ID,
		installer.Name,
		nullIfEmpty(installer.Document),
		installer.Email,
		installer.Phone,
		installer.City,
		installer.State,
		installer.Notes,
		installer.Active,
		installer.UpdatedAt,
	)

	return err
}

func (r *repository) Delete(ctx context.Context, installerID int64) error {
	const query = `
		DELETE FROM installers
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, installerID)
	return err
}

func nullIfEmpty(value string) interface{} {
	if value == "" {
		return nil
	}

	return value
}
