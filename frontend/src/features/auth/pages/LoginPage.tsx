import LockRoundedIcon from "@mui/icons-material/LockRounded";
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
import { Link as RouterLink, useNavigate } from "react-router-dom";
import { z } from "zod";
import { useAuth } from "../hooks/useAuth";

const loginSchema = z.object({
  email: z.string().email("Informe um e-mail valido"),
  password: z.string().min(1, "Informe sua senha"),
});

type LoginFormValues = z.infer<typeof loginSchema>;

function getLoginErrorMessage(error: unknown) {
  if (isAxiosError<{ message?: string }>(error)) {
    return (
      error.response?.data?.message ?? "Nao foi possivel entrar no sistema."
    );
  }

  return "Nao foi possivel entrar no sistema.";
}

export function LoginPage() {
  const navigate = useNavigate();
  const { login } = useAuth();
  const [submitError, setSubmitError] = useState<string | null>(null);
  const {
    formState: { errors, isSubmitting },
    handleSubmit,
    register,
  } = useForm<LoginFormValues>({
    defaultValues: {
      email: "admin@local.dev",
      password: "123456",
    },
    resolver: zodResolver(loginSchema),
  });

  const onSubmit = async (values: LoginFormValues) => {
    try {
      setSubmitError(null);
      await login(values);

      navigate("/", { replace: true });
    } catch (error) {
      setSubmitError(getLoginErrorMessage(error));
    }
  };

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
                <LockRoundedIcon />
              </Box>
              <Typography variant="h3">Entrar no painel</Typography>
              <Typography color="text.secondary" variant="body1">
                Use seu acesso administrativo para entrar no sistema de gestao
                de orcamentos.
              </Typography>
            </Box>

            <Alert severity="info" variant="outlined">
              Credenciais iniciais de desenvolvimento: `admin@local.dev` e
              `123456`.
            </Alert>

            {submitError ? <Alert severity="error">{submitError}</Alert> : null}

            <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
              <TextField
                error={Boolean(errors.email)}
                fullWidth
                helperText={errors.email?.message}
                label="E-mail"
                placeholder="admin@local.dev"
                type="email"
                {...register("email")}
              />
              <TextField
                error={Boolean(errors.password)}
                fullWidth
                helperText={errors.password?.message}
                label="Senha"
                placeholder="Digite sua senha"
                type="password"
                {...register("password")}
              />
              <Button
                disabled={isSubmitting}
                size="large"
                type="submit"
                variant="contained"
              >
                {isSubmitting ? "Entrando..." : "Entrar"}
              </Button>
            </Box>

            <Typography color="text.secondary" variant="body2">
              Quer revisar a base do painel? Veja a{" "}
              <Typography
                component={RouterLink}
                sx={{ color: "primary.main", textDecoration: "none" }}
                to="/"
              >
                estrutura inicial
              </Typography>
              .
            </Typography>
          </Box>
        </Paper>
      </Container>
    </Box>
  );
}
