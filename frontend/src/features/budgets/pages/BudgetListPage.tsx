import AddRoundedIcon from "@mui/icons-material/AddRounded";
import DeleteOutlineRoundedIcon from "@mui/icons-material/DeleteOutlineRounded";
import EditRoundedIcon from "@mui/icons-material/EditRounded";
import SearchRoundedIcon from "@mui/icons-material/SearchRounded";
import UploadFileRoundedIcon from "@mui/icons-material/UploadFileRounded";
import {
  Alert,
  Box,
  Button,
  Chip,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
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
import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { isAxiosError } from "axios";
import {
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
  type UIEvent,
} from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { SectionCard } from "../../../components/common/SectionCard";
import { useAuth } from "../../auth/hooks/useAuth";
import {
  deleteBudgetRequest,
  getBudgetCatalogsRequest,
  getBudgetListCatalogsRequest,
  getBudgetListRequest,
} from "../api/budgets";
import type {
  BudgetCatalogItem,
  BudgetListFilters,
  BudgetSortBy,
  BudgetSortOrder,
} from "../types/budget";

const currencyFormatter = new Intl.NumberFormat("pt-BR", {
  style: "currency",
  currency: "BRL",
});

const dateFormatter = new Intl.DateTimeFormat("pt-BR", {
  dateStyle: "short",
});

const dateTimeFormatter = new Intl.DateTimeFormat("pt-BR", {
  dateStyle: "short",
  timeStyle: "short",
});

const decimalFormatter = new Intl.NumberFormat("pt-BR", {
  maximumFractionDigits: 2,
  minimumFractionDigits: 2,
});

const pageSize = 20;

const defaultFilters: BudgetListFilters = {
  budgetNumber: "",
  yearBudget: "",
  statusId: "",
  installerId: "",
  salespersonId: "",
  page: 1,
  pageSize,
  sortBy: "sent_at",
  sortOrder: "desc",
};

function isSortBy(value: string | null): value is BudgetSortBy {
  return (
    value === "sent_at" ||
    value === "gross_value" ||
    value === "created_at" ||
    value === "updated_at" ||
    value === "year_budget" ||
    value === "budget_number"
  );
}

function isSortOrder(value: string | null): value is BudgetSortOrder {
  return value === "asc" || value === "desc";
}

function parsePositiveInteger(value: string | null, fallback: number) {
  const parsedValue = Number(value);

  if (!Number.isInteger(parsedValue) || parsedValue <= 0) {
    return fallback;
  }

  return parsedValue;
}

function getFiltersFromSearchParams(
  searchParams: URLSearchParams,
): BudgetListFilters {
  const sortBy = searchParams.get("sortBy");
  const sortOrder = searchParams.get("sortOrder");

  return {
    budgetNumber:
      searchParams.get("budgetNumber") ?? defaultFilters.budgetNumber,
    yearBudget: searchParams.get("yearBudget") ?? defaultFilters.yearBudget,
    statusId: searchParams.get("statusId") ?? defaultFilters.statusId,
    installerId: searchParams.get("installerId") ?? defaultFilters.installerId,
    salespersonId:
      searchParams.get("salespersonId") ?? defaultFilters.salespersonId,
    page: parsePositiveInteger(searchParams.get("page"), defaultFilters.page),
    pageSize: parsePositiveInteger(
      searchParams.get("pageSize"),
      defaultFilters.pageSize,
    ),
    sortBy: isSortBy(sortBy) ? sortBy : defaultFilters.sortBy,
    sortOrder: isSortOrder(sortOrder) ? sortOrder : defaultFilters.sortOrder,
  };
}

function buildSearchParams(filters: BudgetListFilters) {
  const nextSearchParams = new URLSearchParams();

  if (filters.budgetNumber) {
    nextSearchParams.set("budgetNumber", filters.budgetNumber);
  }
  if (filters.yearBudget) {
    nextSearchParams.set("yearBudget", filters.yearBudget);
  }
  if (filters.statusId) {
    nextSearchParams.set("statusId", filters.statusId);
  }
  if (filters.installerId) {
    nextSearchParams.set("installerId", filters.installerId);
  }
  if (filters.salespersonId) {
    nextSearchParams.set("salespersonId", filters.salespersonId);
  }
  if (filters.page !== defaultFilters.page) {
    nextSearchParams.set("page", String(filters.page));
  }
  if (filters.pageSize !== defaultFilters.pageSize) {
    nextSearchParams.set("pageSize", String(filters.pageSize));
  }
  if (filters.sortBy !== defaultFilters.sortBy) {
    nextSearchParams.set("sortBy", filters.sortBy);
  }
  if (filters.sortOrder !== defaultFilters.sortOrder) {
    nextSearchParams.set("sortOrder", filters.sortOrder);
  }

  return nextSearchParams;
}

function getBudgetErrorMessage(error: unknown) {
  if (isAxiosError<{ message?: string }>(error)) {
    return (
      error.response?.data?.message ??
      "Nao foi possivel carregar os orcamentos."
    );
  }

  return "Nao foi possivel carregar os orcamentos.";
}

function getBudgetDeleteErrorMessage(error: unknown) {
  if (isAxiosError<{ message?: string }>(error)) {
    return (
      error.response?.data?.message ?? "Nao foi possivel excluir o orcamento."
    );
  }

  return "Nao foi possivel excluir o orcamento.";
}

function formatOptionalCurrency(value: number | null) {
  if (value === null) {
    return "Nao informado";
  }

  return currencyFormatter.format(value);
}

function formatOptionalText(value: string) {
  return value.trim() ? value : "Nao informado";
}

function createCatalogMap(items: BudgetCatalogItem[]) {
  return new Map(items.map((item) => [item.id, item.name]));
}

function formatCatalogName(
  value: number | null,
  catalogMap: Map<number, string>,
  fallbackWhenMissing = "Nao informado",
) {
  if (value === null) {
    return fallbackWhenMissing;
  }

  return catalogMap.get(value) ?? `ID ${value}`;
}

function formatResolvedCatalogName(
  value: number | null,
  explicitName: string | null,
  catalogMap: Map<number, string>,
  fallbackWhenMissing = "Nao informado",
) {
  if (explicitName !== null && explicitName.trim()) {
    return explicitName;
  }

  return formatCatalogName(value, catalogMap, fallbackWhenMissing);
}

const tableHeadCellSx = {
  backgroundColor: "rgba(37, 99, 235, 0.08)",
  borderBottomColor: "primary.main",
  borderBottomWidth: 2,
  color: "text.primary",
  fontSize: "0.75rem",
  fontWeight: 700,
  letterSpacing: "0.04em",
  py: 1.5,
  textTransform: "uppercase",
  whiteSpace: "nowrap",
};

const tableDetailCellSx = {
  color: "text.secondary",
  fontSize: "0.78rem",
  lineHeight: 1.45,
  py: 1.25,
  verticalAlign: "top",
};

const stickyIdColumnWidth = 72;
const stickyBudgetColumnWidth = 150;
const frozenColumnsWidth = stickyIdColumnWidth + stickyBudgetColumnWidth;
const tableMaxHeight = "calc(100vh - 280px)";

const compactFilterFieldSx = {
  width: "100%",
  "@media (min-width:900px)": {
    width: "auto",
  },
};

function resetTableRowHeight(row: HTMLTableRowElement | null) {
  if (!row) {
    return;
  }

  row.style.height = "";
}

function applyTableRowHeight(row: HTMLTableRowElement | null, height: number) {
  if (!row) {
    return;
  }

  row.style.height = `${height}px`;
}

export function BudgetListPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { user } = useAuth();
  const [searchParams, setSearchParams] = useSearchParams();
  const isAdmin = user?.role === "admin";
  const filters = useMemo(
    () => getFiltersFromSearchParams(searchParams),
    [searchParams],
  );
  const effectiveFilters = useMemo(
    () =>
      isAdmin
        ? filters
        : {
            ...filters,
            salespersonId: "",
          },
    [filters, isAdmin],
  );
  const [budgetPendingDelete, setBudgetPendingDelete] = useState<{
    id: number;
    budgetNumber: string;
  } | null>(null);
  const [deleteError, setDeleteError] = useState<string | null>(null);
  const frozenTableContainerRef = useRef<HTMLDivElement | null>(null);
  const mainTableContainerRef = useRef<HTMLDivElement | null>(null);
  const frozenHeaderRowRef = useRef<HTMLTableRowElement | null>(null);
  const mainHeaderRowRef = useRef<HTMLTableRowElement | null>(null);
  const frozenBodyRowRefs = useRef<Array<HTMLTableRowElement | null>>([]);
  const mainBodyRowRefs = useRef<Array<HTMLTableRowElement | null>>([]);
  const scrollSyncSourceRef = useRef<"frozen" | "main" | null>(null);
  const [draftFilters, setDraftFilters] = useState(() => ({
    budgetNumber: effectiveFilters.budgetNumber,
    yearBudget: effectiveFilters.yearBudget,
    statusId: effectiveFilters.statusId,
    installerId: effectiveFilters.installerId,
    salespersonId: effectiveFilters.salespersonId,
  }));

  useEffect(() => {
    if (isAdmin || !filters.salespersonId) {
      return;
    }

    setSearchParams(buildSearchParams(effectiveFilters), { replace: true });
  }, [effectiveFilters, filters.salespersonId, isAdmin, setSearchParams]);

  const budgetListQuery = useQuery({
    queryKey: ["budgets", user?.id ?? "anonymous", effectiveFilters],
    queryFn: () => getBudgetListRequest(effectiveFilters),
    placeholderData: keepPreviousData,
  });
  const budgetCatalogsQuery = useQuery({
    queryKey: ["budget-catalogs", user?.id ?? "anonymous"],
    queryFn: isAdmin ? getBudgetCatalogsRequest : getBudgetListCatalogsRequest,
    staleTime: 1000 * 60 * 5,
  });
  const deleteBudgetMutation = useMutation({
    mutationFn: deleteBudgetRequest,
    onSuccess: async () => {
      const shouldGoToPreviousPage =
        effectiveFilters.page > 1 &&
        (budgetListQuery.data?.items.length ?? 0) === 1;

      if (shouldGoToPreviousPage) {
        setSearchParams(
          buildSearchParams({
            ...effectiveFilters,
            page: effectiveFilters.page - 1,
          }),
        );
      }

      await queryClient.invalidateQueries({ queryKey: ["budgets"] });
      setBudgetPendingDelete(null);
      setDeleteError(null);
    },
  });
  const statusMap = useMemo(
    () => createCatalogMap(budgetCatalogsQuery.data?.statuses ?? []),
    [budgetCatalogsQuery.data?.statuses],
  );
  const priorityMap = useMemo(
    () => createCatalogMap(budgetCatalogsQuery.data?.priorities ?? []),
    [budgetCatalogsQuery.data?.priorities],
  );
  const installerMap = useMemo(
    () => createCatalogMap(budgetCatalogsQuery.data?.installers ?? []),
    [budgetCatalogsQuery.data?.installers],
  );
  const projectMap = useMemo(
    () => createCatalogMap(budgetCatalogsQuery.data?.projects ?? []),
    [budgetCatalogsQuery.data?.projects],
  );
  const salespersonMap = useMemo(
    () => createCatalogMap(budgetCatalogsQuery.data?.salespeople ?? []),
    [budgetCatalogsQuery.data?.salespeople],
  );
  const contactMap = useMemo(
    () => createCatalogMap(budgetCatalogsQuery.data?.contacts ?? []),
    [budgetCatalogsQuery.data?.contacts],
  );
  const lossReasonMap = useMemo(
    () => createCatalogMap(budgetCatalogsQuery.data?.lossReasons ?? []),
    [budgetCatalogsQuery.data?.lossReasons],
  );

  const totalPages = Math.max(
    1,
    Math.ceil((budgetListQuery.data?.total ?? 0) / filters.pageSize),
  );
  const budgetItems = useMemo(
    () => budgetListQuery.data?.items ?? [],
    [budgetListQuery.data?.items],
  );

  useEffect(() => {
    frozenBodyRowRefs.current = frozenBodyRowRefs.current.slice(
      0,
      budgetItems.length,
    );
    mainBodyRowRefs.current = mainBodyRowRefs.current.slice(
      0,
      budgetItems.length,
    );
  }, [budgetItems.length]);

  useLayoutEffect(() => {
    const syncRowHeights = () => {
      resetTableRowHeight(frozenHeaderRowRef.current);
      resetTableRowHeight(mainHeaderRowRef.current);
      frozenBodyRowRefs.current.forEach((row) => resetTableRowHeight(row));
      mainBodyRowRefs.current.forEach((row) => resetTableRowHeight(row));

      const headerHeight = Math.max(
        frozenHeaderRowRef.current?.getBoundingClientRect().height ?? 0,
        mainHeaderRowRef.current?.getBoundingClientRect().height ?? 0,
      );

      if (headerHeight > 0) {
        applyTableRowHeight(frozenHeaderRowRef.current, headerHeight);
        applyTableRowHeight(mainHeaderRowRef.current, headerHeight);
      }

      budgetItems.forEach((_, index) => {
        const frozenRow = frozenBodyRowRefs.current[index] ?? null;
        const mainRow = mainBodyRowRefs.current[index] ?? null;
        const rowHeight = Math.max(
          frozenRow?.getBoundingClientRect().height ?? 0,
          mainRow?.getBoundingClientRect().height ?? 0,
        );

        if (rowHeight > 0) {
          applyTableRowHeight(frozenRow, rowHeight);
          applyTableRowHeight(mainRow, rowHeight);
        }
      });
    };

    syncRowHeights();
    window.addEventListener("resize", syncRowHeights);

    return () => {
      window.removeEventListener("resize", syncRowHeights);
    };
  }, [budgetCatalogsQuery.data, budgetItems, isAdmin]);

  const handleMainTableScroll = (event: UIEvent<HTMLDivElement>) => {
    if (scrollSyncSourceRef.current === "frozen") {
      scrollSyncSourceRef.current = null;
      return;
    }

    if (!frozenTableContainerRef.current) {
      return;
    }

    scrollSyncSourceRef.current = "main";
    frozenTableContainerRef.current.scrollTop = event.currentTarget.scrollTop;
    requestAnimationFrame(() => {
      if (scrollSyncSourceRef.current === "main") {
        scrollSyncSourceRef.current = null;
      }
    });
  };

  const handleFrozenTableScroll = (event: UIEvent<HTMLDivElement>) => {
    if (scrollSyncSourceRef.current === "main") {
      scrollSyncSourceRef.current = null;
      return;
    }

    if (!mainTableContainerRef.current) {
      return;
    }

    scrollSyncSourceRef.current = "frozen";
    mainTableContainerRef.current.scrollTop = event.currentTarget.scrollTop;
    requestAnimationFrame(() => {
      if (scrollSyncSourceRef.current === "frozen") {
        scrollSyncSourceRef.current = null;
      }
    });
  };

  const handleDraftChange = (
    field: keyof typeof draftFilters,
    value: string,
  ) => {
    setDraftFilters((currentDraft) => ({
      ...currentDraft,
      [field]: value,
    }));
  };

  const handleApplyFilters = () => {
    const nextFilters: BudgetListFilters = {
      ...effectiveFilters,
      ...draftFilters,
      page: 1,
      salespersonId: isAdmin ? draftFilters.salespersonId : "",
    };

    setSearchParams(buildSearchParams(nextFilters));
  };

  const handleClearFilters = () => {
    setDraftFilters({
      budgetNumber: defaultFilters.budgetNumber,
      yearBudget: defaultFilters.yearBudget,
      statusId: defaultFilters.statusId,
      installerId: defaultFilters.installerId,
      salespersonId: isAdmin ? defaultFilters.salespersonId : "",
    });
    setSearchParams(
      buildSearchParams({
        ...defaultFilters,
        salespersonId: isAdmin ? defaultFilters.salespersonId : "",
      }),
    );
  };

  const handlePageChange = (
    _event: React.ChangeEvent<unknown>,
    value: number,
  ) => {
    setSearchParams(
      buildSearchParams({
        ...effectiveFilters,
        page: value,
      }),
    );
  };

  const handleSortByChange = (value: BudgetSortBy) => {
    const nextFilters: BudgetListFilters = {
      ...effectiveFilters,
      page: 1,
      sortBy: value,
    };

    setSearchParams(buildSearchParams(nextFilters));
  };

  const handleSortOrderChange = (value: BudgetSortOrder) => {
    const nextFilters: BudgetListFilters = {
      ...effectiveFilters,
      page: 1,
      sortOrder: value,
    };

    setSearchParams(buildSearchParams(nextFilters));
  };

  const handleOpenDeleteDialog = (budgetId: number, budgetNumber: string) => {
    setDeleteError(null);
    setBudgetPendingDelete({
      id: budgetId,
      budgetNumber,
    });
  };

  const handleCloseDeleteDialog = () => {
    if (deleteBudgetMutation.isPending) {
      return;
    }

    setBudgetPendingDelete(null);
    setDeleteError(null);
  };

  const handleConfirmDelete = async () => {
    if (budgetPendingDelete === null) {
      return;
    }

    try {
      setDeleteError(null);
      await deleteBudgetMutation.mutateAsync(budgetPendingDelete.id);
    } catch (error) {
      setDeleteError(getBudgetDeleteErrorMessage(error));
    }
  };

  return (
    <Box sx={{ display: "flex", flexDirection: "column", gap: 3, minWidth: 0 }}>
      {isAdmin ? (
        <Box
          sx={{
            display: "flex",
            flexDirection: { sm: "row", xs: "column" },
            gap: 1.5,
            justifyContent: "flex-start",
            mt: { md: 1.5, xs: 1 },
            width: "100%",
          }}
        >
          <Button
            onClick={() => navigate("/budgets/import")}
            startIcon={<UploadFileRoundedIcon />}
            variant="outlined"
          >
            Importar planilha
          </Button>
          <Button
            onClick={() => navigate("/budgets/new")}
            startIcon={<AddRoundedIcon />}
            variant="contained"
          >
            Novo orçamento
          </Button>
        </Box>
      ) : null}

      <SectionCard
        description="Use os filtros abaixo para consultar budgets reais na API."
        title="Filtros"
      >
        <Box
          sx={{
            display: "grid",
            gap: 2,
            gridTemplateColumns: {
              lg: "minmax(240px, 300px) minmax(100px, 120px) minmax(130px, 160px) minmax(140px, 180px) minmax(140px, 180px)",
              md: "repeat(2, minmax(0, 1fr))",
              xs: "minmax(0, 1fr)",
            },
            justifyContent: "flex-start",
          }}
        >
          <TextField
            label="Nro. orçamento"
            onChange={(event) =>
              handleDraftChange("budgetNumber", event.target.value)
            }
            placeholder="Ex: BGT-2026-001"
            size="small"
            sx={compactFilterFieldSx}
            value={draftFilters.budgetNumber}
          />
          <TextField
            label="Ano"
            onChange={(event) =>
              handleDraftChange("yearBudget", event.target.value)
            }
            placeholder="Ex: 2026"
            size="small"
            slotProps={{ htmlInput: { min: 0 } }}
            sx={compactFilterFieldSx}
            type="number"
            value={draftFilters.yearBudget}
          />
          <TextField
            label="Status"
            onChange={(event) =>
              handleDraftChange("statusId", event.target.value)
            }
            select
            size="small"
            sx={compactFilterFieldSx}
            value={draftFilters.statusId}
          >
            <MenuItem value="">Todos</MenuItem>
            {(budgetCatalogsQuery.data?.statuses ?? []).map((status) => (
              <MenuItem key={status.id} value={String(status.id)}>
                {status.name}
              </MenuItem>
            ))}
          </TextField>
          <TextField
            label="Instalador"
            onChange={(event) =>
              handleDraftChange("installerId", event.target.value)
            }
            select
            size="small"
            sx={compactFilterFieldSx}
            value={draftFilters.installerId}
          >
            <MenuItem value="">Todos</MenuItem>
            {(budgetCatalogsQuery.data?.installers ?? []).map((installer) => (
              <MenuItem key={installer.id} value={String(installer.id)}>
                {installer.name}
              </MenuItem>
            ))}
          </TextField>
          {isAdmin ? (
            <TextField
              label="Vendedor"
              onChange={(event) =>
                handleDraftChange("salespersonId", event.target.value)
              }
              select
              size="small"
              sx={compactFilterFieldSx}
              value={draftFilters.salespersonId}
            >
              <MenuItem value="">Todos</MenuItem>
              {(budgetCatalogsQuery.data?.salespeople ?? []).map(
                (salesperson) => (
                  <MenuItem key={salesperson.id} value={String(salesperson.id)}>
                    {salesperson.name}
                  </MenuItem>
                ),
              )}
            </TextField>
          ) : null}
        </Box>

        <Box
          sx={{
            display: "grid",
            gap: 2,
            gridTemplateColumns: {
              lg: "minmax(180px, 220px) minmax(160px, 180px) auto auto",
              md: "repeat(2, minmax(0, 1fr))",
              xs: "minmax(0, 1fr)",
            },
            justifyContent: "flex-start",
          }}
        >
          <TextField
            label="Ordenar por"
            onChange={(event) =>
              handleSortByChange(event.target.value as BudgetSortBy)
            }
            select
            size="small"
            sx={compactFilterFieldSx}
            value={filters.sortBy}
          >
            <MenuItem value="sent_at">Data de envio</MenuItem>
            <MenuItem value="gross_value">Valor bruto</MenuItem>
            <MenuItem value="created_at">Criado em</MenuItem>
            <MenuItem value="updated_at">Atualizado em</MenuItem>
            <MenuItem value="year_budget">Ano</MenuItem>
            <MenuItem value="budget_number">Numero</MenuItem>
          </TextField>
          <TextField
            label="Ordem"
            onChange={(event) =>
              handleSortOrderChange(event.target.value as BudgetSortOrder)
            }
            select
            size="small"
            sx={compactFilterFieldSx}
            value={filters.sortOrder}
          >
            <MenuItem value="desc">Decrescente</MenuItem>
            <MenuItem value="asc">Crescente</MenuItem>
          </TextField>
          <Button
            onClick={handleApplyFilters}
            startIcon={<SearchRoundedIcon />}
            sx={{ minWidth: 140 }}
            variant="contained"
          >
            Filtrar
          </Button>
          <Button
            onClick={handleClearFilters}
            sx={{ minWidth: 120 }}
            variant="outlined"
          >
            Limpar
          </Button>
        </Box>
      </SectionCard>

      <SectionCard>
        {budgetListQuery.isLoading ? (
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

        {budgetListQuery.isError ? (
          <Alert severity="error">
            {getBudgetErrorMessage(budgetListQuery.error)}
          </Alert>
        ) : null}

        {budgetCatalogsQuery.isError ? (
          <Alert severity="warning" variant="outlined">
            Nao foi possivel carregar alguns catalogos. Alguns campos podem
            aparecer com ID.
          </Alert>
        ) : null}

        {deleteError ? <Alert severity="error">{deleteError}</Alert> : null}

        {!budgetListQuery.isLoading && !budgetListQuery.isError ? (
          <Box
            sx={{
              display: "flex",
              flexDirection: "column",
              gap: 2,
              minWidth: 0,
            }}
          >
            <Box
              sx={{
                alignItems: { md: "center", xs: "flex-start" },
                display: "flex",
                flexDirection: { md: "row", xs: "column" },
                gap: 1,
                justifyContent: "flex-end",
              }}
            >
              <Typography color="text.secondary" variant="body2">
                Pagina {budgetListQuery.data?.page ?? filters.page} de{" "}
                {totalPages}
              </Typography>
            </Box>

            {budgetListQuery.data?.items.length ? (
              <Box
                sx={{
                  border: (theme) => `1px solid ${theme.palette.divider}`,
                  borderRadius: 3,
                  display: "flex",
                  minWidth: 0,
                  overflow: "hidden",
                  width: "100%",
                }}
              >
                <TableContainer
                  onScroll={handleFrozenTableScroll}
                  ref={frozenTableContainerRef}
                  sx={{
                    borderRight: (theme) =>
                      `1px solid ${theme.palette.divider}`,
                    flex: "0 0 auto",
                    maxHeight: tableMaxHeight,
                    overflowX: "hidden",
                    overflowY: "auto",
                    scrollbarWidth: "none",
                    width: frozenColumnsWidth,
                    "&::-webkit-scrollbar": {
                      display: "none",
                    },
                  }}
                >
                  <Table
                    size="small"
                    stickyHeader
                    sx={{
                      borderCollapse: "separate",
                      borderSpacing: 0,
                      tableLayout: "fixed",
                      width: frozenColumnsWidth,
                      "& .MuiTableCell-stickyHeader": {
                        backgroundColor: "background.paper",
                        top: 0,
                        zIndex: 2,
                      },
                    }}
                  >
                    <TableHead>
                      <TableRow ref={frozenHeaderRowRef}>
                        <TableCell
                          sx={{
                            ...tableHeadCellSx,
                            minWidth: stickyIdColumnWidth,
                            width: stickyIdColumnWidth,
                          }}
                        >
                          ID
                        </TableCell>
                        <TableCell
                          sx={{
                            ...tableHeadCellSx,
                            minWidth: stickyBudgetColumnWidth,
                            width: stickyBudgetColumnWidth,
                          }}
                        >
                          Orçamento
                        </TableCell>
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {budgetItems.map((budget, index) => (
                        <TableRow
                          hover
                          key={budget.id}
                          ref={(element) => {
                            frozenBodyRowRefs.current[index] = element;
                          }}
                        >
                          <TableCell
                            sx={{ ...tableDetailCellSx, whiteSpace: "nowrap" }}
                          >
                            {budget.id}
                          </TableCell>
                          <TableCell
                            sx={{
                              ...tableDetailCellSx,
                              whiteSpace: "nowrap",
                            }}
                          >
                            <Box
                              sx={{
                                display: "flex",
                                flexDirection: "column",
                                gap: 0.5,
                              }}
                            >
                              <Typography
                                sx={{ fontSize: "0.8rem", fontWeight: 600 }}
                                variant="body2"
                              >
                                {budget.budgetNumber}
                              </Typography>
                            </Box>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </TableContainer>

                <TableContainer
                  onScroll={handleMainTableScroll}
                  ref={mainTableContainerRef}
                  sx={{
                    flex: 1,
                    maxHeight: tableMaxHeight,
                    minWidth: 0,
                    overflow: "auto",
                  }}
                >
                  <Table
                    size="small"
                    stickyHeader
                    sx={{
                      borderCollapse: "separate",
                      borderSpacing: 0,
                      minWidth: 1600,
                      "& .MuiTableCell-stickyHeader": {
                        backgroundColor: "background.paper",
                        top: 0,
                        zIndex: 2,
                      },
                    }}
                  >
                    <TableHead>
                      <TableRow ref={mainHeaderRowRef}>
                        <TableCell sx={tableHeadCellSx}>Ano</TableCell>
                        <TableCell sx={tableHeadCellSx}>Revisão</TableCell>
                        <TableCell sx={tableHeadCellSx}>Envio</TableCell>
                        <TableCell sx={tableHeadCellSx}>Status</TableCell>
                        <TableCell sx={tableHeadCellSx}>Prioridade</TableCell>
                        <TableCell sx={tableHeadCellSx}>Instalador</TableCell>
                        <TableCell sx={tableHeadCellSx}>Projeto</TableCell>
                        <TableCell sx={tableHeadCellSx}>Vendedor</TableCell>
                        <TableCell sx={tableHeadCellSx}>Contato</TableCell>
                        <TableCell sx={tableHeadCellSx}>
                          Motivo de perda
                        </TableCell>
                        <TableCell sx={tableHeadCellSx}>Designer</TableCell>
                        <TableCell sx={tableHeadCellSx}>Concorrente</TableCell>
                        <TableCell align="right" sx={tableHeadCellSx}>
                          Preço concorrente
                        </TableCell>
                        <TableCell sx={tableHeadCellSx}>
                          Especificações
                        </TableCell>
                        <TableCell sx={tableHeadCellSx}>
                          Follow-up atual
                        </TableCell>
                        <TableCell align="right" sx={tableHeadCellSx}>
                          Área m²
                        </TableCell>
                        <TableCell align="right" sx={tableHeadCellSx}>
                          Comissão
                        </TableCell>
                        <TableCell align="right" sx={tableHeadCellSx}>
                          Valor bruto
                        </TableCell>
                        <TableCell sx={tableHeadCellSx}>Criado em</TableCell>
                        <TableCell sx={tableHeadCellSx}>
                          Atualizado em
                        </TableCell>
                        {isAdmin ? (
                          <TableCell sx={tableHeadCellSx}>Ações</TableCell>
                        ) : null}
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {budgetItems.map((budget, index) => (
                        <TableRow
                          hover
                          key={budget.id}
                          ref={(element) => {
                            mainBodyRowRefs.current[index] = element;
                          }}
                        >
                          <TableCell
                            sx={{ ...tableDetailCellSx, whiteSpace: "nowrap" }}
                          >
                            {budget.yearBudget}
                          </TableCell>
                          <TableCell
                            sx={{ ...tableDetailCellSx, whiteSpace: "nowrap" }}
                          >
                            {budget.revision}
                          </TableCell>
                          <TableCell sx={tableDetailCellSx}>
                            {dateFormatter.format(new Date(budget.sentAt))}
                          </TableCell>
                          <TableCell sx={tableDetailCellSx}>
                            <Chip
                              color="primary"
                              label={formatCatalogName(
                                budget.statusId,
                                statusMap,
                              )}
                              size="small"
                              sx={{ fontSize: "0.72rem", height: 24 }}
                              variant="outlined"
                            />
                          </TableCell>
                          <TableCell
                            sx={{ ...tableDetailCellSx, whiteSpace: "nowrap" }}
                          >
                            {formatCatalogName(budget.priorityId, priorityMap)}
                          </TableCell>
                          <TableCell sx={tableDetailCellSx}>
                            {formatCatalogName(
                              budget.installerId,
                              installerMap,
                              "Sem instalador vinculado",
                            )}
                          </TableCell>
                          <TableCell
                            sx={{ ...tableDetailCellSx, whiteSpace: "nowrap" }}
                          >
                            {formatResolvedCatalogName(
                              budget.projectId,
                              budget.projectName,
                              projectMap,
                            )}
                          </TableCell>
                          <TableCell
                            sx={{ ...tableDetailCellSx, whiteSpace: "nowrap" }}
                          >
                            {formatResolvedCatalogName(
                              budget.salespersonId,
                              budget.salespersonName,
                              salespersonMap,
                            )}
                          </TableCell>
                          <TableCell
                            sx={{ ...tableDetailCellSx, whiteSpace: "nowrap" }}
                          >
                            {formatResolvedCatalogName(
                              budget.contactId,
                              budget.contactName,
                              contactMap,
                            )}
                          </TableCell>
                          <TableCell
                            sx={{ ...tableDetailCellSx, whiteSpace: "nowrap" }}
                          >
                            {formatCatalogName(
                              budget.lossReasonId,
                              lossReasonMap,
                            )}
                          </TableCell>
                          <TableCell sx={tableDetailCellSx}>
                            {formatOptionalText(budget.designerName)}
                          </TableCell>
                          <TableCell sx={tableDetailCellSx}>
                            {formatOptionalText(budget.competitorName)}
                          </TableCell>
                          <TableCell
                            align="right"
                            sx={{ ...tableDetailCellSx, whiteSpace: "nowrap" }}
                          >
                            {formatOptionalCurrency(budget.competitorPrice)}
                          </TableCell>
                          <TableCell
                            sx={{ ...tableDetailCellSx, minWidth: 220 }}
                          >
                            {formatOptionalText(budget.specificationDetails)}
                          </TableCell>
                          <TableCell
                            sx={{ ...tableDetailCellSx, minWidth: 220 }}
                          >
                            {formatOptionalText(budget.currentFollowUp)}
                          </TableCell>
                          <TableCell
                            align="right"
                            sx={{ ...tableDetailCellSx, whiteSpace: "nowrap" }}
                          >
                            {decimalFormatter.format(budget.areaM2)}
                          </TableCell>
                          <TableCell
                            align="right"
                            sx={{ ...tableDetailCellSx, whiteSpace: "nowrap" }}
                          >
                            {currencyFormatter.format(budget.commissionValue)}
                          </TableCell>
                          <TableCell align="right" sx={tableDetailCellSx}>
                            {currencyFormatter.format(budget.grossValue)}
                          </TableCell>
                          <TableCell
                            sx={{ ...tableDetailCellSx, whiteSpace: "nowrap" }}
                          >
                            {dateTimeFormatter.format(
                              new Date(budget.createdAt),
                            )}
                          </TableCell>
                          <TableCell
                            sx={{ ...tableDetailCellSx, whiteSpace: "nowrap" }}
                          >
                            {dateTimeFormatter.format(
                              new Date(budget.updatedAt),
                            )}
                          </TableCell>
                          {isAdmin ? (
                            <TableCell
                              sx={{
                                ...tableDetailCellSx,
                                whiteSpace: "nowrap",
                              }}
                            >
                              <Button
                                onClick={() =>
                                  navigate(`/budgets/${budget.id}/edit`)
                                }
                                size="small"
                                startIcon={<EditRoundedIcon />}
                                variant="text"
                              >
                                Editar
                              </Button>
                              <Button
                                color="error"
                                onClick={() =>
                                  handleOpenDeleteDialog(
                                    budget.id,
                                    budget.budgetNumber,
                                  )
                                }
                                size="small"
                                startIcon={<DeleteOutlineRoundedIcon />}
                                variant="text"
                              >
                                Excluir
                              </Button>
                            </TableCell>
                          ) : null}
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </TableContainer>
              </Box>
            ) : (
              <Alert severity="info" variant="outlined">
                Nenhum orcamento encontrado para os filtros informados.
              </Alert>
            )}

            <Box
              sx={{
                alignItems: { md: "center", xs: "flex-start" },
                display: "flex",
                flexDirection: { md: "row", xs: "column" },
                gap: 1.5,
                justifyContent: "space-between",
                pt: 1,
              }}
            >
              <Typography color="text.secondary" variant="body2">
                {budgetListQuery.data?.total ?? 0} resultado(s) encontrado(s)
              </Typography>
              <Pagination
                color="primary"
                count={totalPages}
                onChange={handlePageChange}
                page={budgetListQuery.data?.page ?? filters.page}
                shape="rounded"
              />
            </Box>
          </Box>
        ) : null}
      </SectionCard>

      <Dialog
        onClose={handleCloseDeleteDialog}
        open={budgetPendingDelete !== null}
      >
        <DialogTitle>Excluir orçamento</DialogTitle>
        <DialogContent>
          <DialogContentText>
            Confirma a exclusao do orçamento{" "}
            <strong>{budgetPendingDelete?.budgetNumber ?? ""}</strong>?
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button
            disabled={deleteBudgetMutation.isPending}
            onClick={handleCloseDeleteDialog}
            variant="outlined"
          >
            Cancelar
          </Button>
          <Button
            color="error"
            disabled={deleteBudgetMutation.isPending}
            onClick={handleConfirmDelete}
            variant="contained"
          >
            {deleteBudgetMutation.isPending ? "Excluindo..." : "Excluir"}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
