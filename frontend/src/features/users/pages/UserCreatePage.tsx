import { useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "react-router-dom";
import { createUserRequest } from "../api/users";
import { UserForm, type UserFormValues } from "../components/UserForm";
import type { CreateUserPayload } from "../types/user";

const defaultUserFormValues: UserFormValues = {
  email: "",
  name: "",
  password: "",
  passwordConfirm: "",
  role: "user",
  username: "",
};

export function UserCreatePage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const handleSubmit = async (payload: CreateUserPayload) => {
    await createUserRequest(payload);
    await queryClient.invalidateQueries({ queryKey: ["users"] });
    navigate("/users");
  };

  return (
    <UserForm
      initialValues={defaultUserFormValues}
      onCancel={() => navigate("/users")}
      onSubmit={handleSubmit}
      submitLabel="Salvar usuario"
      subtitle="Cadastre um novo acesso e defina se ele sera administrativo ou operacional."
      title="Novo usuario"
    />
  );
}
