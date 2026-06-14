package dto

import "time"

type CreateBudgetRequest struct {
	BudgetNumber         string    `json:"budget_number" validate:"required,min=1,max=50"`
	YearBudget           int       `json:"year_budget" validate:"required"`
	Revision             int       `json:"revision"`
	SentAt               time.Time `json:"sent_at"`
	GrossValue           float64   `json:"gross_value"`
	CommissionValue      float64   `json:"commission_value"`
	AreaM2               float64   `json:"area_m2"`
	StatusID             int64     `json:"status_id" validate:"required"`
	PriorityID           *int64    `json:"priority_id"`
	InstallerID          *int64    `json:"installer_id"`
	ProjectID            *int64    `json:"project_id"`
	SalespersonID        *int64    `json:"salesperson_id"`
	ContactID            *int64    `json:"contact_id"`
	LossReasonID         *int64    `json:"loss_reason_id"`
	CompetitorName       string    `json:"competitor_name" validate:"max=150"`
	CompetitorPrice      *float64  `json:"competitor_price"`
	DesignerName         string    `json:"designer_name" validate:"max=150"`
	SpecificationDetails string    `json:"specification_details"`
	CurrentFollowUp      string    `json:"current_follow_up"`
}

type UpdateBudgetRequest struct {
	BudgetNumber         string    `json:"budget_number" validate:"required,min=1,max=50"`
	YearBudget           int       `json:"year_budget" validate:"required"`
	Revision             int       `json:"revision"`
	SentAt               time.Time `json:"sent_at"`
	GrossValue           float64   `json:"gross_value"`
	CommissionValue      float64   `json:"commission_value"`
	AreaM2               float64   `json:"area_m2"`
	StatusID             int64     `json:"status_id" validate:"required"`
	PriorityID           *int64    `json:"priority_id"`
	InstallerID          *int64    `json:"installer_id"`
	ProjectID            *int64    `json:"project_id"`
	SalespersonID        *int64    `json:"salesperson_id"`
	ContactID            *int64    `json:"contact_id"`
	LossReasonID         *int64    `json:"loss_reason_id"`
	CompetitorName       string    `json:"competitor_name" validate:"max=150"`
	CompetitorPrice      *float64  `json:"competitor_price"`
	DesignerName         string    `json:"designer_name" validate:"max=150"`
	SpecificationDetails string    `json:"specification_details"`
	CurrentFollowUp      string    `json:"current_follow_up"`
}

type ListBudgetsFilters struct {
	BudgetNumber            string
	YearBudget              *int
	StatusID                *int64
	SalespersonID           *int64
	RestrictedSalespersonID *int64
	InstallerID             *int64
	PriorityID              *int64
	ProjectTypeID           *int64
	DesignerName            string
	CompetitorName          string
	SentAtFrom              *time.Time
	SentAtTo                *time.Time
	GrossValueMin           *float64
	GrossValueMax           *float64
	Page                    int
	PageSize                int
	SortBy                  string
	SortOrder               string
}

type BudgetResponse struct {
	ID                   int64     `json:"id"`
	BudgetNumber         string    `json:"budget_number"`
	YearBudget           int       `json:"year_budget"`
	Revision             int       `json:"revision"`
	SentAt               time.Time `json:"sent_at"`
	GrossValue           float64   `json:"gross_value"`
	CommissionValue      float64   `json:"commission_value"`
	AreaM2               float64   `json:"area_m2"`
	StatusID             int64     `json:"status_id"`
	PriorityID           *int64    `json:"priority_id,omitempty"`
	InstallerID          *int64    `json:"installer_id,omitempty"`
	ProjectID            *int64    `json:"project_id,omitempty"`
	SalespersonID        *int64    `json:"salesperson_id,omitempty"`
	ContactID            *int64    `json:"contact_id,omitempty"`
	LossReasonID         *int64    `json:"loss_reason_id,omitempty"`
	CompetitorName       string    `json:"competitor_name"`
	CompetitorPrice      *float64  `json:"competitor_price,omitempty"`
	DesignerName         string    `json:"designer_name"`
	ProjectName          *string   `json:"project_name,omitempty"`
	SalespersonName      *string   `json:"salesperson_name,omitempty"`
	ContactName          *string   `json:"contact_name,omitempty"`
	SpecificationDetails string    `json:"specification_details"`
	CurrentFollowUp      string    `json:"current_follow_up"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type ListBudgetsResponse struct {
	Items    []BudgetResponse `json:"items"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
	Total    int64            `json:"total"`
}
