package model

import "time"

type ConversationType string

const (
	ConversationTypeDirect ConversationType = "direct"
)

type ConversationModel struct {
	ID              int64
	Type            ConversationType
	CreatedByUserID int64
	ProjectID       *int64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type ConversationParticipantModel struct {
	ID                int64
	ConversationID    int64
	UserID            int64
	LastReadMessageID *int64
	JoinedAt          time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type ConversationMessageModel struct {
	ID             int64
	ConversationID int64
	SenderUserID   int64
	Body           string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type ConversationListItemModel struct {
	ConversationID      int64
	Type                ConversationType
	UpdatedAt           time.Time
	ProjectID           *int64
	ProjectCode         *string
	ProjectName         *string
	ParticipantUserID   int64
	ParticipantName     string
	ParticipantUsername string
	ParticipantRole     UserRole
	LastMessageID       *int64
	LastMessageBody     *string
	LastMessageAt       *time.Time
	LastMessageSenderID *int64
	UnreadCount         int64
}

type ConversationMessageDetailsModel struct {
	MessageID      int64
	ConversationID int64
	SenderUserID   int64
	SenderName     string
	SenderUsername string
	SenderRole     UserRole
	Body           string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
