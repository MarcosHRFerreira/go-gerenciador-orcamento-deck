package dto

import "time"

type CreateBudgetStatusRequest struct {
	Code        string `json:"code" validate:"required,min=2,max=50"`
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description"`
	IsFinal     bool   `json:"is_final"`
	SortOrder   int    `json:"sort_order"`
}

type UpdateBudgetStatusRequest struct {
	Code        string `json:"code" validate:"required,min=2,max=50"`
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description"`
	IsFinal     bool   `json:"is_final"`
	SortOrder   int    `json:"sort_order"`
}

type BudgetStatusResponse struct {
	ID          int64     `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsFinal     bool      `json:"is_final"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateLossReasonRequest struct {
	Code        string `json:"code" validate:"required,min=2,max=50"`
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description"`
	Active      bool   `json:"active"`
}

type UpdateLossReasonRequest struct {
	Code        string `json:"code" validate:"required,min=2,max=50"`
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description"`
	Active      bool   `json:"active"`
}

type LossReasonResponse struct {
	ID          int64     `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreatePriorityRequest struct {
	Code   string `json:"code" validate:"required,min=2,max=50"`
	Name   string `json:"name" validate:"required,min=2,max=100"`
	Weight int    `json:"weight" validate:"required"`
}

type UpdatePriorityRequest struct {
	Code   string `json:"code" validate:"required,min=2,max=50"`
	Name   string `json:"name" validate:"required,min=2,max=100"`
	Weight int    `json:"weight" validate:"required"`
}

type PriorityResponse struct {
	ID        int64     `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Weight    int       `json:"weight"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateSalespersonRequest struct {
	Name  string `json:"name" validate:"required,min=3,max=150"`
	Email string `json:"email" validate:"required,email"`
	Phone string `json:"phone" validate:"required,min=8,max=30"`
}

type UpdateSalespersonRequest struct {
	Name   string `json:"name" validate:"required,min=3,max=150"`
	Email  string `json:"email" validate:"required,email"`
	Phone  string `json:"phone" validate:"required,min=8,max=30"`
	Active bool   `json:"active"`
}

type SalespersonResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateInstallerRequest struct {
	Name     string `json:"name" validate:"required,min=3,max=150"`
	Document string `json:"document" validate:"max=30"`
	Email    string `json:"email" validate:"required,email"`
	Phone    string `json:"phone" validate:"required,min=8,max=30"`
	City     string `json:"city" validate:"max=100"`
	State    string `json:"state" validate:"max=50"`
	Notes    string `json:"notes"`
}

type UpdateInstallerRequest struct {
	Name     string `json:"name" validate:"required,min=3,max=150"`
	Document string `json:"document" validate:"max=30"`
	Email    string `json:"email" validate:"required,email"`
	Phone    string `json:"phone" validate:"required,min=8,max=30"`
	City     string `json:"city" validate:"max=100"`
	State    string `json:"state" validate:"max=50"`
	Notes    string `json:"notes"`
	Active   bool   `json:"active"`
}

type InstallerResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Document  string    `json:"document"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	City      string    `json:"city"`
	State     string    `json:"state"`
	Notes     string    `json:"notes"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateContactRequest struct {
	InstallerID int64  `json:"installer_id" validate:"required"`
	Name        string `json:"name" validate:"required,min=3,max=150"`
	Email       string `json:"email" validate:"required,email"`
	Phone       string `json:"phone" validate:"required,min=8,max=30"`
	Role        string `json:"role" validate:"max=100"`
	IsPrimary   bool   `json:"is_primary"`
}

type UpdateContactRequest struct {
	InstallerID int64  `json:"installer_id" validate:"required"`
	Name        string `json:"name" validate:"required,min=3,max=150"`
	Email       string `json:"email" validate:"required,email"`
	Phone       string `json:"phone" validate:"required,min=8,max=30"`
	Role        string `json:"role" validate:"max=100"`
	IsPrimary   bool   `json:"is_primary"`
}

type ContactResponse struct {
	ID          int64     `json:"id"`
	InstallerID int64     `json:"installer_id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
	Role        string    `json:"role"`
	IsPrimary   bool      `json:"is_primary"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateProjectTypeRequest struct {
	Code        string `json:"code" validate:"required,min=2,max=50"`
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description"`
}

type UpdateProjectTypeRequest struct {
	Code        string `json:"code" validate:"required,min=2,max=50"`
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description"`
}

type ProjectTypeResponse struct {
	ID          int64     `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateProjectRequest struct {
	Name          string `json:"name" validate:"required,min=3,max=200"`
	ProjectTypeID *int64 `json:"project_type_id"`
	City          string `json:"city" validate:"max=100"`
	State         string `json:"state" validate:"max=50"`
	Notes         string `json:"notes"`
}

type UpdateProjectRequest struct {
	Name          string `json:"name" validate:"required,min=3,max=200"`
	ProjectTypeID *int64 `json:"project_type_id"`
	City          string `json:"city" validate:"max=100"`
	State         string `json:"state" validate:"max=50"`
	Notes         string `json:"notes"`
}

type ProjectResponse struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	ProjectTypeID *int64    `json:"project_type_id,omitempty"`
	City          string    `json:"city"`
	State         string    `json:"state"`
	Notes         string    `json:"notes"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
