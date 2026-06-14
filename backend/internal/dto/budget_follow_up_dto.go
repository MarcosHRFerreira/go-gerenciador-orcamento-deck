package dto

import "time"

type CreateBudgetFollowUpRequest struct {
	Notes      string     `json:"notes" validate:"required,min=3"`
	FollowUpAt *time.Time `json:"follow_up_at"`
}

type BudgetFollowUpResponse struct {
	ID              int64     `json:"id"`
	BudgetID        int64     `json:"budget_id"`
	CreatedByUserID int64     `json:"created_by_user_id"`
	Notes           string    `json:"notes"`
	FollowUpAt      time.Time `json:"follow_up_at"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
