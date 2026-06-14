package salesperson

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/apperror"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	salespersonrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/salesperson"
	"github.com/jackc/pgx/v5/pgconn"
)

type Service interface {
	Create(ctx context.Context, req *dto.CreateSalespersonRequest) (int64, error)
	List(ctx context.Context) ([]dto.SalespersonResponse, error)
	GetByID(ctx context.Context, salespersonID int64) (*dto.SalespersonResponse, error)
	Update(ctx context.Context, salespersonID int64, req *dto.UpdateSalespersonRequest) error
	Delete(ctx context.Context, salespersonID int64) error
}

type service struct {
	repo salespersonrepository.Repository
}

func NewService(repo salespersonrepository.Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, req *dto.CreateSalespersonRequest) (int64, error) {
	email := strings.TrimSpace(req.Email)
	existingItem, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return 0, apperror.Internal("failed to check existing salesperson", err)
	}
	if existingItem != nil {
		return 0, apperror.Conflict("salesperson already exists")
	}

	now := time.Now()
	id, err := s.repo.Create(ctx, &model.SalespersonModel{
		Name:      strings.TrimSpace(req.Name),
		Email:     email,
		Phone:     strings.TrimSpace(req.Phone),
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return 0, mapSalespersonPersistenceError("create", err)
	}

	return id, nil
}

func (s *service) List(ctx context.Context) ([]dto.SalespersonResponse, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, apperror.Internal("failed to list salespeople", err)
	}

	response := make([]dto.SalespersonResponse, 0, len(items))
	for _, item := range items {
		response = append(response, toResponse(item))
	}

	return response, nil
}

func (s *service) GetByID(ctx context.Context, salespersonID int64) (*dto.SalespersonResponse, error) {
	if salespersonID <= 0 {
		return nil, apperror.BadRequest("salesperson_id is required")
	}

	item, err := s.repo.GetByID(ctx, salespersonID)
	if err != nil {
		return nil, apperror.Internal("failed to get salesperson", err)
	}
	if item == nil {
		return nil, apperror.NotFound("salesperson not found")
	}

	response := toResponse(*item)
	return &response, nil
}

func (s *service) Update(ctx context.Context, salespersonID int64, req *dto.UpdateSalespersonRequest) error {
	if salespersonID <= 0 {
		return apperror.BadRequest("salesperson_id is required")
	}

	currentItem, err := s.repo.GetByID(ctx, salespersonID)
	if err != nil {
		return apperror.Internal("failed to check salesperson", err)
	}
	if currentItem == nil {
		return apperror.NotFound("salesperson not found")
	}

	email := strings.TrimSpace(req.Email)
	existingItem, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return apperror.Internal("failed to check existing salesperson", err)
	}
	if existingItem != nil && existingItem.ID != salespersonID {
		return apperror.Conflict("salesperson already exists")
	}

	err = s.repo.Update(ctx, &model.SalespersonModel{
		ID:        salespersonID,
		Name:      strings.TrimSpace(req.Name),
		Email:     email,
		Phone:     strings.TrimSpace(req.Phone),
		Active:    req.Active,
		UpdatedAt: time.Now(),
	})
	if err != nil {
		return mapSalespersonPersistenceError("update", err)
	}

	return nil
}

func (s *service) Delete(ctx context.Context, salespersonID int64) error {
	if salespersonID <= 0 {
		return apperror.BadRequest("salesperson_id is required")
	}

	item, err := s.repo.GetByID(ctx, salespersonID)
	if err != nil {
		return apperror.Internal("failed to check salesperson", err)
	}
	if item == nil {
		return apperror.NotFound("salesperson not found")
	}

	if err := s.repo.Delete(ctx, salespersonID); err != nil {
		return mapSalespersonPersistenceError("delete", err)
	}

	return nil
}

func toResponse(item model.SalespersonModel) dto.SalespersonResponse {
	return dto.SalespersonResponse{
		ID:        item.ID,
		Name:      item.Name,
		Email:     item.Email,
		Phone:     item.Phone,
		Active:    item.Active,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}

func mapSalespersonPersistenceError(action string, err error) error {
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) {
		switch pgError.ConstraintName {
		case "salespeople_email_key":
			return apperror.Conflict("salesperson already exists")
		}

		if pgError.Code == "23505" {
			return apperror.Conflict("salesperson already exists")
		}
	}

	return apperror.Internal("failed to "+action+" salesperson", err)
}
