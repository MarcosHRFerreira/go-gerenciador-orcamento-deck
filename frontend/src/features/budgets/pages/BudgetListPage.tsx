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
import { alpha } from "@mui/material/styles";
import type { SxProps, Theme } from "@mui/material/styles";
import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { isAxiosError } from "axios";
import { useEffect, useMemo, useRef, useState, type UIEvent } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import {
  compactFilterFieldSx,
  FilterField,
  filterGroupSx,
  filterSectionCardSx,
  filterGroupTitleSx,
} from "../../../components/common/FilterField";
import { PageHeader } from "../../../components/common/PageHeader";
import { SectionCard } from "../../../components/common/SectionCard";
import { useAuth } from "../../auth/hooks/useAuth";
import {
  deleteBudgetRequest,
  getBudgetCatalogsRequest,
  getBudgetManagementCatalogsRequest,
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

const defaultPageSize = 50;
const budgetGridBlue = "var(--app-accent-text)";
const budgetTintedDarkText = "#020617";

const defaultFilters: BudgetListFilters = {
  budgetNumber: "",
  sourceCompany: "",
  yearBudget: "",
  statusId: "",
  installerId: "",
  systemTypeId: "",
  projectCode: "",
  projectId: "",
  projectName: "",
  salespersonId: "",
  estimatorId: "",
  sentAtFrom: "",
  sentAtTo: "",
  page: 1,
  pageSize: defaultPageSize,
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
    systemTypeId:
      searchParams.get("systemTypeId") ?? defaultFilters.systemTypeId,
    projectCode: searchParams.get("projectCode") ?? defaultFilters.projectCode,
    projectId: searchParams.get("projectId") ?? defaultFilters.projectId,
    projectName: searchParams.get("projectName") ?? defaultFilters.projectName,
    salespersonId:
      searchParams.get("salespersonId") ?? defaultFilters.salespersonId,
    estimatorId: searchParams.get("estimatorId") ?? defaultFilters.estimatorId,
    sentAtFrom: searchParams.get("sentAtFrom") ?? defaultFilters.sentAtFrom,
    sentAtTo: searchParams.get("sentAtTo") ?? defaultFilters.sentAtTo,
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
  if (filters.systemTypeId) {
    nextSearchParams.set("systemTypeId", filters.systemTypeId);
  }
  if (filters.projectCode) {
    nextSearchParams.set("projectCode", filters.projectCode);
  }
  if (filters.projectId) {
    nextSearchParams.set("projectId", filters.projectId);
  }
  if (filters.projectName) {
    nextSearchParams.set("projectName", filters.projectName);
  }
  if (filters.salespersonId) {
    nextSearchParams.set("salespersonId", filters.salespersonId);
  }
  if (filters.estimatorId) {
    nextSearchParams.set("estimatorId", filters.estimatorId);
  }
  if (filters.sentAtFrom) {
    nextSearchParams.set("sentAtFrom", filters.sentAtFrom);
  }
  if (filters.sentAtTo) {
    nextSearchParams.set("sentAtTo", filters.sentAtTo);
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
      "Não foi possível carregar os orçamentos."
    );
  }

  return "Não foi possível carregar os orçamentos.";
}

function getBudgetDeleteErrorMessage(error: unknown) {
  if (isAxiosError<{ message?: string }>(error)) {
    return (
      error.response?.data?.message ?? "Não foi possível excluir o orçamento."
    );
  }

  return "Não foi possível excluir o orçamento.";
}

function formatOptionalCurrency(value: number | null) {
  if (value === null) {
    return "Não informado";
  }

  return currencyFormatter.format(value);
}

function formatOptionalText(value: string) {
  return value.trim() ? value : "Não informado";
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
  fallbackWhenMissing = "Não informado",
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
  fallbackWhenMissing = "Não informado",
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
  projectCode: string;
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
  backgroundColor: (theme: Theme) =>
    theme.palette.mode === "dark"
      ? alpha(theme.palette.primary.light, 0.16)
      : "#DBEAFE",
  borderBottomColor: (theme: Theme) =>
    theme.palette.mode === "dark"
      ? theme.palette.primary.light
      : budgetGridBlue,
  borderBottomStyle: "solid",
  borderBottomWidth: 2,
  color: (theme: Theme) =>
    theme.palette.mode === "dark"
      ? theme.palette.primary.light
      : budgetGridBlue,
  fontSize: "0.75rem",
  fontWeight: 800,
  letterSpacing: "0.05em",
  py: 1.2,
  textTransform: "uppercase",
  whiteSpace: "nowrap",
};

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
};

const singleLineTableCellSx = {
  ...tableDetailCellSx,
  maxWidth: 220,
  overflow: "hidden",
  textOverflow: "ellipsis",
  whiteSpace: "nowrap",
};

const premiumBudgetAlertSx = {
  borderRadius: 3,
  boxShadow: "0 14px 30px rgba(30, 58, 138, 0.08)",
  "& .MuiAlert-message": {
    fontWeight: 600,
  },
} as const;

const budgetActionPanelSx = {
  backgroundColor: (theme: Theme) =>
    theme.palette.mode === "dark"
      ? alpha(theme.palette.primary.light, 0.9)
      : alpha(theme.palette.common.white, 0.42),
  border: "1px solid",
  borderColor: (theme: Theme) =>
    theme.palette.mode === "dark"
      ? alpha(theme.palette.primary.dark, 0.22)
      : alpha(theme.palette.primary.main, 0.14),
  borderRadius: 4,
  boxShadow: (theme: Theme) =>
    theme.palette.mode === "dark"
      ? "0 18px 34px rgba(2, 6, 23, 0.22)"
      : `0 14px 28px ${alpha(theme.palette.primary.main, 0.06)}`,
  display: "flex",
  flexDirection: "column",
  gap: 2,
  p: 2.25,
} as const;

const budgetActionPanelTitleSx = {
  color: (theme: Theme) =>
    theme.palette.mode === "dark"
      ? budgetTintedDarkText
      : theme.palette.primary.dark,
  fontSize: "0.8rem",
  fontWeight: 800,
  letterSpacing: "0.04em",
  lineHeight: 1.2,
  textTransform: "uppercase",
} as const;

const budgetActionPanelBodySx = {
  color: (theme: Theme) =>
    theme.palette.mode === "dark"
      ? alpha(budgetTintedDarkText, 0.82)
      : theme.palette.text.secondary,
  fontWeight: 600,
} as const;

const budgetFiltersGridSx = {
  display: "grid",
  gap: 2,
  gridTemplateColumns: {
    xl: "minmax(0, 1.45fr) minmax(0, 1fr)",
    lg: "repeat(2, minmax(0, 1fr))",
    xs: "minmax(0, 1fr)",
  },
} as const;

const budgetIdentificationGridSx = {
  display: "grid",
  gap: 2,
  gridTemplateColumns: {
    xl: "repeat(4, minmax(0, 1fr))",
    lg: "repeat(3, minmax(0, 1fr))",
    sm: "repeat(2, minmax(0, 1fr))",
    xs: "minmax(0, 1fr)",
  },
} as const;

const budgetSecondaryFiltersGridSx = {
  display: "grid",
  gap: 2,
  gridTemplateColumns: {
    sm: "repeat(2, minmax(0, 1fr))",
    xs: "minmax(0, 1fr)",
  },
} as const;

const budgetWideFilterGroupSx: SxProps<Theme> = {
  ...(filterGroupSx as Record<string, unknown>),
  gridColumn: {
    xl: "span 2",
    xs: "auto",
  },
};

const budgetActionPanelContentSx = {
  alignItems: {
    xl: "start",
    xs: "stretch",
  },
  columnGap: 2.5,
  display: "grid",
  gridTemplateColumns: {
    xl: "minmax(0, 1.5fr) minmax(320px, 0.9fr)",
    xs: "minmax(0, 1fr)",
  },
  rowGap: 2.25,
} as const;

const budgetActionControlsSx = {
  alignItems: "end",
  display: "grid",
  gap: 1.5,
  gridTemplateColumns: {
    sm: "repeat(2, minmax(0, 1fr))",
    xs: "minmax(0, 1fr)",
  },
} as const;

const budgetPrimaryActionButtonSx = {
  boxShadow: (theme: Theme) =>
    theme.palette.mode === "dark"
      ? "0 16px 28px rgba(2, 6, 23, 0.2)"
      : `0 14px 28px ${alpha(theme.palette.primary.main, 0.22)}`,
  color: (theme: Theme) =>
    theme.palette.mode === "dark" ? budgetTintedDarkText : undefined,
  fontWeight: 800,
  minHeight: 44,
  minWidth: 148,
} as const;

const budgetSecondaryActionButtonSx = {
  backgroundColor: (theme: Theme) =>
    theme.palette.mode === "dark"
      ? alpha(theme.palette.common.white, 0.2)
      : "transparent",
  borderColor: (theme: Theme) =>
    theme.palette.mode === "dark"
      ? alpha(budgetTintedDarkText, 0.28)
      : alpha(theme.palette.primary.main, 0.28),
  color: (theme: Theme) =>
    theme.palette.mode === "dark" ? budgetTintedDarkText : budgetGridBlue,
  fontWeight: 800,
  minHeight: 44,
  minWidth: 128,
  "&:hover": {
    backgroundColor: (theme: Theme) =>
      theme.palette.mode === "dark"
        ? alpha(theme.palette.common.white, 0.3)
        : alpha(theme.palette.primary.main, 0.04),
    borderColor: (theme: Theme) =>
      theme.palette.mode === "dark"
        ? alpha(budgetTintedDarkText, 0.4)
        : alpha(theme.palette.primary.main, 0.36),
  },
} as const;

const budgetViewToggleSx = {
  backgroundColor: (theme: Theme) =>
    theme.palette.mode === "dark"
      ? alpha(theme.palette.primary.light, 0.88)
      : alpha(theme.palette.common.white, 0.46),
  borderRadius: 999,
  boxShadow: (theme: Theme) =>
    theme.palette.mode === "dark"
      ? "0 14px 28px rgba(2, 6, 23, 0.18)"
      : `0 12px 24px ${alpha(theme.palette.primary.main, 0.08)}`,
  p: 0.35,
  "& .MuiToggleButton-root": {
    borderColor: (theme: Theme) =>
      theme.palette.mode === "dark"
        ? alpha(budgetTintedDarkText, 0.2)
        : alpha(theme.palette.primary.main, 0.24),
    borderRadius: 999,
    color: (theme: Theme) =>
      theme.palette.mode === "dark" ? budgetTintedDarkText : budgetGridBlue,
    fontWeight: 800,
    px: 1.75,
  },
  "& .Mui-selected": {
    backgroundColor: (theme: Theme) =>
      theme.palette.mode === "dark"
        ? alpha(theme.palette.common.white, 0.32)
        : alpha(theme.palette.primary.main, 0.14),
    boxShadow: (theme: Theme) =>
      theme.palette.mode === "dark"
        ? "0 10px 18px rgba(2, 6, 23, 0.16)"
        : `0 10px 18px ${alpha(theme.palette.primary.main, 0.12)}`,
    color: (theme: Theme) =>
      theme.palette.mode === "dark" ? budgetTintedDarkText : budgetGridBlue,
  },
} as const;

const budgetPaginationSx = {
  "& .MuiPaginationItem-root": {
    color: "text.primary",
    fontWeight: 700,
  },
  "& .Mui-selected": {
    color: (theme: Theme) =>
      theme.palette.mode === "dark" ? budgetTintedDarkText : undefined,
  },
} as const;

const rowNumberColumnWidth = 68;
const budgetNumberColumnWidth = 140;
const floatingBudgetMirrorColumnWidth = 156;
const tableMaxHeight = "calc(100vh - 280px)";

export function BudgetListPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { user } = useAuth();
  const [searchParams, setSearchParams] = useSearchParams();
  const isAdmin = user?.role === "admin";
  const isEstimatorUser =
    user?.role === "user" && user.user_kind === "estimator";
  const canManageBudgetScreen = isAdmin || isEstimatorUser;
  const canCreateBudget = isAdmin || isEstimatorUser;
  const canDeleteBudget = isAdmin || isEstimatorUser;
  const filters = useMemo(
    () => getFiltersFromSearchParams(searchParams),
    [searchParams],
  );
  const effectiveFilters = useMemo(
    () =>
      canManageBudgetScreen
        ? filters
        : {
            ...filters,
            salespersonId: "",
            estimatorId: "",
          },
    [canManageBudgetScreen, filters],
  );
  const [budgetPendingDelete, setBudgetPendingDelete] = useState<{
    id: number;
    budgetNumber: string;
  } | null>(null);
  const [deleteError, setDeleteError] = useState<string | null>(null);
  const listTableContainerRef = useRef<HTMLDivElement | null>(null);
  const [showFloatingBudgetMirror, setShowFloatingBudgetMirror] =
    useState(false);
  const [draftFilters, setDraftFilters] = useState(() => ({
    budgetNumber: effectiveFilters.budgetNumber,
    sourceCompany: effectiveFilters.sourceCompany,
    yearBudget: effectiveFilters.yearBudget,
    statusId: effectiveFilters.statusId,
    installerId: effectiveFilters.installerId,
    systemTypeId: effectiveFilters.systemTypeId,
    projectCode: effectiveFilters.projectCode,
    projectName: effectiveFilters.projectName,
    salespersonId: effectiveFilters.salespersonId,
    estimatorId: effectiveFilters.estimatorId,
    sentAtFrom: effectiveFilters.sentAtFrom,
    sentAtTo: effectiveFilters.sentAtTo,
  }));
  const [viewMode, setViewMode] = useState<BudgetViewMode>("list");

  useEffect(() => {
    if (
      canManageBudgetScreen ||
      (!filters.salespersonId && !filters.estimatorId)
    ) {
      return;
    }

    setSearchParams(buildSearchParams(effectiveFilters), { replace: true });
  }, [
    effectiveFilters,
    filters.estimatorId,
    filters.salespersonId,
    canManageBudgetScreen,
    setSearchParams,
  ]);

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
  const hasProjectCodeFilter = useMemo(
    () => normalizeText(projectBudgetFilters.projectCode) !== "",
    [projectBudgetFilters.projectCode],
  );
  const budgetCatalogsQuery = useQuery({
    queryKey: ["budget-catalogs", user?.id ?? "anonymous"],
    queryFn: canManageBudgetScreen
      ? getBudgetManagementCatalogsRequest
      : getBudgetListCatalogsRequest,
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
  const productLineMap = useMemo(
    () => createCatalogMap(budgetCatalogsQuery.data?.productLines ?? []),
    [budgetCatalogsQuery.data?.productLines],
  );
  const systemTypeMap = useMemo(
    () => createCatalogMap(budgetCatalogsQuery.data?.systemTypes ?? []),
    [budgetCatalogsQuery.data?.systemTypes],
  );
  const projectMap = useMemo(
    () => createCatalogMap(budgetCatalogsQuery.data?.projects ?? []),
    [budgetCatalogsQuery.data?.projects],
  );
  const salespersonMap = useMemo(
    () => createCatalogMap(budgetCatalogsQuery.data?.salespeople ?? []),
    [budgetCatalogsQuery.data?.salespeople],
  );
  const estimatorMap = useMemo(
    () => createCatalogMap(budgetCatalogsQuery.data?.estimators ?? []),
    [budgetCatalogsQuery.data?.estimators],
  );
  const contactMap = useMemo(
    () => createCatalogMap(budgetCatalogsQuery.data?.contacts ?? []),
    [budgetCatalogsQuery.data?.contacts],
  );
  const lossReasonMap = useMemo(
    () => createCatalogMap(budgetCatalogsQuery.data?.lossReasons ?? []),
    [budgetCatalogsQuery.data?.lossReasons],
  );
  const selectedProjectFilterLabel = useMemo(() => {
    if (!effectiveFilters.projectId) {
      return effectiveFilters.projectName.trim();
    }

    return formatCatalogName(
      Number(effectiveFilters.projectId),
      projectMap,
      effectiveFilters.projectName.trim() ||
        `Obra #${effectiveFilters.projectId}`,
    );
  }, [effectiveFilters.projectId, effectiveFilters.projectName, projectMap]);
  const appliedFilterChips = useMemo(() => {
    const chips: string[] = [];

    if (effectiveFilters.budgetNumber.trim()) {
      chips.push(`Orçamento ${effectiveFilters.budgetNumber.trim()}`);
    }
    if (effectiveFilters.yearBudget.trim()) {
      chips.push(`Ano ${effectiveFilters.yearBudget.trim()}`);
    }
    if (effectiveFilters.sourceCompany.trim()) {
      chips.push(`Empresa ${effectiveFilters.sourceCompany.trim()}`);
    }
    if (effectiveFilters.projectCode.trim()) {
      chips.push(`Obra ${effectiveFilters.projectCode.trim()}`);
    }
    if (effectiveFilters.projectId) {
      chips.push(`Obra vinculada ${selectedProjectFilterLabel}`);
    }
    if (effectiveFilters.projectName.trim()) {
      chips.push(`Nome da obra ${effectiveFilters.projectName.trim()}`);
    }
    if (effectiveFilters.statusId) {
      chips.push(
        `Status ${formatCatalogName(Number(effectiveFilters.statusId), statusMap, "Status")}`,
      );
    }
    if (effectiveFilters.installerId) {
      chips.push(
        `Instalador ${formatCatalogName(Number(effectiveFilters.installerId), installerMap, "Instalador")}`,
      );
    }
    if (effectiveFilters.systemTypeId) {
      chips.push(
        `Tipo ${formatCatalogName(Number(effectiveFilters.systemTypeId), systemTypeMap, "Tipo de Sistema")}`,
      );
    }
    if (effectiveFilters.salespersonId) {
      chips.push(
        `Vendedor ${formatCatalogName(Number(effectiveFilters.salespersonId), salespersonMap, "Vendedor")}`,
      );
    }
    if (effectiveFilters.estimatorId) {
      chips.push(
        `Orçamentista ${formatCatalogName(Number(effectiveFilters.estimatorId), estimatorMap, "Orçamentista")}`,
      );
    }
    if (effectiveFilters.sentAtFrom || effectiveFilters.sentAtTo) {
      chips.push(
        `Período ${effectiveFilters.sentAtFrom || "início"} até ${effectiveFilters.sentAtTo || "hoje"}`,
      );
    }

    return chips;
  }, [
    effectiveFilters.budgetNumber,
    effectiveFilters.yearBudget,
    effectiveFilters.sourceCompany,
    effectiveFilters.projectCode,
    effectiveFilters.projectId,
    effectiveFilters.projectName,
    effectiveFilters.statusId,
    effectiveFilters.installerId,
    effectiveFilters.systemTypeId,
    effectiveFilters.salespersonId,
    effectiveFilters.estimatorId,
    effectiveFilters.sentAtFrom,
    effectiveFilters.sentAtTo,
    estimatorMap,
    installerMap,
    selectedProjectFilterLabel,
    salespersonMap,
    statusMap,
    systemTypeMap,
  ]);

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
        "Obra não informada",
      );
      const projectCode = normalizeText(budget.projectCode)
        ? (budget.projectCode?.trim() ?? "")
        : "Código não informado";
      const groupKey = `project-${budget.projectId}`;
      const existingGroup = currentGroups.get(groupKey);

      if (existingGroup) {
        existingGroup.items.push(budget);
        return currentGroups;
      }

      currentGroups.set(groupKey, {
        key: groupKey,
        projectId: budget.projectId,
        projectCode,
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
      .filter((group) => hasProjectCodeFilter || group.needsAttention)
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
  }, [
    hasProjectCodeFilter,
    projectBudgetItems,
    projectMap,
    systemTypeMap,
    statusMap,
  ]);
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
    setShowFloatingBudgetMirror(false);
    if (!listTableContainerRef.current) {
      return;
    }

    listTableContainerRef.current.scrollLeft = 0;
  }, [filters.page, filters.pageSize, isProjectView]);

  const handleListTableScroll = (event: UIEvent<HTMLDivElement>) => {
    setShowFloatingBudgetMirror(event.currentTarget.scrollLeft > 12);
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
      salespersonId: canManageBudgetScreen ? draftFilters.salespersonId : "",
      estimatorId: canManageBudgetScreen ? draftFilters.estimatorId : "",
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
      systemTypeId: defaultFilters.systemTypeId,
      projectCode: defaultFilters.projectCode,
      projectName: defaultFilters.projectName,
      salespersonId: canManageBudgetScreen ? defaultFilters.salespersonId : "",
      estimatorId: canManageBudgetScreen ? defaultFilters.estimatorId : "",
      sentAtFrom: defaultFilters.sentAtFrom,
      sentAtTo: defaultFilters.sentAtTo,
    });
    setSearchParams(
      buildSearchParams({
        ...defaultFilters,
        salespersonId: canManageBudgetScreen
          ? defaultFilters.salespersonId
          : "",
        estimatorId: canManageBudgetScreen ? defaultFilters.estimatorId : "",
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

  const handlePageSizeChange = (value: number) => {
    setSearchParams(
      buildSearchParams({
        ...effectiveFilters,
        page: 1,
        pageSize: value,
      }),
    );
  };

  const handleBudgetRowDoubleClick = (budgetId: number) => {
    if (!user) {
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
      <PageHeader
        action={
          canCreateBudget ? (
            <Box
              sx={{
                display: "flex",
                flexDirection: { sm: "row", xs: "column" },
                gap: 1.25,
                width: "100%",
              }}
            >
              {isAdmin ? (
                <Button
                  onClick={() => navigate("/budgets/import")}
                  startIcon={<UploadFileRoundedIcon />}
                  sx={budgetSecondaryActionButtonSx}
                  variant="outlined"
                >
                  Importar planilha
                </Button>
              ) : null}
              <Button
                onClick={() => navigate("/budgets/new")}
                startIcon={<AddRoundedIcon />}
                variant="contained"
              >
                Novo orçamento
              </Button>
            </Box>
          ) : undefined
        }
        description="Gerencie a operação comercial com filtros avançados, leitura por obra e visualização detalhada dos orçamentos mais relevantes."
        title="Orçamentos"
      />

      <SectionCard
        description="Filtre por identificação, classificação, responsáveis e período para localizar rapidamente os orçamentos certos."
        sx={filterSectionCardSx}
        title="Filtros"
      >
        <Box
          sx={{
            display: "grid",
            gap: 2.5,
          }}
        >
          <Box sx={budgetFiltersGridSx}>
            <Box sx={budgetWideFilterGroupSx}>
              <Typography sx={filterGroupTitleSx} variant="subtitle2">
                Identificação
              </Typography>
              <Typography sx={budgetActionPanelBodySx} variant="body2">
                Use o campo `Obra` para buscar por qualquer parte do nome ou do
                código e localizar rapidamente os orçamentos relacionados.
              </Typography>
              <Box sx={budgetIdentificationGridSx}>
                <Box
                  sx={{
                    gridColumn: {
                      xl: "span 2",
                      xs: "auto",
                    },
                  }}
                >
                  <FilterField label="Obra">
                    <TextField
                      onChange={(event) =>
                        handleDraftChange("projectName", event.target.value)
                      }
                      placeholder="Digite parte do nome ou codigo da obra"
                      size="small"
                      sx={compactFilterFieldSx}
                      value={draftFilters.projectName}
                    />
                  </FilterField>
                </Box>
                <FilterField label="Número do orçamento">
                  <TextField
                    onChange={(event) =>
                      handleDraftChange("budgetNumber", event.target.value)
                    }
                    placeholder="Ex: BGT-2026-001"
                    size="small"
                    sx={compactFilterFieldSx}
                    value={draftFilters.budgetNumber}
                  />
                </FilterField>
                <FilterField label="Ano">
                  <TextField
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
                </FilterField>
                <FilterField label="Empresa">
                  <TextField
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
                </FilterField>
                <FilterField label="Código da Obra">
                  <TextField
                    onChange={(event) =>
                      handleDraftChange("projectCode", event.target.value)
                    }
                    placeholder="Digite parte do código da obra"
                    size="small"
                    sx={compactFilterFieldSx}
                    value={draftFilters.projectCode}
                  />
                </FilterField>
              </Box>
            </Box>

            <Box sx={filterGroupSx}>
              <Typography sx={filterGroupTitleSx} variant="subtitle2">
                Classificação
              </Typography>
              <Box sx={budgetSecondaryFiltersGridSx}>
                <FilterField label="Status">
                  <TextField
                    onChange={(event) =>
                      handleDraftChange("statusId", event.target.value)
                    }
                    select
                    size="small"
                    sx={compactFilterFieldSx}
                    value={draftFilters.statusId}
                  >
                    <MenuItem value="">Todos</MenuItem>
                    {(budgetCatalogsQuery.data?.statuses ?? []).map(
                      (status) => (
                        <MenuItem key={status.id} value={String(status.id)}>
                          {status.name}
                        </MenuItem>
                      ),
                    )}
                  </TextField>
                </FilterField>
                <FilterField label="Instalador">
                  <TextField
                    onChange={(event) =>
                      handleDraftChange("installerId", event.target.value)
                    }
                    select
                    size="small"
                    sx={compactFilterFieldSx}
                    value={draftFilters.installerId}
                  >
                    <MenuItem value="">Todos</MenuItem>
                    {(budgetCatalogsQuery.data?.installers ?? []).map(
                      (installer) => (
                        <MenuItem
                          key={installer.id}
                          value={String(installer.id)}
                        >
                          {installer.name}
                        </MenuItem>
                      ),
                    )}
                  </TextField>
                </FilterField>
                <FilterField label="Tipo de Sistema">
                  <TextField
                    onChange={(event) =>
                      handleDraftChange("systemTypeId", event.target.value)
                    }
                    select
                    size="small"
                    sx={compactFilterFieldSx}
                    value={draftFilters.systemTypeId}
                  >
                    <MenuItem value="">Todos</MenuItem>
                    {(budgetCatalogsQuery.data?.systemTypes ?? []).map(
                      (systemType) => (
                        <MenuItem
                          key={systemType.id}
                          value={String(systemType.id)}
                        >
                          {systemType.name}
                        </MenuItem>
                      ),
                    )}
                  </TextField>
                </FilterField>
              </Box>
            </Box>

            <Box sx={filterGroupSx}>
              <Typography sx={filterGroupTitleSx} variant="subtitle2">
                Responsáveis e período
              </Typography>
              <Box sx={budgetSecondaryFiltersGridSx}>
                {canManageBudgetScreen ? (
                  <FilterField label="Vendedor">
                    <TextField
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
                          <MenuItem
                            key={salesperson.id}
                            value={String(salesperson.id)}
                          >
                            {salesperson.name}
                          </MenuItem>
                        ),
                      )}
                    </TextField>
                  </FilterField>
                ) : null}
                {canManageBudgetScreen ? (
                  <FilterField label="Orçamentista">
                    <TextField
                      onChange={(event) =>
                        handleDraftChange("estimatorId", event.target.value)
                      }
                      select
                      size="small"
                      sx={compactFilterFieldSx}
                      value={draftFilters.estimatorId}
                    >
                      <MenuItem value="">Todos</MenuItem>
                      {(budgetCatalogsQuery.data?.estimators ?? []).map(
                        (estimator) => (
                          <MenuItem
                            key={estimator.id}
                            value={String(estimator.id)}
                          >
                            {estimator.name}
                          </MenuItem>
                        ),
                      )}
                    </TextField>
                  </FilterField>
                ) : null}
                <FilterField label="Envio de">
                  <TextField
                    onChange={(event) =>
                      handleDraftChange("sentAtFrom", event.target.value)
                    }
                    size="small"
                    sx={compactFilterFieldSx}
                    type="date"
                    value={draftFilters.sentAtFrom}
                  />
                </FilterField>
                <FilterField label="Envio até">
                  <TextField
                    onChange={(event) =>
                      handleDraftChange("sentAtTo", event.target.value)
                    }
                    size="small"
                    sx={compactFilterFieldSx}
                    type="date"
                    value={draftFilters.sentAtTo}
                  />
                </FilterField>
              </Box>
            </Box>
          </Box>

          <Box sx={budgetActionPanelSx}>
            <Box sx={budgetActionPanelContentSx}>
              <Box sx={{ display: "flex", flexDirection: "column", gap: 1 }}>
                <Typography sx={budgetActionPanelTitleSx} variant="subtitle2">
                  Filtros ativos
                </Typography>
                {effectiveFilters.projectId ? (
                  <Alert
                    severity="info"
                    sx={{
                      borderRadius: 3,
                      fontWeight: 700,
                      "& .MuiAlert-message": {
                        display: "flex",
                        flexDirection: "column",
                        gap: 0.3,
                      },
                    }}
                  >
                    <Typography sx={{ fontWeight: 800 }} variant="body2">
                      Exibindo somente os orçamentos da obra selecionada.
                    </Typography>
                    <Typography variant="body2">
                      {selectedProjectFilterLabel}
                    </Typography>
                  </Alert>
                ) : null}
                {appliedFilterChips.length > 0 ? (
                  <Box sx={{ display: "flex", flexWrap: "wrap", gap: 1 }}>
                    {appliedFilterChips.map((item) => (
                      <Chip
                        key={item}
                        label={item}
                        size="small"
                        sx={{
                          "& .MuiChip-label": { fontWeight: 700 },
                          backgroundColor: (theme) =>
                            theme.palette.mode === "dark"
                              ? alpha(theme.palette.common.white, 0.28)
                              : alpha(theme.palette.primary.main, 0.09),
                          borderColor: (theme) =>
                            theme.palette.mode === "dark"
                              ? alpha(budgetTintedDarkText, 0.2)
                              : alpha(theme.palette.primary.main, 0.18),
                          color: (theme) =>
                            theme.palette.mode === "dark"
                              ? budgetTintedDarkText
                              : budgetGridBlue,
                        }}
                        variant="outlined"
                      />
                    ))}
                  </Box>
                ) : (
                  <Typography sx={budgetActionPanelBodySx} variant="body2">
                    Nenhum filtro aplicado no momento. Use os campos acima para
                    refinar a consulta.
                  </Typography>
                )}
              </Box>

              <Box sx={budgetActionControlsSx}>
                <FilterField label="Ordenar por">
                  <TextField
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
                    <MenuItem value="budget_number">Número</MenuItem>
                  </TextField>
                </FilterField>
                <FilterField label="Ordem">
                  <TextField
                    onChange={(event) =>
                      handleSortOrderChange(
                        event.target.value as BudgetSortOrder,
                      )
                    }
                    select
                    size="small"
                    sx={compactFilterFieldSx}
                    value={filters.sortOrder}
                  >
                    <MenuItem value="desc">Decrescente</MenuItem>
                    <MenuItem value="asc">Crescente</MenuItem>
                  </TextField>
                </FilterField>
                <Button
                  onClick={handleApplyFilters}
                  startIcon={<SearchRoundedIcon />}
                  sx={budgetPrimaryActionButtonSx}
                  variant="contained"
                >
                  Filtrar
                </Button>
                <Button
                  onClick={handleClearFilters}
                  sx={budgetSecondaryActionButtonSx}
                  variant="outlined"
                >
                  Limpar
                </Button>
              </Box>
            </Box>
          </Box>
        </Box>
      </SectionCard>

      <SectionCard
        sx={{
          background: (theme) =>
            `linear-gradient(135deg, ${alpha(theme.palette.info.main, 0.05)} 0%, ${alpha(theme.palette.primary.main, 0.03)} 100%)`,
          border: "1px solid",
          borderColor: (theme) => alpha(theme.palette.info.main, 0.16),
          boxShadow: (theme) =>
            `0 14px 28px ${alpha(theme.palette.info.main, 0.08)}`,
        }}
      >
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
          <Alert severity="error" sx={premiumBudgetAlertSx}>
            {getBudgetErrorMessage(activeBudgetQuery.error)}
          </Alert>
        ) : null}

        {budgetCatalogsQuery.isError ? (
          <Alert
            severity="warning"
            sx={premiumBudgetAlertSx}
            variant="outlined"
          >
            Não foi possível carregar alguns catálogos. Alguns campos podem
            aparecer com ID.
          </Alert>
        ) : null}

        {deleteError ? (
          <Alert severity="error" sx={premiumBudgetAlertSx}>
            {deleteError}
          </Alert>
        ) : null}

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
              <Box sx={{ display: "flex", flexDirection: "column", gap: 0.9 }}>
                <Box sx={{ display: "flex", flexWrap: "wrap", gap: 1 }}>
                  <Chip
                    color="primary"
                    label={isProjectView ? "Visão por obra" : "Lista completa"}
                    size="small"
                    variant="outlined"
                  />
                  {!isProjectView ? (
                    <Chip
                      label={`Página ${budgetListQuery.data?.page ?? filters.page} de ${totalPages}`}
                      size="small"
                      variant="outlined"
                    />
                  ) : null}
                  <Chip
                    label={
                      isProjectView
                        ? `${projectBudgetListQuery.data?.total ?? 0} orçamento(s) no recorte`
                        : `${budgetListQuery.data?.total ?? 0} orçamento(s) no recorte`
                    }
                    size="small"
                    variant="outlined"
                  />
                </Box>
                <Typography
                  color="text.primary"
                  sx={{ fontWeight: 600, maxWidth: 780 }}
                  variant="body2"
                >
                  {isProjectView
                    ? hasProjectCodeFilter
                      ? `${groupedProjectsCount} obra(s) encontrada(s) para o código filtrado.`
                      : `${groupedProjectsCount} obra(s) sem PEDIDO definido entre as 20 mais recentes, com prioridade para quem tem mais orçamentos.`
                    : "A listagem abaixo prioriza leitura operacional, comparação rápida e acesso direto às ações principais."}
                </Typography>
              </Box>

              <Box
                sx={{
                  alignItems: { sm: "center", xs: "stretch" },
                  display: "flex",
                  flexDirection: { sm: "row", xs: "column" },
                  gap: 1.5,
                }}
              >
                {!isProjectView ? (
                  <FilterField label="Linhas por página">
                    <TextField
                      onChange={(event) =>
                        handlePageSizeChange(Number(event.target.value))
                      }
                      select
                      size="small"
                      sx={{ ...compactFilterFieldSx, minWidth: 160 }}
                      value={filters.pageSize}
                    >
                      <MenuItem value={50}>50</MenuItem>
                      <MenuItem value={100}>100</MenuItem>
                    </TextField>
                  </FilterField>
                ) : null}
                <ToggleButtonGroup
                  exclusive
                  onChange={handleViewModeChange}
                  size="small"
                  sx={budgetViewToggleSx}
                  value={viewMode}
                >
                  <ToggleButton value="project">
                    <ApartmentRoundedIcon fontSize="small" sx={{ mr: 1 }} />
                    Por obra
                  </ToggleButton>
                  <ToggleButton value="list">
                    <TableRowsRoundedIcon fontSize="small" sx={{ mr: 1 }} />
                    Lista completa
                  </ToggleButton>
                </ToggleButtonGroup>
              </Box>
            </Box>

            {isProjectView ? (
              <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
                <Alert
                  severity="info"
                  sx={premiumBudgetAlertSx}
                  variant="outlined"
                >
                  {hasProjectCodeFilter ? (
                    <>
                      Como há filtro por <strong>Código da Obra</strong>, a
                      visualização exibe também obras que já possuem{" "}
                      <strong>PEDIDO</strong> definido.
                    </>
                  ) : (
                    <>
                      Nesta visualização aparecem apenas obras sem{" "}
                      <strong>PEDIDO</strong> definido, limitados aos 20 mais
                      recentes pela última atualização dos orçamentos do grupo e
                      ordenados no topo por quantidade de orçamentos vinculados.
                    </>
                  )}
                </Alert>
                {projectBudgetItems.length === 0 ? (
                  <Alert
                    severity="info"
                    sx={premiumBudgetAlertSx}
                    variant="outlined"
                  >
                    Nenhum orçamento foi encontrado para os filtros informados.
                  </Alert>
                ) : projectGroups.length === 0 ? (
                  <Alert
                    severity="info"
                    sx={premiumBudgetAlertSx}
                    variant="outlined"
                  >
                    {hasProjectCodeFilter ? (
                      <>
                        Foram encontrados{" "}
                        <strong>
                          {projectBudgetItems.length} orçamento(s)
                        </strong>{" "}
                        na busca atual, mas nenhuma obra agrupável corresponde
                        ao código informado.
                      </>
                    ) : (
                      <>
                        Foram encontrados{" "}
                        <strong>
                          {projectBudgetItems.length} orçamento(s)
                        </strong>{" "}
                        na busca atual, mas todas as obras já possuem{" "}
                        <strong>PEDIDO</strong> definido e por isso não aparecem
                        nesta visualização.
                      </>
                    )}
                  </Alert>
                ) : (
                  projectGroups.map((group) => (
                    <Accordion
                      defaultExpanded={group.items.length > 1}
                      disableGutters
                      key={group.key}
                      sx={{
                        background: (theme) =>
                          `linear-gradient(135deg, ${alpha(theme.palette.primary.main, 0.045)} 0%, ${alpha(theme.palette.info.main, 0.02)} 100%)`,
                        border: (theme) =>
                          `1px solid ${alpha(theme.palette.primary.main, 0.18)}`,
                        boxShadow: (theme) =>
                          `0 12px 26px ${alpha(theme.palette.primary.main, 0.06)}`,
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
                            <Typography
                              color={budgetGridBlue}
                              sx={{ fontWeight: 800 }}
                              variant="h6"
                            >
                              {group.projectName}
                            </Typography>
                            <Typography
                              color="text.primary"
                              sx={{ fontWeight: 500 }}
                              variant="body2"
                            >
                              {`${group.projectCode} - ${group.items.length} orçamento(s) vinculado(s) - ${group.items[0]?.sourceCompany || "Não informado"}`}
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
                              label={`${group.items.length} orçamento(s)`}
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
                                Abrir obra
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
                                      return alpha(
                                        theme.palette.success.main,
                                        0.14,
                                      );
                                    }

                                    if (statusCategory === "cancelado") {
                                      return alpha(
                                        theme.palette.grey[500],
                                        0.12,
                                      );
                                    }

                                    return theme.palette.mode === "dark"
                                      ? alpha(theme.palette.info.light, 0.05)
                                      : alpha(theme.palette.info.main, 0.035);
                                  },
                                  border: (theme) => {
                                    if (isWinner) {
                                      return `1px solid ${theme.palette.success.main}`;
                                    }

                                    return `1px solid ${alpha(theme.palette.primary.main, 0.14)}`;
                                  },
                                  boxShadow: (theme) =>
                                    `0 10px 24px ${alpha(theme.palette.primary.main, 0.05)}`,
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
                                      color={budgetGridBlue}
                                      sx={{ fontWeight: 800 }}
                                      variant="subtitle1"
                                    >
                                      {budget.budgetNumber}
                                    </Typography>
                                    <Typography
                                      color="text.primary"
                                      sx={{ fontWeight: 500 }}
                                      variant="body2"
                                    >
                                      ID {budget.id} - envio{" "}
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
                                        label="Vencedor da obra"
                                        size="small"
                                      />
                                    ) : null}
                                    {statusCategory === "cancelado" ? (
                                      <Chip
                                        label="Sem necessidade de atenção"
                                        size="small"
                                        variant="outlined"
                                      />
                                    ) : null}
                                  </Box>
                                </Box>

                                <Box
                                  sx={{
                                    "& .MuiTypography-body2": {
                                      color: "text.primary",
                                      fontWeight: 600,
                                    },
                                    "& .MuiTypography-caption": {
                                      color: budgetGridBlue,
                                      fontSize: "0.76rem",
                                      fontWeight: 700,
                                    },
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
                                      Orçamentista
                                    </Typography>
                                    <Typography variant="body2">
                                      {formatResolvedCatalogName(
                                        budget.estimatorId,
                                        budget.estimatorName,
                                        estimatorMap,
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
                                      Tipo de Sistema
                                    </Typography>
                                    <Typography variant="body2">
                                      {formatResolvedCatalogName(
                                        budget.systemTypeId,
                                        budget.systemTypeName,
                                        systemTypeMap,
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
                                      Área m2
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
                                      Projetista
                                    </Typography>
                                    <Typography variant="body2">
                                      {formatOptionalText(
                                        budget.projetistaName,
                                      )}
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

                                {user ? (
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
                                    {canDeleteBudget ? (
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
                                    ) : null}
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
                  onScroll={handleListTableScroll}
                  ref={listTableContainerRef}
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
                      minWidth: 1980,
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
                        <TableCell sx={tableHeadCellSx}>Ano</TableCell>
                        <TableCell sx={tableHeadCellSx}>Empresa</TableCell>
                        <TableCell sx={tableHeadCellSx}>Revisão</TableCell>
                        <TableCell sx={tableHeadCellSx}>Envio</TableCell>
                        <TableCell sx={tableHeadCellSx}>Status</TableCell>
                        <TableCell sx={tableHeadCellSx}>Prioridade</TableCell>
                        <TableCell sx={tableHeadCellSx}>Instalador</TableCell>
                        <TableCell sx={tableHeadCellSx}>
                          Linha de produtos
                        </TableCell>
                        <TableCell sx={tableHeadCellSx}>
                          Tipo de Sistema
                        </TableCell>
                        <TableCell sx={tableHeadCellSx}>Obra</TableCell>
                        <TableCell sx={tableHeadCellSx}>Construtora</TableCell>
                        <TableCell sx={tableHeadCellSx}>Vendedor</TableCell>
                        <TableCell sx={tableHeadCellSx}>Orçamentista</TableCell>
                        <TableCell sx={tableHeadCellSx}>Contato</TableCell>
                        <TableCell sx={tableHeadCellSx}>
                          Motivo de perda
                        </TableCell>
                        <TableCell sx={tableHeadCellSx}>Projetista</TableCell>
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
                          Área m2
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
                        {user ? (
                          <TableCell sx={tableHeadCellSx}>Ações</TableCell>
                        ) : null}
                        <TableCell
                          sx={{
                            ...tableHeadCellSx,
                            borderLeft: (theme) =>
                              showFloatingBudgetMirror
                                ? `1px solid ${theme.palette.divider}`
                                : "none",
                            boxShadow: showFloatingBudgetMirror
                              ? "-10px 0 24px rgba(15, 23, 42, 0.12)"
                              : "none",
                            minWidth: floatingBudgetMirrorColumnWidth,
                            opacity: showFloatingBudgetMirror ? 1 : 0,
                            pointerEvents: showFloatingBudgetMirror
                              ? "auto"
                              : "none",
                            position: "sticky",
                            right: 0,
                            transform: showFloatingBudgetMirror
                              ? "translateX(0)"
                              : "translateX(12px)",
                            transition:
                              "opacity 0.2s ease, transform 0.2s ease, box-shadow 0.2s ease",
                            width: floatingBudgetMirrorColumnWidth,
                            zIndex: 4,
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
                          sx={{
                            cursor: user ? "pointer" : "default",
                          }}
                        >
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
                            {(filters.page - 1) * filters.pageSize + index + 1}
                          </TableCell>
                          <TableCell
                            sx={{
                              ...singleLineTableCellSx,
                              fontWeight: 600,
                              maxWidth: budgetNumberColumnWidth,
                              minWidth: budgetNumberColumnWidth,
                              width: budgetNumberColumnWidth,
                            }}
                            title={budget.budgetNumber}
                          >
                            {budget.budgetNumber}
                          </TableCell>
                          <TableCell sx={singleLineTableCellSx}>
                            {budget.yearBudget}
                          </TableCell>
                          <TableCell
                            sx={singleLineTableCellSx}
                            title={budget.sourceCompany || "Não informado"}
                          >
                            {budget.sourceCompany || "Não informado"}
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
                              sx={{
                                fontSize: "0.7rem",
                                fontWeight: 700,
                                height: 22,
                              }}
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
                              budget.productLineId,
                              budget.productLineName,
                              productLineMap,
                            )}
                          >
                            {formatResolvedCatalogName(
                              budget.productLineId,
                              budget.productLineName,
                              productLineMap,
                            )}
                          </TableCell>
                          <TableCell
                            sx={singleLineTableCellSx}
                            title={formatResolvedCatalogName(
                              budget.systemTypeId,
                              budget.systemTypeName,
                              systemTypeMap,
                            )}
                          >
                            {formatResolvedCatalogName(
                              budget.systemTypeId,
                              budget.systemTypeName,
                              systemTypeMap,
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
                            title={formatOptionalText(
                              budget.constructionCompany,
                            )}
                          >
                            {formatOptionalText(budget.constructionCompany)}
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
                              budget.estimatorId,
                              budget.estimatorName,
                              estimatorMap,
                            )}
                          >
                            {formatResolvedCatalogName(
                              budget.estimatorId,
                              budget.estimatorName,
                              estimatorMap,
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
                            title={formatOptionalText(budget.projetistaName)}
                          >
                            {formatOptionalText(budget.projetistaName)}
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
                          {user ? (
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
                              {canDeleteBudget ? (
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
                              ) : null}
                            </TableCell>
                          ) : null}
                          <TableCell
                            sx={{
                              ...tableDetailCellSx,
                              backgroundColor: "background.paper",
                              borderLeft: (theme) =>
                                showFloatingBudgetMirror
                                  ? `1px solid ${theme.palette.divider}`
                                  : "none",
                              boxShadow: showFloatingBudgetMirror
                                ? "-10px 0 24px rgba(15, 23, 42, 0.08)"
                                : "none",
                              opacity: showFloatingBudgetMirror ? 1 : 0,
                              pointerEvents: showFloatingBudgetMirror
                                ? "auto"
                                : "none",
                              position: "sticky",
                              right: 0,
                              textAlign: "center",
                              transform: showFloatingBudgetMirror
                                ? "translateX(0)"
                                : "translateX(12px)",
                              transition:
                                "opacity 0.2s ease, transform 0.2s ease, box-shadow 0.2s ease",
                              width: floatingBudgetMirrorColumnWidth,
                              zIndex: 1,
                            }}
                          >
                            <Chip
                              label={budget.budgetNumber}
                              size="small"
                              sx={{
                                fontSize: "0.68rem",
                                fontWeight: 700,
                                maxWidth: "100%",
                                "& .MuiChip-label": {
                                  overflow: "hidden",
                                  px: 1.25,
                                  textOverflow: "ellipsis",
                                },
                              }}
                              title={budget.budgetNumber}
                              variant="outlined"
                            />
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </TableContainer>
              </Box>
            ) : (
              <Alert severity="info" variant="outlined">
                Nenhum orçamento encontrado para os filtros informados.
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
              <Typography
                color="text.primary"
                sx={{ fontWeight: 600 }}
                variant="body2"
              >
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
                  sx={budgetPaginationSx}
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
            Confirma a exclusão do orçamento{" "}
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
