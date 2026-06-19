package user

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
)

type Repository interface {
	CountUsers(ctx context.Context) (int64, error)
	CountActiveAdmins(ctx context.Context) (int64, error)
	CreateUser(ctx context.Context, user *model.UserModel) (int64, error)
	GetUserByEmail(ctx context.Context, email string) (*model.UserModel, error)
	GetUserByUsername(ctx context.Context, username string) (*model.UserModel, error)
	GetUserByEmailOrUsername(ctx context.Context, email string, username string) (*model.UserModel, error)
	GetUserByID(ctx context.Context, userID int64) (*model.UserModel, error)
	GetUsersByIDs(ctx context.Context, userIDs []int64) ([]model.UserModel, error)
	ListUsers(ctx context.Context) ([]model.UserModel, error)
	ListActiveUsers(ctx context.Context) ([]model.UserModel, error)
	UpdateUser(ctx context.Context, userID int64, name string, email string, username string, role model.UserRole, userKind model.UserKind, updatedAt time.Time) error
	UpdateUserRole(ctx context.Context, userID int64, role model.UserRole, userKind model.UserKind, updatedAt time.Time) error
	UpdateUserActive(ctx context.Context, userID int64, active bool, updatedAt time.Time) error
	UpdateUserPassword(ctx context.Context, userID int64, passwordHash string, mustChangePassword bool, updatedAt time.Time) error
	GetActiveRefreshTokenByHash(ctx context.Context, refreshTokenHash string, now time.Time) (*model.RefreshTokenModel, error)
	StoreRefreshToken(ctx context.Context, token *model.RefreshTokenModel) error
	DeleteRefreshTokensByUserID(ctx context.Context, userID int64) error
	DeleteRefreshTokenByHash(ctx context.Context, refreshTokenHash string) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) CountUsers(ctx context.Context) (int64, error) {
	const query = `SELECT COUNT(*) FROM users`

	var total int64
	if err := r.db.QueryRowContext(ctx, query).Scan(&total); err != nil {
		return 0, err
	}

	return total, nil
}

func (r *repository) CountActiveAdmins(ctx context.Context) (int64, error) {
	const query = `SELECT COUNT(*) FROM users WHERE role = 'admin' AND active = TRUE`

	var total int64
	if err := r.db.QueryRowContext(ctx, query).Scan(&total); err != nil {
		return 0, err
	}

	return total, nil
}

func (r *repository) CreateUser(ctx context.Context, user *model.UserModel) (int64, error) {
	const query = `
		INSERT INTO users (
			name,
			email,
			username,
			password_hash,
			role,
			user_kind,
			active,
			must_change_password,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`

	var userID int64
	err := r.db.QueryRowContext(
		ctx,
		query,
		user.Name,
		user.Email,
		user.Username,
		user.PasswordHash,
		user.Role,
		nullableUserKind(user.UserKind),
		user.Active,
		user.MustChangePassword,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&userID)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func (r *repository) GetUserByEmail(ctx context.Context, email string) (*model.UserModel, error) {
	const query = `
		SELECT id, name, email, username, password_hash, role, user_kind, active, must_change_password, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	return r.getOne(ctx, query, email)
}

func (r *repository) GetUserByUsername(ctx context.Context, username string) (*model.UserModel, error) {
	const query = `
		SELECT id, name, email, username, password_hash, role, user_kind, active, must_change_password, created_at, updated_at
		FROM users
		WHERE username = $1
	`

	return r.getOne(ctx, query, username)
}

func (r *repository) GetUserByEmailOrUsername(ctx context.Context, email string, username string) (*model.UserModel, error) {
	const query = `
		SELECT id, name, email, username, password_hash, role, user_kind, active, must_change_password, created_at, updated_at
		FROM users
		WHERE email = $1 OR username = $2
	`

	return r.getOne(ctx, query, email, username)
}

func (r *repository) GetUserByID(ctx context.Context, userID int64) (*model.UserModel, error) {
	const query = `
		SELECT id, name, email, username, password_hash, role, user_kind, active, must_change_password, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	return r.getOne(ctx, query, userID)
}

func (r *repository) GetUsersByIDs(ctx context.Context, userIDs []int64) ([]model.UserModel, error) {
	if len(userIDs) == 0 {
		return []model.UserModel{}, nil
	}

	query := `
		SELECT id, name, email, username, password_hash, role, user_kind, active, must_change_password, created_at, updated_at
		FROM users
		WHERE id IN (`

	args := make([]interface{}, 0, len(userIDs))
	placeholders := make([]string, 0, len(userIDs))
	for index, userID := range userIDs {
		args = append(args, userID)
		placeholders = append(placeholders, fmt.Sprintf("$%d", index+1))
	}

	query += strings.Join(placeholders, ", ")
	query += `) ORDER BY id ASC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanUserRows(rows)
}

func (r *repository) ListUsers(ctx context.Context) ([]model.UserModel, error) {
	const query = `
		SELECT id, name, email, username, password_hash, role, user_kind, active, must_change_password, created_at, updated_at
		FROM users
		ORDER BY id ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanUserRows(rows)
}

func (r *repository) ListActiveUsers(ctx context.Context) ([]model.UserModel, error) {
	const query = `
		SELECT id, name, email, username, password_hash, role, user_kind, active, must_change_password, created_at, updated_at
		FROM users
		WHERE active = TRUE
		ORDER BY id ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanUserRows(rows)
}

func (r *repository) UpdateUser(ctx context.Context, userID int64, name string, email string, username string, role model.UserRole, userKind model.UserKind, updatedAt time.Time) error {
	const query = `
		UPDATE users
		SET name = $2, email = $3, username = $4, role = $5, user_kind = $6, updated_at = $7
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, userID, name, email, username, role, nullableUserKind(userKind), updatedAt)
	return err
}

func (r *repository) UpdateUserRole(ctx context.Context, userID int64, role model.UserRole, userKind model.UserKind, updatedAt time.Time) error {
	const query = `
		UPDATE users
		SET role = $2, user_kind = $3, updated_at = $4
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, userID, role, nullableUserKind(userKind), updatedAt)
	return err
}

func (r *repository) UpdateUserActive(ctx context.Context, userID int64, active bool, updatedAt time.Time) error {
	const query = `
		UPDATE users
		SET active = $2, updated_at = $3
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, userID, active, updatedAt)
	return err
}

func (r *repository) UpdateUserPassword(ctx context.Context, userID int64, passwordHash string, mustChangePassword bool, updatedAt time.Time) error {
	const query = `
		UPDATE users
		SET password_hash = $2, must_change_password = $3, updated_at = $4
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, userID, passwordHash, mustChangePassword, updatedAt)
	return err
}

func (r *repository) GetActiveRefreshTokenByHash(ctx context.Context, refreshTokenHash string, now time.Time) (*model.RefreshTokenModel, error) {
	const query = `
		SELECT id, user_id, refresh_token, expired_at, created_at, updated_at
		FROM refresh_tokens
		WHERE refresh_token = $1 AND expired_at > $2
	`

	row := r.db.QueryRowContext(ctx, query, refreshTokenHash, now)

	var token model.RefreshTokenModel
	err := row.Scan(
		&token.ID,
		&token.UserID,
		&token.RefreshToken,
		&token.ExpiredAt,
		&token.CreatedAt,
		&token.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &token, nil
}

func (r *repository) StoreRefreshToken(ctx context.Context, token *model.RefreshTokenModel) error {
	const query = `
		INSERT INTO refresh_tokens (
			user_id,
			refresh_token,
			expired_at,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		token.UserID,
		token.RefreshToken,
		token.ExpiredAt,
		token.CreatedAt,
		token.UpdatedAt,
	)

	return err
}

func (r *repository) DeleteRefreshTokensByUserID(ctx context.Context, userID int64) error {
	const query = `DELETE FROM refresh_tokens WHERE user_id = $1`

	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

func (r *repository) DeleteRefreshTokenByHash(ctx context.Context, refreshTokenHash string) error {
	const query = `DELETE FROM refresh_tokens WHERE refresh_token = $1`

	_, err := r.db.ExecContext(ctx, query, refreshTokenHash)
	return err
}

func (r *repository) getOne(ctx context.Context, query string, args ...interface{}) (*model.UserModel, error) {
	row := r.db.QueryRowContext(ctx, query, args...)

	var user model.UserModel
	var userKind sql.NullString
	err := row.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&userKind,
		&user.Active,
		&user.MustChangePassword,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	user.UserKind = normalizeUserKind(userKind)

	return &user, nil
}

func scanUserRows(rows *sql.Rows) ([]model.UserModel, error) {
	users := make([]model.UserModel, 0)
	for rows.Next() {
		var user model.UserModel
		var userKind sql.NullString
		if err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.Username,
			&user.PasswordHash,
			&user.Role,
			&userKind,
			&user.Active,
			&user.MustChangePassword,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, err
		}

		user.UserKind = normalizeUserKind(userKind)
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func nullableUserKind(userKind model.UserKind) interface{} {
	if userKind == "" {
		return nil
	}

	return string(userKind)
}

func normalizeUserKind(userKind sql.NullString) model.UserKind {
	if !userKind.Valid {
		return ""
	}

	return model.UserKind(userKind.String)
}
