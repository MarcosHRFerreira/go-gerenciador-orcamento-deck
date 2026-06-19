package budgetfollowup

import (
	"context"
	"strings"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/accessscope"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/apperror"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	budgetrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/budget"
	budgetfollowuprepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/budgetfollowup"
	estimatorrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/estimator"
	salespersonrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/salesperson"
	userrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/user"
)

type Service interface {
	Create(ctx context.Context, budgetID int64, userID int64, role model.UserRole, username string, req *dto.CreateBudgetFollowUpRequest) (int64, error)
	ListByBudgetID(ctx context.Context, budgetID int64, role model.UserRole, username string) ([]dto.BudgetFollowUpResponse, error)
}

type service struct {
	repo            budgetfollowuprepository.Repository
	budgetRepo      budgetrepository.Repository
	userRepo        userrepository.Repository
	salespersonRepo salespersonrepository.Repository
	estimatorRepo   estimatorrepository.Repository
}

func NewService(
	repo budgetfollowuprepository.Repository,
	budgetRepo budgetrepository.Repository,
	userRepo userrepository.Repository,
	salespersonRepo salespersonrepository.Repository,
	estimatorRepo estimatorrepository.Repository,
) Service {
	return &service{
		repo:            repo,
		budgetRepo:      budgetRepo,
		userRepo:        userRepo,
		salespersonRepo: salespersonRepo,
		estimatorRepo:   estimatorRepo,
	}
}

func (s *service) Create(ctx context.Context, budgetID int64, userID int64, role model.UserRole, username string, req *dto.CreateBudgetFollowUpRequest) (int64, error) {
	if budgetID <= 0 {
		return 0, apperror.BadRequest("budget_id e obrigatorio")
	}

	if userID <= 0 {
		return 0, apperror.Unauthorized("Usuario autenticado obrigatorio")
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

	notes := strings.TrimSpace(req.Notes)
	if notes == "" {
		return 0, apperror.BadRequest("Observacoes obrigatorias")
	}

	now := time.Now()
	followUpAt := now
	if req.FollowUpAt != nil && !req.FollowUpAt.IsZero() {
		followUpAt = *req.FollowUpAt
	}

	id, err := s.repo.Create(ctx, &model.BudgetFollowUpModel{
		BudgetID:        budgetID,
		CreatedByUserID: userID,
		Notes:           notes,
		FollowUpAt:      followUpAt,
		CreatedAt:       now,
		UpdatedAt:       now,
	})
	if err != nil {
		return 0, apperror.Internal("failed to create budget follow up", err)
	}

	if err := s.budgetRepo.UpdateCurrentFollowUp(ctx, budgetID, notes, now); err != nil {
		return 0, apperror.Internal("failed to sync current follow up", err)
	}

	return id, nil
}

func (s *service) ListByBudgetID(ctx context.Context, budgetID int64, role model.UserRole, username string) ([]dto.BudgetFollowUpResponse, error) {
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
		return nil, apperror.Internal("failed to list budget follow ups", err)
	}

	response := make([]dto.BudgetFollowUpResponse, 0, len(items))
	for _, item := range items {
		response = append(response, dto.BudgetFollowUpResponse{
			ID:              item.ID,
			BudgetID:        item.BudgetID,
			CreatedByUserID: item.CreatedByUserID,
			Notes:           item.Notes,
			FollowUpAt:      item.FollowUpAt,
			CreatedAt:       item.CreatedAt,
			UpdatedAt:       item.UpdatedAt,
		})
	}

	return response, nil
}
