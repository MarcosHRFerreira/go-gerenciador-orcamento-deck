import ApartmentRoundedIcon from "@mui/icons-material/ApartmentRounded";
import AttachMoneyRoundedIcon from "@mui/icons-material/AttachMoneyRounded";
import AssignmentTurnedInRoundedIcon from "@mui/icons-material/AssignmentTurnedInRounded";
import DescriptionRoundedIcon from "@mui/icons-material/DescriptionRounded";
import InsightsRoundedIcon from "@mui/icons-material/InsightsRounded";
import TrendingUpRoundedIcon from "@mui/icons-material/TrendingUpRounded";
import {
  Alert,
  Box,
  Chip,
  CircularProgress,
  Divider,
  LinearProgress,
  MenuItem,
  TextField,
  Typography,
} from "@mui/material";
import { useQuery } from "@tanstack/react-query";
import { useEffect, useMemo, useState } from "react";
import { PageHeader } from "../../../components/common/PageHeader";
import { SectionCard } from "../../../components/common/SectionCard";
import { getBudgetProjectListRequest } from "../../budgets/api/budgets";
import type {
  BudgetListFilters,
  BudgetListItem,
} from "../../budgets/types/budget";

const currencyFormatter = new Intl.NumberFormat("pt-BR", {
  style: "currency",
  currency: "BRL",
});

const dateFormatter = new Intl.DateTimeFormat("pt-BR", {
  dateStyle: "short",
});

type DashboardCompanyFilter = "" | "Rocktec" | "Trox";

type DashboardMetricCard = {
  key: string;
  label: string;
  value: string;
  helper: string;
  icon: typeof TrendingUpRoundedIcon;
};

type DashboardBreakdownItem = {
  label: string;
  budgetCount: number;
  grossValue: number;
};

type DashboardRankingItem = {
  label: string;
  budgetCount: number;
  grossValue: number;
};

type DashboardMonthlyItem = {
  key: string;
  label: string;
  budgetCount: number;
  grossValue: number;
};

const dashboardQueryFiltersBase: BudgetListFilters = {
  budgetNumber: "",
  sourceCompany: "",
  yearBudget: "",
  statusId: "",
  installerId: "",
  projectName: "",
  salespersonId: "",
  page: 1,
  pageSize: 100,
  sortBy: "updated_at",
  sortOrder: "desc",
};

const monthOptions = [
  { value: "1", label: "Janeiro" },
  { value: "2", label: "Fevereiro" },
  { value: "3", label: "Marco" },
  { value: "4", label: "Abril" },
  { value: "5", label: "Maio" },
  { value: "6", label: "Junho" },
  { value: "7", label: "Julho" },
  { value: "8", label: "Agosto" },
  { value: "9", label: "Setembro" },
  { value: "10", label: "Outubro" },
  { value: "11", label: "Novembro" },
  { value: "12", label: "Dezembro" },
] as const;

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
  if (
    normalizedValue === "nao informado" ||
    normalizedValue === "não informado" ||
    normalizedValue === "-"
  ) {
    return fallback;
  }

  return trimmedValue;
}

function getStatusCategory(statusName: string | null) {
  const normalizedStatusName = normalizeText(statusName);

  if (normalizedStatusName === "pedido") {
    return "pedido";
  }
  if (normalizedStatusName === "cancelado") {
    return "cancelado";
  }
  if (normalizedStatusName === "orcamento") {
    return "orcamento";
  }

  return "other";
}

function createRankingItems(
  budgets: BudgetListItem[],
  selector: (budget: BudgetListItem) => string,
  fallback: string,
) {
  const groupedItems = budgets.reduce<Map<string, DashboardRankingItem>>(
    (currentMap, budget) => {
      const label = normalizeDisplayValue(selector(budget), fallback);
      const existingItem = currentMap.get(label);

      if (existingItem) {
        existingItem.budgetCount += 1;
        existingItem.grossValue += budget.grossValue;
        return currentMap;
      }

      currentMap.set(label, {
        label,
        budgetCount: 1,
        grossValue: budget.grossValue,
      });
      return currentMap;
    },
    new Map<string, DashboardRankingItem>(),
  );

  return Array.from(groupedItems.values())
    .sort((firstItem, secondItem) => {
      if (secondItem.budgetCount !== firstItem.budgetCount) {
        return secondItem.budgetCount - firstItem.budgetCount;
      }

      if (secondItem.grossValue !== firstItem.grossValue) {
        return secondItem.grossValue - firstItem.grossValue;
      }

      return firstItem.label.localeCompare(secondItem.label);
    })
    .slice(0, 5);
}

function formatPercentage(value: number) {
  return `${value.toFixed(1)}%`;
}

function formatCompactCurrency(value: number) {
  if (value >= 1_000_000) {
    return `R$ ${(value / 1_000_000).toFixed(1)} mi`;
  }

  if (value >= 1_000) {
    return `R$ ${(value / 1_000).toFixed(1)} mil`;
  }

  return currencyFormatter.format(value);
}

function buildDashboardScopeLabel(
  sourceCompany: DashboardCompanyFilter,
  selectedYear: string,
  selectedMonth: string,
) {
  const companyLabel = sourceCompany || "todas as empresas";
  const yearLabel = selectedYear || "todos os anos";
  const monthLabel =
    monthOptions.find((item) => item.value === selectedMonth)?.label ??
    "todos os meses";

  return `${companyLabel}, ${yearLabel} e ${monthLabel}`;
}

export function DashboardPage() {
  const [sourceCompany, setSourceCompany] =
    useState<DashboardCompanyFilter>("");
  const [selectedYear, setSelectedYear] = useState("");
  const [selectedMonth, setSelectedMonth] = useState("");

  const dashboardQueryFilters = useMemo<BudgetListFilters>(
    () => ({
      ...dashboardQueryFiltersBase,
      sourceCompany,
    }),
    [sourceCompany],
  );

  const budgetsQuery = useQuery({
    queryKey: ["dashboard", "budgets", dashboardQueryFilters],
    queryFn: () => getBudgetProjectListRequest(dashboardQueryFilters),
  });

  const fetchedBudgetItems = useMemo(
    () => budgetsQuery.data?.items ?? [],
    [budgetsQuery.data?.items],
  );

  const availableYears = useMemo(() => {
    const years = fetchedBudgetItems.reduce<Set<string>>(
      (currentYears, budget) => {
        const sentAtDate = new Date(budget.sentAt);
        if (!Number.isNaN(sentAtDate.getTime())) {
          currentYears.add(String(sentAtDate.getFullYear()));
        }
        return currentYears;
      },
      new Set<string>(),
    );

    return Array.from(years).sort((firstYear, secondYear) =>
      secondYear.localeCompare(firstYear),
    );
  }, [fetchedBudgetItems]);

  useEffect(() => {
    if (selectedYear && !availableYears.includes(selectedYear)) {
      setSelectedYear("");
      setSelectedMonth("");
    }
  }, [availableYears, selectedYear]);

  const budgetItems = useMemo(() => {
    return fetchedBudgetItems.filter((budget) => {
      const sentAtDate = new Date(budget.sentAt);
      if (Number.isNaN(sentAtDate.getTime())) {
        return false;
      }

      const budgetYear = String(sentAtDate.getFullYear());
      const budgetMonth = String(sentAtDate.getMonth() + 1);

      if (selectedYear && budgetYear !== selectedYear) {
        return false;
      }

      if (selectedMonth && budgetMonth !== selectedMonth) {
        return false;
      }

      return true;
    });
  }, [fetchedBudgetItems, selectedMonth, selectedYear]);

  const dashboardData = useMemo(() => {
    const totalBudgets = budgetItems.length;
    const totalGrossValue = budgetItems.reduce(
      (currentTotal, budget) => currentTotal + budget.grossValue,
      0,
    );
    const pedidoCount = budgetItems.filter(
      (budget) => getStatusCategory(budget.statusName) === "pedido",
    ).length;
    const canceladoCount = budgetItems.filter(
      (budget) => getStatusCategory(budget.statusName) === "cancelado",
    ).length;
    const orcamentoCount = budgetItems.filter(
      (budget) => getStatusCategory(budget.statusName) === "orcamento",
    ).length;
    const activeCount = totalBudgets - canceladoCount;
    const averageTicket =
      totalBudgets === 0 ? 0 : totalGrossValue / totalBudgets;
    const conversionRate =
      totalBudgets === 0 ? 0 : (pedidoCount / totalBudgets) * 100;

    const companyBreakdownMap = budgetItems.reduce<
      Map<string, DashboardBreakdownItem>
    >((currentMap, budget) => {
      const label = normalizeDisplayValue(
        budget.sourceCompany,
        "Origem nao informada",
      );
      const existingItem = currentMap.get(label);

      if (existingItem) {
        existingItem.budgetCount += 1;
        existingItem.grossValue += budget.grossValue;
        return currentMap;
      }

      currentMap.set(label, {
        label,
        budgetCount: 1,
        grossValue: budget.grossValue,
      });
      return currentMap;
    }, new Map<string, DashboardBreakdownItem>());

    const companyBreakdown = Array.from(companyBreakdownMap.values()).sort(
      (firstItem, secondItem) => secondItem.budgetCount - firstItem.budgetCount,
    );

    const monthlyEvolutionMap = budgetItems.reduce<
      Map<string, DashboardMonthlyItem>
    >((currentMap, budget) => {
      const date = new Date(budget.sentAt);
      const key = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}`;
      const label = date.toLocaleDateString("pt-BR", {
        month: "short",
        year: "2-digit",
      });
      const existingItem = currentMap.get(key);

      if (existingItem) {
        existingItem.budgetCount += 1;
        existingItem.grossValue += budget.grossValue;
        return currentMap;
      }

      currentMap.set(key, {
        key,
        label,
        budgetCount: 1,
        grossValue: budget.grossValue,
      });
      return currentMap;
    }, new Map<string, DashboardMonthlyItem>());

    const monthlyEvolution = Array.from(monthlyEvolutionMap.values())
      .sort((firstItem, secondItem) =>
        firstItem.key.localeCompare(secondItem.key),
      )
      .slice(-6);

    const topSalespeople = createRankingItems(
      budgetItems,
      (budget) => budget.salespersonName ?? "",
      "Sem vendedor",
    );
    const topProjects = createRankingItems(
      budgetItems,
      (budget) => budget.projectName ?? "",
      "Sem obra vinculada",
    );

    const recentBudgets = [...budgetItems]
      .sort(
        (firstItem, secondItem) =>
          new Date(secondItem.updatedAt).getTime() -
          new Date(firstItem.updatedAt).getTime(),
      )
      .slice(0, 6);

    const metricCards: DashboardMetricCard[] = [
      {
        key: "total-budgets",
        label: "Orcamentos monitorados",
        value: String(totalBudgets),
        helper: "Volume total da empresa ou visao consolidada",
        icon: DescriptionRoundedIcon,
      },
      {
        key: "gross-value",
        label: "Valor bruto total",
        value: formatCompactCurrency(totalGrossValue),
        helper: "Soma do valor bruto da carteira atual",
        icon: AttachMoneyRoundedIcon,
      },
      {
        key: "ticket-medio",
        label: "Ticket medio",
        value: formatCompactCurrency(averageTicket),
        helper: "Media de valor por orcamento",
        icon: TrendingUpRoundedIcon,
      },
      {
        key: "pedido-conversion",
        label: "Conversao em pedido",
        value: formatPercentage(conversionRate),
        helper: `${pedidoCount} pedido(s) em ${totalBudgets} orcamento(s)`,
        icon: AssignmentTurnedInRoundedIcon,
      },
    ];

    return {
      activeCount,
      canceladoCount,
      companyBreakdown,
      conversionRate,
      metricCards,
      monthlyEvolution,
      orcamentoCount,
      pedidoCount,
      recentBudgets,
      topProjects,
      topSalespeople,
    };
  }, [budgetItems]);

  return (
    <Box sx={{ display: "flex", flexDirection: "column", gap: 3 }}>
      <PageHeader
        action={
          <Box
            sx={{
              display: "grid",
              gap: 1.5,
              gridTemplateColumns: {
                lg: "repeat(3, minmax(180px, 220px))",
                sm: "repeat(2, minmax(180px, 220px))",
                xs: "minmax(0, 1fr)",
              },
              width: { md: "auto", xs: "100%" },
            }}
          >
            <TextField
              label="Empresa"
              onChange={(event) =>
                setSourceCompany(event.target.value as DashboardCompanyFilter)
              }
              select
              size="small"
              value={sourceCompany}
            >
              <MenuItem value="">Todas as empresas</MenuItem>
              <MenuItem value="Rocktec">ROCKTEC</MenuItem>
              <MenuItem value="Trox">TROX</MenuItem>
            </TextField>
            <TextField
              label="Ano"
              onChange={(event) => setSelectedYear(event.target.value)}
              select
              size="small"
              value={selectedYear}
            >
              <MenuItem value="">Todos os anos</MenuItem>
              {availableYears.map((year) => (
                <MenuItem key={year} value={year}>
                  {year}
                </MenuItem>
              ))}
            </TextField>
            <TextField
              label="Mes"
              onChange={(event) => setSelectedMonth(event.target.value)}
              select
              size="small"
              value={selectedMonth}
            >
              <MenuItem value="">Todos os meses</MenuItem>
              {monthOptions.map((month) => (
                <MenuItem key={month.value} value={month.value}>
                  {month.label}
                </MenuItem>
              ))}
            </TextField>
          </Box>
        }
        description="Painel multiempresa com leitura por empresa, ano e mes usando os dados reais da carteira de orcamentos."
        title="Dashboard"
      />

      {budgetsQuery.isLoading ? <LinearProgress /> : null}

      {budgetsQuery.isError ? (
        <Alert severity="error" variant="outlined">
          Nao foi possivel carregar os dados do dashboard.
        </Alert>
      ) : null}

      {!budgetsQuery.isLoading && !budgetsQuery.isError ? (
        <Alert severity="info" variant="outlined">
          {`Os indicadores abaixo refletem o recorte de ${buildDashboardScopeLabel(
            sourceCompany,
            selectedYear,
            selectedMonth,
          )}.`}
        </Alert>
      ) : null}

      <Box
        sx={{
          display: "grid",
          gap: 3,
          gridTemplateColumns: {
            lg: "repeat(4, minmax(0, 1fr))",
            md: "repeat(2, minmax(0, 1fr))",
            xs: "minmax(0, 1fr)",
          },
        }}
      >
        {(dashboardData?.metricCards ?? []).map((metric) => (
          <Box key={metric.label}>
            <SectionCard>
              <Box sx={{ display: "flex", flexDirection: "column", gap: 1.5 }}>
                <Chip
                  color="primary"
                  icon={<metric.icon />}
                  label="Resumo"
                  sx={{ alignSelf: "flex-start" }}
                  variant="outlined"
                />
                <Typography color="text.secondary" variant="body2">
                  {metric.label}
                </Typography>
                <Typography variant="h3">{metric.value}</Typography>
                <Typography color="text.secondary" variant="body2">
                  {metric.helper}
                </Typography>
              </Box>
            </SectionCard>
          </Box>
        ))}
      </Box>

      <Box
        sx={{
          display: "grid",
          gap: 3,
          gridTemplateColumns: {
            lg: "minmax(0, 1.25fr) minmax(0, 0.75fr)",
            xs: "minmax(0, 1fr)",
          },
        }}
      >
        <SectionCard
          description="Distribuicao da carteira por empresa dentro do filtro atual."
          title="Comparativo por empresa"
        >
          {dashboardData?.companyBreakdown.length ? (
            <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
              {dashboardData.companyBreakdown.map((item) => (
                <Box
                  key={item.label}
                  sx={{
                    alignItems: "center",
                    display: "grid",
                    gap: 1.5,
                    gridTemplateColumns: {
                      md: "minmax(180px, 220px) minmax(0, 1fr) auto auto",
                      xs: "minmax(0, 1fr)",
                    },
                  }}
                >
                  <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
                    <ApartmentRoundedIcon color="action" fontSize="small" />
                    <Typography sx={{ fontWeight: 700 }} variant="body2">
                      {item.label}
                    </Typography>
                  </Box>
                  <Box
                    sx={{
                      bgcolor: "action.hover",
                      borderRadius: 999,
                      height: 10,
                      overflow: "hidden",
                    }}
                  >
                    <Box
                      sx={{
                        bgcolor: "primary.main",
                        height: "100%",
                        width: `${Math.max(
                          dashboardData.metricCards[0]?.value === "0"
                            ? 0
                            : (item.budgetCount /
                                Number(dashboardData.metricCards[0]?.value)) *
                                100,
                          4,
                        )}%`,
                      }}
                    />
                  </Box>
                  <Typography variant="body2">
                    {item.budgetCount} orcamento(s)
                  </Typography>
                  <Typography
                    sx={{ fontWeight: 700, whiteSpace: "nowrap" }}
                    variant="body2"
                  >
                    {formatCompactCurrency(item.grossValue)}
                  </Typography>
                </Box>
              ))}
            </Box>
          ) : (
            <Alert severity="info" variant="outlined">
              Nenhum orcamento encontrado para o filtro atual.
            </Alert>
          )}
        </SectionCard>

        <SectionCard
          description="Leitura rapida do funil atual no escopo selecionado."
          title="Status da carteira"
        >
          <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
            {[
              {
                color: "success.main",
                label: "Pedidos",
                value: dashboardData?.pedidoCount ?? 0,
              },
              {
                color: "primary.main",
                label: "Orcamentos",
                value: dashboardData?.orcamentoCount ?? 0,
              },
              {
                color: "text.secondary",
                label: "Cancelados",
                value: dashboardData?.canceladoCount ?? 0,
              },
              {
                color: "warning.main",
                label: "Ativos",
                value: dashboardData?.activeCount ?? 0,
              },
            ].map((item) => (
              <Box
                key={item.label}
                sx={{
                  alignItems: "center",
                  display: "flex",
                  justifyContent: "space-between",
                }}
              >
                <Typography color="text.secondary" variant="body2">
                  {item.label}
                </Typography>
                <Chip
                  label={String(item.value)}
                  size="small"
                  sx={{
                    bgcolor: item.color,
                    color: "common.white",
                    fontWeight: 700,
                  }}
                />
              </Box>
            ))}
          </Box>
        </SectionCard>
      </Box>

      <Box
        sx={{
          display: "grid",
          gap: 3,
          gridTemplateColumns: {
            lg: "repeat(3, minmax(0, 1fr))",
            xs: "minmax(0, 1fr)",
          },
        }}
      >
        <SectionCard
          description="Top vendedores por quantidade de orcamentos e valor bruto."
          title="Ranking de vendedores"
        >
          <Box sx={{ display: "flex", flexDirection: "column", gap: 1.5 }}>
            {(dashboardData?.topSalespeople ?? []).map((item, index) => (
              <Box key={item.label}>
                <Box
                  sx={{
                    alignItems: "center",
                    display: "flex",
                    justifyContent: "space-between",
                  }}
                >
                  <Typography sx={{ fontWeight: 700 }} variant="body2">
                    {`${index + 1}. ${item.label}`}
                  </Typography>
                  <Typography color="text.secondary" variant="caption">
                    {item.budgetCount} orc.
                  </Typography>
                </Box>
                <Typography color="text.secondary" variant="caption">
                  {currencyFormatter.format(item.grossValue)}
                </Typography>
              </Box>
            ))}
          </Box>
        </SectionCard>

        <SectionCard
          description="Obras com maior volume dentro do recorte atual."
          title="Top obras"
        >
          <Box sx={{ display: "flex", flexDirection: "column", gap: 1.5 }}>
            {(dashboardData?.topProjects ?? []).map((item, index) => (
              <Box key={item.label}>
                <Box
                  sx={{
                    alignItems: "center",
                    display: "flex",
                    justifyContent: "space-between",
                  }}
                >
                  <Typography sx={{ fontWeight: 700 }} variant="body2">
                    {`${index + 1}. ${item.label}`}
                  </Typography>
                  <Typography color="text.secondary" variant="caption">
                    {item.budgetCount} orc.
                  </Typography>
                </Box>
                <Typography color="text.secondary" variant="caption">
                  {currencyFormatter.format(item.grossValue)}
                </Typography>
              </Box>
            ))}
          </Box>
        </SectionCard>

        <SectionCard
          description="Evolucao recente considerando a data de envio do orcamento."
          title="Ultimos 6 meses"
        >
          <Box sx={{ display: "flex", flexDirection: "column", gap: 1.5 }}>
            {(dashboardData?.monthlyEvolution ?? []).map((item) => (
              <Box key={item.key}>
                <Box
                  sx={{
                    alignItems: "center",
                    display: "flex",
                    justifyContent: "space-between",
                  }}
                >
                  <Typography
                    sx={{ textTransform: "capitalize" }}
                    variant="body2"
                  >
                    {item.label}
                  </Typography>
                  <Typography color="text.secondary" variant="caption">
                    {item.budgetCount} orc.
                  </Typography>
                </Box>
                <Typography color="text.secondary" variant="caption">
                  {currencyFormatter.format(item.grossValue)}
                </Typography>
              </Box>
            ))}
          </Box>
        </SectionCard>
      </Box>

      <SectionCard
        description="Ultimos orcamentos atualizados dentro do filtro selecionado."
        title="Movimentacao recente"
      >
        <Box
          sx={{
            display: "flex",
            flexDirection: "column",
            gap: 1.5,
          }}
        >
          {(dashboardData?.recentBudgets ?? []).length ? (
            dashboardData.recentBudgets.map((budget, index) => (
              <Box key={budget.id}>
                {index > 0 ? <Divider sx={{ mb: 1.5 }} /> : null}
                <Box
                  sx={{
                    alignItems: { md: "center", xs: "flex-start" },
                    display: "flex",
                    flexDirection: { md: "row", xs: "column" },
                    gap: 1,
                    justifyContent: "space-between",
                  }}
                >
                  <Box sx={{ minWidth: 0 }}>
                    <Typography sx={{ fontWeight: 700 }} variant="body2">
                      {budget.budgetNumber}
                    </Typography>
                    <Typography color="text.secondary" variant="caption">
                      {`${normalizeDisplayValue(
                        budget.sourceCompany,
                        "Origem nao informada",
                      )} · ${normalizeDisplayValue(
                        budget.projectName,
                        "Sem obra vinculada",
                      )}`}
                    </Typography>
                  </Box>
                  <Box
                    sx={{
                      display: "flex",
                      flexDirection: { sm: "row", xs: "column" },
                      gap: 1,
                    }}
                  >
                    <Chip
                      icon={<InsightsRoundedIcon />}
                      label={budget.statusName ?? "Nao informado"}
                      size="small"
                      variant="outlined"
                    />
                    <Chip
                      label={currencyFormatter.format(budget.grossValue)}
                      size="small"
                      variant="outlined"
                    />
                    <Chip
                      label={`Atualizado em ${dateFormatter.format(
                        new Date(budget.updatedAt),
                      )}`}
                      size="small"
                      variant="outlined"
                    />
                  </Box>
                </Box>
              </Box>
            ))
          ) : budgetsQuery.isLoading ? (
            <Box
              sx={{
                alignItems: "center",
                display: "flex",
                justifyContent: "center",
                minHeight: 120,
              }}
            >
              <CircularProgress size={28} />
            </Box>
          ) : (
            <Alert severity="info" variant="outlined">
              Nenhuma movimentacao encontrada para o filtro atual.
            </Alert>
          )}
        </Box>
      </SectionCard>
    </Box>
  );
}
