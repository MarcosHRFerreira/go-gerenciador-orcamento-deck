package budgetstatushistory

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/accessscope"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/apperror"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	budgetrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/budget"
	budgetstatusrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/budgetstatus"
	budgetstatushistoryrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/budgetstatushistory"
	salespersonrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/salesperson"
)

type Service interface {
	ChangeStatus(ctx context.Context, budgetID int64, userID int64, role model.UserRole, username string, req *dto.ChangeBudgetStatusRequest) (int64, error)
	ListByBudgetID(ctx context.Context, budgetID int64, role model.UserRole, username string) ([]dto.BudgetStatusHistoryResponse, error)
}

type service struct {
	repo             budgetstatushistoryrepository.Repository
	budgetRepo       budgetrepository.Repository
	budgetStatusRepo budgetstatusrepository.Repository
	salespersonRepo  salespersonrepository.Repository
}

func NewService(
	repo budgetstatushistoryrepository.Repository,
	budgetRepo budgetrepository.Repository,
	budgetStatusRepo budgetstatusrepository.Repository,
	salespersonRepo salespersonrepository.Repository,
) Service {
	return &service{
		repo:             repo,
		budgetRepo:       budgetRepo,
		budgetStatusRepo: budgetStatusRepo,
		salespersonRepo:  salespersonRepo,
	}
}

func (s *service) ChangeStatus(ctx context.Context, budgetID int64, userID int64, role model.UserRole, username string, req *dto.ChangeBudgetStatusRequest) (int64, error) {
	if budgetID <= 0 {
		return 0, apperror.BadRequest("budget_id e obrigatorio")
	}

	if userID <= 0 {
		return 0, apperror.Unauthorized("Usuario autenticado obrigatorio")
	}

	if req.StatusID <= 0 {
		return 0, apperror.BadRequest("status_id e obrigatorio")
	}

	restrictedSalespersonID, err := accessscope.ResolveRestrictedSalespersonID(ctx, role, username, s.salespersonRepo)
	if err != nil {
		return 0, err
	}

	budget, err := s.budgetRepo.GetByIDScoped(ctx, budgetID, restrictedSalespersonID)
	if err != nil {
		return 0, apperror.Internal("failed to check budget", err)
	}
	if budget == nil {
		return 0, apperror.NotFound("Orcamento nao encontrado")
	}

	status, err := s.budgetStatusRepo.GetByID(ctx, req.StatusID)
	if err != nil {
		return 0, apperror.Internal("failed to check budget status", err)
	}
	if status == nil {
		return 0, apperror.BadRequest("Status de orcamento nao encontrado")
	}

	if budget.StatusID == req.StatusID {
		return 0, apperror.Conflict("O orcamento ja possui o status informado")
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

func (s *service) ListByBudgetID(ctx context.Context, budgetID int64, role model.UserRole, username string) ([]dto.BudgetStatusHistoryResponse, error) {
	if budgetID <= 0 {
		return nil, apperror.BadRequest("budget_id e obrigatorio")
	}

	restrictedSalespersonID, err := accessscope.ResolveRestrictedSalespersonID(ctx, role, username, s.salespersonRepo)
	if err != nil {
		return nil, err
	}

	budget, err := s.budgetRepo.GetByIDScoped(ctx, budgetID, restrictedSalespersonID)
	if err != nil {
		return nil, apperror.Internal("failed to check budget", err)
	}
	if budget == nil {
		return nil, apperror.NotFound("Orcamento nao encontrado")
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
