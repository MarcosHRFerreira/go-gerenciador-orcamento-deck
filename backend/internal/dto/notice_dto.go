package dto

import "time"

type CreateNoticeRequest struct {
	Title            string     `json:"title" validate:"required,min=3,max=180"`
	Body             string     `json:"body" validate:"required,min=3,max=5000"`
	ScopeType        string     `json:"scope_type" validate:"required,oneof=all users"`
	Priority         string     `json:"priority" validate:"required,oneof=info warning critical"`
	Pinned           bool       `json:"pinned"`
	ExpiresAt        *time.Time `json:"expires_at,omitempty"`
	RecipientUserIDs []int64    `json:"recipient_user_ids"`
}

type CreateNoticeResponse struct {
	ID int64 `json:"id"`
}

type NoticeResponse struct {
	ID                int64      `json:"id"`
	Title             string     `json:"title"`
	Body              string     `json:"body"`
	ScopeType         string     `json:"scope_type"`
	Priority          string     `json:"priority"`
	Pinned            bool       `json:"pinned"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty"`
	CreatedByUserID   int64      `json:"created_by_user_id"`
	CreatedByUserName string     `json:"created_by_user_name"`
	ReadAt            *time.Time `json:"read_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type NoticeUnreadCountResponse struct {
	Count int64 `json:"count"`
}
