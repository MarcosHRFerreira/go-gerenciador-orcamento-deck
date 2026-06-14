package project

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/apperror"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	projectrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/project"
	projecttyperepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/projecttype"
)

type Service interface {
	Create(ctx context.Context, req *dto.CreateProjectRequest) (int64, error)
	List(ctx context.Context) ([]dto.ProjectResponse, error)
	GetByID(ctx context.Context, projectID int64) (*dto.ProjectResponse, error)
	Update(ctx context.Context, projectID int64, req *dto.UpdateProjectRequest) error
	Delete(ctx context.Context, projectID int64) error
}

type service struct {
	repo            projectrepository.Repository
	projectTypeRepo projecttyperepository.Repository
}

func NewService(repo projectrepository.Repository, projectTypeRepo projecttyperepository.Repository) Service {
	return &service{
		repo:            repo,
		projectTypeRepo: projectTypeRepo,
	}
}

func (s *service) Create(ctx context.Context, req *dto.CreateProjectRequest) (int64, error) {
	projectTypeID, err := s.normalizeProjectTypeID(ctx, req.ProjectTypeID)
	if err != nil {
		return 0, err
	}

	now := time.Now()
	id, err := s.repo.Create(ctx, &model.ProjectModel{
		Name:          strings.TrimSpace(req.Name),
		ProjectTypeID: projectTypeID,
		City:          strings.TrimSpace(req.City),
		State:         strings.TrimSpace(req.State),
		Notes:         strings.TrimSpace(req.Notes),
		CreatedAt:     now,
		UpdatedAt:     now,
	})
	if err != nil {
		return 0, apperror.Internal("failed to create project", err)
	}

	return id, nil
}

func (s *service) List(ctx context.Context) ([]dto.ProjectResponse, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, apperror.Internal("failed to list projects", err)
	}

	response := make([]dto.ProjectResponse, 0, len(items))
	for _, item := range items {
		response = append(response, toResponse(item))
	}

	return response, nil
}

func (s *service) GetByID(ctx context.Context, projectID int64) (*dto.ProjectResponse, error) {
	if projectID <= 0 {
		return nil, apperror.BadRequest("project_id is required")
	}

	item, err := s.repo.GetByID(ctx, projectID)
	if err != nil {
		return nil, apperror.Internal("failed to get project", err)
	}
	if item == nil {
		return nil, apperror.NotFound("project not found")
	}

	response := toResponse(*item)
	return &response, nil
}

func (s *service) Update(ctx context.Context, projectID int64, req *dto.UpdateProjectRequest) error {
	if projectID <= 0 {
		return apperror.BadRequest("project_id is required")
	}

	currentItem, err := s.repo.GetByID(ctx, projectID)
	if err != nil {
		return apperror.Internal("failed to check project", err)
	}
	if currentItem == nil {
		return apperror.NotFound("project not found")
	}

	projectTypeID, err := s.normalizeProjectTypeID(ctx, req.ProjectTypeID)
	if err != nil {
		return err
	}

	err = s.repo.Update(ctx, &model.ProjectModel{
		ID:            projectID,
		Name:          strings.TrimSpace(req.Name),
		ProjectTypeID: projectTypeID,
		City:          strings.TrimSpace(req.City),
		State:         strings.TrimSpace(req.State),
		Notes:         strings.TrimSpace(req.Notes),
		UpdatedAt:     time.Now(),
	})
	if err != nil {
		return apperror.Internal("failed to update project", err)
	}

	return nil
}

func (s *service) Delete(ctx context.Context, projectID int64) error {
	if projectID <= 0 {
		return apperror.BadRequest("project_id is required")
	}

	item, err := s.repo.GetByID(ctx, projectID)
	if err != nil {
		return apperror.Internal("failed to check project", err)
	}
	if item == nil {
		return apperror.NotFound("project not found")
	}

	if err := s.repo.Delete(ctx, projectID); err != nil {
		return apperror.Internal("failed to delete project", err)
	}

	return nil
}

func (s *service) normalizeProjectTypeID(ctx context.Context, projectTypeID *int64) (sql.NullInt64, error) {
	if projectTypeID == nil {
		return sql.NullInt64{}, nil
	}

	if *projectTypeID <= 0 {
		return sql.NullInt64{}, apperror.BadRequest("project_type_id must be a valid integer")
	}

	projectType, err := s.projectTypeRepo.GetByID(ctx, *projectTypeID)
	if err != nil {
		return sql.NullInt64{}, apperror.Internal("failed to check project type", err)
	}
	if projectType == nil {
		return sql.NullInt64{}, apperror.BadRequest("project type not found")
	}

	return sql.NullInt64{
		Int64: *projectTypeID,
		Valid: true,
	}, nil
}

func toResponse(item model.ProjectModel) dto.ProjectResponse {
	var projectTypeID *int64
	if item.ProjectTypeID.Valid {
		projectTypeID = &item.ProjectTypeID.Int64
	}

	return dto.ProjectResponse{
		ID:            item.ID,
		Name:          item.Name,
		ProjectTypeID: projectTypeID,
		City:          item.City,
		State:         item.State,
		Notes:         item.Notes,
		CreatedAt:     item.CreatedAt,
		UpdatedAt:     item.UpdatedAt,
	}
}
