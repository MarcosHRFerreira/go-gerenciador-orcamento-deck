package model

import (
	"database/sql"
	"time"
)

type BudgetImportBatchModel struct {
	ID               int64
	PreviewID        string
	FileName         string
	SourceCompany    string
	SourceLayout     string
	Status           string
	ExecutedByUserID sql.NullInt64
	StartedAt        time.Time
	FinishedAt       sql.NullTime
	RowsProcessed    int
	BudgetsCreated   int
	BudgetsUpdated   int
	BudgetsIgnored   int
	RowsFailed       int
	CatalogsCreated  int
	ResultMessage    string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type BudgetImportRowRawModel struct {
	ID                int64
	ImportBatchID     int64
	RowNumber         int
	BudgetNumber      string
	Status            string
	Action            string
	RawRowData        []byte
	NormalizedRowData []byte
	Messages          []byte
	BudgetID          sql.NullInt64
	CreatedAt         time.Time
}
