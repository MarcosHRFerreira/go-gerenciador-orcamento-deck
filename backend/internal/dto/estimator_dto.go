package dto

import "time"

type CreateEstimatorRequest struct {
	Code   string `json:"code" validate:"max=30"`
	Name   string `json:"name" validate:"required,min=3,max=150"`
	Email  string `json:"email" validate:"omitempty,email,max=250"`
	Phone  string `json:"phone" validate:"max=50"`
	Notes  string `json:"notes"`
	UserID *int64 `json:"user_id"`
}

type UpdateEstimatorRequest struct {
	Code   string `json:"code" validate:"required,max=30"`
	Name   string `json:"name" validate:"required,min=3,max=150"`
	Email  string `json:"email" validate:"omitempty,email,max=250"`
	Phone  string `json:"phone" validate:"max=50"`
	Active bool   `json:"active"`
	Notes  string `json:"notes"`
	UserID *int64 `json:"user_id"`
}

type EstimatorResponse struct {
	ID        int64     `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Active    bool      `json:"active"`
	Notes     string    `json:"notes"`
	UserID    *int64    `json:"user_id,omitempty"`
	UserName  *string   `json:"user_name,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
