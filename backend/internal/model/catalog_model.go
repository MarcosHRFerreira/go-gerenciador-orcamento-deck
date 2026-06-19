package model

import (
	"database/sql"
	"time"
)

type BudgetStatusModel struct {
	ID          int64
	Code        string
	Name        string
	Description string
	IsFinal     bool
	SortOrder   int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type LossReasonModel struct {
	ID          int64
	Code        string
	Name        string
	Description string
	Active      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type PriorityModel struct {
	ID        int64
	Code      string
	Name      string
	Weight    int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type SalespersonModel struct {
	ID        int64
	Name      string
	Email     string
	Phone     string
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type InstallerModel struct {
	ID        int64
	Name      string
	Document  string
	Email     string
	Phone     string
	City      string
	State     string
	Notes     string
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ContactModel struct {
	ID          int64
	InstallerID int64
	Name        string
	Email       string
	Phone       string
	Role        string
	IsPrimary   bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ProjectTypeModel struct {
	ID          int64
	Code        string
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ProjectModel struct {
	ID            int64
	Code          string
	Name          string
	ProjectTypeID sql.NullInt64
	City          string
	State         string
	Notes         string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type ProductLineModel struct {
	ID          int64
	Code        string
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type SystemTypeModel struct {
	ID          int64
	Code        string
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
