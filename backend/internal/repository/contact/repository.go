package contact

import (
	"context"
	"database/sql"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
)

type Repository interface {
	Create(ctx context.Context, contact *model.ContactModel) (int64, error)
	List(ctx context.Context, installerID *int64) ([]model.ContactModel, error)
	GetByID(ctx context.Context, contactID int64) (*model.ContactModel, error)
	GetByInstallerAndEmail(ctx context.Context, installerID int64, email string) (*model.ContactModel, error)
	GetByInstallerAndPhone(ctx context.Context, installerID int64, phone string) (*model.ContactModel, error)
	Update(ctx context.Context, contact *model.ContactModel) error
	Delete(ctx context.Context, contactID int64) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, contact *model.ContactModel) (int64, error) {
	const query = `
		INSERT INTO contacts (installer_id, name, email, phone, role, is_primary, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRowContext(
		ctx,
		query,
		contact.InstallerID,
		contact.Name,
		contact.Email,
		contact.Phone,
		contact.Role,
		contact.IsPrimary,
		contact.CreatedAt,
		contact.UpdatedAt,
	).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *repository) List(ctx context.Context, installerID *int64) ([]model.ContactModel, error) {
	query := `
		SELECT id, installer_id, name, email, phone, role, is_primary, created_at, updated_at
		FROM contacts
	`

	args := make([]interface{}, 0, 1)
	if installerID != nil {
		query += ` WHERE installer_id = $1`
		args = append(args, *installerID)
	}

	query += ` ORDER BY name ASC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.ContactModel, 0)
	for rows.Next() {
		var item model.ContactModel
		if err := rows.Scan(
			&item.ID,
			&item.InstallerID,
			&item.Name,
			&item.Email,
			&item.Phone,
			&item.Role,
			&item.IsPrimary,
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

func (r *repository) GetByID(ctx context.Context, contactID int64) (*model.ContactModel, error) {
	const query = `
		SELECT id, installer_id, name, email, phone, role, is_primary, created_at, updated_at
		FROM contacts
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, contactID)
	var item model.ContactModel
	err := row.Scan(
		&item.ID,
		&item.InstallerID,
		&item.Name,
		&item.Email,
		&item.Phone,
		&item.Role,
		&item.IsPrimary,
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

func (r *repository) GetByInstallerAndEmail(ctx context.Context, installerID int64, email string) (*model.ContactModel, error) {
	const query = `
		SELECT id, installer_id, name, email, phone, role, is_primary, created_at, updated_at
		FROM contacts
		WHERE installer_id = $1 AND email = $2
	`

	row := r.db.QueryRowContext(ctx, query, installerID, email)
	var item model.ContactModel
	err := row.Scan(
		&item.ID,
		&item.InstallerID,
		&item.Name,
		&item.Email,
		&item.Phone,
		&item.Role,
		&item.IsPrimary,
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

func (r *repository) GetByInstallerAndPhone(ctx context.Context, installerID int64, phone string) (*model.ContactModel, error) {
	const query = `
		SELECT id, installer_id, name, email, phone, role, is_primary, created_at, updated_at
		FROM contacts
		WHERE installer_id = $1 AND phone = $2
	`

	row := r.db.QueryRowContext(ctx, query, installerID, phone)
	var item model.ContactModel
	err := row.Scan(
		&item.ID,
		&item.InstallerID,
		&item.Name,
		&item.Email,
		&item.Phone,
		&item.Role,
		&item.IsPrimary,
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

func (r *repository) Update(ctx context.Context, contact *model.ContactModel) error {
	const query = `
		UPDATE contacts
		SET installer_id = $2, name = $3, email = $4, phone = $5, role = $6, is_primary = $7, updated_at = $8
		WHERE id = $1
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		contact.ID,
		contact.InstallerID,
		contact.Name,
		contact.Email,
		contact.Phone,
		contact.Role,
		contact.IsPrimary,
		contact.UpdatedAt,
	)

	return err
}

func (r *repository) Delete(ctx context.Context, contactID int64) error {
	const query = `
		DELETE FROM contacts
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, contactID)
	return err
}
