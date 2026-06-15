import ArrowBackRoundedIcon from "@mui/icons-material/ArrowBackRounded";
import SaveRoundedIcon from "@mui/icons-material/SaveRounded";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  Alert,
  Box,
  Button,
  CircularProgress,
  MenuItem,
  TextField,
} from "@mui/material";
import { useQuery } from "@tanstack/react-query";
import { isAxiosError } from "axios";
import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { PageHeader } from "../../../components/common/PageHeader";
import { SectionCard } from "../../../components/common/SectionCard";
import { getBudgetCatalogsRequest } from "../api/budgets";
import type { BudgetCreatePayload } from "../types/budget";
import type { BudgetFormValues } from "./budgetFormValues";
import { z as schema } from "zod";

const budgetFormSchema = schema.object({
  areaM2: schema
    .string()
    .trim()
    .min(1, "Informe a área em m²")
    .refine(
      (value) => isValidNonNegativeNumber(value),
      "Informe uma área válida",
    ),
  budgetNumber: schema
    .string()
    .trim()
    .min(1, "Informe o número do orçamento")
    .max(50, "O número do orçamento deve ter no máximo 50 caracteres"),
  commissionValue: schema
    .string()
    .trim()
    .min(1, "Informe a comissão")
    .refine(
      (value) => isValidNonNegativeNumber(value),
      "Informe uma comissão válida",
    ),
  competitorName: schema
    .string()
    .trim()
    .max(150, "O concorrente deve ter no máximo 150 caracteres"),
  competitorPrice: schema
    .string()
    .trim()
    .refine(
      (value) => isValidOptionalNonNegativeNumber(value),
      "Informe um preço concorrente válido",
    ),
  contactId: schema
    .string()
    .trim()
    .refine(
      (value) => isValidOptionalPositiveInteger(value),
      "Selecione um contato válido",
    ),
  currentFollowUp: schema.string().trim(),
  designerName: schema
    .string()
    .trim()
    .max(150, "O designer deve ter no máximo 150 caracteres"),
  grossValue: schema
    .string()
    .trim()
    .min(1, "Informe o valor bruto")
    .refine(
      (value) => isValidPositiveNumber(value),
      "Informe um valor bruto válido",
    ),
  installerId: schema
    .string()
    .trim()
    .refine(
      (value) => isValidOptionalPositiveInteger(value),
      "Selecione um instalador válido",
    ),
  lossReasonId: schema
    .string()
    .trim()
    .refine(
      (value) => isValidOptionalPositiveInteger(value),
      "Selecione um motivo de perda válido",
    ),
  priorityId: schema
    .string()
    .trim()
    .refine(
      (value) => isValidOptionalPositiveInteger(value),
      "Selecione uma prioridade válida",
    ),
  projectId: schema
    .string()
    .trim()
    .refine(
      (value) => isValidOptionalPositiveInteger(value),
      "Selecione um projeto válido",
    ),
  revision: schema
    .string()
    .trim()
    .min(1, "Informe a revisão")
    .refine(
      (value) => isValidNonNegativeInteger(value),
      "Informe uma revisão válida",
    ),
  salespersonId: schema
    .string()
    .trim()
    .refine(
      (value) => isValidOptionalPositiveInteger(value),
      "Selecione um vendedor válido",
    ),
  sentAt: schema
    .string()
    .trim()
    .min(1, "Informe a data de envio")
    .refine(
      (value) => isValidDateTime(value),
      "Informe uma data de envio válida",
    ),
  specificationDetails: schema.string().trim(),
  statusId: schema
    .string()
    .trim()
    .min(1, "Selecione o status")
    .refine(
      (value) => isValidPositiveInteger(value),
      "Selecione um status válido",
    ),
  yearBudget: schema
    .string()
    .trim()
    .min(1, "Informe o ano")
    .refine((value) => isValidPositiveInteger(value), "Informe um ano válido"),
});

type BudgetFormProps = {
  backLabel?: string;
  initialDataError?: string | null;
  initialValues: BudgetFormValues;
  isInitialDataLoading?: boolean;
  lockedProjectId?: number | null;
  lockedProjectLabel?: string | null;
  mode: "create" | "edit";
  onCancel: () => void;
  onSubmit: (payload: BudgetCreatePayload) => Promise<void>;
  submitLabel: string;
  subtitle: string;
  title: string;
};

function isValidPositiveInteger(value: string) {
  const parsedValue = Number(value);

  return Number.isInteger(parsedValue) && parsedValue > 0;
}

function isValidNonNegativeInteger(value: string) {
  const parsedValue = Number(value);

  return Number.isInteger(parsedValue) && parsedValue >= 0;
}

function isValidPositiveNumber(value: string) {
  const parsedValue = Number(value);

  return Number.isFinite(parsedValue) && parsedValue > 0;
}

function isValidNonNegativeNumber(value: string) {
  const parsedValue = Number(value);

  return Number.isFinite(parsedValue) && parsedValue >= 0;
}

function isValidOptionalPositiveInteger(value: string) {
  return value === "" || isValidPositiveInteger(value);
}

function isValidOptionalNonNegativeNumber(value: string) {
  return value === "" || isValidNonNegativeNumber(value);
}

function isValidDateTime(value: string) {
  return !Number.isNaN(new Date(value).getTime());
}

function parseInteger(value: string) {
  return Number.parseInt(value, 10);
}

function parseDecimal(value: string) {
  return Number(value);
}

function parseOptionalInteger(value: string) {
  if (!value) {
    return null;
  }

  return Number.parseInt(value, 10);
}

function parseOptionalDecimal(value: string) {
  if (!value) {
    return null;
  }

  return Number(value);
}

function getBudgetSubmitErrorMessage(mode: "create" | "edit", error: unknown) {
  if (isAxiosError<{ message?: string }>(error)) {
    return (
      error.response?.data?.message ??
      (mode === "create"
        ? "Não foi possível cadastrar o orçamento."
        : "Não foi possível atualizar o orçamento.")
    );
  }

  return mode === "create"
    ? "Não foi possível cadastrar o orçamento."
    : "Não foi possível atualizar o orçamento.";
}

function mapFormValuesToPayload(values: BudgetFormValues): BudgetCreatePayload {
  return {
    areaM2: parseDecimal(values.areaM2),
    budgetNumber: values.budgetNumber.trim(),
    commissionValue: parseDecimal(values.commissionValue),
    competitorName: values.competitorName.trim(),
    competitorPrice: parseOptionalDecimal(values.competitorPrice),
    contactId: parseOptionalInteger(values.contactId),
    currentFollowUp: values.currentFollowUp.trim(),
    designerName: values.designerName.trim(),
    grossValue: parseDecimal(values.grossValue),
    installerId: parseOptionalInteger(values.installerId),
    lossReasonId: parseOptionalInteger(values.lossReasonId),
    priorityId: parseOptionalInteger(values.priorityId),
    projectId: parseOptionalInteger(values.projectId),
    revision: parseInteger(values.revision),
    salespersonId: parseOptionalInteger(values.salespersonId),
    sentAt: new Date(values.sentAt).toISOString(),
    specificationDetails: values.specificationDetails.trim(),
    statusId: parseInteger(values.statusId),
    yearBudget: parseInteger(values.yearBudget),
  };
}

export function BudgetForm({
  backLabel = "Voltar",
  initialDataError = null,
  initialValues,
  isInitialDataLoading = false,
  lockedProjectId = null,
  lockedProjectLabel = null,
  mode,
  onCancel,
  onSubmit,
  submitLabel,
  subtitle,
  title,
}: BudgetFormProps) {
  const [submitError, setSubmitError] = useState<string | null>(null);
  const budgetCatalogsQuery = useQuery({
    queryKey: ["budget-catalogs"],
    queryFn: getBudgetCatalogsRequest,
    staleTime: 1000 * 60 * 5,
  });
  const {
    formState: { errors, isSubmitting },
    handleSubmit,
    register,
    reset,
  } = useForm<BudgetFormValues>({
    defaultValues: initialValues,
    resolver: zodResolver(budgetFormSchema),
  });

  useEffect(() => {
    reset(initialValues);
  }, [initialValues, reset]);

  const handleFormSubmit = async (values: BudgetFormValues) => {
    try {
      setSubmitError(null);
      await onSubmit(mapFormValuesToPayload(values));
    } catch (error) {
      setSubmitError(getBudgetSubmitErrorMessage(mode, error));
    }
  };

  return (
    <Box sx={{ display: "flex", flexDirection: "column", gap: 3 }}>
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

      {isInitialDataLoading ? (
        <SectionCard>
          <Box
            sx={{
              alignItems: "center",
              display: "flex",
              justifyContent: "center",
              minHeight: 240,
            }}
          >
            <CircularProgress />
          </Box>
        </SectionCard>
      ) : null}

      {!isInitialDataLoading ? (
        <Box
          component="form"
          onSubmit={handleSubmit(handleFormSubmit)}
          sx={{ display: "flex", flexDirection: "column", gap: 3 }}
        >
          {initialDataError ? (
            <Alert severity="error">{initialDataError}</Alert>
          ) : null}

          {budgetCatalogsQuery.isError ? (
            <Alert severity="error">
              Não foi possível carregar os catálogos necessários para o
              formulário.
            </Alert>
          ) : null}

          {submitError ? <Alert severity="error">{submitError}</Alert> : null}

          <SectionCard
            description="Campos principais para identificação e envio do orçamento."
            title="Dados principais"
          >
            <Box
              sx={{
                display: "grid",
                gap: 2,
                gridTemplateColumns: {
                  lg: "repeat(5, minmax(0, 1fr))",
                  md: "repeat(2, minmax(0, 1fr))",
                  xs: "minmax(0, 1fr)",
                },
              }}
            >
              <TextField
                error={Boolean(errors.budgetNumber)}
                helperText={errors.budgetNumber?.message}
                label="Número do orçamento"
                placeholder="Ex: BGT-2026-004"
                {...register("budgetNumber")}
              />
              <TextField
                error={Boolean(errors.yearBudget)}
                helperText={errors.yearBudget?.message}
                label="Ano"
                type="number"
                {...register("yearBudget")}
              />
              <TextField
                error={Boolean(errors.revision)}
                helperText={errors.revision?.message}
                label="Revisão"
                type="number"
                {...register("revision")}
              />
              <TextField
                error={Boolean(errors.sentAt)}
                helperText={errors.sentAt?.message}
                label="Data de envio"
                slotProps={{ inputLabel: { shrink: true } }}
                type="datetime-local"
                {...register("sentAt")}
              />
              <TextField
                error={Boolean(errors.statusId)}
                helperText={errors.statusId?.message}
                label="Status"
                select
                {...register("statusId")}
              >
                <MenuItem value="">Selecione</MenuItem>
                {(budgetCatalogsQuery.data?.statuses ?? []).map((status) => (
                  <MenuItem key={status.id} value={String(status.id)}>
                    {status.name}
                  </MenuItem>
                ))}
              </TextField>
            </Box>
          </SectionCard>

          <SectionCard
            description="Valores, área e dados comerciais do orçamento."
            title="Informações comerciais"
          >
            <Box
              sx={{
                display: "grid",
                gap: 2,
                gridTemplateColumns: {
                  lg: "repeat(3, minmax(0, 1fr))",
                  md: "repeat(2, minmax(0, 1fr))",
                  xs: "minmax(0, 1fr)",
                },
              }}
            >
              <TextField
                error={Boolean(errors.grossValue)}
                helperText={errors.grossValue?.message}
                label="Valor bruto"
                placeholder="0,00"
                type="number"
                {...register("grossValue")}
              />
              <TextField
                error={Boolean(errors.commissionValue)}
                helperText={errors.commissionValue?.message}
                label="Comissão"
                placeholder="0,00"
                type="number"
                {...register("commissionValue")}
              />
              <TextField
                error={Boolean(errors.areaM2)}
                helperText={errors.areaM2?.message}
                label="Área m²"
                placeholder="0,00"
                type="number"
                {...register("areaM2")}
              />
              <TextField
                error={Boolean(errors.designerName)}
                helperText={errors.designerName?.message}
                label="Designer"
                {...register("designerName")}
              />
              <TextField
                error={Boolean(errors.competitorName)}
                helperText={errors.competitorName?.message}
                label="Concorrente"
                {...register("competitorName")}
              />
              <TextField
                error={Boolean(errors.competitorPrice)}
                helperText={errors.competitorPrice?.message}
                label="Preço concorrente"
                placeholder="0,00"
                type="number"
                {...register("competitorPrice")}
              />
            </Box>
          </SectionCard>

          <SectionCard
            description="Vincule o orçamento aos cadastros auxiliares disponíveis."
            title="Relacionamentos"
          >
            <Box
              sx={{
                display: "grid",
                gap: 2,
                gridTemplateColumns: {
                  lg: "repeat(3, minmax(0, 1fr))",
                  md: "repeat(2, minmax(0, 1fr))",
                  xs: "minmax(0, 1fr)",
                },
              }}
            >
              <TextField
                error={Boolean(errors.priorityId)}
                helperText={errors.priorityId?.message}
                label="Prioridade"
                select
                {...register("priorityId")}
              >
                <MenuItem value="">Não informar</MenuItem>
                {(budgetCatalogsQuery.data?.priorities ?? []).map(
                  (priority) => (
                    <MenuItem key={priority.id} value={String(priority.id)}>
                      {priority.name}
                    </MenuItem>
                  ),
                )}
              </TextField>
              <TextField
                error={Boolean(errors.installerId)}
                helperText={errors.installerId?.message}
                label="Instalador"
                select
                {...register("installerId")}
              >
                <MenuItem value="">Não informar</MenuItem>
                {(budgetCatalogsQuery.data?.installers ?? []).map(
                  (installer) => (
                    <MenuItem key={installer.id} value={String(installer.id)}>
                      {installer.name}
                    </MenuItem>
                  ),
                )}
              </TextField>
              <TextField
                disabled={lockedProjectId !== null}
                error={Boolean(errors.projectId)}
                helperText={
                  errors.projectId?.message ??
                  (lockedProjectId !== null
                    ? `Projeto predefinido neste fluxo: ${lockedProjectLabel ?? `#${lockedProjectId}`}.`
                    : undefined)
                }
                label="Projeto"
                select
                {...register("projectId")}
              >
                <MenuItem value="">Não informar</MenuItem>
                {(budgetCatalogsQuery.data?.projects ?? []).map((project) => (
                  <MenuItem key={project.id} value={String(project.id)}>
                    {project.name}
                  </MenuItem>
                ))}
              </TextField>
              <TextField
                error={Boolean(errors.salespersonId)}
                helperText={errors.salespersonId?.message}
                label="Vendedor"
                select
                {...register("salespersonId")}
              >
                <MenuItem value="">Não informar</MenuItem>
                {(budgetCatalogsQuery.data?.salespeople ?? []).map(
                  (salesperson) => (
                    <MenuItem
                      key={salesperson.id}
                      value={String(salesperson.id)}
                    >
                      {salesperson.name}
                    </MenuItem>
                  ),
                )}
              </TextField>
              <TextField
                error={Boolean(errors.contactId)}
                helperText={errors.contactId?.message}
                label="Contato"
                select
                {...register("contactId")}
              >
                <MenuItem value="">Não informar</MenuItem>
                {(budgetCatalogsQuery.data?.contacts ?? []).map((contact) => (
                  <MenuItem key={contact.id} value={String(contact.id)}>
                    {contact.name}
                  </MenuItem>
                ))}
              </TextField>
              <TextField
                error={Boolean(errors.lossReasonId)}
                helperText={errors.lossReasonId?.message}
                label="Motivo de perda"
                select
                {...register("lossReasonId")}
              >
                <MenuItem value="">Não informar</MenuItem>
                {(budgetCatalogsQuery.data?.lossReasons ?? []).map(
                  (lossReason) => (
                    <MenuItem key={lossReason.id} value={String(lossReason.id)}>
                      {lossReason.name}
                    </MenuItem>
                  ),
                )}
              </TextField>
            </Box>
          </SectionCard>

          <SectionCard
            description="Descreva detalhes técnicos e o acompanhamento atual do orçamento."
            title="Detalhes"
          >
            <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
              <TextField
                error={Boolean(errors.specificationDetails)}
                helperText={errors.specificationDetails?.message}
                label="Especificações"
                minRows={4}
                multiline
                {...register("specificationDetails")}
              />
              <TextField
                error={Boolean(errors.currentFollowUp)}
                helperText={errors.currentFollowUp?.message}
                label="Follow-up atual"
                minRows={4}
                multiline
                {...register("currentFollowUp")}
              />
            </Box>
          </SectionCard>

          <Box
            sx={{
              display: "flex",
              gap: 2,
              justifyContent: "flex-end",
              pb: 1,
            }}
          >
            <Button
              disabled={isSubmitting}
              onClick={onCancel}
              variant="outlined"
            >
              Cancelar
            </Button>
            <Button
              disabled={
                isSubmitting ||
                budgetCatalogsQuery.isError ||
                initialDataError !== null
              }
              startIcon={<SaveRoundedIcon />}
              type="submit"
              variant="contained"
            >
              {isSubmitting ? "Salvando..." : submitLabel}
            </Button>
          </Box>
        </Box>
      ) : null}
    </Box>
  );
}
