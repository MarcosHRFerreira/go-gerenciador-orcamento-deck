import ArrowBackRoundedIcon from "@mui/icons-material/ArrowBackRounded";
import SaveRoundedIcon from "@mui/icons-material/SaveRounded";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  Alert,
  Autocomplete,
  Box,
  Button,
  CircularProgress,
  MenuItem,
  TextField,
  Typography,
} from "@mui/material";
import { alpha } from "@mui/material/styles";
import { useQuery } from "@tanstack/react-query";
import { isAxiosError } from "axios";
import { useEffect, useMemo, useState } from "react";
import { Controller, useForm } from "react-hook-form";
import { PageHeader } from "../../../components/common/PageHeader";
import { SectionCard } from "../../../components/common/SectionCard";
import { useAuth } from "../../auth/hooks/useAuth";
import {
  getBudgetCatalogsRequest,
  getBudgetListCatalogsRequest,
} from "../api/budgets";
import { listProjectsRequest } from "../../projects/api/projects";
import type { BudgetCreatePayload } from "../types/budget";
import type { BudgetFormValues } from "./budgetFormValues";
import { z as schema } from "zod";
import type { ReactNode } from "react";

const budgetGridBlue = "#1E3A8A";
const budgetFormSectionCardSx = {
  background: (theme: {
    palette: {
      info: { main: string };
      primary: { main: string };
    };
  }) =>
    `linear-gradient(135deg, ${alpha(theme.palette.primary.main, 0.07)} 0%, ${alpha(theme.palette.info.main, 0.035)} 100%)`,
  border: "1px solid",
  borderColor: (theme: { palette: { primary: { main: string } } }) =>
    alpha(theme.palette.primary.main, 0.16),
  boxShadow: (theme: { palette: { primary: { main: string } } }) =>
    `0 12px 24px ${alpha(theme.palette.primary.main, 0.07)}`,
  "& .MuiTypography-h5": {
    color: budgetGridBlue,
    fontWeight: 800,
  },
  "& .MuiTypography-body2": {
    color: "text.primary",
  },
};

const budgetFieldContainerSx = {
  display: "grid",
  gap: 0.75,
  minWidth: 0,
};

const budgetFieldLabelSx = {
  color: budgetGridBlue,
  fontSize: "0.82rem",
  fontWeight: 700,
  lineHeight: 1.2,
};

type BudgetFieldProps = {
  children: ReactNode;
  label: string;
};

function BudgetField({ children, label }: BudgetFieldProps) {
  return (
    <Box sx={budgetFieldContainerSx}>
      <Typography sx={budgetFieldLabelSx}>{label}</Typography>
      {children}
    </Box>
  );
}

const budgetFormSchema = schema.object({
  areaM2: schema
    .string()
    .trim()
    .min(1, "Informe a área em m2")
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
  projetistaName: schema
    .string()
    .trim()
    .max(150, "O projetista deve ter no máximo 150 caracteres"),
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
  systemTypeId: schema
    .string()
    .trim()
    .refine(
      (value) => isValidOptionalPositiveInteger(value),
      "Selecione um tipo de sistema válido",
    ),
  productLineId: schema
    .string()
    .trim()
    .refine(
      (value) => isValidOptionalPositiveInteger(value),
      "Selecione uma linha de produtos válida",
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
      "Selecione uma obra válida",
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
  estimatorId: schema
    .string()
    .trim()
    .refine(
      (value) => isValidOptionalPositiveInteger(value),
      "Selecione um orçamentista válido",
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
  currentProjectId?: number | null;
  currentProjectLabel?: string | null;
  lockedProjectId?: number | null;
  lockedProjectLabel?: string | null;
  mode: "create" | "edit";
  onCancel: () => void;
  onSubmit: (payload: BudgetCreatePayload) => Promise<void>;
  submitLabel: string;
  subtitle: string;
  title: string;
};

type ProjectOption = {
  id: number;
  name: string;
};

function isValidPositiveInteger(value: string) {
  const parsedValue = Number(value);

  return Number.isInteger(parsedValue) && parsedValue > 0;
}

function isValidNonNegativeInteger(value: string) {
  const parsedValue = Number(value);

  return Number.isInteger(parsedValue) && parsedValue >= 0;
}

function normalizeDecimalString(value: string) {
  const trimmedValue = value.trim().replace(/\s+/g, "");

  if (trimmedValue.includes(",") && trimmedValue.includes(".")) {
    return trimmedValue.replaceAll(".", "").replace(",", ".");
  }

  if (trimmedValue.includes(",")) {
    return trimmedValue.replace(",", ".");
  }

  return trimmedValue;
}

function isValidPositiveNumber(value: string) {
  const parsedValue = Number(normalizeDecimalString(value));

  return Number.isFinite(parsedValue) && parsedValue > 0;
}

function isValidNonNegativeNumber(value: string) {
  const parsedValue = Number(normalizeDecimalString(value));

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
  return Number(normalizeDecimalString(value));
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

  return Number(normalizeDecimalString(value));
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
    projetistaName: values.projetistaName.trim(),
    grossValue: parseDecimal(values.grossValue),
    installerId: parseOptionalInteger(values.installerId),
    lossReasonId: parseOptionalInteger(values.lossReasonId),
    priorityId: parseOptionalInteger(values.priorityId),
    productLineId: parseOptionalInteger(values.productLineId),
    systemTypeId: parseOptionalInteger(values.systemTypeId),
    projectId: parseOptionalInteger(values.projectId),
    revision: parseInteger(values.revision),
    salespersonId: parseOptionalInteger(values.salespersonId),
    estimatorId: parseOptionalInteger(values.estimatorId),
    sentAt: new Date(values.sentAt).toISOString(),
    specificationDetails: values.specificationDetails.trim(),
    statusId: parseInteger(values.statusId),
    yearBudget: parseInteger(values.yearBudget),
  };
}

export function BudgetForm({
  backLabel = "Voltar",
  currentProjectId = null,
  currentProjectLabel = null,
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
  const { user } = useAuth();
  const isAdmin = user?.role === "admin";
  const isEstimatorUser =
    user?.role === "user" && user.user_kind === "estimator";
  const canManageBudgetAssignments = isAdmin || isEstimatorUser;
  const budgetCatalogsQuery = useQuery({
    queryKey: ["budget-catalogs"],
    queryFn: canManageBudgetAssignments
      ? getBudgetCatalogsRequest
      : getBudgetListCatalogsRequest,
    staleTime: 1000 * 60 * 5,
  });
  const projectsQuery = useQuery({
    enabled: canManageBudgetAssignments,
    queryFn: listProjectsRequest,
    queryKey: ["projects", "budget-form"],
    staleTime: 1000 * 60 * 5,
  });
  const {
    control,
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

  const projectOptions = useMemo<ProjectOption[]>(() => {
    const optionsMap = new Map<number, string>();

    const appendOption = (
      projectId: number | null,
      projectName: string | null,
    ) => {
      if (projectId === null) {
        return;
      }

      const trimmedProjectName = projectName?.trim() ?? "";
      const normalizedName =
        trimmedProjectName.length > 0
          ? trimmedProjectName
          : `Obra #${projectId}`;

      optionsMap.set(projectId, normalizedName);
    };

    (budgetCatalogsQuery.data?.projects ?? []).forEach((project) => {
      appendOption(project.id, project.name);
    });
    (projectsQuery.data ?? []).forEach((project) => {
      appendOption(project.id, project.name);
    });
    appendOption(currentProjectId, currentProjectLabel);
    appendOption(lockedProjectId, lockedProjectLabel);

    return Array.from(optionsMap.entries()).map(([id, name]) => ({
      id,
      name,
    }));
  }, [
    budgetCatalogsQuery.data?.projects,
    currentProjectId,
    currentProjectLabel,
    lockedProjectId,
    lockedProjectLabel,
    projectsQuery.data,
  ]);

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
      <Box
        sx={{
          background: (theme) =>
            `linear-gradient(135deg, ${alpha(theme.palette.primary.main, 0.08)} 0%, ${alpha(theme.palette.info.main, 0.04)} 100%)`,
          border: "1px solid",
          borderColor: (theme) => alpha(theme.palette.primary.main, 0.18),
          borderRadius: 4,
          boxShadow: (theme) =>
            `0 14px 28px ${alpha(theme.palette.primary.main, 0.08)}`,
          p: { md: 3, xs: 2.5 },
          "& .MuiTypography-h3": {
            color: budgetGridBlue,
            fontWeight: 800,
          },
          "& .MuiTypography-body1": {
            color: "text.primary",
            lineHeight: 1.7,
          },
        }}
      >
        <PageHeader
          action={
            <Button
              onClick={onCancel}
              startIcon={<ArrowBackRoundedIcon />}
              sx={{
                borderColor: (theme) => alpha(theme.palette.primary.main, 0.28),
                color: budgetGridBlue,
                fontWeight: 700,
              }}
              variant="outlined"
            >
              {backLabel}
            </Button>
          }
          description={subtitle}
          title={title}
        />
      </Box>

      {isInitialDataLoading ? (
        <SectionCard sx={budgetFormSectionCardSx}>
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
            <Alert
              severity="error"
              sx={{
                "& .MuiAlert-message": {
                  fontWeight: 600,
                },
              }}
            >
              {initialDataError}
            </Alert>
          ) : null}

          {budgetCatalogsQuery.isError ? (
            <Alert
              severity="error"
              sx={{
                "& .MuiAlert-message": {
                  fontWeight: 600,
                },
              }}
            >
              Não foi possível carregar os catálogos necessários para o
              formulário.
            </Alert>
          ) : null}

          {submitError ? (
            <Alert
              severity="error"
              sx={{
                "& .MuiAlert-message": {
                  fontWeight: 600,
                },
              }}
            >
              {submitError}
            </Alert>
          ) : null}
          <SectionCard
            description="Campos principais para identificação e envio do orçamento."
            sx={budgetFormSectionCardSx}
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
              <BudgetField label="Número do orçamento">
                <TextField
                  error={Boolean(errors.budgetNumber)}
                  helperText={
                    errors.budgetNumber?.message ??
                    (mode === "edit" ? "Campo protegido na edição." : undefined)
                  }
                  placeholder="Ex: BGT-2026-004"
                  slotProps={{
                    input: {
                      readOnly: mode === "edit",
                    },
                  }}
                  {...register("budgetNumber")}
                />
              </BudgetField>
              <BudgetField label="Ano">
                <TextField
                  error={Boolean(errors.yearBudget)}
                  helperText={errors.yearBudget?.message}
                  type="number"
                  {...register("yearBudget")}
                />
              </BudgetField>
              <BudgetField label="Revisão">
                <TextField
                  error={Boolean(errors.revision)}
                  helperText={errors.revision?.message}
                  type="number"
                  {...register("revision")}
                />
              </BudgetField>
              <BudgetField label="Data de envio">
                <TextField
                  error={Boolean(errors.sentAt)}
                  helperText={errors.sentAt?.message}
                  type="datetime-local"
                  {...register("sentAt")}
                />
              </BudgetField>
              <BudgetField label="Status">
                <Controller
                  control={control}
                  name="statusId"
                  render={({ field }) => (
                    <TextField
                      error={Boolean(errors.statusId)}
                      helperText={errors.statusId?.message}
                      select
                      {...field}
                      value={field.value ?? ""}
                    >
                      <MenuItem value="">Selecione</MenuItem>
                      {(budgetCatalogsQuery.data?.statuses ?? []).map(
                        (status) => (
                          <MenuItem key={status.id} value={String(status.id)}>
                            {status.name}
                          </MenuItem>
                        ),
                      )}
                    </TextField>
                  )}
                />
              </BudgetField>
            </Box>
          </SectionCard>

          <SectionCard
            description="Valores, área e dados comerciais do orçamento."
            sx={budgetFormSectionCardSx}
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
              <BudgetField label="Valor bruto">
                <TextField
                  error={Boolean(errors.grossValue)}
                  helperText={errors.grossValue?.message}
                  slotProps={{ htmlInput: { inputMode: "decimal" } }}
                  placeholder="0,00"
                  type="text"
                  {...register("grossValue")}
                />
              </BudgetField>
              <BudgetField label="Comissão">
                <TextField
                  error={Boolean(errors.commissionValue)}
                  helperText={errors.commissionValue?.message}
                  slotProps={{ htmlInput: { inputMode: "decimal" } }}
                  placeholder="0,00"
                  type="text"
                  {...register("commissionValue")}
                />
              </BudgetField>
              <BudgetField label="Área m2">
                <TextField
                  error={Boolean(errors.areaM2)}
                  helperText={errors.areaM2?.message}
                  slotProps={{ htmlInput: { inputMode: "decimal" } }}
                  placeholder="0,00"
                  type="text"
                  {...register("areaM2")}
                />
              </BudgetField>
              <BudgetField label="Projetista">
                <TextField
                  error={Boolean(errors.projetistaName)}
                  helperText={errors.projetistaName?.message}
                  {...register("projetistaName")}
                />
              </BudgetField>
              <BudgetField label="Concorrente">
                <TextField
                  error={Boolean(errors.competitorName)}
                  helperText={errors.competitorName?.message}
                  {...register("competitorName")}
                />
              </BudgetField>
              <BudgetField label="Preço concorrente">
                <TextField
                  error={Boolean(errors.competitorPrice)}
                  helperText={errors.competitorPrice?.message}
                  slotProps={{ htmlInput: { inputMode: "decimal" } }}
                  placeholder="0,00"
                  type="text"
                  {...register("competitorPrice")}
                />
              </BudgetField>
            </Box>
          </SectionCard>

          <SectionCard
            description="Vincule o orçamento aos cadastros auxiliares disponíveis."
            sx={budgetFormSectionCardSx}
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
              <BudgetField label="Prioridade">
                <TextField
                  error={Boolean(errors.priorityId)}
                  helperText={errors.priorityId?.message}
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
              </BudgetField>
              <BudgetField label="Instalador">
                <TextField
                  error={Boolean(errors.installerId)}
                  helperText={errors.installerId?.message}
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
              </BudgetField>
              <BudgetField label="Linha de produtos">
                <TextField
                  error={Boolean(errors.productLineId)}
                  helperText={errors.productLineId?.message}
                  select
                  {...register("productLineId")}
                >
                  <MenuItem value="">Não informar</MenuItem>
                  {(budgetCatalogsQuery.data?.productLines ?? []).map(
                    (productLine) => (
                      <MenuItem
                        key={productLine.id}
                        value={String(productLine.id)}
                      >
                        {productLine.name}
                      </MenuItem>
                    ),
                  )}
                </TextField>
              </BudgetField>
              <BudgetField label="Tipo de Sistema">
                <TextField
                  error={Boolean(errors.systemTypeId)}
                  helperText={errors.systemTypeId?.message}
                  select
                  {...register("systemTypeId")}
                >
                  <MenuItem value="">Não informar</MenuItem>
                  {(budgetCatalogsQuery.data?.systemTypes ?? []).map(
                    (systemType) => (
                      <MenuItem
                        key={systemType.id}
                        value={String(systemType.id)}
                      >
                        {systemType.name}
                      </MenuItem>
                    ),
                  )}
                </TextField>
              </BudgetField>
              <BudgetField label="Obra">
                <Controller
                  control={control}
                  name="projectId"
                  render={({ field }) => {
                    const selectedProject =
                      projectOptions.find(
                        (project) => String(project.id) === field.value,
                      ) ?? null;

                    return (
                      <Autocomplete<ProjectOption, false, false, false>
                        disabled={lockedProjectId !== null}
                        getOptionLabel={(option) => option.name}
                        isOptionEqualToValue={(option, value) =>
                          option.id === value.id
                        }
                        loading={projectsQuery.isLoading}
                        noOptionsText="Nenhuma obra encontrada"
                        onChange={(_, selectedOption) => {
                          field.onChange(
                            selectedOption === null
                              ? ""
                              : String(selectedOption.id),
                          );
                        }}
                        options={projectOptions}
                        renderInput={(params) => (
                          <TextField
                            {...params}
                            error={Boolean(errors.projectId)}
                            helperText={
                              errors.projectId?.message ??
                              (lockedProjectId !== null
                                ? `Obra predefinida neste fluxo: ${lockedProjectLabel ?? `#${lockedProjectId}`}.`
                                : "Digite para filtrar as obras.")
                            }
                          />
                        )}
                        value={selectedProject}
                      />
                    );
                  }}
                />
              </BudgetField>
              {canManageBudgetAssignments ? (
                <BudgetField label="Vendedor">
                  <TextField
                    error={Boolean(errors.salespersonId)}
                    helperText={errors.salespersonId?.message}
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
                </BudgetField>
              ) : null}
              {canManageBudgetAssignments ? (
                <BudgetField label="Orçamentista">
                  <TextField
                    error={Boolean(errors.estimatorId)}
                    helperText={errors.estimatorId?.message}
                    select
                    {...register("estimatorId")}
                  >
                    <MenuItem value="">Não informar</MenuItem>
                    {(budgetCatalogsQuery.data?.estimators ?? []).map(
                      (estimator) => (
                        <MenuItem
                          key={estimator.id}
                          value={String(estimator.id)}
                        >
                          {estimator.name}
                        </MenuItem>
                      ),
                    )}
                  </TextField>
                </BudgetField>
              ) : null}
              <BudgetField label="Contato">
                <TextField
                  error={Boolean(errors.contactId)}
                  helperText={errors.contactId?.message}
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
              </BudgetField>
              <BudgetField label="Motivo de perda">
                <TextField
                  error={Boolean(errors.lossReasonId)}
                  helperText={errors.lossReasonId?.message}
                  select
                  {...register("lossReasonId")}
                >
                  <MenuItem value="">Não informar</MenuItem>
                  {(budgetCatalogsQuery.data?.lossReasons ?? []).map(
                    (lossReason) => (
                      <MenuItem
                        key={lossReason.id}
                        value={String(lossReason.id)}
                      >
                        {lossReason.name}
                      </MenuItem>
                    ),
                  )}
                </TextField>
              </BudgetField>
            </Box>
          </SectionCard>

          <SectionCard
            description="Descreva detalhes técnicos e o acompanhamento atual do orçamento."
            sx={budgetFormSectionCardSx}
            title="Detalhes"
          >
            <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
              <BudgetField label="Especificações">
                <TextField
                  error={Boolean(errors.specificationDetails)}
                  helperText={errors.specificationDetails?.message}
                  minRows={4}
                  multiline
                  {...register("specificationDetails")}
                />
              </BudgetField>
              <BudgetField label="Follow-up atual">
                <TextField
                  error={Boolean(errors.currentFollowUp)}
                  helperText={errors.currentFollowUp?.message}
                  minRows={4}
                  multiline
                  {...register("currentFollowUp")}
                />
              </BudgetField>
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
              sx={{
                borderColor: (theme) => alpha(theme.palette.primary.main, 0.28),
                color: budgetGridBlue,
                fontWeight: 700,
              }}
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
              sx={{
                boxShadow: (theme) =>
                  `0 12px 24px ${alpha(theme.palette.primary.main, 0.22)}`,
                fontWeight: 700,
              }}
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
