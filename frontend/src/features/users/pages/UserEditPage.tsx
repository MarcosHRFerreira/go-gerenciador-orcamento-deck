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
      throw new Error("Payload inválido para edição de usuário.");
    }

    if (!isValidUserId) {
      throw new Error("Usuário inválido para edição.");
    }

    await updateUserRequest(parsedUserId, payload);
    await queryClient.invalidateQueries({ queryKey: ["users"] });
    navigate("/users");
  };

  if (!isValidUserId) {
    return (
      <Box sx={{ display: "flex", flexDirection: "column", gap: 3 }}>
        <PageHeader
          description="O identificador informado para edição não é válido."
          title="Editar usuário"
        />
        <SectionCard
          description="Revise a navegação e tente novamente pela listagem de usuários."
          title="Usuário inválido"
        >
          <Alert severity="error">
            Não foi possível identificar o usuário.
          </Alert>
        </SectionCard>
      </Box>
    );
  }

  if (usersQuery.isLoading) {
    return (
      <Box sx={{ display: "flex", flexDirection: "column", gap: 3 }}>
        <PageHeader
          description="Carregando os dados cadastrais para edição."
          title="Editar usuário"
        />
        <SectionCard
          description="Aguarde enquanto os dados do usuário são carregados."
          title="Dados do usuário"
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
          description="Não foi possível carregar o usuário solicitado."
          title="Editar usuário"
        />
        <SectionCard
          description="Volte para a listagem e selecione novamente o usuário."
          title="Falha ao carregar"
        >
          <Alert severity="error">Usuário não encontrado para edição.</Alert>
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
      submitLabel="Salvar alterações"
      subtitle="Atualize nome, e-mail, username e perfil do usuário selecionado."
      title="Editar usuário"
    />
  );
}
