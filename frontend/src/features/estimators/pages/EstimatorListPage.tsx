import AddRoundedIcon from "@mui/icons-material/AddRounded";
import BlockRoundedIcon from "@mui/icons-material/BlockRounded";
import DeleteOutlineRoundedIcon from "@mui/icons-material/DeleteOutlineRounded";
import EditRoundedIcon from "@mui/icons-material/EditRounded";
import PublishedWithChangesRoundedIcon from "@mui/icons-material/PublishedWithChangesRounded";
import {
  Alert,
  Box,
  Button,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  LinearProgress,
  MenuItem,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TextField,
  Typography,
} from "@mui/material";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { isAxiosError } from "axios";
import { useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  compactFilterFieldSx,
  FilterField,
  filterGroupSx,
  filterGroupTitleSx,
  filterSectionCardSx,
} from "../../../components/common/FilterField";
import { PageHeader } from "../../../components/common/PageHeader";
import { SectionCard } from "../../../components/common/SectionCard";
import {
  deleteEstimatorRequest,
  listEstimatorsRequest,
  updateEstimatorRequest,
} from "../api/estimators";
import type {
  EstimatorItem,
  EstimatorListFilters,
  UpdateEstimatorPayload,
} from "../types/estimator";

const defaultFilters: EstimatorListFilters = {
  search: "",
  status: "all",
  link: "all",
};

const dateTimeFormatter = new Intl.DateTimeFormat("pt-BR", {
  dateStyle: "short",
  timeStyle: "short",
});

const tableHeadCellSx = {
  background: "linear-gradient(180deg, #1E3A8A 0%, #1D4ED8 100%)",
  borderBottomColor: "#1E40AF",
  borderBottomWidth: 2,
  color: "common.white",
  fontSize: "0.75rem",
  fontWeight: 700,
  letterSpacing: "0.04em",
  py: 1.5,
  textTransform: "uppercase",
  whiteSpace: "nowrap",
};

const tableDetailCellSx = {
  color: "text.secondary",
  fontSize: "0.82rem",
  lineHeight: 1.45,
  py: 1.25,
  verticalAlign: "top",
};

function formatDateTime(value: string) {
  return dateTimeFormatter.format(new Date(value));
}

function getStatusChipColor(active: boolean) {
  return active ? "success" : "default";
}

function getStatusLabel(active: boolean) {
  return active ? "Ativo" : "Inativo";
}

function getMutationErrorMessage(error: unknown, fallbackMessage: string) {
  if (isAxiosError<{ message?: string }>(error)) {
    return error.response?.data?.message ?? fallbackMessage;
  }

  return fallbackMessage;
}

function mapUpdatePayload(
  estimator: EstimatorItem,
  payload: Pick<
    UpdateEstimatorPayload,
    "code" | "email" | "name" | "notes" | "phone" | "userId"
  >,
  active = estimator.active,
): UpdateEstimatorPayload {
  return {
    ...payload,
    active,
  };
}

export default function EstimatorListPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [filters, setFilters] = useState<EstimatorListFilters>(defaultFilters);
  const [feedbackMessage, setFeedbackMessage] = useState<string | null>(null);
  const [feedbackError, setFeedbackError] = useState<string | null>(null);
  const [pendingDelete, setPendingDelete] = useState<EstimatorItem | null>(
    null,
  );

  const estimatorsQuery = useQuery({
    queryKey: ["estimators"],
    queryFn: listEstimatorsRequest,
  });

  const updateMutation = useMutation({
    mutationFn: ({
      estimatorId,
      payload,
    }: {
      estimatorId: number;
      payload: UpdateEstimatorPayload;
    }) => updateEstimatorRequest(estimatorId, payload),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["estimators"] });
      setFeedbackError(null);
      setFeedbackMessage("Orçamentista atualizado com sucesso.");
    },
    onError: (error) => {
      setFeedbackMessage(null);
      setFeedbackError(
        getMutationErrorMessage(
          error,
          "Não foi possível atualizar o orçamentista.",
        ),
      );
    },
  });

  const deleteMutation = useMutation({
    mutationFn: deleteEstimatorRequest,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["estimators"] });
      setFeedbackError(null);
      setFeedbackMessage("Orçamentista removido com sucesso.");
      setPendingDelete(null);
    },
    onError: (error) => {
      setFeedbackMessage(null);
      setFeedbackError(
        getMutationErrorMessage(
          error,
          "Não foi possível remover o orçamentista.",
        ),
      );
    },
  });

  const filteredEstimators = useMemo(() => {
    const normalizedSearch = filters.search.trim().toLocaleLowerCase("pt-BR");

    return (estimatorsQuery.data ?? [])
      .filter((item) => {
        const matchesSearch =
          normalizedSearch.length === 0 ||
          item.code.toLocaleLowerCase("pt-BR").includes(normalizedSearch) ||
          item.name.toLocaleLowerCase("pt-BR").includes(normalizedSearch) ||
          item.email.toLocaleLowerCase("pt-BR").includes(normalizedSearch) ||
          item.phone.toLocaleLowerCase("pt-BR").includes(normalizedSearch) ||
          item.userName
            ?.toLocaleLowerCase("pt-BR")
            .includes(normalizedSearch) === true;
        const matchesStatus =
          filters.status === "all" ||
          (filters.status === "active" && item.active) ||
          (filters.status === "inactive" && !item.active);
        const matchesLink =
          filters.link === "all" ||
          (filters.link === "linked" && item.userId !== null) ||
          (filters.link === "unlinked" && item.userId === null);

        return matchesSearch && matchesStatus && matchesLink;
      })
      .sort((firstItem, secondItem) =>
        firstItem.name.localeCompare(secondItem.name, "pt-BR"),
      );
  }, [estimatorsQuery.data, filters.link, filters.search, filters.status]);

  const isMutating = updateMutation.isPending || deleteMutation.isPending;

  const handleOpenCreate = () => {
    setFeedbackMessage(null);
    setFeedbackError(null);
    navigate("/estimators/new");
  };

  const handleOpenEdit = (estimator: EstimatorItem) => {
    setFeedbackMessage(null);
    setFeedbackError(null);
    navigate(`/estimators/${estimator.id}/edit`);
  };

  const handleToggleActive = async (estimator: EstimatorItem) => {
    await updateMutation.mutateAsync({
      estimatorId: estimator.id,
      payload: mapUpdatePayload(estimator, {
        code: estimator.code,
        name: estimator.name,
        email: estimator.email,
        phone: estimator.phone,
        notes: estimator.notes,
        userId: estimator.userId,
      }, !estimator.active),
    });
    setFeedbackError(null);
    setFeedbackMessage(
      estimator.active
        ? "Orçamentista desativado com sucesso."
        : "Orçamentista ativado com sucesso.",
    );
  };

  const handleDelete = async () => {
    if (!pendingDelete) {
      return;
    }

    await deleteMutation.mutateAsync(pendingDelete.id);
  };

  return (
    <Box sx={{ display: "flex", flexDirection: "column", gap: 3 }}>
      <PageHeader
        action={
          <Button
            onClick={handleOpenCreate}
            startIcon={<AddRoundedIcon />}
            variant="contained"
          >
            Novo orçamentista
          </Button>
        }
        description="Cadastre e gerencie os orçamentistas usados na governança operacional dos orçamentos."
        title="Orçamentistas"
      />

      <SectionCard
        description="Use os filtros para localizar rapidamente um orçamentista pelo código, nome, e-mail ou usuário vinculado."
        sx={filterSectionCardSx}
        title="Filtros"
      >
        <Box
          sx={{
            display: "grid",
            gap: 2,
            gridTemplateColumns: {
              lg: "minmax(0, 1.4fr) minmax(0, 1.6fr)",
              md: "repeat(2, minmax(0, 1fr))",
              xs: "minmax(0, 1fr)",
            },
          }}
        >
          <Box sx={filterGroupSx}>
            <Typography sx={filterGroupTitleSx} variant="subtitle2">
              Identificação
            </Typography>
            <FilterField label="Buscar por código, nome, e-mail ou usuário">
              <TextField
                onChange={(event) =>
                  setFilters((currentFilters) => ({
                    ...currentFilters,
                    search: event.target.value,
                  }))
                }
                size="small"
                sx={compactFilterFieldSx}
                value={filters.search}
              />
            </FilterField>
          </Box>
          <Box sx={filterGroupSx}>
            <Typography sx={filterGroupTitleSx} variant="subtitle2">
              Classificação
            </Typography>
            <Box
              sx={{
                display: "grid",
                gap: 2,
                gridTemplateColumns: {
                  sm: "repeat(2, minmax(0, 1fr))",
                  xs: "minmax(0, 1fr)",
                },
              }}
            >
              <FilterField label="Status">
                <TextField
                  onChange={(event) =>
                    setFilters((currentFilters) => ({
                      ...currentFilters,
                      status: event.target
                        .value as EstimatorListFilters["status"],
                    }))
                  }
                  select
                  size="small"
                  sx={compactFilterFieldSx}
                  value={filters.status}
                >
                  <MenuItem value="all">Todos</MenuItem>
                  <MenuItem value="active">Ativo</MenuItem>
                  <MenuItem value="inactive">Inativo</MenuItem>
                </TextField>
              </FilterField>
              <FilterField label="Vínculo de usuário">
                <TextField
                  onChange={(event) =>
                    setFilters((currentFilters) => ({
                      ...currentFilters,
                      link: event.target.value as EstimatorListFilters["link"],
                    }))
                  }
                  select
                  size="small"
                  sx={compactFilterFieldSx}
                  value={filters.link}
                >
                  <MenuItem value="all">Todos</MenuItem>
                  <MenuItem value="linked">Com usuário</MenuItem>
                  <MenuItem value="unlinked">Sem usuário</MenuItem>
                </TextField>
              </FilterField>
            </Box>
          </Box>
        </Box>
      </SectionCard>

      <SectionCard
        description={`${filteredEstimators.length} orçamentista(s) encontrado(s) com os filtros atuais.`}
        title="Listagem"
      >
        {estimatorsQuery.isLoading ? <LinearProgress /> : null}
        {estimatorsQuery.isError ? (
          <Alert severity="error">
            Não foi possível carregar a listagem de orçamentistas.
          </Alert>
        ) : null}
        {feedbackMessage ? (
          <Alert severity="success">{feedbackMessage}</Alert>
        ) : null}
        {feedbackError ? <Alert severity="error">{feedbackError}</Alert> : null}

        <TableContainer>
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell sx={tableHeadCellSx}>Código</TableCell>
                <TableCell sx={tableHeadCellSx}>Nome</TableCell>
                <TableCell sx={tableHeadCellSx}>Contato</TableCell>
                <TableCell sx={tableHeadCellSx}>Usuário vinculado</TableCell>
                <TableCell sx={tableHeadCellSx}>Status</TableCell>
                <TableCell sx={tableHeadCellSx}>Atualizado em</TableCell>
                <TableCell align="right" sx={tableHeadCellSx}>
                  Ações
                </TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {filteredEstimators.map((item) => (
                <TableRow hover key={item.id}>
                  <TableCell sx={tableDetailCellSx}>
                    <Typography sx={{ fontWeight: 700 }} variant="body2">
                      {item.code}
                    </Typography>
                  </TableCell>
                  <TableCell sx={tableDetailCellSx}>
                    <Box
                      sx={{
                        display: "flex",
                        flexDirection: "column",
                        gap: 0.25,
                      }}
                    >
                      <Typography sx={{ fontWeight: 600 }} variant="body2">
                        {item.name}
                      </Typography>
                      {item.notes.trim() ? (
                        <Typography color="text.secondary" variant="caption">
                          {item.notes}
                        </Typography>
                      ) : null}
                    </Box>
                  </TableCell>
                  <TableCell sx={tableDetailCellSx}>
                    <Box
                      sx={{
                        display: "flex",
                        flexDirection: "column",
                        gap: 0.25,
                      }}
                    >
                      <Typography variant="body2">
                        {item.email || "E-mail não informado"}
                      </Typography>
                      <Typography color="text.secondary" variant="caption">
                        {item.phone || "Telefone não informado"}
                      </Typography>
                    </Box>
                  </TableCell>
                  <TableCell sx={tableDetailCellSx}>
                    {item.userName ? (
                      <Chip
                        color="primary"
                        label={item.userName}
                        size="small"
                        variant="outlined"
                      />
                    ) : (
                      "Não vinculado"
                    )}
                  </TableCell>
                  <TableCell sx={tableDetailCellSx}>
                    <Chip
                      color={getStatusChipColor(item.active)}
                      label={getStatusLabel(item.active)}
                      size="small"
                      variant={item.active ? "filled" : "outlined"}
                    />
                  </TableCell>
                  <TableCell sx={tableDetailCellSx}>
                    {formatDateTime(item.updatedAt)}
                  </TableCell>
                  <TableCell align="right" sx={tableDetailCellSx}>
                    <Box
                      sx={{
                        display: "flex",
                        flexWrap: "wrap",
                        gap: 1,
                        justifyContent: "flex-end",
                      }}
                    >
                      <Button
                        disabled={isMutating}
                        onClick={() => handleOpenEdit(item)}
                        size="small"
                        startIcon={<EditRoundedIcon />}
                        variant="text"
                      >
                        Editar
                      </Button>
                      <Button
                        color={item.active ? "warning" : "success"}
                        disabled={isMutating}
                        onClick={() => {
                          void handleToggleActive(item);
                        }}
                        size="small"
                        startIcon={
                          item.active ? (
                            <BlockRoundedIcon />
                          ) : (
                            <PublishedWithChangesRoundedIcon />
                          )
                        }
                        variant="text"
                      >
                        {item.active ? "Desativar" : "Ativar"}
                      </Button>
                      <Button
                        color="error"
                        disabled={isMutating}
                        onClick={() => setPendingDelete(item)}
                        size="small"
                        startIcon={<DeleteOutlineRoundedIcon />}
                        variant="text"
                      >
                        Excluir
                      </Button>
                    </Box>
                  </TableCell>
                </TableRow>
              ))}
              {!estimatorsQuery.isLoading && filteredEstimators.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={7} sx={tableDetailCellSx}>
                    Nenhum orçamentista encontrado com os filtros informados.
                  </TableCell>
                </TableRow>
              ) : null}
            </TableBody>
          </Table>
        </TableContainer>
      </SectionCard>

      <Dialog
        onClose={() => {
          if (!deleteMutation.isPending) {
            setPendingDelete(null);
          }
        }}
        open={pendingDelete !== null}
      >
        <DialogTitle>Confirmar exclusão</DialogTitle>
        <DialogContent>
          <DialogContentText>
            {pendingDelete
              ? `Confirma a exclusão do orçamentista ${pendingDelete.name}?`
              : ""}
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button
            disabled={deleteMutation.isPending}
            onClick={() => setPendingDelete(null)}
            variant="outlined"
          >
            Cancelar
          </Button>
          <Button
            color="error"
            disabled={deleteMutation.isPending}
            onClick={() => {
              void handleDelete();
            }}
            variant="contained"
          >
            {deleteMutation.isPending ? "Excluindo..." : "Excluir"}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
