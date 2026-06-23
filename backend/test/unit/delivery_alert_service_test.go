package unit

import (
	"context"
	"testing"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/config"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	deliveryalertservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/deliveryalert"
)

type deliveryAlertConversationRepositoryStub struct {
	capturedActorUserID       int64
	capturedParticipantUserID int64
	capturedProjectID         *int64
	capturedInitialMessage    string
	createConversationID      int64
	createMessageID           int64
	createErr                 error
}

func (s *deliveryAlertConversationRepositoryStub) CreateOrAppendDirect(_ context.Context, actorUserID int64, participantUserID int64, projectID *int64, initialMessage string, _ time.Time) (int64, int64, error) {
	s.capturedActorUserID = actorUserID
	s.capturedParticipantUserID = participantUserID
	s.capturedProjectID = projectID
	s.capturedInitialMessage = initialMessage
	return s.createConversationID, s.createMessageID, s.createErr
}

func (s *deliveryAlertConversationRepositoryStub) ListByUser(_ context.Context, _ int64) ([]model.ConversationListItemModel, error) {
	return nil, nil
}

func (s *deliveryAlertConversationRepositoryStub) ListMessagesByConversation(_ context.Context, _ int64, _ int64) ([]model.ConversationMessageDetailsModel, error) {
	return nil, nil
}

func (s *deliveryAlertConversationRepositoryStub) CountUnreadByUser(_ context.Context, _ int64) (int64, error) {
	return 0, nil
}

func (s *deliveryAlertConversationRepositoryStub) SendMessage(_ context.Context, _ int64, _ int64, _ string, _ time.Time) (int64, error) {
	return 0, nil
}

func (s *deliveryAlertConversationRepositoryStub) MarkAsRead(_ context.Context, _ int64, _ int64, _ time.Time) error {
	return nil
}

type deliveryAlertRepositoryStub struct {
	candidates         []model.DeliveryAlertCandidateModel
	listErr            error
	createEventID      int64
	createEventErr     error
	capturedCreateEvent *model.DeliveryAlertEventModel
}

func (s *deliveryAlertRepositoryStub) ListDueInTwoDaysCandidates(_ context.Context, _ time.Time) ([]model.DeliveryAlertCandidateModel, error) {
	return s.candidates, s.listErr
}

func (s *deliveryAlertRepositoryStub) CreateEvent(_ context.Context, item *model.DeliveryAlertEventModel) (int64, error) {
	s.capturedCreateEvent = item
	return s.createEventID, s.createEventErr
}

type noticeRepositoryStub struct {
	createID                  int64
	createErr                 error
	capturedNotice            *model.NoticeModel
	capturedRecipientUserIDs  []int64
}

func (s *noticeRepositoryStub) Create(_ context.Context, notice *model.NoticeModel, recipientUserIDs []int64) (int64, error) {
	s.capturedNotice = notice
	s.capturedRecipientUserIDs = append([]int64(nil), recipientUserIDs...)
	return s.createID, s.createErr
}

func (s *noticeRepositoryStub) ListByUser(_ context.Context, _ int64, _ string, _ time.Time) ([]model.NoticeListItemModel, error) {
	return nil, nil
}

func (s *noticeRepositoryStub) GetByIDForUser(_ context.Context, _ int64, _ int64, _ time.Time) (*model.NoticeListItemModel, error) {
	return nil, nil
}

func (s *noticeRepositoryStub) CountUnreadByUser(_ context.Context, _ int64, _ time.Time) (int64, error) {
	return 0, nil
}

func (s *noticeRepositoryStub) MarkAsRead(_ context.Context, _ int64, _ int64, _ time.Time) error {
	return nil
}

func TestDeliveryAlertServiceRunShouldCreateConversationAndAdminNotice(t *testing.T) {
	projectID := int64(22)
	projectCode := "OBR-0022"
	projectName := "Edificio Central"

	conversationRepo := &deliveryAlertConversationRepositoryStub{
		createConversationID: 9,
		createMessageID:      14,
	}
	deliveryAlertRepo := &deliveryAlertRepositoryStub{
		candidates: []model.DeliveryAlertCandidateModel{
			{
				BudgetID:            101,
				BudgetNumber:        "10019",
				ProjectID:           &projectID,
				ProjectCode:         &projectCode,
				ProjectName:         &projectName,
				RecipientUserID:     31,
				RecipientUserName:   "Guilherme",
				RecipientUsername:   "guilherme",
				DeliveryDate:        mustParseDate(t, "2026-06-25"),
				DaysUntilDelivery:   2,
				ConstructionCompany: "Construtora XPTO",
			},
		},
		createEventID: 77,
	}
	noticeRepo := &noticeRepositoryStub{
		createID: 55,
	}
	userRepo := &userRepositoryStub{
		getUserByUsernameItem: &model.UserModel{
			ID:       1,
			Name:     "Admin Sistema",
			Username: "admin_alerta",
			Active:   true,
			Role:     model.RoleAdmin,
		},
		listActiveUsersItems: []model.UserModel{
			{ID: 1, Name: "Admin Sistema", Active: true, Role: model.RoleAdmin},
			{ID: 2, Name: "Admin Gestor", Active: true, Role: model.RoleAdmin},
			{ID: 31, Name: "Guilherme", Active: true, Role: model.RoleUser},
		},
	}

	service := deliveryalertservice.NewService(
		&config.Config{DeliveryAlertSenderUsername: "admin_alerta"},
		conversationRepo,
		deliveryAlertRepo,
		noticeRepo,
		userRepo,
	)

	err := service.Run(context.Background())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if conversationRepo.capturedActorUserID != 1 {
		t.Fatalf("expected sender user id 1, got %d", conversationRepo.capturedActorUserID)
	}
	if conversationRepo.capturedParticipantUserID != 31 {
		t.Fatalf("expected recipient user id 31, got %d", conversationRepo.capturedParticipantUserID)
	}
	if conversationRepo.capturedInitialMessage == "" {
		t.Fatal("expected automatic conversation message to be created")
	}
	if noticeRepo.capturedNotice == nil {
		t.Fatal("expected admin notice to be created")
	}
	if noticeRepo.capturedNotice.ScopeType != model.NoticeScopeUsers {
		t.Fatalf("expected directed admin notice, got %s", noticeRepo.capturedNotice.ScopeType)
	}
	if noticeRepo.capturedNotice.Priority != model.NoticePriorityWarning {
		t.Fatalf("expected warning priority, got %s", noticeRepo.capturedNotice.Priority)
	}
	if noticeRepo.capturedNotice.Title != "Alerta automatico de entrega do orcamento 10019" {
		t.Fatalf("unexpected notice title: %s", noticeRepo.capturedNotice.Title)
	}
	if len(noticeRepo.capturedRecipientUserIDs) != 2 {
		t.Fatalf("expected 2 admin recipients, got %d", len(noticeRepo.capturedRecipientUserIDs))
	}
	if noticeRepo.capturedRecipientUserIDs[0] != 1 || noticeRepo.capturedRecipientUserIDs[1] != 2 {
		t.Fatalf("unexpected admin recipients: %#v", noticeRepo.capturedRecipientUserIDs)
	}
	if deliveryAlertRepo.capturedCreateEvent == nil {
		t.Fatal("expected delivery alert event to be created")
	}
	if deliveryAlertRepo.capturedCreateEvent.ConversationID != 9 {
		t.Fatalf("expected conversation id 9, got %d", deliveryAlertRepo.capturedCreateEvent.ConversationID)
	}
	if deliveryAlertRepo.capturedCreateEvent.MessageID != 14 {
		t.Fatalf("expected message id 14, got %d", deliveryAlertRepo.capturedCreateEvent.MessageID)
	}
}

func mustParseDate(t *testing.T, value string) time.Time {
	t.Helper()

	parsedValue, err := time.Parse("2006-01-02", value)
	if err != nil {
		t.Fatalf("expected valid date %s, got error %v", value, err)
	}

	return parsedValue
}
