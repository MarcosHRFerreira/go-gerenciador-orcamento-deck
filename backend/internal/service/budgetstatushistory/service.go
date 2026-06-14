package budgetstatushistory

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/apperror"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	budgetrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/budget"
	budgetstatushistoryrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/budgetstatushistory"
	budgetstatusrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/budgetstatus"
)

type Service interface {
	ChangeStatus(ctx context.Context, budgetID int64, userID int64, req *dto.ChangeBudgetStatusRequest) (int64, error)
	ListByBudgetID(ctx context.Context, budgetID int64) ([]dto.BudgetStatusHistoryResponse, error)
}

type service struct {
	repo             budgetstatushistoryrepository.Repository
	budgetRepo       budgetrepository.Repository
	budgetStatusRepo budgetstatusrepository.Repository
}

func NewService(
	repo budgetstatushistoryrepository.Repository,
	budgetRepo budgetrepository.Repository,
	budgetStatusRepo budgetstatusrepository.Repository,
) Service {
	return &service{
		repo:             repo,
		budgetRepo:       budgetRepo,
		budgetStatusRepo: budgetStatusRepo,
	}
}

func (s *service) ChangeStatus(ctx context.Context, budgetID int64, userID int64, req *dto.ChangeBudgetStatusRequest) (int64, error) {
	if budgetID <= 0 {
		return 0, apperror.BadRequest("budget_id is required")
	}

	if userID <= 0 {
		return 0, apperror.Unauthorized("authenticated user is required")
	}

	if req.StatusID <= 0 {
		return 0, apperror.BadRequest("status_id is required")
	}

	budget, err := s.budgetRepo.GetByID(ctx, budgetID)
	if err != nil {
		return 0, apperror.Internal("failed to check budget", err)
	}
	if budget == nil {
		return 0, apperror.NotFound("budget not found")
	}

	status, err := s.budgetStatusRepo.GetByID(ctx, req.StatusID)
	if err != nil {
		return 0, apperror.Internal("failed to check budget status", err)
	}
	if status == nil {
		return 0, apperror.BadRequest("budget status not found")
	}

	if budget.StatusID == req.StatusID {
		return 0, apperror.Conflict("budget already has informed status")
	}

	notes := strings.TrimSpace(req.Notes)
	now := time.Now()
	id, err := s.repo.Create(ctx, &model.BudgetStatusHistoryModel{
		BudgetID: budgetID,
		FromStatusID: sql.NullInt64{
			Int64: budget.StatusID,
			Valid: true,
		},
		ToStatusID:      req.StatusID,
		ChangedByUserID: userID,
		Notes:           notes,
		ChangedAt:       now,
		CreatedAt:       now,
		UpdatedAt:       now,
	})
	if err != nil {
		return 0, apperror.Internal("failed to create budget status history", err)
	}

	if err := s.budgetRepo.UpdateStatus(ctx, budgetID, req.StatusID, now); err != nil {
		return 0, apperror.Internal("failed to update budget status", err)
	}

	return id, nil
}

func (s *service) ListByBudgetID(ctx context.Context, budgetID int64) ([]dto.BudgetStatusHistoryResponse, error) {
	if budgetID <= 0 {
		return nil, apperror.BadRequest("budget_id is required")
	}

	budget, err := s.budgetRepo.GetByID(ctx, budgetID)
	if err != nil {
		return nil, apperror.Internal("failed to check budget", err)
	}
	if budget == nil {
		return nil, apperror.NotFound("budget not found")
	}

	items, err := s.repo.ListByBudgetID(ctx, budgetID)
	if err != nil {
		return nil, apperror.Internal("failed to list budget status history", err)
	}

	response := make([]dto.BudgetStatusHistoryResponse, 0, len(items))
	for _, item := range items {
		response = append(response, dto.BudgetStatusHistoryResponse{
			ID:              item.ID,
			BudgetID:        item.BudgetID,
			FromStatusID:    nullableInt64Pointer(item.FromStatusID),
			ToStatusID:      item.ToStatusID,
			ChangedByUserID: item.ChangedByUserID,
			Notes:           item.Notes,
			ChangedAt:       item.ChangedAt,
			CreatedAt:       item.CreatedAt,
			UpdatedAt:       item.UpdatedAt,
		})
	}

	return response, nil
}

func nullableInt64Pointer(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}

	return &value.Int64
}
