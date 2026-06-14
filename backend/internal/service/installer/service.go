package installer

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/apperror"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	installerrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/installer"
	"github.com/jackc/pgx/v5/pgconn"
)

type Service interface {
	Create(ctx context.Context, req *dto.CreateInstallerRequest) (int64, error)
	List(ctx context.Context) ([]dto.InstallerResponse, error)
	GetByID(ctx context.Context, installerID int64) (*dto.InstallerResponse, error)
	Update(ctx context.Context, installerID int64, req *dto.UpdateInstallerRequest) error
	Delete(ctx context.Context, installerID int64) error
}

type service struct {
	repo installerrepository.Repository
}

func NewService(repo installerrepository.Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, req *dto.CreateInstallerRequest) (int64, error) {
	document := strings.TrimSpace(req.Document)
	if document != "" {
		existingItem, err := s.repo.GetByDocument(ctx, document)
		if err != nil {
			return 0, apperror.Internal("failed to check existing installer", err)
		}
		if existingItem != nil {
			return 0, apperror.Conflict("Instalador ja existe")
		}
	}

	now := time.Now()
	id, err := s.repo.Create(ctx, &model.InstallerModel{
		Name:      strings.TrimSpace(req.Name),
		Document:  document,
		Email:     strings.TrimSpace(req.Email),
		Phone:     strings.TrimSpace(req.Phone),
		City:      strings.TrimSpace(req.City),
		State:     strings.TrimSpace(req.State),
		Notes:     strings.TrimSpace(req.Notes),
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return 0, mapInstallerPersistenceError("create", err)
	}

	return id, nil
}

func (s *service) List(ctx context.Context) ([]dto.InstallerResponse, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, apperror.Internal("failed to list installers", err)
	}

	response := make([]dto.InstallerResponse, 0, len(items))
	for _, item := range items {
		response = append(response, toResponse(item))
	}

	return response, nil
}

func (s *service) GetByID(ctx context.Context, installerID int64) (*dto.InstallerResponse, error) {
	if installerID <= 0 {
		return nil, apperror.BadRequest("installer_id e obrigatorio")
	}

	item, err := s.repo.GetByID(ctx, installerID)
	if err != nil {
		return nil, apperror.Internal("failed to get installer", err)
	}
	if item == nil {
		return nil, apperror.NotFound("Instalador nao encontrado")
	}

	response := toResponse(*item)
	return &response, nil
}

func (s *service) Update(ctx context.Context, installerID int64, req *dto.UpdateInstallerRequest) error {
	if installerID <= 0 {
		return apperror.BadRequest("installer_id e obrigatorio")
	}

	currentItem, err := s.repo.GetByID(ctx, installerID)
	if err != nil {
		return apperror.Internal("failed to check installer", err)
	}
	if currentItem == nil {
		return apperror.NotFound("Instalador nao encontrado")
	}

	document := strings.TrimSpace(req.Document)
	if document != "" {
		existingItem, existsErr := s.repo.GetByDocument(ctx, document)
		if existsErr != nil {
			return apperror.Internal("failed to check existing installer", existsErr)
		}
		if existingItem != nil && existingItem.ID != installerID {
			return apperror.Conflict("Instalador ja existe")
		}
	}

	err = s.repo.Update(ctx, &model.InstallerModel{
		ID:        installerID,
		Name:      strings.TrimSpace(req.Name),
		Document:  document,
		Email:     strings.TrimSpace(req.Email),
		Phone:     strings.TrimSpace(req.Phone),
		City:      strings.TrimSpace(req.City),
		State:     strings.TrimSpace(req.State),
		Notes:     strings.TrimSpace(req.Notes),
		Active:    req.Active,
		UpdatedAt: time.Now(),
	})
	if err != nil {
		return mapInstallerPersistenceError("update", err)
	}

	return nil
}

func (s *service) Delete(ctx context.Context, installerID int64) error {
	if installerID <= 0 {
		return apperror.BadRequest("installer_id e obrigatorio")
	}

	item, err := s.repo.GetByID(ctx, installerID)
	if err != nil {
		return apperror.Internal("failed to check installer", err)
	}
	if item == nil {
		return apperror.NotFound("Instalador nao encontrado")
	}

	if err := s.repo.Delete(ctx, installerID); err != nil {
		return mapInstallerPersistenceError("delete", err)
	}

	return nil
}

func toResponse(item model.InstallerModel) dto.InstallerResponse {
	return dto.InstallerResponse{
		ID:        item.ID,
		Name:      item.Name,
		Document:  item.Document,
		Email:     item.Email,
		Phone:     item.Phone,
		City:      item.City,
		State:     item.State,
		Notes:     item.Notes,
		Active:    item.Active,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}

func mapInstallerPersistenceError(action string, err error) error {
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) {
		switch pgError.ConstraintName {
		case "installers_document_key":
			return apperror.Conflict("Instalador ja existe")
		}

		if pgError.Code == "23505" {
			return apperror.Conflict("Instalador ja existe")
		}
	}

	return apperror.Internal("failed to "+action+" installer", err)
}
