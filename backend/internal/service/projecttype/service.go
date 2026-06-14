package projecttype

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/apperror"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	projecttyperepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/projecttype"
	"github.com/jackc/pgx/v5/pgconn"
)

type Service interface {
	Create(ctx context.Context, req *dto.CreateProjectTypeRequest) (int64, error)
	List(ctx context.Context) ([]dto.ProjectTypeResponse, error)
	GetByID(ctx context.Context, projectTypeID int64) (*dto.ProjectTypeResponse, error)
	Update(ctx context.Context, projectTypeID int64, req *dto.UpdateProjectTypeRequest) error
	Delete(ctx context.Context, projectTypeID int64) error
}

type service struct {
	repo projecttyperepository.Repository
}

func NewService(repo projecttyperepository.Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, req *dto.CreateProjectTypeRequest) (int64, error) {
	code := strings.TrimSpace(req.Code)
	name := strings.TrimSpace(req.Name)

	existingItem, err := s.repo.GetByCodeOrName(ctx, code, name)
	if err != nil {
		return 0, apperror.Internal("failed to check existing project type", err)
	}
	if existingItem != nil {
		return 0, apperror.Conflict("Tipo de projeto ja existe")
	}

	now := time.Now()
	id, err := s.repo.Create(ctx, &model.ProjectTypeModel{
		Code:        code,
		Name:        name,
		Description: strings.TrimSpace(req.Description),
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		return 0, mapProjectTypePersistenceError("create", err)
	}

	return id, nil
}

func (s *service) List(ctx context.Context) ([]dto.ProjectTypeResponse, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, apperror.Internal("failed to list project types", err)
	}

	response := make([]dto.ProjectTypeResponse, 0, len(items))
	for _, item := range items {
		response = append(response, toResponse(item))
	}

	return response, nil
}

func (s *service) GetByID(ctx context.Context, projectTypeID int64) (*dto.ProjectTypeResponse, error) {
	if projectTypeID <= 0 {
		return nil, apperror.BadRequest("project_type_id e obrigatorio")
	}

	item, err := s.repo.GetByID(ctx, projectTypeID)
	if err != nil {
		return nil, apperror.Internal("failed to get project type", err)
	}
	if item == nil {
		return nil, apperror.NotFound("Tipo de projeto nao encontrado")
	}

	response := toResponse(*item)
	return &response, nil
}

func (s *service) Update(ctx context.Context, projectTypeID int64, req *dto.UpdateProjectTypeRequest) error {
	if projectTypeID <= 0 {
		return apperror.BadRequest("project_type_id e obrigatorio")
	}

	currentItem, err := s.repo.GetByID(ctx, projectTypeID)
	if err != nil {
		return apperror.Internal("failed to check project type", err)
	}
	if currentItem == nil {
		return apperror.NotFound("Tipo de projeto nao encontrado")
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
		return apperror.Internal("failed to check existing project type", err)
	}
	if existingItem != nil && existingItem.ID != projectTypeID {
		return apperror.Conflict("Tipo de projeto ja existe")
	}

	err = s.repo.Update(ctx, &model.ProjectTypeModel{
		ID:          projectTypeID,
		Code:        code,
		Name:        name,
		Description: strings.TrimSpace(req.Description),
		UpdatedAt:   time.Now(),
	})
	if err != nil {
		return mapProjectTypePersistenceError("update", err)
	}

	return nil
}

func (s *service) Delete(ctx context.Context, projectTypeID int64) error {
	if projectTypeID <= 0 {
		return apperror.BadRequest("project_type_id e obrigatorio")
	}

	item, err := s.repo.GetByID(ctx, projectTypeID)
	if err != nil {
		return apperror.Internal("failed to check project type", err)
	}
	if item == nil {
		return apperror.NotFound("Tipo de projeto nao encontrado")
	}

	if err := s.repo.Delete(ctx, projectTypeID); err != nil {
		return mapProjectTypePersistenceError("delete", err)
	}

	return nil
}

func toResponse(item model.ProjectTypeModel) dto.ProjectTypeResponse {
	return dto.ProjectTypeResponse{
		ID:          item.ID,
		Code:        item.Code,
		Name:        item.Name,
		Description: item.Description,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}

func mapProjectTypePersistenceError(action string, err error) error {
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) {
		switch pgError.ConstraintName {
		case "project_types_code_key", "project_types_name_key":
			return apperror.Conflict("Tipo de projeto ja existe")
		}

		if pgError.Code == "23505" {
			return apperror.Conflict("Tipo de projeto ja existe")
		}
	}

	return apperror.Internal("failed to "+action+" project type", err)
}
