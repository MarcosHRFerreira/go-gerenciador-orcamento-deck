import ArrowForwardRoundedIcon from "@mui/icons-material/ArrowForwardRounded";
import QueryStatsRoundedIcon from "@mui/icons-material/QueryStatsRounded";
import ShieldRoundedIcon from "@mui/icons-material/ShieldRounded";
import LockRoundedIcon from "@mui/icons-material/LockRounded";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  Alert,
  Box,
  Button,
  Container,
  Divider,
  Paper,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { alpha } from "@mui/material/styles";
import { isAxiosError } from "axios";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { Link as RouterLink, useNavigate } from "react-router-dom";
import { z } from "zod";
import BrandLogo from "../../../components/common/BrandLogo";
import { useAuth } from "../hooks/useAuth";

const loginSchema = z.object({
  email: z.string().email("Informe um e-mail válido"),
  password: z.string().min(1, "Informe sua senha"),
});

type LoginFormValues = z.infer<typeof loginSchema>;

function getLoginErrorMessage(error: unknown) {
  if (isAxiosError<{ message?: string }>(error)) {
    return (
      error.response?.data?.message ?? "Não foi possível entrar no sistema."
    );
  }

  return "Não foi possível entrar no sistema.";
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
      email: "",
      password: "",
    },
    resolver: zodResolver(loginSchema),
  });

  const onSubmit = async (values: LoginFormValues) => {
    try {
      setSubmitError(null);
      await login({
        ...values,
        email: values.email.trim(),
      });

      navigate("/budgets", { replace: true });
    } catch (error) {
      setSubmitError(getLoginErrorMessage(error));
    }
  };

  return (
    <Box
      sx={{
        alignItems: "center",
        background:
          "radial-gradient(circle at top left, rgba(37, 99, 235, 0.2), transparent 28%), radial-gradient(circle at bottom right, rgba(14, 165, 233, 0.18), transparent 26%), linear-gradient(135deg, #08111F 0%, #0F172A 45%, #F8FAFC 100%)",
        display: "flex",
        minHeight: "100vh",
        position: "relative",
        py: { md: 4, xs: 3 },
        "&::before": {
          background:
            "radial-gradient(circle, rgba(255,255,255,0.08) 0%, rgba(255,255,255,0) 72%)",
          content: '""',
          height: 420,
          left: "-8%",
          pointerEvents: "none",
          position: "absolute",
          top: "-10%",
          width: 420,
        },
        "&::after": {
          background:
            "radial-gradient(circle, rgba(37,99,235,0.2) 0%, rgba(37,99,235,0) 72%)",
          bottom: "-14%",
          content: '""',
          height: 360,
          pointerEvents: "none",
          position: "absolute",
          right: "-10%",
          width: 360,
        },
      }}
    >
      <Container maxWidth="lg" sx={{ position: "relative", zIndex: 1 }}>
        <Paper
          sx={{
            backdropFilter: "blur(16px)",
            backgroundColor: alpha("#FFFFFF", 0.92),
            border: `1px solid ${alpha("#FFFFFF", 0.2)}`,
            borderRadius: { md: 8, xs: 5 },
            boxShadow: "0 28px 80px rgba(15, 23, 42, 0.28)",
            overflow: "hidden",
            p: 0,
          }}
        >
          <Box
            sx={{
              display: "grid",
              gridTemplateColumns: { md: "1.15fr 0.85fr", xs: "1fr" },
              minHeight: { md: 680, xs: "auto" },
            }}
          >
            <Box
              sx={{
                background:
                  "linear-gradient(160deg, #08111F 0%, #0F172A 42%, #1E3A8A 100%)",
                color: "#FFFFFF",
                display: "flex",
                flexDirection: "column",
                gap: 3,
                justifyContent: "space-between",
                p: { md: 5, xs: 3 },
                position: "relative",
              }}
            >
              <Box
                sx={{
                  background:
                    "radial-gradient(circle at top, rgba(59, 130, 246, 0.28) 0%, rgba(59, 130, 246, 0) 62%)",
                  inset: 0,
                  pointerEvents: "none",
                  position: "absolute",
                }}
              />

              <Stack spacing={3} sx={{ position: "relative", zIndex: 1 }}>
                <Box
                  sx={{
                    alignItems: "center",
                    display: "flex",
                    gap: 1.5,
                    justifyContent: { md: "flex-start", xs: "center" },
                  }}
                >
                  <BrandLogo
                    imageSx={{ width: { md: 240, xs: 180 } }}
                    wrapperSx={{
                      border: `1px solid ${alpha("#FFFFFF", 0.12)}`,
                      boxShadow: "0 24px 60px rgba(0, 0, 0, 0.28)",
                      p: 1.25,
                    }}
                  />
                </Box>

                <Stack spacing={1.5}>
                  <Typography
                    sx={{
                      color: alpha("#FFFFFF", 0.76),
                      fontSize: "0.78rem",
                      fontWeight: 700,
                      letterSpacing: "0.18em",
                      textTransform: "uppercase",
                    }}
                  >
                    Plataforma Deck
                  </Typography>
                  <Typography
                    sx={{
                      fontSize: { md: "2.8rem", xs: "2rem" },
                      fontWeight: 800,
                      letterSpacing: "-0.03em",
                      lineHeight: 1.05,
                      maxWidth: 520,
                    }}
                  >
                    Gestão de orçamentos com leitura premium e foco operacional.
                  </Typography>
                  <Typography
                    sx={{
                      color: alpha("#FFFFFF", 0.72),
                      maxWidth: 500,
                    }}
                    variant="body1"
                  >
                    Acesse um ambiente desenhado para acompanhar oportunidades,
                    organizar o funil comercial e manter a equipe alinhada com a
                    operação da Deck com mais clareza e consistência visual.
                  </Typography>
                </Stack>
              </Stack>

              <Stack
                direction={{ md: "row", xs: "column" }}
                spacing={2}
                sx={{ position: "relative", zIndex: 1 }}
              >
                <Paper
                  sx={{
                    backgroundColor: alpha("#FFFFFF", 0.06),
                    border: `1px solid ${alpha("#FFFFFF", 0.08)}`,
                    borderRadius: 4,
                    color: "#FFFFFF",
                    flex: 1,
                    p: 2.5,
                  }}
                >
                  <Stack spacing={1.5}>
                    <Box
                      sx={{
                        alignItems: "center",
                        backgroundColor: alpha("#FFFFFF", 0.08),
                        borderRadius: 2.5,
                        display: "inline-flex",
                        height: 44,
                        justifyContent: "center",
                        width: 44,
                      }}
                    >
                      <ShieldRoundedIcon sx={{ color: "#93C5FD" }} />
                    </Box>
                    <Typography sx={{ fontWeight: 700 }} variant="h5">
                      Acesso protegido
                    </Typography>
                    <Typography
                      sx={{ color: alpha("#FFFFFF", 0.7) }}
                      variant="body2"
                    >
                      Sessão segura, controle de perfil e rastreabilidade nas
                      operações mais sensíveis do sistema.
                    </Typography>
                  </Stack>
                </Paper>

                <Paper
                  sx={{
                    backgroundColor: alpha("#FFFFFF", 0.06),
                    border: `1px solid ${alpha("#FFFFFF", 0.08)}`,
                    borderRadius: 4,
                    color: "#FFFFFF",
                    flex: 1,
                    p: 2.5,
                  }}
                >
                  <Stack spacing={1.5}>
                    <Box
                      sx={{
                        alignItems: "center",
                        backgroundColor: alpha("#FFFFFF", 0.08),
                        borderRadius: 2.5,
                        display: "inline-flex",
                        height: 44,
                        justifyContent: "center",
                        width: 44,
                      }}
                    >
                      <QueryStatsRoundedIcon sx={{ color: "#93C5FD" }} />
                    </Box>
                    <Typography sx={{ fontWeight: 700 }} variant="h5">
                      Operação centralizada
                    </Typography>
                    <Typography
                      sx={{ color: alpha("#FFFFFF", 0.7) }}
                      variant="body2"
                    >
                      Consultas, acompanhamento e importação de planilhas em um
                      painel mais limpo e objetivo.
                    </Typography>
                  </Stack>
                </Paper>
              </Stack>
            </Box>

            <Box
              sx={{
                background:
                  "linear-gradient(180deg, rgba(255,255,255,0.98) 0%, rgba(248,250,252,0.98) 100%)",
                display: "flex",
                flexDirection: "column",
                justifyContent: "center",
                p: { md: 5, xs: 3 },
              }}
            >
              <Box
                component="form"
                onSubmit={handleSubmit(onSubmit)}
                sx={{ display: "flex", flexDirection: "column", gap: 3 }}
              >
                <Stack spacing={2}>
                  <Box
                    sx={{
                      alignItems: "center",
                      display: "inline-flex",
                      gap: 1.5,
                    }}
                  >
                    <Box
                      sx={{
                        alignItems: "center",
                        background:
                          "linear-gradient(135deg, #0F172A 0%, #1E3A8A 100%)",
                        borderRadius: 3,
                        color: "#FFFFFF",
                        display: "inline-flex",
                        height: 52,
                        justifyContent: "center",
                        width: 52,
                      }}
                    >
                      <LockRoundedIcon />
                    </Box>
                    <Box>
                      <Typography
                        sx={{
                          color: "primary.main",
                          fontSize: "0.78rem",
                          fontWeight: 700,
                          letterSpacing: "0.16em",
                          textTransform: "uppercase",
                        }}
                      >
                        Área segura
                      </Typography>
                      <Typography variant="h3">Entrar no painel</Typography>
                    </Box>
                  </Box>

                  <Typography color="text.secondary" variant="body1">
                    Entre com seu acesso para continuar a operação comercial da
                    Deck com mais contexto e organização.
                  </Typography>
                </Stack>

                {submitError ? (
                  <Alert
                    severity="error"
                    sx={{
                      borderRadius: 3,
                    }}
                  >
                    {submitError}
                  </Alert>
                ) : null}

                <Stack spacing={2}>
                  <TextField
                    autoComplete="username"
                    error={Boolean(errors.email)}
                    fullWidth
                    helperText={errors.email?.message}
                    label="E-mail"
                    placeholder="você@deck.com.br"
                    type="email"
                    sx={{
                      "& .MuiOutlinedInput-root": {
                        pl: 0.5,
                      },
                    }}
                    {...register("email")}
                  />
                  <TextField
                    autoComplete="current-password"
                    error={Boolean(errors.password)}
                    fullWidth
                    helperText={errors.password?.message}
                    label="Senha"
                    placeholder="Digite sua senha"
                    type="password"
                    sx={{
                      "& .MuiOutlinedInput-root": {
                        pl: 0.5,
                      },
                    }}
                    {...register("password")}
                  />

                  <Button
                    disabled={isSubmitting}
                    endIcon={<ArrowForwardRoundedIcon />}
                    size="large"
                    sx={{
                      background:
                        "linear-gradient(135deg, #0F172A 0%, #1E3A8A 100%)",
                      boxShadow: "0 14px 28px rgba(30, 58, 138, 0.24)",
                      minHeight: 52,
                    }}
                    type="submit"
                    variant="contained"
                  >
                    {isSubmitting ? "Entrando..." : "Acessar sistema"}
                  </Button>
                </Stack>

                <Divider />

                <Stack
                  direction={{ sm: "row", xs: "column" }}
                  spacing={1.5}
                  sx={{ justifyContent: "space-between" }}
                >
                  <Typography color="text.secondary" variant="body2">
                    Ambiente preparado para acompanhamento comercial e
                    orçamentos.
                  </Typography>
                  <Typography color="text.secondary" variant="body2">
                    Quer explorar antes?{" "}
                    <Box
                      component={RouterLink}
                      sx={{
                        color: "primary.main",
                        display: "inline",
                        fontWeight: 700,
                        textDecoration: "none",
                      }}
                      to="/dashboard"
                    >
                      Ver estrutura inicial
                    </Box>
                  </Typography>
                </Stack>
              </Box>
            </Box>
          </Box>
        </Paper>
      </Container>
    </Box>
  );
}
