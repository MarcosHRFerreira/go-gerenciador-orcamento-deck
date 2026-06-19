export type DashboardSalespeopleFilters = {
  sourceCompany: string;
  salespersonId: string;
  year: string;
  month: string;
};

export type DashboardSummary = {
  activeSalespeople: number;
  totalBudgets: number;
  totalGrossValue: number;
  averageTicket: number;
  totalNegotiationGrossValue: number;
  conversionRate: number;
  valueConversionRate: number;
  wonBudgets: number;
  negotiationBudgets: number;
  lostBudgets: number;
  stalledBudgetsCount: number;
};

export type DashboardSalespersonSummary = {
  label: string;
  budgetCount: number;
  grossValue: number;
  negotiationBudgetCount: number;
  negotiationGrossValue: number;
  wonBudgetCount: number;
  stalledBudgetCount: number;
  averageTicket: number;
  conversionRate: number;
  lastActivityAt: string | null;
};

export type DashboardEstimatorSummary = {
  label: string;
  budgetCount: number;
  grossValue: number;
  negotiationBudgetCount: number;
  negotiationGrossValue: number;
  wonBudgetCount: number;
  stalledBudgetCount: number;
  averageTicket: number;
  conversionRate: number;
  lastActivityAt: string | null;
};

export type DashboardSalespersonFunnelItem = {
  label: string;
  totalBudgets: number;
  negotiationBudgets: number;
  wonBudgets: number;
  lostBudgets: number;
  conversionRate: number;
};

export type DashboardStaleBudgetItem = {
  id: number;
  budgetNumber: string;
  salespersonLabel: string;
  projectLabel: string;
  statusLabel: string;
  constructionCompanyLabel: string;
  grossValue: number;
  lastActivityAt: string;
  stalledDays: number;
};

export type DashboardMonthlyEvolutionItem = {
  monthKey: string;
  monthLabel: string;
  budgetCount: number;
  grossValue: number;
  wonBudgetCount: number;
  wonGrossValue: number;
};

export type DashboardEntityPerformanceItem = {
  label: string;
  budgetCount: number;
  wonBudgetCount: number;
  lostBudgetCount: number;
  grossValue: number;
  wonGrossValue: number;
  conversionRate: number;
  valueConversionRate: number;
  lastActivityAt: string | null;
};

export type DashboardLossReasonItem = {
  label: string;
  lostBudgetCount: number;
  grossValue: number;
  averageTicket: number;
};

export type DashboardClosingTimeItem = {
  label: string;
  budgetCount: number;
  averageClosingDays: number;
  grossValue: number;
};

export type DashboardTechnicalSummary = {
  activeEstimators: number;
  budgetsWithEstimator: number;
  budgetsWithoutEstimator: number;
  coverageRate: number;
  totalGrossValue: number;
  averageTicket: number;
  totalNegotiationGrossValue: number;
  wonBudgets: number;
  negotiationBudgets: number;
  lostBudgets: number;
  stalledBudgetsCount: number;
  conversionRate: number;
};

export type DashboardTechnicalOverview = {
  summary: DashboardTechnicalSummary;
  topEstimatorsByValue: DashboardEstimatorSummary[];
  topEstimatorsByBudgetCount: DashboardEstimatorSummary[];
  topEstimatorsByAverageTicket: DashboardEstimatorSummary[];
  recentEstimators: DashboardEstimatorSummary[];
};

export type DashboardSalespeopleData = {
  availableYears: string[];
  summary: DashboardSummary;
  topSalespeopleByValue: DashboardSalespersonSummary[];
  topSalespeopleByBudgetCount: DashboardSalespersonSummary[];
  topSalespeopleByConversion: DashboardSalespersonSummary[];
  topSalespeopleByAverageTicket: DashboardSalespersonSummary[];
  negotiationPipeline: DashboardSalespersonSummary[];
  recentSalespeople: DashboardSalespersonSummary[];
  salespersonFunnel: DashboardSalespersonFunnelItem[];
  staleBudgets: DashboardStaleBudgetItem[];
  monthlyEvolution: DashboardMonthlyEvolutionItem[];
  topConstructionCompanies: DashboardEntityPerformanceItem[];
  topProjects: DashboardEntityPerformanceItem[];
  topLossReasons: DashboardLossReasonItem[];
  averageClosingTimes: DashboardClosingTimeItem[];
  technicalOverview: DashboardTechnicalOverview;
};
