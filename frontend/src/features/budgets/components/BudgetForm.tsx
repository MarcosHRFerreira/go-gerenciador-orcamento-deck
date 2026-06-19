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
} from "@mui/material";
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

const budgetFormSchema = schema.object({
  areaM2: schema
    .string()
    .trim()
    .min(1, "Informe a area em m2")
    .refine(
      (value) => isValidNonNegativeNumber(value),
      "Informe uma area valida",
    ),
  budgetNumber: schema
    .string()
    .trim()
    .min(1, "Informe o numero do orcamento")
    .max(50, "O numero do orcamento deve ter no maximo 50 caracteres"),
  commissionValue: schema
    .string()
    .trim()
    .min(1, "Informe a comissao")
    .refine(
      (value) => isValidNonNegativeNumber(value),
      "Informe uma comissao valida",
    ),
  competitorName: schema
    .string()
    .trim()
    .max(150, "O concorrente deve ter no maximo 150 caracteres"),
  competitorPrice: schema
    .string()
    .trim()
    .refine(
      (value) => isValidOptionalNonNegativeNumber(value),
      "Informe um preco concorrente valido",
    ),
  contactId: schema
    .string()
    .trim()
    .refine(
      (value) => isValidOptionalPositiveInteger(value),
      "Selecione um contato valido",
    ),
  currentFollowUp: schema.string().trim(),
  projetistaName: schema
    .string()
    .trim()
    .max(150, "O projetista deve ter no maximo 150 caracteres"),
  grossValue: schema
    .string()
    .trim()
    .min(1, "Informe o valor bruto")
    .refine(
      (value) => isValidPositiveNumber(value),
      "Informe um valor bruto valido",
    ),
  installerId: schema
    .string()
    .trim()
    .refine(
      (value) => isValidOptionalPositiveInteger(value),
      "Selecione um instalador valido",
    ),
  productLineId: schema
    .string()
    .trim()
    .refine(
      (value) => isValidOptionalPositiveInteger(value),
      "Selecione uma linha de produtos valida",
    ),
  lossReasonId: schema
    .string()
    .trim()
    .refine(
      (value) => isValidOptionalPositiveInteger(value),
      "Selecione um motivo de perda valido",
    ),
  priorityId: schema
    .string()
    .trim()
    .refine(
      (value) => isValidOptionalPositiveInteger(value),
      "Selecione uma prioridade valida",
    ),
  projectId: schema
    .string()
    .trim()
    .refine(
      (value) => isValidOptionalPositiveInteger(value),
      "Selecione uma obra valida",
    ),
  revision: schema
    .string()
    .trim()
    .min(1, "Informe a revisao")
    .refine(
      (value) => isValidNonNegativeInteger(value),
      "Informe uma revisao valida",
    ),
  salespersonId: schema
    .string()
    .trim()
    .refine(
      (value) => isValidOptionalPositiveInteger(value),
      "Selecione um vendedor valido",
    ),
  estimatorId: schema
    .string()
    .trim()
    .refine(
      (value) => isValidOptionalPositiveInteger(value),
      "Selecione um orçamentista valido",
    ),
  sentAt: schema
    .string()
    .trim()
    .min(1, "Informe a data de envio")
    .refine(
      (value) => isValidDateTime(value),
      "Informe uma data de envio valida",
    ),
  specificationDetails: schema.string().trim(),
  statusId: schema
    .string()
    .trim()
    .min(1, "Selecione o status")
    .refine(
      (value) => isValidPositiveInteger(value),
      "Selecione um status valido",
    ),
  yearBudget: schema
    .string()
    .trim()
    .min(1, "Informe o ano")
    .refine((value) => isValidPositiveInteger(value), "Informe um ano valido"),
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
        ? "Nao foi possivel cadastrar o orcamento."
        : "Nao foi possivel atualizar o orcamento.")
    );
  }

  return mode === "create"
    ? "Nao foi possivel cadastrar o orcamento."
    : "Nao foi possivel atualizar o orcamento.";
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
              Nao foi possivel carregar os catalogos necessarios para o
              formulario.
            </Alert>
          ) : null}

          {submitError ? <Alert severity="error">{submitError}</Alert> : null}
          <SectionCard
            description="Campos principais para identificacao e envio do orcamento."
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
                label="Numero do orcamento"
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
                label="Revisao"
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
            description="Valores, area e dados comerciais do orcamento."
            title="Informacoes comerciais"
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
                slotProps={{ htmlInput: { inputMode: "decimal" } }}
                placeholder="0,00"
                type="text"
                {...register("grossValue")}
              />
              <TextField
                error={Boolean(errors.commissionValue)}
                helperText={errors.commissionValue?.message}
                label="Comissao"
                slotProps={{ htmlInput: { inputMode: "decimal" } }}
                placeholder="0,00"
                type="text"
                {...register("commissionValue")}
              />
              <TextField
                error={Boolean(errors.areaM2)}
                helperText={errors.areaM2?.message}
                label="Area m2"
                slotProps={{ htmlInput: { inputMode: "decimal" } }}
                placeholder="0,00"
                type="text"
                {...register("areaM2")}
              />
              <TextField
                error={Boolean(errors.projetistaName)}
                helperText={errors.projetistaName?.message}
                label="Projetista"
                {...register("projetistaName")}
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
                label="Preco concorrente"
                slotProps={{ htmlInput: { inputMode: "decimal" } }}
                placeholder="0,00"
                type="text"
                {...register("competitorPrice")}
              />
            </Box>
          </SectionCard>

          <SectionCard
            description="Vincule o orcamento aos cadastros auxiliares disponiveis."
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
                <MenuItem value="">Nao informar</MenuItem>
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
                <MenuItem value="">Nao informar</MenuItem>
                {(budgetCatalogsQuery.data?.installers ?? []).map(
                  (installer) => (
                    <MenuItem key={installer.id} value={String(installer.id)}>
                      {installer.name}
                    </MenuItem>
                  ),
                )}
              </TextField>
              <TextField
                error={Boolean(errors.productLineId)}
                helperText={errors.productLineId?.message}
                label="Linha de produtos"
                select
                {...register("productLineId")}
              >
                <MenuItem value="">Nao informar</MenuItem>
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
                          label="Obra"
                        />
                      )}
                      value={selectedProject}
                    />
                  );
                }}
              />
              {canManageBudgetAssignments ? (
                <TextField
                  error={Boolean(errors.salespersonId)}
                  helperText={errors.salespersonId?.message}
                  label="Vendedor"
                  select
                  {...register("salespersonId")}
                >
                  <MenuItem value="">Nao informar</MenuItem>
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
              ) : null}
              {canManageBudgetAssignments ? (
                <TextField
                  error={Boolean(errors.estimatorId)}
                  helperText={errors.estimatorId?.message}
                  label="Orçamentista"
                  select
                  {...register("estimatorId")}
                >
                  <MenuItem value="">Nao informar</MenuItem>
                  {(budgetCatalogsQuery.data?.estimators ?? []).map(
                    (estimator) => (
                      <MenuItem key={estimator.id} value={String(estimator.id)}>
                        {estimator.name}
                      </MenuItem>
                    ),
                  )}
                </TextField>
              ) : null}
              <TextField
                error={Boolean(errors.contactId)}
                helperText={errors.contactId?.message}
                label="Contato"
                select
                {...register("contactId")}
              >
                <MenuItem value="">Nao informar</MenuItem>
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
                <MenuItem value="">Nao informar</MenuItem>
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
            description="Descreva detalhes tecnicos e o acompanhamento atual do orcamento."
            title="Detalhes"
          >
            <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
              <TextField
                error={Boolean(errors.specificationDetails)}
                helperText={errors.specificationDetails?.message}
                label="Especificacoes"
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
