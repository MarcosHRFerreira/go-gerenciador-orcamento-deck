package lossreason

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/apperror"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	lossreasonrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/lossreason"
	"github.com/jackc/pgx/v5/pgconn"
)

type Service interface {
	Create(ctx context.Context, req *dto.CreateLossReasonRequest) (int64, error)
	List(ctx context.Context) ([]dto.LossReasonResponse, error)
	GetByID(ctx context.Context, reasonID int64) (*dto.LossReasonResponse, error)
	Update(ctx context.Context, reasonID int64, req *dto.UpdateLossReasonRequest) error
	Delete(ctx context.Context, reasonID int64) error
}

type service struct {
	repo lossreasonrepository.Repository
}

func NewService(repo lossreasonrepository.Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, req *dto.CreateLossReasonRequest) (int64, error) {
	code := strings.TrimSpace(req.Code)
	name := strings.TrimSpace(req.Name)

	existingItem, err := s.repo.GetByCodeOrName(ctx, code, name)
	if err != nil {
		return 0, apperror.Internal("failed to check existing loss reason", err)
	}
	if existingItem != nil {
		return 0, apperror.Conflict("loss reason already exists")
	}

	now := time.Now()
	id, err := s.repo.Create(ctx, &model.LossReasonModel{
		Code:        code,
		Name:        name,
		Description: strings.TrimSpace(req.Description),
		Active:      req.Active,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		return 0, mapLossReasonPersistenceError("create", err)
	}

	return id, nil
}

func (s *service) List(ctx context.Context) ([]dto.LossReasonResponse, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, apperror.Internal("failed to list loss reasons", err)
	}

	response := make([]dto.LossReasonResponse, 0, len(items))
	for _, item := range items {
		response = append(response, toResponse(item))
	}

	return response, nil
}

func (s *service) GetByID(ctx context.Context, reasonID int64) (*dto.LossReasonResponse, error) {
	if reasonID <= 0 {
		return nil, apperror.BadRequest("reason_id is required")
	}

	item, err := s.repo.GetByID(ctx, reasonID)
	if err != nil {
		return nil, apperror.Internal("failed to get loss reason", err)
	}
	if item == nil {
		return nil, apperror.NotFound("loss reason not found")
	}

	response := toResponse(*item)
	return &response, nil
}

func (s *service) Update(ctx context.Context, reasonID int64, req *dto.UpdateLossReasonRequest) error {
	if reasonID <= 0 {
		return apperror.BadRequest("reason_id is required")
	}

	currentItem, err := s.repo.GetByID(ctx, reasonID)
	if err != nil {
		return apperror.Internal("failed to check loss reason", err)
	}
	if currentItem == nil {
		return apperror.NotFound("loss reason not found")
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
		return apperror.Internal("failed to check existing loss reason", err)
	}
	if existingItem != nil && existingItem.ID != reasonID {
		return apperror.Conflict("loss reason already exists")
	}

	err = s.repo.Update(ctx, &model.LossReasonModel{
		ID:          reasonID,
		Code:        code,
		Name:        name,
		Description: strings.TrimSpace(req.Description),
		Active:      req.Active,
		UpdatedAt:   time.Now(),
	})
	if err != nil {
		return mapLossReasonPersistenceError("update", err)
	}

	return nil
}

func (s *service) Delete(ctx context.Context, reasonID int64) error {
	if reasonID <= 0 {
		return apperror.BadRequest("reason_id is required")
	}

	item, err := s.repo.GetByID(ctx, reasonID)
	if err != nil {
		return apperror.Internal("failed to check loss reason", err)
	}
	if item == nil {
		return apperror.NotFound("loss reason not found")
	}

	if err := s.repo.Delete(ctx, reasonID); err != nil {
		return mapLossReasonPersistenceError("delete", err)
	}

	return nil
}

func toResponse(item model.LossReasonModel) dto.LossReasonResponse {
	return dto.LossReasonResponse{
		ID:          item.ID,
		Code:        item.Code,
		Name:        item.Name,
		Description: item.Description,
		Active:      item.Active,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}

func mapLossReasonPersistenceError(action string, err error) error {
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) {
		switch pgError.ConstraintName {
		case "loss_reasons_code_key", "loss_reasons_name_key":
			return apperror.Conflict("loss reason already exists")
		}

		if pgError.Code == "23505" {
			return apperror.Conflict("loss reason already exists")
		}
	}

	return apperror.Internal("failed to "+action+" loss reason", err)
}
