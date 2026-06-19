import AddRoundedIcon from "@mui/icons-material/AddRounded";
import BlockRoundedIcon from "@mui/icons-material/BlockRounded";
import DeleteOutlineRoundedIcon from "@mui/icons-material/DeleteOutlineRounded";
import EditRoundedIcon from "@mui/icons-material/EditRounded";
import PublishedWithChangesRoundedIcon from "@mui/icons-material/PublishedWithChangesRounded";
import { zodResolver } from "@hookform/resolvers/zod";
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
import { useEffect, useMemo, useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import {
  compactFilterFieldSx,
  FilterField,
  filterGroupSx,
  filterGroupTitleSx,
  filterSectionCardSx,
} from "../../../components/common/FilterField";
import { PageHeader } from "../../../components/common/PageHeader";
import { SectionCard } from "../../../components/common/SectionCard";
import { listUsersRequest } from "../../users/api/users";
import type { UserItem } from "../../users/types/user";
import {
  createEstimatorRequest,
  deleteEstimatorRequest,
  getNextEstimatorCodeRequest,
  listEstimatorsRequest,
  updateEstimatorRequest,
} from "../api/estimators";
import type {
  CreateEstimatorPayload,
  EstimatorItem,
  EstimatorListFilters,
  UpdateEstimatorPayload,
} from "../types/estimator";

const defaultFilters: EstimatorListFilters = {
  search: "",
  status: "all",
  link: "all",
};

const estimatorSchema = z.object({
  code: z
    .string()
    .trim()
    .min(1, "Informe o código")
    .max(30, "O código deve ter no máximo 30 caracteres"),
  name: z
    .string()
    .trim()
    .min(3, "Informe um nome com pelo menos 3 caracteres")
    .max(150, "O nome deve ter no máximo 150 caracteres"),
  email: z
    .string()
    .trim()
    .refine(
      (value) => value === "" || z.string().email().safeParse(value).success,
      "Informe um e-mail válido",
    ),
  phone: z
    .string()
    .trim()
    .max(50, "O telefone deve ter no máximo 50 caracteres"),
  notes: z.string().trim(),
  userId: z.string().trim(),
});

type EstimatorFormValues = z.infer<typeof estimatorSchema>;
type EstimatorDialogMode = "create" | "edit";

type EstimatorDialogState = {
  mode: EstimatorDialogMode;
  estimator: EstimatorItem | null;
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

function mapCreatePayload(values: EstimatorFormValues): CreateEstimatorPayload {
  return {
    code: values.code.trim(),
    name: values.name.trim(),
    email: values.email.trim(),
    phone: values.phone.trim(),
    notes: values.notes.trim(),
    userId: values.userId ? Number(values.userId) : null,
  };
}

function mapUpdatePayload(
  estimator: EstimatorItem,
  values: EstimatorFormValues,
): UpdateEstimatorPayload {
  return {
    ...mapCreatePayload(values),
    active: estimator.active,
  };
}

function getDialogTitle(dialogState: EstimatorDialogState | null) {
  return dialogState?.mode === "edit"
    ? "Editar orçamentista"
    : "Novo orçamentista";
}

function getDialogSubmitLabel(dialogState: EstimatorDialogState | null) {
  return dialogState?.mode === "edit"
    ? "Salvar alterações"
    : "Cadastrar orçamentista";
}

function getLinkedEstimatorUsers(users: UserItem[]) {
  return users
    .filter(
      (item) =>
        item.role === "user" && item.userKind === "estimator" && item.active,
    )
    .sort((firstItem, secondItem) =>
      firstItem.name.localeCompare(secondItem.name, "pt-BR"),
    );
}

export default function EstimatorListPage() {
  const queryClient = useQueryClient();
  const [filters, setFilters] = useState<EstimatorListFilters>(defaultFilters);
  const [feedbackMessage, setFeedbackMessage] = useState<string | null>(null);
  const [feedbackError, setFeedbackError] = useState<string | null>(null);
  const [dialogState, setDialogState] = useState<EstimatorDialogState | null>(
    null,
  );
  const [pendingDelete, setPendingDelete] = useState<EstimatorItem | null>(
    null,
  );
  const {
    formState: { errors, isSubmitting },
    handleSubmit,
    register,
    reset,
    setValue,
  } = useForm<EstimatorFormValues>({
    defaultValues: {
      code: "",
      name: "",
      email: "",
      phone: "",
      notes: "",
      userId: "",
    },
    resolver: zodResolver(estimatorSchema),
  });

  const estimatorsQuery = useQuery({
    queryKey: ["estimators"],
    queryFn: listEstimatorsRequest,
  });

  const usersQuery = useQuery({
    queryKey: ["users", "estimator-links"],
    queryFn: listUsersRequest,
    staleTime: 1000 * 60 * 5,
  });

  const nextCodeQuery = useQuery({
    queryKey: ["estimators", "next-code", dialogState?.mode],
    queryFn: getNextEstimatorCodeRequest,
    enabled: dialogState?.mode === "create",
    staleTime: 0,
  });

  const createMutation = useMutation({
    mutationFn: createEstimatorRequest,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["estimators"] });
      setFeedbackError(null);
      setFeedbackMessage("Orçamentista cadastrado com sucesso.");
      setDialogState(null);
      reset();
    },
    onError: (error) => {
      setFeedbackMessage(null);
      setFeedbackError(
        getMutationErrorMessage(
          error,
          "Não foi possível cadastrar o orçamentista.",
        ),
      );
    },
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
      setDialogState(null);
      reset();
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

  const linkedUsers = useMemo(
    () => getLinkedEstimatorUsers(usersQuery.data ?? []),
    [usersQuery.data],
  );

  useEffect(() => {
    if (!dialogState) {
      reset({
        code: "",
        name: "",
        email: "",
        phone: "",
        notes: "",
        userId: "",
      });
      return;
    }

    if (dialogState.estimator) {
      reset({
        code: dialogState.estimator.code,
        name: dialogState.estimator.name,
        email: dialogState.estimator.email,
        phone: dialogState.estimator.phone,
        notes: dialogState.estimator.notes,
        userId:
          dialogState.estimator.userId === null
            ? ""
            : String(dialogState.estimator.userId),
      });
      return;
    }

    reset({
      code: nextCodeQuery.data ?? "",
      name: "",
      email: "",
      phone: "",
      notes: "",
      userId: "",
    });
  }, [dialogState, nextCodeQuery.data, reset]);

  useEffect(() => {
    if (dialogState?.mode !== "create" || !nextCodeQuery.data) {
      return;
    }

    setValue("code", nextCodeQuery.data, { shouldValidate: false });
  }, [dialogState?.mode, nextCodeQuery.data, setValue]);

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

  const isMutating =
    createMutation.isPending ||
    updateMutation.isPending ||
    deleteMutation.isPending;

  const handleOpenCreate = () => {
    setFeedbackMessage(null);
    setFeedbackError(null);
    setDialogState({
      mode: "create",
      estimator: null,
    });
  };

  const handleOpenEdit = (estimator: EstimatorItem) => {
    setFeedbackMessage(null);
    setFeedbackError(null);
    setDialogState({
      mode: "edit",
      estimator,
    });
  };

  const handleToggleActive = async (estimator: EstimatorItem) => {
    await updateMutation.mutateAsync({
      estimatorId: estimator.id,
      payload: {
        code: estimator.code,
        name: estimator.name,
        email: estimator.email,
        phone: estimator.phone,
        active: !estimator.active,
        notes: estimator.notes,
        userId: estimator.userId,
      },
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

  const onSubmit = async (values: EstimatorFormValues) => {
    if (dialogState?.mode === "edit" && dialogState.estimator) {
      await updateMutation.mutateAsync({
        estimatorId: dialogState.estimator.id,
        payload: mapUpdatePayload(dialogState.estimator, values),
      });
      return;
    }

    await createMutation.mutateAsync(mapCreatePayload(values));
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
                      status: event.target.value as EstimatorListFilters["status"],
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
        fullWidth
        maxWidth="sm"
        onClose={() => {
          if (!isSubmitting && !isMutating) {
            setDialogState(null);
          }
        }}
        open={dialogState !== null}
      >
        <Box
          component="form"
          onSubmit={handleSubmit((values) => {
            void onSubmit(values);
          })}
        >
          <DialogTitle>{getDialogTitle(dialogState)}</DialogTitle>
          <DialogContent
            sx={{ display: "flex", flexDirection: "column", gap: 2, pt: 1 }}
          >
            <TextField
              autoFocus
              error={Boolean(errors.code)}
              helperText={
                errors.code?.message ??
                (nextCodeQuery.isLoading && dialogState?.mode === "create"
                  ? "Gerando próximo código..."
                  : undefined)
              }
              label="Código"
              {...register("code")}
            />
            <TextField
              error={Boolean(errors.name)}
              helperText={errors.name?.message}
              label="Nome"
              {...register("name")}
            />
            <TextField
              error={Boolean(errors.email)}
              helperText={errors.email?.message}
              label="E-mail"
              type="email"
              {...register("email")}
            />
            <TextField
              error={Boolean(errors.phone)}
              helperText={errors.phone?.message}
              label="Telefone"
              {...register("phone")}
            />
            <TextField
              error={Boolean(errors.userId)}
              helperText={
                errors.userId?.message ??
                "Vincule opcionalmente um usuário com role user e user_kind estimator."
              }
              label="Usuário vinculado"
              select
              {...register("userId")}
            >
              <MenuItem value="">Não vincular</MenuItem>
              {linkedUsers.map((user) => (
                <MenuItem key={user.id} value={String(user.id)}>
                  {user.name} ({user.username})
                </MenuItem>
              ))}
            </TextField>
            <TextField
              error={Boolean(errors.notes)}
              helperText={errors.notes?.message}
              label="Observações"
              minRows={3}
              multiline
              {...register("notes")}
            />
          </DialogContent>
          <DialogActions>
            <Button
              disabled={isSubmitting || isMutating}
              onClick={() => setDialogState(null)}
              variant="outlined"
            >
              Cancelar
            </Button>
            <Button
              disabled={isSubmitting || isMutating}
              type="submit"
              variant="contained"
            >
              {isSubmitting || isMutating
                ? "Salvando..."
                : getDialogSubmitLabel(dialogState)}
            </Button>
          </DialogActions>
        </Box>
      </Dialog>

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
