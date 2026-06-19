package accessscope

import (
	"context"
	"errors"
	"strings"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/apperror"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	estimatorrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/estimator"
	salespersonrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/salesperson"
	userrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/user"
)

type BudgetScope struct {
	UserKind                model.UserKind
	RestrictedSalespersonID *int64
	RestrictedEstimatorID   *int64
}

func ResolveBudgetScope(
	ctx context.Context,
	role model.UserRole,
	username string,
	userRepo userrepository.Repository,
	salespersonRepo salespersonrepository.Repository,
	estimatorRepo estimatorrepository.Repository,
) (*BudgetScope, error) {
	if role == model.RoleAdmin {
		return &BudgetScope{}, nil
	}

	normalizedUsername := strings.TrimSpace(username)
	if normalizedUsername == "" {
		return zeroBudgetScope(model.UserKindSalesperson), nil
	}

	if userRepo == nil {
		return nil, apperror.Internal("failed to resolve budget scope", errors.New("user repository is nil"))
	}

	user, err := userRepo.GetUserByUsername(ctx, normalizedUsername)
	if err != nil {
		return nil, apperror.Internal("failed to resolve budget scope", err)
	}
	if user == nil || !user.Active {
		return zeroBudgetScope(model.UserKindSalesperson), nil
	}

	userKind := normalizeOperationalUserKind(user.UserKind)
	if userKind == model.UserKindEstimator {
		return &BudgetScope{
			UserKind: model.UserKindEstimator,
		}, nil
	}

	if salespersonRepo == nil {
		return nil, apperror.Internal("failed to resolve budget scope", errors.New("salesperson repository is nil"))
	}

	salesperson, salespersonErr := salespersonRepo.GetByUsername(ctx, normalizedUsername)
	if salespersonErr != nil {
		return nil, apperror.Internal("failed to resolve budget scope", salespersonErr)
	}
	if salesperson == nil || !salesperson.Active {
		return zeroBudgetScope(model.UserKindSalesperson), nil
	}

	return &BudgetScope{
		UserKind:                model.UserKindSalesperson,
		RestrictedSalespersonID: &salesperson.ID,
	}, nil
}

func zeroBudgetScope(userKind model.UserKind) *BudgetScope {
	zeroValue := zeroRestrictedID()
	scope := &BudgetScope{
		UserKind: userKind,
	}
	if userKind == model.UserKindEstimator {
		scope.RestrictedEstimatorID = zeroValue
		return scope
	}

	scope.RestrictedSalespersonID = zeroValue
	return scope
}

func zeroRestrictedID() *int64 {
	value := int64(0)
	return &value
}

func normalizeOperationalUserKind(userKind model.UserKind) model.UserKind {
	if userKind == model.UserKindEstimator {
		return model.UserKindEstimator
	}

	return model.UserKindSalesperson
}
