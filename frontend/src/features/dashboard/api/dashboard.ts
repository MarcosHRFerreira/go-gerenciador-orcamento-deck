import { api } from "../../../lib/axios/api";
import { AxiosError } from "axios";
import { getBudgetProjectListRequest } from "../../budgets/api/budgets";
import type {
  BudgetListFilters,
  BudgetListItem,
} from "../../budgets/types/budget";
import type {
  DashboardClosingTimeItem,
  DashboardEstimatorSummary,
  DashboardEntityPerformanceItem,
  DashboardLossReasonItem,
  DashboardMonthlyEvolutionItem,
  DashboardSalespeopleData,
  DashboardSalespeopleFilters,
  DashboardSalespersonFunnelItem,
  DashboardSalespersonSummary,
  DashboardStaleBudgetItem,
  DashboardSummary,
  DashboardTechnicalOverview,
  DashboardTechnicalSummary,
} from "../types/dashboard";

const monthLabels = [
  "Jan",
  "Fev",
  "Mar",
  "Abr",
  "Mai",
  "Jun",
  "Jul",
  "Ago",
  "Set",
  "Out",
  "Nov",
  "Dez",
] as const;

type DashboardSummaryApiResponse = {
  active_salespeople: number;
  total_budgets: number;
  total_gross_value: number;
  average_ticket: number;
  total_negotiation_gross_value: number;
  conversion_rate: number;
  value_conversion_rate: number;
  won_budgets: number;
  negotiation_budgets: number;
  lost_budgets: number;
  stalled_budgets_count: number;
};

type DashboardSalespersonSummaryApiResponse = {
  label: string;
  budget_count: number;
  gross_value: number;
  negotiation_budget_count: number;
  negotiation_gross_value: number;
  won_budget_count: number;
  stalled_budget_count: number;
  average_ticket: number;
  conversion_rate: number;
  last_activity_at?: string | null;
};

type DashboardEstimatorSummaryApiResponse = {
  label: string;
  budget_count: number;
  gross_value: number;
  negotiation_budget_count: number;
  negotiation_gross_value: number;
  won_budget_count: number;
  stalled_budget_count: number;
  average_ticket: number;
  conversion_rate: number;
  last_activity_at?: string | null;
};

type DashboardSalespersonFunnelApiResponse = {
  label: string;
  total_budgets: number;
  negotiation_budgets: number;
  won_budgets: number;
  lost_budgets: number;
  conversion_rate: number;
};

type DashboardStaleBudgetApiResponse = {
  id: number;
  budget_number: string;
  salesperson_label: string;
  project_label: string;
  status_label: string;
  construction_company_label: string;
  gross_value: number;
  last_activity_at: string;
  stalled_days: number;
};

type DashboardMonthlyEvolutionApiResponse = {
  month_key: string;
  month_label: string;
  budget_count: number;
  gross_value: number;
  won_budget_count: number;
  won_gross_value: number;
};

type DashboardEntityPerformanceApiResponse = {
  label: string;
  project_id?: number | null;
  budget_count: number;
  won_budget_count: number;
  lost_budget_count: number;
  gross_value: number;
  won_gross_value: number;
  conversion_rate: number;
  value_conversion_rate: number;
  last_activity_at?: string | null;
};

type DashboardLossReasonApiResponse = {
  label: string;
  lost_budget_count: number;
  gross_value: number;
  average_ticket: number;
};

type DashboardClosingTimeApiResponse = {
  label: string;
  budget_count: number;
  average_closing_days: number;
  gross_value: number;
};

type DashboardTechnicalSummaryApiResponse = {
  active_estimators: number;
  budgets_with_estimator: number;
  budgets_without_estimator: number;
  coverage_rate: number;
  total_gross_value: number;
  average_ticket: number;
  total_negotiation_gross_value: number;
  won_budgets: number;
  negotiation_budgets: number;
  lost_budgets: number;
  stalled_budgets_count: number;
  conversion_rate: number;
};

type DashboardTechnicalOverviewApiResponse = {
  summary: DashboardTechnicalSummaryApiResponse;
  top_estimators_by_value: DashboardEstimatorSummaryApiResponse[];
  top_estimators_by_budget_count: DashboardEstimatorSummaryApiResponse[];
  top_estimators_by_average_ticket: DashboardEstimatorSummaryApiResponse[];
  recent_estimators: DashboardEstimatorSummaryApiResponse[];
};

type DashboardSalespeopleApiResponse = {
  available_years: number[];
  summary: DashboardSummaryApiResponse;
  top_salespeople_by_value: DashboardSalespersonSummaryApiResponse[];
  top_salespeople_by_budget_count: DashboardSalespersonSummaryApiResponse[];
  top_salespeople_by_conversion: DashboardSalespersonSummaryApiResponse[];
  top_salespeople_by_average_ticket: DashboardSalespersonSummaryApiResponse[];
  negotiation_pipeline: DashboardSalespersonSummaryApiResponse[];
  recent_salespeople: DashboardSalespersonSummaryApiResponse[];
  salesperson_funnel: DashboardSalespersonFunnelApiResponse[];
  stale_budgets: DashboardStaleBudgetApiResponse[];
  monthly_evolution: DashboardMonthlyEvolutionApiResponse[];
  top_construction_companies: DashboardEntityPerformanceApiResponse[];
  top_projects: DashboardEntityPerformanceApiResponse[];
  top_loss_reasons: DashboardLossReasonApiResponse[];
  average_closing_times: DashboardClosingTimeApiResponse[];
  technical_overview: DashboardTechnicalOverviewApiResponse;
};

function mapSummary(response: DashboardSummaryApiResponse): DashboardSummary {
  return {
    activeSalespeople: response.active_salespeople,
    averageTicket: response.average_ticket,
    conversionRate: response.conversion_rate,
    lostBudgets: response.lost_budgets,
    negotiationBudgets: response.negotiation_budgets,
    stalledBudgetsCount: response.stalled_budgets_count,
    totalBudgets: response.total_budgets,
    totalGrossValue: response.total_gross_value,
    totalNegotiationGrossValue: response.total_negotiation_gross_value,
    valueConversionRate: response.value_conversion_rate,
    wonBudgets: response.won_budgets,
  };
}

function mapSalespersonSummary(
  response: DashboardSalespersonSummaryApiResponse,
): DashboardSalespersonSummary {
  return {
    averageTicket: response.average_ticket,
    budgetCount: response.budget_count,
    conversionRate: response.conversion_rate,
    grossValue: response.gross_value,
    label: response.label,
    lastActivityAt: response.last_activity_at ?? null,
    negotiationBudgetCount: response.negotiation_budget_count,
    negotiationGrossValue: response.negotiation_gross_value,
    stalledBudgetCount: response.stalled_budget_count,
    wonBudgetCount: response.won_budget_count,
  };
}

function mapEstimatorSummary(
  response: DashboardEstimatorSummaryApiResponse,
): DashboardEstimatorSummary {
  return {
    averageTicket: response.average_ticket,
    budgetCount: response.budget_count,
    conversionRate: response.conversion_rate,
    grossValue: response.gross_value,
    label: response.label,
    lastActivityAt: response.last_activity_at ?? null,
    negotiationBudgetCount: response.negotiation_budget_count,
    negotiationGrossValue: response.negotiation_gross_value,
    stalledBudgetCount: response.stalled_budget_count,
    wonBudgetCount: response.won_budget_count,
  };
}

function mapSalespersonFunnel(
  response: DashboardSalespersonFunnelApiResponse,
): DashboardSalespersonFunnelItem {
  return {
    conversionRate: response.conversion_rate,
    label: response.label,
    lostBudgets: response.lost_budgets,
    negotiationBudgets: response.negotiation_budgets,
    totalBudgets: response.total_budgets,
    wonBudgets: response.won_budgets,
  };
}

function mapStaleBudget(
  response: DashboardStaleBudgetApiResponse,
): DashboardStaleBudgetItem {
  return {
    budgetNumber: response.budget_number,
    constructionCompanyLabel: response.construction_company_label,
    grossValue: response.gross_value,
    id: response.id,
    lastActivityAt: response.last_activity_at,
    projectLabel: response.project_label,
    salespersonLabel: response.salesperson_label,
    stalledDays: response.stalled_days,
    statusLabel: response.status_label,
  };
}

function mapMonthlyEvolution(
  response: DashboardMonthlyEvolutionApiResponse,
): DashboardMonthlyEvolutionItem {
  return {
    budgetCount: response.budget_count,
    grossValue: response.gross_value,
    monthKey: response.month_key,
    monthLabel: response.month_label,
    wonBudgetCount: response.won_budget_count,
    wonGrossValue: response.won_gross_value,
  };
}

function mapEntityPerformance(
  response: DashboardEntityPerformanceApiResponse,
): DashboardEntityPerformanceItem {
  return {
    budgetCount: response.budget_count,
    conversionRate: response.conversion_rate,
    grossValue: response.gross_value,
    label: response.label,
    lastActivityAt: response.last_activity_at ?? null,
    lostBudgetCount: response.lost_budget_count,
    projectId: response.project_id ?? null,
    valueConversionRate: response.value_conversion_rate,
    wonBudgetCount: response.won_budget_count,
    wonGrossValue: response.won_gross_value,
  };
}

function mapLossReasonSummary(
  response: DashboardLossReasonApiResponse,
): DashboardLossReasonItem {
  return {
    averageTicket: response.average_ticket,
    grossValue: response.gross_value,
    label: response.label,
    lostBudgetCount: response.lost_budget_count,
  };
}

function mapClosingTimeSummary(
  response: DashboardClosingTimeApiResponse,
): DashboardClosingTimeItem {
  return {
    averageClosingDays: response.average_closing_days,
    budgetCount: response.budget_count,
    grossValue: response.gross_value,
    label: response.label,
  };
}

function mapTechnicalSummary(
  response?: DashboardTechnicalSummaryApiResponse,
): DashboardTechnicalSummary {
  return {
    activeEstimators: response?.active_estimators ?? 0,
    averageTicket: response?.average_ticket ?? 0,
    budgetsWithEstimator: response?.budgets_with_estimator ?? 0,
    budgetsWithoutEstimator: response?.budgets_without_estimator ?? 0,
    conversionRate: response?.conversion_rate ?? 0,
    coverageRate: response?.coverage_rate ?? 0,
    lostBudgets: response?.lost_budgets ?? 0,
    negotiationBudgets: response?.negotiation_budgets ?? 0,
    stalledBudgetsCount: response?.stalled_budgets_count ?? 0,
    totalGrossValue: response?.total_gross_value ?? 0,
    totalNegotiationGrossValue: response?.total_negotiation_gross_value ?? 0,
    wonBudgets: response?.won_budgets ?? 0,
  };
}

function mapTechnicalOverview(
  response?: DashboardTechnicalOverviewApiResponse,
): DashboardTechnicalOverview {
  return {
    recentEstimators: (response?.recent_estimators ?? []).map(
      mapEstimatorSummary,
    ),
    summary: mapTechnicalSummary(response?.summary),
    topEstimatorsByAverageTicket: (
      response?.top_estimators_by_average_ticket ?? []
    ).map(mapEstimatorSummary),
    topEstimatorsByBudgetCount: (
      response?.top_estimators_by_budget_count ?? []
    ).map(mapEstimatorSummary),
    topEstimatorsByValue: (response?.top_estimators_by_value ?? []).map(
      mapEstimatorSummary,
    ),
  };
}

function buildDashboardParams(filters: DashboardSalespeopleFilters) {
  return {
    installer_id: filters.installerId || undefined,
    month: filters.month || undefined,
    salesperson_id: filters.salespersonId || undefined,
    source_company: filters.sourceCompany || undefined,
    status_id: filters.statusId || undefined,
    year: filters.year || undefined,
  };
}

function buildLegacyDashboardBudgetFilters(
  filters: DashboardSalespeopleFilters,
): BudgetListFilters {
  return {
    budgetNumber: "",
    installerId: filters.installerId,
    page: 1,
    pageSize: 100,
    projectCode: "",
    projectName: "",
    salespersonId: filters.salespersonId,
    estimatorId: "",
    sentAtFrom: "",
    sentAtTo: "",
    systemTypeId: "",
    sortBy: "updated_at",
    sortOrder: "desc",
    sourceCompany: filters.sourceCompany,
    statusId: filters.statusId,
    yearBudget: "",
  };
}

function normalizeText(value: string | null | undefined) {
  return (value ?? "")
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "")
    .trim()
    .toLowerCase();
}

function normalizeDisplayValue(
  value: string | null | undefined,
  fallback: string,
) {
  const trimmedValue = (value ?? "").trim();
  if (!trimmedValue) {
    return fallback;
  }

  const normalizedValue = normalizeText(trimmedValue);
  if (normalizedValue === "nao informado" || normalizedValue === "-") {
    return fallback;
  }

  return trimmedValue;
}

function getStatusCategory(statusName: string | null) {
  const normalizedStatusName = normalizeText(statusName);

  if (normalizedStatusName === "pedido") {
    return "won";
  }
  if (normalizedStatusName === "cancelado") {
    return "lost";
  }

  return "negotiation";
}

function getLastActivityDate(budget: BudgetListItem) {
  const updatedAt = new Date(budget.updatedAt);
  if (!Number.isNaN(updatedAt.getTime())) {
    return updatedAt;
  }

  const sentAt = new Date(budget.sentAt);
  if (!Number.isNaN(sentAt.getTime())) {
    return sentAt;
  }

  return null;
}

function getDaysSince(date: Date, referenceDate: Date) {
  const millisecondsPerDay = 1000 * 60 * 60 * 24;
  return Math.max(
    0,
    Math.floor((referenceDate.getTime() - date.getTime()) / millisecondsPerDay),
  );
}

function getMonthKey(date: Date) {
  const month = String(date.getMonth() + 1).padStart(2, "0");
  return `${date.getFullYear()}-${month}`;
}

function getMonthLabel(date: Date) {
  return `${monthLabels[date.getMonth()]}/${date.getFullYear()}`;
}

function sortSalespersonSummaries(
  items: DashboardSalespersonSummary[],
  compare: (
    firstItem: DashboardSalespersonSummary,
    secondItem: DashboardSalespersonSummary,
  ) => number,
) {
  return [...items].sort(compare).slice(0, 10);
}

function sortEstimatorSummaries(
  items: DashboardEstimatorSummary[],
  compare: (
    firstItem: DashboardEstimatorSummary,
    secondItem: DashboardEstimatorSummary,
  ) => number,
) {
  return [...items].sort(compare).slice(0, 10);
}

function sortEntityPerformance(
  items: DashboardEntityPerformanceItem[],
  compare: (
    firstItem: DashboardEntityPerformanceItem,
    secondItem: DashboardEntityPerformanceItem,
  ) => number,
) {
  return [...items].sort(compare).slice(0, 10);
}

function getComparableSalespersonSummaries(
  items: DashboardSalespersonSummary[],
) {
  const comparableItems = items.filter((item) => item.budgetCount >= 2);

  if (comparableItems.length > 0) {
    return comparableItems;
  }

  return items;
}

function getComparableEstimatorSummaries(items: DashboardEstimatorSummary[]) {
  const comparableItems = items.filter((item) => item.budgetCount >= 2);

  if (comparableItems.length > 0) {
    return comparableItems;
  }

  return items;
}

function buildEntityPerformanceFromBudgetItems(
  budgetItems: BudgetListItem[],
  getLabel: (budget: BudgetListItem) => string,
) {
  const entityMap = budgetItems.reduce<
    Map<string, DashboardEntityPerformanceItem>
  >((currentMap, budget) => {
    const label = getLabel(budget);
    const statusCategory = getStatusCategory(budget.statusName);
    const lastActivityDate = getLastActivityDate(budget);
    const existingItem = currentMap.get(label);

    if (existingItem) {
      existingItem.budgetCount += 1;
      existingItem.grossValue += budget.grossValue;
      if (statusCategory === "won") {
        existingItem.wonBudgetCount += 1;
        existingItem.wonGrossValue += budget.grossValue;
      }
      if (statusCategory === "lost") {
        existingItem.lostBudgetCount += 1;
      }
      if (
        lastActivityDate !== null &&
        (existingItem.lastActivityAt === null ||
          lastActivityDate.getTime() >
            new Date(existingItem.lastActivityAt).getTime())
      ) {
        existingItem.lastActivityAt = lastActivityDate.toISOString();
      }
      return currentMap;
    }

    currentMap.set(label, {
      budgetCount: 1,
      conversionRate: 0,
      grossValue: budget.grossValue,
      label,
      lastActivityAt: lastActivityDate?.toISOString() ?? null,
      lostBudgetCount: statusCategory === "lost" ? 1 : 0,
      valueConversionRate: 0,
      wonBudgetCount: statusCategory === "won" ? 1 : 0,
      wonGrossValue: statusCategory === "won" ? budget.grossValue : 0,
    });
    return currentMap;
  }, new Map<string, DashboardEntityPerformanceItem>());

  return Array.from(entityMap.values()).map((item) => ({
    ...item,
    conversionRate:
      item.budgetCount === 0
        ? 0
        : (item.wonBudgetCount / item.budgetCount) * 100,
    valueConversionRate:
      item.grossValue === 0 ? 0 : (item.wonGrossValue / item.grossValue) * 100,
  }));
}

function buildLossReasonSummariesFromBudgetItems(
  budgetItems: BudgetListItem[],
) {
  const lossReasonMap = budgetItems
    .filter((budget) => getStatusCategory(budget.statusName) === "lost")
    .reduce<Map<string, DashboardLossReasonItem>>((currentMap, budget) => {
      const label = normalizeDisplayValue(
        budget.lossReasonName,
        "Motivo nao informado",
      );
      const existingItem = currentMap.get(label);
      if (existingItem) {
        existingItem.lostBudgetCount += 1;
        existingItem.grossValue += budget.grossValue;
        return currentMap;
      }

      currentMap.set(label, {
        averageTicket: 0,
        grossValue: budget.grossValue,
        label,
        lostBudgetCount: 1,
      });
      return currentMap;
    }, new Map<string, DashboardLossReasonItem>());

  return Array.from(lossReasonMap.values())
    .map((item) => ({
      ...item,
      averageTicket:
        item.lostBudgetCount === 0 ? 0 : item.grossValue / item.lostBudgetCount,
    }))
    .sort((firstItem, secondItem) => {
      if (secondItem.grossValue !== firstItem.grossValue) {
        return secondItem.grossValue - firstItem.grossValue;
      }
      return secondItem.lostBudgetCount - firstItem.lostBudgetCount;
    })
    .slice(0, 10);
}

function buildClosingTimeSummariesFromBudgetItems(
  budgetItems: BudgetListItem[],
) {
  const closedBudgets = budgetItems
    .map((budget) => {
      const statusCategory = getStatusCategory(budget.statusName);
      if (statusCategory !== "won" && statusCategory !== "lost") {
        return null;
      }

      const sentAtDate = new Date(budget.sentAt);
      const closingDate = getLastActivityDate(budget);
      if (
        Number.isNaN(sentAtDate.getTime()) ||
        closingDate === null ||
        Number.isNaN(closingDate.getTime())
      ) {
        return null;
      }

      return {
        closingDays: getDaysSince(sentAtDate, closingDate),
        grossValue: budget.grossValue,
        statusCategory,
      };
    })
    .filter(
      (
        item,
      ): item is {
        closingDays: number;
        grossValue: number;
        statusCategory: "won" | "lost";
      } => item !== null,
    );

  const buildSummary = (
    label: string,
    items: typeof closedBudgets,
  ): DashboardClosingTimeItem | null => {
    if (items.length === 0) {
      return null;
    }

    const budgetCount = items.length;
    const grossValue = items.reduce(
      (currentTotal, item) => currentTotal + item.grossValue,
      0,
    );
    const totalClosingDays = items.reduce(
      (currentTotal, item) => currentTotal + item.closingDays,
      0,
    );

    return {
      averageClosingDays: totalClosingDays / budgetCount,
      budgetCount,
      grossValue,
      label,
    };
  };

  return [
    buildSummary("Geral", closedBudgets),
    buildSummary(
      "Pedidos",
      closedBudgets.filter((item) => item.statusCategory === "won"),
    ),
    buildSummary(
      "Cancelados",
      closedBudgets.filter((item) => item.statusCategory === "lost"),
    ),
  ].filter((item): item is DashboardClosingTimeItem => item !== null);
}

function buildDashboardFromBudgetItems(
  filters: DashboardSalespeopleFilters,
  budgetItems: BudgetListItem[],
): DashboardSalespeopleData {
  const availableYears = Array.from(
    budgetItems.reduce<Set<string>>((currentYears, budget) => {
      const sentAtDate = new Date(budget.sentAt);
      if (!Number.isNaN(sentAtDate.getTime())) {
        currentYears.add(String(sentAtDate.getFullYear()));
      }
      return currentYears;
    }, new Set<string>()),
  ).sort((firstYear, secondYear) => secondYear.localeCompare(firstYear));

  const filteredBudgetItems = budgetItems.filter((budget) => {
    const sentAtDate = new Date(budget.sentAt);
    if (Number.isNaN(sentAtDate.getTime())) {
      return false;
    }

    const budgetYear = String(sentAtDate.getFullYear());
    const budgetMonth = String(sentAtDate.getMonth() + 1);

    if (filters.year && budgetYear !== filters.year) {
      return false;
    }
    if (filters.month && budgetMonth !== filters.month) {
      return false;
    }

    return true;
  });

  const now = new Date();
  const totalBudgets = filteredBudgetItems.length;
  const totalGrossValue = filteredBudgetItems.reduce(
    (currentTotal, budget) => currentTotal + budget.grossValue,
    0,
  );
  const wonGrossValue = filteredBudgetItems.reduce(
    (currentTotal, budget) =>
      getStatusCategory(budget.statusName) === "won"
        ? currentTotal + budget.grossValue
        : currentTotal,
    0,
  );
  const wonBudgets = filteredBudgetItems.filter(
    (budget) => getStatusCategory(budget.statusName) === "won",
  ).length;
  const lostBudgets = filteredBudgetItems.filter(
    (budget) => getStatusCategory(budget.statusName) === "lost",
  ).length;
  const negotiationBudgets = filteredBudgetItems.filter(
    (budget) => getStatusCategory(budget.statusName) === "negotiation",
  ).length;
  const totalNegotiationGrossValue = filteredBudgetItems.reduce(
    (currentTotal, budget) =>
      getStatusCategory(budget.statusName) === "negotiation"
        ? currentTotal + budget.grossValue
        : currentTotal,
    0,
  );
  const averageTicket = totalBudgets === 0 ? 0 : totalGrossValue / totalBudgets;
  const conversionRate =
    totalBudgets === 0 ? 0 : (wonBudgets / totalBudgets) * 100;
  const valueConversionRate =
    totalGrossValue === 0 ? 0 : (wonGrossValue / totalGrossValue) * 100;

  const salespersonSummaryMap = filteredBudgetItems.reduce<
    Map<string, DashboardSalespersonSummary>
  >((currentMap, budget) => {
    const label = normalizeDisplayValue(budget.salespersonName, "Sem vendedor");
    const existingItem = currentMap.get(label);
    const statusCategory = getStatusCategory(budget.statusName);
    const lastActivityDate = getLastActivityDate(budget);
    const lastActivityAt = lastActivityDate?.toISOString() ?? null;
    const stalledDays =
      lastActivityDate === null ? 0 : getDaysSince(lastActivityDate, now);
    const isStalled = statusCategory === "negotiation" && stalledDays >= 7;

    if (existingItem) {
      existingItem.budgetCount += 1;
      existingItem.grossValue += budget.grossValue;
      if (statusCategory === "negotiation") {
        existingItem.negotiationBudgetCount += 1;
        existingItem.negotiationGrossValue += budget.grossValue;
      }
      if (statusCategory === "won") {
        existingItem.wonBudgetCount += 1;
      }
      if (isStalled) {
        existingItem.stalledBudgetCount += 1;
      }
      if (
        lastActivityAt !== null &&
        (existingItem.lastActivityAt === null ||
          new Date(lastActivityAt).getTime() >
            new Date(existingItem.lastActivityAt).getTime())
      ) {
        existingItem.lastActivityAt = lastActivityAt;
      }
      return currentMap;
    }

    currentMap.set(label, {
      averageTicket: 0,
      budgetCount: 1,
      conversionRate: 0,
      grossValue: budget.grossValue,
      label,
      lastActivityAt,
      negotiationBudgetCount: statusCategory === "negotiation" ? 1 : 0,
      negotiationGrossValue:
        statusCategory === "negotiation" ? budget.grossValue : 0,
      stalledBudgetCount: isStalled ? 1 : 0,
      wonBudgetCount: statusCategory === "won" ? 1 : 0,
    });
    return currentMap;
  }, new Map<string, DashboardSalespersonSummary>());

  const salespersonSummaries = Array.from(salespersonSummaryMap.values())
    .map((item) => ({
      ...item,
      averageTicket:
        item.budgetCount === 0 ? 0 : item.grossValue / item.budgetCount,
      conversionRate:
        item.budgetCount === 0
          ? 0
          : (item.wonBudgetCount / item.budgetCount) * 100,
    }))
    .sort((firstItem, secondItem) =>
      firstItem.label.localeCompare(secondItem.label),
    );
  const estimatorSummaryMap = filteredBudgetItems.reduce<
    Map<string, DashboardEstimatorSummary>
  >((currentMap, budget) => {
    if (budget.estimatorId === null) {
      return currentMap;
    }

    const label = normalizeDisplayValue(
      budget.estimatorName,
      "Orcamentista sem nome",
    );
    const existingItem = currentMap.get(label);
    const statusCategory = getStatusCategory(budget.statusName);
    const lastActivityDate = getLastActivityDate(budget);
    const lastActivityAt = lastActivityDate?.toISOString() ?? null;
    const stalledDays =
      lastActivityDate === null ? 0 : getDaysSince(lastActivityDate, now);
    const isStalled = statusCategory === "negotiation" && stalledDays >= 7;

    if (existingItem) {
      existingItem.budgetCount += 1;
      existingItem.grossValue += budget.grossValue;
      if (statusCategory === "negotiation") {
        existingItem.negotiationBudgetCount += 1;
        existingItem.negotiationGrossValue += budget.grossValue;
      }
      if (statusCategory === "won") {
        existingItem.wonBudgetCount += 1;
      }
      if (isStalled) {
        existingItem.stalledBudgetCount += 1;
      }
      if (
        lastActivityAt !== null &&
        (existingItem.lastActivityAt === null ||
          new Date(lastActivityAt).getTime() >
            new Date(existingItem.lastActivityAt).getTime())
      ) {
        existingItem.lastActivityAt = lastActivityAt;
      }
      return currentMap;
    }

    currentMap.set(label, {
      averageTicket: 0,
      budgetCount: 1,
      conversionRate: 0,
      grossValue: budget.grossValue,
      label,
      lastActivityAt,
      negotiationBudgetCount: statusCategory === "negotiation" ? 1 : 0,
      negotiationGrossValue:
        statusCategory === "negotiation" ? budget.grossValue : 0,
      stalledBudgetCount: isStalled ? 1 : 0,
      wonBudgetCount: statusCategory === "won" ? 1 : 0,
    });
    return currentMap;
  }, new Map<string, DashboardEstimatorSummary>());
  const estimatorSummaries = Array.from(estimatorSummaryMap.values())
    .map((item) => ({
      ...item,
      averageTicket:
        item.budgetCount === 0 ? 0 : item.grossValue / item.budgetCount,
      conversionRate:
        item.budgetCount === 0
          ? 0
          : (item.wonBudgetCount / item.budgetCount) * 100,
    }))
    .sort((firstItem, secondItem) =>
      firstItem.label.localeCompare(secondItem.label),
    );

  const staleBudgets = filteredBudgetItems
    .map<DashboardStaleBudgetItem | null>((budget) => {
      const statusCategory = getStatusCategory(budget.statusName);
      if (statusCategory !== "negotiation") {
        return null;
      }

      const lastActivityDate = getLastActivityDate(budget);
      if (lastActivityDate === null) {
        return null;
      }

      const stalledDays = getDaysSince(lastActivityDate, now);
      if (stalledDays < 7) {
        return null;
      }

      return {
        budgetNumber: budget.budgetNumber,
        constructionCompanyLabel: normalizeDisplayValue(
          budget.constructionCompany,
          "Construtora nao informada",
        ),
        grossValue: budget.grossValue,
        id: budget.id,
        lastActivityAt: lastActivityDate.toISOString(),
        projectLabel: normalizeDisplayValue(
          budget.projectName,
          "Sem obra vinculada",
        ),
        salespersonLabel: normalizeDisplayValue(
          budget.salespersonName,
          "Sem vendedor",
        ),
        stalledDays,
        statusLabel: normalizeDisplayValue(
          budget.statusName,
          "Status nao informado",
        ),
      };
    })
    .filter((item): item is DashboardStaleBudgetItem => item !== null)
    .sort((firstItem, secondItem) => {
      if (secondItem.stalledDays !== firstItem.stalledDays) {
        return secondItem.stalledDays - firstItem.stalledDays;
      }
      return secondItem.grossValue - firstItem.grossValue;
    })
    .slice(0, 10);

  const monthlyEvolution = Array.from(
    filteredBudgetItems
      .reduce<Map<string, DashboardMonthlyEvolutionItem>>(
        (currentMap, budget) => {
          const sentAtDate = new Date(budget.sentAt);
          if (Number.isNaN(sentAtDate.getTime())) {
            return currentMap;
          }

          const monthKey = getMonthKey(sentAtDate);
          const statusCategory = getStatusCategory(budget.statusName);
          const existingItem = currentMap.get(monthKey);

          if (existingItem) {
            existingItem.budgetCount += 1;
            existingItem.grossValue += budget.grossValue;
            if (statusCategory === "won") {
              existingItem.wonBudgetCount += 1;
              existingItem.wonGrossValue += budget.grossValue;
            }
            return currentMap;
          }

          currentMap.set(monthKey, {
            budgetCount: 1,
            grossValue: budget.grossValue,
            monthKey,
            monthLabel: getMonthLabel(sentAtDate),
            wonBudgetCount: statusCategory === "won" ? 1 : 0,
            wonGrossValue: statusCategory === "won" ? budget.grossValue : 0,
          });
          return currentMap;
        },
        new Map<string, DashboardMonthlyEvolutionItem>(),
      )
      .values(),
  )
    .sort((firstItem, secondItem) =>
      firstItem.monthKey.localeCompare(secondItem.monthKey),
    )
    .slice(-12);
  const efficiencyBase =
    getComparableSalespersonSummaries(salespersonSummaries);
  const constructionCompanyPerformance = buildEntityPerformanceFromBudgetItems(
    filteredBudgetItems,
    (budget) =>
      normalizeDisplayValue(
        budget.constructionCompany,
        "Construtora nao informada",
      ),
  );
  const projectPerformance = buildEntityPerformanceFromBudgetItems(
    filteredBudgetItems,
    (budget) => normalizeDisplayValue(budget.projectName, "Sem obra vinculada"),
  );
  const topLossReasons =
    buildLossReasonSummariesFromBudgetItems(filteredBudgetItems);
  const averageClosingTimes =
    buildClosingTimeSummariesFromBudgetItems(filteredBudgetItems);
  const technicalEfficiencyBase =
    getComparableEstimatorSummaries(estimatorSummaries);
  const budgetsWithEstimator = estimatorSummaries.reduce(
    (currentTotal, item) => currentTotal + item.budgetCount,
    0,
  );
  const technicalGrossValue = estimatorSummaries.reduce(
    (currentTotal, item) => currentTotal + item.grossValue,
    0,
  );
  const technicalNegotiationGrossValue = estimatorSummaries.reduce(
    (currentTotal, item) => currentTotal + item.negotiationGrossValue,
    0,
  );
  const technicalWonBudgets = estimatorSummaries.reduce(
    (currentTotal, item) => currentTotal + item.wonBudgetCount,
    0,
  );
  const technicalNegotiationBudgets = estimatorSummaries.reduce(
    (currentTotal, item) => currentTotal + item.negotiationBudgetCount,
    0,
  );
  const technicalStalledBudgetsCount = estimatorSummaries.reduce(
    (currentTotal, item) => currentTotal + item.stalledBudgetCount,
    0,
  );
  const budgetsWithoutEstimator = Math.max(
    0,
    totalBudgets - budgetsWithEstimator,
  );
  const technicalLostBudgets = Math.max(
    0,
    budgetsWithEstimator - technicalNegotiationBudgets - technicalWonBudgets,
  );
  const technicalAverageTicket =
    budgetsWithEstimator === 0 ? 0 : technicalGrossValue / budgetsWithEstimator;
  const technicalCoverageRate =
    totalBudgets === 0 ? 0 : (budgetsWithEstimator / totalBudgets) * 100;
  const technicalConversionRate =
    budgetsWithEstimator === 0
      ? 0
      : (technicalWonBudgets / budgetsWithEstimator) * 100;

  return {
    averageClosingTimes,
    availableYears,
    monthlyEvolution,
    negotiationPipeline: sortSalespersonSummaries(
      salespersonSummaries.filter((item) => item.negotiationBudgetCount > 0),
      (firstItem, secondItem) => {
        if (
          secondItem.negotiationGrossValue !== firstItem.negotiationGrossValue
        ) {
          return (
            secondItem.negotiationGrossValue - firstItem.negotiationGrossValue
          );
        }
        return (
          secondItem.negotiationBudgetCount - firstItem.negotiationBudgetCount
        );
      },
    ),
    recentSalespeople: sortSalespersonSummaries(
      salespersonSummaries.filter((item) => item.lastActivityAt !== null),
      (firstItem, secondItem) =>
        new Date(secondItem.lastActivityAt ?? 0).getTime() -
        new Date(firstItem.lastActivityAt ?? 0).getTime(),
    ),
    salespersonFunnel: [...salespersonSummaries]
      .map<DashboardSalespersonFunnelItem>((item) => ({
        conversionRate: item.conversionRate,
        label: item.label,
        lostBudgets: Math.max(
          0,
          item.budgetCount - item.negotiationBudgetCount - item.wonBudgetCount,
        ),
        negotiationBudgets: item.negotiationBudgetCount,
        totalBudgets: item.budgetCount,
        wonBudgets: item.wonBudgetCount,
      }))
      .sort((firstItem, secondItem) => {
        if (secondItem.totalBudgets !== firstItem.totalBudgets) {
          return secondItem.totalBudgets - firstItem.totalBudgets;
        }
        return secondItem.wonBudgets - firstItem.wonBudgets;
      })
      .slice(0, 10),
    staleBudgets,
    summary: {
      activeSalespeople: salespersonSummaries.length,
      averageTicket,
      conversionRate,
      lostBudgets,
      negotiationBudgets,
      stalledBudgetsCount: staleBudgets.length,
      totalBudgets,
      totalGrossValue,
      totalNegotiationGrossValue,
      valueConversionRate,
      wonBudgets,
    },
    technicalOverview: {
      recentEstimators: sortEstimatorSummaries(
        estimatorSummaries.filter((item) => item.lastActivityAt !== null),
        (firstItem, secondItem) =>
          new Date(secondItem.lastActivityAt ?? 0).getTime() -
          new Date(firstItem.lastActivityAt ?? 0).getTime(),
      ),
      summary: {
        activeEstimators: estimatorSummaries.length,
        averageTicket: technicalAverageTicket,
        budgetsWithEstimator,
        budgetsWithoutEstimator,
        conversionRate: technicalConversionRate,
        coverageRate: technicalCoverageRate,
        lostBudgets: technicalLostBudgets,
        negotiationBudgets: technicalNegotiationBudgets,
        stalledBudgetsCount: technicalStalledBudgetsCount,
        totalGrossValue: technicalGrossValue,
        totalNegotiationGrossValue: technicalNegotiationGrossValue,
        wonBudgets: technicalWonBudgets,
      },
      topEstimatorsByAverageTicket: sortEstimatorSummaries(
        technicalEfficiencyBase,
        (firstItem, secondItem) => {
          if (secondItem.averageTicket !== firstItem.averageTicket) {
            return secondItem.averageTicket - firstItem.averageTicket;
          }
          return secondItem.grossValue - firstItem.grossValue;
        },
      ),
      topEstimatorsByBudgetCount: sortEstimatorSummaries(
        estimatorSummaries,
        (firstItem, secondItem) => {
          if (secondItem.budgetCount !== firstItem.budgetCount) {
            return secondItem.budgetCount - firstItem.budgetCount;
          }
          return secondItem.grossValue - firstItem.grossValue;
        },
      ),
      topEstimatorsByValue: sortEstimatorSummaries(
        estimatorSummaries,
        (firstItem, secondItem) => {
          if (secondItem.grossValue !== firstItem.grossValue) {
            return secondItem.grossValue - firstItem.grossValue;
          }
          return secondItem.budgetCount - firstItem.budgetCount;
        },
      ),
    },
    topConstructionCompanies: sortEntityPerformance(
      constructionCompanyPerformance,
      (firstItem, secondItem) => {
        if (secondItem.grossValue !== firstItem.grossValue) {
          return secondItem.grossValue - firstItem.grossValue;
        }
        return secondItem.budgetCount - firstItem.budgetCount;
      },
    ),
    topLossReasons,
    topProjects: sortEntityPerformance(
      projectPerformance,
      (firstItem, secondItem) => {
        if (secondItem.grossValue !== firstItem.grossValue) {
          return secondItem.grossValue - firstItem.grossValue;
        }
        return secondItem.budgetCount - firstItem.budgetCount;
      },
    ),
    topSalespeopleByAverageTicket: sortSalespersonSummaries(
      efficiencyBase,
      (firstItem, secondItem) => {
        if (secondItem.averageTicket !== firstItem.averageTicket) {
          return secondItem.averageTicket - firstItem.averageTicket;
        }
        return secondItem.grossValue - firstItem.grossValue;
      },
    ),
    topSalespeopleByBudgetCount: sortSalespersonSummaries(
      salespersonSummaries,
      (firstItem, secondItem) => {
        if (secondItem.budgetCount !== firstItem.budgetCount) {
          return secondItem.budgetCount - firstItem.budgetCount;
        }
        return secondItem.grossValue - firstItem.grossValue;
      },
    ),
    topSalespeopleByValue: sortSalespersonSummaries(
      salespersonSummaries,
      (firstItem, secondItem) => {
        if (secondItem.grossValue !== firstItem.grossValue) {
          return secondItem.grossValue - firstItem.grossValue;
        }
        return secondItem.budgetCount - firstItem.budgetCount;
      },
    ),
    topSalespeopleByConversion: sortSalespersonSummaries(
      efficiencyBase,
      (firstItem, secondItem) => {
        if (secondItem.conversionRate !== firstItem.conversionRate) {
          return secondItem.conversionRate - firstItem.conversionRate;
        }
        if (secondItem.wonBudgetCount !== firstItem.wonBudgetCount) {
          return secondItem.wonBudgetCount - firstItem.wonBudgetCount;
        }
        return secondItem.budgetCount - firstItem.budgetCount;
      },
    ),
  };
}

export async function getSalespeopleDashboardRequest(
  filters: DashboardSalespeopleFilters,
): Promise<DashboardSalespeopleData> {
  const params = buildDashboardParams(filters);

  try {
    const response = await api.get<DashboardSalespeopleApiResponse>(
      "/dashboard/salespeople",
      {
        params,
      },
    );

    return {
      availableYears: (response.data.available_years ?? []).map((item) =>
        String(item),
      ),
      monthlyEvolution: (response.data.monthly_evolution ?? []).map(
        mapMonthlyEvolution,
      ),
      negotiationPipeline: (response.data.negotiation_pipeline ?? []).map(
        mapSalespersonSummary,
      ),
      recentSalespeople: (response.data.recent_salespeople ?? []).map(
        mapSalespersonSummary,
      ),
      salespersonFunnel: (response.data.salesperson_funnel ?? []).map(
        mapSalespersonFunnel,
      ),
      averageClosingTimes: (response.data.average_closing_times ?? []).map(
        mapClosingTimeSummary,
      ),
      staleBudgets: (response.data.stale_budgets ?? []).map(mapStaleBudget),
      summary: mapSummary(response.data.summary),
      technicalOverview: mapTechnicalOverview(response.data.technical_overview),
      topConstructionCompanies: (
        response.data.top_construction_companies ?? []
      ).map(mapEntityPerformance),
      topLossReasons: (response.data.top_loss_reasons ?? []).map(
        mapLossReasonSummary,
      ),
      topProjects: (response.data.top_projects ?? []).map(mapEntityPerformance),
      topSalespeopleByAverageTicket: (
        response.data.top_salespeople_by_average_ticket ?? []
      ).map(mapSalespersonSummary),
      topSalespeopleByBudgetCount: (
        response.data.top_salespeople_by_budget_count ?? []
      ).map(mapSalespersonSummary),
      topSalespeopleByConversion: (
        response.data.top_salespeople_by_conversion ?? []
      ).map(mapSalespersonSummary),
      topSalespeopleByValue: (response.data.top_salespeople_by_value ?? []).map(
        mapSalespersonSummary,
      ),
    };
  } catch (error) {
    const axiosError = error as AxiosError;

    if (axiosError.response?.status === 404) {
      const legacyFilters = buildLegacyDashboardBudgetFilters(filters);
      const legacyResponse = await getBudgetProjectListRequest(legacyFilters);
      const fallbackData = buildDashboardFromBudgetItems(
        filters,
        legacyResponse.items,
      );

      return fallbackData;
    }

    throw error;
  }
}
