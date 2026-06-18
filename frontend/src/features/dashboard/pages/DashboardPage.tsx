import AttachMoneyRoundedIcon from "@mui/icons-material/AttachMoneyRounded";
import AssignmentTurnedInRoundedIcon from "@mui/icons-material/AssignmentTurnedInRounded";
import DescriptionRoundedIcon from "@mui/icons-material/DescriptionRounded";
import DownloadRoundedIcon from "@mui/icons-material/DownloadRounded";
import InsightsRoundedIcon from "@mui/icons-material/InsightsRounded";
import OpenInNewRoundedIcon from "@mui/icons-material/OpenInNewRounded";
import TrendingUpRoundedIcon from "@mui/icons-material/TrendingUpRounded";
import {
  Alert,
  Box,
  Button,
  Chip,
  Divider,
  LinearProgress,
  MenuItem,
  TextField,
  Typography,
} from "@mui/material";
import { useQuery } from "@tanstack/react-query";
import { useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { PageHeader } from "../../../components/common/PageHeader";
import { SectionCard } from "../../../components/common/SectionCard";
import { getBudgetCatalogsRequest } from "../../budgets/api/budgets";
import type { BudgetCatalogItem } from "../../budgets/types/budget";
import { getSalespeopleDashboardRequest } from "../api/dashboard";
import type {
  DashboardMonthlyEvolutionItem,
  DashboardSalespeopleFilters,
  DashboardSalespersonFunnelItem,
  DashboardSalespersonSummary,
  DashboardStaleBudgetItem,
} from "../types/dashboard";

const currencyFormatter = new Intl.NumberFormat("pt-BR", {
  style: "currency",
  currency: "BRL",
});

const dateFormatter = new Intl.DateTimeFormat("pt-BR", {
  dateStyle: "short",
});

const monthYearFormatter = new Intl.DateTimeFormat("pt-BR", {
  month: "short",
  year: "numeric",
});

type DashboardCompanyFilter = "" | "Rocktec" | "Trox";

type DashboardMetricCard = {
  key: string;
  label: string;
  value: string;
  helper: string;
  icon: typeof TrendingUpRoundedIcon;
};

type BudgetListNavigationOptions = {
  budgetNumber?: string;
  salespersonId?: string;
  sourceCompany: DashboardCompanyFilter;
  statusId?: string;
  year: string;
  month: string;
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

function formatDateOrFallback(value: string | null, fallback: string) {
  if (value === null) {
    return fallback;
  }

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return fallback;
  }

  return dateFormatter.format(date);
}

function formatMonthKeyLabel(value: string) {
  const date = new Date(`${value}-01T00:00:00`);
  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return monthYearFormatter.format(date);
}

function formatSignedCurrency(value: number) {
  const absoluteValue = currencyFormatter.format(Math.abs(value));

  if (value > 0) {
    return `+${absoluteValue}`;
  }
  if (value < 0) {
    return `-${absoluteValue}`;
  }

  return absoluteValue;
}

function formatSignedInteger(value: number) {
  if (value > 0) {
    return `+${value}`;
  }

  return String(value);
}

function formatSignedPercentagePoints(value: number) {
  if (value > 0) {
    return `+${value.toFixed(1)} p.p.`;
  }
  if (value < 0) {
    return `${value.toFixed(1)} p.p.`;
  }

  return "0,0 p.p.";
}

function getDeltaColor(value: number) {
  if (value > 0) {
    return "success";
  }
  if (value < 0) {
    return "error";
  }

  return "default";
}

function getMonthDateRange(year: string, month: string) {
  if (!year || !month) {
    return null;
  }

  const normalizedMonth = month.padStart(2, "0");
  const monthStart = new Date(`${year}-${normalizedMonth}-01T00:00:00`);
  if (Number.isNaN(monthStart.getTime())) {
    return null;
  }

  const monthEnd = new Date(
    monthStart.getFullYear(),
    monthStart.getMonth() + 1,
    0,
  );
  const from = `${year}-${normalizedMonth}-01`;
  const to = `${year}-${normalizedMonth}-${String(monthEnd.getDate()).padStart(2, "0")}`;

  return { from, to };
}

function buildBudgetListSearchParams({
  budgetNumber,
  month,
  salespersonId,
  sourceCompany,
  statusId,
  year,
}: BudgetListNavigationOptions) {
  const searchParams = new URLSearchParams();

  if (budgetNumber) {
    searchParams.set("budgetNumber", budgetNumber);
  }
  if (sourceCompany) {
    searchParams.set("sourceCompany", sourceCompany);
  }
  if (year) {
    searchParams.set("yearBudget", year);
  }
  if (salespersonId) {
    searchParams.set("salespersonId", salespersonId);
  }
  if (statusId) {
    searchParams.set("statusId", statusId);
  }

  const monthDateRange = getMonthDateRange(year, month);
  if (monthDateRange !== null) {
    searchParams.set("sentAtFrom", monthDateRange.from);
    searchParams.set("sentAtTo", monthDateRange.to);
  }

  searchParams.set("page", "1");
  searchParams.set("pageSize", "50");
  searchParams.set("sortBy", "updated_at");
  searchParams.set("sortOrder", "desc");

  return searchParams;
}

function escapeCsvValue(value: string) {
  return `"${value.replaceAll('"', '""')}"`;
}

function createCsvLine(values: Array<string | number>) {
  return values.map((value) => escapeCsvValue(String(value))).join(";");
}

function normalizeLookupKey(value: string) {
  return value
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "")
    .trim()
    .toLowerCase();
}

function buildDashboardScopeLabel(
  sourceCompany: DashboardCompanyFilter,
  selectedYear: string,
  selectedMonth: string,
  salespersonLabel: string,
) {
  const companyLabel = sourceCompany || "todas as empresas";
  const yearLabel = selectedYear || "todos os anos";
  const monthLabel =
    monthOptions.find((item) => item.value === selectedMonth)?.label ??
    "todos os meses";
  const salespersonScope = salespersonLabel || "todos os vendedores";

  return `${companyLabel}, ${yearLabel}, ${monthLabel} e ${salespersonScope}`;
}

export function DashboardPage() {
  const navigate = useNavigate();
  const [sourceCompany, setSourceCompany] =
    useState<DashboardCompanyFilter>("");
  const [selectedYear, setSelectedYear] = useState("");
  const [selectedMonth, setSelectedMonth] = useState("");
  const [selectedSalespersonId, setSelectedSalespersonId] = useState("");

  const dashboardFilters = useMemo<DashboardSalespeopleFilters>(
    () => ({
      sourceCompany,
      salespersonId: selectedSalespersonId,
      year: selectedYear,
      month: selectedMonth,
    }),
    [selectedMonth, selectedSalespersonId, selectedYear, sourceCompany],
  );

  const budgetCatalogsQuery = useQuery({
    queryKey: ["dashboard", "budget-catalogs"],
    queryFn: getBudgetCatalogsRequest,
    staleTime: 1000 * 60 * 5,
  });

  const dashboardQuery = useQuery({
    queryKey: ["dashboard", "salespeople", dashboardFilters],
    queryFn: () => getSalespeopleDashboardRequest(dashboardFilters),
  });

  const availableYears = useMemo(() => {
    return dashboardQuery.data?.availableYears ?? [];
  }, [dashboardQuery.data?.availableYears]);

  useEffect(() => {
    if (selectedYear && !availableYears.includes(selectedYear)) {
      setSelectedYear("");
      setSelectedMonth("");
    }
  }, [availableYears, selectedYear]);

  const salespersonOptions = useMemo<BudgetCatalogItem[]>(() => {
    return [...(budgetCatalogsQuery.data?.salespeople ?? [])].sort((a, b) =>
      a.name.localeCompare(b.name),
    );
  }, [budgetCatalogsQuery.data?.salespeople]);

  const selectedSalespersonLabel = useMemo(() => {
    return (
      salespersonOptions.find(
        (item) => String(item.id) === selectedSalespersonId,
      )?.name ?? ""
    );
  }, [selectedSalespersonId, salespersonOptions]);

  const salespersonIdByName = useMemo(() => {
    return salespersonOptions.reduce<Map<string, string>>(
      (currentMap, item) => {
        currentMap.set(item.name, String(item.id));
        return currentMap;
      },
      new Map<string, string>(),
    );
  }, [salespersonOptions]);

  const budgetItems = useMemo(() => {
    return dashboardQuery.data;
  }, [dashboardQuery.data]);

  const statusIdByNormalizedName = useMemo(() => {
    return (budgetCatalogsQuery.data?.statuses ?? []).reduce<
      Map<string, string>
    >((currentMap, item) => {
      currentMap.set(normalizeLookupKey(item.name), String(item.id));
      return currentMap;
    }, new Map<string, string>());
  }, [budgetCatalogsQuery.data?.statuses]);

  const dashboardData = useMemo(() => {
    const summary = budgetItems?.summary;
    const monthlyEvolution = budgetItems?.monthlyEvolution ?? [];
    const maxMonthlyGrossValue = monthlyEvolution.reduce(
      (currentMax, item) => Math.max(currentMax, item.grossValue),
      0,
    );

    const metricCards: DashboardMetricCard[] = [
      {
        key: "active-salespeople",
        label: "Vendedores ativos",
        value: String(summary?.activeSalespeople ?? 0),
        helper: "Quantidade de vendedores com orcamentos no recorte atual",
        icon: InsightsRoundedIcon,
      },
      {
        key: "total-budgets",
        label: "Orcamentos monitorados",
        value: String(summary?.totalBudgets ?? 0),
        helper: "Volume total de orcamentos no recorte atual",
        icon: DescriptionRoundedIcon,
      },
      {
        key: "gross-value",
        label: "Valor bruto total",
        value: formatCompactCurrency(summary?.totalGrossValue ?? 0),
        helper: "Soma do valor bruto da carteira comercial",
        icon: AttachMoneyRoundedIcon,
      },
      {
        key: "ticket-medio",
        label: "Ticket medio",
        value: formatCompactCurrency(summary?.averageTicket ?? 0),
        helper: "Media de valor por orcamento do periodo",
        icon: TrendingUpRoundedIcon,
      },
      {
        key: "negotiation-gross-value",
        label: "Valor em negociacao",
        value: formatCompactCurrency(summary?.totalNegotiationGrossValue ?? 0),
        helper: `${summary?.negotiationBudgets ?? 0} orcamento(s) ainda em carteira`,
        icon: AttachMoneyRoundedIcon,
      },
      {
        key: "pedido-conversion",
        label: "Conversao em pedido",
        value: formatPercentage(summary?.conversionRate ?? 0),
        helper: `${summary?.wonBudgets ?? 0} pedido(s) em ${summary?.totalBudgets ?? 0} orcamento(s)`,
        icon: AssignmentTurnedInRoundedIcon,
      },
      {
        key: "stalled-budgets",
        label: "Orcamentos parados",
        value: String(summary?.stalledBudgetsCount ?? 0),
        helper: "Oportunidades em negociacao sem atividade ha 7 dias ou mais",
        icon: DescriptionRoundedIcon,
      },
    ];

    return {
      conversionRate: summary?.conversionRate ?? 0,
      lostBudgets: summary?.lostBudgets ?? 0,
      maxMonthlyGrossValue,
      metricCards,
      monthlyEvolution,
      negotiationBudgets: summary?.negotiationBudgets ?? 0,
      negotiationPipeline: budgetItems?.negotiationPipeline ?? [],
      recentSalespeople: budgetItems?.recentSalespeople ?? [],
      salespersonFunnel: budgetItems?.salespersonFunnel ?? [],
      staleBudgets: budgetItems?.staleBudgets ?? [],
      topSalespeopleByAverageTicket:
        budgetItems?.topSalespeopleByAverageTicket ?? [],
      topSalespeopleByBudgetCount:
        budgetItems?.topSalespeopleByBudgetCount ?? [],
      topSalespeopleByConversion:
        budgetItems?.topSalespeopleByConversion ?? [],
      topSalespeopleByValue: budgetItems?.topSalespeopleByValue ?? [],
      wonBudgets: summary?.wonBudgets ?? 0,
    };
  }, [budgetItems]);

  const dashboardInsights = useMemo(() => {
    const monthlyEvolution = dashboardData.monthlyEvolution;
    if (monthlyEvolution.length < 2) {
      return {
        monthComparison: null,
      };
    }

    const currentMonth = monthlyEvolution[monthlyEvolution.length - 1];
    const previousMonth = monthlyEvolution[monthlyEvolution.length - 2];
    const currentConversionRate =
      currentMonth.budgetCount === 0
        ? 0
        : (currentMonth.wonBudgetCount / currentMonth.budgetCount) * 100;
    const previousConversionRate =
      previousMonth.budgetCount === 0
        ? 0
        : (previousMonth.wonBudgetCount / previousMonth.budgetCount) * 100;

    return {
      monthComparison: {
        currentMonthLabel: formatMonthKeyLabel(currentMonth.monthKey),
        previousMonthLabel: formatMonthKeyLabel(previousMonth.monthKey),
        currentMonth,
        previousMonth,
        budgetDelta: currentMonth.budgetCount - previousMonth.budgetCount,
        grossValueDelta: currentMonth.grossValue - previousMonth.grossValue,
        conversionDelta: currentConversionRate - previousConversionRate,
      },
    };
  }, [dashboardData.monthlyEvolution]);

  const handleOpenBudgetList = ({
    budgetNumber,
    salespersonId,
    statusId,
  }: Partial<BudgetListNavigationOptions> = {}) => {
    const searchParams = buildBudgetListSearchParams({
      budgetNumber,
      month: selectedMonth,
      salespersonId: salespersonId ?? selectedSalespersonId,
      sourceCompany,
      statusId,
      year: selectedYear,
    });

    navigate(`/budgets?${searchParams.toString()}`);
  };

  const handleExportDashboardCsv = () => {
    const csvLines = [
      createCsvLine([
        "Escopo",
        buildDashboardScopeLabel(
          sourceCompany,
          selectedYear,
          selectedMonth,
          selectedSalespersonLabel,
        ),
      ]),
      "",
      createCsvLine([
        "Resumo",
        "Vendedores ativos",
        "Orcamentos monitorados",
        "Valor bruto total",
        "Ticket medio",
        "Valor em negociacao",
        "Conversao",
        "Orcamentos parados",
      ]),
      createCsvLine([
        "Resumo",
        dashboardData.metricCards[0]?.value ?? "0",
        dashboardData.metricCards[1]?.value ?? "0",
        dashboardData.metricCards[2]?.value ?? "R$ 0,00",
        dashboardData.metricCards[3]?.value ?? "R$ 0,00",
        dashboardData.metricCards[4]?.value ?? "R$ 0,00",
        dashboardData.metricCards[5]?.value ?? "0,0%",
        dashboardData.metricCards[6]?.value ?? "0",
      ]),
      "",
      createCsvLine([
        "Top por valor",
        "Vendedor",
        "Orcamentos",
        "Valor bruto",
        "Ticket medio",
        "Ultima atividade",
      ]),
      ...dashboardData.topSalespeopleByValue.map((item) =>
        createCsvLine([
          "Top por valor",
          item.label,
          item.budgetCount,
          currencyFormatter.format(item.grossValue),
          currencyFormatter.format(item.averageTicket),
          formatDateOrFallback(item.lastActivityAt, "Nao informada"),
        ]),
      ),
      "",
      createCsvLine([
        "Carteira em negociacao",
        "Vendedor",
        "Em aberto",
        "Valor",
        "Parados",
        "Ultima atividade",
      ]),
      ...dashboardData.negotiationPipeline.map((item) =>
        createCsvLine([
          "Carteira em negociacao",
          item.label,
          item.negotiationBudgetCount,
          currencyFormatter.format(item.negotiationGrossValue),
          item.stalledBudgetCount,
          formatDateOrFallback(item.lastActivityAt, "Nao informada"),
        ]),
      ),
      "",
      createCsvLine([
        "Orcamentos parados",
        "Orcamento",
        "Vendedor",
        "Obra",
        "Construtora",
        "Valor",
        "Dias parados",
      ]),
      ...dashboardData.staleBudgets.map((item) =>
        createCsvLine([
          "Orcamentos parados",
          item.budgetNumber,
          item.salespersonLabel,
          item.projectLabel,
          item.constructionCompanyLabel,
          currencyFormatter.format(item.grossValue),
          item.stalledDays,
        ]),
      ),
      "",
      createCsvLine([
        "Evolucao mensal",
        "Mes",
        "Orcamentos",
        "Valor bruto",
        "Pedidos",
        "Valor convertido",
      ]),
      ...dashboardData.monthlyEvolution.map((item) =>
        createCsvLine([
          "Evolucao mensal",
          item.monthLabel,
          item.budgetCount,
          currencyFormatter.format(item.grossValue),
          item.wonBudgetCount,
          currencyFormatter.format(item.wonGrossValue),
        ]),
      ),
    ].join("\n");

    const csvBlob = new Blob([`\uFEFF${csvLines}`], {
      type: "text/csv;charset=utf-8;",
    });
    const downloadUrl = window.URL.createObjectURL(csvBlob);
    const downloadLink = document.createElement("a");
    const fileScope = [
      sourceCompany || "todas-empresas",
      selectedYear || "todos-anos",
      selectedMonth || "todos-meses",
      selectedSalespersonLabel || "todos-vendedores",
    ]
      .join("-")
      .replaceAll(" ", "-")
      .toLowerCase();

    downloadLink.href = downloadUrl;
    downloadLink.download = `dashboard-vendedores-${fileScope}.csv`;
    document.body.appendChild(downloadLink);
    downloadLink.click();
    document.body.removeChild(downloadLink);
    window.URL.revokeObjectURL(downloadUrl);
  };

  return (
    <Box sx={{ display: "flex", flexDirection: "column", gap: 3 }}>
      <PageHeader
        action={
          <Box
            sx={{
              display: "grid",
              gap: 1.5,
              gridTemplateColumns: {
                lg: "repeat(4, minmax(180px, 220px))",
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
            <TextField
              label="Vendedor"
              onChange={(event) => setSelectedSalespersonId(event.target.value)}
              select
              size="small"
              value={selectedSalespersonId}
            >
              <MenuItem value="">Todos os vendedores</MenuItem>
              {salespersonOptions.map((salesperson) => (
                <MenuItem key={salesperson.id} value={String(salesperson.id)}>
                  {salesperson.name}
                </MenuItem>
              ))}
            </TextField>
          </Box>
        }
        description="Painel administrativo com foco em performance comercial por vendedor, carteira em negociacao e oportunidades paradas."
        title="Dashboard administrativo"
      />

      {dashboardQuery.isLoading ? <LinearProgress /> : null}

      {dashboardQuery.isError ? (
        <Alert severity="error" variant="outlined">
          Nao foi possivel carregar os dados do dashboard.
        </Alert>
      ) : null}

      {!dashboardQuery.isLoading && !dashboardQuery.isError ? (
        <Alert severity="info" variant="outlined">
          {`Os indicadores abaixo refletem o recorte de ${buildDashboardScopeLabel(
            sourceCompany,
            selectedYear,
            selectedMonth,
            selectedSalespersonLabel,
          )}.`}
        </Alert>
      ) : null}

      {!dashboardQuery.isLoading && !dashboardQuery.isError ? (
        <SectionCard
          description="Atalhos para aprofundar a analise do dashboard e exportar o recorte atual."
          title="Acoes rapidas"
        >
          <Box
            sx={{
              display: "flex",
              flexWrap: "wrap",
              gap: 1.5,
            }}
          >
            <Button
              onClick={() => handleOpenBudgetList()}
              startIcon={<OpenInNewRoundedIcon />}
              variant="contained"
            >
              Abrir lista filtrada
            </Button>
            <Button
              onClick={() =>
                handleOpenBudgetList({
                  statusId:
                    statusIdByNormalizedName.get(
                      normalizeLookupKey("Em Negociacao"),
                    ) ?? "",
                })
              }
              startIcon={<OpenInNewRoundedIcon />}
              variant="outlined"
            >
              Ver em negociacao
            </Button>
            <Button
              onClick={() =>
                handleOpenBudgetList({
                  statusId:
                    statusIdByNormalizedName.get(
                      normalizeLookupKey("Pedido"),
                    ) ?? "",
                })
              }
              startIcon={<OpenInNewRoundedIcon />}
              variant="outlined"
            >
              Ver pedidos
            </Button>
            <Button
              onClick={() =>
                handleOpenBudgetList({
                  statusId:
                    statusIdByNormalizedName.get(
                      normalizeLookupKey("Cancelado"),
                    ) ?? "",
                })
              }
              startIcon={<OpenInNewRoundedIcon />}
              variant="outlined"
            >
              Ver cancelados
            </Button>
            <Button
              onClick={handleExportDashboardCsv}
              startIcon={<DownloadRoundedIcon />}
              variant="outlined"
            >
              Exportar CSV
            </Button>
          </Box>
        </SectionCard>
      ) : null}

      <Box
        sx={{
          display: "grid",
          gap: 3,
          gridTemplateColumns: {
            lg: "repeat(3, minmax(0, 1fr))",
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
          description="Top 10 vendedores ordenados pelo valor bruto total do recorte atual."
          title="Top 10 por valor"
        >
          {(dashboardData?.topSalespeopleByValue ?? []).length ? (
            <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
              {dashboardData.topSalespeopleByValue.map((item, index) => (
                <Box
                  key={item.label}
                  sx={{
                    display: "flex",
                    flexDirection: "column",
                    gap: 0.5,
                  }}
                >
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
                    <Box sx={{ display: "flex", gap: 1 }}>
                      <Chip
                        label={`${item.budgetCount} orc.`}
                        size="small"
                        variant="outlined"
                      />
                      <Button
                        onClick={() =>
                          handleOpenBudgetList({
                            salespersonId:
                              salespersonIdByName.get(item.label) ?? undefined,
                          })
                        }
                        size="small"
                        startIcon={<OpenInNewRoundedIcon />}
                        variant="text"
                      >
                        Abrir
                      </Button>
                    </Box>
                  </Box>
                  <Box
                    sx={{
                      alignItems: "center",
                      display: "flex",
                      flexWrap: "wrap",
                      gap: 1,
                    }}
                  >
                    <Chip
                      label={currencyFormatter.format(item.grossValue)}
                      size="small"
                      variant="outlined"
                    />
                    <Chip
                      label={`Ticket ${formatCompactCurrency(item.averageTicket)}`}
                      size="small"
                      variant="outlined"
                    />
                    <Chip
                      label={`Ultima atividade ${formatDateOrFallback(item.lastActivityAt, "Nao informada")}`}
                      size="small"
                      variant="outlined"
                    />
                  </Box>
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
          description="Top 10 vendedores ordenados pelo volume de orcamentos."
          title="Top 10 por quantidade"
        >
          {(dashboardData?.topSalespeopleByBudgetCount ?? []).length ? (
            <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
              {dashboardData.topSalespeopleByBudgetCount.map((item, index) => (
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
                    <Chip
                      label={`${item.budgetCount} orc.`}
                      size="small"
                      variant="outlined"
                    />
                  </Box>
                  <Box
                    sx={{
                      alignItems: "center",
                      display: "flex",
                      flexWrap: "wrap",
                      gap: 1,
                      mt: 0.5,
                    }}
                  >
                    <Chip
                      label={currencyFormatter.format(item.grossValue)}
                      size="small"
                      variant="outlined"
                    />
                    <Chip
                      color="success"
                      label={`Conversao ${formatPercentage(item.conversionRate)}`}
                      size="small"
                      variant="outlined"
                    />
                    <Chip
                      label={`Ultima atividade ${formatDateOrFallback(item.lastActivityAt, "Nao informada")}`}
                      size="small"
                      variant="outlined"
                    />
                  </Box>
                </Box>
              ))}
            </Box>
          ) : (
            <Alert severity="info" variant="outlined">
              Nenhum orcamento encontrado para o filtro atual.
            </Alert>
          )}
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
          description="Comparativo entre os dois meses mais recentes do recorte atual para identificar aceleracao ou perda de ritmo."
          title="Tendencia mensal"
        >
          {dashboardInsights.monthComparison !== null ? (
            <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
              <Alert severity="info" variant="outlined">
                {`Comparando ${dashboardInsights.monthComparison.currentMonthLabel} com ${dashboardInsights.monthComparison.previousMonthLabel}.`}
              </Alert>
              <Box
                sx={{
                  display: "grid",
                  gap: 1.5,
                  gridTemplateColumns: {
                    md: "repeat(3, minmax(0, 1fr))",
                    xs: "minmax(0, 1fr)",
                  },
                }}
              >
                <Box
                  sx={{
                    border: (theme) => `1px solid ${theme.palette.divider}`,
                    borderRadius: 2,
                    p: 2,
                  }}
                >
                  <Typography color="text.secondary" variant="body2">
                    Orcamentos
                  </Typography>
                  <Typography sx={{ mt: 0.5 }} variant="h5">
                    {dashboardInsights.monthComparison.currentMonth.budgetCount}
                  </Typography>
                  <Chip
                    color={getDeltaColor(
                      dashboardInsights.monthComparison.budgetDelta,
                    )}
                    label={`${formatSignedInteger(
                      dashboardInsights.monthComparison.budgetDelta,
                    )} vs anterior`}
                    size="small"
                    sx={{ mt: 1 }}
                    variant="outlined"
                  />
                </Box>
                <Box
                  sx={{
                    border: (theme) => `1px solid ${theme.palette.divider}`,
                    borderRadius: 2,
                    p: 2,
                  }}
                >
                  <Typography color="text.secondary" variant="body2">
                    Valor bruto
                  </Typography>
                  <Typography sx={{ mt: 0.5 }} variant="h5">
                    {formatCompactCurrency(
                      dashboardInsights.monthComparison.currentMonth.grossValue,
                    )}
                  </Typography>
                  <Chip
                    color={getDeltaColor(
                      dashboardInsights.monthComparison.grossValueDelta,
                    )}
                    label={`${formatSignedCurrency(
                      dashboardInsights.monthComparison.grossValueDelta,
                    )} vs anterior`}
                    size="small"
                    sx={{ mt: 1 }}
                    variant="outlined"
                  />
                </Box>
                <Box
                  sx={{
                    border: (theme) => `1px solid ${theme.palette.divider}`,
                    borderRadius: 2,
                    p: 2,
                  }}
                >
                  <Typography color="text.secondary" variant="body2">
                    Conversao
                  </Typography>
                  <Typography sx={{ mt: 0.5 }} variant="h5">
                    {formatPercentage(
                      dashboardInsights.monthComparison.currentMonth.budgetCount ===
                        0
                        ? 0
                        : (dashboardInsights.monthComparison.currentMonth
                            .wonBudgetCount /
                            dashboardInsights.monthComparison.currentMonth
                              .budgetCount) *
                            100,
                    )}
                  </Typography>
                  <Chip
                    color={getDeltaColor(
                      dashboardInsights.monthComparison.conversionDelta,
                    )}
                    label={`${formatSignedPercentagePoints(
                      dashboardInsights.monthComparison.conversionDelta,
                    )} vs anterior`}
                    size="small"
                    sx={{ mt: 1 }}
                    variant="outlined"
                  />
                </Box>
              </Box>
            </Box>
          ) : (
            <Alert severity="info" variant="outlined">
              Sao necessarios pelo menos dois meses no recorte atual para montar
              o comparativo de tendencia.
            </Alert>
          )}
        </SectionCard>

        <SectionCard
          description="Vendedores com melhor conversao no periodo, priorizando quem tem pelo menos dois orcamentos no recorte."
          title="Top conversao"
        >
          {(dashboardData?.topSalespeopleByConversion ?? []).length ? (
            <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
              {dashboardData.topSalespeopleByConversion.map((item, index) => (
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
                    <Chip
                      color="success"
                      label={formatPercentage(item.conversionRate)}
                      size="small"
                      variant="outlined"
                    />
                  </Box>
                  <Box
                    sx={{
                      alignItems: "center",
                      display: "flex",
                      flexWrap: "wrap",
                      gap: 1,
                      mt: 0.5,
                    }}
                  >
                    <Chip
                      label={`${item.wonBudgetCount} pedido(s)`}
                      size="small"
                      variant="outlined"
                    />
                    <Chip
                      label={`${item.budgetCount} orc.`}
                      size="small"
                      variant="outlined"
                    />
                    <Chip
                      label={`Ticket ${formatCompactCurrency(item.averageTicket)}`}
                      size="small"
                      variant="outlined"
                    />
                  </Box>
                </Box>
              ))}
            </Box>
          ) : (
            <Alert severity="info" variant="outlined">
              Nenhum vendedor elegivel para o ranking de conversao no filtro atual.
            </Alert>
          )}
        </SectionCard>

        <SectionCard
          description="Vendedores com maior ticket medio no periodo, considerando base minima para reduzir distorcao."
          title="Top ticket medio"
        >
          {(dashboardData?.topSalespeopleByAverageTicket ?? []).length ? (
            <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
              {dashboardData.topSalespeopleByAverageTicket.map((item, index) => (
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
                    <Chip
                      color="primary"
                      label={formatCompactCurrency(item.averageTicket)}
                      size="small"
                      variant="outlined"
                    />
                  </Box>
                  <Box
                    sx={{
                      alignItems: "center",
                      display: "flex",
                      flexWrap: "wrap",
                      gap: 1,
                      mt: 0.5,
                    }}
                  >
                    <Chip
                      label={`${item.budgetCount} orc.`}
                      size="small"
                      variant="outlined"
                    />
                    <Chip
                      label={currencyFormatter.format(item.grossValue)}
                      size="small"
                      variant="outlined"
                    />
                    <Chip
                      label={`Conversao ${formatPercentage(item.conversionRate)}`}
                      size="small"
                      variant="outlined"
                    />
                  </Box>
                </Box>
              ))}
            </Box>
          ) : (
            <Alert severity="info" variant="outlined">
              Nenhum vendedor elegivel para o ranking de ticket medio no filtro atual.
            </Alert>
          )}
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
          description="Carteira ainda em negociacao, com foco no valor e no volume por vendedor."
          title="Carteira em negociacao"
        >
          {(dashboardData?.negotiationPipeline ?? []).length ? (
            <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
              {dashboardData.negotiationPipeline.map((item) => (
                <Box key={item.label}>
                  <Box
                    sx={{
                      alignItems: "center",
                      display: "flex",
                      justifyContent: "space-between",
                    }}
                  >
                    <Typography sx={{ fontWeight: 700 }} variant="body2">
                      {item.label}
                    </Typography>
                    <Box sx={{ display: "flex", gap: 1 }}>
                      <Chip
                        color="warning"
                        label={`${item.negotiationBudgetCount} em aberto`}
                        size="small"
                        variant="outlined"
                      />
                      <Button
                        onClick={() =>
                          handleOpenBudgetList({
                            salespersonId:
                              salespersonIdByName.get(item.label) ?? undefined,
                            statusId:
                              statusIdByNormalizedName.get(
                                normalizeLookupKey("Em Negociacao"),
                              ) ?? undefined,
                          })
                        }
                        size="small"
                        startIcon={<OpenInNewRoundedIcon />}
                        variant="text"
                      >
                        Abrir
                      </Button>
                    </Box>
                  </Box>
                  <Box
                    sx={{
                      alignItems: "center",
                      display: "flex",
                      flexWrap: "wrap",
                      gap: 1,
                      mt: 0.5,
                    }}
                  >
                    <Chip
                      label={currencyFormatter.format(
                        item.negotiationGrossValue,
                      )}
                      size="small"
                      variant="outlined"
                    />
                    <Chip
                      label={`${item.stalledBudgetCount} parado(s)`}
                      size="small"
                      variant="outlined"
                    />
                    <Chip
                      label={`Ultima atividade ${formatDateOrFallback(item.lastActivityAt, "Nao informada")}`}
                      size="small"
                      variant="outlined"
                    />
                  </Box>
                </Box>
              ))}
            </Box>
          ) : (
            <Alert severity="info" variant="outlined">
              Nenhuma carteira em negociacao encontrada para o filtro atual.
            </Alert>
          )}
        </SectionCard>

        <SectionCard
          description="Comparativo do funil principal por vendedor com foco em negociacao, pedidos e perdas."
          title="Funil por vendedor"
        >
          {(dashboardData?.salespersonFunnel ?? []).length ? (
            <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
              {dashboardData.salespersonFunnel.map((item) => {
                const wonWidth =
                  item.totalBudgets === 0
                    ? 0
                    : (item.wonBudgets / item.totalBudgets) * 100;
                const negotiationWidth =
                  item.totalBudgets === 0
                    ? 0
                    : (item.negotiationBudgets / item.totalBudgets) * 100;
                const lostWidth =
                  item.totalBudgets === 0
                    ? 0
                    : (item.lostBudgets / item.totalBudgets) * 100;

                return (
                  <Box key={item.label}>
                    <Box
                      sx={{
                        alignItems: "center",
                        display: "flex",
                        justifyContent: "space-between",
                      }}
                    >
                      <Typography sx={{ fontWeight: 700 }} variant="body2">
                        {item.label}
                      </Typography>
                      <Chip
                        label={`${item.totalBudgets} orc.`}
                        size="small"
                        variant="outlined"
                      />
                    </Box>
                    <Box
                      sx={{
                        bgcolor: "action.hover",
                        borderRadius: 999,
                        display: "flex",
                        height: 10,
                        mt: 1,
                        overflow: "hidden",
                      }}
                    >
                      <Box
                        sx={{
                          bgcolor: "success.main",
                          width: `${wonWidth}%`,
                        }}
                      />
                      <Box
                        sx={{
                          bgcolor: "warning.main",
                          width: `${negotiationWidth}%`,
                        }}
                      />
                      <Box
                        sx={{
                          bgcolor: "text.disabled",
                          width: `${lostWidth}%`,
                        }}
                      />
                    </Box>
                    <Box
                      sx={{
                        alignItems: "center",
                        display: "flex",
                        flexWrap: "wrap",
                        gap: 1,
                        mt: 1,
                      }}
                    >
                      <Chip
                        color="success"
                        label={`Pedido ${item.wonBudgets}`}
                        size="small"
                        variant="outlined"
                      />
                      <Chip
                        color="warning"
                        label={`Negociacao ${item.negotiationBudgets}`}
                        size="small"
                        variant="outlined"
                      />
                      <Chip
                        label={`Cancelado ${item.lostBudgets}`}
                        size="small"
                        variant="outlined"
                      />
                      <Chip
                        label={`Conversao ${formatPercentage(item.conversionRate)}`}
                        size="small"
                        variant="outlined"
                      />
                    </Box>
                  </Box>
                );
              })}
            </Box>
          ) : (
            <Alert severity="info" variant="outlined">
              Nenhum dado de funil encontrado para o filtro atual.
            </Alert>
          )}
        </SectionCard>

        <SectionCard
          description="Leitura rapida da atividade mais recente por vendedor dentro do recorte."
          title="Ultima atividade"
        >
          {(dashboardData?.recentSalespeople ?? []).length ? (
            <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
              {dashboardData.recentSalespeople.map((item) => (
                <Box
                  key={item.label}
                  sx={{
                    alignItems: "center",
                    display: "flex",
                    justifyContent: "space-between",
                  }}
                >
                  <Box>
                    <Typography sx={{ fontWeight: 700 }} variant="body2">
                      {item.label}
                    </Typography>
                    <Typography color="text.secondary" variant="caption">
                      {`${item.budgetCount} orcamento(s) · ${currencyFormatter.format(item.grossValue)}`}
                    </Typography>
                  </Box>
                  <Chip
                    icon={<InsightsRoundedIcon />}
                    label={formatDateOrFallback(
                      item.lastActivityAt,
                      "Nao informada",
                    )}
                    size="small"
                    variant="outlined"
                  />
                </Box>
              ))}
            </Box>
          ) : (
            <Alert severity="info" variant="outlined">
              Nenhuma atividade encontrada para o filtro atual.
            </Alert>
          )}
        </SectionCard>

        <SectionCard
          description="Oportunidades em negociacao sem atualizacao recente, priorizadas pelos casos mais antigos."
          title="Orcamentos parados"
        >
          {(dashboardData?.staleBudgets ?? []).length ? (
            <Box sx={{ display: "flex", flexDirection: "column", gap: 1.5 }}>
              {dashboardData.staleBudgets.map((item, index) => (
                <Box key={item.id}>
                  {index > 0 ? <Divider sx={{ mb: 1.5 }} /> : null}
                  <Box
                    sx={{
                      display: "flex",
                      flexDirection: "column",
                      gap: 0.75,
                    }}
                  >
                    <Box
                      sx={{
                        alignItems: "center",
                        display: "flex",
                        justifyContent: "space-between",
                      }}
                    >
                      <Typography sx={{ fontWeight: 700 }} variant="body2">
                        {item.budgetNumber}
                      </Typography>
                      <Box sx={{ display: "flex", gap: 1 }}>
                        <Chip
                          color="warning"
                          label={`${item.stalledDays} dia(s)`}
                          size="small"
                          variant="outlined"
                        />
                        <Button
                          onClick={() =>
                            handleOpenBudgetList({
                              budgetNumber: item.budgetNumber,
                            })
                          }
                          size="small"
                          startIcon={<OpenInNewRoundedIcon />}
                          variant="text"
                        >
                          Abrir
                        </Button>
                      </Box>
                    </Box>
                    <Typography color="text.secondary" variant="caption">
                      {`${item.salespersonLabel} · ${item.projectLabel}`}
                    </Typography>
                    <Typography color="text.secondary" variant="caption">
                      {`${item.constructionCompanyLabel} · ${item.statusLabel}`}
                    </Typography>
                    <Box
                      sx={{
                        alignItems: "center",
                        display: "flex",
                        flexWrap: "wrap",
                        gap: 1,
                      }}
                    >
                      <Chip
                        label={currencyFormatter.format(item.grossValue)}
                        size="small"
                        variant="outlined"
                      />
                      <Chip
                        label={`Ultima atividade ${formatDateOrFallback(item.lastActivityAt, "Nao informada")}`}
                        size="small"
                        variant="outlined"
                      />
                    </Box>
                  </Box>
                </Box>
              ))}
            </Box>
          ) : (
            <Alert severity="success" variant="outlined">
              Nenhum orcamento parado encontrado para o filtro atual.
            </Alert>
          )}
        </SectionCard>
      </Box>

      <Box
        sx={{
          display: "grid",
          gap: 3,
          gridTemplateColumns: {
            lg: "minmax(0, 1.4fr) minmax(0, 0.6fr)",
            xs: "minmax(0, 1fr)",
          },
        }}
      >
        <SectionCard
          description="Evolucao do volume orcado e do valor convertido ao longo do tempo dentro do recorte atual."
          title="Evolucao mensal"
        >
          {(dashboardData?.monthlyEvolution ?? []).length ? (
            <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
              {dashboardData.monthlyEvolution.map((item) => {
                const progressValue =
                  (dashboardData.maxMonthlyGrossValue ?? 0) === 0
                    ? 0
                    : (item.grossValue / dashboardData.maxMonthlyGrossValue) *
                      100;

                return (
                  <Box key={item.monthKey}>
                    <Box
                      sx={{
                        alignItems: "center",
                        display: "flex",
                        justifyContent: "space-between",
                      }}
                    >
                      <Typography sx={{ fontWeight: 700 }} variant="body2">
                        {item.monthLabel}
                      </Typography>
                      <Chip
                        label={`${item.budgetCount} orc.`}
                        size="small"
                        variant="outlined"
                      />
                    </Box>
                    <LinearProgress
                      sx={{ borderRadius: 999, height: 8, mt: 1 }}
                      value={progressValue}
                      variant="determinate"
                    />
                    <Box
                      sx={{
                        alignItems: "center",
                        display: "flex",
                        flexWrap: "wrap",
                        gap: 1,
                        mt: 1,
                      }}
                    >
                      <Chip
                        label={currencyFormatter.format(item.grossValue)}
                        size="small"
                        variant="outlined"
                      />
                      <Chip
                        color="success"
                        label={`Pedido ${currencyFormatter.format(item.wonGrossValue)}`}
                        size="small"
                        variant="outlined"
                      />
                      <Chip
                        label={`${item.wonBudgetCount} pedido(s)`}
                        size="small"
                        variant="outlined"
                      />
                    </Box>
                  </Box>
                );
              })}
            </Box>
          ) : (
            <Alert severity="info" variant="outlined">
              Nenhuma evolucao mensal encontrada para o filtro atual.
            </Alert>
          )}
        </SectionCard>

        <SectionCard
          description="Leitura consolidada do funil principal dentro do periodo selecionado."
          title="Resumo da carteira"
        >
          <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
            {[
              {
                color: "success.main",
                label: "Pedidos",
                value: dashboardData?.wonBudgets ?? 0,
              },
              {
                color: "warning.main",
                label: "Em negociacao",
                value: dashboardData?.negotiationBudgets ?? 0,
              },
              {
                color: "text.secondary",
                label: "Cancelados",
                value: dashboardData?.lostBudgets ?? 0,
              },
              {
                color: "primary.main",
                label: "Conversao",
                value: formatPercentage(dashboardData?.conversionRate ?? 0),
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
    </Box>
  );
}
