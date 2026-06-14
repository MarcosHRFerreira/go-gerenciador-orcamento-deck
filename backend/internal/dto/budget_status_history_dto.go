package dto

import "time"

type ChangeBudgetStatusRequest struct {
	StatusID int64  `json:"status_id" validate:"required"`
	Notes    string `json:"notes" validate:"max=1000"`
}

type BudgetStatusHistoryResponse struct {
	ID              int64     `json:"id"`
	BudgetID        int64     `json:"budget_id"`
	FromStatusID    *int64    `json:"from_status_id,omitempty"`
	ToStatusID      int64     `json:"to_status_id"`
	ChangedByUserID int64     `json:"changed_by_user_id"`
	Notes           string    `json:"notes"`
	ChangedAt       time.Time `json:"changed_at"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
