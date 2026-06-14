package unit

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
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
	getByIDItem *model.BudgetStatusModel
	getByIDErr  error
}

func (s *budgetStatusRepositoryStub) Create(_ context.Context, _ *model.BudgetStatusModel) (int64, error) {
	return 0, errors.New("not implemented")
}

func (s *budgetStatusRepositoryStub) List(_ context.Context) ([]model.BudgetStatusModel, error) {
	return nil, errors.New("not implemented")
}

func (s *budgetStatusRepositoryStub) GetByCodeOrName(_ context.Context, _ string, _ string) (*model.BudgetStatusModel, error) {
	return nil, errors.New("not implemented")
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
	service := budgetfollowupservice.NewService(followUpRepo, budgetRepo)
	followUpAt := time.Date(2026, time.July, 1, 15, 0, 0, 0, time.UTC)

	id, err := service.Create(context.Background(), 8, 21, &dto.CreateBudgetFollowUpRequest{
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
	service := budgetfollowupservice.NewService(&budgetFollowUpRepositoryStub{}, &budgetRepositoryStub{})

	_, err := service.Create(context.Background(), 8, 21, &dto.CreateBudgetFollowUpRequest{
		Notes: "retorno",
	})

	assertAppError(t, err, 404, "budget not found")
}

func TestBudgetFollowUpServiceCreateShouldReturnUnauthorizedWhenUserIsMissing(t *testing.T) {
	service := budgetfollowupservice.NewService(&budgetFollowUpRepositoryStub{}, &budgetRepositoryStub{})

	_, err := service.Create(context.Background(), 8, 0, &dto.CreateBudgetFollowUpRequest{
		Notes: "retorno",
	})

	assertAppError(t, err, 401, "authenticated user is required")
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
	})

	items, err := service.ListByBudgetID(context.Background(), 8)

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
	historyRepo := &budgetStatusHistoryRepositoryStub{
		createID: 44,
	}
	budgetRepo := &budgetRepositoryStub{
		getByIDItem: &model.BudgetModel{
			ID:       9,
			StatusID: 1,
		},
	}
	statusRepo := &budgetStatusRepositoryStub{
		getByIDItem: &model.BudgetStatusModel{ID: 2},
	}
	service := budgetstatushistoryservice.NewService(historyRepo, budgetRepo, statusRepo)

	id, err := service.ChangeStatus(context.Background(), 9, 30, &dto.ChangeBudgetStatusRequest{
		StatusID: 2,
		Notes:    "  status atualizado  ",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != 44 {
		t.Fatalf("expected id 44, got %d", id)
	}
	if historyRepo.capturedCreateItem == nil {
		t.Fatal("expected history item to be captured")
	}
	if !historyRepo.capturedCreateItem.FromStatusID.Valid || historyRepo.capturedCreateItem.FromStatusID.Int64 != 1 {
		t.Fatalf("expected from status 1, got %+v", historyRepo.capturedCreateItem.FromStatusID)
	}
	if historyRepo.capturedCreateItem.ToStatusID != 2 {
		t.Fatalf("expected to status 2, got %d", historyRepo.capturedCreateItem.ToStatusID)
	}
	if historyRepo.capturedCreateItem.Notes != "status atualizado" {
		t.Fatalf("expected trimmed notes, got %s", historyRepo.capturedCreateItem.Notes)
	}
	if budgetRepo.capturedStatusID != 2 {
		t.Fatalf("expected synced status id 2, got %d", budgetRepo.capturedStatusID)
	}
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
	)

	_, err := service.ChangeStatus(context.Background(), 9, 30, &dto.ChangeBudgetStatusRequest{
		StatusID: 2,
	})

	assertAppError(t, err, 409, "budget already has informed status")
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
	)

	_, err := service.ChangeStatus(context.Background(), 9, 30, &dto.ChangeBudgetStatusRequest{
		StatusID: 2,
	})

	assertAppError(t, err, 400, "budget status not found")
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
	)

	items, err := service.ListByBudgetID(context.Background(), 9)

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
