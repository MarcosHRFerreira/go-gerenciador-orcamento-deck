package contact

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/apperror"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	contactrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/contact"
	installerrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/installer"
	"github.com/jackc/pgx/v5/pgconn"
)

type Service interface {
	Create(ctx context.Context, req *dto.CreateContactRequest) (int64, error)
	List(ctx context.Context, installerIDRaw string) ([]dto.ContactResponse, error)
	GetByID(ctx context.Context, contactID int64) (*dto.ContactResponse, error)
	Update(ctx context.Context, contactID int64, req *dto.UpdateContactRequest) error
	Delete(ctx context.Context, contactID int64) error
}

type service struct {
	repo          contactrepository.Repository
	installerRepo installerrepository.Repository
}

func NewService(repo contactrepository.Repository, installerRepo installerrepository.Repository) Service {
	return &service{
		repo:          repo,
		installerRepo: installerRepo,
	}
}

func (s *service) Create(ctx context.Context, req *dto.CreateContactRequest) (int64, error) {
	if err := s.validateInstallerExists(ctx, req.InstallerID); err != nil {
		return 0, err
	}
	if err := s.validateUniqueFields(ctx, req.InstallerID, strings.TrimSpace(req.Email), strings.TrimSpace(req.Phone), 0); err != nil {
		return 0, err
	}

	now := time.Now()
	id, err := s.repo.Create(ctx, &model.ContactModel{
		InstallerID: req.InstallerID,
		Name:        strings.TrimSpace(req.Name),
		Email:       strings.TrimSpace(req.Email),
		Phone:       strings.TrimSpace(req.Phone),
		Role:        strings.TrimSpace(req.Role),
		IsPrimary:   req.IsPrimary,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		return 0, mapContactPersistenceError("create", err)
	}

	return id, nil
}

func (s *service) List(ctx context.Context, installerIDRaw string) ([]dto.ContactResponse, error) {
	var installerID *int64
	if installerIDRaw != "" {
		parsedValue, err := strconv.ParseInt(installerIDRaw, 10, 64)
		if err != nil {
			return nil, apperror.BadRequest("installer_id deve ser um inteiro valido")
		}
		if parsedValue <= 0 {
			return nil, apperror.BadRequest("installer_id deve ser um inteiro valido")
		}

		installerID = &parsedValue
	}

	items, err := s.repo.List(ctx, installerID)
	if err != nil {
		return nil, apperror.Internal("failed to list contacts", err)
	}

	response := make([]dto.ContactResponse, 0, len(items))
	for _, item := range items {
		response = append(response, toResponse(item))
	}

	return response, nil
}

func (s *service) GetByID(ctx context.Context, contactID int64) (*dto.ContactResponse, error) {
	if contactID <= 0 {
		return nil, apperror.BadRequest("contact_id e obrigatorio")
	}

	item, err := s.repo.GetByID(ctx, contactID)
	if err != nil {
		return nil, apperror.Internal("failed to get contact", err)
	}
	if item == nil {
		return nil, apperror.NotFound("Contato nao encontrado")
	}

	response := toResponse(*item)
	return &response, nil
}

func (s *service) Update(ctx context.Context, contactID int64, req *dto.UpdateContactRequest) error {
	if contactID <= 0 {
		return apperror.BadRequest("contact_id e obrigatorio")
	}

	currentItem, err := s.repo.GetByID(ctx, contactID)
	if err != nil {
		return apperror.Internal("failed to check contact", err)
	}
	if currentItem == nil {
		return apperror.NotFound("Contato nao encontrado")
	}

	if err := s.validateInstallerExists(ctx, req.InstallerID); err != nil {
		return err
	}

	email := strings.TrimSpace(req.Email)
	phone := strings.TrimSpace(req.Phone)
	if err := s.validateUniqueFields(ctx, req.InstallerID, email, phone, contactID); err != nil {
		return err
	}

	err = s.repo.Update(ctx, &model.ContactModel{
		ID:          contactID,
		InstallerID: req.InstallerID,
		Name:        strings.TrimSpace(req.Name),
		Email:       email,
		Phone:       phone,
		Role:        strings.TrimSpace(req.Role),
		IsPrimary:   req.IsPrimary,
		UpdatedAt:   time.Now(),
	})
	if err != nil {
		return mapContactPersistenceError("update", err)
	}

	return nil
}

func (s *service) Delete(ctx context.Context, contactID int64) error {
	if contactID <= 0 {
		return apperror.BadRequest("contact_id e obrigatorio")
	}

	item, err := s.repo.GetByID(ctx, contactID)
	if err != nil {
		return apperror.Internal("failed to check contact", err)
	}
	if item == nil {
		return apperror.NotFound("Contato nao encontrado")
	}

	if err := s.repo.Delete(ctx, contactID); err != nil {
		return mapContactPersistenceError("delete", err)
	}

	return nil
}

func (s *service) validateInstallerExists(ctx context.Context, installerID int64) error {
	if installerID <= 0 {
		return apperror.BadRequest("installer_id e obrigatorio")
	}

	installer, err := s.installerRepo.GetByID(ctx, installerID)
	if err != nil {
		return apperror.Internal("failed to check installer", err)
	}
	if installer == nil {
		return apperror.BadRequest("Instalador nao encontrado")
	}

	return nil
}

func (s *service) validateUniqueFields(ctx context.Context, installerID int64, email string, phone string, contactID int64) error {
	existingByEmail, err := s.repo.GetByInstallerAndEmail(ctx, installerID, email)
	if err != nil {
		return apperror.Internal("failed to check existing contact", err)
	}
	if existingByEmail != nil && existingByEmail.ID != contactID {
		return apperror.Conflict("Contato ja existe")
	}

	existingByPhone, err := s.repo.GetByInstallerAndPhone(ctx, installerID, phone)
	if err != nil {
		return apperror.Internal("failed to check existing contact", err)
	}
	if existingByPhone != nil && existingByPhone.ID != contactID {
		return apperror.Conflict("Contato ja existe")
	}

	return nil
}

func toResponse(item model.ContactModel) dto.ContactResponse {
	return dto.ContactResponse{
		ID:          item.ID,
		InstallerID: item.InstallerID,
		Name:        item.Name,
		Email:       item.Email,
		Phone:       item.Phone,
		Role:        item.Role,
		IsPrimary:   item.IsPrimary,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}

func mapContactPersistenceError(action string, err error) error {
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) {
		switch pgError.ConstraintName {
		case "uq_contacts_installer_email", "uq_contacts_installer_phone":
			return apperror.Conflict("Contato ja existe")
		}

		if pgError.Code == "23505" {
			return apperror.Conflict("Contato ja existe")
		}
	}

	return apperror.Internal("failed to "+action+" contact", err)
}
