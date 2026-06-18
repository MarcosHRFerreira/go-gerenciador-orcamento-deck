package unit

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	budgetrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/budget"
	budgetfollowupservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/budgetfollowup"
	budgetstatushistoryservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/budgetstatushistory"
)

type budgetFollowUpRepositoryStub struct {
	createID           int64
	createErr          error
	listItems          []model.BudgetFollowUpModel
	listErr            error
	capturedCreateItem *model.BudgetFollowUpModel
}

func (s *budgetFollowUpRepositoryStub) Create(_ context.Context, item *model.BudgetFollowUpModel) (int64, error) {
	s.capturedCreateItem = item
	return s.createID, s.createErr
}

func (s *budgetFollowUpRepositoryStub) ListByBudgetID(_ context.Context, _ int64) ([]model.BudgetFollowUpModel, error) {
	return s.listItems, s.listErr
}

type budgetStatusHistoryRepositoryStub struct {
	createID           int64
	createErr          error
	listItems          []model.BudgetStatusHistoryModel
	listErr            error
	capturedCreateItem *model.BudgetStatusHistoryModel
}

func (s *budgetStatusHistoryRepositoryStub) Create(_ context.Context, item *model.BudgetStatusHistoryModel) (int64, error) {
	s.capturedCreateItem = item
	return s.createID, s.createErr
}

func (s *budgetStatusHistoryRepositoryStub) ListByBudgetID(_ context.Context, _ int64) ([]model.BudgetStatusHistoryModel, error) {
	return s.listItems, s.listErr
}

type budgetStatusRepositoryStub struct {
	getByIDItem         *model.BudgetStatusModel
	getByIDErr          error
	getByCodeOrNameItem *model.BudgetStatusModel
	getByCodeOrNameErr  error
}

func (s *budgetStatusRepositoryStub) Create(_ context.Context, _ *model.BudgetStatusModel) (int64, error) {
	return 0, errors.New("not implemented")
}

func (s *budgetStatusRepositoryStub) List(_ context.Context) ([]model.BudgetStatusModel, error) {
	return nil, errors.New("not implemented")
}

func (s *budgetStatusRepositoryStub) GetByCodeOrName(_ context.Context, _ string, _ string) (*model.BudgetStatusModel, error) {
	return s.getByCodeOrNameItem, s.getByCodeOrNameErr
}

func (s *budgetStatusRepositoryStub) GetByID(_ context.Context, _ int64) (*model.BudgetStatusModel, error) {
	return s.getByIDItem, s.getByIDErr
}

func (s *budgetStatusRepositoryStub) Update(_ context.Context, _ *model.BudgetStatusModel) error {
	return errors.New("not implemented")
}

func (s *budgetStatusRepositoryStub) Delete(_ context.Context, _ int64) error {
	return errors.New("not implemented")
}

func TestBudgetFollowUpServiceCreateShouldTrimNotesAndSyncCurrentFollowUp(t *testing.T) {
	followUpRepo := &budgetFollowUpRepositoryStub{
		createID: 33,
	}
	budgetRepo := &budgetRepositoryStub{
		getByIDItem: &model.BudgetModel{ID: 8},
	}
	service := budgetfollowupservice.NewService(followUpRepo, budgetRepo, &salespersonRepositoryStub{})
	followUpAt := time.Date(2026, time.July, 1, 15, 0, 0, 0, time.UTC)

	id, err := service.Create(context.Background(), 8, 21, model.RoleAdmin, "", &dto.CreateBudgetFollowUpRequest{
		Notes:      "  retorno na proxima semana  ",
		FollowUpAt: &followUpAt,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != 33 {
		t.Fatalf("expected id 33, got %d", id)
	}
	if followUpRepo.capturedCreateItem == nil {
		t.Fatal("expected follow up item to be captured")
	}
	if followUpRepo.capturedCreateItem.Notes != "retorno na proxima semana" {
		t.Fatalf("expected trimmed notes, got %s", followUpRepo.capturedCreateItem.Notes)
	}
	if !followUpRepo.capturedCreateItem.FollowUpAt.Equal(followUpAt) {
		t.Fatalf("expected follow_up_at %v, got %v", followUpAt, followUpRepo.capturedCreateItem.FollowUpAt)
	}
	if budgetRepo.capturedCurrentFollowUp != "retorno na proxima semana" {
		t.Fatalf("expected synced current follow up, got %s", budgetRepo.capturedCurrentFollowUp)
	}
}

func TestBudgetFollowUpServiceCreateShouldReturnNotFoundWhenBudgetDoesNotExist(t *testing.T) {
	service := budgetfollowupservice.NewService(&budgetFollowUpRepositoryStub{}, &budgetRepositoryStub{}, &salespersonRepositoryStub{})

	_, err := service.Create(context.Background(), 8, 21, model.RoleAdmin, "", &dto.CreateBudgetFollowUpRequest{
		Notes: "retorno",
	})

	assertAppError(t, err, 404, "Orcamento nao encontrado")
}

func TestBudgetFollowUpServiceCreateShouldReturnUnauthorizedWhenUserIsMissing(t *testing.T) {
	service := budgetfollowupservice.NewService(&budgetFollowUpRepositoryStub{}, &budgetRepositoryStub{}, &salespersonRepositoryStub{})

	_, err := service.Create(context.Background(), 8, 0, model.RoleAdmin, "", &dto.CreateBudgetFollowUpRequest{
		Notes: "retorno",
	})

	assertAppError(t, err, 401, "Usuario autenticado obrigatorio")
}

func TestBudgetFollowUpServiceListShouldMapItems(t *testing.T) {
	followUpAt := time.Date(2026, time.July, 2, 10, 0, 0, 0, time.UTC)
	service := budgetfollowupservice.NewService(&budgetFollowUpRepositoryStub{
		listItems: []model.BudgetFollowUpModel{
			{
				ID:              5,
				BudgetID:        8,
				CreatedByUserID: 21,
				Notes:           "retorno registrado",
				FollowUpAt:      followUpAt,
			},
		},
	}, &budgetRepositoryStub{
		getByIDItem: &model.BudgetModel{ID: 8},
	}, &salespersonRepositoryStub{})

	items, err := service.ListByBudgetID(context.Background(), 8, model.RoleAdmin, "")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Notes != "retorno registrado" {
		t.Fatalf("expected notes retorno registrado, got %s", items[0].Notes)
	}
	if !items[0].FollowUpAt.Equal(followUpAt) {
		t.Fatalf("expected follow up at %v, got %v", followUpAt, items[0].FollowUpAt)
	}
}

func TestBudgetStatusHistoryServiceChangeStatusShouldCreateHistoryAndSyncBudget(t *testing.T) {
	budgetRepo := &budgetRepositoryStub{
		changeStatusID: 44,
		getByIDItem: &model.BudgetModel{
			ID:       9,
			StatusID: 1,
		},
	}
	statusRepo := &budgetStatusRepositoryStub{
		getByIDItem: &model.BudgetStatusModel{ID: 2},
	}
	historyRepo := &budgetStatusHistoryRepositoryStub{}
	service := budgetstatushistoryservice.NewService(historyRepo, budgetRepo, statusRepo, &salespersonRepositoryStub{})

	id, err := service.ChangeStatus(context.Background(), 9, 30, model.RoleAdmin, "", &dto.ChangeBudgetStatusRequest{
		StatusID: 2,
		Notes:    "  status atualizado  ",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != 44 {
		t.Fatalf("expected id 44, got %d", id)
	}
	if budgetRepo.capturedChangeStatusParams == nil {
		t.Fatal("expected change status params to be captured")
	}
	if budgetRepo.capturedChangeStatusParams.BudgetID != 9 {
		t.Fatalf("expected budget id 9, got %d", budgetRepo.capturedChangeStatusParams.BudgetID)
	}
	if budgetRepo.capturedChangeStatusParams.StatusID != 2 {
		t.Fatalf("expected to status 2, got %d", budgetRepo.capturedChangeStatusParams.StatusID)
	}
	if budgetRepo.capturedChangeStatusParams.Notes != "status atualizado" {
		t.Fatalf("expected trimmed notes, got %s", budgetRepo.capturedChangeStatusParams.Notes)
	}
	if budgetRepo.capturedChangeStatusParams.UserID != 30 {
		t.Fatalf("expected user id 30, got %d", budgetRepo.capturedChangeStatusParams.UserID)
	}
	if budgetRepo.capturedChangeStatusParams.EnforceProjectWinnerRule {
		t.Fatal("expected project winner rule to be disabled")
	}
}

func TestBudgetStatusHistoryServiceChangeStatusShouldCancelOtherProjectBudgetsWhenMarkedAsPedido(t *testing.T) {
	cancelledStatusID := int64(3)
	budgetRepo := &budgetRepositoryStub{
		changeStatusID: 55,
		getByIDItem: &model.BudgetModel{
			ID:        9,
			StatusID:  1,
			ProjectID: sql.NullInt64{Int64: 77, Valid: true},
		},
	}
	statusRepo := &budgetStatusRepositoryStub{
		getByIDItem:         &model.BudgetStatusModel{ID: 2, Code: "PEDIDO", Name: "Pedido"},
		getByCodeOrNameItem: &model.BudgetStatusModel{ID: cancelledStatusID, Code: "CANCELADO", Name: "Cancelado"},
	}
	service := budgetstatushistoryservice.NewService(&budgetStatusHistoryRepositoryStub{}, budgetRepo, statusRepo, &salespersonRepositoryStub{})

	id, err := service.ChangeStatus(context.Background(), 9, 30, model.RoleAdmin, "", &dto.ChangeBudgetStatusRequest{
		StatusID: 2,
		Notes:    "Virou pedido",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != 55 {
		t.Fatalf("expected id 55, got %d", id)
	}
	if budgetRepo.capturedChangeStatusParams == nil {
		t.Fatal("expected change status params to be captured")
	}
	if !budgetRepo.capturedChangeStatusParams.EnforceProjectWinnerRule {
		t.Fatal("expected project winner rule to be enabled")
	}
	if budgetRepo.capturedChangeStatusParams.CancelledStatusID != cancelledStatusID {
		t.Fatalf("expected cancelled status id %d, got %d", cancelledStatusID, budgetRepo.capturedChangeStatusParams.CancelledStatusID)
	}
}

func TestBudgetStatusHistoryServiceChangeStatusShouldReturnConflictWhenProjectAlreadyHasPedido(t *testing.T) {
	service := budgetstatushistoryservice.NewService(
		&budgetStatusHistoryRepositoryStub{},
		&budgetRepositoryStub{
			changeStatusErr: budgetrepository.ErrProjectAlreadyHasPedido,
			getByIDItem: &model.BudgetModel{
				ID:        9,
				StatusID:  1,
				ProjectID: sql.NullInt64{Int64: 77, Valid: true},
			},
		},
		&budgetStatusRepositoryStub{
			getByIDItem:         &model.BudgetStatusModel{ID: 2, Code: "PEDIDO", Name: "Pedido"},
			getByCodeOrNameItem: &model.BudgetStatusModel{ID: 3, Code: "CANCELADO", Name: "Cancelado"},
		},
		&salespersonRepositoryStub{},
	)

	_, err := service.ChangeStatus(context.Background(), 9, 30, model.RoleAdmin, "", &dto.ChangeBudgetStatusRequest{
		StatusID: 2,
	})

	assertAppError(t, err, 409, "Ja existe outro orcamento da obra marcado como PEDIDO")
}

func TestBudgetStatusHistoryServiceChangeStatusShouldReturnBadRequestWhenCancelledStatusDoesNotExist(t *testing.T) {
	service := budgetstatushistoryservice.NewService(
		&budgetStatusHistoryRepositoryStub{},
		&budgetRepositoryStub{
			getByIDItem: &model.BudgetModel{
				ID:        9,
				StatusID:  1,
				ProjectID: sql.NullInt64{Int64: 77, Valid: true},
			},
		},
		&budgetStatusRepositoryStub{
			getByIDItem: &model.BudgetStatusModel{ID: 2, Code: "PEDIDO", Name: "Pedido"},
		},
		&salespersonRepositoryStub{},
	)

	_, err := service.ChangeStatus(context.Background(), 9, 30, model.RoleAdmin, "", &dto.ChangeBudgetStatusRequest{
		StatusID: 2,
	})

	assertAppError(t, err, 400, "Status CANCELADO nao encontrado")
}

func TestBudgetStatusHistoryServiceChangeStatusShouldReturnConflictWhenStatusIsTheSame(t *testing.T) {
	service := budgetstatushistoryservice.NewService(
		&budgetStatusHistoryRepositoryStub{},
		&budgetRepositoryStub{
			getByIDItem: &model.BudgetModel{
				ID:       9,
				StatusID: 2,
			},
		},
		&budgetStatusRepositoryStub{
			getByIDItem: &model.BudgetStatusModel{ID: 2},
		},
		&salespersonRepositoryStub{},
	)

	_, err := service.ChangeStatus(context.Background(), 9, 30, model.RoleAdmin, "", &dto.ChangeBudgetStatusRequest{
		StatusID: 2,
	})

	assertAppError(t, err, 409, "O orcamento ja possui o status informado")
}

func TestBudgetStatusHistoryServiceChangeStatusShouldReturnBadRequestWhenStatusDoesNotExist(t *testing.T) {
	service := budgetstatushistoryservice.NewService(
		&budgetStatusHistoryRepositoryStub{},
		&budgetRepositoryStub{
			getByIDItem: &model.BudgetModel{
				ID:       9,
				StatusID: 1,
			},
		},
		&budgetStatusRepositoryStub{},
		&salespersonRepositoryStub{},
	)

	_, err := service.ChangeStatus(context.Background(), 9, 30, model.RoleAdmin, "", &dto.ChangeBudgetStatusRequest{
		StatusID: 2,
	})

	assertAppError(t, err, 400, "Status de orcamento nao encontrado")
}

func TestBudgetStatusHistoryServiceListShouldMapItems(t *testing.T) {
	changedAt := time.Date(2026, time.August, 1, 9, 0, 0, 0, time.UTC)
	service := budgetstatushistoryservice.NewService(
		&budgetStatusHistoryRepositoryStub{
			listItems: []model.BudgetStatusHistoryModel{
				{
					ID:           7,
					BudgetID:     9,
					FromStatusID: sql.NullInt64{Int64: 1, Valid: true},
					ToStatusID:   2,
					Notes:        "mudanca",
					ChangedAt:    changedAt,
				},
			},
		},
		&budgetRepositoryStub{
			getByIDItem: &model.BudgetModel{ID: 9},
		},
		&budgetStatusRepositoryStub{},
		&salespersonRepositoryStub{},
	)

	items, err := service.ListByBudgetID(context.Background(), 9, model.RoleAdmin, "")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].FromStatusID == nil || *items[0].FromStatusID != 1 {
		t.Fatalf("expected from status 1, got %v", items[0].FromStatusID)
	}
	if items[0].ToStatusID != 2 {
		t.Fatalf("expected to status 2, got %d", items[0].ToStatusID)
	}
	if !items[0].ChangedAt.Equal(changedAt) {
		t.Fatalf("expected changed at %v, got %v", changedAt, items[0].ChangedAt)
	}
}
