import AddRoundedIcon from "@mui/icons-material/AddRounded";
import DeleteOutlineRoundedIcon from "@mui/icons-material/DeleteOutlineRounded";
import EditRoundedIcon from "@mui/icons-material/EditRounded";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  Alert,
  Box,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  LinearProgress,
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
import {
  createSystemTypeRequest,
  deleteSystemTypeRequest,
  listSystemTypesRequest,
  updateSystemTypeRequest,
} from "../api/systemTypes";
import type {
  CreateSystemTypePayload,
  SystemTypeItem,
  SystemTypeListFilters,
  UpdateSystemTypePayload,
} from "../types/systemType";

const defaultFilters: SystemTypeListFilters = {
  search: "",
};

const systemTypeSchema = z.object({
  code: z
    .string()
    .trim()
    .min(2, "Informe um código com pelo menos 2 caracteres")
    .max(50, "O código deve ter no máximo 50 caracteres"),
  name: z
    .string()
    .trim()
    .min(2, "Informe um nome com pelo menos 2 caracteres")
    .max(150, "O nome deve ter no máximo 150 caracteres"),
  description: z
    .string()
    .trim()
    .max(500, "A descrição deve ter no máximo 500 caracteres"),
});

type SystemTypeFormValues = z.infer<typeof systemTypeSchema>;

type SystemTypeDialogMode = "create" | "edit";

type SystemTypeDialogState = {
  mode: SystemTypeDialogMode;
  systemType: SystemTypeItem | null;
};

const dateTimeFormatter = new Intl.DateTimeFormat("pt-BR", {
  dateStyle: "short",
  timeStyle: "short",
});

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
  fontSize: "0.82rem",
  lineHeight: 1.45,
  py: 1.25,
  verticalAlign: "top",
};

function formatDateTime(value: string) {
  return dateTimeFormatter.format(new Date(value));
}

function getMutationErrorMessage(error: unknown, fallbackMessage: string) {
  if (isAxiosError<{ message?: string }>(error)) {
    return error.response?.data?.message ?? fallbackMessage;
  }

  return fallbackMessage;
}

function mapCreatePayload(
  values: SystemTypeFormValues,
): CreateSystemTypePayload {
  return {
    code: values.code.trim(),
    name: values.name.trim(),
    description: values.description.trim(),
  };
}

function mapUpdatePayload(
  values: SystemTypeFormValues,
): UpdateSystemTypePayload {
  return {
    code: values.code.trim(),
    name: values.name.trim(),
    description: values.description.trim(),
  };
}

function getDialogTitle(dialogState: SystemTypeDialogState | null) {
  return dialogState?.mode === "edit"
    ? "Editar tipo de sistema"
    : "Novo tipo de sistema";
}

function getDialogSubmitLabel(dialogState: SystemTypeDialogState | null) {
  return dialogState?.mode === "edit"
    ? "Salvar alteracoes"
    : "Cadastrar tipo de sistema";
}

export default function SystemTypeListPage() {
  const queryClient = useQueryClient();
  const [filters, setFilters] = useState<SystemTypeListFilters>(defaultFilters);
  const [feedbackMessage, setFeedbackMessage] = useState<string | null>(null);
  const [feedbackError, setFeedbackError] = useState<string | null>(null);
  const [deleteErrorMessage, setDeleteErrorMessage] = useState<string | null>(
    null,
  );
  const [dialogState, setDialogState] = useState<SystemTypeDialogState | null>(
    null,
  );
  const [pendingDelete, setPendingDelete] = useState<SystemTypeItem | null>(
    null,
  );
  const {
    formState: { errors, isSubmitting },
    handleSubmit,
    register,
    reset,
  } = useForm<SystemTypeFormValues>({
    defaultValues: {
      code: "",
      description: "",
      name: "",
    },
    resolver: zodResolver(systemTypeSchema),
  });

  const systemTypesQuery = useQuery({
    queryKey: ["system-types"],
    queryFn: listSystemTypesRequest,
  });

  const invalidateRelatedQueries = async () => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: ["system-types"] }),
      queryClient.invalidateQueries({ queryKey: ["budget-catalogs"] }),
      queryClient.invalidateQueries({ queryKey: ["budget-list-catalogs"] }),
    ]);
  };

  const createMutation = useMutation({
    mutationFn: createSystemTypeRequest,
    onSuccess: async () => {
      await invalidateRelatedQueries();
      setFeedbackError(null);
      setFeedbackMessage("Tipo de sistema cadastrado com sucesso.");
      setDialogState(null);
      reset();
    },
    onError: (error) => {
      setFeedbackMessage(null);
      setFeedbackError(
        getMutationErrorMessage(
          error,
          "Não foi possível cadastrar o tipo de sistema.",
        ),
      );
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({
      payload,
      systemTypeId,
    }: {
      payload: UpdateSystemTypePayload;
      systemTypeId: number;
    }) => updateSystemTypeRequest(systemTypeId, payload),
    onSuccess: async () => {
      await invalidateRelatedQueries();
      setFeedbackError(null);
      setFeedbackMessage("Tipo de sistema atualizado com sucesso.");
      setDialogState(null);
      reset();
    },
    onError: (error) => {
      setFeedbackMessage(null);
      setFeedbackError(
        getMutationErrorMessage(
          error,
          "Não foi possível atualizar o tipo de sistema.",
        ),
      );
    },
  });

  const deleteMutation = useMutation({
    mutationFn: deleteSystemTypeRequest,
    onSuccess: async () => {
      await invalidateRelatedQueries();
      setDeleteErrorMessage(null);
      setFeedbackError(null);
      setFeedbackMessage("Tipo de sistema removido com sucesso.");
      setPendingDelete(null);
    },
    onError: (error) => {
      setDeleteErrorMessage(
        getMutationErrorMessage(
          error,
          "Não foi possível remover o tipo de sistema.",
        ),
      );
      setFeedbackMessage(null);
      setFeedbackError(
        getMutationErrorMessage(
          error,
          "Não foi possível remover o tipo de sistema.",
        ),
      );
    },
  });

  useEffect(() => {
    if (!dialogState) {
      reset({
        code: "",
        description: "",
        name: "",
      });
      return;
    }

    if (dialogState.systemType) {
      reset({
        code: dialogState.systemType.code,
        description: dialogState.systemType.description,
        name: dialogState.systemType.name,
      });
      return;
    }

    reset({
      code: "",
      description: "",
      name: "",
    });
  }, [dialogState, reset]);

  const filteredSystemTypes = useMemo(() => {
    const normalizedSearch = filters.search.trim().toLocaleLowerCase("pt-BR");

    return (systemTypesQuery.data ?? [])
      .filter((item) => {
        if (normalizedSearch.length === 0) {
          return true;
        }

        return (
          item.code.toLocaleLowerCase("pt-BR").includes(normalizedSearch) ||
          item.name.toLocaleLowerCase("pt-BR").includes(normalizedSearch) ||
          item.description.toLocaleLowerCase("pt-BR").includes(normalizedSearch)
        );
      })
      .sort((firstItem, secondItem) =>
        firstItem.name.localeCompare(secondItem.name, "pt-BR"),
      );
  }, [filters.search, systemTypesQuery.data]);

  const isMutating =
    createMutation.isPending ||
    updateMutation.isPending ||
    deleteMutation.isPending;

  const handleOpenCreate = () => {
    setFeedbackMessage(null);
    setFeedbackError(null);
    setDialogState({
      mode: "create",
      systemType: null,
    });
  };

  const handleOpenEdit = (systemType: SystemTypeItem) => {
    setFeedbackMessage(null);
    setFeedbackError(null);
    setDeleteErrorMessage(null);
    setDialogState({
      mode: "edit",
      systemType,
    });
  };

  const handleDelete = async () => {
    if (!pendingDelete) {
      return;
    }

    await deleteMutation.mutateAsync(pendingDelete.id);
  };

  const onSubmit = async (values: SystemTypeFormValues) => {
    if (dialogState?.mode === "edit" && dialogState.systemType) {
      await updateMutation.mutateAsync({
        payload: mapUpdatePayload(values),
        systemTypeId: dialogState.systemType.id,
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
            Novo tipo de sistema
          </Button>
        }
        description="Cadastre e gerencie os tipos de sistema usados nos orçamentos."
        title="Tipos de Sistema"
      />

      <SectionCard
        description="Use a busca para localizar rapidamente um tipo de sistema por código, nome ou descrição."
        sx={filterSectionCardSx}
        title="Filtros"
      >
        <Box sx={filterGroupSx}>
          <Typography sx={filterGroupTitleSx} variant="subtitle2">
            Identificação
          </Typography>
          <FilterField label="Buscar por código, nome ou descrição">
            <TextField
              fullWidth
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
      </SectionCard>

      <SectionCard
        description={`${filteredSystemTypes.length} tipo(s) de sistema encontrado(s) com os filtros atuais.`}
        title="Listagem"
      >
        {systemTypesQuery.isLoading ? <LinearProgress /> : null}

        {feedbackMessage ? (
          <Alert severity="success">{feedbackMessage}</Alert>
        ) : null}
        {feedbackError ? <Alert severity="error">{feedbackError}</Alert> : null}

        {systemTypesQuery.isError ? (
          <Alert severity="error">
            {getMutationErrorMessage(
              systemTypesQuery.error,
              "Não foi possível carregar os tipos de sistema.",
            )}
          </Alert>
        ) : null}

        {!systemTypesQuery.isLoading && !systemTypesQuery.isError ? (
          filteredSystemTypes.length > 0 ? (
            <TableContainer>
              <Table size="small">
                <TableHead>
                  <TableRow>
                    <TableCell sx={tableHeadCellSx}>Código</TableCell>
                    <TableCell sx={tableHeadCellSx}>Nome</TableCell>
                    <TableCell sx={tableHeadCellSx}>Descrição</TableCell>
                    <TableCell sx={tableHeadCellSx}>Criado em</TableCell>
                    <TableCell sx={tableHeadCellSx}>Atualizado em</TableCell>
                    <TableCell sx={tableHeadCellSx}>Acoes</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {filteredSystemTypes.map((systemType) => (
                    <TableRow key={systemType.id}>
                      <TableCell sx={tableDetailCellSx}>
                        {systemType.code}
                      </TableCell>
                      <TableCell sx={tableDetailCellSx}>
                        {systemType.name}
                      </TableCell>
                      <TableCell sx={tableDetailCellSx}>
                        {systemType.description || "Sem descrição"}
                      </TableCell>
                      <TableCell sx={tableDetailCellSx}>
                        {formatDateTime(systemType.createdAt)}
                      </TableCell>
                      <TableCell sx={tableDetailCellSx}>
                        {formatDateTime(systemType.updatedAt)}
                      </TableCell>
                      <TableCell
                        sx={{
                          ...tableDetailCellSx,
                          whiteSpace: "nowrap",
                        }}
                      >
                        <Button
                          onClick={() => handleOpenEdit(systemType)}
                          size="small"
                          startIcon={<EditRoundedIcon />}
                          sx={{ minWidth: "auto", px: 0.75, py: 0.25 }}
                          variant="text"
                        >
                          Editar
                        </Button>
                        <Button
                          color="error"
                          onClick={() => {
                            setDeleteErrorMessage(null);
                            setPendingDelete(systemType);
                          }}
                          size="small"
                          startIcon={<DeleteOutlineRoundedIcon />}
                          sx={{ minWidth: "auto", px: 0.75, py: 0.25 }}
                          variant="text"
                        >
                          Excluir
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          ) : (
            <Typography color="text.secondary" variant="body2">
              Nenhum tipo de sistema encontrado.
            </Typography>
          )
        ) : null}
      </SectionCard>

      <Dialog
        fullWidth
        maxWidth="sm"
        onClose={() => {
          if (!isMutating && !isSubmitting) {
            setDialogState(null);
          }
        }}
        open={dialogState !== null}
      >
        <Box component="form" noValidate onSubmit={handleSubmit(onSubmit)}>
          <DialogTitle>{getDialogTitle(dialogState)}</DialogTitle>
          <DialogContent
            sx={{
              display: "grid",
              gap: 2,
              pt: "8px !important",
            }}
          >
            <TextField
              error={Boolean(errors.code)}
              helperText={errors.code?.message}
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
              error={Boolean(errors.description)}
              helperText={errors.description?.message}
              label="Descrição"
              minRows={3}
              multiline
              {...register("description")}
            />
          </DialogContent>
          <DialogActions>
            <Button
              disabled={isMutating || isSubmitting}
              onClick={() => setDialogState(null)}
              variant="outlined"
            >
              Cancelar
            </Button>
            <Button
              disabled={isMutating || isSubmitting}
              type="submit"
              variant="contained"
            >
              {getDialogSubmitLabel(dialogState)}
            </Button>
          </DialogActions>
        </Box>
      </Dialog>

      <Dialog
        onClose={() => {
          if (!deleteMutation.isPending) {
            setDeleteErrorMessage(null);
            setPendingDelete(null);
          }
        }}
        open={pendingDelete !== null}
      >
        <DialogTitle>Excluir tipo de sistema</DialogTitle>
        <DialogContent>
          <Alert severity="warning" sx={{ mb: 2 }}>
            Se este tipo de sistema estiver vinculado a orcamentos, a exclusao
            vai remover essa classificacao desses registros e eles ficarao sem
            tipo de sistema.
          </Alert>
          {deleteErrorMessage ? (
            <Alert severity="error" sx={{ mb: 2 }}>
              {deleteErrorMessage}
            </Alert>
          ) : null}
          <DialogContentText>
            Confirma a exclusao do tipo de sistema{" "}
            <strong>{pendingDelete?.name ?? ""}</strong>?
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
            onClick={handleDelete}
            variant="contained"
          >
            {deleteMutation.isPending ? "Excluindo..." : "Excluir"}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
