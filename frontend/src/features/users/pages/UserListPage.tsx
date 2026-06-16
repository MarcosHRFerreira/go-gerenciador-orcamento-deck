import AddRoundedIcon from "@mui/icons-material/AddRounded";
import AdminPanelSettingsRoundedIcon from "@mui/icons-material/AdminPanelSettingsRounded";
import BlockRoundedIcon from "@mui/icons-material/BlockRounded";
import EditRoundedIcon from "@mui/icons-material/EditRounded";
import LockResetRoundedIcon from "@mui/icons-material/LockResetRounded";
import PersonOutlineRoundedIcon from "@mui/icons-material/PersonOutlineRounded";
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
import { useMemo, useState } from "react";
import { useForm } from "react-hook-form";
import { useNavigate } from "react-router-dom";
import { z } from "zod";
import { PageHeader } from "../../../components/common/PageHeader";
import { SectionCard } from "../../../components/common/SectionCard";
import {
  createStrongPasswordSchema,
  passwordStrengthHint,
} from "../../../shared/validation/password";
import { useAuth } from "../../auth/hooks/useAuth";
import {
  listUsersRequest,
  resetUserPasswordRequest,
  updateUserActiveRequest,
  updateUserRoleRequest,
} from "../api/users";
import type {
  ResetUserPasswordPayload,
  UserItem,
  UserListFilters,
  UserRole,
} from "../types/user";

const dateTimeFormatter = new Intl.DateTimeFormat("pt-BR", {
  dateStyle: "short",
  timeStyle: "short",
});

type PendingUserAction =
  | {
      type: "role";
      user: UserItem;
      nextRole: UserRole;
    }
  | {
      type: "active";
      user: UserItem;
      nextActive: boolean;
    };

const resetPasswordSchema = z
  .object({
    password: createStrongPasswordSchema("A senha temporaria"),
    passwordConfirm: z.string().min(8, "Confirme a senha temporaria"),
  })
  .refine((values) => values.password === values.passwordConfirm, {
    message: "A confirmacao deve ser igual a senha temporaria",
    path: ["passwordConfirm"],
  });

type ResetPasswordFormValues = z.infer<typeof resetPasswordSchema>;

const defaultFilters: UserListFilters = {
  role: "all",
  search: "",
  status: "all",
};

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

function getRoleLabel(role: UserRole) {
  return role === "admin" ? "Administrador" : "Usuario";
}

function getRoleChipColor(role: UserRole) {
  return role === "admin" ? "primary" : "default";
}

function getStatusLabel(active: boolean) {
  return active ? "Ativo" : "Inativo";
}

function getStatusChipColor(active: boolean) {
  return active ? "success" : "default";
}

function getMutationErrorMessage(error: unknown) {
  if (isAxiosError<{ message?: string }>(error)) {
    return (
      error.response?.data?.message ??
      "Nao foi possivel atualizar os dados do usuario."
    );
  }

  return "Nao foi possivel atualizar os dados do usuario.";
}

function getRoleActionLabel(nextRole: UserRole) {
  return nextRole === "admin" ? "Tornar administrador" : "Tornar usuario";
}

function getPendingActionDescription(action: PendingUserAction) {
  if (action.type === "role") {
    return `Confirma a alteracao do perfil de ${action.user.name} para ${getRoleLabel(action.nextRole).toLowerCase()}?`;
  }

  return action.nextActive
    ? `Confirma a reativacao do usuario ${action.user.name}?`
    : `Confirma a desativacao do usuario ${action.user.name}?`;
}

function mapResetPasswordFormValues(
  values: ResetPasswordFormValues,
): ResetUserPasswordPayload {
  return {
    password: values.password,
    passwordConfirm: values.passwordConfirm,
  };
}

export function UserListPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { user: currentUser } = useAuth();
  const [filters, setFilters] = useState<UserListFilters>(defaultFilters);
  const [feedbackMessage, setFeedbackMessage] = useState<string | null>(null);
  const [feedbackError, setFeedbackError] = useState<string | null>(null);
  const [pendingAction, setPendingAction] = useState<PendingUserAction | null>(
    null,
  );
  const [resetPasswordUser, setResetPasswordUser] = useState<UserItem | null>(
    null,
  );
  const {
    formState: { errors: resetPasswordErrors },
    handleSubmit: handleResetPasswordSubmit,
    register: registerResetPassword,
    reset: resetResetPasswordForm,
  } = useForm<ResetPasswordFormValues>({
    defaultValues: {
      password: "",
      passwordConfirm: "",
    },
    resolver: zodResolver(resetPasswordSchema),
  });

  const usersQuery = useQuery({
    queryKey: ["users"],
    queryFn: listUsersRequest,
  });

  const updateRoleMutation = useMutation({
    mutationFn: ({
      nextRole,
      userId,
    }: {
      nextRole: UserRole;
      userId: number;
    }) => updateUserRoleRequest(userId, { role: nextRole }),
    onSuccess: async (_, variables) => {
      await queryClient.invalidateQueries({ queryKey: ["users"] });
      setFeedbackError(null);
      setFeedbackMessage(
        `Perfil atualizado para ${getRoleLabel(variables.nextRole).toLowerCase()}.`,
      );
      setPendingAction(null);
    },
    onError: (error) => {
      setFeedbackMessage(null);
      setFeedbackError(getMutationErrorMessage(error));
    },
  });

  const updateActiveMutation = useMutation({
    mutationFn: ({ active, userId }: { active: boolean; userId: number }) =>
      updateUserActiveRequest(userId, { active }),
    onSuccess: async (_, variables) => {
      await queryClient.invalidateQueries({ queryKey: ["users"] });
      setFeedbackError(null);
      setFeedbackMessage(
        variables.active
          ? "Usuario reativado com sucesso."
          : "Usuario desativado com sucesso.",
      );
      setPendingAction(null);
    },
    onError: (error) => {
      setFeedbackMessage(null);
      setFeedbackError(getMutationErrorMessage(error));
    },
  });

  const resetPasswordMutation = useMutation({
    mutationFn: ({
      payload,
      userId,
    }: {
      payload: ResetUserPasswordPayload;
      userId: number;
      userName: string;
    }) => resetUserPasswordRequest(userId, payload),
    onSuccess: async (_, variables) => {
      await queryClient.invalidateQueries({ queryKey: ["users"] });
      setFeedbackError(null);
      setFeedbackMessage(
        `Senha resetada para ${variables.userName}. No proximo acesso, o usuario devera trocar a senha.`,
      );
      setResetPasswordUser(null);
      resetResetPasswordForm();
    },
    onError: (error) => {
      setFeedbackMessage(null);
      setFeedbackError(getMutationErrorMessage(error));
    },
  });

  const filteredUsers = useMemo(() => {
    const items = usersQuery.data ?? [];
    const normalizedSearch = filters.search.trim().toLocaleLowerCase("pt-BR");

    return items
      .filter((item) => {
        const matchesSearch =
          normalizedSearch.length === 0 ||
          item.name.toLocaleLowerCase("pt-BR").includes(normalizedSearch) ||
          item.email.toLocaleLowerCase("pt-BR").includes(normalizedSearch) ||
          item.username.toLocaleLowerCase("pt-BR").includes(normalizedSearch);
        const matchesRole =
          filters.role === "all" || item.role === filters.role;
        const matchesStatus =
          filters.status === "all" ||
          (filters.status === "active" && item.active) ||
          (filters.status === "inactive" && !item.active);

        return matchesSearch && matchesRole && matchesStatus;
      })
      .sort((firstItem, secondItem) =>
        firstItem.name.localeCompare(secondItem.name, "pt-BR"),
      );
  }, [filters.role, filters.search, filters.status, usersQuery.data]);

  const isMutating =
    updateRoleMutation.isPending ||
    updateActiveMutation.isPending ||
    resetPasswordMutation.isPending;

  const handleConfirmAction = async () => {
    if (!pendingAction) {
      return;
    }

    if (pendingAction.type === "role") {
      await updateRoleMutation.mutateAsync({
        nextRole: pendingAction.nextRole,
        userId: pendingAction.user.id,
      });
      return;
    }

    await updateActiveMutation.mutateAsync({
      active: pendingAction.nextActive,
      userId: pendingAction.user.id,
    });
  };

  const handleResetPassword = async (values: ResetPasswordFormValues) => {
    if (!resetPasswordUser) {
      return;
    }

    await resetPasswordMutation.mutateAsync({
      payload: mapResetPasswordFormValues(values),
      userId: resetPasswordUser.id,
      userName: resetPasswordUser.name,
    });
  };

  return (
    <Box sx={{ display: "flex", flexDirection: "column", gap: 3 }}>
      <PageHeader
        action={
          <Box sx={{ width: "100%" }}>
            <Button
              onClick={() => navigate("/users/new")}
              startIcon={<AddRoundedIcon />}
              variant="contained"
            >
              Novo usuario
            </Button>
          </Box>
        }
        description="Gerencie acessos, perfil administrativo e status dos usuarios do sistema."
        title="Usuarios"
      />

      <SectionCard
        description="Use os filtros para localizar usuarios rapidamente antes de alterar perfil ou status."
        title="Filtros"
      >
        <Box
          sx={{
            display: "grid",
            gap: 2,
            gridTemplateColumns: {
              lg: "2fr 1fr 1fr",
              md: "repeat(2, minmax(0, 1fr))",
              xs: "minmax(0, 1fr)",
            },
          }}
        >
          <TextField
            label="Buscar por nome, e-mail ou username"
            onChange={(event) =>
              setFilters((currentFilters) => ({
                ...currentFilters,
                search: event.target.value,
              }))
            }
            size="small"
            value={filters.search}
          />
          <TextField
            label="Perfil"
            onChange={(event) =>
              setFilters((currentFilters) => ({
                ...currentFilters,
                role: event.target.value as UserListFilters["role"],
              }))
            }
            select
            size="small"
            value={filters.role}
          >
            <MenuItem value="all">Todos</MenuItem>
            <MenuItem value="admin">Administrador</MenuItem>
            <MenuItem value="user">Usuario</MenuItem>
          </TextField>
          <TextField
            label="Status"
            onChange={(event) =>
              setFilters((currentFilters) => ({
                ...currentFilters,
                status: event.target.value as UserListFilters["status"],
              }))
            }
            select
            size="small"
            value={filters.status}
          >
            <MenuItem value="all">Todos</MenuItem>
            <MenuItem value="active">Ativo</MenuItem>
            <MenuItem value="inactive">Inativo</MenuItem>
          </TextField>
        </Box>
      </SectionCard>

      <SectionCard
        description={`${
          filteredUsers.length
        } usuario(s) encontrado(s) com os filtros atuais.`}
        title="Listagem"
      >
        {usersQuery.isLoading ? <LinearProgress /> : null}
        {usersQuery.isError ? (
          <Alert severity="error">
            Nao foi possivel carregar a listagem de usuarios.
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
                <TableCell sx={tableHeadCellSx}>Username</TableCell>
                <TableCell sx={tableHeadCellSx}>Perfil</TableCell>
                <TableCell sx={tableHeadCellSx}>Status</TableCell>
                <TableCell sx={tableHeadCellSx}>Atualizado em</TableCell>
                <TableCell align="right" sx={tableHeadCellSx}>
                  Acoes
                </TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {filteredUsers.map((item) => {
                const isSelf = currentUser?.id === item.id;
                const nextRole: UserRole =
                  item.role === "admin" ? "user" : "admin";
                const nextActive = !item.active;

                return (
                  <TableRow hover key={item.id}>
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
                        <Typography color="text.secondary" variant="caption">
                          ID {item.id}
                        </Typography>
                      </Box>
                    </TableCell>
                    <TableCell sx={tableDetailCellSx}>{item.email}</TableCell>
                    <TableCell sx={tableDetailCellSx}>
                      {item.username}
                    </TableCell>
                    <TableCell sx={tableDetailCellSx}>
                      <Chip
                        color={getRoleChipColor(item.role)}
                        label={getRoleLabel(item.role)}
                        size="small"
                        variant={item.role === "admin" ? "filled" : "outlined"}
                      />
                    </TableCell>
                    <TableCell sx={tableDetailCellSx}>
                      <Box
                        sx={{
                          display: "flex",
                          flexDirection: "column",
                          gap: 0.75,
                        }}
                      >
                        <Chip
                          color={getStatusChipColor(item.active)}
                          label={getStatusLabel(item.active)}
                          size="small"
                          variant={item.active ? "filled" : "outlined"}
                        />
                        {item.mustChangePassword ? (
                          <Chip
                            color="warning"
                            label="Troca pendente"
                            size="small"
                            variant="outlined"
                          />
                        ) : null}
                      </Box>
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
                          onClick={() => navigate(`/users/${item.id}/edit`)}
                          size="small"
                          startIcon={<EditRoundedIcon />}
                          variant="text"
                        >
                          Editar
                        </Button>
                        <Button
                          disabled={isMutating || isSelf}
                          onClick={() =>
                            setPendingAction({
                              nextRole,
                              type: "role",
                              user: item,
                            })
                          }
                          size="small"
                          startIcon={
                            nextRole === "admin" ? (
                              <AdminPanelSettingsRoundedIcon />
                            ) : (
                              <PersonOutlineRoundedIcon />
                            )
                          }
                          variant="text"
                        >
                          {getRoleActionLabel(nextRole)}
                        </Button>
                        <Button
                          color="warning"
                          disabled={isMutating || isSelf}
                          onClick={() => {
                            resetResetPasswordForm();
                            setResetPasswordUser(item);
                          }}
                          size="small"
                          startIcon={<LockResetRoundedIcon />}
                          variant="text"
                        >
                          Resetar senha
                        </Button>
                        <Button
                          color={item.active ? "error" : "success"}
                          disabled={isMutating || isSelf}
                          onClick={() =>
                            setPendingAction({
                              nextActive,
                              type: "active",
                              user: item,
                            })
                          }
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
                      </Box>
                    </TableCell>
                  </TableRow>
                );
              })}
              {!usersQuery.isLoading && filteredUsers.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={7} sx={tableDetailCellSx}>
                    Nenhum usuario encontrado com os filtros informados.
                  </TableCell>
                </TableRow>
              ) : null}
            </TableBody>
          </Table>
        </TableContainer>
      </SectionCard>

      <Dialog
        onClose={() => {
          if (!isMutating) {
            setResetPasswordUser(null);
            resetResetPasswordForm();
          }
        }}
        open={resetPasswordUser !== null}
      >
        <Box
          component="form"
          onSubmit={handleResetPasswordSubmit((values) => {
            void handleResetPassword(values);
          })}
        >
          <DialogTitle>Resetar senha</DialogTitle>
          <DialogContent
            sx={{ display: "flex", flexDirection: "column", gap: 2, pt: 1 }}
          >
            <DialogContentText>
              {resetPasswordUser
                ? `Defina uma senha temporaria para ${resetPasswordUser.name}. No proximo acesso, o usuario sera obrigado a trocar a senha.`
                : ""}
            </DialogContentText>
            <Alert severity="info" variant="outlined">
              {passwordStrengthHint}
            </Alert>
            <TextField
              autoFocus
              error={Boolean(resetPasswordErrors.password)}
              helperText={resetPasswordErrors.password?.message}
              label="Senha temporaria"
              type="password"
              {...registerResetPassword("password")}
            />
            <TextField
              error={Boolean(resetPasswordErrors.passwordConfirm)}
              helperText={resetPasswordErrors.passwordConfirm?.message}
              label="Confirmar senha temporaria"
              type="password"
              {...registerResetPassword("passwordConfirm")}
            />
          </DialogContent>
          <DialogActions>
            <Button
              disabled={isMutating}
              onClick={() => {
                setResetPasswordUser(null);
                resetResetPasswordForm();
              }}
              variant="outlined"
            >
              Cancelar
            </Button>
            <Button disabled={isMutating} type="submit" variant="contained">
              {isMutating ? "Salvando..." : "Confirmar reset"}
            </Button>
          </DialogActions>
        </Box>
      </Dialog>

      <Dialog
        onClose={() => {
          if (!isMutating) {
            setPendingAction(null);
          }
        }}
        open={pendingAction !== null}
      >
        <DialogTitle>Confirmar alteracao</DialogTitle>
        <DialogContent>
          <DialogContentText>
            {pendingAction ? getPendingActionDescription(pendingAction) : ""}
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button
            disabled={isMutating}
            onClick={() => setPendingAction(null)}
            variant="outlined"
          >
            Cancelar
          </Button>
          <Button
            disabled={isMutating}
            onClick={() => {
              void handleConfirmAction();
            }}
            variant="contained"
          >
            {isMutating ? "Salvando..." : "Confirmar"}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
