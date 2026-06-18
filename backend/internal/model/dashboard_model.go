package model

import (
	"database/sql"
	"time"
)

type DashboardBudgetSnapshotModel struct {
	ID                   int64
	BudgetNumber         string
	SentAt               time.Time
	GrossValue           float64
	ConstructionCompany  string
	UpdatedAt            time.Time
	StatusName           sql.NullString
	ProjectName          sql.NullString
	SalespersonName      sql.NullString
}
