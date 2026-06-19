import UploadFileRoundedIcon from "@mui/icons-material/UploadFileRounded";
import VisibilityRoundedIcon from "@mui/icons-material/VisibilityRounded";
import TaskAltRoundedIcon from "@mui/icons-material/TaskAltRounded";
import {
  Alert,
  Box,
  Button,
  Checkbox,
  Chip,
  FormControlLabel,
  LinearProgress,
  MenuItem,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  TextField,
  Typography,
} from "@mui/material";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { isAxiosError } from "axios";
import { useEffect, useMemo, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { PageHeader } from "../../../components/common/PageHeader";
import { SectionCard } from "../../../components/common/SectionCard";
import {
  executeBudgetImportRequest,
  getBudgetImportStatusRequest,
  previewBudgetImportRequest,
} from "../api/budgets";
import type {
  BudgetImportExecutionResult,
  BudgetImportPreviewOptions,
  BudgetImportPreviewResult,
} from "../types/budget";

function getImportErrorMessage(error: unknown, fallbackMessage: string) {
  if (isAxiosError<{ message?: string }>(error)) {
    if (error.response?.status === 401 || error.response?.status === 403) {
      return "Sua sessão expirou. Entre novamente no sistema.";
    }

    if (
      error.response?.status === 404 &&
      error.response?.data?.message === "Preview nao encontrado"
    ) {
      return "Este preview ja foi utilizado ou expirou. Gere um novo preview antes de importar novamente.";
    }

    if (!error.response && error.message === "Network Error") {
      return "Sua sessão expirou. Entre novamente no sistema.";
    }

    return error.response?.data?.message ?? fallbackMessage;
  }

  return fallbackMessage;
}

function isPreviewUnavailableError(error: unknown) {
  return (
    isAxiosError<{ message?: string }>(error) &&
    error.response?.status === 404 &&
    error.response?.data?.message === "Preview nao encontrado"
  );
}

function formatDateTime(value: string) {
  return new Intl.DateTimeFormat("pt-BR", {
    dateStyle: "short",
    timeStyle: "short",
  }).format(new Date(value));
}

function createSummaryItems(previewResult: BudgetImportPreviewResult) {
  return [
    { label: "Linhas lidas", value: previewResult.summary.rowsRead },
    { label: "Linhas válidas", value: previewResult.summary.rowsValid },
    {
      label: "Com alerta",
      value: previewResult.summary.rowsWithWarning,
    },
    { label: "Com erro", value: previewResult.summary.rowsWithError },
    { label: "Novos orçamentos", value: previewResult.summary.newBudgets },
    {
      label: "Orçamentos existentes",
      value: previewResult.summary.existingBudgets,
    },
  ];
}

function createCatalogItems(previewResult: BudgetImportPreviewResult) {
  return [
    {
      label: "Status",
      value: previewResult.catalogActions.budgetStatusesToCreate,
    },
    {
      label: "Prioridades",
      value: previewResult.catalogActions.prioritiesToCreate,
    },
    {
      label: "Instaladores",
      value: previewResult.catalogActions.installersToCreate,
    },
    {
      label: "Linhas de produtos",
      value: previewResult.catalogActions.productLinesToCreate,
    },
    {
      label: "Obras",
      value: previewResult.catalogActions.projectsToCreate,
    },
    {
      label: "Tipos de obra",
      value: previewResult.catalogActions.projectTypesToCreate,
    },
    {
      label: "Vendedores",
      value: previewResult.catalogActions.salespeopleToCreate,
    },
    {
      label: "Contatos",
      value: previewResult.catalogActions.contactsToCreate,
    },
    {
      label: "Motivos de perda",
      value: previewResult.catalogActions.lossReasonsToCreate,
    },
  ];
}

function getExecutionSuccessMessage(
  fileName: string,
  executionResult: BudgetImportExecutionResult,
) {
  const { budgetsCreated, budgetsIgnored, budgetsUpdated } =
    executionResult.summary;

  return `Arquivo ${fileName} importado com sucesso. ${budgetsCreated} orçamento(s) criado(s), ${budgetsUpdated} atualizado(s) e ${budgetsIgnored} ignorado(s).`;
}

function isImportProcessing(
  executionResult: BudgetImportExecutionResult | null,
) {
  return executionResult?.status === "processing";
}

function getImportProgressValue(
  executionResult: BudgetImportExecutionResult | null,
) {
  if (executionResult === null) {
    return null;
  }

  const { rowsExpected, rowsProcessed } = executionResult.summary;
  if (rowsExpected <= 0) {
    return null;
  }

  return Math.min(100, Math.round((rowsProcessed / rowsExpected) * 100));
}

export function BudgetImportPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const executingPreviewIdRef = useRef<string | null>(null);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [previewOptions, setPreviewOptions] =
    useState<BudgetImportPreviewOptions>({
      duplicateStrategy: "ignore",
      createMissingCatalogs: true,
      useDefaultNotInformed: true,
    });
  const [previewResult, setPreviewResult] =
    useState<BudgetImportPreviewResult | null>(null);
  const [executionResult, setExecutionResult] =
    useState<BudgetImportExecutionResult | null>(null);
  const [activeImportId, setActiveImportId] = useState<string | null>(null);
  const [processInfo, setProcessInfo] = useState<string | null>(null);
  const [previewError, setPreviewError] = useState<string | null>(null);
  const [executionError, setExecutionError] = useState<string | null>(null);

  const requestPreview = async () => {
    if (selectedFile === null) {
      throw new Error("Selecione um arquivo .xlsx para gerar o preview.");
    }

    return previewBudgetImportRequest(selectedFile, previewOptions);
  };

  const requestExecuteImport = async (previewId: string) => {
    return executeBudgetImportRequest({
      previewId,
    });
  };

  const previewMutation = useMutation({
    mutationFn: requestPreview,
    onSuccess: (result) => {
      executingPreviewIdRef.current = null;
      setPreviewResult(result);
      setExecutionResult(null);
      setActiveImportId(null);
      setProcessInfo(null);
      setPreviewError(null);
      setExecutionError(null);
    },
  });

  const executeMutation = useMutation({
    mutationFn: async ({ previewId }: { previewId: string }) =>
      requestExecuteImport(previewId),
    onSuccess: (result) => {
      executingPreviewIdRef.current = result.previewId;
      setExecutionResult(result);
      setActiveImportId(result.importId);
      setProcessInfo(
        "Importacao iniciada. O sistema esta processando a planilha em segundo plano.",
      );
      setExecutionError(null);
    },
  });

  const importStatusQuery = useQuery({
    queryKey: ["budget-import-status", activeImportId],
    queryFn: async () => getBudgetImportStatusRequest(activeImportId ?? ""),
    enabled: activeImportId !== null && isImportProcessing(executionResult),
    refetchInterval: 2000,
  });

  const summaryItems = useMemo(
    () => (previewResult === null ? [] : createSummaryItems(previewResult)),
    [previewResult],
  );
  const catalogItems = useMemo(
    () => (previewResult === null ? [] : createCatalogItems(previewResult)),
    [previewResult],
  );
  const executionProgressValue = useMemo(
    () => getImportProgressValue(executionResult),
    [executionResult],
  );
  const executionIsProcessing = isImportProcessing(executionResult);

  useEffect(() => {
    const statusResult = importStatusQuery.data;

    if (statusResult === undefined) {
      return;
    }

    setExecutionResult(statusResult);
    if (isImportProcessing(statusResult)) {
      return;
    }

    setActiveImportId(null);
    setProcessInfo(null);
    void queryClient.invalidateQueries({ queryKey: ["budgets"] });
  }, [importStatusQuery.data, queryClient]);

  useEffect(() => {
    if (
      importStatusQuery.error === null ||
      importStatusQuery.error === undefined
    ) {
      return;
    }

    setActiveImportId(null);
    setProcessInfo(null);
    setExecutionError(
      getImportErrorMessage(
        importStatusQuery.error,
        "Nao foi possivel consultar o status da importacao.",
      ),
    );
  }, [importStatusQuery.error]);

  const handleChooseFile = () => {
    fileInputRef.current?.click();
  };

  const handleFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const nextFile = event.target.files?.[0] ?? null;
    setSelectedFile(nextFile);
    executingPreviewIdRef.current = null;
    setPreviewResult(null);
    setExecutionResult(null);
    setActiveImportId(null);
    setProcessInfo(null);
    setPreviewError(null);
    setExecutionError(null);
  };

  const handlePreview = async () => {
    try {
      setProcessInfo(null);
      setPreviewError(null);
      await previewMutation.mutateAsync();
    } catch (error) {
      const fallbackMessage =
        error instanceof Error
          ? error.message
          : "Nao foi possivel gerar o preview da planilha.";
      setPreviewError(getImportErrorMessage(error, fallbackMessage));
    }
  };

  const handleExecuteImport = async () => {
    if (previewResult === null) {
      return;
    }

    if (executingPreviewIdRef.current === previewResult.previewId) {
      return;
    }

    executingPreviewIdRef.current = previewResult.previewId;
    try {
      setProcessInfo(null);
      setExecutionError(null);
      await executeMutation.mutateAsync({
        previewId: previewResult.previewId,
      });
    } catch (error) {
      executingPreviewIdRef.current = null;

      if (isPreviewUnavailableError(error) && selectedFile !== null) {
        try {
          const refreshedPreview = await requestPreview();
          executingPreviewIdRef.current = null;
          setPreviewResult(refreshedPreview);
          setExecutionResult(null);
          setPreviewError(null);
          setProcessInfo(
            "O preview anterior nao estava mais disponivel. O sistema esta regenerando os dados e tentando concluir a importacao automaticamente.",
          );
          const executionResponse = await requestExecuteImport(
            refreshedPreview.previewId,
          );
          executingPreviewIdRef.current = executionResponse.previewId;
          setExecutionResult(executionResponse);
          setActiveImportId(executionResponse.importId);
          setProcessInfo(
            "O preview anterior nao estava mais disponivel. A importacao foi reiniciada com um preview atualizado e esta sendo processada em segundo plano.",
          );
          setExecutionError(null);
          return;
        } catch (recoveryError) {
          setProcessInfo(null);
          setExecutionError(
            getImportErrorMessage(
              recoveryError,
              "Nao foi possivel concluir a importacao automaticamente. Gere um novo preview e tente novamente.",
            ),
          );
          return;
        }
      }

      setProcessInfo(null);
      setExecutionError(
        getImportErrorMessage(error, "Nao foi possivel executar a importacao."),
      );
    }
  };

  return (
    <Box sx={{ display: "flex", flexDirection: "column", gap: 3 }}>
      <PageHeader
        action={
          <Button onClick={() => navigate("/budgets")} variant="outlined">
            Voltar para orçamentos
          </Button>
        }
        description="Envie uma planilha .xlsx, revise o layout identificado automaticamente e confirme a carga somente depois da validacao."
        title="Importar planilha de orçamentos"
      />

      <SectionCard
        description="A importacao identifica automaticamente o layout compativel e exige um arquivo .xlsx."
        title="Arquivo e opções"
      >
        <Box sx={{ display: "flex", flexDirection: "column", gap: 2.5 }}>
          <input
            accept=".xlsx"
            hidden
            onChange={handleFileChange}
            ref={fileInputRef}
            type="file"
          />

          <Box
            sx={{
              alignItems: { md: "center", xs: "stretch" },
              display: "flex",
              flexDirection: { md: "row", xs: "column" },
              gap: 2,
            }}
          >
            <Button
              onClick={handleChooseFile}
              startIcon={<UploadFileRoundedIcon />}
              variant="outlined"
            >
              Selecionar planilha
            </Button>
            <Typography color="text.secondary" variant="body2">
              {selectedFile?.name ?? "Nenhum arquivo selecionado."}
            </Typography>
          </Box>

          <Box
            sx={{
              display: "grid",
              gap: 2,
              gridTemplateColumns: {
                md: "minmax(220px, 280px) auto auto",
                xs: "minmax(0, 1fr)",
              },
            }}
          >
            <TextField
              helperText="Configuracao fixa para proteger ajustes manuais ja feitos no sistema."
              label="Duplicidade"
              size="small"
              slotProps={{ input: { readOnly: true } }}
              value="Ignorar existente"
            />
            <FormControlLabel
              control={
                <Checkbox
                  checked={previewOptions.createMissingCatalogs}
                  onChange={(event) =>
                    setPreviewOptions((currentValue) => ({
                      ...currentValue,
                      createMissingCatalogs: event.target.checked,
                    }))
                  }
                />
              }
              label="Criar catálogos ausentes"
            />
            <FormControlLabel
              control={
                <Checkbox
                  checked={previewOptions.useDefaultNotInformed}
                  onChange={(event) =>
                    setPreviewOptions((currentValue) => ({
                      ...currentValue,
                      useDefaultNotInformed: event.target.checked,
                    }))
                  }
                />
              }
              label="Usar Nao informado"
            />
          </Box>

          {processInfo ? <Alert severity="info">{processInfo}</Alert> : null}
          {previewError ? <Alert severity="error">{previewError}</Alert> : null}
          {executionError ? (
            <Alert severity="error">{executionError}</Alert>
          ) : null}
          {previewMutation.isPending ? (
            <Box sx={{ display: "flex", flexDirection: "column", gap: 1 }}>
              <Typography color="text.secondary" variant="body2">
                Gerando preview da planilha. Aguarde a identificacao do layout e
                a leitura das linhas importaveis.
              </Typography>
              <LinearProgress />
            </Box>
          ) : null}
          {executeMutation.isPending ? (
            <Box sx={{ display: "flex", flexDirection: "column", gap: 1 }}>
              <Typography color="text.secondary" variant="body2">
                Iniciando a importacao. Aguarde a confirmacao do processamento.
              </Typography>
              <LinearProgress />
            </Box>
          ) : null}

          <Box
            sx={{
              display: "flex",
              flexDirection: { sm: "row", xs: "column" },
              gap: 1.5,
            }}
          >
            <Button
              disabled={
                selectedFile === null ||
                previewMutation.isPending ||
                executionIsProcessing
              }
              onClick={handlePreview}
              startIcon={<VisibilityRoundedIcon />}
              variant="contained"
            >
              {previewMutation.isPending
                ? "Gerando preview..."
                : "Gerar preview"}
            </Button>
            <Button
              color="success"
              disabled={
                previewResult === null ||
                executeMutation.isPending ||
                executionIsProcessing ||
                executionResult?.previewId === previewResult.previewId
              }
              onClick={handleExecuteImport}
              startIcon={<TaskAltRoundedIcon />}
              variant="contained"
            >
              {executeMutation.isPending
                ? "Iniciando..."
                : "Confirmar importação"}
            </Button>
          </Box>
        </Box>
      </SectionCard>

      {executionIsProcessing && executionResult ? (
        <SectionCard
          description={`Importação ${executionResult.importId} iniciada em ${formatDateTime(executionResult.startedAt)}.`}
          title="Processamento da importação"
        >
          <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
            <Alert severity="info">
              {processInfo ?? executionResult.result.message}
            </Alert>
            <Box sx={{ display: "flex", flexDirection: "column", gap: 1 }}>
              <Typography color="text.secondary" variant="body2">
                {executionResult.summary.rowsExpected > 0
                  ? `${executionResult.summary.rowsProcessed} de ${executionResult.summary.rowsExpected} linha(s) processada(s).`
                  : "Processando as linhas da planilha..."}
              </Typography>
              <LinearProgress
                value={executionProgressValue ?? 0}
                variant={
                  executionProgressValue === null
                    ? "indeterminate"
                    : "determinate"
                }
              />
            </Box>
          </Box>
        </SectionCard>
      ) : null}

      {executionResult && !executionIsProcessing ? (
        <SectionCard
          description={`Importação ${executionResult.importId} finalizada em ${formatDateTime(executionResult.finishedAt)}.`}
          title="Resultado da importação"
        >
          <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
            <Alert
              severity={
                executionResult.status === "failed" ? "error" : "success"
              }
            >
              {executionResult.status === "failed"
                ? "A importacao nao foi concluida. Revise a mensagem abaixo e tente novamente."
                : getExecutionSuccessMessage(
                    previewResult?.fileName ?? "selecionado",
                    executionResult,
                  )}
            </Alert>
            <Alert severity="info" variant="outlined">
              {executionResult.result.message}
            </Alert>
            {executionResult.summary.rowsFailed > 0 ? (
              <Alert severity="warning">
                {`${executionResult.summary.rowsFailed} linha(s) com inconsistência não foram importadas, mas o restante do arquivo foi processado com sucesso.`}
              </Alert>
            ) : null}
            <Box
              sx={{
                display: "grid",
                gap: 1.5,
                gridTemplateColumns: {
                  lg: "repeat(6, minmax(0, 1fr))",
                  md: "repeat(3, minmax(0, 1fr))",
                  xs: "repeat(2, minmax(0, 1fr))",
                },
              }}
            >
              <Box
                sx={{
                  border: "1px solid",
                  borderColor: "divider",
                  borderRadius: 3,
                  p: 2,
                }}
              >
                <Typography color="text.secondary" variant="body2">
                  Processadas
                </Typography>
                <Typography variant="h6">
                  {executionResult.summary.rowsProcessed}
                </Typography>
              </Box>
              <Box
                sx={{
                  border: "1px solid",
                  borderColor: "divider",
                  borderRadius: 3,
                  p: 2,
                }}
              >
                <Typography color="text.secondary" variant="body2">
                  Criadas
                </Typography>
                <Typography variant="h6">
                  {executionResult.summary.budgetsCreated}
                </Typography>
              </Box>
              <Box
                sx={{
                  border: "1px solid",
                  borderColor: "divider",
                  borderRadius: 3,
                  p: 2,
                }}
              >
                <Typography color="text.secondary" variant="body2">
                  Atualizadas
                </Typography>
                <Typography variant="h6">
                  {executionResult.summary.budgetsUpdated}
                </Typography>
              </Box>
              <Box
                sx={{
                  border: "1px solid",
                  borderColor: "divider",
                  borderRadius: 3,
                  p: 2,
                }}
              >
                <Typography color="text.secondary" variant="body2">
                  Ignoradas
                </Typography>
                <Typography variant="h6">
                  {executionResult.summary.budgetsIgnored}
                </Typography>
              </Box>
              <Box
                sx={{
                  border: "1px solid",
                  borderColor: "divider",
                  borderRadius: 3,
                  p: 2,
                }}
              >
                <Typography color="text.secondary" variant="body2">
                  Falhas
                </Typography>
                <Typography variant="h6">
                  {executionResult.summary.rowsFailed}
                </Typography>
              </Box>
              <Box
                sx={{
                  border: "1px solid",
                  borderColor: "divider",
                  borderRadius: 3,
                  p: 2,
                }}
              >
                <Typography color="text.secondary" variant="body2">
                  Catálogos criados
                </Typography>
                <Typography variant="h6">
                  {executionResult.summary.catalogsCreated}
                </Typography>
              </Box>
            </Box>
          </Box>
        </SectionCard>
      ) : null}

      {previewResult ? (
        <SectionCard
          description={`Arquivo ${previewResult.fileName} | Layout ${previewResult.layout.name} | Aba ${previewResult.sheetName} | Cabecalho na linha ${previewResult.headerRow}`}
          title="Preview da importação"
        >
          <Box sx={{ display: "flex", flexDirection: "column", gap: 2.5 }}>
            <Alert severity="info" variant="outlined">
              {`Layout identificado: ${previewResult.layout.name} | Origem: ${previewResult.layout.sourceCompany}. ${previewResult.layout.description}`}
            </Alert>

            <Box
              sx={{
                display: "grid",
                gap: 1.5,
                gridTemplateColumns: {
                  lg: "repeat(3, minmax(0, 1fr))",
                  xs: "minmax(0, 1fr)",
                },
              }}
            >
              <Box
                sx={{
                  border: "1px solid",
                  borderColor: "divider",
                  borderRadius: 3,
                  p: 2,
                }}
              >
                <Typography color="text.secondary" variant="body2">
                  Layout detectado
                </Typography>
                <Typography variant="h6">
                  {previewResult.layout.name}
                </Typography>
                <Typography color="text.secondary" variant="body2">
                  {previewResult.layout.key}
                </Typography>
              </Box>
              <Box
                sx={{
                  border: "1px solid",
                  borderColor: "divider",
                  borderRadius: 3,
                  p: 2,
                }}
              >
                <Typography color="text.secondary" variant="body2">
                  Empresa de origem
                </Typography>
                <Typography variant="h6">
                  {previewResult.layout.sourceCompany}
                </Typography>
                <Typography color="text.secondary" variant="body2">
                  Metadado gravado no lote e nos orcamentos importados
                </Typography>
              </Box>
              <Box
                sx={{
                  border: "1px solid",
                  borderColor: "divider",
                  borderRadius: 3,
                  p: 2,
                }}
              >
                <Typography color="text.secondary" variant="body2">
                  Estrategia aplicada
                </Typography>
                <Typography variant="h6">Ignorar duplicados</Typography>
                <Typography color="text.secondary" variant="body2">
                  {previewResult.options.createMissingCatalogs
                    ? "Cria catalogos ausentes durante a carga"
                    : "Nao cria catalogos ausentes durante a carga"}
                </Typography>
              </Box>
            </Box>

            <Box
              sx={{
                display: "grid",
                gap: 1.5,
                gridTemplateColumns: {
                  lg: "repeat(6, minmax(0, 1fr))",
                  md: "repeat(3, minmax(0, 1fr))",
                  xs: "repeat(2, minmax(0, 1fr))",
                },
              }}
            >
              {summaryItems.map((item) => (
                <Box
                  key={item.label}
                  sx={{
                    bgcolor: "rgba(37, 99, 235, 0.06)",
                    borderRadius: 3,
                    p: 2,
                  }}
                >
                  <Typography color="text.secondary" variant="body2">
                    {item.label}
                  </Typography>
                  <Typography variant="h5">{item.value}</Typography>
                </Box>
              ))}
            </Box>

            <Box sx={{ display: "flex", flexDirection: "column", gap: 1.5 }}>
              <Typography variant="h6">Governanca da importacao</Typography>
              <Box
                sx={{
                  display: "grid",
                  gap: 1.5,
                  gridTemplateColumns: {
                    lg: "repeat(2, minmax(0, 1fr))",
                    xs: "minmax(0, 1fr)",
                  },
                }}
              >
                <Box
                  sx={{
                    border: "1px solid",
                    borderColor: "divider",
                    borderRadius: 3,
                    p: 2,
                  }}
                >
                  <Typography color="text.secondary" variant="body2">
                    Escopo da duplicidade
                  </Typography>
                  <Typography variant="h6">
                    {previewResult.governance.duplicateScope}
                  </Typography>
                  <Typography color="text.secondary" variant="body2">
                    {previewResult.governance.duplicatePolicy}
                  </Typography>
                </Box>
                <Box
                  sx={{
                    border: "1px solid",
                    borderColor: "divider",
                    borderRadius: 3,
                    p: 2,
                  }}
                >
                  <Typography color="text.secondary" variant="body2">
                    Politica para valores ausentes
                  </Typography>
                  <Typography variant="body1">
                    {previewResult.governance.missingValuePolicy}
                  </Typography>
                  <Typography color="text.secondary" variant="body2">
                    {previewResult.governance.legacyMatchingScope}
                  </Typography>
                </Box>
              </Box>
              <Box sx={{ display: "flex", flexWrap: "wrap", gap: 1 }}>
                {previewResult.governance.defaultCatalogs.map((catalogName) => (
                  <Chip
                    key={`governance-${catalogName}`}
                    color="default"
                    label={catalogName}
                    size="small"
                    variant="outlined"
                  />
                ))}
              </Box>
            </Box>

            <Box
              sx={{
                display: "grid",
                gap: 1.5,
                gridTemplateColumns: {
                  lg: "repeat(4, minmax(0, 1fr))",
                  md: "repeat(2, minmax(0, 1fr))",
                  xs: "minmax(0, 1fr)",
                },
              }}
            >
              {catalogItems.map((item) => (
                <Box
                  key={item.label}
                  sx={{
                    border: "1px solid",
                    borderColor: "divider",
                    borderRadius: 3,
                    p: 2,
                  }}
                >
                  <Typography color="text.secondary" variant="body2">
                    {item.label}
                  </Typography>
                  <Typography variant="h6">{item.value}</Typography>
                </Box>
              ))}
            </Box>

            {previewResult.fieldGroups.length > 0 ? (
              <Box sx={{ display: "flex", flexDirection: "column", gap: 1.5 }}>
                <Typography variant="h6">Mapeamento do layout</Typography>
                <Typography color="text.secondary" variant="body2">
                  Veja quais campos entram no dominio principal e quais ficam
                  preservados apenas para rastreabilidade do lote.
                </Typography>
                <Box
                  sx={{
                    display: "grid",
                    gap: 1.5,
                    gridTemplateColumns: {
                      lg: "repeat(3, minmax(0, 1fr))",
                      md: "repeat(2, minmax(0, 1fr))",
                      xs: "minmax(0, 1fr)",
                    },
                  }}
                >
                  {previewResult.fieldGroups.map((group) => (
                    <Box
                      key={group.key}
                      sx={{
                        border: "1px solid",
                        borderColor: "divider",
                        borderRadius: 3,
                        p: 2,
                      }}
                    >
                      <Typography variant="subtitle1">{group.title}</Typography>
                      <Typography color="text.secondary" variant="body2">
                        {group.description}
                      </Typography>
                      <Box
                        sx={{
                          display: "flex",
                          flexWrap: "wrap",
                          gap: 1,
                          mt: 1.5,
                        }}
                      >
                        {group.fields.map((field) => (
                          <Chip
                            key={`${group.key}-${field}`}
                            label={field}
                            size="small"
                            variant="outlined"
                          />
                        ))}
                      </Box>
                    </Box>
                  ))}
                </Box>
              </Box>
            ) : null}

            {previewResult.warnings.length > 0 ? (
              <Alert severity="warning">
                {previewResult.warnings.map((item) => item.message).join(" | ")}
              </Alert>
            ) : null}

            {previewResult.summary.rowsWithError > 0 ? (
              <Alert severity="warning">
                {`${previewResult.summary.rowsWithError} linha(s) com inconsistência não serão importadas. As demais linhas válidas podem ser confirmadas normalmente.`}
              </Alert>
            ) : null}

            {previewResult.errors.length > 0 ? (
              <Alert severity="error">
                {previewResult.errors.map((item) => item.message).join(" | ")}
              </Alert>
            ) : null}

            <Box sx={{ display: "flex", flexWrap: "wrap", gap: 1 }}>
              <Chip
                color="info"
                label={`${previewResult.summary.rowsEmptyIgnored} linha(s) vazia(s) ignorada(s)`}
                variant="outlined"
              />
              <Chip
                color="primary"
                label={`${previewResult.summary.newBudgets} novo(s) orçamento(s)`}
                variant="outlined"
              />
              <Chip
                color="secondary"
                label={`${previewResult.summary.existingBudgets} orçamento(s) existente(s)`}
                variant="outlined"
              />
            </Box>

            <Box sx={{ overflowX: "auto" }}>
              <Table size="small">
                <TableHead>
                  <TableRow>
                    <TableCell>Linha</TableCell>
                    <TableCell>Orçamento</TableCell>
                    <TableCell>Status</TableCell>
                    <TableCell>Ação</TableCell>
                    <TableCell>Mensagens</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {previewResult.sampleRows.map((item) => (
                    <TableRow key={`${item.rowNumber}-${item.budgetNumber}`}>
                      <TableCell>{item.rowNumber}</TableCell>
                      <TableCell>
                        {item.budgetNumber || "Nao informado"}
                      </TableCell>
                      <TableCell>
                        <Chip
                          color={item.status === "error" ? "error" : "default"}
                          label={item.status}
                          size="small"
                          variant="outlined"
                        />
                      </TableCell>
                      <TableCell>{item.action}</TableCell>
                      <TableCell>{item.messages.join(" | ") || "-"}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </Box>

            {previewResult.inconsistencyRows.length > 0 ? (
              <Box sx={{ display: "flex", flexDirection: "column", gap: 1.5 }}>
                <Typography variant="h6">Linhas com inconsistência</Typography>
                <Typography color="text.secondary" variant="body2">
                  Essas linhas serão ignoradas na importação e servem como aviso
                  para correção da planilha.
                </Typography>
                <Box sx={{ overflowX: "auto" }}>
                  <Table size="small">
                    <TableHead>
                      <TableRow>
                        <TableCell>Linha</TableCell>
                        <TableCell>Orçamento</TableCell>
                        <TableCell>Motivo</TableCell>
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {previewResult.inconsistencyRows.map((item) => (
                        <TableRow
                          key={`inconsistency-${item.rowNumber}-${item.budgetNumber}`}
                        >
                          <TableCell>{item.rowNumber}</TableCell>
                          <TableCell>
                            {item.budgetNumber || "Nao informado"}
                          </TableCell>
                          <TableCell>
                            {item.messages.join(" | ") || "-"}
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </Box>
              </Box>
            ) : null}
          </Box>
        </SectionCard>
      ) : null}
    </Box>
  );
}
