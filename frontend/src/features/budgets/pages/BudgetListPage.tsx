import AddRoundedIcon from "@mui/icons-material/AddRounded";
import ApartmentRoundedIcon from "@mui/icons-material/ApartmentRounded";
import DeleteOutlineRoundedIcon from "@mui/icons-material/DeleteOutlineRounded";
import EditRoundedIcon from "@mui/icons-material/EditRounded";
import ExpandMoreRoundedIcon from "@mui/icons-material/ExpandMoreRounded";
import VisibilityRoundedIcon from "@mui/icons-material/VisibilityRounded";
import SearchRoundedIcon from "@mui/icons-material/SearchRounded";
import TableRowsRoundedIcon from "@mui/icons-material/TableRowsRounded";
import UploadFileRoundedIcon from "@mui/icons-material/UploadFileRounded";
import {
  Accordion,
  AccordionDetails,
  AccordionSummary,
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
  ToggleButton,
  ToggleButtonGroup,
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
  getBudgetProjectListRequest,
} from "../api/budgets";
import type {
  BudgetCatalogItem,
  BudgetListFilters,
  BudgetListItem,
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
  sourceCompany: "",
  yearBudget: "",
  statusId: "",
  installerId: "",
  projectName: "",
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
    sourceCompany:
      searchParams.get("sourceCompany") ?? defaultFilters.sourceCompany,
    yearBudget: searchParams.get("yearBudget") ?? defaultFilters.yearBudget,
    statusId: searchParams.get("statusId") ?? defaultFilters.statusId,
    installerId: searchParams.get("installerId") ?? defaultFilters.installerId,
    projectName: searchParams.get("projectName") ?? defaultFilters.projectName,
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
  if (filters.sourceCompany) {
    nextSearchParams.set("sourceCompany", filters.sourceCompany);
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
  if (filters.projectName) {
    nextSearchParams.set("projectName", filters.projectName);
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

function normalizeText(value: string | null | undefined) {
  return (value ?? "")
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "")
    .trim()
    .toLowerCase();
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

function hasProjectAssociation(
  budget: Pick<BudgetListItem, "projectId" | "projectName">,
) {
  if (budget.projectId === null) {
    return false;
  }

  return normalizeText(budget.projectName) !== "nao informado";
}

type BudgetViewMode = "project" | "list";

type BudgetStatusCategory = "pedido" | "cancelado" | "orcamento" | "other";

type BudgetProjectGroup = {
  key: string;
  projectId: number | null;
  projectName: string;
  items: BudgetListItem[];
  latestActivityAt: string;
  winnerBudgetId: number | null;
  cancelledCount: number;
  activeCount: number;
  needsAttention: boolean;
};

function normalizeValue(value: string | null | undefined) {
  return (value ?? "")
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "")
    .trim()
    .toLowerCase();
}

function getBudgetStatusCategory(statusName: string): BudgetStatusCategory {
  const normalizedStatusName = normalizeValue(statusName);

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

const tableHeadCellSx = {
  backgroundColor: "rgba(37, 99, 235, 0.08)",
  borderBottomColor: "primary.main",
  borderBottomStyle: "solid",
  borderBottomWidth: 2,
  color: "text.primary",
  fontSize: "0.72rem",
  fontWeight: 700,
  letterSpacing: "0.04em",
  py: 1,
  textTransform: "uppercase",
  whiteSpace: "nowrap",
};

const tableDetailCellSx = {
  borderBottomColor: "divider",
  borderBottomStyle: "solid",
  borderBottomWidth: 1,
  color: "text.secondary",
  fontSize: "0.74rem",
  lineHeight: 1.25,
  py: 0.7,
  verticalAlign: "middle",
};

const singleLineTableCellSx = {
  ...tableDetailCellSx,
  maxWidth: 220,
  overflow: "hidden",
  textOverflow: "ellipsis",
  whiteSpace: "nowrap",
};

const frozenBudgetTableCellSx = {
  ...singleLineTableCellSx,
  height: "100%",
  px: 1.5,
  py: 2.1,
  textAlign: "center",
};

const stickyBudgetColumnWidth = 140;
const frozenColumnsWidth = stickyBudgetColumnWidth;
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
  const frozenTableRef = useRef<HTMLTableElement | null>(null);
  const mainTableRef = useRef<HTMLTableElement | null>(null);
  const frozenHeaderRowRef = useRef<HTMLTableRowElement | null>(null);
  const mainHeaderRowRef = useRef<HTMLTableRowElement | null>(null);
  const frozenBodyRowRefs = useRef<Array<HTMLTableRowElement | null>>([]);
  const mainBodyRowRefs = useRef<Array<HTMLTableRowElement | null>>([]);
  const scrollSyncSourceRef = useRef<"frozen" | "main" | null>(null);
  const [draftFilters, setDraftFilters] = useState(() => ({
    budgetNumber: effectiveFilters.budgetNumber,
    sourceCompany: effectiveFilters.sourceCompany,
    yearBudget: effectiveFilters.yearBudget,
    statusId: effectiveFilters.statusId,
    installerId: effectiveFilters.installerId,
    projectName: effectiveFilters.projectName,
    salespersonId: effectiveFilters.salespersonId,
  }));
  const [viewMode, setViewMode] = useState<BudgetViewMode>("project");

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
  const projectBudgetFilters = useMemo(
    () => ({
      ...effectiveFilters,
      page: 1,
      pageSize: 100,
    }),
    [effectiveFilters],
  );
  const projectBudgetListQuery = useQuery({
    enabled: viewMode === "project",
    queryKey: [
      "budgets",
      "project-view",
      user?.id ?? "anonymous",
      projectBudgetFilters,
    ],
    queryFn: () => getBudgetProjectListRequest(projectBudgetFilters),
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
  const projectBudgetItems = useMemo(
    () => projectBudgetListQuery.data?.items ?? [],
    [projectBudgetListQuery.data?.items],
  );
  const projectGroups = useMemo<BudgetProjectGroup[]>(() => {
    const groupedBudgets = projectBudgetItems.reduce<
      Map<string, BudgetProjectGroup>
    >((currentGroups, budget) => {
      if (!hasProjectAssociation(budget)) {
        return currentGroups;
      }

      const projectName = formatResolvedCatalogName(
        budget.projectId,
        budget.projectName,
        projectMap,
        "Projeto nao informado",
      );
      const groupKey = `project-${budget.projectId}`;
      const existingGroup = currentGroups.get(groupKey);

      if (existingGroup) {
        existingGroup.items.push(budget);
        return currentGroups;
      }

      currentGroups.set(groupKey, {
        key: groupKey,
        projectId: budget.projectId,
        projectName,
        items: [budget],
        latestActivityAt: budget.updatedAt,
        winnerBudgetId: null,
        cancelledCount: 0,
        activeCount: 0,
        needsAttention: true,
      });

      return currentGroups;
    }, new Map<string, BudgetProjectGroup>());

    const recentGroups = Array.from(groupedBudgets.values())
      .map((group) => {
        const winner = group.items.find((budget) => {
          const statusLabel = formatCatalogName(budget.statusId, statusMap);
          return getBudgetStatusCategory(statusLabel) === "pedido";
        });
        const cancelledCount = group.items.filter((budget) => {
          const statusLabel = formatCatalogName(budget.statusId, statusMap);
          return getBudgetStatusCategory(statusLabel) === "cancelado";
        }).length;
        const activeCount = group.items.length - cancelledCount;
        const latestActivityAt = group.items.reduce((currentLatest, budget) => {
          if (new Date(budget.updatedAt) > new Date(currentLatest)) {
            return budget.updatedAt;
          }

          return currentLatest;
        }, group.items[0]?.updatedAt ?? group.latestActivityAt);

        return {
          ...group,
          items: [...group.items].sort((firstItem, secondItem) =>
            firstItem.budgetNumber.localeCompare(secondItem.budgetNumber),
          ),
          latestActivityAt,
          winnerBudgetId: winner?.id ?? null,
          cancelledCount,
          activeCount,
          needsAttention: winner === undefined,
        };
      })
      .filter((group) => group.needsAttention)
      .sort((firstGroup, secondGroup) => {
        const latestActivityDifference =
          new Date(secondGroup.latestActivityAt).getTime() -
          new Date(firstGroup.latestActivityAt).getTime();

        if (latestActivityDifference !== 0) {
          return latestActivityDifference;
        }

        if (firstGroup.projectId === null && secondGroup.projectId !== null) {
          return 1;
        }

        if (firstGroup.projectId !== null && secondGroup.projectId === null) {
          return -1;
        }

        return firstGroup.projectName.localeCompare(secondGroup.projectName);
      })
      .slice(0, 20);

    return recentGroups.sort((firstGroup, secondGroup) => {
      const budgetCountDifference =
        secondGroup.items.length - firstGroup.items.length;

      if (budgetCountDifference !== 0) {
        return budgetCountDifference;
      }

      const latestActivityDifference =
        new Date(secondGroup.latestActivityAt).getTime() -
        new Date(firstGroup.latestActivityAt).getTime();

      if (latestActivityDifference !== 0) {
        return latestActivityDifference;
      }

      if (firstGroup.projectId === null && secondGroup.projectId !== null) {
        return 1;
      }

      if (firstGroup.projectId !== null && secondGroup.projectId === null) {
        return -1;
      }

      return firstGroup.projectName.localeCompare(secondGroup.projectName);
    });
  }, [projectBudgetItems, projectMap, statusMap]);
  const groupedProjectsCount = useMemo(
    () => projectGroups.length,
    [projectGroups],
  );
  const isProjectView = viewMode === "project";
  const activeBudgetItems = isProjectView ? projectBudgetItems : budgetItems;
  const activeBudgetQuery = isProjectView
    ? projectBudgetListQuery
    : budgetListQuery;

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
    let isCancelled = false;

    const syncRowHeights = () => {
      if (isCancelled) {
        return;
      }

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

    const scheduleSyncRowHeights = () => {
      requestAnimationFrame(() => {
        requestAnimationFrame(() => {
          syncRowHeights();
        });
      });
    };

    syncRowHeights();
    scheduleSyncRowHeights();
    window.addEventListener("resize", scheduleSyncRowHeights);

    const resizeObserver =
      typeof ResizeObserver === "undefined"
        ? null
        : new ResizeObserver(() => {
            scheduleSyncRowHeights();
          });

    [
      frozenTableContainerRef.current,
      mainTableContainerRef.current,
      frozenTableRef.current,
      mainTableRef.current,
    ].forEach((element) => {
      if (element !== null) {
        resizeObserver?.observe(element);
      }
    });

    if ("fonts" in document) {
      void document.fonts.ready.then(() => {
        scheduleSyncRowHeights();
      });
    }

    return () => {
      isCancelled = true;
      resizeObserver?.disconnect();
      window.removeEventListener("resize", scheduleSyncRowHeights);
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
      sourceCompany: defaultFilters.sourceCompany,
      yearBudget: defaultFilters.yearBudget,
      statusId: defaultFilters.statusId,
      installerId: defaultFilters.installerId,
      projectName: defaultFilters.projectName,
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

  const handleBudgetRowDoubleClick = (budgetId: number) => {
    if (!isAdmin) {
      return;
    }

    navigate(`/budgets/${budgetId}/edit`);
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

  const handleViewModeChange = (
    _event: React.MouseEvent<HTMLElement>,
    nextViewMode: BudgetViewMode | null,
  ) => {
    if (nextViewMode === null) {
      return;
    }

    setViewMode(nextViewMode);
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
              lg: "minmax(220px, 280px) minmax(100px, 120px) minmax(140px, 170px) minmax(130px, 160px) minmax(140px, 180px) minmax(180px, 240px) minmax(140px, 180px)",
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
            label="Empresa"
            onChange={(event) =>
              handleDraftChange("sourceCompany", event.target.value)
            }
            select
            size="small"
            sx={compactFilterFieldSx}
            value={draftFilters.sourceCompany}
          >
            <MenuItem value="">Todas</MenuItem>
            <MenuItem value="Rocktec">ROCKTEC</MenuItem>
            <MenuItem value="Trox">TROX</MenuItem>
          </TextField>
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
          <TextField
            label="Obra"
            onChange={(event) =>
              handleDraftChange("projectName", event.target.value)
            }
            placeholder="Digite parte do nome da obra"
            size="small"
            sx={compactFilterFieldSx}
            value={draftFilters.projectName}
          />
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
        {activeBudgetQuery.isLoading ? (
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

        {activeBudgetQuery.isError ? (
          <Alert severity="error">
            {getBudgetErrorMessage(activeBudgetQuery.error)}
          </Alert>
        ) : null}

        {budgetCatalogsQuery.isError ? (
          <Alert severity="warning" variant="outlined">
            Nao foi possivel carregar alguns catalogos. Alguns campos podem
            aparecer com ID.
          </Alert>
        ) : null}

        {deleteError ? <Alert severity="error">{deleteError}</Alert> : null}

        {!activeBudgetQuery.isLoading && !activeBudgetQuery.isError ? (
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
                alignItems: { lg: "center", xs: "flex-start" },
                display: "flex",
                flexDirection: { lg: "row", xs: "column" },
                gap: 1.5,
                justifyContent: "space-between",
              }}
            >
              <Box>
                {!isProjectView ? (
                  <Typography color="text.secondary" variant="body2">
                    Pagina {budgetListQuery.data?.page ?? filters.page} de{" "}
                    {totalPages}
                  </Typography>
                ) : null}
                <Typography color="text.secondary" variant="body2">
                  {isProjectView
                    ? `${groupedProjectsCount} projeto(s) sem PEDIDO definido entre os 20 mais recentes, com prioridade para quem tem mais orcamentos`
                    : `${budgetListQuery.data?.total ?? 0} orçamento(s) encontrado(s)`}
                </Typography>
              </Box>

              <ToggleButtonGroup
                exclusive
                onChange={handleViewModeChange}
                size="small"
                value={viewMode}
              >
                <ToggleButton value="project">
                  <ApartmentRoundedIcon fontSize="small" sx={{ mr: 1 }} />
                  Por projeto
                </ToggleButton>
                <ToggleButton value="list">
                  <TableRowsRoundedIcon fontSize="small" sx={{ mr: 1 }} />
                  Lista completa
                </ToggleButton>
              </ToggleButtonGroup>
            </Box>

            {isProjectView ? (
              <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
                <Alert severity="info" variant="outlined">
                  Nesta visualizacao aparecem apenas projetos sem{" "}
                  <strong>PEDIDO</strong> definido, limitados aos 20 mais
                  recentes pela ultima atualizacao dos orcamentos do grupo e
                  ordenados no topo por quantidade de orcamentos vinculados.
                </Alert>

                {projectBudgetItems.length === 0 ? (
                  <Alert severity="info" variant="outlined">
                    Nenhum orçamento foi encontrado para os filtros informados.
                  </Alert>
                ) : projectGroups.length === 0 ? (
                  <Alert severity="info" variant="outlined">
                    Foram encontrados{" "}
                    <strong>{projectBudgetItems.length} orçamento(s)</strong> na
                    busca atual, mas todos os projetos ja possuem{" "}
                    <strong>PEDIDO</strong> definido e por isso nao aparecem
                    nesta visualizacao.
                  </Alert>
                ) : (
                  projectGroups.map((group) => (
                    <Accordion
                      defaultExpanded={group.items.length > 1}
                      disableGutters
                      key={group.key}
                      sx={{
                        border: (theme) => `1px solid ${theme.palette.divider}`,
                        borderRadius: "20px !important",
                        overflow: "hidden",
                        "&::before": {
                          display: "none",
                        },
                      }}
                    >
                      <AccordionSummary expandIcon={<ExpandMoreRoundedIcon />}>
                        <Box
                          sx={{
                            alignItems: { md: "center", xs: "flex-start" },
                            display: "flex",
                            flexDirection: { md: "row", xs: "column" },
                            gap: 1.5,
                            justifyContent: "space-between",
                            width: "100%",
                          }}
                        >
                          <Box sx={{ minWidth: 0 }}>
                            <Typography sx={{ fontWeight: 700 }} variant="h6">
                              {group.projectName}
                            </Typography>
                            <Typography color="text.secondary" variant="body2">
                              {`Projeto #${group.projectId} · ${group.items.length} orcamento(s) vinculado(s) · ${group.items[0]?.sourceCompany || "Nao informado"}`}
                            </Typography>
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
                              label={`${group.items.length} orcamento(s)`}
                              size="small"
                              variant="outlined"
                            />
                            <Chip
                              color={
                                group.needsAttention ? "warning" : "success"
                              }
                              label={
                                group.needsAttention
                                  ? "Sem PEDIDO definido"
                                  : "Com vencedor definido"
                              }
                              size="small"
                              variant="outlined"
                            />
                            <Chip
                              label={`${group.cancelledCount} cancelado(s)`}
                              size="small"
                              variant="outlined"
                            />
                            {isAdmin && group.projectId !== null ? (
                              <Button
                                onClick={(event) => {
                                  event.stopPropagation();
                                  navigate(`/projects/${group.projectId}`);
                                }}
                                size="small"
                                startIcon={<VisibilityRoundedIcon />}
                                variant="text"
                              >
                                Abrir projeto
                              </Button>
                            ) : null}
                          </Box>
                        </Box>
                      </AccordionSummary>

                      <AccordionDetails>
                        <Box
                          sx={{
                            display: "grid",
                            gap: 1.5,
                            gridTemplateColumns: "minmax(0, 1fr)",
                          }}
                        >
                          {group.items.map((budget) => {
                            const statusLabel = formatCatalogName(
                              budget.statusId,
                              statusMap,
                            );
                            const statusCategory =
                              getBudgetStatusCategory(statusLabel);
                            const isWinner = group.winnerBudgetId === budget.id;

                            return (
                              <Box
                                key={budget.id}
                                sx={{
                                  backgroundColor: (theme) => {
                                    if (isWinner) {
                                      return theme.palette.success.main + "12";
                                    }

                                    if (statusCategory === "cancelado") {
                                      return theme.palette.grey[500] + "12";
                                    }

                                    return theme.palette.background.paper;
                                  },
                                  border: (theme) => {
                                    if (isWinner) {
                                      return `1px solid ${theme.palette.success.main}`;
                                    }

                                    return `1px solid ${theme.palette.divider}`;
                                  },
                                  borderRadius: 3,
                                  p: 2,
                                }}
                              >
                                <Box
                                  sx={{
                                    alignItems: {
                                      md: "center",
                                      xs: "flex-start",
                                    },
                                    display: "flex",
                                    flexDirection: {
                                      md: "row",
                                      xs: "column",
                                    },
                                    gap: 1.5,
                                    justifyContent: "space-between",
                                    mb: 1.5,
                                  }}
                                >
                                  <Box sx={{ minWidth: 0 }}>
                                    <Typography
                                      sx={{ fontWeight: 700 }}
                                      variant="subtitle1"
                                    >
                                      {budget.budgetNumber}
                                    </Typography>
                                    <Typography
                                      color="text.secondary"
                                      variant="body2"
                                    >
                                      ID {budget.id} · envio{" "}
                                      {dateFormatter.format(
                                        new Date(budget.sentAt),
                                      )}
                                    </Typography>
                                  </Box>

                                  <Box
                                    sx={{
                                      display: "flex",
                                      flexWrap: "wrap",
                                      gap: 1,
                                    }}
                                  >
                                    <Chip
                                      label={
                                        budget.sourceCompany || "Sem origem"
                                      }
                                      size="small"
                                      variant="outlined"
                                    />
                                    <Chip
                                      color={
                                        isWinner
                                          ? "success"
                                          : statusCategory === "cancelado"
                                            ? "default"
                                            : "primary"
                                      }
                                      label={statusLabel}
                                      size="small"
                                      variant={isWinner ? "filled" : "outlined"}
                                    />
                                    {isWinner ? (
                                      <Chip
                                        color="success"
                                        label="Vencedor do projeto"
                                        size="small"
                                      />
                                    ) : null}
                                    {statusCategory === "cancelado" ? (
                                      <Chip
                                        label="Sem necessidade de atencao"
                                        size="small"
                                        variant="outlined"
                                      />
                                    ) : null}
                                  </Box>
                                </Box>

                                <Box
                                  sx={{
                                    display: "grid",
                                    gap: 1.5,
                                    gridTemplateColumns: {
                                      lg: "repeat(4, minmax(0, 1fr))",
                                      md: "repeat(2, minmax(0, 1fr))",
                                      xs: "minmax(0, 1fr)",
                                    },
                                  }}
                                >
                                  <Box>
                                    <Typography
                                      color="text.secondary"
                                      variant="caption"
                                    >
                                      Valor bruto
                                    </Typography>
                                    <Typography variant="body2">
                                      {currencyFormatter.format(
                                        budget.grossValue,
                                      )}
                                    </Typography>
                                  </Box>
                                  <Box>
                                    <Typography
                                      color="text.secondary"
                                      variant="caption"
                                    >
                                      Vendedor
                                    </Typography>
                                    <Typography variant="body2">
                                      {formatResolvedCatalogName(
                                        budget.salespersonId,
                                        budget.salespersonName,
                                        salespersonMap,
                                      )}
                                    </Typography>
                                  </Box>
                                  <Box>
                                    <Typography
                                      color="text.secondary"
                                      variant="caption"
                                    >
                                      Contato
                                    </Typography>
                                    <Typography variant="body2">
                                      {formatResolvedCatalogName(
                                        budget.contactId,
                                        budget.contactName,
                                        contactMap,
                                      )}
                                    </Typography>
                                  </Box>
                                  <Box>
                                    <Typography
                                      color="text.secondary"
                                      variant="caption"
                                    >
                                      Instalador
                                    </Typography>
                                    <Typography variant="body2">
                                      {formatCatalogName(
                                        budget.installerId,
                                        installerMap,
                                        "Sem instalador vinculado",
                                      )}
                                    </Typography>
                                  </Box>
                                  <Box>
                                    <Typography
                                      color="text.secondary"
                                      variant="caption"
                                    >
                                      Prioridade
                                    </Typography>
                                    <Typography variant="body2">
                                      {formatCatalogName(
                                        budget.priorityId,
                                        priorityMap,
                                      )}
                                    </Typography>
                                  </Box>
                                  <Box>
                                    <Typography
                                      color="text.secondary"
                                      variant="caption"
                                    >
                                      Area m2
                                    </Typography>
                                    <Typography variant="body2">
                                      {decimalFormatter.format(budget.areaM2)}
                                    </Typography>
                                  </Box>
                                  <Box>
                                    <Typography
                                      color="text.secondary"
                                      variant="caption"
                                    >
                                      Designer
                                    </Typography>
                                    <Typography variant="body2">
                                      {formatOptionalText(budget.designerName)}
                                    </Typography>
                                  </Box>
                                  <Box>
                                    <Typography
                                      color="text.secondary"
                                      variant="caption"
                                    >
                                      Follow-up atual
                                    </Typography>
                                    <Typography variant="body2">
                                      {formatOptionalText(
                                        budget.currentFollowUp,
                                      )}
                                    </Typography>
                                  </Box>
                                </Box>

                                {isAdmin ? (
                                  <Box
                                    sx={{
                                      display: "flex",
                                      gap: 1,
                                      justifyContent: "flex-end",
                                      mt: 1.5,
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
                                  </Box>
                                ) : null}
                              </Box>
                            );
                          })}
                        </Box>
                      </AccordionDetails>
                    </Accordion>
                  ))
                )}
              </Box>
            ) : activeBudgetItems.length ? (
              <Box
                sx={{
                  alignItems: "stretch",
                  border: (theme) => `1px solid ${theme.palette.divider}`,
                  borderRadius: 3,
                  display: "flex",
                  height: tableMaxHeight,
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
                    boxSizing: "border-box",
                    flex: "0 0 auto",
                    height: "100%",
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
                    ref={frozenTableRef}
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
                            minWidth: stickyBudgetColumnWidth,
                            textAlign: "center",
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
                          onDoubleClick={() =>
                            handleBudgetRowDoubleClick(budget.id)
                          }
                          ref={(element) => {
                            frozenBodyRowRefs.current[index] = element;
                          }}
                          sx={{ cursor: isAdmin ? "pointer" : "default" }}
                        >
                          <TableCell sx={frozenBudgetTableCellSx}>
                            <Box
                              sx={{
                                alignItems: "center",
                                display: "flex",
                                flexDirection: "column",
                                gap: 0.5,
                                height: "100%",
                                justifyContent: "center",
                                minWidth: 0,
                              }}
                            >
                              <Typography
                                noWrap
                                sx={{
                                  fontSize: "0.75rem",
                                  fontWeight: 600,
                                  textAlign: "center",
                                  width: "100%",
                                }}
                                title={budget.budgetNumber}
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
                    boxSizing: "border-box",
                    flex: 1,
                    height: "100%",
                    minWidth: 0,
                    overflow: "auto",
                  }}
                >
                  <Table
                    ref={mainTableRef}
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
                        <TableCell sx={tableHeadCellSx}>Empresa</TableCell>
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
                          onDoubleClick={() =>
                            handleBudgetRowDoubleClick(budget.id)
                          }
                          ref={(element) => {
                            mainBodyRowRefs.current[index] = element;
                          }}
                          sx={{ cursor: isAdmin ? "pointer" : "default" }}
                        >
                          <TableCell sx={singleLineTableCellSx}>
                            {budget.yearBudget}
                          </TableCell>
                          <TableCell
                            sx={singleLineTableCellSx}
                            title={budget.sourceCompany || "Nao informado"}
                          >
                            {budget.sourceCompany || "Nao informado"}
                          </TableCell>
                          <TableCell sx={singleLineTableCellSx}>
                            {budget.revision}
                          </TableCell>
                          <TableCell sx={singleLineTableCellSx}>
                            {dateFormatter.format(new Date(budget.sentAt))}
                          </TableCell>
                          <TableCell sx={tableDetailCellSx}>
                            <Chip
                              color="primary"
                              label={formatResolvedCatalogName(
                                budget.statusId,
                                budget.statusName,
                                statusMap,
                              )}
                              size="small"
                              sx={{ fontSize: "0.68rem", height: 20 }}
                              variant="outlined"
                            />
                          </TableCell>
                          <TableCell
                            sx={singleLineTableCellSx}
                            title={formatResolvedCatalogName(
                              budget.priorityId,
                              budget.priorityName,
                              priorityMap,
                            )}
                          >
                            {formatResolvedCatalogName(
                              budget.priorityId,
                              budget.priorityName,
                              priorityMap,
                            )}
                          </TableCell>
                          <TableCell
                            sx={singleLineTableCellSx}
                            title={formatResolvedCatalogName(
                              budget.installerId,
                              budget.installerName,
                              installerMap,
                              "Sem instalador vinculado",
                            )}
                          >
                            {formatResolvedCatalogName(
                              budget.installerId,
                              budget.installerName,
                              installerMap,
                              "Sem instalador vinculado",
                            )}
                          </TableCell>
                          <TableCell
                            sx={singleLineTableCellSx}
                            title={formatResolvedCatalogName(
                              budget.projectId,
                              budget.projectName,
                              projectMap,
                            )}
                          >
                            {formatResolvedCatalogName(
                              budget.projectId,
                              budget.projectName,
                              projectMap,
                            )}
                          </TableCell>
                          <TableCell
                            sx={singleLineTableCellSx}
                            title={formatResolvedCatalogName(
                              budget.salespersonId,
                              budget.salespersonName,
                              salespersonMap,
                            )}
                          >
                            {formatResolvedCatalogName(
                              budget.salespersonId,
                              budget.salespersonName,
                              salespersonMap,
                            )}
                          </TableCell>
                          <TableCell
                            sx={singleLineTableCellSx}
                            title={formatResolvedCatalogName(
                              budget.contactId,
                              budget.contactName,
                              contactMap,
                            )}
                          >
                            {formatResolvedCatalogName(
                              budget.contactId,
                              budget.contactName,
                              contactMap,
                            )}
                          </TableCell>
                          <TableCell
                            sx={singleLineTableCellSx}
                            title={formatResolvedCatalogName(
                              budget.lossReasonId,
                              budget.lossReasonName,
                              lossReasonMap,
                            )}
                          >
                            {formatResolvedCatalogName(
                              budget.lossReasonId,
                              budget.lossReasonName,
                              lossReasonMap,
                            )}
                          </TableCell>
                          <TableCell
                            sx={singleLineTableCellSx}
                            title={formatOptionalText(budget.designerName)}
                          >
                            {formatOptionalText(budget.designerName)}
                          </TableCell>
                          <TableCell
                            sx={singleLineTableCellSx}
                            title={formatOptionalText(budget.competitorName)}
                          >
                            {formatOptionalText(budget.competitorName)}
                          </TableCell>
                          <TableCell align="right" sx={singleLineTableCellSx}>
                            {formatOptionalCurrency(budget.competitorPrice)}
                          </TableCell>
                          <TableCell
                            sx={{ ...singleLineTableCellSx, minWidth: 220 }}
                            title={formatOptionalText(
                              budget.specificationDetails,
                            )}
                          >
                            {formatOptionalText(budget.specificationDetails)}
                          </TableCell>
                          <TableCell
                            sx={{ ...singleLineTableCellSx, minWidth: 220 }}
                            title={formatOptionalText(budget.currentFollowUp)}
                          >
                            {formatOptionalText(budget.currentFollowUp)}
                          </TableCell>
                          <TableCell align="right" sx={singleLineTableCellSx}>
                            {decimalFormatter.format(budget.areaM2)}
                          </TableCell>
                          <TableCell align="right" sx={singleLineTableCellSx}>
                            {currencyFormatter.format(budget.commissionValue)}
                          </TableCell>
                          <TableCell align="right" sx={singleLineTableCellSx}>
                            {currencyFormatter.format(budget.grossValue)}
                          </TableCell>
                          <TableCell sx={singleLineTableCellSx}>
                            {dateTimeFormatter.format(
                              new Date(budget.createdAt),
                            )}
                          </TableCell>
                          <TableCell sx={singleLineTableCellSx}>
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
                                sx={{ minWidth: "auto", px: 0.75, py: 0.25 }}
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
                                sx={{ minWidth: "auto", px: 0.75, py: 0.25 }}
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
                {isProjectView
                  ? `${projectBudgetListQuery.data?.total ?? 0} orçamento(s) carregado(s) para agrupamento`
                  : `${budgetListQuery.data?.total ?? 0} resultado(s) encontrado(s)`}
              </Typography>
              {!isProjectView ? (
                <Pagination
                  color="primary"
                  count={totalPages}
                  onChange={handlePageChange}
                  page={budgetListQuery.data?.page ?? filters.page}
                  shape="rounded"
                />
              ) : null}
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
