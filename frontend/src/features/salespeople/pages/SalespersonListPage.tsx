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
import {
  createSalespersonRequest,
  deleteSalespersonRequest,
  listSalespeopleRequest,
  updateSalespersonRequest,
} from "../api/salespeople";
import type {
  CreateSalespersonPayload,
  SalespersonItem,
  SalespersonListFilters,
  UpdateSalespersonPayload,
} from "../types/salesperson";

const defaultFilters: SalespersonListFilters = {
  search: "",
  status: "all",
};

const salespersonSchema = z.object({
  name: z
    .string()
    .trim()
    .min(3, "Informe um nome com pelo menos 3 caracteres")
    .max(150, "O nome deve ter no máximo 150 caracteres"),
  email: z.string().trim().email("Informe um e-mail válido"),
  phone: z
    .string()
    .trim()
    .min(8, "Informe um telefone com pelo menos 8 caracteres")
    .max(30, "O telefone deve ter no máximo 30 caracteres"),
});

type SalespersonFormValues = z.infer<typeof salespersonSchema>;

type SalespersonDialogMode = "create" | "edit";

type SalespersonDialogState = {
  mode: SalespersonDialogMode;
  salesperson: SalespersonItem | null;
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

function mapCreatePayload(
  values: SalespersonFormValues,
): CreateSalespersonPayload {
  return {
    email: values.email.trim(),
    name: values.name.trim(),
    phone: values.phone.trim(),
  };
}

function mapUpdatePayload(
  salesperson: SalespersonItem,
  values: SalespersonFormValues,
): UpdateSalespersonPayload {
  return {
    active: salesperson.active,
    email: values.email.trim(),
    name: values.name.trim(),
    phone: values.phone.trim(),
  };
}

function getDialogTitle(dialogState: SalespersonDialogState | null) {
  return dialogState?.mode === "edit" ? "Editar vendedor" : "Novo vendedor";
}

function getDialogSubmitLabel(dialogState: SalespersonDialogState | null) {
  return dialogState?.mode === "edit"
    ? "Salvar alterações"
    : "Cadastrar vendedor";
}

export default function SalespersonListPage() {
  const queryClient = useQueryClient();
  const [filters, setFilters] =
    useState<SalespersonListFilters>(defaultFilters);
  const [feedbackMessage, setFeedbackMessage] = useState<string | null>(null);
  const [feedbackError, setFeedbackError] = useState<string | null>(null);
  const [dialogState, setDialogState] = useState<SalespersonDialogState | null>(
    null,
  );
  const [pendingDelete, setPendingDelete] = useState<SalespersonItem | null>(
    null,
  );
  const {
    formState: { errors, isSubmitting },
    handleSubmit,
    register,
    reset,
  } = useForm<SalespersonFormValues>({
    defaultValues: {
      email: "",
      name: "",
      phone: "",
    },
    resolver: zodResolver(salespersonSchema),
  });

  const salespeopleQuery = useQuery({
    queryKey: ["salespeople"],
    queryFn: listSalespeopleRequest,
  });

  const createMutation = useMutation({
    mutationFn: createSalespersonRequest,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["salespeople"] });
      setFeedbackError(null);
      setFeedbackMessage("Vendedor cadastrado com sucesso.");
      setDialogState(null);
      reset();
    },
    onError: (error) => {
      setFeedbackMessage(null);
      setFeedbackError(
        getMutationErrorMessage(
          error,
          "Não foi possível cadastrar o vendedor.",
        ),
      );
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({
      payload,
      salespersonId,
    }: {
      payload: UpdateSalespersonPayload;
      salespersonId: number;
    }) => updateSalespersonRequest(salespersonId, payload),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["salespeople"] });
      setFeedbackError(null);
      setFeedbackMessage("Vendedor atualizado com sucesso.");
      setDialogState(null);
      reset();
    },
    onError: (error) => {
      setFeedbackMessage(null);
      setFeedbackError(
        getMutationErrorMessage(
          error,
          "Não foi possível atualizar o vendedor.",
        ),
      );
    },
  });

  const deleteMutation = useMutation({
    mutationFn: deleteSalespersonRequest,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["salespeople"] });
      setFeedbackError(null);
      setFeedbackMessage("Vendedor removido com sucesso.");
      setPendingDelete(null);
    },
    onError: (error) => {
      setFeedbackMessage(null);
      setFeedbackError(
        getMutationErrorMessage(error, "Não foi possível remover o vendedor."),
      );
    },
  });

  useEffect(() => {
    if (!dialogState) {
      reset({
        email: "",
        name: "",
        phone: "",
      });
      return;
    }

    if (dialogState.salesperson) {
      reset({
        email: dialogState.salesperson.email,
        name: dialogState.salesperson.name,
        phone: dialogState.salesperson.phone,
      });
      return;
    }

    reset({
      email: "",
      name: "",
      phone: "",
    });
  }, [dialogState, reset]);

  const filteredSalespeople = useMemo(() => {
    const normalizedSearch = filters.search.trim().toLocaleLowerCase("pt-BR");

    return (salespeopleQuery.data ?? [])
      .filter((item) => {
        const matchesSearch =
          normalizedSearch.length === 0 ||
          item.name.toLocaleLowerCase("pt-BR").includes(normalizedSearch) ||
          item.email.toLocaleLowerCase("pt-BR").includes(normalizedSearch) ||
          item.phone.toLocaleLowerCase("pt-BR").includes(normalizedSearch);
        const matchesStatus =
          filters.status === "all" ||
          (filters.status === "active" && item.active) ||
          (filters.status === "inactive" && !item.active);

        return matchesSearch && matchesStatus;
      })
      .sort((firstItem, secondItem) =>
        firstItem.name.localeCompare(secondItem.name, "pt-BR"),
      );
  }, [filters.search, filters.status, salespeopleQuery.data]);

  const isMutating =
    createMutation.isPending ||
    updateMutation.isPending ||
    deleteMutation.isPending;

  const handleOpenCreate = () => {
    setFeedbackMessage(null);
    setFeedbackError(null);
    setDialogState({
      mode: "create",
      salesperson: null,
    });
  };

  const handleOpenEdit = (salesperson: SalespersonItem) => {
    setFeedbackMessage(null);
    setFeedbackError(null);
    setDialogState({
      mode: "edit",
      salesperson,
    });
  };

  const handleToggleActive = async (salesperson: SalespersonItem) => {
    await updateMutation.mutateAsync({
      payload: {
        active: !salesperson.active,
        email: salesperson.email,
        name: salesperson.name,
        phone: salesperson.phone,
      },
      salespersonId: salesperson.id,
    });
    setFeedbackError(null);
    setFeedbackMessage(
      salesperson.active
        ? "Vendedor desativado com sucesso."
        : "Vendedor ativado com sucesso.",
    );
  };

  const handleDelete = async () => {
    if (!pendingDelete) {
      return;
    }

    await deleteMutation.mutateAsync(pendingDelete.id);
  };

  const onSubmit = async (values: SalespersonFormValues) => {
    if (dialogState?.mode === "edit" && dialogState.salesperson) {
      await updateMutation.mutateAsync({
        payload: mapUpdatePayload(dialogState.salesperson, values),
        salespersonId: dialogState.salesperson.id,
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
            Novo vendedor
          </Button>
        }
        description="Cadastre e gerencie os vendedores usados nos orçamentos do sistema."
        title="Vendedores"
      />

      <SectionCard
        description="Use os filtros para localizar rapidamente um vendedor pelo nome, e-mail ou telefone."
        sx={filterSectionCardSx}
        title="Filtros"
      >
        <Box
          sx={{
            display: "grid",
            gap: 2,
            gridTemplateColumns: {
              md: "minmax(0, 1.6fr) minmax(0, 1fr)",
              xs: "minmax(0, 1fr)",
            },
          }}
        >
          <Box sx={filterGroupSx}>
            <Typography sx={filterGroupTitleSx} variant="subtitle2">
              Identificação
            </Typography>
            <FilterField label="Buscar por nome, e-mail ou telefone">
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
            <FilterField label="Status">
              <TextField
                onChange={(event) =>
                  setFilters((currentFilters) => ({
                    ...currentFilters,
                    status: event.target
                      .value as SalespersonListFilters["status"],
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
          </Box>
        </Box>
      </SectionCard>

      <SectionCard
        description={`${filteredSalespeople.length} vendedor(es) encontrado(s) com os filtros atuais.`}
        title="Listagem"
      >
        {salespeopleQuery.isLoading ? <LinearProgress /> : null}
        {salespeopleQuery.isError ? (
          <Alert severity="error">
            Não foi possível carregar a listagem de vendedores.
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
                <TableCell sx={tableHeadCellSx}>Nome</TableCell>
                <TableCell sx={tableHeadCellSx}>E-mail</TableCell>
                <TableCell sx={tableHeadCellSx}>Telefone</TableCell>
                <TableCell sx={tableHeadCellSx}>Status</TableCell>
                <TableCell sx={tableHeadCellSx}>Atualizado em</TableCell>
                <TableCell align="right" sx={tableHeadCellSx}>
                  Ações
                </TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {filteredSalespeople.map((item) => (
                <TableRow hover key={item.id}>
                  <TableCell sx={tableDetailCellSx}>
                    <Typography sx={{ fontWeight: 600 }} variant="body2">
                      {item.name}
                    </Typography>
                  </TableCell>
                  <TableCell sx={tableDetailCellSx}>{item.email}</TableCell>
                  <TableCell sx={tableDetailCellSx}>{item.phone}</TableCell>
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
              {!salespeopleQuery.isLoading &&
              filteredSalespeople.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={6} sx={tableDetailCellSx}>
                    Nenhum vendedor encontrado com os filtros informados.
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
              ? `Confirma a exclusão do vendedor ${pendingDelete.name}?`
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
