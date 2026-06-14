package model

import (
	"database/sql"
	"time"
)

type BudgetStatusHistoryModel struct {
	ID              int64
	BudgetID        int64
	FromStatusID    sql.NullInt64
	ToStatusID      int64
	ChangedByUserID int64
	Notes           string
	ChangedAt       time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
