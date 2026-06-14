package budgetfollowup

import (
	"context"
	"strings"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/apperror"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	budgetrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/budget"
	budgetfollowuprepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/budgetfollowup"
)

type Service interface {
	Create(ctx context.Context, budgetID int64, userID int64, req *dto.CreateBudgetFollowUpRequest) (int64, error)
	ListByBudgetID(ctx context.Context, budgetID int64) ([]dto.BudgetFollowUpResponse, error)
}

type service struct {
	repo       budgetfollowuprepository.Repository
	budgetRepo budgetrepository.Repository
}

func NewService(repo budgetfollowuprepository.Repository, budgetRepo budgetrepository.Repository) Service {
	return &service{
		repo:       repo,
		budgetRepo: budgetRepo,
	}
}

func (s *service) Create(ctx context.Context, budgetID int64, userID int64, req *dto.CreateBudgetFollowUpRequest) (int64, error) {
	if budgetID <= 0 {
		return 0, apperror.BadRequest("budget_id is required")
	}

	if userID <= 0 {
		return 0, apperror.Unauthorized("authenticated user is required")
	}

	budget, err := s.budgetRepo.GetByID(ctx, budgetID)
	if err != nil {
		return 0, apperror.Internal("failed to check budget", err)
	}
	if budget == nil {
		return 0, apperror.NotFound("budget not found")
	}

	notes := strings.TrimSpace(req.Notes)
	if notes == "" {
		return 0, apperror.BadRequest("notes is required")
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

func (s *service) ListByBudgetID(ctx context.Context, budgetID int64) ([]dto.BudgetFollowUpResponse, error) {
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
