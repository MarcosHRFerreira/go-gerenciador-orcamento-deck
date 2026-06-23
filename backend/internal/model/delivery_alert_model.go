package model

import "time"

type DeliveryAlertType string

const (
	DeliveryAlertTypeDueInTwoDays DeliveryAlertType = "delivery_due_in_2_days"
)

type DeliveryAlertCandidateModel struct {
	BudgetID            int64
	BudgetNumber        string
	ProjectID           *int64
	ProjectCode         *string
	ProjectName         *string
	SalespersonID       int64
	SalespersonName     string
	SalespersonEmail    string
	RecipientUserID     int64
	RecipientUserName   string
	RecipientUsername   string
	DeliveryDate        time.Time
	DaysUntilDelivery   int
	ConstructionCompany string
}

type DeliveryAlertEventModel struct {
	ID              int64
	BudgetID        int64
	RecipientUserID int64
	ConversationID  int64
	MessageID       int64
	AlertType       DeliveryAlertType
	DeliveryDate    time.Time
	SentAt          time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
