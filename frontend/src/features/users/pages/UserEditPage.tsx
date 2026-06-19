import { Alert, Box, LinearProgress } from "@mui/material";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useMemo } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { PageHeader } from "../../../components/common/PageHeader";
import { SectionCard } from "../../../components/common/SectionCard";
import { listUsersRequest, updateUserRequest } from "../api/users";
import { UserForm, type UserFormValues } from "../components/UserForm";
import type { CreateUserPayload, UpdateUserPayload } from "../types/user";

const defaultUserFormValues: UserFormValues = {
  email: "",
  name: "",
  password: "",
  passwordConfirm: "",
  role: "user",
  userKind: "salesperson",
  username: "",
};

export function UserEditPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { userId } = useParams();

  const parsedUserId = Number(userId);
  const isValidUserId = Number.isInteger(parsedUserId) && parsedUserId > 0;

  const usersQuery = useQuery({
    enabled: isValidUserId,
    queryFn: listUsersRequest,
    queryKey: ["users"],
  });

  const currentUser = useMemo(() => {
    if (!isValidUserId) {
      return null;
    }

    const users = usersQuery.data ?? [];

    return users.find((item) => item.id === parsedUserId) ?? null;
  }, [isValidUserId, parsedUserId, usersQuery.data]);

  const initialValues = useMemo<UserFormValues>(() => {
    if (currentUser === null) {
      return defaultUserFormValues;
    }

    return {
      email: currentUser.email,
      name: currentUser.name,
      password: "",
      passwordConfirm: "",
      role: currentUser.role,
      userKind: currentUser.userKind ?? "salesperson",
      username: currentUser.username,
    };
  }, [currentUser]);

  const handleSubmit = async (
    payload: CreateUserPayload | UpdateUserPayload,
  ) => {
    if ("password" in payload) {
      throw new Error("Payload invalido para edicao de usuario.");
    }

    if (!isValidUserId) {
      throw new Error("Usuario invalido para edicao.");
    }

    await updateUserRequest(parsedUserId, payload);
    await queryClient.invalidateQueries({ queryKey: ["users"] });
    navigate("/users");
  };

  if (!isValidUserId) {
    return (
      <Box sx={{ display: "flex", flexDirection: "column", gap: 3 }}>
        <PageHeader
          description="O identificador informado para edicao nao e valido."
          title="Editar usuario"
        />
        <SectionCard
          description="Revise a navegacao e tente novamente pela listagem de usuarios."
          title="Usuario invalido"
        >
          <Alert severity="error">Nao foi possivel identificar o usuario.</Alert>
        </SectionCard>
      </Box>
    );
  }

  if (usersQuery.isLoading) {
    return (
      <Box sx={{ display: "flex", flexDirection: "column", gap: 3 }}>
        <PageHeader
          description="Carregando os dados cadastrais para edicao."
          title="Editar usuario"
        />
        <SectionCard
          description="Aguarde enquanto os dados do usuario sao carregados."
          title="Dados do usuario"
        >
          <LinearProgress />
        </SectionCard>
      </Box>
    );
  }

  if (usersQuery.isError || currentUser === null) {
    return (
      <Box sx={{ display: "flex", flexDirection: "column", gap: 3 }}>
        <PageHeader
          description="Nao foi possivel carregar o usuario solicitado."
          title="Editar usuario"
        />
        <SectionCard
          description="Volte para a listagem e selecione novamente o usuario."
          title="Falha ao carregar"
        >
          <Alert severity="error">Usuario nao encontrado para edicao.</Alert>
        </SectionCard>
      </Box>
    );
  }

  return (
    <UserForm
      initialValues={initialValues}
      mode="edit"
      onCancel={() => navigate("/users")}
      onSubmit={handleSubmit}
      submitLabel="Salvar alteracoes"
      subtitle="Atualize nome, e-mail, username e perfil do usuario selecionado."
      title="Editar usuario"
    />
  );
}
