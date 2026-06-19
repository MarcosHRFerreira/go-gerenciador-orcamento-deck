package productline

import (
	"context"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/apperror"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	productlinerepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/productline"
)

type Service interface {
	List(ctx context.Context) ([]dto.ProductLineResponse, error)
}

type service struct {
	repo productlinerepository.Repository
}

func NewService(repo productlinerepository.Repository) Service {
	return &service{repo: repo}
}

func (s *service) List(ctx context.Context) ([]dto.ProductLineResponse, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, apperror.Internal("failed to list product lines", err)
	}

	response := make([]dto.ProductLineResponse, 0, len(items))
	for _, item := range items {
		response = append(response, toResponse(item))
	}

	return response, nil
}

func toResponse(item model.ProductLineModel) dto.ProductLineResponse {
	return dto.ProductLineResponse{
		ID:          item.ID,
		Code:        item.Code,
		Name:        item.Name,
		Description: item.Description,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}
