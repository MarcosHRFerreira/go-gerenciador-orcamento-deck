package model

import "time"

type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
)

type UserModel struct {
	ID                 int64
	Name               string
	Email              string
	Username           string
	PasswordHash       string
	Role               UserRole
	Active             bool
	MustChangePassword bool
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type RefreshTokenModel struct {
	ID           int64
	UserID       int64
	RefreshToken string
	ExpiredAt    time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
