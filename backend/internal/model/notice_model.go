package model

import "time"

type NoticeScopeType string

const (
	NoticeScopeAllUsers NoticeScopeType = "all"
	NoticeScopeUsers    NoticeScopeType = "users"
)

type NoticePriority string

const (
	NoticePriorityInfo     NoticePriority = "info"
	NoticePriorityWarning  NoticePriority = "warning"
	NoticePriorityCritical NoticePriority = "critical"
)

type NoticeModel struct {
	ID              int64
	Title           string
	Body            string
	ScopeType       NoticeScopeType
	Priority        NoticePriority
	Pinned          bool
	ExpiresAt       *time.Time
	CreatedByUserID int64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type NoticeRecipientModel struct {
	ID        int64
	NoticeID  int64
	UserID    int64
	ReadAt    *time.Time
	HiddenAt  *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

type NoticeListItemModel struct {
	NoticeModel
	ReadAt            *time.Time
	CreatedByUserName string
	RecipientID       int64
}
