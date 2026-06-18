package dashboard

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/accessscope"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/apperror"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	dashboardrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/dashboard"
	salespersonrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/salesperson"
)

type Service interface {
	GetSalespeopleDashboard(
		ctx context.Context,
		filters *dto.DashboardSalespeopleFilters,
		role model.UserRole,
		username string,
	) (*dto.SalespeopleDashboardResponse, error)
}

type service struct {
	repo            dashboardrepository.Repository
	salespersonRepo salespersonrepository.Repository
	now             func() time.Time
}

func NewService(
	repo dashboardrepository.Repository,
	salespersonRepo salespersonrepository.Repository,
) Service {
	return NewServiceWithClock(repo, salespersonRepo, time.Now)
}

func NewServiceWithClock(
	repo dashboardrepository.Repository,
	salespersonRepo salespersonrepository.Repository,
	now func() time.Time,
) Service {
	return &service{
		repo:            repo,
		salespersonRepo: salespersonRepo,
		now:             now,
	}
}

func (s *service) GetSalespeopleDashboard(
	ctx context.Context,
	filters *dto.DashboardSalespeopleFilters,
	role model.UserRole,
	username string,
) (*dto.SalespeopleDashboardResponse, error) {
	normalizedFilters, err := normalizeDashboardFilters(filters)
	if err != nil {
		return nil, err
	}

	restrictedSalespersonID, err := accessscope.ResolveRestrictedSalespersonID(
		ctx,
		role,
		username,
		s.salespersonRepo,
	)
	if err != nil {
		return nil, err
	}
	normalizedFilters.RestrictedSalespersonID = restrictedSalespersonID

	summary, err := s.repo.GetSummary(ctx, normalizedFilters)
	if err != nil {
		return nil, apperror.Internal("failed to load dashboard summary", err)
	}

	availableYears, err := s.repo.ListAvailableYears(ctx, normalizedFilters)
	if err != nil {
		return nil, apperror.Internal("failed to load dashboard years", err)
	}

	salespersonSummaries, err := s.repo.ListSalespersonSummaries(ctx, normalizedFilters)
	if err != nil {
		return nil, apperror.Internal("failed to load dashboard salespeople", err)
	}

	staleBudgets, err := s.repo.ListStaleBudgets(ctx, normalizedFilters, 10)
	if err != nil {
		return nil, apperror.Internal("failed to load dashboard stale budgets", err)
	}

	monthlyEvolution, err := s.repo.ListMonthlyEvolution(ctx, normalizedFilters, 12)
	if err != nil {
		return nil, apperror.Internal("failed to load dashboard monthly evolution", err)
	}

	return buildSalespeopleDashboardResponse(
		summary,
		availableYears,
		salespersonSummaries,
		staleBudgets,
		monthlyEvolution,
	), nil
}

func normalizeDashboardFilters(
	filters *dto.DashboardSalespeopleFilters,
) (*dto.DashboardSalespeopleFilters, error) {
	if filters == nil {
		return &dto.DashboardSalespeopleFilters{}, nil
	}

	normalized := *filters
	normalized.SourceCompany = strings.TrimSpace(filters.SourceCompany)

	if normalized.SalespersonID != nil && *normalized.SalespersonID <= 0 {
		return nil, apperror.BadRequest("salesperson_id deve ser maior que zero")
	}
	if normalized.Year != nil && *normalized.Year <= 0 {
		return nil, apperror.BadRequest("year deve ser maior que zero")
	}
	if normalized.Month != nil && (*normalized.Month < 1 || *normalized.Month > 12) {
		return nil, apperror.BadRequest("month deve estar entre 1 e 12")
	}

	return &normalized, nil
}

func buildSalespeopleDashboardResponse(
	summary *dto.DashboardSummaryResponse,
	availableYears []int,
	salespersonSummaries []dto.DashboardSalespersonSummaryResponse,
	staleBudgets []dto.DashboardStaleBudgetResponse,
	monthlyEvolution []dto.DashboardMonthlyEvolutionResponse,
) *dto.SalespeopleDashboardResponse {
	response := &dto.SalespeopleDashboardResponse{
		AvailableYears:             availableYears,
		MonthlyEvolution:           monthlyEvolution,
		NegotiationPipeline:        make([]dto.DashboardSalespersonSummaryResponse, 0),
		RecentSalespeople:          make([]dto.DashboardSalespersonSummaryResponse, 0),
		SalespersonFunnel:          make([]dto.DashboardSalespersonFunnelResponse, 0),
		StaleBudgets:               staleBudgets,
		TopSalespeopleByBudgetCount: make([]dto.DashboardSalespersonSummaryResponse, 0),
		TopSalespeopleByConversion: make([]dto.DashboardSalespersonSummaryResponse, 0),
		TopSalespeopleByAverageTicket: make([]dto.DashboardSalespersonSummaryResponse, 0),
		TopSalespeopleByValue:      make([]dto.DashboardSalespersonSummaryResponse, 0),
	}

	if summary == nil {
		summary = &dto.DashboardSummaryResponse{}
	}
	summary.ActiveSalespeople = len(salespersonSummaries)
	response.Summary = *summary

	if len(salespersonSummaries) == 0 {
		return response
	}

	sort.Slice(salespersonSummaries, func(firstIndex int, secondIndex int) bool {
		return salespersonSummaries[firstIndex].Label < salespersonSummaries[secondIndex].Label
	})
	response.TopSalespeopleByValue = limitSalespersonSummaries(
		salespersonSummaries,
		func(firstItem dto.DashboardSalespersonSummaryResponse, secondItem dto.DashboardSalespersonSummaryResponse) bool {
			if firstItem.GrossValue != secondItem.GrossValue {
				return firstItem.GrossValue > secondItem.GrossValue
			}

			return firstItem.BudgetCount > secondItem.BudgetCount
		},
		10,
	)
	response.TopSalespeopleByBudgetCount = limitSalespersonSummaries(
		salespersonSummaries,
		func(firstItem dto.DashboardSalespersonSummaryResponse, secondItem dto.DashboardSalespersonSummaryResponse) bool {
			if firstItem.BudgetCount != secondItem.BudgetCount {
				return firstItem.BudgetCount > secondItem.BudgetCount
			}

			return firstItem.GrossValue > secondItem.GrossValue
		},
		10,
	)
	efficiencyBase := getComparableSalespeopleForEfficiency(salespersonSummaries)
	response.TopSalespeopleByConversion = limitSalespersonSummaries(
		efficiencyBase,
		func(firstItem dto.DashboardSalespersonSummaryResponse, secondItem dto.DashboardSalespersonSummaryResponse) bool {
			if firstItem.ConversionRate != secondItem.ConversionRate {
				return firstItem.ConversionRate > secondItem.ConversionRate
			}
			if firstItem.WonBudgetCount != secondItem.WonBudgetCount {
				return firstItem.WonBudgetCount > secondItem.WonBudgetCount
			}

			return firstItem.BudgetCount > secondItem.BudgetCount
		},
		10,
	)
	response.TopSalespeopleByAverageTicket = limitSalespersonSummaries(
		efficiencyBase,
		func(firstItem dto.DashboardSalespersonSummaryResponse, secondItem dto.DashboardSalespersonSummaryResponse) bool {
			if firstItem.AverageTicket != secondItem.AverageTicket {
				return firstItem.AverageTicket > secondItem.AverageTicket
			}
			if firstItem.GrossValue != secondItem.GrossValue {
				return firstItem.GrossValue > secondItem.GrossValue
			}

			return firstItem.BudgetCount > secondItem.BudgetCount
		},
		10,
	)
	response.NegotiationPipeline = limitSalespersonSummaries(
		filterSalespersonSummaries(
			salespersonSummaries,
			func(item dto.DashboardSalespersonSummaryResponse) bool {
				return item.NegotiationBudgetCount > 0
			},
		),
		func(firstItem dto.DashboardSalespersonSummaryResponse, secondItem dto.DashboardSalespersonSummaryResponse) bool {
			if firstItem.NegotiationGrossValue != secondItem.NegotiationGrossValue {
				return firstItem.NegotiationGrossValue > secondItem.NegotiationGrossValue
			}

			return firstItem.NegotiationBudgetCount > secondItem.NegotiationBudgetCount
		},
		10,
	)
	response.RecentSalespeople = limitSalespersonSummaries(
		filterSalespersonSummaries(
			salespersonSummaries,
			func(item dto.DashboardSalespersonSummaryResponse) bool {
				return item.LastActivityAt != nil
			},
		),
		func(firstItem dto.DashboardSalespersonSummaryResponse, secondItem dto.DashboardSalespersonSummaryResponse) bool {
			if firstItem.LastActivityAt == nil {
				return false
			}
			if secondItem.LastActivityAt == nil {
				return true
			}

			return firstItem.LastActivityAt.After(*secondItem.LastActivityAt)
		},
		10,
	)
	response.SalespersonFunnel = buildSalespersonFunnel(salespersonSummaries)

	return response
}

func getComparableSalespeopleForEfficiency(
	items []dto.DashboardSalespersonSummaryResponse,
) []dto.DashboardSalespersonSummaryResponse {
	comparableItems := filterSalespersonSummaries(
		items,
		func(item dto.DashboardSalespersonSummaryResponse) bool {
			return item.BudgetCount >= 2
		},
	)

	if len(comparableItems) > 0 {
		return comparableItems
	}

	return items
}

func limitSalespersonSummaries(
	items []dto.DashboardSalespersonSummaryResponse,
	sortLess func(firstItem dto.DashboardSalespersonSummaryResponse, secondItem dto.DashboardSalespersonSummaryResponse) bool,
	limit int,
) []dto.DashboardSalespersonSummaryResponse {
	clonedItems := append([]dto.DashboardSalespersonSummaryResponse{}, items...)
	sort.Slice(clonedItems, func(firstIndex int, secondIndex int) bool {
		return sortLess(clonedItems[firstIndex], clonedItems[secondIndex])
	})

	if len(clonedItems) > limit {
		return clonedItems[:limit]
	}

	return clonedItems
}

func filterSalespersonSummaries(
	items []dto.DashboardSalespersonSummaryResponse,
	matches func(item dto.DashboardSalespersonSummaryResponse) bool,
) []dto.DashboardSalespersonSummaryResponse {
	filteredItems := make([]dto.DashboardSalespersonSummaryResponse, 0)
	for _, item := range items {
		if matches(item) {
			filteredItems = append(filteredItems, item)
		}
	}

	return filteredItems
}

func buildSalespersonFunnel(
	items []dto.DashboardSalespersonSummaryResponse,
) []dto.DashboardSalespersonFunnelResponse {
	funnelItems := make([]dto.DashboardSalespersonFunnelResponse, 0, len(items))
	for _, item := range items {
		funnelItems = append(funnelItems, dto.DashboardSalespersonFunnelResponse{
			ConversionRate:     item.ConversionRate,
			Label:              item.Label,
			LostBudgets:        maxInt(0, item.BudgetCount-item.NegotiationBudgetCount-item.WonBudgetCount),
			NegotiationBudgets: item.NegotiationBudgetCount,
			TotalBudgets:       item.BudgetCount,
			WonBudgets:         item.WonBudgetCount,
		})
	}

	sort.Slice(funnelItems, func(firstIndex int, secondIndex int) bool {
		if funnelItems[firstIndex].TotalBudgets != funnelItems[secondIndex].TotalBudgets {
			return funnelItems[firstIndex].TotalBudgets > funnelItems[secondIndex].TotalBudgets
		}

		return funnelItems[firstIndex].WonBudgets > funnelItems[secondIndex].WonBudgets
	})
	if len(funnelItems) > 10 {
		return funnelItems[:10]
	}

	return funnelItems
}

func maxInt(firstValue int, secondValue int) int {
	if firstValue > secondValue {
		return firstValue
	}

	return secondValue
}
