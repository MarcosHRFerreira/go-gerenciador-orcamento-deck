import ArrowBackRoundedIcon from "@mui/icons-material/ArrowBackRounded";
import EditRoundedIcon from "@mui/icons-material/EditRounded";
import FolderOpenRoundedIcon from "@mui/icons-material/FolderOpenRounded";
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
  List,
  ListItem,
  ListItemText,
  LinearProgress,
  MenuItem,
  TextField,
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
import { PageHeader } from "../../../components/common/PageHeader";
import { SectionCard } from "../../../components/common/SectionCard";
import {
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

  return trimmedValue ? trimmedValue : "Nao informado";
}

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

function createNameMap<T extends { id: number; name: string }>(items: T[]) {
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
  installerId: "",
  page: 1,
  pageSize: 100,
  projectId: String(projectId),
  projectName: "",
  salespersonId: "",
  sourceCompany: "",
  sortBy: "sent_at",
  sortOrder: "desc",
  statusId: "",
  yearBudget: "",
});

const associationCandidateFilters: BudgetListFilters = {
  budgetNumber: "",
  installerId: "",
  page: 1,
  pageSize: 100,
  projectName: "",
  salespersonId: "",
  sourceCompany: "",
  sortBy: "updated_at",
  sortOrder: "desc",
  statusId: "",
  yearBudget: "",
};

function mapBudgetDetailToPayload(
  budget: BudgetDetailItem,
  projectId: number | null,
): BudgetCreatePayload {
  return {
    areaM2: budget.areaM2,
    budgetNumber: budget.budgetNumber,
    commissionValue: budget.commissionValue,
    competitorName: budget.competitorName,
    competitorPrice: budget.competitorPrice,
    contactId: budget.contactId,
    currentFollowUp: budget.currentFollowUp,
    designerName: budget.designerName,
    grossValue: budget.grossValue,
    installerId: budget.installerId,
    lossReasonId: budget.lossReasonId,
    priorityId: budget.priorityId,
    projectId,
    revision: budget.revision,
    salespersonId: budget.salespersonId,
    sentAt: budget.sentAt,
    specificationDetails: budget.specificationDetails,
    statusId: budget.statusId,
    yearBudget: budget.yearBudget,
  };
}

export default function ProjectDetailPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { projectId: rawProjectId } = useParams();
  const projectId = Number(rawProjectId);
  const hasValidProjectId = Number.isInteger(projectId) && projectId > 0;
  const [associationDialogOpen, setAssociationDialogOpen] = useState(false);
  const [associationCandidateId, setAssociationCandidateId] = useState("");
  const [associationFeedback, setAssociationFeedback] = useState<string | null>(
    null,
  );
  const [associationError, setAssociationError] = useState<string | null>(null);
  const [candidateSearch, setCandidateSearch] = useState("");
  const [pendingDisassociationBudget, setPendingDisassociationBudget] =
    useState<BudgetListItem | null>(null);

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
      createNameMap<BudgetCatalogItem>(
        budgetCatalogsQuery.data?.statuses ?? [],
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
              budget.designerName,
              budget.competitorName,
              budget.salespersonName ?? "",
            ].join(" "),
          );

          return searchableContent.includes(normalizedSearch);
        }),
    [associationCandidatesQuery.data?.items, candidateSearch],
  );

  const invalidateProjectQueries = async (budgetId: number) => {
    await queryClient.invalidateQueries({ queryKey: ["budgets"] });
    await queryClient.invalidateQueries({ queryKey: ["budget", budgetId] });
    await queryClient.invalidateQueries({
      queryKey: ["project-budgets", projectId],
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
        `Orcamento ${budget.budgetNumber} associado ao projeto com sucesso.`,
      );
      setAssociationCandidateId("");
      setCandidateSearch("");
      setAssociationDialogOpen(false);
    },
    onError: (error) => {
      setAssociationFeedback(null);
      setAssociationError(
        getErrorMessage(
          error,
          "Nao foi possivel associar o orcamento ao projeto.",
        ),
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
        `Orcamento ${budget.budgetNumber} removido do projeto com sucesso.`,
      );
      setPendingDisassociationBudget(null);
    },
    onError: (error) => {
      setAssociationFeedback(null);
      setAssociationError(
        getErrorMessage(
          error,
          "Nao foi possivel remover o orcamento do projeto.",
        ),
      );
      setPendingDisassociationBudget(null);
    },
  });

  const handleAssociateBudget = async () => {
    const selectedBudgetId = Number(associationCandidateId);

    if (!Number.isInteger(selectedBudgetId) || selectedBudgetId <= 0) {
      setAssociationFeedback(null);
      setAssociationError("Selecione um orcamento valido para associar.");
      return;
    }

    await associateBudgetMutation.mutateAsync(selectedBudgetId);
  };

  if (!hasValidProjectId) {
    return (
      <Box sx={{ p: { md: 3, xs: 2 } }}>
        <Alert severity="error">Projeto invalido.</Alert>
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

  const project = projectQuery.data;

  return (
    <Box sx={{ display: "flex", flexDirection: "column", gap: 3 }}>
      <PageHeader
        action={
          <Box sx={{ display: "flex", flexWrap: "wrap", gap: 1 }}>
            <Button
              onClick={() =>
                navigate(`/budgets/new?projectId=${projectId}&returnTo=project`)
              }
              startIcon={<NoteAddRoundedIcon />}
              variant="contained"
            >
              Novo orcamento
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
              Associar orcamento
            </Button>
            <Button
              onClick={() => navigate("/budgets")}
              startIcon={<ArrowBackRoundedIcon />}
              variant="outlined"
            >
              Voltar para orcamentos
            </Button>
          </Box>
        }
        description="Acompanhe os orcamentos vinculados ao projeto, o vencedor definido e os itens que ja foram encerrados."
        title={project?.name ?? "Projeto"}
      />

      {isLoading ? <LinearProgress /> : null}

      {associationFeedback ? (
        <Alert severity="success" variant="outlined">
          {associationFeedback}
        </Alert>
      ) : null}

      {associationError ? (
        <Alert severity="error" variant="outlined">
          {associationError}
        </Alert>
      ) : null}

      {hasError ? (
        <Alert severity="error" variant="outlined">
          {getErrorMessage(
            projectQuery.error ??
              projectTypesQuery.error ??
              budgetCatalogsQuery.error ??
              projectBudgetsQuery.error,
            "Nao foi possivel carregar os dados do projeto.",
          )}
        </Alert>
      ) : null}

      {project ? (
        <>
          <SectionCard
            description="Dados principais do projeto e indicadores consolidados do grupo de orcamentos."
            title="Resumo do projeto"
          >
            <Box
              sx={{
                display: "grid",
                gap: 2,
                gridTemplateColumns: {
                  lg: "repeat(4, minmax(0, 1fr))",
                  md: "repeat(2, minmax(0, 1fr))",
                  xs: "minmax(0, 1fr)",
                },
              }}
            >
              <Box>
                <Typography color="text.secondary" variant="caption">
                  Projeto
                </Typography>
                <Typography variant="body1">{project.name}</Typography>
              </Box>
              <Box>
                <Typography color="text.secondary" variant="caption">
                  Tipo de projeto
                </Typography>
                <Typography variant="body1">
                  {formatCatalogName(project.projectTypeId, projectTypeMap)}
                </Typography>
              </Box>
              <Box>
                <Typography color="text.secondary" variant="caption">
                  Cidade / Estado
                </Typography>
                <Typography variant="body1">
                  {project.city || project.state
                    ? [project.city, project.state].filter(Boolean).join(" / ")
                    : "Nao informado"}
                </Typography>
              </Box>
              <Box>
                <Typography color="text.secondary" variant="caption">
                  Ultima atualizacao
                </Typography>
                <Typography variant="body1">
                  {dateTimeFormatter.format(new Date(project.updatedAt))}
                </Typography>
              </Box>
            </Box>

            <Box sx={{ display: "flex", flexWrap: "wrap", gap: 1 }}>
              <Chip
                color={winnerBudget ? "success" : "warning"}
                label={
                  winnerBudget
                    ? `Vencedor: ${winnerBudget.budgetNumber}`
                    : "Sem PEDIDO definido"
                }
              />
              <Chip
                label={`${projectBudgets.length} orcamento(s)`}
                variant="outlined"
              />
              <Chip
                label={`${cancelledBudgetsCount} cancelado(s)`}
                variant="outlined"
              />
              <Chip
                label={`${budgetsRequiringAttentionCount} com atencao`}
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
                Observacoes
              </Typography>
              <Typography sx={{ whiteSpace: "pre-wrap" }} variant="body2">
                {formatOptionalText(project.notes)}
              </Typography>
            </Box>
          </SectionCard>

          <SectionCard
            description="Lista de orcamentos atualmente associados ao projeto, com acoes para editar, abrir na listagem e remover o vinculo."
            title="Orcamentos vinculados"
          >
            {projectBudgets.length === 0 ? (
              <Alert severity="info" variant="outlined">
                Nenhum orcamento vinculado a este projeto ate o momento.
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
                            Designer
                          </Typography>
                          <Typography variant="body2">
                            {formatOptionalText(budget.designerName)}
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
                          onClick={() => navigate(`/budgets/${budget.id}/edit`)}
                          size="small"
                          startIcon={<EditRoundedIcon />}
                          variant="text"
                        >
                          Editar orcamento
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
                          Remover do projeto
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
                            Historico de status
                          </Typography>
                        </Box>

                        {budgetHistoryQuery?.isLoading ? (
                          <Typography color="text.secondary" variant="body2">
                            Carregando historico...
                          </Typography>
                        ) : budgetHistoryQuery?.isError ? (
                          <Alert severity="warning" variant="outlined">
                            Nao foi possivel carregar o historico deste
                            orcamento.
                          </Alert>
                        ) : budgetHistory.length === 0 ? (
                          <Typography color="text.secondary" variant="body2">
                            Nenhuma alteracao de status registrada ate o
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
                                        por usuario #
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
      >
        <DialogTitle>Associar orcamento ao projeto</DialogTitle>
        <DialogContent
          sx={{ display: "flex", flexDirection: "column", gap: 2, pt: 2 }}
        >
          <DialogContentText>
            Selecione um orcamento sem projeto vinculado para adiciona-lo a este
            projeto.
          </DialogContentText>

          <TextField
            label="Buscar por numero ou responsavel"
            onChange={(event) => setCandidateSearch(event.target.value)}
            value={candidateSearch}
          />

          <TextField
            helperText={
              associationCandidatesQuery.isLoading
                ? "Carregando orcamentos disponiveis..."
                : availableAssociationCandidates.length === 0
                  ? "Nenhum orcamento disponivel para associacao."
                  : "Selecione um orcamento da lista."
            }
            label="Orcamento disponivel"
            onChange={(event) => setAssociationCandidateId(event.target.value)}
            select
            value={associationCandidateId}
          >
            <MenuItem value="">Selecione</MenuItem>
            {availableAssociationCandidates.map((budget) => (
              <MenuItem key={budget.id} value={String(budget.id)}>
                {budget.budgetNumber} ·{" "}
                {formatOptionalText(budget.salespersonName)}
              </MenuItem>
            ))}
          </TextField>
        </DialogContent>
        <DialogActions>
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
        maxWidth="xs"
        onClose={() => {
          if (disassociateBudgetMutation.isPending) {
            return;
          }

          setPendingDisassociationBudget(null);
        }}
        open={pendingDisassociationBudget !== null}
      >
        <DialogTitle>Remover orcamento do projeto</DialogTitle>
        <DialogContent sx={{ pt: 2 }}>
          <DialogContentText>
            {pendingDisassociationBudget
              ? `Confirma a remocao do orcamento ${pendingDisassociationBudget.budgetNumber} deste projeto?`
              : ""}
          </DialogContentText>
        </DialogContent>
        <DialogActions>
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
