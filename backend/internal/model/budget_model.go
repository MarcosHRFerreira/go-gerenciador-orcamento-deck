package model

import (
	"database/sql"
	"time"
)

type BudgetModel struct {
	ID                   int64
	BudgetNumber         string
	YearBudget           int
	Revision             int
	SentAt               time.Time
	GrossValue           float64
	CommissionValue      float64
	AreaM2               float64
	StatusID             int64
	PriorityID           sql.NullInt64
	InstallerID          sql.NullInt64
	ProjectID            sql.NullInt64
	SalespersonID        sql.NullInt64
	ContactID            sql.NullInt64
	LossReasonID         sql.NullInt64
	CompetitorName       string
	CompetitorPrice      sql.NullFloat64
	DesignerName         string
	SpecificationDetails string
	CurrentFollowUp      string
	CreatedAt            time.Time
	UpdatedAt            time.Time
	StatusName           sql.NullString
	PriorityName         sql.NullString
	InstallerName        sql.NullString
	ProjectName          sql.NullString
	SalespersonName      sql.NullString
	ContactName          sql.NullString
	LossReasonName       sql.NullString
}
