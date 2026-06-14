package accessscope

import (
	"context"
	"errors"
	"strings"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/apperror"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	salespersonrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/salesperson"
)

func ResolveRestrictedSalespersonID(
	ctx context.Context,
	role model.UserRole,
	username string,
	salespersonRepo salespersonrepository.Repository,
) (*int64, error) {
	if role == model.RoleAdmin {
		return nil, nil
	}

	if salespersonRepo == nil {
		return nil, apperror.Internal("failed to resolve salesperson scope", errors.New("salesperson repository is nil"))
	}

	normalizedUsername := strings.TrimSpace(username)
	if normalizedUsername == "" {
		return zeroSalespersonID(), nil
	}

	salesperson, err := salespersonRepo.GetByUsername(ctx, normalizedUsername)
	if err != nil {
		return nil, apperror.Internal("failed to resolve salesperson scope", err)
	}
	if salesperson == nil || !salesperson.Active {
		return zeroSalespersonID(), nil
	}

	return &salesperson.ID, nil
}

func zeroSalespersonID() *int64 {
	value := int64(0)
	return &value
}
