package estimator

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/apperror"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	estimatorrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/estimator"
	userrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/user"
	"github.com/jackc/pgx/v5/pgconn"
)

type Service interface {
	Create(ctx context.Context, req *dto.CreateEstimatorRequest) (int64, error)
	GetNextCode(ctx context.Context) (string, error)
	List(ctx context.Context) ([]dto.EstimatorResponse, error)
	GetByID(ctx context.Context, estimatorID int64) (*dto.EstimatorResponse, error)
	Update(ctx context.Context, estimatorID int64, req *dto.UpdateEstimatorRequest) error
	Delete(ctx context.Context, estimatorID int64) error
}

type service struct {
	repo     estimatorrepository.Repository
	userRepo userrepository.Repository
}

func NewService(repo estimatorrepository.Repository, userRepo userrepository.Repository) Service {
	return &service{
		repo:     repo,
		userRepo: userRepo,
	}
}

func (s *service) Create(ctx context.Context, req *dto.CreateEstimatorRequest) (int64, error) {
	code := strings.TrimSpace(req.Code)
	var err error
	if code == "" {
		code, err = s.repo.GetNextCode(ctx)
		if err != nil {
			return 0, apperror.Internal("failed to generate estimator code", err)
		}
	}

	userID, err := s.normalizeUserID(ctx, req.UserID, 0)
	if err != nil {
		return 0, err
	}

	existingItem, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return 0, apperror.Internal("failed to check existing estimator", err)
	}
	if existingItem != nil {
		return 0, apperror.Conflict("Orcamentista ja existe")
	}

	if userID.Valid {
		existingByUserID, existingByUserIDErr := s.repo.GetByUserID(ctx, userID.Int64)
		if existingByUserIDErr != nil {
			return 0, apperror.Internal("failed to check linked user", existingByUserIDErr)
		}
		if existingByUserID != nil {
			return 0, apperror.Conflict("Usuario ja vinculado a outro orcamentista")
		}
	}

	now := time.Now()
	id, err := s.repo.Create(ctx, &model.EstimatorModel{
		Code:      code,
		Name:      strings.TrimSpace(req.Name),
		Email:     strings.TrimSpace(req.Email),
		Phone:     strings.TrimSpace(req.Phone),
		Active:    true,
		Notes:     strings.TrimSpace(req.Notes),
		UserID:    userID,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return 0, mapEstimatorPersistenceError("create", err)
	}

	return id, nil
}

func (s *service) GetNextCode(ctx context.Context) (string, error) {
	code, err := s.repo.GetNextCode(ctx)
	if err != nil {
		return "", apperror.Internal("failed to generate estimator code", err)
	}

	return code, nil
}

func (s *service) List(ctx context.Context) ([]dto.EstimatorResponse, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, apperror.Internal("failed to list estimators", err)
	}

	response := make([]dto.EstimatorResponse, 0, len(items))
	for _, item := range items {
		response = append(response, toResponse(item))
	}

	return response, nil
}

func (s *service) GetByID(ctx context.Context, estimatorID int64) (*dto.EstimatorResponse, error) {
	if estimatorID <= 0 {
		return nil, apperror.BadRequest("estimator_id e obrigatorio")
	}

	item, err := s.repo.GetByID(ctx, estimatorID)
	if err != nil {
		return nil, apperror.Internal("failed to get estimator", err)
	}
	if item == nil {
		return nil, apperror.NotFound("Orcamentista nao encontrado")
	}

	response := toResponse(*item)
	return &response, nil
}

func (s *service) Update(ctx context.Context, estimatorID int64, req *dto.UpdateEstimatorRequest) error {
	if estimatorID <= 0 {
		return apperror.BadRequest("estimator_id e obrigatorio")
	}

	currentItem, err := s.repo.GetByID(ctx, estimatorID)
	if err != nil {
		return apperror.Internal("failed to check estimator", err)
	}
	if currentItem == nil {
		return apperror.NotFound("Orcamentista nao encontrado")
	}

	code := strings.TrimSpace(req.Code)
	existingItem, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return apperror.Internal("failed to check existing estimator", err)
	}
	if existingItem != nil && existingItem.ID != estimatorID {
		return apperror.Conflict("Orcamentista ja existe")
	}

	userID, err := s.normalizeUserID(ctx, req.UserID, estimatorID)
	if err != nil {
		return err
	}

	if currentItem.Code == code &&
		currentItem.Name == strings.TrimSpace(req.Name) &&
		currentItem.Email == strings.TrimSpace(req.Email) &&
		currentItem.Phone == strings.TrimSpace(req.Phone) &&
		currentItem.Active == req.Active &&
		currentItem.Notes == strings.TrimSpace(req.Notes) &&
		currentItem.UserID == userID {
		return nil
	}

	err = s.repo.Update(ctx, &model.EstimatorModel{
		ID:        estimatorID,
		Code:      code,
		Name:      strings.TrimSpace(req.Name),
		Email:     strings.TrimSpace(req.Email),
		Phone:     strings.TrimSpace(req.Phone),
		Active:    req.Active,
		Notes:     strings.TrimSpace(req.Notes),
		UserID:    userID,
		UpdatedAt: time.Now(),
	})
	if err != nil {
		return mapEstimatorPersistenceError("update", err)
	}

	return nil
}

func (s *service) Delete(ctx context.Context, estimatorID int64) error {
	if estimatorID <= 0 {
		return apperror.BadRequest("estimator_id e obrigatorio")
	}

	item, err := s.repo.GetByID(ctx, estimatorID)
	if err != nil {
		return apperror.Internal("failed to check estimator", err)
	}
	if item == nil {
		return apperror.NotFound("Orcamentista nao encontrado")
	}

	if err := s.repo.Delete(ctx, estimatorID); err != nil {
		return mapEstimatorPersistenceError("delete", err)
	}

	return nil
}

func (s *service) normalizeUserID(ctx context.Context, userID *int64, currentEstimatorID int64) (sql.NullInt64, error) {
	if userID == nil {
		return sql.NullInt64{}, nil
	}

	if *userID <= 0 {
		return sql.NullInt64{}, apperror.BadRequest("user_id deve ser um inteiro valido")
	}

	user, err := s.userRepo.GetUserByID(ctx, *userID)
	if err != nil {
		return sql.NullInt64{}, apperror.Internal("failed to check linked user", err)
	}
	if user == nil {
		return sql.NullInt64{}, apperror.BadRequest("Usuario informado nao foi encontrado")
	}
	if user.Role != model.RoleUser {
		return sql.NullInt64{}, apperror.BadRequest("Somente usuarios com perfil user podem ser vinculados ao orcamentista")
	}
	if user.UserKind != model.UserKindEstimator {
		return sql.NullInt64{}, apperror.BadRequest("Usuario informado precisa ter user_kind estimator")
	}

	existingByUserID, err := s.repo.GetByUserID(ctx, *userID)
	if err != nil {
		return sql.NullInt64{}, apperror.Internal("failed to check linked user", err)
	}
	if existingByUserID != nil && existingByUserID.ID != currentEstimatorID {
		return sql.NullInt64{}, apperror.Conflict("Usuario ja vinculado a outro orcamentista")
	}

	return sql.NullInt64{
		Int64: *userID,
		Valid: true,
	}, nil
}

func toResponse(item model.EstimatorModel) dto.EstimatorResponse {
	var userID *int64
	if item.UserID.Valid {
		userID = &item.UserID.Int64
	}

	var userName *string
	if item.UserName.Valid {
		userName = &item.UserName.String
	}

	return dto.EstimatorResponse{
		ID:        item.ID,
		Code:      item.Code,
		Name:      item.Name,
		Email:     item.Email,
		Phone:     item.Phone,
		Active:    item.Active,
		Notes:     item.Notes,
		UserID:    userID,
		UserName:  userName,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}

func mapEstimatorPersistenceError(action string, err error) error {
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) {
		switch pgError.ConstraintName {
		case "uq_estimators_code":
			return apperror.Conflict("Orcamentista ja existe")
		case "uq_estimators_user_id":
			return apperror.Conflict("Usuario ja vinculado a outro orcamentista")
		case "fk_estimators_user_id":
			return apperror.BadRequest("Usuario informado nao foi encontrado")
		}

		if pgError.Code == "23505" {
			return apperror.Conflict("Orcamentista ja existe")
		}
	}

	return apperror.Internal("failed to "+action+" estimator", err)
}
