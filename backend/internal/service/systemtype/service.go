package systemtype

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/apperror"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	systemtyperepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/systemtype"
	"github.com/jackc/pgx/v5/pgconn"
)

type Service interface {
	Create(ctx context.Context, req *dto.CreateSystemTypeRequest) (int64, error)
	List(ctx context.Context) ([]dto.SystemTypeResponse, error)
	GetByID(ctx context.Context, systemTypeID int64) (*dto.SystemTypeResponse, error)
	Update(ctx context.Context, systemTypeID int64, req *dto.UpdateSystemTypeRequest) error
	Delete(ctx context.Context, systemTypeID int64) error
}

type service struct {
	repo systemtyperepository.Repository
}

func NewService(repo systemtyperepository.Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, req *dto.CreateSystemTypeRequest) (int64, error) {
	code := strings.TrimSpace(req.Code)
	name := strings.TrimSpace(req.Name)
	description := strings.TrimSpace(req.Description)

	existingItem, err := s.repo.GetByCodeOrName(ctx, code, name)
	if err != nil {
		return 0, apperror.Internal("failed to check existing system type", err)
	}
	if existingItem != nil {
		return 0, apperror.Conflict("Tipo de sistema ja existe")
	}

	now := time.Now()
	id, err := s.repo.Create(ctx, &model.SystemTypeModel{
		Code:        code,
		Name:        name,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		return 0, mapSystemTypePersistenceError("create", err)
	}

	return id, nil
}

func (s *service) List(ctx context.Context) ([]dto.SystemTypeResponse, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, apperror.Internal("failed to list system types", err)
	}

	response := make([]dto.SystemTypeResponse, 0, len(items))
	for _, item := range items {
		response = append(response, toResponse(item))
	}

	return response, nil
}

func (s *service) GetByID(ctx context.Context, systemTypeID int64) (*dto.SystemTypeResponse, error) {
	if systemTypeID <= 0 {
		return nil, apperror.BadRequest("system_type_id e obrigatorio")
	}

	item, err := s.repo.GetByID(ctx, systemTypeID)
	if err != nil {
		return nil, apperror.Internal("failed to get system type", err)
	}
	if item == nil {
		return nil, apperror.NotFound("Tipo de sistema nao encontrado")
	}

	response := toResponse(*item)
	return &response, nil
}

func (s *service) Update(ctx context.Context, systemTypeID int64, req *dto.UpdateSystemTypeRequest) error {
	if systemTypeID <= 0 {
		return apperror.BadRequest("system_type_id e obrigatorio")
	}

	currentItem, err := s.repo.GetByID(ctx, systemTypeID)
	if err != nil {
		return apperror.Internal("failed to check system type", err)
	}
	if currentItem == nil {
		return apperror.NotFound("Tipo de sistema nao encontrado")
	}

	code := strings.TrimSpace(req.Code)
	name := strings.TrimSpace(req.Name)
	description := strings.TrimSpace(req.Description)

	existingItem, err := s.repo.GetByCodeOrName(ctx, code, name)
	if err != nil {
		return apperror.Internal("failed to check existing system type", err)
	}
	if existingItem != nil && existingItem.ID != systemTypeID {
		return apperror.Conflict("Tipo de sistema ja existe")
	}

	err = s.repo.Update(ctx, &model.SystemTypeModel{
		ID:          systemTypeID,
		Code:        code,
		Name:        name,
		Description: description,
		UpdatedAt:   time.Now(),
	})
	if err != nil {
		return mapSystemTypePersistenceError("update", err)
	}

	return nil
}

func (s *service) Delete(ctx context.Context, systemTypeID int64) error {
	if systemTypeID <= 0 {
		return apperror.BadRequest("system_type_id e obrigatorio")
	}

	item, err := s.repo.GetByID(ctx, systemTypeID)
	if err != nil {
		return apperror.Internal("failed to check system type", err)
	}
	if item == nil {
		return apperror.NotFound("Tipo de sistema nao encontrado")
	}

	if err := s.repo.Delete(ctx, systemTypeID); err != nil {
		return mapSystemTypePersistenceError("delete", err)
	}

	return nil
}

func toResponse(item model.SystemTypeModel) dto.SystemTypeResponse {
	return dto.SystemTypeResponse{
		ID:          item.ID,
		Code:        item.Code,
		Name:        item.Name,
		Description: item.Description,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}

func mapSystemTypePersistenceError(action string, err error) error {
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) {
		switch pgError.ConstraintName {
		case "system_types_code_key", "system_types_name_key":
			return apperror.Conflict("Tipo de sistema ja existe")
		case "fk_budgets_system_type_id":
			return apperror.BadRequest("Tipo de sistema esta vinculado a orcamentos")
		}

		switch pgError.Code {
		case "23503":
			return apperror.BadRequest("Tipo de sistema esta vinculado a orcamentos")
		case "23505":
			return apperror.Conflict("Tipo de sistema ja existe")
		}
	}

	return apperror.Internal("failed to "+action+" system type", err)
}
