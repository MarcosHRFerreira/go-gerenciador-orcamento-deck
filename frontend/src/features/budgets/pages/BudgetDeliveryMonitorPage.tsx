import LaunchRoundedIcon from "@mui/icons-material/LaunchRounded";
import LocalShippingRoundedIcon from "@mui/icons-material/LocalShippingRounded";
import SearchRoundedIcon from "@mui/icons-material/SearchRounded";
import {
  Alert,
  Box,
  Button,
  Chip,
  CircularProgress,
  MenuItem,
  Pagination,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TextField,
  Typography,
} from "@mui/material";
import { alpha } from "@mui/material/styles";
import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { isAxiosError } from "axios";
import { useEffect, useMemo, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import {
  compactFilterFieldSx,
  FilterField,
  filterGroupSx,
  filterGroupTitleSx,
  filterSectionCardSx,
} from "../../../components/common/FilterField";
import { PageHeader } from "../../../components/common/PageHeader";
import { SectionCard } from "../../../components/common/SectionCard";
import { useAuth } from "../../auth/hooks/useAuth";
import {
  getBudgetCatalogsRequest,
  getBudgetDeliveryMonitorRequest,
  getBudgetListCatalogsRequest,
} from "../api/budgets";
import type {
  BudgetCatalogItem,
  BudgetDeliveryMonitorFilters,
  BudgetDeliveryMonitorItem,
  BudgetDeliveryStatus,
} from "../types/budget";

const dateFormatter = new Intl.DateTimeFormat("pt-BR", {
  dateStyle: "short",
});

const dateTimeFormatter = new Intl.DateTimeFormat("pt-BR", {
  dateStyle: "short",
  timeStyle: "short",
});

const defaultPageSize = 25;

const defaultFilters: BudgetDeliveryMonitorFilters = {
  budgetNumber: "",
  projectName: "",
  salespersonId: "",
  statusId: "",
  deliveryDateFrom: "",
  deliveryDateTo: "",
  deliveryStatus: "",
  missingDeliveryDate: false,
  page: 1,
  pageSize: defaultPageSize,
};

const deliveryStatusOptions: Array<{
  label: string;
  value: BudgetDeliveryStatus;
}> = [
  { value: "overdue", label: "Atrasado" },
  { value: "due_today", label: "Entrega hoje" },
  { value: "due_in_1_day", label: "Entrega em 1 dia" },
  { value: "due_in_2_days", label: "Entrega em 2 dias" },
  { value: "future", label: "Entrega futura" },
  { value: "missing_delivery_date", label: "Pedido sem data" },
];

const pageSizeOptions = [10, 25, 50] as const;

const monitorSummaryGridSx = {
  display: "grid",
  gap: 2,
  gridTemplateColumns: {
    lg: "repeat(5, minmax(0, 1fr))",
    md: "repeat(3, minmax(0, 1fr))",
    sm: "repeat(2, minmax(0, 1fr))",
    xs: "minmax(0, 1fr)",
  },
} as const;

const monitorSummaryCardSx = {
  border: "1px solid",
  borderColor: (theme: { palette: { primary: { main: string } } }) =>
    alpha(theme.palette.primary.main, 0.12),
  borderRadius: 3,
  display: "grid",
  gap: 0.5,
  p: 2,
} as const;

const monitorFilterGridSx = {
  display: "grid",
  gap: 2,
  gridTemplateColumns: {
    xl: "repeat(4, minmax(0, 1fr))",
    md: "repeat(3, minmax(0, 1fr))",
    sm: "repeat(2, minmax(0, 1fr))",
    xs: "minmax(0, 1fr)",
  },
} as const;

const tableHeadCellSx = {
  backgroundColor: (theme: {
    palette: { mode: string; primary: { light: string } };
  }) =>
    theme.palette.mode === "dark"
      ? alpha(theme.palette.primary.light, 0.16)
      : "#DBEAFE",
  borderBottomColor: (theme: {
    palette: { mode: string; primary: { light: string } };
  }) =>
    theme.palette.mode === "dark" ? theme.palette.primary.light : "#1D4ED8",
  borderBottomStyle: "solid",
  borderBottomWidth: 2,
  color: (theme: { palette: { mode: string; primary: { light: string } } }) =>
    theme.palette.mode === "dark" ? theme.palette.primary.light : "#1D4ED8",
  fontSize: "0.75rem",
  fontWeight: 800,
  letterSpacing: "0.05em",
  py: 1.2,
  textTransform: "uppercase",
  whiteSpace: "nowrap",
} as const;

const tableDetailCellSx = {
  borderBottomColor: "divider",
  borderBottomStyle: "solid",
  borderBottomWidth: 1,
  color: "text.primary",
  fontSize: "0.78rem",
  fontWeight: 500,
  lineHeight: 1.35,
  py: 0.8,
  verticalAlign: "middle",
} as const;

const singleLineTableCellSx = {
  ...tableDetailCellSx,
  maxWidth: 220,
  overflow: "hidden",
  textOverflow: "ellipsis",
  whiteSpace: "nowrap",
} as const;

const premiumBudgetAlertSx = {
  borderRadius: 3,
  boxShadow: "0 14px 30px rgba(30, 58, 138, 0.08)",
  "& .MuiAlert-message": {
    fontWeight: 600,
  },
} as const;

const budgetPaginationSx = {
  "& .MuiPaginationItem-root": {
    color: "text.primary",
    fontWeight: 700,
  },
} as const;

const rowNumberColumnWidth = 68;
const budgetNumberColumnWidth = 140;
const tableMaxHeight = "calc(100vh - 300px)";

function parsePositiveInteger(value: string | null, fallback: number) {
  const parsedValue = Number(value);

  if (!Number.isInteger(parsedValue) || parsedValue <= 0) {
    return fallback;
  }

  return parsedValue;
}

function isDeliveryStatus(value: string | null): value is BudgetDeliveryStatus {
  return deliveryStatusOptions.some((option) => option.value === value);
}

function getFiltersFromSearchParams(
  searchParams: URLSearchParams,
): BudgetDeliveryMonitorFilters {
  const deliveryStatus = searchParams.get("deliveryStatus");

  return {
    budgetNumber:
      searchParams.get("budgetNumber") ?? defaultFilters.budgetNumber,
    projectName: searchParams.get("projectName") ?? defaultFilters.projectName,
    salespersonId:
      searchParams.get("salespersonId") ?? defaultFilters.salespersonId,
    statusId: searchParams.get("statusId") ?? defaultFilters.statusId,
    deliveryDateFrom:
      searchParams.get("deliveryDateFrom") ?? defaultFilters.deliveryDateFrom,
    deliveryDateTo:
      searchParams.get("deliveryDateTo") ?? defaultFilters.deliveryDateTo,
    deliveryStatus: isDeliveryStatus(deliveryStatus)
      ? deliveryStatus
      : defaultFilters.deliveryStatus,
    missingDeliveryDate: searchParams.get("missingDeliveryDate") === "true",
    page: parsePositiveInteger(searchParams.get("page"), defaultFilters.page),
    pageSize: parsePositiveInteger(
      searchParams.get("pageSize"),
      defaultFilters.pageSize,
    ),
  };
}

function buildSearchParams(filters: BudgetDeliveryMonitorFilters) {
  const nextSearchParams = new URLSearchParams();

  if (filters.budgetNumber) {
    nextSearchParams.set("budgetNumber", filters.budgetNumber);
  }
  if (filters.projectName) {
    nextSearchParams.set("projectName", filters.projectName);
  }
  if (filters.salespersonId) {
    nextSearchParams.set("salespersonId", filters.salespersonId);
  }
  if (filters.statusId) {
    nextSearchParams.set("statusId", filters.statusId);
  }
  if (filters.deliveryDateFrom) {
    nextSearchParams.set("deliveryDateFrom", filters.deliveryDateFrom);
  }
  if (filters.deliveryDateTo) {
    nextSearchParams.set("deliveryDateTo", filters.deliveryDateTo);
  }
  if (filters.deliveryStatus) {
    nextSearchParams.set("deliveryStatus", filters.deliveryStatus);
  }
  if (filters.missingDeliveryDate) {
    nextSearchParams.set("missingDeliveryDate", "true");
  }
  if (filters.page !== defaultFilters.page) {
    nextSearchParams.set("page", String(filters.page));
  }
  if (filters.pageSize !== defaultFilters.pageSize) {
    nextSearchParams.set("pageSize", String(filters.pageSize));
  }

  return nextSearchParams;
}

function getBudgetErrorMessage(error: unknown) {
  if (isAxiosError<{ message?: string }>(error)) {
    return (
      error.response?.data?.message ??
      "Não foi possível carregar o acompanhamento de entregas."
    );
  }

  return "Não foi possível carregar o acompanhamento de entregas.";
}

function formatDate(value: string | null, fallback = "Nao informada") {
  if (!value) {
    return fallback;
  }

  const parsedDate = new Date(`${value}T00:00:00`);
  if (Number.isNaN(parsedDate.getTime())) {
    return fallback;
  }

  return dateFormatter.format(parsedDate);
}

function formatDateTime(value: string) {
  const parsedDate = new Date(value);
  if (Number.isNaN(parsedDate.getTime())) {
    return "Nao informado";
  }

  return dateTimeFormatter.format(parsedDate);
}

function formatProjectLabel(item: BudgetDeliveryMonitorItem) {
  if (item.projectName && item.projectCode) {
    return `${item.projectCode} - ${item.projectName}`;
  }
  if (item.projectName) {
    return item.projectName;
  }
  if (item.projectCode) {
    return item.projectCode;
  }

  return "Nao informada";
}

function formatDaysUntilDelivery(value: number | null) {
  if (value === null) {
    return "Sem data";
  }
  if (value === 0) {
    return "Hoje";
  }
  if (value === 1) {
    return "1 dia";
  }
  if (value < 0) {
    return `${Math.abs(value)} dia(s) atrasado(s)`;
  }

  return `${value} dias`;
}

function getDeliveryStatusChipColor(status: BudgetDeliveryStatus) {
  switch (status) {
    case "overdue":
      return "error";
    case "due_today":
      return "warning";
    case "due_in_1_day":
    case "due_in_2_days":
      return "info";
    case "future":
      return "success";
    case "missing_delivery_date":
      return "default";
    default:
      return "default";
  }
}

function createCatalogMap(items: BudgetCatalogItem[]) {
  return new Map(items.map((item) => [String(item.id), item.name]));
}

export function BudgetDeliveryMonitorPage() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const { user } = useAuth();
  const isAdmin = user?.role === "admin";
  const appliedFilters = useMemo(
    () => getFiltersFromSearchParams(searchParams),
    [searchParams],
  );
  const [draftFilters, setDraftFilters] =
    useState<BudgetDeliveryMonitorFilters>(appliedFilters);

  const monitorQuery = useQuery({
    queryKey: ["budget-delivery-monitor", appliedFilters],
    queryFn: () => getBudgetDeliveryMonitorRequest(appliedFilters),
    placeholderData: keepPreviousData,
  });

  const catalogsQuery = useQuery({
    queryKey: ["budget-delivery-monitor-catalogs", isAdmin],
    queryFn: isAdmin ? getBudgetCatalogsRequest : getBudgetListCatalogsRequest,
    placeholderData: keepPreviousData,
  });

  useEffect(() => {
    setDraftFilters(appliedFilters);
  }, [appliedFilters]);

  const statusCatalogMap = useMemo(
    () => createCatalogMap(catalogsQuery.data?.statuses ?? []),
    [catalogsQuery.data?.statuses],
  );

  const totalPages = Math.max(
    1,
    Math.ceil((monitorQuery.data?.total ?? 0) / appliedFilters.pageSize),
  );

  const handleDraftChange = <Key extends keyof BudgetDeliveryMonitorFilters>(
    key: Key,
    value: BudgetDeliveryMonitorFilters[Key],
  ) => {
    setDraftFilters((currentFilters) => ({
      ...currentFilters,
      [key]: value,
    }));
  };

  const handleApplyFilters = () => {
    setSearchParams(
      buildSearchParams({
        ...draftFilters,
        page: 1,
      }),
    );
  };

  const handleClearFilters = () => {
    setDraftFilters(defaultFilters);
    setSearchParams(buildSearchParams(defaultFilters));
  };

  const handleChangePage = (
    _event: React.ChangeEvent<unknown>,
    page: number,
  ) => {
    setSearchParams(
      buildSearchParams({
        ...appliedFilters,
        page,
      }),
    );
  };

  const handlePageSizeChange = (value: string) => {
    setSearchParams(
      buildSearchParams({
        ...appliedFilters,
        page: 1,
        pageSize: Number(value),
      }),
    );
  };

  return (
    <Box sx={{ display: "grid", gap: 3 }}>
      <PageHeader
        title="Acompanhamento de entregas"
        description="Monitore pedidos atrasados, proximos da entrega e registros sem data informada."
        action={
          <Button
            onClick={() => navigate("/budgets")}
            startIcon={<LocalShippingRoundedIcon />}
            variant="outlined"
          >
            Voltar para orcamentos
          </Button>
        }
      />

      <SectionCard
        title="Resumo operacional"
        description="Cards principais para priorizar as pendencias de entrega."
        sx={filterSectionCardSx}
      >
        <Box sx={monitorSummaryGridSx}>
          {[
            {
              helper: "Pedidos no contexto atual",
              label: "Total monitorado",
              value: monitorQuery.data?.summary.total ?? 0,
            },
            {
              helper: "Precisam de acao imediata",
              label: "Atrasados",
              value: monitorQuery.data?.summary.overdueCount ?? 0,
            },
            {
              helper: "Entrega prevista para hoje",
              label: "Entrega hoje",
              value: monitorQuery.data?.summary.dueTodayCount ?? 0,
            },
            {
              helper: "Entrega em 1 ou 2 dias",
              label: "Proximos 2 dias",
              value: monitorQuery.data?.summary.dueInUpTo2DaysCount ?? 0,
            },
            {
              helper: "Pendencia de preenchimento",
              label: "Sem data",
              value: monitorQuery.data?.summary.missingDeliveryCount ?? 0,
            },
          ].map((card) => (
            <Box key={card.label} sx={monitorSummaryCardSx}>
              <Typography color="text.secondary" variant="body2">
                {card.label}
              </Typography>
              <Typography sx={{ fontWeight: 800 }} variant="h4">
                {card.value}
              </Typography>
              <Typography color="text.secondary" variant="caption">
                {card.helper}
              </Typography>
            </Box>
          ))}
        </Box>
      </SectionCard>

      <SectionCard
        title="Filtros"
        description="Use os filtros para localizar rapidamente os pedidos que exigem acompanhamento."
        sx={filterSectionCardSx}
      >
        <Box sx={filterGroupSx}>
          <Typography sx={filterGroupTitleSx} variant="subtitle2">
            Consulta operacional
          </Typography>

          <Box sx={monitorFilterGridSx}>
            <FilterField label="Obra">
              <TextField
                value={draftFilters.projectName}
                onChange={(event) =>
                  handleDraftChange("projectName", event.target.value)
                }
                placeholder="Nome ou codigo da obra"
                size="small"
                sx={compactFilterFieldSx}
              />
            </FilterField>

            <FilterField label="Orcamento">
              <TextField
                value={draftFilters.budgetNumber}
                onChange={(event) =>
                  handleDraftChange("budgetNumber", event.target.value)
                }
                placeholder="Numero do orcamento"
                size="small"
                sx={compactFilterFieldSx}
              />
            </FilterField>

            {isAdmin ? (
              <FilterField label="Vendedor">
                <TextField
                  select
                  value={draftFilters.salespersonId}
                  onChange={(event) =>
                    handleDraftChange("salespersonId", event.target.value)
                  }
                  size="small"
                  sx={compactFilterFieldSx}
                >
                  <MenuItem value="">Todos</MenuItem>
                  {(catalogsQuery.data?.salespeople ?? []).map((item) => (
                    <MenuItem key={item.id} value={String(item.id)}>
                      {item.name}
                    </MenuItem>
                  ))}
                </TextField>
              </FilterField>
            ) : null}

            <FilterField label="Status do orcamento">
              <TextField
                select
                value={draftFilters.statusId}
                onChange={(event) =>
                  handleDraftChange("statusId", event.target.value)
                }
                size="small"
                sx={compactFilterFieldSx}
              >
                <MenuItem value="">Pedido</MenuItem>
                {(catalogsQuery.data?.statuses ?? []).map((item) => (
                  <MenuItem key={item.id} value={String(item.id)}>
                    {item.name}
                  </MenuItem>
                ))}
              </TextField>
            </FilterField>

            <FilterField label="Situacao da entrega">
              <TextField
                select
                value={draftFilters.deliveryStatus}
                onChange={(event) =>
                  handleDraftChange(
                    "deliveryStatus",
                    event.target
                      .value as BudgetDeliveryMonitorFilters["deliveryStatus"],
                  )
                }
                size="small"
                sx={compactFilterFieldSx}
              >
                <MenuItem value="">Todas</MenuItem>
                {deliveryStatusOptions.map((option) => (
                  <MenuItem key={option.value} value={option.value}>
                    {option.label}
                  </MenuItem>
                ))}
              </TextField>
            </FilterField>

            <FilterField label="Entrega de">
              <TextField
                type="date"
                value={draftFilters.deliveryDateFrom}
                onChange={(event) =>
                  handleDraftChange("deliveryDateFrom", event.target.value)
                }
                size="small"
                sx={compactFilterFieldSx}
              />
            </FilterField>

            <FilterField label="Entrega ate">
              <TextField
                type="date"
                value={draftFilters.deliveryDateTo}
                onChange={(event) =>
                  handleDraftChange("deliveryDateTo", event.target.value)
                }
                size="small"
                sx={compactFilterFieldSx}
              />
            </FilterField>

            <FilterField label="Pendencia rapida">
              <Button
                onClick={() =>
                  handleDraftChange(
                    "missingDeliveryDate",
                    !draftFilters.missingDeliveryDate,
                  )
                }
                variant={
                  draftFilters.missingDeliveryDate ? "contained" : "outlined"
                }
              >
                {draftFilters.missingDeliveryDate
                  ? "Somente pedidos sem data"
                  : "Filtrar pedidos sem data"}
              </Button>
            </FilterField>
          </Box>

          <Box
            sx={{
              display: "flex",
              flexWrap: "wrap",
              gap: 1.5,
              justifyContent: "flex-end",
            }}
          >
            <Button onClick={handleClearFilters} variant="outlined">
              Limpar
            </Button>
            <Button
              onClick={handleApplyFilters}
              startIcon={<SearchRoundedIcon />}
              variant="contained"
            >
              Filtrar
            </Button>
          </Box>
        </Box>
      </SectionCard>

      <SectionCard
        title="Pedidos monitorados"
        description="Lista inicial com destaque visual para prioridade de atendimento."
        sx={{
          background: (theme) =>
            `linear-gradient(135deg, ${alpha(theme.palette.info.main, 0.05)} 0%, ${alpha(theme.palette.primary.main, 0.03)} 100%)`,
          border: "1px solid",
          borderColor: (theme) => alpha(theme.palette.info.main, 0.16),
          boxShadow: (theme) =>
            `0 14px 28px ${alpha(theme.palette.info.main, 0.08)}`,
        }}
      >
        {monitorQuery.isError ? (
          <Alert severity="error" sx={premiumBudgetAlertSx}>
            {getBudgetErrorMessage(monitorQuery.error)}
          </Alert>
        ) : null}

        {monitorQuery.isLoading ? (
          <Box
            sx={{
              alignItems: "center",
              display: "flex",
              justifyContent: "center",
              minHeight: 240,
            }}
          >
            <CircularProgress />
          </Box>
        ) : null}

        {!monitorQuery.isLoading && !monitorQuery.isError ? (
          <>
            <Box
              sx={{
                alignItems: { lg: "center", xs: "flex-start" },
                display: "flex",
                flexDirection: { lg: "row", xs: "column" },
                gap: 1.5,
                justifyContent: "space-between",
                mb: 2,
              }}
            >
              <Box sx={{ display: "flex", flexDirection: "column", gap: 0.9 }}>
                <Box sx={{ display: "flex", flexWrap: "wrap", gap: 1 }}>
                  <Chip
                    color="primary"
                    label="Lista completa"
                    size="small"
                    variant="outlined"
                  />
                  <Chip
                    label={`Página ${monitorQuery.data?.page ?? appliedFilters.page} de ${totalPages}`}
                    size="small"
                    variant="outlined"
                  />
                  <Chip
                    label={`${monitorQuery.data?.total ?? 0} pedido(s) no recorte`}
                    size="small"
                    variant="outlined"
                  />
                </Box>
                <Typography
                  color="text.primary"
                  sx={{ fontWeight: 600, maxWidth: 780 }}
                  variant="body2"
                >
                  A listagem abaixo prioriza leitura operacional, comparação
                  rápida e acesso direto às ações principais.
                </Typography>
              </Box>

              <FilterField label="Linhas por pagina">
                <TextField
                  select
                  value={String(appliedFilters.pageSize)}
                  onChange={(event) => handlePageSizeChange(event.target.value)}
                  size="small"
                  sx={{ ...compactFilterFieldSx, minWidth: 160 }}
                >
                  {pageSizeOptions.map((option) => (
                    <MenuItem key={option} value={String(option)}>
                      {option}
                    </MenuItem>
                  ))}
                </TextField>
              </FilterField>
            </Box>

            {monitorQuery.data?.items.length ? (
              <>
                <Box
                  sx={{
                    border: (theme) =>
                      `1px solid ${alpha(theme.palette.primary.main, 0.18)}`,
                    borderRadius: 3,
                    boxShadow: (theme) =>
                      `0 12px 24px ${alpha(theme.palette.primary.main, 0.06)}`,
                    height: tableMaxHeight,
                    minWidth: 0,
                    overflow: "hidden",
                    width: "100%",
                  }}
                >
                  <TableContainer
                    sx={{
                      boxSizing: "border-box",
                      height: "100%",
                      overflow: "auto",
                    }}
                  >
                    <Table
                      size="small"
                      stickyHeader
                      sx={{
                        borderCollapse: "separate",
                        borderSpacing: 0,
                        minWidth: 1480,
                        "& .MuiTableHead-root": {
                          position: "relative",
                          zIndex: 3,
                        },
                        "& .MuiTableCell-stickyHeader": {
                          backgroundColor: (theme) =>
                            theme.palette.background.paper,
                          backgroundImage: (theme) =>
                            theme.palette.mode === "dark"
                              ? `linear-gradient(180deg, ${alpha(theme.palette.background.paper, 0.98)} 0%, ${alpha(theme.palette.primary.light, 0.12)} 100%)`
                              : "linear-gradient(180deg, rgba(219, 234, 254, 0.98) 0%, rgba(191, 219, 254, 0.98) 100%)",
                          boxShadow: (theme) =>
                            theme.palette.mode === "dark"
                              ? `inset 0 -1px 0 ${alpha(theme.palette.primary.light, 0.26)}, 0 10px 18px rgba(2, 6, 23, 0.28)`
                              : "inset 0 -1px 0 rgba(30, 58, 138, 0.22), 0 10px 18px rgba(15, 23, 42, 0.08)",
                          top: 0,
                          zIndex: 4,
                        },
                        "& .MuiTableBody-root .MuiTableRow-root:hover": {
                          backgroundColor: (theme) =>
                            alpha(theme.palette.info.main, 0.06),
                        },
                      }}
                    >
                      <TableHead>
                        <TableRow>
                          <TableCell
                            align="center"
                            sx={{
                              ...tableHeadCellSx,
                              minWidth: rowNumberColumnWidth,
                              width: rowNumberColumnWidth,
                            }}
                          >
                            #
                          </TableCell>
                          <TableCell
                            sx={{
                              ...tableHeadCellSx,
                              minWidth: budgetNumberColumnWidth,
                              width: budgetNumberColumnWidth,
                            }}
                          >
                            Orçamento
                          </TableCell>
                          <TableCell sx={tableHeadCellSx}>Obra</TableCell>
                          <TableCell sx={tableHeadCellSx}>Empresa</TableCell>
                          <TableCell sx={tableHeadCellSx}>Vendedor</TableCell>
                          <TableCell sx={tableHeadCellSx}>Status</TableCell>
                          <TableCell sx={tableHeadCellSx}>
                            Data de entrega
                          </TableCell>
                          <TableCell sx={tableHeadCellSx}>Dias</TableCell>
                          <TableCell sx={tableHeadCellSx}>Situação</TableCell>
                          <TableCell sx={tableHeadCellSx}>
                            Atualização
                          </TableCell>
                          <TableCell sx={tableHeadCellSx}>Ações</TableCell>
                        </TableRow>
                      </TableHead>
                      <TableBody>
                        {monitorQuery.data.items.map((item, index) => (
                          <TableRow hover key={item.id}>
                            <TableCell
                              align="center"
                              sx={{
                                ...singleLineTableCellSx,
                                fontWeight: 700,
                                maxWidth: rowNumberColumnWidth,
                                minWidth: rowNumberColumnWidth,
                                width: rowNumberColumnWidth,
                              }}
                            >
                              {(appliedFilters.page - 1) *
                                appliedFilters.pageSize +
                                index +
                                1}
                            </TableCell>
                            <TableCell
                              sx={{
                                ...singleLineTableCellSx,
                                fontWeight: 600,
                                maxWidth: budgetNumberColumnWidth,
                                minWidth: budgetNumberColumnWidth,
                                width: budgetNumberColumnWidth,
                              }}
                              title={item.budgetNumber}
                            >
                              {item.budgetNumber}
                            </TableCell>
                            <TableCell
                              sx={singleLineTableCellSx}
                              title={formatProjectLabel(item)}
                            >
                              {formatProjectLabel(item)}
                            </TableCell>
                            <TableCell
                              sx={singleLineTableCellSx}
                              title={
                                item.constructionCompany || "Nao informada"
                              }
                            >
                              {item.constructionCompany || "Nao informada"}
                            </TableCell>
                            <TableCell
                              sx={singleLineTableCellSx}
                              title={item.salespersonName || "Nao informado"}
                            >
                              {item.salespersonName || "Nao informado"}
                            </TableCell>
                            <TableCell sx={tableDetailCellSx}>
                              <Chip
                                color="primary"
                                label={
                                  item.statusName ??
                                  statusCatalogMap.get(String(item.statusId)) ??
                                  "Nao informado"
                                }
                                size="small"
                                sx={{
                                  fontSize: "0.7rem",
                                  fontWeight: 700,
                                  height: 22,
                                }}
                                variant="outlined"
                              />
                            </TableCell>
                            <TableCell sx={singleLineTableCellSx}>
                              {formatDate(item.deliveryDate)}
                            </TableCell>
                            <TableCell sx={singleLineTableCellSx}>
                              {formatDaysUntilDelivery(item.daysUntilDelivery)}
                            </TableCell>
                            <TableCell sx={tableDetailCellSx}>
                              <Chip
                                color={getDeliveryStatusChipColor(
                                  item.deliveryStatus,
                                )}
                                label={item.deliveryStatusLabel}
                                size="small"
                                sx={{
                                  fontSize: "0.7rem",
                                  fontWeight: 700,
                                  height: 22,
                                }}
                                variant={
                                  item.deliveryStatus ===
                                  "missing_delivery_date"
                                    ? "outlined"
                                    : "filled"
                                }
                              />
                            </TableCell>
                            <TableCell sx={singleLineTableCellSx}>
                              {formatDateTime(item.updatedAt)}
                            </TableCell>
                            <TableCell
                              sx={{
                                ...tableDetailCellSx,
                                whiteSpace: "nowrap",
                              }}
                            >
                              <Button
                                onClick={() =>
                                  navigate(`/budgets/${item.id}/edit`)
                                }
                                size="small"
                                startIcon={<LaunchRoundedIcon />}
                                sx={{ minWidth: "auto", px: 0.75, py: 0.25 }}
                                variant="text"
                              >
                                Editar
                              </Button>
                              {item.projectId ? (
                                <Button
                                  onClick={() =>
                                    navigate(`/projects/${item.projectId}`)
                                  }
                                  size="small"
                                  sx={{ minWidth: "auto", px: 0.75, py: 0.25 }}
                                  variant="text"
                                >
                                  Obra
                                </Button>
                              ) : null}
                            </TableCell>
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                  </TableContainer>
                </Box>

                <Box
                  sx={{
                    alignItems: { md: "center", xs: "flex-start" },
                    display: "flex",
                    flexDirection: { md: "row", xs: "column" },
                    gap: 1.5,
                    justifyContent: "space-between",
                    mt: 2,
                  }}
                >
                  <Typography
                    color="text.primary"
                    sx={{ fontWeight: 600 }}
                    variant="body2"
                  >
                    {monitorQuery.data?.total ?? 0} resultado(s) encontrado(s)
                  </Typography>
                  <Pagination
                    color="primary"
                    count={totalPages}
                    onChange={handleChangePage}
                    page={appliedFilters.page}
                    shape="rounded"
                    sx={budgetPaginationSx}
                  />
                </Box>
              </>
            ) : (
              <Alert
                severity="info"
                sx={premiumBudgetAlertSx}
                variant="outlined"
              >
                Nenhum pedido foi encontrado com os filtros informados.
              </Alert>
            )}
          </>
        ) : null}
      </SectionCard>
    </Box>
  );
}

export default BudgetDeliveryMonitorPage;
