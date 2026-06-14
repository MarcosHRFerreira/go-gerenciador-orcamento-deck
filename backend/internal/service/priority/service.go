package priority

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/apperror"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	priorityrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/priority"
	"github.com/jackc/pgx/v5/pgconn"
)

type Service interface {
	Create(ctx context.Context, req *dto.CreatePriorityRequest) (int64, error)
	List(ctx context.Context) ([]dto.PriorityResponse, error)
	GetByID(ctx context.Context, priorityID int64) (*dto.PriorityResponse, error)
	Update(ctx context.Context, priorityID int64, req *dto.UpdatePriorityRequest) error
	Delete(ctx context.Context, priorityID int64) error
}

type service struct {
	repo priorityrepository.Repository
}

func NewService(repo priorityrepository.Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, req *dto.CreatePriorityRequest) (int64, error) {
	code := strings.TrimSpace(req.Code)
	name := strings.TrimSpace(req.Name)

	existingItem, err := s.repo.GetByCodeOrName(ctx, code, name)
	if err != nil {
		return 0, apperror.Internal("failed to check existing priority", err)
	}
	if existingItem != nil {
		return 0, apperror.Conflict("Prioridade ja existe")
	}

	now := time.Now()
	id, err := s.repo.Create(ctx, &model.PriorityModel{
		Code:      code,
		Name:      name,
		Weight:    req.Weight,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return 0, mapPriorityPersistenceError("create", err)
	}

	return id, nil
}

func (s *service) List(ctx context.Context) ([]dto.PriorityResponse, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, apperror.Internal("failed to list priorities", err)
	}

	response := make([]dto.PriorityResponse, 0, len(items))
	for _, item := range items {
		response = append(response, toResponse(item))
	}

	return response, nil
}

func (s *service) GetByID(ctx context.Context, priorityID int64) (*dto.PriorityResponse, error) {
	if priorityID <= 0 {
		return nil, apperror.BadRequest("priority_id e obrigatorio")
	}

	item, err := s.repo.GetByID(ctx, priorityID)
	if err != nil {
		return nil, apperror.Internal("failed to get priority", err)
	}
	if item == nil {
		return nil, apperror.NotFound("Prioridade nao encontrada")
	}

	response := toResponse(*item)
	return &response, nil
}

func (s *service) Update(ctx context.Context, priorityID int64, req *dto.UpdatePriorityRequest) error {
	if priorityID <= 0 {
		return apperror.BadRequest("priority_id e obrigatorio")
	}

	currentItem, err := s.repo.GetByID(ctx, priorityID)
	if err != nil {
		return apperror.Internal("failed to check priority", err)
	}
	if currentItem == nil {
		return apperror.NotFound("Prioridade nao encontrada")
	}

	code := strings.TrimSpace(req.Code)
	if code == "" {
		return apperror.BadRequest("code e obrigatorio")
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return apperror.BadRequest("name e obrigatorio")
	}

	existingItem, err := s.repo.GetByCodeOrName(ctx, code, name)
	if err != nil {
		return apperror.Internal("failed to check existing priority", err)
	}
	if existingItem != nil && existingItem.ID != priorityID {
		return apperror.Conflict("Prioridade ja existe")
	}

	err = s.repo.Update(ctx, &model.PriorityModel{
		ID:        priorityID,
		Code:      code,
		Name:      name,
		Weight:    req.Weight,
		UpdatedAt: time.Now(),
	})
	if err != nil {
		return mapPriorityPersistenceError("update", err)
	}

	return nil
}

func (s *service) Delete(ctx context.Context, priorityID int64) error {
	if priorityID <= 0 {
		return apperror.BadRequest("priority_id e obrigatorio")
	}

	item, err := s.repo.GetByID(ctx, priorityID)
	if err != nil {
		return apperror.Internal("failed to check priority", err)
	}
	if item == nil {
		return apperror.NotFound("Prioridade nao encontrada")
	}

	if err := s.repo.Delete(ctx, priorityID); err != nil {
		return mapPriorityPersistenceError("delete", err)
	}

	return nil
}

func toResponse(item model.PriorityModel) dto.PriorityResponse {
	return dto.PriorityResponse{
		ID:        item.ID,
		Code:      item.Code,
		Name:      item.Name,
		Weight:    item.Weight,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}

func mapPriorityPersistenceError(action string, err error) error {
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) {
		switch pgError.ConstraintName {
		case "priorities_code_key", "priorities_name_key":
			return apperror.Conflict("Prioridade ja existe")
		}

		if pgError.Code == "23505" {
			return apperror.Conflict("Prioridade ja existe")
		}
	}

	return apperror.Internal("failed to "+action+" priority", err)
}
