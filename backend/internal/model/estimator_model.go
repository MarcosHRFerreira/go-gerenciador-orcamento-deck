package model

import (
	"database/sql"
	"time"
)

type EstimatorModel struct {
	ID        int64
	Code      string
	Name      string
	Email     string
	Phone     string
	Active    bool
	Notes     string
	UserID    sql.NullInt64
	CreatedAt time.Time
	UpdatedAt time.Time
	UserName  sql.NullString
}
