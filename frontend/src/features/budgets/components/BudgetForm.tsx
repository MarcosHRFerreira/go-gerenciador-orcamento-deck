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
import { Controller, useForm, useWatch, type Path } from "react-hook-form";
import { PageHeader } from "../../../components/common/PageHeader";
import { SectionCard } from "../../../components/common/SectionCard";
import { useAuth } from "../../auth/hooks/useAuth";
import {
  getBudgetCatalogsRequest,
  getBudgetListCatalogsRequest,
} from "../api/budgets";
import { listProjectsRequest } from "../../projects/api/projects";
import type { BudgetCatalogItem, BudgetCreatePayload } from "../types/budget";
import type { BudgetFormValues } from "./budgetFormValues";
import {
  getBudgetStatusDisplayName,
  getFactorFieldLabel,
  isWonStatusLabel,
} from "../utils/businessTerms";
import {
  getPriorityLabelByGrossValue,
  parseBudgetDecimalInput,
  resolvePriorityIdByGrossValue,
} from "../utils/priorityRanges";
import { z as schema } from "zod";
import type { ReactNode } from "react";

const budgetGridBlue = "var(--app-accent-text)";
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
    .min(1, "Informe o fator")
    .refine(
      (value) => isValidNonNegativeNumber(value),
      "Informe um fator válido",
    ),
  constructionCompany: schema
    .string()
    .trim()
    .max(200, "A construtora deve ter no máximo 200 caracteres"),
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
  deliveryDate: schema
    .string()
    .trim()
    .refine(
      (value) => value === "" || isValidDate(value),
      "Informe uma data de entrega válida",
    ),
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
  sourceCompany: schema.string().trim(),
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
  currentContactId?: number | null;
  currentContactLabel?: string | null;
  currentEstimatorId?: number | null;
  currentEstimatorLabel?: string | null;
  currentInstallerId?: number | null;
  currentInstallerLabel?: string | null;
  initialDataError?: string | null;
  initialValues: BudgetFormValues;
  isInitialDataLoading?: boolean;
  currentLossReasonId?: number | null;
  currentLossReasonLabel?: string | null;
  currentProductLineId?: number | null;
  currentProductLineLabel?: string | null;
  currentProjectId?: number | null;
  currentProjectLabel?: string | null;
  currentSalespersonId?: number | null;
  currentSalespersonLabel?: string | null;
  currentSystemTypeId?: number | null;
  currentSystemTypeLabel?: string | null;
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

function mergeCatalogItemsWithCurrentValue(
  items: BudgetCatalogItem[],
  currentId: number | null,
  currentLabel: string | null,
  fallbackPrefix: string,
) {
  const optionsMap = new Map<number, string>();

  items.forEach((item) => {
    optionsMap.set(item.id, item.name);
  });

  if (currentId !== null) {
    const trimmedLabel = currentLabel?.trim() ?? "";
    optionsMap.set(
      currentId,
      trimmedLabel.length > 0
        ? trimmedLabel
        : `${fallbackPrefix} #${currentId}`,
    );
  }

  return Array.from(optionsMap.entries()).map(([id, name]) => ({
    id,
    name,
  }));
}

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

function isValidDate(value: string) {
  return /^\d{4}-\d{2}-\d{2}$/.test(value) && !Number.isNaN(Date.parse(value));
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
    constructionCompany: values.constructionCompany.trim(),
    competitorName: values.competitorName.trim(),
    competitorPrice: parseOptionalDecimal(values.competitorPrice),
    contactId: parseOptionalInteger(values.contactId),
    currentFollowUp: values.currentFollowUp.trim(),
    deliveryDate: values.deliveryDate.trim() || null,
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
  currentContactId = null,
  currentContactLabel = null,
  currentEstimatorId = null,
  currentEstimatorLabel = null,
  currentInstallerId = null,
  currentInstallerLabel = null,
  currentProjectId = null,
  currentProjectLabel = null,
  currentLossReasonId = null,
  currentLossReasonLabel = null,
  currentProductLineId = null,
  currentProductLineLabel = null,
  currentSalespersonId = null,
  currentSalespersonLabel = null,
  currentSystemTypeId = null,
  currentSystemTypeLabel = null,
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
    setValue,
    watch,
  } = useForm<BudgetFormValues>({
    defaultValues: initialValues,
    resolver: zodResolver(budgetFormSchema),
  });
  const selectedStatusId = useWatch({
    control,
    name: "statusId",
  });
  const deliveryDateValue = useWatch({
    control,
    name: "deliveryDate",
  });
  const grossValueInput = useWatch({
    control,
    name: "grossValue",
  });

  const getTextFieldBinding = (fieldName: Path<BudgetFormValues>) => {
    const fieldRegistration = register(fieldName);

    return {
      inputRef: fieldRegistration.ref,
      name: fieldRegistration.name,
      onBlur: fieldRegistration.onBlur,
      onChange: fieldRegistration.onChange,
      value: watch(fieldName) ?? "",
    };
  };

  useEffect(() => {
    reset(initialValues);
  }, [initialValues, reset]);

  const derivedPriorityLabel = useMemo(() => {
    const parsedGrossValue = parseBudgetDecimalInput(grossValueInput ?? "");
    if (parsedGrossValue === null) {
      return "";
    }

    return getPriorityLabelByGrossValue(parsedGrossValue);
  }, [grossValueInput]);

  useEffect(() => {
    const parsedGrossValue = parseBudgetDecimalInput(grossValueInput ?? "");
    if (parsedGrossValue === null) {
      setValue("priorityId", "", { shouldDirty: false, shouldValidate: false });
      return;
    }

    const resolvedPriorityId = resolvePriorityIdByGrossValue(
      parsedGrossValue,
      budgetCatalogsQuery.data?.priorities ?? [],
    );

    setValue(
      "priorityId",
      resolvedPriorityId === null ? "" : String(resolvedPriorityId),
      {
        shouldDirty: false,
        shouldValidate: false,
      },
    );
  }, [budgetCatalogsQuery.data?.priorities, grossValueInput, setValue]);

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
  const installerOptions = useMemo(
    () =>
      mergeCatalogItemsWithCurrentValue(
        budgetCatalogsQuery.data?.installers ?? [],
        currentInstallerId,
        currentInstallerLabel,
        "Instalador",
      ),
    [
      budgetCatalogsQuery.data?.installers,
      currentInstallerId,
      currentInstallerLabel,
    ],
  );
  const productLineOptions = useMemo(
    () =>
      mergeCatalogItemsWithCurrentValue(
        budgetCatalogsQuery.data?.productLines ?? [],
        currentProductLineId,
        currentProductLineLabel,
        "Linha de produtos",
      ),
    [
      budgetCatalogsQuery.data?.productLines,
      currentProductLineId,
      currentProductLineLabel,
    ],
  );
  const systemTypeOptions = useMemo(
    () =>
      mergeCatalogItemsWithCurrentValue(
        budgetCatalogsQuery.data?.systemTypes ?? [],
        currentSystemTypeId,
        currentSystemTypeLabel,
        "Tipo de Sistema",
      ),
    [
      budgetCatalogsQuery.data?.systemTypes,
      currentSystemTypeId,
      currentSystemTypeLabel,
    ],
  );
  const salespersonOptions = useMemo(
    () =>
      mergeCatalogItemsWithCurrentValue(
        budgetCatalogsQuery.data?.salespeople ?? [],
        currentSalespersonId,
        currentSalespersonLabel,
        "Vendedor",
      ),
    [
      budgetCatalogsQuery.data?.salespeople,
      currentSalespersonId,
      currentSalespersonLabel,
    ],
  );
  const estimatorOptions = useMemo(
    () =>
      mergeCatalogItemsWithCurrentValue(
        budgetCatalogsQuery.data?.estimators ?? [],
        currentEstimatorId,
        currentEstimatorLabel,
        "Orcamentista",
      ),
    [
      budgetCatalogsQuery.data?.estimators,
      currentEstimatorId,
      currentEstimatorLabel,
    ],
  );
  const contactOptions = useMemo(
    () =>
      mergeCatalogItemsWithCurrentValue(
        budgetCatalogsQuery.data?.contacts ?? [],
        currentContactId,
        currentContactLabel,
        "Contato",
      ),
    [budgetCatalogsQuery.data?.contacts, currentContactId, currentContactLabel],
  );
  const lossReasonOptions = useMemo(
    () =>
      mergeCatalogItemsWithCurrentValue(
        budgetCatalogsQuery.data?.lossReasons ?? [],
        currentLossReasonId,
        currentLossReasonLabel,
        "Motivo de perda",
      ),
    [
      budgetCatalogsQuery.data?.lossReasons,
      currentLossReasonId,
      currentLossReasonLabel,
    ],
  );

  const handleFormSubmit = async (values: BudgetFormValues) => {
    try {
      setSubmitError(null);
      await onSubmit(mapFormValuesToPayload(values));
    } catch (error) {
      setSubmitError(getBudgetSubmitErrorMessage(mode, error));
    }
  };

  const selectedStatusName = useMemo(() => {
    if (!selectedStatusId) {
      return null;
    }

    const selectedStatus = (budgetCatalogsQuery.data?.statuses ?? []).find(
      (status) => String(status.id) === selectedStatusId,
    );

    return selectedStatus?.name ?? null;
  }, [budgetCatalogsQuery.data?.statuses, selectedStatusId]);

  const isPedidoStatus =
    selectedStatusName !== null && isWonStatusLabel(selectedStatusName);
  const shouldShowWonStatusAlert =
    isPedidoStatus && initialDataError === null && !isInitialDataLoading;

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
          <input type="hidden" {...register("areaM2")} />
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

          {shouldShowWonStatusAlert ? (
            <Alert
              severity="warning"
              sx={{
                "& .MuiAlert-message": {
                  fontWeight: 600,
                },
              }}
            >
              {`Ao marcar este orçamento como Fechado, o sistema poderá cancelar automaticamente os orçamentos em aberto de outros instaladores da mesma obra. Orçamentos do mesmo instalador podem permanecer ativos por representarem escopos complementares. Um aviso informativo será enviado ao vendedor e aos administradores.${deliveryDateValue ? " A data de entrega segue recomendada para apoiar o acompanhamento operacional deste fechado." : " Se houver entrega prevista, informe a data para apoiar o acompanhamento operacional deste fechado."}`}
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
                  lg: "repeat(6, minmax(0, 1fr))",
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
                  {...getTextFieldBinding("budgetNumber")}
                />
              </BudgetField>
              <BudgetField label="Ano">
                <TextField
                  error={Boolean(errors.yearBudget)}
                  helperText={errors.yearBudget?.message}
                  type="number"
                  {...getTextFieldBinding("yearBudget")}
                />
              </BudgetField>
              <BudgetField label="Revisão">
                <TextField
                  error={Boolean(errors.revision)}
                  helperText={errors.revision?.message}
                  type="number"
                  {...getTextFieldBinding("revision")}
                />
              </BudgetField>
              <BudgetField label="Data de envio">
                <TextField
                  error={Boolean(errors.sentAt)}
                  helperText={errors.sentAt?.message}
                  type="datetime-local"
                  {...getTextFieldBinding("sentAt")}
                />
              </BudgetField>
              <BudgetField label="Empresa">
                <TextField
                  helperText={
                    mode === "edit"
                      ? "Campo exibido para conferência da origem do orçamento."
                      : "Preenchido automaticamente pela origem do orçamento."
                  }
                  slotProps={{
                    input: {
                      readOnly: true,
                    },
                  }}
                  {...getTextFieldBinding("sourceCompany")}
                />
              </BudgetField>
              <BudgetField label="Data de entrega">
                <TextField
                  error={Boolean(errors.deliveryDate)}
                  helperText={
                    errors.deliveryDate?.message ??
                    (isPedidoStatus
                      ? "Recomendado para acompanhamento operacional do fechado."
                      : undefined)
                  }
                  type="date"
                  {...getTextFieldBinding("deliveryDate")}
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
                            {getBudgetStatusDisplayName(status.name)}
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
            description="Valores e dados comerciais do orçamento."
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
                  {...getTextFieldBinding("grossValue")}
                />
              </BudgetField>
              <BudgetField label={getFactorFieldLabel()}>
                <TextField
                  error={Boolean(errors.commissionValue)}
                  helperText={errors.commissionValue?.message}
                  slotProps={{ htmlInput: { inputMode: "decimal" } }}
                  placeholder="0,00"
                  type="text"
                  {...getTextFieldBinding("commissionValue")}
                />
              </BudgetField>
              <BudgetField label="Construtora">
                <TextField
                  error={Boolean(errors.constructionCompany)}
                  helperText={errors.constructionCompany?.message}
                  placeholder="Nome da construtora"
                  {...getTextFieldBinding("constructionCompany")}
                />
              </BudgetField>
              <BudgetField label="Projetista">
                <TextField
                  error={Boolean(errors.projetistaName)}
                  helperText={errors.projetistaName?.message}
                  {...getTextFieldBinding("projetistaName")}
                />
              </BudgetField>
              <BudgetField label="Concorrente">
                <TextField
                  error={Boolean(errors.competitorName)}
                  helperText={errors.competitorName?.message}
                  {...getTextFieldBinding("competitorName")}
                />
              </BudgetField>
              <BudgetField label="Preço concorrente">
                <TextField
                  error={Boolean(errors.competitorPrice)}
                  helperText={errors.competitorPrice?.message}
                  slotProps={{ htmlInput: { inputMode: "decimal" } }}
                  placeholder="0,00"
                  type="text"
                  {...getTextFieldBinding("competitorPrice")}
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
                <input type="hidden" {...register("priorityId")} />
                <TextField
                  helperText={
                    derivedPriorityLabel.length > 0
                      ? "Classificada automaticamente pelo valor bruto."
                      : "Informe o valor bruto para classificar a prioridade."
                  }
                  slotProps={{
                    input: {
                      readOnly: true,
                    },
                  }}
                  value={derivedPriorityLabel}
                />
              </BudgetField>
              <BudgetField label="Instalador">
                <TextField
                  error={Boolean(errors.installerId)}
                  helperText={errors.installerId?.message}
                  select
                  {...getTextFieldBinding("installerId")}
                >
                  <MenuItem value="">Não informar</MenuItem>
                  {installerOptions.map((installer) => (
                    <MenuItem key={installer.id} value={String(installer.id)}>
                      {installer.name}
                    </MenuItem>
                  ))}
                </TextField>
              </BudgetField>
              <BudgetField label="Linha de produtos">
                <TextField
                  error={Boolean(errors.productLineId)}
                  helperText={errors.productLineId?.message}
                  select
                  {...getTextFieldBinding("productLineId")}
                >
                  <MenuItem value="">Não informar</MenuItem>
                  {productLineOptions.map((productLine) => (
                    <MenuItem
                      key={productLine.id}
                      value={String(productLine.id)}
                    >
                      {productLine.name}
                    </MenuItem>
                  ))}
                </TextField>
              </BudgetField>
              <BudgetField label="Tipo de Sistema">
                <TextField
                  error={Boolean(errors.systemTypeId)}
                  helperText={errors.systemTypeId?.message}
                  select
                  {...getTextFieldBinding("systemTypeId")}
                >
                  <MenuItem value="">Não informar</MenuItem>
                  {systemTypeOptions.map((systemType) => (
                    <MenuItem key={systemType.id} value={String(systemType.id)}>
                      {systemType.name}
                    </MenuItem>
                  ))}
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
                    {...getTextFieldBinding("salespersonId")}
                  >
                    <MenuItem value="">Não informar</MenuItem>
                    {salespersonOptions.map((salesperson) => (
                      <MenuItem
                        key={salesperson.id}
                        value={String(salesperson.id)}
                      >
                        {salesperson.name}
                      </MenuItem>
                    ))}
                  </TextField>
                </BudgetField>
              ) : null}
              {canManageBudgetAssignments ? (
                <BudgetField label="Orçamentista">
                  <TextField
                    error={Boolean(errors.estimatorId)}
                    helperText={errors.estimatorId?.message}
                    select
                    {...getTextFieldBinding("estimatorId")}
                  >
                    <MenuItem value="">Não informar</MenuItem>
                    {estimatorOptions.map((estimator) => (
                      <MenuItem key={estimator.id} value={String(estimator.id)}>
                        {estimator.name}
                      </MenuItem>
                    ))}
                  </TextField>
                </BudgetField>
              ) : null}
              <BudgetField label="Contato">
                <TextField
                  error={Boolean(errors.contactId)}
                  helperText={errors.contactId?.message}
                  select
                  {...getTextFieldBinding("contactId")}
                >
                  <MenuItem value="">Não informar</MenuItem>
                  {contactOptions.map((contact) => (
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
                  {...getTextFieldBinding("lossReasonId")}
                >
                  <MenuItem value="">Não informar</MenuItem>
                  {lossReasonOptions.map((lossReason) => (
                    <MenuItem key={lossReason.id} value={String(lossReason.id)}>
                      {lossReason.name}
                    </MenuItem>
                  ))}
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
                  {...getTextFieldBinding("specificationDetails")}
                />
              </BudgetField>
              <BudgetField label="Follow-up atual">
                <TextField
                  error={Boolean(errors.currentFollowUp)}
                  helperText={errors.currentFollowUp?.message}
                  minRows={4}
                  multiline
                  {...getTextFieldBinding("currentFollowUp")}
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
