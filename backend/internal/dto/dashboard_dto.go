package dto

import "time"

type DashboardSalespeopleFilters struct {
	SourceCompany           string
	SalespersonID           *int64
	InstallerID             *int64
	StatusID                *int64
	RestrictedSalespersonID *int64
	Year                    *int
	Month                   *int
}

type DashboardSummaryResponse struct {
	ActiveSalespeople          int     `json:"active_salespeople"`
	TotalBudgets               int     `json:"total_budgets"`
	TotalGrossValue            float64 `json:"total_gross_value"`
	AverageTicket              float64 `json:"average_ticket"`
	TotalNegotiationGrossValue float64 `json:"total_negotiation_gross_value"`
	ConversionRate             float64 `json:"conversion_rate"`
	ValueConversionRate        float64 `json:"value_conversion_rate"`
	WonBudgets                 int     `json:"won_budgets"`
	NegotiationBudgets         int     `json:"negotiation_budgets"`
	LostBudgets                int     `json:"lost_budgets"`
	StalledBudgetsCount        int     `json:"stalled_budgets_count"`
}

type DashboardSalespersonSummaryResponse struct {
	Label                  string     `json:"label"`
	BudgetCount            int        `json:"budget_count"`
	GrossValue             float64    `json:"gross_value"`
	NegotiationBudgetCount int        `json:"negotiation_budget_count"`
	NegotiationGrossValue  float64    `json:"negotiation_gross_value"`
	WonBudgetCount         int        `json:"won_budget_count"`
	StalledBudgetCount     int        `json:"stalled_budget_count"`
	AverageTicket          float64    `json:"average_ticket"`
	ConversionRate         float64    `json:"conversion_rate"`
	LastActivityAt         *time.Time `json:"last_activity_at,omitempty"`
}

type DashboardEstimatorSummaryResponse struct {
	Label                  string     `json:"label"`
	BudgetCount            int        `json:"budget_count"`
	GrossValue             float64    `json:"gross_value"`
	NegotiationBudgetCount int        `json:"negotiation_budget_count"`
	NegotiationGrossValue  float64    `json:"negotiation_gross_value"`
	WonBudgetCount         int        `json:"won_budget_count"`
	StalledBudgetCount     int        `json:"stalled_budget_count"`
	AverageTicket          float64    `json:"average_ticket"`
	ConversionRate         float64    `json:"conversion_rate"`
	LastActivityAt         *time.Time `json:"last_activity_at,omitempty"`
}

type DashboardSalespersonFunnelResponse struct {
	Label              string  `json:"label"`
	TotalBudgets       int     `json:"total_budgets"`
	NegotiationBudgets int     `json:"negotiation_budgets"`
	WonBudgets         int     `json:"won_budgets"`
	LostBudgets        int     `json:"lost_budgets"`
	ConversionRate     float64 `json:"conversion_rate"`
}

type DashboardStaleBudgetResponse struct {
	ID                       int64     `json:"id"`
	BudgetNumber             string    `json:"budget_number"`
	SalespersonLabel         string    `json:"salesperson_label"`
	ProjectLabel             string    `json:"project_label"`
	StatusLabel              string    `json:"status_label"`
	ConstructionCompanyLabel string    `json:"construction_company_label"`
	GrossValue               float64   `json:"gross_value"`
	LastActivityAt           time.Time `json:"last_activity_at"`
	StalledDays              int       `json:"stalled_days"`
}

type DashboardMonthlyEvolutionResponse struct {
	MonthKey       string  `json:"month_key"`
	MonthLabel     string  `json:"month_label"`
	BudgetCount    int     `json:"budget_count"`
	GrossValue     float64 `json:"gross_value"`
	WonBudgetCount int     `json:"won_budget_count"`
	WonGrossValue  float64 `json:"won_gross_value"`
}

type DashboardEntityPerformanceResponse struct {
	Label               string     `json:"label"`
	ProjectID           *int64     `json:"project_id,omitempty"`
	BudgetCount         int        `json:"budget_count"`
	WonBudgetCount      int        `json:"won_budget_count"`
	LostBudgetCount     int        `json:"lost_budget_count"`
	GrossValue          float64    `json:"gross_value"`
	WonGrossValue       float64    `json:"won_gross_value"`
	ConversionRate      float64    `json:"conversion_rate"`
	ValueConversionRate float64    `json:"value_conversion_rate"`
	LastActivityAt      *time.Time `json:"last_activity_at,omitempty"`
}

type DashboardLossReasonSummaryResponse struct {
	Label           string  `json:"label"`
	LostBudgetCount int     `json:"lost_budget_count"`
	GrossValue      float64 `json:"gross_value"`
	AverageTicket   float64 `json:"average_ticket"`
}

type DashboardClosingTimeSummaryResponse struct {
	Label              string  `json:"label"`
	BudgetCount        int     `json:"budget_count"`
	AverageClosingDays float64 `json:"average_closing_days"`
	GrossValue         float64 `json:"gross_value"`
}

type DashboardTechnicalSummaryResponse struct {
	ActiveEstimators           int     `json:"active_estimators"`
	BudgetsWithEstimator       int     `json:"budgets_with_estimator"`
	BudgetsWithoutEstimator    int     `json:"budgets_without_estimator"`
	CoverageRate               float64 `json:"coverage_rate"`
	TotalGrossValue            float64 `json:"total_gross_value"`
	AverageTicket              float64 `json:"average_ticket"`
	TotalNegotiationGrossValue float64 `json:"total_negotiation_gross_value"`
	WonBudgets                 int     `json:"won_budgets"`
	NegotiationBudgets         int     `json:"negotiation_budgets"`
	LostBudgets                int     `json:"lost_budgets"`
	StalledBudgetsCount        int     `json:"stalled_budgets_count"`
	ConversionRate             float64 `json:"conversion_rate"`
}

type DashboardTechnicalOverviewResponse struct {
	Summary                      DashboardTechnicalSummaryResponse   `json:"summary"`
	TopEstimatorsByValue         []DashboardEstimatorSummaryResponse `json:"top_estimators_by_value"`
	TopEstimatorsByBudgetCount   []DashboardEstimatorSummaryResponse `json:"top_estimators_by_budget_count"`
	TopEstimatorsByAverageTicket []DashboardEstimatorSummaryResponse `json:"top_estimators_by_average_ticket"`
	RecentEstimators             []DashboardEstimatorSummaryResponse `json:"recent_estimators"`
}

type SalespeopleDashboardResponse struct {
	AvailableYears                []int                                 `json:"available_years"`
	Summary                       DashboardSummaryResponse              `json:"summary"`
	TopSalespeopleByValue         []DashboardSalespersonSummaryResponse `json:"top_salespeople_by_value"`
	TopSalespeopleByBudgetCount   []DashboardSalespersonSummaryResponse `json:"top_salespeople_by_budget_count"`
	TopSalespeopleByConversion    []DashboardSalespersonSummaryResponse `json:"top_salespeople_by_conversion"`
	TopSalespeopleByAverageTicket []DashboardSalespersonSummaryResponse `json:"top_salespeople_by_average_ticket"`
	NegotiationPipeline           []DashboardSalespersonSummaryResponse `json:"negotiation_pipeline"`
	RecentSalespeople             []DashboardSalespersonSummaryResponse `json:"recent_salespeople"`
	SalespersonFunnel             []DashboardSalespersonFunnelResponse  `json:"salesperson_funnel"`
	StaleBudgets                  []DashboardStaleBudgetResponse        `json:"stale_budgets"`
	MonthlyEvolution              []DashboardMonthlyEvolutionResponse   `json:"monthly_evolution"`
	TopConstructionCompanies      []DashboardEntityPerformanceResponse  `json:"top_construction_companies"`
	TopProjects                   []DashboardEntityPerformanceResponse  `json:"top_projects"`
	TopLossReasons                []DashboardLossReasonSummaryResponse  `json:"top_loss_reasons"`
	AverageClosingTimes           []DashboardClosingTimeSummaryResponse `json:"average_closing_times"`
	TechnicalOverview             DashboardTechnicalOverviewResponse    `json:"technical_overview"`
}
