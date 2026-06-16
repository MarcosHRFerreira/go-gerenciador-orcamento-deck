import ArrowBackRoundedIcon from "@mui/icons-material/ArrowBackRounded";
import SaveRoundedIcon from "@mui/icons-material/SaveRounded";
import { zodResolver } from "@hookform/resolvers/zod";
import { Alert, Box, Button, MenuItem, TextField } from "@mui/material";
import { isAxiosError } from "axios";
import { useState } from "react";
import { useForm, type Resolver } from "react-hook-form";
import { z } from "zod";
import { PageHeader } from "../../../components/common/PageHeader";
import { SectionCard } from "../../../components/common/SectionCard";
import {
  createStrongPasswordSchema,
  passwordStrengthHint,
} from "../../../shared/validation/password";
import type {
  CreateUserPayload,
  UpdateUserPayload,
  UserRole,
} from "../types/user";

const baseUserFormSchema = z.object({
  email: z.string().trim().email("Informe um e-mail valido"),
  name: z
    .string()
    .trim()
    .min(1, "Informe o nome")
    .max(150, "O nome deve ter no maximo 150 caracteres"),
  role: z.enum(["admin", "user"]),
  username: z
    .string()
    .trim()
    .min(1, "Informe o username")
    .max(100, "O username deve ter no maximo 100 caracteres"),
});

const createUserFormSchema = baseUserFormSchema
  .extend({
    password: createStrongPasswordSchema("A senha"),
    passwordConfirm: z.string().min(8, "Confirme a senha"),
  })
  .refine((values) => values.password === values.passwordConfirm, {
    message: "A confirmacao de senha deve ser igual a senha",
    path: ["passwordConfirm"],
  });

export type UserFormValues = z.infer<typeof createUserFormSchema>;

const editUserFormSchema = baseUserFormSchema.extend({
  password: z.string(),
  passwordConfirm: z.string(),
});

type UserFormProps = {
  backLabel?: string;
  initialValues: UserFormValues;
  mode: "create" | "edit";
  onCancel: () => void;
  onSubmit: (payload: CreateUserPayload | UpdateUserPayload) => Promise<void>;
  submitLabel: string;
  subtitle: string;
  title: string;
};

function getSubmitErrorMessage(error: unknown) {
  if (isAxiosError<{ message?: string }>(error)) {
    return (
      error.response?.data?.message ?? "Nao foi possivel salvar o usuario."
    );
  }

  return "Nao foi possivel salvar o usuario.";
}

function getRoleLabel(role: UserRole) {
  return role === "admin" ? "Administrador" : "Usuario";
}

function mapFormValuesToPayload(
  values: UserFormValues,
  mode: "create" | "edit",
): CreateUserPayload | UpdateUserPayload {
  const basePayload = {
    email: values.email.trim(),
    name: values.name.trim(),
    role: values.role,
    username: values.username.trim(),
  };

  if (mode === "edit") {
    return basePayload;
  }

  return {
    ...basePayload,
    password: values.password,
    passwordConfirm: values.passwordConfirm,
  };
}

export function UserForm({
  backLabel = "Voltar",
  initialValues,
  mode,
  onCancel,
  onSubmit,
  submitLabel,
  subtitle,
  title,
}: UserFormProps) {
  const [submitError, setSubmitError] = useState<string | null>(null);
  const resolver =
    mode === "create"
      ? zodResolver(createUserFormSchema)
      : zodResolver(editUserFormSchema);
  const {
    formState: { errors, isSubmitting },
    handleSubmit,
    register,
  } = useForm<UserFormValues>({
    defaultValues: initialValues,
    resolver: resolver as Resolver<UserFormValues>,
  });

  const handleFormSubmit = async (values: UserFormValues) => {
    try {
      setSubmitError(null);
      await onSubmit(mapFormValuesToPayload(values, mode));
    } catch (error) {
      setSubmitError(getSubmitErrorMessage(error));
    }
  };

  return (
    <Box component="form" onSubmit={handleSubmit(handleFormSubmit)}>
      <PageHeader
        action={
          <Button
            onClick={onCancel}
            startIcon={<ArrowBackRoundedIcon />}
            variant="outlined"
          >
            {backLabel}
          </Button>
        }
        description={subtitle}
        title={title}
      />

      <Box sx={{ mt: 3 }}>
        <SectionCard
          description={
            mode === "create"
              ? "Cadastre acessos administrativos ou operacionais com o perfil adequado."
              : "Atualize os dados cadastrais e o perfil de acesso do usuario."
          }
          title="Dados do usuario"
        >
          {submitError ? <Alert severity="error">{submitError}</Alert> : null}
          {mode === "create" ? (
            <Alert severity="info" variant="outlined">
              {passwordStrengthHint}
            </Alert>
          ) : null}

          <Box
            sx={{
              display: "grid",
              gap: 2,
              gridTemplateColumns: {
                md: "repeat(2, minmax(0, 1fr))",
                xs: "minmax(0, 1fr)",
              },
            }}
          >
            <TextField
              error={Boolean(errors.name)}
              fullWidth
              helperText={errors.name?.message}
              label="Nome"
              placeholder="Maria Souza"
              {...register("name")}
            />
            <TextField
              error={Boolean(errors.username)}
              fullWidth
              helperText={errors.username?.message}
              label="Username"
              placeholder="maria.souza"
              {...register("username")}
            />
            <TextField
              error={Boolean(errors.email)}
              fullWidth
              helperText={errors.email?.message}
              label="E-mail"
              placeholder="maria@empresa.com"
              type="email"
              {...register("email")}
            />
            <TextField
              error={Boolean(errors.role)}
              fullWidth
              helperText={errors.role?.message}
              label="Perfil"
              select
              {...register("role")}
            >
              {(["user", "admin"] as const).map((role) => (
                <MenuItem key={role} value={role}>
                  {getRoleLabel(role)}
                </MenuItem>
              ))}
            </TextField>
            {mode === "create" ? (
              <TextField
                error={Boolean(errors.password)}
                fullWidth
                helperText={errors.password?.message}
                label="Senha"
                placeholder="Digite a senha inicial"
                type="password"
                {...register("password")}
              />
            ) : null}
            {mode === "create" ? (
              <TextField
                error={Boolean(errors.passwordConfirm)}
                fullWidth
                helperText={errors.passwordConfirm?.message}
                label="Confirmar senha"
                placeholder="Repita a senha"
                type="password"
                {...register("passwordConfirm")}
              />
            ) : null}
          </Box>

          <Box
            sx={{
              display: "flex",
              gap: 1.5,
              justifyContent: "flex-end",
              pt: 1,
            }}
          >
            <Button onClick={onCancel} variant="text">
              Cancelar
            </Button>
            <Button
              disabled={isSubmitting}
              startIcon={<SaveRoundedIcon />}
              type="submit"
              variant="contained"
            >
              {isSubmitting ? "Salvando..." : submitLabel}
            </Button>
          </Box>
        </SectionCard>
      </Box>
    </Box>
  );
}
