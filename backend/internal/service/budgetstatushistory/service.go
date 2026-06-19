package budgetstatushistory

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/accessscope"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/apperror"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	budgetrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/budget"
	budgetstatusrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/budgetstatus"
	budgetstatushistoryrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/budgetstatushistory"
	estimatorrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/estimator"
	salespersonrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/salesperson"
	userrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/user"
)

type Service interface {
	ChangeStatus(ctx context.Context, budgetID int64, userID int64, role model.UserRole, username string, req *dto.ChangeBudgetStatusRequest) (int64, error)
	ListByBudgetID(ctx context.Context, budgetID int64, role model.UserRole, username string) ([]dto.BudgetStatusHistoryResponse, error)
}

type service struct {
	repo             budgetstatushistoryrepository.Repository
	budgetRepo       budgetrepository.Repository
	budgetStatusRepo budgetstatusrepository.Repository
	userRepo         userrepository.Repository
	salespersonRepo  salespersonrepository.Repository
	estimatorRepo    estimatorrepository.Repository
}

func NewService(
	repo budgetstatushistoryrepository.Repository,
	budgetRepo budgetrepository.Repository,
	budgetStatusRepo budgetstatusrepository.Repository,
	userRepo userrepository.Repository,
	salespersonRepo salespersonrepository.Repository,
	estimatorRepo estimatorrepository.Repository,
) Service {
	return &service{
		repo:             repo,
		budgetRepo:       budgetRepo,
		budgetStatusRepo: budgetStatusRepo,
		userRepo:         userRepo,
		salespersonRepo:  salespersonRepo,
		estimatorRepo:    estimatorRepo,
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

	scope, err := accessscope.ResolveBudgetScope(ctx, role, username, s.userRepo, s.salespersonRepo, s.estimatorRepo)
	if err != nil {
		return 0, err
	}

	budget, err := s.budgetRepo.GetByIDScoped(ctx, budgetID, scope.RestrictedSalespersonID, scope.RestrictedEstimatorID)
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
	changeStatusParams := &budgetrepository.ChangeStatusParams{
		BudgetID:  budgetID,
		StatusID:  req.StatusID,
		UserID:    userID,
		Notes:     notes,
		ChangedAt: now,
	}

	if isPedidoStatus(status) && budget.ProjectID.Valid {
		cancelledStatus, getCancelledStatusErr := s.budgetStatusRepo.GetByCodeOrName(ctx, "CANCELADO", "CANCELADO")
		if getCancelledStatusErr != nil {
			return 0, apperror.Internal("failed to check cancelled budget status", getCancelledStatusErr)
		}
		if cancelledStatus == nil {
			return 0, apperror.BadRequest("Status CANCELADO nao encontrado")
		}

		changeStatusParams.EnforceProjectWinnerRule = true
		changeStatusParams.CancelledStatusID = cancelledStatus.ID
	}

	id, err := s.budgetRepo.ChangeStatus(ctx, changeStatusParams)
	if err != nil {
		if errors.Is(err, budgetrepository.ErrProjectAlreadyHasPedido) {
			return 0, apperror.Conflict("Ja existe outro orcamento da obra marcado como PEDIDO")
		}

		return 0, apperror.Internal("failed to change budget status", err)
	}

	return id, nil
}

func (s *service) ListByBudgetID(ctx context.Context, budgetID int64, role model.UserRole, username string) ([]dto.BudgetStatusHistoryResponse, error) {
	if budgetID <= 0 {
		return nil, apperror.BadRequest("budget_id e obrigatorio")
	}

	scope, err := accessscope.ResolveBudgetScope(ctx, role, username, s.userRepo, s.salespersonRepo, s.estimatorRepo)
	if err != nil {
		return nil, err
	}

	budget, err := s.budgetRepo.GetByIDScoped(ctx, budgetID, scope.RestrictedSalespersonID, scope.RestrictedEstimatorID)
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

func isPedidoStatus(status *model.BudgetStatusModel) bool {
	if status == nil {
		return false
	}

	return strings.EqualFold(strings.TrimSpace(status.Code), "PEDIDO") || strings.EqualFold(strings.TrimSpace(status.Name), "PEDIDO")
}
