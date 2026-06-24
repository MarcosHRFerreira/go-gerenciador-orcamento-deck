import ArrowBackRoundedIcon from "@mui/icons-material/ArrowBackRounded";
import ContentCopyRoundedIcon from "@mui/icons-material/ContentCopyRounded";
import EditRoundedIcon from "@mui/icons-material/EditRounded";
import EmojiEventsRoundedIcon from "@mui/icons-material/EmojiEventsRounded";
import FolderOpenRoundedIcon from "@mui/icons-material/FolderOpenRounded";
import ForumRoundedIcon from "@mui/icons-material/ForumRounded";
import LinkRoundedIcon from "@mui/icons-material/LinkRounded";
import LinkOffRoundedIcon from "@mui/icons-material/LinkOffRounded";
import NoteAddRoundedIcon from "@mui/icons-material/NoteAddRounded";
import TimelineRoundedIcon from "@mui/icons-material/TimelineRounded";
import {
  Alert,
  Box,
  Button,
  Chip,
  Divider,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  IconButton,
  List,
  ListItem,
  ListItemText,
  LinearProgress,
  MenuItem,
  TextField,
  Tooltip,
  Typography,
} from "@mui/material";
import {
  useMutation,
  useQueries,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { isAxiosError } from "axios";
import { useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { alpha } from "@mui/material/styles";
import {
  compactFilterFieldSx,
  FilterField,
  filterGroupSx,
  filterGroupTitleSx,
} from "../../../components/common/FilterField";
import { PageHeader } from "../../../components/common/PageHeader";
import { SectionCard } from "../../../components/common/SectionCard";
import { useAuth } from "../../auth/hooks/useAuth";
import {
  electBudgetWinnerRequest,
  getBudgetByIdRequest,
  getBudgetListCatalogsRequest,
  getBudgetListRequest,
  getBudgetStatusHistoryRequest,
  updateBudgetRequest,
} from "../../budgets/api/budgets";
import type {
  BudgetCatalogItem,
  BudgetCreatePayload,
  BudgetDetailItem,
  BudgetListFilters,
  BudgetListItem,
  BudgetStatusHistoryItem,
} from "../../budgets/types/budget";
import {
  getBudgetStatusDisplayName,
  getWonStatusSingularLabel,
  isWonStatusLabel,
} from "../../budgets/utils/businessTerms";
import {
  getProjectByIdRequest,
  listProjectTypesRequest,
} from "../api/projects";
import type { ProjectTypeCatalogItem } from "../types/project";

const currencyFormatter = new Intl.NumberFormat("pt-BR", {
  currency: "BRL",
  style: "currency",
});

const dateFormatter = new Intl.DateTimeFormat("pt-BR", {
  dateStyle: "short",
});

const dateTimeFormatter = new Intl.DateTimeFormat("pt-BR", {
  dateStyle: "short",
  timeStyle: "short",
});

type BudgetStatusCategory = "pedido" | "cancelado" | "orcamento" | "other";

function formatOptionalText(value: string | null | undefined) {
  const trimmedValue = value?.trim();

  return trimmedValue ? trimmedValue : "Não informado";
}

function normalizeValue(value: string | null | undefined) {
  return (value ?? "")
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "")
    .trim()
    .toLowerCase();
}

function getBudgetStatusCategory(statusName: string): BudgetStatusCategory {
  if (isWonStatusLabel(statusName)) {
    return "pedido";
  }

  const normalizedStatusName = normalizeValue(statusName);

  if (normalizedStatusName === "cancelado") {
    return "cancelado";
  }

  if (normalizedStatusName === "orcamento") {
    return "orcamento";
  }

  return "other";
}

function createNameMap<T extends { id: number; name: string }>(items: T[]) {
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

  return catalogMap.get(value) ?? `#${value}`;
}

function getErrorMessage(error: unknown, fallbackMessage: string) {
  if (isAxiosError<{ message?: string }>(error)) {
    return error.response?.data?.message ?? fallbackMessage;
  }

  return fallbackMessage;
}

function formatHistoryTransition(
  item: BudgetStatusHistoryItem,
  statusMap: Map<number, string>,
) {
  const fromStatusLabel =
    item.fromStatusId === null
      ? "Status inicial"
      : formatCatalogName(item.fromStatusId, statusMap);
  const toStatusLabel = formatCatalogName(item.toStatusId, statusMap);

  return `${fromStatusLabel} -> ${toStatusLabel}`;
}

const projectBudgetsFilters = (projectId: number): BudgetListFilters => ({
  budgetNumber: "",
  grossValueMax: "",
  grossValueMin: "",
  installerId: "",
  priorityId: "",
  page: 1,
  pageSize: 100,
  projectId: String(projectId),
  projectCode: "",
  projectName: "",
  salespersonId: "",
  estimatorId: "",
  sentAtFrom: "",
  sentAtTo: "",
  systemTypeId: "",
  sourceCompany: "",
  sortBy: "sent_at",
  sortOrder: "desc",
  statusId: "",
  yearBudget: "",
});

const associationCandidateFilters: BudgetListFilters = {
  budgetNumber: "",
  grossValueMax: "",
  grossValueMin: "",
  installerId: "",
  priorityId: "",
  page: 1,
  pageSize: 100,
  projectCode: "",
  projectName: "",
  salespersonId: "",
  estimatorId: "",
  sentAtFrom: "",
  sentAtTo: "",
  systemTypeId: "",
  sourceCompany: "",
  sortBy: "updated_at",
  sortOrder: "desc",
  statusId: "",
  yearBudget: "",
};

const premiumFeedbackAlertSx = {
  borderRadius: 3,
  boxShadow: "0 14px 30px rgba(30, 58, 138, 0.08)",
  "& .MuiAlert-message": {
    fontWeight: 600,
  },
} as const;

const premiumDialogSx = {
  "& .MuiDialog-paper": {
    backdropFilter: "blur(12px)",
    background:
      "linear-gradient(180deg, rgba(255,255,255,0.98) 0%, rgba(248,250,252,0.98) 100%)",
    border: "1px solid rgba(30, 58, 138, 0.12)",
    borderRadius: 4,
    boxShadow: "0 28px 60px rgba(15, 23, 42, 0.18)",
    overflow: "hidden",
  },
} as const;

const premiumDialogTitleSx = {
  background:
    "linear-gradient(135deg, rgba(30,58,138,0.09) 0%, rgba(14,165,233,0.045) 100%)",
  borderBottom: "1px solid rgba(30, 58, 138, 0.1)",
  color: "var(--app-accent-text)",
  fontWeight: 800,
  px: 3,
  py: 2.25,
} as const;

const premiumDialogActionsSx = {
  borderTop: "1px solid rgba(30, 58, 138, 0.08)",
  px: 3,
  pb: 3,
  pt: 2,
} as const;

function mapBudgetDetailToPayload(
  budget: BudgetDetailItem,
  projectId: number | null,
): BudgetCreatePayload {
  return {
    areaM2: budget.areaM2,
    budgetNumber: budget.budgetNumber,
    commissionValue: budget.commissionValue,
    constructionCompany: budget.constructionCompany,
    competitorName: budget.competitorName,
    competitorPrice: budget.competitorPrice,
    contactId: budget.contactId,
    currentFollowUp: budget.currentFollowUp,
    deliveryDate: budget.deliveryDate,
    projetistaName: budget.projetistaName,
    grossValue: budget.grossValue,
    installerId: budget.installerId,
    lossReasonId: budget.lossReasonId,
    priorityId: budget.priorityId,
    productLineId: budget.productLineId,
    systemTypeId: budget.systemTypeId,
    projectId,
    revision: budget.revision,
    salespersonId: budget.salespersonId,
    estimatorId: budget.estimatorId,
    sentAt: budget.sentAt,
    specificationDetails: budget.specificationDetails,
    statusId: budget.statusId,
    yearBudget: budget.yearBudget,
  };
}

export default function ProjectDetailPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { user } = useAuth();
  const { projectId: rawProjectId } = useParams();
  const projectId = Number(rawProjectId);
  const hasValidProjectId = Number.isInteger(projectId) && projectId > 0;
  const canCreateBudget =
    user?.role === "admin" ||
    (user?.role === "user" && user.user_kind === "estimator");
  const [associationDialogOpen, setAssociationDialogOpen] = useState(false);
  const [associationCandidateId, setAssociationCandidateId] = useState("");
  const [associationFeedback, setAssociationFeedback] = useState<string | null>(
    null,
  );
  const [associationError, setAssociationError] = useState<string | null>(null);
  const [copyFeedback, setCopyFeedback] = useState<string | null>(null);
  const [copyError, setCopyError] = useState<string | null>(null);
  const [candidateSearch, setCandidateSearch] = useState("");
  const [pendingDisassociationBudget, setPendingDisassociationBudget] =
    useState<BudgetListItem | null>(null);
  const [pendingWinnerBudget, setPendingWinnerBudget] =
    useState<BudgetListItem | null>(null);
  const [winnerNotes, setWinnerNotes] = useState("");

  const projectQuery = useQuery({
    enabled: hasValidProjectId,
    queryFn: () => getProjectByIdRequest(projectId),
    queryKey: ["project", projectId],
  });

  const projectTypesQuery = useQuery({
    enabled: hasValidProjectId,
    queryFn: listProjectTypesRequest,
    queryKey: ["project-types"],
  });

  const budgetCatalogsQuery = useQuery({
    enabled: hasValidProjectId,
    queryFn: getBudgetListCatalogsRequest,
    queryKey: ["budget-list-catalogs"],
  });

  const projectBudgetsQuery = useQuery({
    enabled: hasValidProjectId,
    queryFn: () => getBudgetListRequest(projectBudgetsFilters(projectId)),
    queryKey: ["project-budgets", projectId],
  });
  const associationCandidatesQuery = useQuery({
    enabled: hasValidProjectId && associationDialogOpen,
    queryFn: () => getBudgetListRequest(associationCandidateFilters),
    queryKey: ["budget-association-candidates"],
  });
  const projectBudgets = projectBudgetsQuery.data?.items ?? [];
  const budgetHistoryQueries = useQueries({
    queries: projectBudgets.map((budget) => ({
      enabled: hasValidProjectId,
      queryFn: () => getBudgetStatusHistoryRequest(budget.id),
      queryKey: ["budget-status-history", budget.id],
    })),
  });

  const projectTypeMap = useMemo(
    () => createNameMap<ProjectTypeCatalogItem>(projectTypesQuery.data ?? []),
    [projectTypesQuery.data],
  );
  const statusMap = useMemo(
    () =>
      new Map(
        (budgetCatalogsQuery.data?.statuses ?? []).map((item) => [
          item.id,
          getBudgetStatusDisplayName(item.name),
        ]),
      ),
    [budgetCatalogsQuery.data?.statuses],
  );
  const priorityMap = useMemo(
    () =>
      createNameMap<BudgetCatalogItem>(
        budgetCatalogsQuery.data?.priorities ?? [],
      ),
    [budgetCatalogsQuery.data?.priorities],
  );
  const installerMap = useMemo(
    () =>
      createNameMap<BudgetCatalogItem>(
        budgetCatalogsQuery.data?.installers ?? [],
      ),
    [budgetCatalogsQuery.data?.installers],
  );
  const systemTypeMap = useMemo(
    () =>
      createNameMap<BudgetCatalogItem>(
        budgetCatalogsQuery.data?.systemTypes ?? [],
      ),
    [budgetCatalogsQuery.data?.systemTypes],
  );

  const winnerBudget = useMemo(
    () =>
      projectBudgets.find((budget) => {
        const statusName = formatCatalogName(budget.statusId, statusMap);
        return getBudgetStatusCategory(statusName) === "pedido";
      }) ?? null,
    [projectBudgets, statusMap],
  );
  const cancelledBudgetsCount = useMemo(
    () =>
      projectBudgets.filter((budget) => {
        const statusName = formatCatalogName(budget.statusId, statusMap);
        return getBudgetStatusCategory(statusName) === "cancelado";
      }).length,
    [projectBudgets, statusMap],
  );
  const budgetsRequiringAttentionCount = useMemo(
    () =>
      projectBudgets.filter((budget) => {
        const statusName = formatCatalogName(budget.statusId, statusMap);
        return getBudgetStatusCategory(statusName) !== "cancelado";
      }).length,
    [projectBudgets, statusMap],
  );
  const totalGrossValue = useMemo(
    () =>
      projectBudgets.reduce(
        (currentTotal, budget) => currentTotal + budget.grossValue,
        0,
      ),
    [projectBudgets],
  );
  const availableAssociationCandidates = useMemo(
    () =>
      (associationCandidatesQuery.data?.items ?? [])
        .filter((budget) => budget.projectId === null)
        .filter((budget) => {
          const normalizedSearch = normalizeValue(candidateSearch);
          if (!normalizedSearch) {
            return true;
          }

          const searchableContent = normalizeValue(
            [
              budget.budgetNumber,
              budget.projetistaName,
              budget.competitorName,
              budget.salespersonName ?? "",
            ].join(" "),
          );

          return searchableContent.includes(normalizedSearch);
        }),
    [associationCandidatesQuery.data?.items, candidateSearch],
  );

  const invalidateProjectQueries = async (budgetId?: number) => {
    await queryClient.invalidateQueries({ queryKey: ["budgets"] });
    if (budgetId !== undefined) {
      await queryClient.invalidateQueries({ queryKey: ["budget", budgetId] });
    }
    await queryClient.invalidateQueries({
      queryKey: ["project-budgets", projectId],
    });
    await queryClient.invalidateQueries({ queryKey: ["project", projectId] });
    await queryClient.invalidateQueries({
      queryKey: ["budget-status-history"],
    });
    await queryClient.invalidateQueries({
      queryKey: ["budget-association-candidates"],
    });
  };

  const associateBudgetMutation = useMutation({
    mutationFn: async (budgetId: number) => {
      const budget = await getBudgetByIdRequest(budgetId);
      await updateBudgetRequest(
        budgetId,
        mapBudgetDetailToPayload(budget, projectId),
      );
      return budget;
    },
    onSuccess: async (budget) => {
      await invalidateProjectQueries(budget.id);
      setAssociationError(null);
      setAssociationFeedback(
        `Orçamento ${budget.budgetNumber} associado à obra com sucesso.`,
      );
      setAssociationCandidateId("");
      setCandidateSearch("");
      setAssociationDialogOpen(false);
    },
    onError: (error) => {
      setAssociationFeedback(null);
      setAssociationError(
        getErrorMessage(error, "Não foi possível associar o orçamento à obra."),
      );
    },
  });

  const disassociateBudgetMutation = useMutation({
    mutationFn: async (budget: BudgetListItem) => {
      const currentBudget = await getBudgetByIdRequest(budget.id);
      await updateBudgetRequest(
        budget.id,
        mapBudgetDetailToPayload(currentBudget, null),
      );
      return budget;
    },
    onSuccess: async (budget) => {
      await invalidateProjectQueries(budget.id);
      setAssociationError(null);
      setAssociationFeedback(
        `Orçamento ${budget.budgetNumber} removido da obra com sucesso.`,
      );
      setPendingDisassociationBudget(null);
    },
    onError: (error) => {
      setAssociationFeedback(null);
      setAssociationError(
        getErrorMessage(error, "Não foi possível remover o orçamento da obra."),
      );
      setPendingDisassociationBudget(null);
    },
  });

  const electWinnerMutation = useMutation({
    mutationFn: async (budget: BudgetListItem) => {
      await electBudgetWinnerRequest(budget.id, {
        notes: winnerNotes.trim(),
      });
      return budget;
    },
    onSuccess: async (budget) => {
      await invalidateProjectQueries(budget.id);
      setAssociationError(null);
      setAssociationFeedback(
        winnerBudget && winnerBudget.id !== budget.id
          ? `Vencedor da obra atualizado para o orçamento ${budget.budgetNumber} com sucesso.`
          : `Orçamento ${budget.budgetNumber} definido como vencedor da obra com sucesso.`,
      );
      setPendingWinnerBudget(null);
      setWinnerNotes("");
    },
    onError: (error) => {
      setAssociationFeedback(null);
      setAssociationError(
        getErrorMessage(error, "Não foi possível definir o vencedor da obra."),
      );
    },
  });

  const handleAssociateBudget = async () => {
    const selectedBudgetId = Number(associationCandidateId);

    if (!Number.isInteger(selectedBudgetId) || selectedBudgetId <= 0) {
      setAssociationFeedback(null);
      setAssociationError("Selecione um orçamento válido para associar.");
      return;
    }

    await associateBudgetMutation.mutateAsync(selectedBudgetId);
  };

  if (!hasValidProjectId) {
    return (
      <Box sx={{ p: { md: 3, xs: 2 } }}>
        <Alert severity="error">Obra inválida.</Alert>
      </Box>
    );
  }

  const isLoading =
    projectQuery.isLoading ||
    projectTypesQuery.isLoading ||
    budgetCatalogsQuery.isLoading ||
    projectBudgetsQuery.isLoading;

  const hasError =
    projectQuery.isError ||
    projectTypesQuery.isError ||
    budgetCatalogsQuery.isError ||
    projectBudgetsQuery.isError;

  const handleCopyProjectCode = async () => {
    if (!project?.code.trim()) {
      return;
    }

    try {
      await navigator.clipboard.writeText(project.code);
      setCopyError(null);
      setCopyFeedback("Código da obra copiado.");
    } catch {
      setCopyFeedback(null);
      setCopyError("Não foi possível copiar o código da obra.");
    }
  };

  const project = projectQuery.data;

  return (
    <Box sx={{ display: "flex", flexDirection: "column", gap: 3 }}>
      <PageHeader
        action={
          <Box sx={{ display: "flex", flexWrap: "wrap", gap: 1 }}>
            {canCreateBudget ? (
              <Button
                onClick={() =>
                  navigate(
                    `/budgets/new?projectId=${projectId}&returnTo=project`,
                  )
                }
                startIcon={<NoteAddRoundedIcon />}
                variant="contained"
              >
                Novo orçamento
              </Button>
            ) : null}
            <Button
              onClick={() =>
                navigate(
                  `/communication?tab=conversations&projectId=${projectId}`,
                )
              }
              startIcon={<ForumRoundedIcon />}
              variant="outlined"
            >
              Conversar sobre a obra
            </Button>
            <Button
              onClick={() => {
                setAssociationFeedback(null);
                setAssociationError(null);
                setAssociationDialogOpen(true);
              }}
              startIcon={<LinkRoundedIcon />}
              variant="outlined"
            >
              Associar orçamento
            </Button>
            <Button
              onClick={() => navigate("/budgets")}
              startIcon={<ArrowBackRoundedIcon />}
              variant="outlined"
            >
              Voltar para orçamentos
            </Button>
          </Box>
        }
        description="Acompanhe os orçamentos vinculados à obra, o vencedor definido e os itens que já foram encerrados."
        title={project?.name ?? "Obra"}
      />

      {isLoading ? <LinearProgress /> : null}

      {associationFeedback ? (
        <Alert
          severity="success"
          sx={premiumFeedbackAlertSx}
          variant="outlined"
        >
          {associationFeedback}
        </Alert>
      ) : null}

      {copyFeedback ? (
        <Alert
          severity="success"
          sx={premiumFeedbackAlertSx}
          variant="outlined"
        >
          {copyFeedback}
        </Alert>
      ) : null}

      {associationError ? (
        <Alert severity="error" sx={premiumFeedbackAlertSx} variant="outlined">
          {associationError}
        </Alert>
      ) : null}

      {copyError ? (
        <Alert severity="error" sx={premiumFeedbackAlertSx} variant="outlined">
          {copyError}
        </Alert>
      ) : null}

      {hasError ? (
        <Alert severity="error" sx={premiumFeedbackAlertSx} variant="outlined">
          {getErrorMessage(
            projectQuery.error ??
              projectTypesQuery.error ??
              budgetCatalogsQuery.error ??
              projectBudgetsQuery.error,
            "Não foi possível carregar os dados da obra.",
          )}
        </Alert>
      ) : null}

      {project ? (
        <>
          <SectionCard
            description="Dados principais da obra e indicadores consolidados do grupo de orçamentos."
            title="Resumo da obra"
          >
            <Box
              sx={{
                display: "grid",
                gap: 1.75,
                gridTemplateColumns: {
                  lg: "repeat(4, minmax(0, 1fr))",
                  md: "repeat(2, minmax(0, 1fr))",
                  xs: "minmax(0, 1fr)",
                },
              }}
            >
              <Box
                sx={{
                  background: (theme) =>
                    `linear-gradient(135deg, ${alpha(theme.palette.primary.main, 0.06)} 0%, ${alpha(theme.palette.info.main, 0.025)} 100%)`,
                  border: "1px solid",
                  borderColor: (theme) =>
                    alpha(theme.palette.primary.main, 0.12),
                  borderRadius: 3,
                  p: 2,
                }}
              >
                <Typography
                  color="text.secondary"
                  sx={{ fontWeight: 700 }}
                  variant="caption"
                >
                  Código
                </Typography>
                <Box sx={{ alignItems: "center", display: "flex", gap: 0.5 }}>
                  <Typography sx={{ fontWeight: 700 }} variant="body1">
                    {project.code}
                  </Typography>
                  <Tooltip title="Copiar código">
                    <IconButton
                      aria-label="Copiar código"
                      onClick={handleCopyProjectCode}
                      size="small"
                    >
                      <ContentCopyRoundedIcon fontSize="small" />
                    </IconButton>
                  </Tooltip>
                </Box>
              </Box>
              <Box
                sx={{
                  background: (theme) =>
                    `linear-gradient(135deg, ${alpha(theme.palette.primary.main, 0.06)} 0%, ${alpha(theme.palette.info.main, 0.025)} 100%)`,
                  border: "1px solid",
                  borderColor: (theme) =>
                    alpha(theme.palette.primary.main, 0.12),
                  borderRadius: 3,
                  p: 2,
                }}
              >
                <Typography
                  color="text.secondary"
                  sx={{ fontWeight: 700 }}
                  variant="caption"
                >
                  Obra
                </Typography>
                <Typography sx={{ fontWeight: 700 }} variant="body1">
                  {project.name}
                </Typography>
              </Box>
              <Box
                sx={{
                  background: (theme) =>
                    `linear-gradient(135deg, ${alpha(theme.palette.primary.main, 0.06)} 0%, ${alpha(theme.palette.info.main, 0.025)} 100%)`,
                  border: "1px solid",
                  borderColor: (theme) =>
                    alpha(theme.palette.primary.main, 0.12),
                  borderRadius: 3,
                  p: 2,
                }}
              >
                <Typography
                  color="text.secondary"
                  sx={{ fontWeight: 700 }}
                  variant="caption"
                >
                  Tipo de obra
                </Typography>
                <Typography sx={{ fontWeight: 700 }} variant="body1">
                  {formatCatalogName(project.projectTypeId, projectTypeMap)}
                </Typography>
              </Box>
              <Box
                sx={{
                  background: (theme) =>
                    `linear-gradient(135deg, ${alpha(theme.palette.primary.main, 0.06)} 0%, ${alpha(theme.palette.info.main, 0.025)} 100%)`,
                  border: "1px solid",
                  borderColor: (theme) =>
                    alpha(theme.palette.primary.main, 0.12),
                  borderRadius: 3,
                  p: 2,
                }}
              >
                <Typography
                  color="text.secondary"
                  sx={{ fontWeight: 700 }}
                  variant="caption"
                >
                  Cidade / Estado
                </Typography>
                <Typography sx={{ fontWeight: 700 }} variant="body1">
                  {project.city || project.state
                    ? [project.city, project.state].filter(Boolean).join(" / ")
                    : "Não informado"}
                </Typography>
              </Box>
              <Box
                sx={{
                  background: (theme) =>
                    `linear-gradient(135deg, ${alpha(theme.palette.primary.main, 0.06)} 0%, ${alpha(theme.palette.info.main, 0.025)} 100%)`,
                  border: "1px solid",
                  borderColor: (theme) =>
                    alpha(theme.palette.primary.main, 0.12),
                  borderRadius: 3,
                  p: 2,
                }}
              >
                <Typography
                  color="text.secondary"
                  sx={{ fontWeight: 700 }}
                  variant="caption"
                >
                  Última atualização
                </Typography>
                <Typography sx={{ fontWeight: 700 }} variant="body1">
                  {dateTimeFormatter.format(new Date(project.updatedAt))}
                </Typography>
              </Box>
            </Box>

            <Box
              sx={{
                display: "grid",
                gap: 1.25,
                gridTemplateColumns: {
                  lg: "repeat(5, minmax(0, 1fr))",
                  md: "repeat(3, minmax(0, 1fr))",
                  xs: "repeat(2, minmax(0, 1fr))",
                },
              }}
            >
              <Chip
                color={winnerBudget ? "success" : "warning"}
                label={
                  winnerBudget
                    ? `Vencedor: ${winnerBudget.budgetNumber}`
                    : `Sem ${getWonStatusSingularLabel().toUpperCase()} definido`
                }
              />
              <Chip
                label={`${projectBudgets.length} orçamento(s)`}
                variant="outlined"
              />
              <Chip
                label={`${cancelledBudgetsCount} cancelado(s)`}
                variant="outlined"
              />
              <Chip
                label={`${budgetsRequiringAttentionCount} com atenção`}
                variant="outlined"
              />
              <Chip
                label={`Valor total ${currencyFormatter.format(totalGrossValue)}`}
                variant="outlined"
              />
            </Box>

            <Divider />

            <Box>
              <Typography color="text.secondary" variant="caption">
                Observações
              </Typography>
              <Typography sx={{ whiteSpace: "pre-wrap" }} variant="body2">
                {formatOptionalText(project.notes)}
              </Typography>
            </Box>
          </SectionCard>

          <SectionCard
            description="Lista de orçamentos atualmente associados à obra, com ações para editar, abrir na listagem e remover o vínculo."
            title="Orçamentos vinculados"
          >
            {projectBudgets.length === 0 ? (
              <Alert severity="info" variant="outlined">
                Nenhum orçamento vinculado a esta obra até o momento.
              </Alert>
            ) : (
              <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
                {projectBudgets.map((budget: BudgetListItem, index) => {
                  const statusName = formatCatalogName(
                    budget.statusId,
                    statusMap,
                  );
                  const budgetHistoryQuery = budgetHistoryQueries[index];
                  const budgetHistory = budgetHistoryQuery?.data ?? [];
                  const statusCategory = getBudgetStatusCategory(statusName);
                  const isWinner = winnerBudget?.id === budget.id;
                  const isWinnerReplacement =
                    winnerBudget !== null && winnerBudget.id !== budget.id;

                  return (
                    <Box
                      key={budget.id}
                      sx={{
                        backgroundColor: (theme) => {
                          if (isWinner) {
                            return `${theme.palette.success.main}12`;
                          }

                          if (statusCategory === "cancelado") {
                            return `${theme.palette.grey[500]}12`;
                          }

                          return theme.palette.background.paper;
                        },
                        border: (theme) =>
                          isWinner
                            ? `1px solid ${theme.palette.success.main}`
                            : `1px solid ${theme.palette.divider}`,
                        borderRadius: 3,
                        p: 2,
                      }}
                    >
                      <Box
                        sx={{
                          alignItems: { md: "center", xs: "flex-start" },
                          display: "flex",
                          flexDirection: { md: "row", xs: "column" },
                          gap: 1.5,
                          justifyContent: "space-between",
                          mb: 1.5,
                        }}
                      >
                        <Box sx={{ minWidth: 0 }}>
                          <Typography sx={{ fontWeight: 700 }} variant="h6">
                            {budget.budgetNumber}
                          </Typography>
                          <Typography color="text.secondary" variant="body2">
                            Enviado em{" "}
                            {dateFormatter.format(new Date(budget.sentAt))}
                          </Typography>
                        </Box>

                        <Box sx={{ display: "flex", flexWrap: "wrap", gap: 1 }}>
                          <Chip
                            color={
                              isWinner
                                ? "success"
                                : statusCategory === "cancelado"
                                  ? "default"
                                  : "primary"
                            }
                            label={statusName}
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
                          display: "grid",
                          gap: 1.5,
                          gridTemplateColumns: {
                            lg: "repeat(4, minmax(0, 1fr))",
                            md: "repeat(2, minmax(0, 1fr))",
                            xs: "minmax(0, 1fr)",
                          },
                          mb: 1.5,
                        }}
                      >
                        <Box>
                          <Typography color="text.secondary" variant="caption">
                            Valor bruto
                          </Typography>
                          <Typography variant="body2">
                            {currencyFormatter.format(budget.grossValue)}
                          </Typography>
                        </Box>
                        <Box>
                          <Typography color="text.secondary" variant="caption">
                            Prioridade
                          </Typography>
                          <Typography variant="body2">
                            {formatCatalogName(budget.priorityId, priorityMap)}
                          </Typography>
                        </Box>
                        <Box>
                          <Typography color="text.secondary" variant="caption">
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
                          <Typography color="text.secondary" variant="caption">
                            Tipo de Sistema
                          </Typography>
                          <Typography variant="body2">
                            {formatCatalogName(
                              budget.systemTypeId,
                              systemTypeMap,
                            )}
                          </Typography>
                        </Box>
                        <Box>
                          <Typography color="text.secondary" variant="caption">
                            Vendedor
                          </Typography>
                          <Typography variant="body2">
                            {formatOptionalText(budget.salespersonName)}
                          </Typography>
                        </Box>
                        <Box>
                          <Typography color="text.secondary" variant="caption">
                            Contato
                          </Typography>
                          <Typography variant="body2">
                            {formatOptionalText(budget.contactName)}
                          </Typography>
                        </Box>
                        <Box>
                          <Typography color="text.secondary" variant="caption">
                            Projetista
                          </Typography>
                          <Typography variant="body2">
                            {formatOptionalText(budget.projetistaName)}
                          </Typography>
                        </Box>
                        <Box>
                          <Typography color="text.secondary" variant="caption">
                            Concorrente
                          </Typography>
                          <Typography variant="body2">
                            {formatOptionalText(budget.competitorName)}
                          </Typography>
                        </Box>
                        <Box>
                          <Typography color="text.secondary" variant="caption">
                            Follow-up atual
                          </Typography>
                          <Typography variant="body2">
                            {formatOptionalText(budget.currentFollowUp)}
                          </Typography>
                        </Box>
                      </Box>

                      <Box
                        sx={{
                          display: "flex",
                          flexWrap: "wrap",
                          gap: 1,
                          justifyContent: "flex-end",
                        }}
                      >
                        <Button
                          color={isWinner ? "success" : "primary"}
                          onClick={() => {
                            setAssociationFeedback(null);
                            setAssociationError(null);
                            setPendingWinnerBudget(budget);
                            setWinnerNotes("");
                          }}
                          size="small"
                          startIcon={<EmojiEventsRoundedIcon />}
                          variant={isWinner ? "contained" : "outlined"}
                        >
                          {isWinner
                            ? "Reaplicar vencedor"
                            : isWinnerReplacement
                              ? "Trocar vencedor para este"
                              : "Eleger vencedor"}
                        </Button>
                        <Button
                          onClick={() => navigate(`/budgets/${budget.id}/edit`)}
                          size="small"
                          startIcon={<EditRoundedIcon />}
                          variant="text"
                        >
                          Editar orçamento
                        </Button>
                        <Button
                          onClick={() => navigate("/budgets")}
                          size="small"
                          startIcon={<FolderOpenRoundedIcon />}
                          variant="text"
                        >
                          Ver na listagem
                        </Button>
                        <Button
                          color="error"
                          onClick={() => setPendingDisassociationBudget(budget)}
                          size="small"
                          startIcon={<LinkOffRoundedIcon />}
                          variant="text"
                        >
                          Remover da obra
                        </Button>
                      </Box>

                      <Divider sx={{ my: 1.5 }} />

                      <Box
                        sx={{
                          display: "flex",
                          flexDirection: "column",
                          gap: 1,
                        }}
                      >
                        <Box
                          sx={{
                            alignItems: "center",
                            display: "flex",
                            gap: 1,
                          }}
                        >
                          <TimelineRoundedIcon
                            color="action"
                            fontSize="small"
                          />
                          <Typography
                            sx={{ fontWeight: 700 }}
                            variant="subtitle2"
                          >
                            Histórico de status
                          </Typography>
                        </Box>

                        {budgetHistoryQuery?.isLoading ? (
                          <Typography color="text.secondary" variant="body2">
                            Carregando histórico...
                          </Typography>
                        ) : budgetHistoryQuery?.isError ? (
                          <Alert severity="warning" variant="outlined">
                            Não foi possível carregar o histórico deste
                            orçamento.
                          </Alert>
                        ) : budgetHistory.length === 0 ? (
                          <Typography color="text.secondary" variant="body2">
                            Nenhuma alteração de status registrada até o
                            momento.
                          </Typography>
                        ) : (
                          <List disablePadding>
                            {budgetHistory.map((historyItem) => (
                              <ListItem
                                disableGutters
                                divider
                                key={historyItem.id}
                                sx={{ alignItems: "flex-start", py: 1 }}
                              >
                                <ListItemText
                                  primary={
                                    <Typography
                                      sx={{
                                        fontSize: "0.9rem",
                                        fontWeight: 600,
                                      }}
                                      variant="body2"
                                    >
                                      {formatHistoryTransition(
                                        historyItem,
                                        statusMap,
                                      )}
                                    </Typography>
                                  }
                                  secondary={
                                    <Box
                                      sx={{
                                        display: "flex",
                                        flexDirection: "column",
                                        gap: 0.5,
                                        mt: 0.5,
                                      }}
                                    >
                                      <Typography
                                        color="text.secondary"
                                        variant="caption"
                                      >
                                        Alterado em{" "}
                                        {dateTimeFormatter.format(
                                          new Date(historyItem.changedAt),
                                        )}{" "}
                                        por usuário #
                                        {historyItem.changedByUserId}
                                      </Typography>
                                      {historyItem.notes ? (
                                        <Typography variant="body2">
                                          {historyItem.notes}
                                        </Typography>
                                      ) : null}
                                    </Box>
                                  }
                                />
                              </ListItem>
                            ))}
                          </List>
                        )}
                      </Box>
                    </Box>
                  );
                })}
              </Box>
            )}
          </SectionCard>
        </>
      ) : null}

      <Dialog
        fullWidth
        maxWidth="sm"
        onClose={() => {
          if (associateBudgetMutation.isPending) {
            return;
          }

          setAssociationDialogOpen(false);
          setAssociationCandidateId("");
          setCandidateSearch("");
        }}
        open={associationDialogOpen}
        sx={premiumDialogSx}
      >
        <DialogTitle sx={premiumDialogTitleSx}>
          Associar orçamento à obra
        </DialogTitle>
        <DialogContent
          sx={{ display: "flex", flexDirection: "column", gap: 2.5, pt: 3 }}
        >
          <DialogContentText sx={{ color: "text.primary", lineHeight: 1.7 }}>
            Selecione um orçamento sem obra vinculada para adicioná-lo a esta
            obra.
          </DialogContentText>
          <Box sx={filterGroupSx}>
            <Typography sx={filterGroupTitleSx} variant="subtitle2">
              Identificação
            </Typography>
            <FilterField label="Buscar por número ou responsável">
              <TextField
                onChange={(event) => setCandidateSearch(event.target.value)}
                size="small"
                sx={compactFilterFieldSx}
                value={candidateSearch}
              />
            </FilterField>
            <FilterField label="Orçamento disponível">
              <TextField
                helperText={
                  associationCandidatesQuery.isLoading
                    ? "Carregando orçamentos disponíveis..."
                    : availableAssociationCandidates.length === 0
                      ? "Nenhum orçamento disponível para associação."
                      : "Selecione um orçamento da lista."
                }
                onChange={(event) =>
                  setAssociationCandidateId(event.target.value)
                }
                select
                size="small"
                sx={compactFilterFieldSx}
                value={associationCandidateId}
              >
                <MenuItem value="">Selecione</MenuItem>
                {availableAssociationCandidates.map((budget) => (
                  <MenuItem key={budget.id} value={String(budget.id)}>
                    {budget.budgetNumber} -{" "}
                    {formatOptionalText(budget.salespersonName)}
                  </MenuItem>
                ))}
              </TextField>
            </FilterField>
          </Box>
        </DialogContent>
        <DialogActions sx={premiumDialogActionsSx}>
          <Button
            disabled={associateBudgetMutation.isPending}
            onClick={() => {
              setAssociationDialogOpen(false);
              setAssociationCandidateId("");
              setCandidateSearch("");
            }}
          >
            Cancelar
          </Button>
          <Button
            disabled={
              associateBudgetMutation.isPending || associationCandidateId === ""
            }
            onClick={() => {
              void handleAssociateBudget();
            }}
            startIcon={<LinkRoundedIcon />}
            variant="contained"
          >
            Associar
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog
        fullWidth
        maxWidth="sm"
        onClose={() => {
          if (electWinnerMutation.isPending) {
            return;
          }

          setPendingWinnerBudget(null);
          setWinnerNotes("");
        }}
        open={pendingWinnerBudget !== null}
        sx={premiumDialogSx}
      >
        <DialogTitle sx={premiumDialogTitleSx}>
          {winnerBudget &&
          pendingWinnerBudget &&
          winnerBudget.id !== pendingWinnerBudget.id
            ? "Trocar vencedor da obra"
            : "Eleger vencedor da obra"}
        </DialogTitle>
        <DialogContent
          sx={{ display: "flex", flexDirection: "column", gap: 2.5, pt: 3 }}
        >
          <DialogContentText sx={{ color: "text.primary", lineHeight: 1.7 }}>
            {pendingWinnerBudget
              ? winnerBudget && winnerBudget.id !== pendingWinnerBudget.id
                ? `Ao confirmar, o orçamento ${pendingWinnerBudget.budgetNumber} passará a ser o novo vencedor da obra. O vencedor atual será substituído e os demais orçamentos da obra ficarão como Cancelado.`
                : `Ao confirmar, o orçamento ${pendingWinnerBudget.budgetNumber} será definido como ${getWonStatusSingularLabel()} e os demais orçamentos desta obra serão alterados para Cancelado.`
              : ""}
          </DialogContentText>
          <TextField
            label={
              winnerBudget &&
              pendingWinnerBudget &&
              winnerBudget.id !== pendingWinnerBudget.id
                ? "Observação da troca"
                : "Observação da eleição"
            }
            multiline
            minRows={3}
            onChange={(event) => setWinnerNotes(event.target.value)}
            placeholder="Opcional. Ex: Orçamento escolhido como vencedor da obra."
            value={winnerNotes}
          />
        </DialogContent>
        <DialogActions sx={premiumDialogActionsSx}>
          <Button
            disabled={electWinnerMutation.isPending}
            onClick={() => {
              setPendingWinnerBudget(null);
              setWinnerNotes("");
            }}
          >
            Cancelar
          </Button>
          <Button
            color="success"
            disabled={
              electWinnerMutation.isPending || pendingWinnerBudget === null
            }
            onClick={() => {
              if (pendingWinnerBudget === null) {
                return;
              }

              void electWinnerMutation.mutateAsync(pendingWinnerBudget);
            }}
            startIcon={<EmojiEventsRoundedIcon />}
            variant="contained"
          >
            Confirmar vencedor
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog
        fullWidth
        maxWidth="xs"
        onClose={() => {
          if (disassociateBudgetMutation.isPending) {
            return;
          }

          setPendingDisassociationBudget(null);
        }}
        open={pendingDisassociationBudget !== null}
        sx={premiumDialogSx}
      >
        <DialogTitle sx={premiumDialogTitleSx}>
          Remover orçamento da obra
        </DialogTitle>
        <DialogContent sx={{ pt: 3 }}>
          <DialogContentText sx={{ color: "text.primary", lineHeight: 1.7 }}>
            {pendingDisassociationBudget
              ? `Confirma a remoção do orçamento ${pendingDisassociationBudget.budgetNumber} desta obra?`
              : ""}
          </DialogContentText>
        </DialogContent>
        <DialogActions sx={premiumDialogActionsSx}>
          <Button
            disabled={disassociateBudgetMutation.isPending}
            onClick={() => setPendingDisassociationBudget(null)}
          >
            Cancelar
          </Button>
          <Button
            color="error"
            disabled={
              disassociateBudgetMutation.isPending ||
              pendingDisassociationBudget === null
            }
            onClick={() => {
              if (pendingDisassociationBudget === null) {
                return;
              }

              void disassociateBudgetMutation.mutateAsync(
                pendingDisassociationBudget,
              );
            }}
            startIcon={<LinkOffRoundedIcon />}
            variant="contained"
          >
            Remover
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
