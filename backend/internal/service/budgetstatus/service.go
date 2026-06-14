package budgetstatus

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/apperror"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	budgetstatusrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/budgetstatus"
	"github.com/jackc/pgx/v5/pgconn"
)

type Service interface {
	Create(ctx context.Context, req *dto.CreateBudgetStatusRequest) (int64, error)
	List(ctx context.Context) ([]dto.BudgetStatusResponse, error)
	GetByID(ctx context.Context, statusID int64) (*dto.BudgetStatusResponse, error)
	Update(ctx context.Context, statusID int64, req *dto.UpdateBudgetStatusRequest) error
	Delete(ctx context.Context, statusID int64) error
}

type service struct {
	repo budgetstatusrepository.Repository
}

func NewService(repo budgetstatusrepository.Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, req *dto.CreateBudgetStatusRequest) (int64, error) {
	code := strings.TrimSpace(req.Code)
	name := strings.TrimSpace(req.Name)
	existingItem, err := s.repo.GetByCodeOrName(ctx, code, name)
	if err != nil {
		return 0, apperror.Internal("failed to check existing budget status", err)
	}
	if existingItem != nil {
		return 0, apperror.Conflict("budget status already exists")
	}

	now := time.Now()
	id, err := s.repo.Create(ctx, &model.BudgetStatusModel{
		Code:        code,
		Name:        name,
		Description: strings.TrimSpace(req.Description),
		IsFinal:     req.IsFinal,
		SortOrder:   req.SortOrder,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		return 0, mapBudgetStatusPersistenceError("create", err)
	}

	return id, nil
}

func (s *service) List(ctx context.Context) ([]dto.BudgetStatusResponse, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, apperror.Internal("failed to list budget statuses", err)
	}

	response := make([]dto.BudgetStatusResponse, 0, len(items))
	for _, item := range items {
		response = append(response, dto.BudgetStatusResponse{
			ID:          item.ID,
			Code:        item.Code,
			Name:        item.Name,
			Description: item.Description,
			IsFinal:     item.IsFinal,
			SortOrder:   item.SortOrder,
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
		})
	}

	return response, nil
}

func (s *service) GetByID(ctx context.Context, statusID int64) (*dto.BudgetStatusResponse, error) {
	if statusID <= 0 {
		return nil, apperror.BadRequest("status_id is required")
	}

	item, err := s.repo.GetByID(ctx, statusID)
	if err != nil {
		return nil, apperror.Internal("failed to get budget status", err)
	}
	if item == nil {
		return nil, apperror.NotFound("budget status not found")
	}

	response := toResponse(*item)
	return &response, nil
}

func (s *service) Update(ctx context.Context, statusID int64, req *dto.UpdateBudgetStatusRequest) error {
	if statusID <= 0 {
		return apperror.BadRequest("status_id is required")
	}

	currentItem, err := s.repo.GetByID(ctx, statusID)
	if err != nil {
		return apperror.Internal("failed to check budget status", err)
	}
	if currentItem == nil {
		return apperror.NotFound("budget status not found")
	}

	code := strings.TrimSpace(req.Code)
	if code == "" {
		return apperror.BadRequest("code is required")
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return apperror.BadRequest("name is required")
	}

	existingItem, err := s.repo.GetByCodeOrName(ctx, code, name)
	if err != nil {
		return apperror.Internal("failed to check existing budget status", err)
	}
	if existingItem != nil && existingItem.ID != statusID {
		return apperror.Conflict("budget status already exists")
	}

	err = s.repo.Update(ctx, &model.BudgetStatusModel{
		ID:          statusID,
		Code:        code,
		Name:        name,
		Description: strings.TrimSpace(req.Description),
		IsFinal:     req.IsFinal,
		SortOrder:   req.SortOrder,
		UpdatedAt:   time.Now(),
	})
	if err != nil {
		return mapBudgetStatusPersistenceError("update", err)
	}

	return nil
}

func (s *service) Delete(ctx context.Context, statusID int64) error {
	if statusID <= 0 {
		return apperror.BadRequest("status_id is required")
	}

	item, err := s.repo.GetByID(ctx, statusID)
	if err != nil {
		return apperror.Internal("failed to check budget status", err)
	}
	if item == nil {
		return apperror.NotFound("budget status not found")
	}

	if err := s.repo.Delete(ctx, statusID); err != nil {
		return mapBudgetStatusPersistenceError("delete", err)
	}

	return nil
}

func toResponse(item model.BudgetStatusModel) dto.BudgetStatusResponse {
	return dto.BudgetStatusResponse{
		ID:          item.ID,
		Code:        item.Code,
		Name:        item.Name,
		Description: item.Description,
		IsFinal:     item.IsFinal,
		SortOrder:   item.SortOrder,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}

func mapBudgetStatusPersistenceError(action string, err error) error {
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) {
		switch pgError.ConstraintName {
		case "budget_statuses_code_key", "budget_statuses_name_key":
			return apperror.Conflict("budget status already exists")
		case "fk_budgets_status_id", "fk_budget_status_history_to_status_id":
			return apperror.Conflict("budget status is being used and cannot be deleted")
		}

		if pgError.Code == "23505" {
			return apperror.Conflict("budget status already exists")
		}
		if pgError.Code == "23503" {
			return apperror.Conflict("budget status is being used and cannot be deleted")
		}
	}

	return apperror.Internal("failed to "+action+" budget status", err)
}
