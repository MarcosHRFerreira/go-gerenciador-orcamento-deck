package model

import "time"

type BudgetFollowUpModel struct {
	ID              int64
	BudgetID        int64
	CreatedByUserID int64
	Notes           string
	FollowUpAt      time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
