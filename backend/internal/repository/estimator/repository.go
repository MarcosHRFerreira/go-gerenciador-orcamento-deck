package estimator

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
)

type Repository interface {
	Create(ctx context.Context, item *model.EstimatorModel) (int64, error)
	GetNextCode(ctx context.Context) (string, error)
	List(ctx context.Context) ([]model.EstimatorModel, error)
	GetByID(ctx context.Context, estimatorID int64) (*model.EstimatorModel, error)
	GetByCode(ctx context.Context, code string) (*model.EstimatorModel, error)
	GetByUserID(ctx context.Context, userID int64) (*model.EstimatorModel, error)
	Update(ctx context.Context, item *model.EstimatorModel) error
	Delete(ctx context.Context, estimatorID int64) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, item *model.EstimatorModel) (int64, error) {
	const query = `
		INSERT INTO estimators (code, name, email, phone, active, notes, user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRowContext(
		ctx,
		query,
		item.Code,
		item.Name,
		item.Email,
		item.Phone,
		item.Active,
		item.Notes,
		nullableInt64(item.UserID),
		item.CreatedAt,
		item.UpdatedAt,
	).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *repository) GetNextCode(ctx context.Context) (string, error) {
	const query = `
		SELECT COALESCE(MAX(CAST(SUBSTRING(code FROM '([0-9]+)$') AS INTEGER)), 0)
		FROM estimators
		WHERE code ~ '^(EST-[0-9]{6}|ESTIMATOR_[0-9]+)$'
	`

	var lastCodeNumber int
	if err := r.db.QueryRowContext(ctx, query).Scan(&lastCodeNumber); err != nil {
		return "", err
	}

	return fmt.Sprintf("EST-%06d", lastCodeNumber+1), nil
}

func (r *repository) List(ctx context.Context) ([]model.EstimatorModel, error) {
	const query = `
		SELECT
			e.id,
			e.code,
			e.name,
			e.email,
			e.phone,
			e.active,
			e.notes,
			e.user_id,
			e.created_at,
			e.updated_at,
			u.name
		FROM estimators e
		LEFT JOIN users u ON u.id = e.user_id
		ORDER BY e.name ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.EstimatorModel, 0)
	for rows.Next() {
		item, scanErr := scanEstimator(rows)
		if scanErr != nil {
			return nil, scanErr
		}

		items = append(items, *item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *repository) GetByID(ctx context.Context, estimatorID int64) (*model.EstimatorModel, error) {
	const query = `
		SELECT
			e.id,
			e.code,
			e.name,
			e.email,
			e.phone,
			e.active,
			e.notes,
			e.user_id,
			e.created_at,
			e.updated_at,
			u.name
		FROM estimators e
		LEFT JOIN users u ON u.id = e.user_id
		WHERE e.id = $1
	`

	return r.getOne(ctx, query, estimatorID)
}

func (r *repository) GetByCode(ctx context.Context, code string) (*model.EstimatorModel, error) {
	const query = `
		SELECT
			e.id,
			e.code,
			e.name,
			e.email,
			e.phone,
			e.active,
			e.notes,
			e.user_id,
			e.created_at,
			e.updated_at,
			u.name
		FROM estimators e
		LEFT JOIN users u ON u.id = e.user_id
		WHERE e.code = $1
	`

	return r.getOne(ctx, query, code)
}

func (r *repository) GetByUserID(ctx context.Context, userID int64) (*model.EstimatorModel, error) {
	const query = `
		SELECT
			e.id,
			e.code,
			e.name,
			e.email,
			e.phone,
			e.active,
			e.notes,
			e.user_id,
			e.created_at,
			e.updated_at,
			u.name
		FROM estimators e
		LEFT JOIN users u ON u.id = e.user_id
		WHERE e.user_id = $1
	`

	return r.getOne(ctx, query, userID)
}

func (r *repository) Update(ctx context.Context, item *model.EstimatorModel) error {
	const query = `
		UPDATE estimators
		SET code = $2, name = $3, email = $4, phone = $5, active = $6, notes = $7, user_id = $8, updated_at = $9
		WHERE id = $1
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		item.ID,
		item.Code,
		item.Name,
		item.Email,
		item.Phone,
		item.Active,
		item.Notes,
		nullableInt64(item.UserID),
		item.UpdatedAt,
	)

	return err
}

func (r *repository) Delete(ctx context.Context, estimatorID int64) error {
	const query = `
		DELETE FROM estimators
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, estimatorID)
	return err
}

func (r *repository) getOne(ctx context.Context, query string, args ...interface{}) (*model.EstimatorModel, error) {
	row := r.db.QueryRowContext(ctx, query, args...)

	item, err := scanEstimator(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return item, nil
}

type estimatorScanner interface {
	Scan(dest ...interface{}) error
}

func scanEstimator(scanner estimatorScanner) (*model.EstimatorModel, error) {
	var item model.EstimatorModel
	err := scanner.Scan(
		&item.ID,
		&item.Code,
		&item.Name,
		&item.Email,
		&item.Phone,
		&item.Active,
		&item.Notes,
		&item.UserID,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.UserName,
	)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func nullableInt64(value sql.NullInt64) interface{} {
	if !value.Valid {
		return nil
	}

	return value.Int64
}
