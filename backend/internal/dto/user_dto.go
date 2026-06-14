package dto

import "time"

type CreateUserRequest struct {
	Name            string `json:"name" validate:"required,min=3"`
	Email           string `json:"email" validate:"required,email"`
	Username        string `json:"username" validate:"required,min=3"`
	Password        string `json:"password" validate:"required,min=8,max=72"`
	PasswordConfirm string `json:"password_confirm" validate:"required,eqfield=Password"`
	Role            string `json:"role" validate:"required,oneof=admin user"`
}

type CreateUserResponse struct {
	ID int64 `json:"id"`
}

type UpdateUserRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=admin user"`
}

type UpdateUserActiveRequest struct {
	Active *bool `json:"active" validate:"required"`
}

type ResetUserPasswordRequest struct {
	Password        string `json:"password" validate:"required,min=8,max=72"`
	PasswordConfirm string `json:"password_confirm" validate:"required,eqfield=Password"`
}

type UserResponse struct {
	ID                 int64     `json:"id"`
	Name               string    `json:"name"`
	Email              string    `json:"email"`
	Username           string    `json:"username"`
	Role               string    `json:"role"`
	Active             bool      `json:"active"`
	MustChangePassword bool      `json:"must_change_password"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
