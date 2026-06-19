import { useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "react-router-dom";
import { createUserRequest } from "../api/users";
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

export function UserCreatePage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const handleSubmit = async (
    payload: CreateUserPayload | UpdateUserPayload,
  ) => {
    if (!("password" in payload)) {
      throw new Error("Payload invalido para criacao de usuario.");
    }

    await createUserRequest(payload);
    await queryClient.invalidateQueries({ queryKey: ["users"] });
    navigate("/users");
  };

  return (
    <UserForm
      initialValues={defaultUserFormValues}
      mode="create"
      onCancel={() => navigate("/users")}
      onSubmit={handleSubmit}
      submitLabel="Salvar usuario"
      subtitle="Cadastre um novo acesso e defina se ele sera administrativo ou operacional."
      title="Novo usuario"
    />
  );
}
