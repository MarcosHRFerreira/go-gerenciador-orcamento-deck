package dto

import "time"

type CreateConversationRequest struct {
	ParticipantUserID int64  `json:"participant_user_id" validate:"required"`
	ProjectID         *int64 `json:"project_id,omitempty"`
	InitialMessage    string `json:"initial_message" validate:"required,min=1,max=5000"`
}

type CreateConversationResponse struct {
	ID int64 `json:"id"`
}

type SendConversationMessageRequest struct {
	Body string `json:"body" validate:"required,min=1,max=5000"`
}

type SendConversationMessageResponse struct {
	ID int64 `json:"id"`
}

type ConversationParticipantResponse struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

type ConversationProjectResponse struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type ConversationListItemResponse struct {
	ID                int64                         `json:"id"`
	Type              string                        `json:"type"`
	UpdatedAt         time.Time                     `json:"updated_at"`
	Project           *ConversationProjectResponse  `json:"project,omitempty"`
	Participant       ConversationParticipantResponse `json:"participant"`
	LastMessageID     *int64                        `json:"last_message_id,omitempty"`
	LastMessageBody   *string                       `json:"last_message_body,omitempty"`
	LastMessageAt     *time.Time                    `json:"last_message_at,omitempty"`
	LastMessageSender *int64                        `json:"last_message_sender,omitempty"`
	UnreadCount       int64                         `json:"unread_count"`
}

type ConversationMessageResponse struct {
	ID             int64                           `json:"id"`
	ConversationID int64                           `json:"conversation_id"`
	Sender         ConversationParticipantResponse `json:"sender"`
	Body           string                          `json:"body"`
	CreatedAt      time.Time                       `json:"created_at"`
	UpdatedAt      time.Time                       `json:"updated_at"`
}

type ConversationUnreadCountResponse struct {
	Count int64 `json:"count"`
}
