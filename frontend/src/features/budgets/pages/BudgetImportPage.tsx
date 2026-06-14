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
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { isAxiosError } from "axios";
import { useMemo, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { PageHeader } from "../../../components/common/PageHeader";
import { SectionCard } from "../../../components/common/SectionCard";
import {
  executeBudgetImportRequest,
  previewBudgetImportRequest,
} from "../api/budgets";
import type {
  BudgetImportExecutionResult,
  BudgetImportPreviewOptions,
  BudgetImportPreviewResult,
} from "../types/budget";

function getImportErrorMessage(error: unknown, fallbackMessage: string) {
  if (isAxiosError<{ message?: string }>(error)) {
    return error.response?.data?.message ?? fallbackMessage;
  }

  return fallbackMessage;
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
      label: "Projetos",
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

export function BudgetImportPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [previewOptions, setPreviewOptions] =
    useState<BudgetImportPreviewOptions>({
      duplicateStrategy: "update",
      createMissingCatalogs: true,
      useDefaultNotInformed: true,
    });
  const [previewResult, setPreviewResult] =
    useState<BudgetImportPreviewResult | null>(null);
  const [executionResult, setExecutionResult] =
    useState<BudgetImportExecutionResult | null>(null);
  const [previewError, setPreviewError] = useState<string | null>(null);
  const [executionError, setExecutionError] = useState<string | null>(null);

  const previewMutation = useMutation({
    mutationFn: async () => {
      if (selectedFile === null) {
        throw new Error("Selecione um arquivo .xlsx para gerar o preview.");
      }

      return previewBudgetImportRequest(selectedFile, previewOptions);
    },
    onSuccess: (result) => {
      setPreviewResult(result);
      setExecutionResult(null);
      setPreviewError(null);
      setExecutionError(null);
    },
  });

  const executeMutation = useMutation({
    mutationFn: async () => {
      if (previewResult === null) {
        throw new Error("Gere o preview antes de confirmar a importação.");
      }

      return executeBudgetImportRequest({
        previewId: previewResult.previewId,
      });
    },
    onSuccess: async (result) => {
      setExecutionResult(result);
      setExecutionError(null);
      await queryClient.invalidateQueries({ queryKey: ["budgets"] });
    },
  });

  const summaryItems = useMemo(
    () => (previewResult === null ? [] : createSummaryItems(previewResult)),
    [previewResult],
  );
  const catalogItems = useMemo(
    () => (previewResult === null ? [] : createCatalogItems(previewResult)),
    [previewResult],
  );

  const handleChooseFile = () => {
    fileInputRef.current?.click();
  };

  const handleFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const nextFile = event.target.files?.[0] ?? null;
    setSelectedFile(nextFile);
    setPreviewResult(null);
    setExecutionResult(null);
    setPreviewError(null);
    setExecutionError(null);
  };

  const handlePreview = async () => {
    try {
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
    try {
      setExecutionError(null);
      await executeMutation.mutateAsync();
    } catch (error) {
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
        description="Envie a aba ORCAMENTOS da planilha, revise o preview e confirme a carga somente depois da validação."
        title="Importar planilha de orçamentos"
      />

      <SectionCard
        description="A importação usa somente a aba ORCAMENTOS e exige um arquivo .xlsx."
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
                md: "minmax(180px, 220px) auto auto",
                xs: "minmax(0, 1fr)",
              },
            }}
          >
            <TextField
              label="Duplicidade"
              onChange={(event) =>
                setPreviewOptions((currentValue) => ({
                  ...currentValue,
                  duplicateStrategy: event.target.value as "ignore" | "update",
                }))
              }
              select
              size="small"
              value={previewOptions.duplicateStrategy}
            >
              <MenuItem value="update">Atualizar existente</MenuItem>
              <MenuItem value="ignore">Ignorar existente</MenuItem>
            </TextField>
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

          {previewError ? <Alert severity="error">{previewError}</Alert> : null}
          {executionError ? (
            <Alert severity="error">{executionError}</Alert>
          ) : null}
          {previewMutation.isPending ? (
            <Box sx={{ display: "flex", flexDirection: "column", gap: 1 }}>
              <Typography color="text.secondary" variant="body2">
                Gerando preview da planilha. Aguarde a leitura da aba
                ORCAMENTOS.
              </Typography>
              <LinearProgress />
            </Box>
          ) : null}
          {executeMutation.isPending ? (
            <Box sx={{ display: "flex", flexDirection: "column", gap: 1 }}>
              <Typography color="text.secondary" variant="body2">
                Importacao em andamento. Aguarde o processamento da planilha.
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
              disabled={selectedFile === null || previewMutation.isPending}
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
              disabled={previewResult === null || executeMutation.isPending}
              onClick={handleExecuteImport}
              startIcon={<TaskAltRoundedIcon />}
              variant="contained"
            >
              {executeMutation.isPending
                ? "Importando..."
                : "Confirmar importação"}
            </Button>
          </Box>
        </Box>
      </SectionCard>

      {executionResult ? (
        <SectionCard
          description={`Importação ${executionResult.importId} finalizada em ${formatDateTime(executionResult.finishedAt)}.`}
          title="Resultado da importação"
        >
          <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
            <Alert
              severity={
                executionResult.summary.rowsFailed > 0 ? "warning" : "success"
              }
            >
              {executionResult.summary.rowsFailed > 0
                ? `${executionResult.result.message} ${executionResult.summary.rowsFailed} linha(s) com inconsistência não foram importadas.`
                : executionResult.result.message}
            </Alert>
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
          description={`Arquivo ${previewResult.fileName} | Aba ${previewResult.sheetName} | Cabeçalho na linha ${previewResult.headerRow}`}
          title="Preview da importação"
        >
          <Box sx={{ display: "flex", flexDirection: "column", gap: 2.5 }}>
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
