import ArrowBackRoundedIcon from "@mui/icons-material/ArrowBackRounded";
import SaveRoundedIcon from "@mui/icons-material/SaveRounded";
import { zodResolver } from "@hookform/resolvers/zod";
import { Alert, Box, Button, MenuItem, TextField } from "@mui/material";
import { isAxiosError } from "axios";
import { useEffect, useMemo, useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { PageHeader } from "../../../components/common/PageHeader";
import { SectionCard } from "../../../components/common/SectionCard";
import type { UserItem } from "../../users/types/user";
import type {
  CreateEstimatorPayload,
  EstimatorItem,
  UpdateEstimatorPayload,
} from "../types/estimator";

export const estimatorFormSchema = z.object({
  code: z
    .string()
    .trim()
    .min(1, "Informe o código")
    .max(30, "O código deve ter no máximo 30 caracteres"),
  name: z
    .string()
    .trim()
    .min(3, "Informe um nome com pelo menos 3 caracteres")
    .max(150, "O nome deve ter no máximo 150 caracteres"),
  email: z
    .string()
    .trim()
    .refine(
      (value) => value === "" || z.string().email().safeParse(value).success,
      "Informe um e-mail válido",
    ),
  phone: z
    .string()
    .trim()
    .max(50, "O telefone deve ter no máximo 50 caracteres"),
  notes: z.string().trim(),
  userId: z.string().trim(),
});

export type EstimatorFormValues = z.infer<typeof estimatorFormSchema>;

export const defaultEstimatorFormValues: EstimatorFormValues = {
  code: "",
  email: "",
  name: "",
  notes: "",
  phone: "",
  userId: "",
};

type EstimatorFormProps = {
  backLabel?: string;
  codeHelpText?: string;
  codeReadOnly?: boolean;
  initialValues: EstimatorFormValues;
  linkedUsers: UserItem[];
  mode: "create" | "edit";
  onCancel: () => void;
  onSubmit: (
    payload: CreateEstimatorPayload | UpdateEstimatorPayload,
  ) => Promise<void>;
  submitLabel: string;
  subtitle: string;
  title: string;
  currentEstimator?: EstimatorItem;
};

function getSubmitErrorMessage(error: unknown) {
  if (isAxiosError<{ message?: string }>(error)) {
    return (
      error.response?.data?.message ??
      "Não foi possível salvar o orçamentista."
    );
  }

  return "Não foi possível salvar o orçamentista.";
}

function mapCreatePayload(values: EstimatorFormValues): CreateEstimatorPayload {
  return {
    code: values.code.trim(),
    email: values.email.trim(),
    name: values.name.trim(),
    notes: values.notes.trim(),
    phone: values.phone.trim(),
    userId: values.userId ? Number(values.userId) : null,
  };
}

function mapUpdatePayload(
  estimator: EstimatorItem,
  values: EstimatorFormValues,
): UpdateEstimatorPayload {
  return {
    ...mapCreatePayload(values),
    active: estimator.active,
  };
}

function getLinkedEstimatorUsers(users: UserItem[]) {
  return users
    .filter(
      (item) =>
        item.role === "user" && item.userKind === "estimator" && item.active,
    )
    .sort((firstItem, secondItem) =>
      firstItem.name.localeCompare(secondItem.name, "pt-BR"),
    );
}

export function mapEstimatorToInitialValues(
  estimator: EstimatorItem | null,
): EstimatorFormValues {
  if (estimator === null) {
    return defaultEstimatorFormValues;
  }

  return {
    code: estimator.code,
    email: estimator.email,
    name: estimator.name,
    notes: estimator.notes,
    phone: estimator.phone,
    userId: estimator.userId === null ? "" : String(estimator.userId),
  };
}

export default function EstimatorForm({
  backLabel = "Voltar",
  codeHelpText,
  codeReadOnly = false,
  currentEstimator,
  initialValues,
  linkedUsers,
  mode,
  onCancel,
  onSubmit,
  submitLabel,
  subtitle,
  title,
}: EstimatorFormProps) {
  const [submitError, setSubmitError] = useState<string | null>(null);
  const availableLinkedUsers = useMemo(
    () => getLinkedEstimatorUsers(linkedUsers),
    [linkedUsers],
  );
  const {
    formState: { errors, isSubmitting },
    handleSubmit,
    register,
    reset,
  } = useForm<EstimatorFormValues>({
    defaultValues: initialValues,
    resolver: zodResolver(estimatorFormSchema),
  });

  useEffect(() => {
    reset(initialValues);
  }, [initialValues, reset]);

  const handleFormSubmit = async (values: EstimatorFormValues) => {
    try {
      setSubmitError(null);

      if (mode === "edit") {
        if (currentEstimator === undefined) {
          throw new Error("Orçamentista não encontrado para edição.");
        }

        await onSubmit(mapUpdatePayload(currentEstimator, values));
        return;
      }

      await onSubmit(mapCreatePayload(values));
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
          description="Mantenha código, contatos, observações e vínculo de usuário alinhados com a operação."
          title="Dados do orçamentista"
        >
          {submitError ? <Alert severity="error">{submitError}</Alert> : null}

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
              autoFocus={!codeReadOnly}
              error={Boolean(errors.code)}
              fullWidth
              helperText={errors.code?.message ?? codeHelpText}
              label="Código"
              placeholder="EST-001"
              slotProps={{
                input: {
                  readOnly: codeReadOnly,
                },
              }}
              {...register("code")}
            />
            <TextField
              autoFocus={codeReadOnly}
              error={Boolean(errors.name)}
              fullWidth
              helperText={errors.name?.message}
              label="Nome"
              placeholder="Maria Souza"
              {...register("name")}
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
              error={Boolean(errors.phone)}
              fullWidth
              helperText={errors.phone?.message}
              label="Telefone"
              placeholder="(11) 99999-9999"
              {...register("phone")}
            />
            <TextField
              error={Boolean(errors.userId)}
              fullWidth
              helperText={
                errors.userId?.message ??
                "Vincule opcionalmente um usuário operacional do tipo orçamentista."
              }
              label="Usuário vinculado"
              select
              {...register("userId")}
            >
              <MenuItem value="">Não vincular</MenuItem>
              {availableLinkedUsers.map((user) => (
                <MenuItem key={user.id} value={String(user.id)}>
                  {user.name} ({user.username})
                </MenuItem>
              ))}
            </TextField>
            <TextField
              error={Boolean(errors.notes)}
              fullWidth
              helperText={errors.notes?.message}
              label="Observações"
              minRows={4}
              multiline
              sx={{ gridColumn: { md: "1 / -1", xs: "auto" } }}
              {...register("notes")}
            />
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
