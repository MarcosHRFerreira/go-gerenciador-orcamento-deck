package unit

import (
	"context"
	"testing"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/dto"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/model"
	dashboardservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/dashboard"
)

type dashboardRepositoryStub struct {
	summary           *dto.DashboardSummaryResponse
	summaryErr        error
	salespeople       []dto.DashboardSalespersonSummaryResponse
	salespeopleErr    error
	staleBudgets      []dto.DashboardStaleBudgetResponse
	staleBudgetsErr   error
	monthlyEvolution  []dto.DashboardMonthlyEvolutionResponse
	monthlyErr        error
	availableYears    []int
	availableYearsErr error
	capturedFilters   *dto.DashboardSalespeopleFilters
}

type dashboardSalespersonRepositoryStub struct {
	getByUsernameItem *model.SalespersonModel
	getByUsernameErr  error
}

func (s *dashboardRepositoryStub) GetSummary(_ context.Context, filters *dto.DashboardSalespeopleFilters) (*dto.DashboardSummaryResponse, error) {
	s.capturedFilters = filters
	return s.summary, s.summaryErr
}

func (s *dashboardRepositoryStub) ListSalespersonSummaries(_ context.Context, _ *dto.DashboardSalespeopleFilters) ([]dto.DashboardSalespersonSummaryResponse, error) {
	return s.salespeople, s.salespeopleErr
}

func (s *dashboardRepositoryStub) ListStaleBudgets(_ context.Context, _ *dto.DashboardSalespeopleFilters, _ int) ([]dto.DashboardStaleBudgetResponse, error) {
	return s.staleBudgets, s.staleBudgetsErr
}

func (s *dashboardRepositoryStub) ListMonthlyEvolution(_ context.Context, _ *dto.DashboardSalespeopleFilters, _ int) ([]dto.DashboardMonthlyEvolutionResponse, error) {
	return s.monthlyEvolution, s.monthlyErr
}

func (s *dashboardRepositoryStub) ListAvailableYears(_ context.Context, _ *dto.DashboardSalespeopleFilters) ([]int, error) {
	return s.availableYears, s.availableYearsErr
}

func (s *dashboardSalespersonRepositoryStub) Create(_ context.Context, _ *model.SalespersonModel) (int64, error) {
	return 0, nil
}

func (s *dashboardSalespersonRepositoryStub) List(_ context.Context) ([]model.SalespersonModel, error) {
	return nil, nil
}

func (s *dashboardSalespersonRepositoryStub) GetByEmail(_ context.Context, _ string) (*model.SalespersonModel, error) {
	return nil, nil
}

func (s *dashboardSalespersonRepositoryStub) GetByUsername(_ context.Context, _ string) (*model.SalespersonModel, error) {
	return s.getByUsernameItem, s.getByUsernameErr
}

func (s *dashboardSalespersonRepositoryStub) GetByID(_ context.Context, _ int64) (*model.SalespersonModel, error) {
	return nil, nil
}

func (s *dashboardSalespersonRepositoryStub) Update(_ context.Context, _ *model.SalespersonModel) error {
	return nil
}

func (s *dashboardSalespersonRepositoryStub) Delete(_ context.Context, _ int64) error {
	return nil
}

func TestDashboardServiceShouldAggregateSummaryAndTimeline(t *testing.T) {
	now := time.Date(2026, time.June, 17, 12, 0, 0, 0, time.UTC)
	repo := &dashboardRepositoryStub{
		availableYears: []int{2026, 2025},
		summary: &dto.DashboardSummaryResponse{
			AverageTicket:              1875,
			ConversionRate:             25,
			LostBudgets:                1,
			NegotiationBudgets:         2,
			StalledBudgetsCount:        2,
			TotalBudgets:               4,
			TotalGrossValue:            7500,
			TotalNegotiationGrossValue: 5000,
			WonBudgets:                 1,
		},
		salespeople: []dto.DashboardSalespersonSummaryResponse{
			{
				AverageTicket:          1500,
				BudgetCount:            2,
				ConversionRate:         50,
				GrossValue:             3000,
				Label:                  "Ana",
				LastActivityAt:         timePointer(time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC)),
				NegotiationBudgetCount: 1,
				NegotiationGrossValue:  2000,
				StalledBudgetCount:     1,
				WonBudgetCount:         1,
			},
			{
				AverageTicket:          2250,
				BudgetCount:            2,
				ConversionRate:         0,
				GrossValue:             4500,
				Label:                  "Bruno",
				LastActivityAt:         timePointer(time.Date(2026, time.May, 20, 0, 0, 0, 0, time.UTC)),
				NegotiationBudgetCount: 1,
				NegotiationGrossValue:  3000,
				StalledBudgetCount:     1,
				WonBudgetCount:         0,
			},
		},
		staleBudgets: []dto.DashboardStaleBudgetResponse{
			{
				BudgetNumber:             "ORC-4",
				ConstructionCompanyLabel: "Construtora D",
				GrossValue:               3000,
				ID:                       4,
				LastActivityAt:           time.Date(2026, time.May, 20, 0, 0, 0, 0, time.UTC),
				ProjectLabel:             "Obra D",
				SalespersonLabel:         "Bruno",
				StalledDays:              28,
				StatusLabel:              "Em Negociacao",
			},
			{
				BudgetNumber:             "ORC-2",
				ConstructionCompanyLabel: "Construtora B",
				GrossValue:               2000,
				ID:                       2,
				LastActivityAt:           time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC),
				ProjectLabel:             "Obra B",
				SalespersonLabel:         "Ana",
				StalledDays:              16,
				StatusLabel:              "Em Negociacao",
			},
		},
		monthlyEvolution: []dto.DashboardMonthlyEvolutionResponse{
			{
				BudgetCount:    1,
				GrossValue:     1000,
				MonthKey:       "2026-01",
				MonthLabel:     "Jan/2026",
				WonBudgetCount: 1,
				WonGrossValue:  1000,
			},
			{
				BudgetCount:    1,
				GrossValue:     2000,
				MonthKey:       "2026-02",
				MonthLabel:     "Fev/2026",
				WonBudgetCount: 0,
				WonGrossValue:  0,
			},
			{
				BudgetCount:    2,
				GrossValue:     4500,
				MonthKey:       "2026-03",
				MonthLabel:     "Mar/2026",
				WonBudgetCount: 0,
				WonGrossValue:  0,
			},
		},
	}
	service := dashboardservice.NewServiceWithClock(
		repo,
		&dashboardSalespersonRepositoryStub{},
		func() time.Time { return now },
	)

	response, err := service.GetSalespeopleDashboard(
		context.Background(),
		&dto.DashboardSalespeopleFilters{},
		model.RoleAdmin,
		"admin",
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if response.Summary.TotalBudgets != 4 {
		t.Fatalf("expected total budgets 4, got %d", response.Summary.TotalBudgets)
	}
	if response.Summary.WonBudgets != 1 {
		t.Fatalf("expected won budgets 1, got %d", response.Summary.WonBudgets)
	}
	if response.Summary.NegotiationBudgets != 2 {
		t.Fatalf("expected negotiation budgets 2, got %d", response.Summary.NegotiationBudgets)
	}
	if response.Summary.LostBudgets != 1 {
		t.Fatalf("expected lost budgets 1, got %d", response.Summary.LostBudgets)
	}
	if response.Summary.StalledBudgetsCount != 2 {
		t.Fatalf("expected stalled budgets 2, got %d", response.Summary.StalledBudgetsCount)
	}
	if len(response.TopSalespeopleByValue) != 2 {
		t.Fatalf("expected 2 salespeople in ranking, got %d", len(response.TopSalespeopleByValue))
	}
	if response.TopSalespeopleByValue[0].Label != "Bruno" {
		t.Fatalf("expected Bruno as top by value, got %s", response.TopSalespeopleByValue[0].Label)
	}
	if len(response.SalespersonFunnel) != 2 {
		t.Fatalf("expected 2 funnel items, got %d", len(response.SalespersonFunnel))
	}
	if response.SalespersonFunnel[0].Label != "Ana" {
		t.Fatalf("expected Ana first in funnel due tie-break, got %s", response.SalespersonFunnel[0].Label)
	}
	if len(response.StaleBudgets) != 2 {
		t.Fatalf("expected 2 stale budgets, got %d", len(response.StaleBudgets))
	}
	if response.StaleBudgets[0].BudgetNumber != "ORC-4" {
		t.Fatalf("expected ORC-4 as stalest budget, got %s", response.StaleBudgets[0].BudgetNumber)
	}
	if len(response.MonthlyEvolution) != 3 {
		t.Fatalf("expected 3 monthly evolution items, got %d", len(response.MonthlyEvolution))
	}
	if response.MonthlyEvolution[2].MonthKey != "2026-03" {
		t.Fatalf("expected last month key 2026-03, got %s", response.MonthlyEvolution[2].MonthKey)
	}
	if len(response.AvailableYears) != 2 || response.AvailableYears[0] != 2026 {
		t.Fatalf("expected available years [2026 2025], got %+v", response.AvailableYears)
	}
	if response.Summary.ActiveSalespeople != 2 {
		t.Fatalf("expected active salespeople 2, got %d", response.Summary.ActiveSalespeople)
	}
	if len(response.TopSalespeopleByConversion) != 2 || response.TopSalespeopleByConversion[0].Label != "Ana" {
		t.Fatalf("expected Ana as top by conversion, got %+v", response.TopSalespeopleByConversion)
	}
	if len(response.TopSalespeopleByAverageTicket) != 2 || response.TopSalespeopleByAverageTicket[0].Label != "Bruno" {
		t.Fatalf("expected Bruno as top by average ticket, got %+v", response.TopSalespeopleByAverageTicket)
	}
}

func TestDashboardServiceShouldRestrictScopeForUserRole(t *testing.T) {
	repo := &dashboardRepositoryStub{}
	service := dashboardservice.NewServiceWithClock(
		repo,
		&dashboardSalespersonRepositoryStub{
			getByUsernameItem: &model.SalespersonModel{
				ID:     77,
				Name:   "vendedor",
				Active: true,
			},
		},
		func() time.Time { return time.Date(2026, time.June, 17, 0, 0, 0, 0, time.UTC) },
	)

	_, err := service.GetSalespeopleDashboard(
		context.Background(),
		&dto.DashboardSalespeopleFilters{},
		model.RoleUser,
		"vendedor",
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.capturedFilters == nil || repo.capturedFilters.RestrictedSalespersonID == nil {
		t.Fatal("expected restricted salesperson id to be captured")
	}
	if *repo.capturedFilters.RestrictedSalespersonID != 77 {
		t.Fatalf("expected restricted salesperson id 77, got %d", *repo.capturedFilters.RestrictedSalespersonID)
	}
}

func timePointer(value time.Time) *time.Time {
	return &value
}
