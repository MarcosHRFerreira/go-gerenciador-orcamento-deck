import LockResetRoundedIcon from "@mui/icons-material/LockResetRounded";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  Alert,
  Box,
  Button,
  Container,
  Paper,
  TextField,
  Typography,
} from "@mui/material";
import { isAxiosError } from "axios";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { useNavigate } from "react-router-dom";
import { z } from "zod";
import BrandLogo from "../../../components/common/BrandLogo";
import {
  createStrongPasswordSchema,
  passwordStrengthHint,
} from "../../../shared/validation/password";
import { changePasswordRequest } from "../api/auth";
import { useAuth } from "../hooks/useAuth";

const changePasswordSchema = z
  .object({
    currentPassword: z.string().min(1, "Informe a senha atual"),
    newPassword: createStrongPasswordSchema("A nova senha"),
    newPasswordConfirm: z.string().min(8, "Confirme a nova senha"),
  })
  .refine((values) => values.newPassword === values.newPasswordConfirm, {
    message: "A confirmação deve ser igual à nova senha",
    path: ["newPasswordConfirm"],
  })
  .refine((values) => values.currentPassword !== values.newPassword, {
    message: "A nova senha deve ser diferente da senha atual",
    path: ["newPassword"],
  });

type ChangePasswordFormValues = z.infer<typeof changePasswordSchema>;

function getChangePasswordErrorMessage(error: unknown) {
  if (isAxiosError<{ message?: string }>(error)) {
    return (
      error.response?.data?.message ?? "Não foi possível atualizar a sua senha."
    );
  }

  return "Não foi possível atualizar a sua senha.";
}

export function ChangePasswordPage() {
  const navigate = useNavigate();
  const { replaceSession, user } = useAuth();
  const [submitError, setSubmitError] = useState<string | null>(null);
  const {
    formState: { errors, isSubmitting },
    handleSubmit,
    register,
  } = useForm<ChangePasswordFormValues>({
    defaultValues: {
      currentPassword: "",
      newPassword: "",
      newPasswordConfirm: "",
    },
    resolver: zodResolver(changePasswordSchema),
  });

  const onSubmit = async (values: ChangePasswordFormValues) => {
    try {
      setSubmitError(null);
      const nextSession = await changePasswordRequest({
        current_password: values.currentPassword,
        new_password: values.newPassword,
        new_password_confirm: values.newPasswordConfirm,
      });

      await replaceSession(nextSession);
      navigate("/budgets", { replace: true });
    } catch (error) {
      setSubmitError(getChangePasswordErrorMessage(error));
    }
  };

  const isFirstAccess = user?.must_change_password ?? false;

  return (
    <Box
      sx={{
        alignItems: "center",
        background: "linear-gradient(180deg, #F8FAFC 0%, #EEF2FF 100%)",
        display: "flex",
        minHeight: "100vh",
        py: 4,
      }}
    >
      <Container maxWidth="sm">
        <Paper sx={{ borderRadius: 6, p: { md: 5, xs: 3 } }}>
          <Box
            component="form"
            onSubmit={handleSubmit(onSubmit)}
            sx={{ display: "flex", flexDirection: "column", gap: 3 }}
          >
            <Box sx={{ display: "flex", flexDirection: "column", gap: 1.5 }}>
              <Box sx={{ display: "flex", justifyContent: "center" }}>
                <BrandLogo
                  imageSx={{ width: { sm: 220, xs: 180 } }}
                  wrapperSx={{
                    border: "1px solid",
                    borderColor: "divider",
                    boxShadow: "0 10px 24px rgba(15, 23, 42, 0.06)",
                    p: 1,
                  }}
                />
              </Box>
              <Box
                sx={{
                  alignItems: "center",
                  bgcolor: "primary.light",
                  borderRadius: 3,
                  color: "primary.main",
                  display: "inline-flex",
                  height: 48,
                  justifyContent: "center",
                  width: 48,
                }}
              >
                <LockResetRoundedIcon />
              </Box>
              <Typography variant="h3">Alterar senha</Typography>
              <Typography color="text.secondary" variant="body1">
                Atualize sua senha para continuar utilizando o sistema.
              </Typography>
            </Box>

            <Alert
              severity={isFirstAccess ? "warning" : "info"}
              variant="outlined"
            >
              {isFirstAccess
                ? "Este é o seu primeiro acesso. Antes de entrar no painel, você precisa definir uma nova senha."
                : "Informe a senha atual e escolha uma nova senha para o seu acesso."}
            </Alert>
            <Alert severity="info" variant="outlined">
              {passwordStrengthHint}
            </Alert>

            {submitError ? <Alert severity="error">{submitError}</Alert> : null}

            <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
              <TextField
                error={Boolean(errors.currentPassword)}
                fullWidth
                helperText={errors.currentPassword?.message}
                label="Senha atual"
                placeholder="Digite sua senha atual"
                type="password"
                {...register("currentPassword")}
              />
              <TextField
                error={Boolean(errors.newPassword)}
                fullWidth
                helperText={errors.newPassword?.message}
                label="Nova senha"
                placeholder="Digite a nova senha"
                type="password"
                {...register("newPassword")}
              />
              <TextField
                error={Boolean(errors.newPasswordConfirm)}
                fullWidth
                helperText={errors.newPasswordConfirm?.message}
                label="Confirmar nova senha"
                placeholder="Repita a nova senha"
                type="password"
                {...register("newPasswordConfirm")}
              />
              <Button
                disabled={isSubmitting}
                size="large"
                type="submit"
                variant="contained"
              >
                {isSubmitting ? "Salvando..." : "Salvar nova senha"}
              </Button>
            </Box>
          </Box>
        </Paper>
      </Container>
    </Box>
  );
}
