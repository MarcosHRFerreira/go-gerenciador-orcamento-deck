import AttachMoneyRoundedIcon from "@mui/icons-material/AttachMoneyRounded";
import AssignmentTurnedInRoundedIcon from "@mui/icons-material/AssignmentTurnedInRounded";
import DescriptionRoundedIcon from "@mui/icons-material/DescriptionRounded";
import DownloadRoundedIcon from "@mui/icons-material/DownloadRounded";
import InsightsRoundedIcon from "@mui/icons-material/InsightsRounded";
import OpenInNewRoundedIcon from "@mui/icons-material/OpenInNewRounded";
import TrendingUpRoundedIcon from "@mui/icons-material/TrendingUpRounded";
import {
  Alert,
  Box,
  Button,
  Chip,
  Divider,
  LinearProgress,
  MenuItem,
  TextField,
  Typography,
} from "@mui/material";
import { AxiosError, isAxiosError } from "axios";
import { useQuery } from "@tanstack/react-query";
import { useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { PageHeader } from "../../../components/common/PageHeader";
import { SectionCard } from "../../../components/common/SectionCard";
import { useAuth } from "../../auth/hooks/useAuth";
import { getBudgetCatalogsRequest } from "../../budgets/api/budgets";
import type { BudgetCatalogItem } from "../../budgets/types/budget";
import { getSalespeopleDashboardRequest } from "../api/dashboard";
import type {
  DashboardEstimatorSummary,
  DashboardMonthlyEvolutionItem,
  DashboardSalespeopleFilters,
  DashboardSalespersonFunnelItem,
  DashboardSalespersonSummary,
  DashboardStaleBudgetItem,
} from "../types/dashboard";

const currencyFormatter = new Intl.NumberFormat("pt-BR", {
  style: "currency",
  currency: "BRL",
});

const dateFormatter = new Intl.DateTimeFormat("pt-BR", {
  dateStyle: "short",
});

const monthYearFormatter = new Intl.DateTimeFormat("pt-BR", {
  month: "short",
  year: "numeric",
});

type DashboardCompanyFilter = "" | "Rocktec" | "Trox";

type DashboardMetricCard = {
  key: string;
  label: string;
  value: string;
  helper: string;
  icon: typeof TrendingUpRoundedIcon;
};

type BudgetListNavigationOptions = {
  budgetNumber?: string;
  salespersonId?: string;
  sourceCompany: DashboardCompanyFilter;
  statusId?: string;
  year: string;
  month: string;
};

type ApiErrorResponse = {
  message?: string;
};

type FilePickerAcceptType = {
  accept: Record<string, string[]>;
  description?: string;
};

type SaveFilePickerOptions = {
  suggestedName?: string;
  types?: FilePickerAcceptType[];
};

type WritableFileStream = {
  close: () => Promise<void>;
  write: (data: Blob) => Promise<void>;
};

type FileHandle = {
  createWritable: () => Promise<WritableFileStream>;
};

type SaveFilePickerWindow = Window & {
  showSaveFilePicker?: (options?: SaveFilePickerOptions) => Promise<FileHandle>;
};

type SpreadsheetValue = string | number | boolean | Date | null;

type SpreadsheetRow = SpreadsheetValue[];

type SpreadsheetCellFormat = "currency" | "date" | "integer" | "percent";

type SpreadsheetSheet = {
  autoFilter?: boolean;
  headerStyle?: "executive" | "table";
  formatsByColumn?: Partial<Record<number, SpreadsheetCellFormat>>;
  headerRowIndex?: number;
  mergeRanges?: string[];
  name: string;
  rows: SpreadsheetRow[];
  sectionRowIndexes?: number[];
  titleRowIndex?: number;
};

const monthOptions = [
  { value: "1", label: "Janeiro" },
  { value: "2", label: "Fevereiro" },
  { value: "3", label: "Marco" },
  { value: "4", label: "Abril" },
  { value: "5", label: "Maio" },
  { value: "6", label: "Junho" },
  { value: "7", label: "Julho" },
  { value: "8", label: "Agosto" },
  { value: "9", label: "Setembro" },
  { value: "10", label: "Outubro" },
  { value: "11", label: "Novembro" },
  { value: "12", label: "Dezembro" },
] as const;

function formatPercentage(value: number) {
  return `${value.toFixed(1)}%`;
}

function formatCompactCurrency(value: number) {
  if (value >= 1_000_000) {
    return `R$ ${(value / 1_000_000).toFixed(1)} mi`;
  }

  if (value >= 1_000) {
    return `R$ ${(value / 1_000).toFixed(1)} mil`;
  }

  return currencyFormatter.format(value);
}

function formatDateOrFallback(value: string | null, fallback: string) {
  if (value === null) {
    return fallback;
  }

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return fallback;
  }

  return dateFormatter.format(date);
}

function formatMonthKeyLabel(value: string) {
  const date = new Date(`${value}-01T00:00:00`);
  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return monthYearFormatter.format(date);
}

function formatSignedCurrency(value: number) {
  const absoluteValue = currencyFormatter.format(Math.abs(value));

  if (value > 0) {
    return `+${absoluteValue}`;
  }
  if (value < 0) {
    return `-${absoluteValue}`;
  }

  return absoluteValue;
}

function formatSignedInteger(value: number) {
  if (value > 0) {
    return `+${value}`;
  }

  return String(value);
}

function formatSignedPercentagePoints(value: number) {
  if (value > 0) {
    return `+${value.toFixed(1)} p.p.`;
  }
  if (value < 0) {
    return `${value.toFixed(1)} p.p.`;
  }

  return "0,0 p.p.";
}

function formatClosingDays(value: number) {
  return `${value.toFixed(1).replace(".", ",")} dia(s)`;
}

function getDeltaColor(value: number) {
  if (value > 0) {
    return "success";
  }
  if (value < 0) {
    return "error";
  }

  return "default";
}

function getDashboardErrorMessage(error: unknown) {
  if (isAxiosError<ApiErrorResponse>(error)) {
    const statusCode = error.response?.status;
    const apiMessage = error.response?.data?.message?.trim();

    if (statusCode === 401) {
      return "Sua sessao expirou. Entre novamente para acessar o dashboard.";
    }
    if (statusCode === 403) {
      return (
        apiMessage ||
        "Voce nao possui permissao para acessar este dashboard administrativo."
      );
    }
    if (statusCode === 400) {
      return (
        apiMessage || "Os filtros informados para o dashboard sao invalidos."
      );
    }

    return apiMessage || "Nao foi possivel carregar os dados do dashboard.";
  }

  return "Nao foi possivel carregar os dados do dashboard.";
}

function getMonthDateRange(year: string, month: string) {
  if (!year || !month) {
    return null;
  }

  const normalizedMonth = month.padStart(2, "0");
  const monthStart = new Date(`${year}-${normalizedMonth}-01T00:00:00`);
  if (Number.isNaN(monthStart.getTime())) {
    return null;
  }

  const monthEnd = new Date(
    monthStart.getFullYear(),
    monthStart.getMonth() + 1,
    0,
  );
  const from = `${year}-${normalizedMonth}-01`;
  const to = `${year}-${normalizedMonth}-${String(monthEnd.getDate()).padStart(2, "0")}`;

  return { from, to };
}

function buildBudgetListSearchParams({
  budgetNumber,
  month,
  salespersonId,
  sourceCompany,
  statusId,
  year,
}: BudgetListNavigationOptions) {
  const searchParams = new URLSearchParams();

  if (budgetNumber) {
    searchParams.set("budgetNumber", budgetNumber);
  }
  if (sourceCompany) {
    searchParams.set("sourceCompany", sourceCompany);
  }
  if (year) {
    searchParams.set("yearBudget", year);
  }
  if (salespersonId) {
    searchParams.set("salespersonId", salespersonId);
  }
  if (statusId) {
    searchParams.set("statusId", statusId);
  }

  const monthDateRange = getMonthDateRange(year, month);
  if (monthDateRange !== null) {
    searchParams.set("sentAtFrom", monthDateRange.from);
    searchParams.set("sentAtTo", monthDateRange.to);
  }

  searchParams.set("page", "1");
  searchParams.set("pageSize", "50");
  searchParams.set("sortBy", "updated_at");
  searchParams.set("sortOrder", "desc");

  return searchParams;
}

function escapeCsvValue(value: string) {
  return `"${value.replaceAll('"', '""')}"`;
}

function createCsvLine(values: Array<string | number>) {
  return values.map((value) => escapeCsvValue(String(value))).join(";");
}

function normalizeLookupKey(value: string) {
  return value
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "")
    .trim()
    .toLowerCase();
}

function buildDashboardScopeLabel(
  sourceCompany: DashboardCompanyFilter,
  selectedYear: string,
  selectedMonth: string,
  salespersonLabel: string,
) {
  const companyLabel = sourceCompany || "todas as empresas";
  const yearLabel = selectedYear || "todos os anos";
  const monthLabel =
    monthOptions.find((item) => item.value === selectedMonth)?.label ??
    "todos os meses";
  const salespersonScope = salespersonLabel || "todos os vendedores";

  return `${companyLabel}, ${yearLabel}, ${monthLabel} e ${salespersonScope}`;
}

function downloadBlob(csvBlob: Blob, fileName: string) {
  const downloadUrl = window.URL.createObjectURL(csvBlob);
  const downloadLink = document.createElement("a");

  downloadLink.href = downloadUrl;
  downloadLink.download = fileName;
  document.body.appendChild(downloadLink);
  downloadLink.click();
  document.body.removeChild(downloadLink);
  window.URL.revokeObjectURL(downloadUrl);
}

async function saveBlobWithPicker(
  fileBlob: Blob,
  fileName: string,
  fileType: string,
  fileExtension: string,
) {
  const pickerWindow = window as SaveFilePickerWindow;
  if (typeof pickerWindow.showSaveFilePicker !== "function") {
    return false;
  }

  try {
    const fileHandle = await pickerWindow.showSaveFilePicker({
      suggestedName: fileName,
      types: [
        {
          accept: {
            [fileType]: [fileExtension],
          },
          description: `Arquivo ${fileExtension.replace(".", "").toUpperCase()}`,
        },
      ],
    });
    const writableStream = await fileHandle.createWritable();
    await writableStream.write(fileBlob);
    await writableStream.close();

    return true;
  } catch (error) {
    if (error instanceof DOMException && error.name === "AbortError") {
      return true;
    }

    return false;
  }
}

function toSpreadsheetDateOrFallback(value: string | null, fallback: string) {
  if (value === null) {
    return fallback;
  }

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return fallback;
  }

  return date;
}

function getSpreadsheetCellDisplayLength(value: SpreadsheetValue) {
  if (value === null) {
    return 0;
  }
  if (value instanceof Date) {
    return dateFormatter.format(value).length;
  }
  if (typeof value === "number") {
    return String(value).length;
  }
  if (typeof value === "boolean") {
    return value ? 4 : 5;
  }

  return value.length;
}

function buildWorksheetColumnWidths(rows: SpreadsheetRow[]) {
  const maxColumns = rows.reduce(
    (currentMax, row) => Math.max(currentMax, row.length),
    0,
  );

  return Array.from({ length: maxColumns }, (_, columnIndex) => {
    const largestWidth = rows.reduce((currentMax, row) => {
      const cellValue = row[columnIndex] ?? null;
      return Math.max(currentMax, getSpreadsheetCellDisplayLength(cellValue));
    }, 0);

    return {
      wch: Math.min(42, Math.max(12, largestWidth + 2)),
    };
  });
}

function getSpreadsheetNumberFormat(format: SpreadsheetCellFormat) {
  if (format === "currency") {
    return '"R$" #,##0.00';
  }
  if (format === "date") {
    return "dd/mm/yyyy";
  }
  if (format === "percent") {
    return "0.0%";
  }

  return "0";
}

function getSpreadsheetBorderStyle(): import("xlsx-js-style").CellStyle["border"] {
  return {
    bottom: {
      color: { rgb: "FFD0D7DE" },
      style: "thin",
    },
    left: {
      color: { rgb: "FFD0D7DE" },
      style: "thin",
    },
    right: {
      color: { rgb: "FFD0D7DE" },
      style: "thin",
    },
    top: {
      color: { rgb: "FFD0D7DE" },
      style: "thin",
    },
  };
}

function getTitleCellStyle(): import("xlsx-js-style").CellStyle {
  return {
    alignment: {
      horizontal: "left",
      vertical: "center",
    },
    fill: {
      fgColor: { rgb: "FF0F4C81" },
      patternType: "solid",
    },
    font: {
      bold: true,
      color: { rgb: "FFFFFFFF" },
      sz: 14,
    },
  };
}

function getSectionCellStyle(): import("xlsx-js-style").CellStyle {
  return {
    alignment: {
      horizontal: "left",
      vertical: "center",
    },
    fill: {
      fgColor: { rgb: "FFDCE6F1" },
      patternType: "solid",
    },
    font: {
      bold: true,
      color: { rgb: "FF1F1F1F" },
      sz: 11,
    },
  };
}

function getHeaderCellStyle(
  headerStyle: SpreadsheetSheet["headerStyle"],
): import("xlsx-js-style").CellStyle {
  const fillColor = headerStyle === "executive" ? "FF4472C4" : "FF5B9BD5";

  return {
    alignment: {
      horizontal: "center",
      vertical: "center",
      wrapText: true,
    },
    border: getSpreadsheetBorderStyle(),
    fill: {
      fgColor: { rgb: fillColor },
      patternType: "solid",
    },
    font: {
      bold: true,
      color: { rgb: "FFFFFFFF" },
    },
  };
}

function getBodyCellStyle(): import("xlsx-js-style").CellStyle {
  return {
    alignment: {
      vertical: "center",
      wrapText: true,
    },
    border: getSpreadsheetBorderStyle(),
  };
}

async function applyWorksheetFormatting(
  XLSX: typeof import("xlsx-js-style"),
  worksheet: import("xlsx-js-style").WorkSheet,
  sheet: SpreadsheetSheet,
) {
  const worksheetRange = worksheet["!ref"];
  if (!worksheetRange) {
    return;
  }

  worksheet["!cols"] = buildWorksheetColumnWidths(sheet.rows);
  worksheet["!rows"] = sheet.rows.map(() => ({ hpt: 22 }));

  if (sheet.mergeRanges && sheet.mergeRanges.length > 0) {
    worksheet["!merges"] = sheet.mergeRanges.map((range) =>
      XLSX.utils.decode_range(range),
    );
  }

  const decodedRange = XLSX.utils.decode_range(worksheetRange);
  const bodyCellStyle = getBodyCellStyle();
  for (
    let rowIndex = decodedRange.s.r;
    rowIndex <= decodedRange.e.r;
    rowIndex += 1
  ) {
    for (
      let columnIndex = decodedRange.s.c;
      columnIndex <= decodedRange.e.c;
      columnIndex += 1
    ) {
      const cellAddress = XLSX.utils.encode_cell({
        c: columnIndex,
        r: rowIndex,
      });
      const cell = worksheet[cellAddress];
      if (!cell) {
        continue;
      }

      cell.s = bodyCellStyle;
    }
  }

  if (typeof sheet.titleRowIndex === "number") {
    const titleCellAddress = XLSX.utils.encode_cell({
      c: decodedRange.s.c,
      r: sheet.titleRowIndex,
    });
    const titleCell = worksheet[titleCellAddress];
    if (titleCell) {
      titleCell.s = getTitleCellStyle();
    }
  }

  sheet.sectionRowIndexes?.forEach((rowIndex) => {
    const sectionCellAddress = XLSX.utils.encode_cell({
      c: decodedRange.s.c,
      r: rowIndex,
    });
    const sectionCell = worksheet[sectionCellAddress];
    if (sectionCell) {
      sectionCell.s = getSectionCellStyle();
    }
  });

  if (
    sheet.autoFilter &&
    typeof sheet.headerRowIndex === "number" &&
    sheet.rows.length > sheet.headerRowIndex
  ) {
    worksheet["!autofilter"] = {
      ref: XLSX.utils.encode_range({
        s: { c: decodedRange.s.c, r: sheet.headerRowIndex },
        e: { c: decodedRange.e.c, r: decodedRange.e.r },
      }),
    };
  }

  if (typeof sheet.headerRowIndex === "number") {
    for (
      let columnIndex = decodedRange.s.c;
      columnIndex <= decodedRange.e.c;
      columnIndex += 1
    ) {
      const headerCellAddress = XLSX.utils.encode_cell({
        c: columnIndex,
        r: sheet.headerRowIndex,
      });
      const headerCell = worksheet[headerCellAddress];
      if (!headerCell) {
        continue;
      }

      headerCell.s = getHeaderCellStyle(sheet.headerStyle);
    }
  }

  if (!sheet.formatsByColumn || typeof sheet.headerRowIndex !== "number") {
    return;
  }

  const headerRowIndex = sheet.headerRowIndex;
  Object.entries(sheet.formatsByColumn).forEach(([columnKey, format]) => {
    const columnIndex = Number(columnKey);
    if (!Number.isInteger(columnIndex) || !format) {
      return;
    }

    for (
      let rowIndex = headerRowIndex + 1;
      rowIndex <= decodedRange.e.r;
      rowIndex += 1
    ) {
      const cellAddress = XLSX.utils.encode_cell({
        c: columnIndex,
        r: rowIndex,
      });
      const cell = worksheet[cellAddress];
      if (!cell) {
        continue;
      }

      if (format === "date") {
        if (cell.t === "d" || cell.t === "n") {
          cell.z = getSpreadsheetNumberFormat(format);
        }
        continue;
      }

      if (cell.t === "n") {
        cell.z = getSpreadsheetNumberFormat(format);
      }
    }
  });
}

function buildDashboardWorkbookSheets({
  dashboardData,
  selectedMonth,
  selectedSalespersonLabel,
  selectedYear,
  sourceCompany,
}: {
  dashboardData: {
    conversionRate: number;
    lostBudgets: number;
    metricCards: DashboardMetricCard[];
    monthlyEvolution: DashboardMonthlyEvolutionItem[];
    negotiationBudgets: number;
    negotiationPipeline: DashboardSalespersonSummary[];
    recentSalespeople: DashboardSalespersonSummary[];
    salespersonFunnel: DashboardSalespersonFunnelItem[];
    staleBudgets: DashboardStaleBudgetItem[];
    averageClosingTimes: Array<{
      averageClosingDays: number;
      budgetCount: number;
      grossValue: number;
      label: string;
    }>;
    topConstructionCompanies: Array<{
      budgetCount: number;
      conversionRate: number;
      grossValue: number;
      label: string;
      lastActivityAt: string | null;
      lostBudgetCount: number;
      valueConversionRate: number;
      wonBudgetCount: number;
      wonGrossValue: number;
    }>;
    topLossReasons: Array<{
      averageTicket: number;
      grossValue: number;
      label: string;
      lostBudgetCount: number;
    }>;
    topProjects: Array<{
      budgetCount: number;
      conversionRate: number;
      grossValue: number;
      label: string;
      lastActivityAt: string | null;
      lostBudgetCount: number;
      valueConversionRate: number;
      wonBudgetCount: number;
      wonGrossValue: number;
    }>;
    topSalespeopleByAverageTicket: DashboardSalespersonSummary[];
    topSalespeopleByBudgetCount: DashboardSalespersonSummary[];
    topSalespeopleByConversion: DashboardSalespersonSummary[];
    topSalespeopleByValue: DashboardSalespersonSummary[];
    technicalOverview: {
      recentEstimators: DashboardEstimatorSummary[];
      summary: {
        activeEstimators: number;
        averageTicket: number;
        budgetsWithEstimator: number;
        budgetsWithoutEstimator: number;
        conversionRate: number;
        coverageRate: number;
        lostBudgets: number;
        negotiationBudgets: number;
        stalledBudgetsCount: number;
        totalGrossValue: number;
        totalNegotiationGrossValue: number;
        wonBudgets: number;
      };
      topEstimatorsByAverageTicket: DashboardEstimatorSummary[];
      topEstimatorsByBudgetCount: DashboardEstimatorSummary[];
      topEstimatorsByValue: DashboardEstimatorSummary[];
    };
    valueConversionRate: number;
    wonBudgets: number;
  };
  selectedMonth: string;
  selectedSalespersonLabel: string;
  selectedYear: string;
  sourceCompany: DashboardCompanyFilter;
}): SpreadsheetSheet[] {
  const scopeLabel = buildDashboardScopeLabel(
    sourceCompany,
    selectedYear,
    selectedMonth,
    selectedSalespersonLabel,
  );
  const topSalespersonByValue = dashboardData.topSalespeopleByValue[0];
  const topSalespersonByBudgetCount =
    dashboardData.topSalespeopleByBudgetCount[0];
  const topSalespersonByConversion =
    dashboardData.topSalespeopleByConversion[0];
  const topSalespersonByAverageTicket =
    dashboardData.topSalespeopleByAverageTicket[0];
  const topConstructionCompany = dashboardData.topConstructionCompanies[0];
  const topProject = dashboardData.topProjects[0];
  const topLossReason = dashboardData.topLossReasons[0];
  const averageClosingTime = dashboardData.averageClosingTimes[0];
  const mostRecentSalesperson = dashboardData.recentSalespeople[0];
  const mostStalledBudget = dashboardData.staleBudgets[0];
  const topEstimatorByValue =
    dashboardData.technicalOverview.topEstimatorsByValue[0];
  const topEstimatorByBudgetCount =
    dashboardData.technicalOverview.topEstimatorsByBudgetCount[0];
  const topEstimatorByAverageTicket =
    dashboardData.technicalOverview.topEstimatorsByAverageTicket[0];
  const mostRecentEstimator =
    dashboardData.technicalOverview.recentEstimators[0];
  const mostStalledSalesperson = [...dashboardData.negotiationPipeline].sort(
    (firstItem, secondItem) => {
      if (firstItem.stalledBudgetCount !== secondItem.stalledBudgetCount) {
        return secondItem.stalledBudgetCount - firstItem.stalledBudgetCount;
      }

      return secondItem.negotiationGrossValue - firstItem.negotiationGrossValue;
    },
  )[0];
  const currentMonth = dashboardData.monthlyEvolution.at(-1);
  const previousMonth =
    dashboardData.monthlyEvolution.length > 1
      ? dashboardData.monthlyEvolution.at(-2)
      : undefined;
  const currentConversionRate =
    currentMonth && currentMonth.budgetCount > 0
      ? (currentMonth.wonBudgetCount / currentMonth.budgetCount) * 100
      : 0;
  const previousConversionRate =
    previousMonth && previousMonth.budgetCount > 0
      ? (previousMonth.wonBudgetCount / previousMonth.budgetCount) * 100
      : 0;
  const budgetDelta =
    currentMonth && previousMonth
      ? currentMonth.budgetCount - previousMonth.budgetCount
      : null;
  const grossValueDelta =
    currentMonth && previousMonth
      ? currentMonth.grossValue - previousMonth.grossValue
      : null;
  const conversionDelta =
    currentMonth && previousMonth
      ? currentConversionRate - previousConversionRate
      : null;

  return [
    {
      headerStyle: "executive",
      name: "Resumo Executivo",
      mergeRanges: ["A1:C1", "A5:C5", "A13:C13", "A24:B24"],
      rows: [
        ["Dashboard administrativo"],
        ["Escopo", scopeLabel],
        ["Gerado em", new Date()],
        [],
        ["Indicadores principais"],
        ["Indicador", "Valor", "Descricao"],
        ...dashboardData.metricCards.map((metric) => [
          metric.label,
          metric.value,
          metric.helper,
        ]),
        [],
        ["Destaques"],
        [
          "Top vendedor por valor",
          topSalespersonByValue?.label ?? "Nao informado",
        ],
        ["Maior valor bruto", topSalespersonByValue?.grossValue ?? null],
        [
          "Melhor conversao por valor",
          `${formatPercentage(dashboardData.valueConversionRate)}`,
        ],
        [
          "Ultima atividade comercial",
          topSalespersonByValue
            ? toSpreadsheetDateOrFallback(
                topSalespersonByValue.lastActivityAt,
                "Nao informada",
              )
            : "Nao informada",
        ],
        [
          "Vendedor com atividade mais recente",
          mostRecentSalesperson?.label ?? "Nao informado",
        ],
        [
          "Orcamento mais parado",
          mostStalledBudget?.budgetNumber ?? "Nao informado",
        ],
        [
          "Dias parados do caso mais antigo",
          mostStalledBudget?.stalledDays ?? null,
        ],
        [],
        ["Novos estrategicos"],
        ["Construtora lider", topConstructionCompany?.label ?? "Nao informado"],
        ["Obra lider", topProject?.label ?? "Nao informado"],
        ["Principal motivo de perda", topLossReason?.label ?? "Nao informado"],
        [
          "Tempo medio de fechamento",
          averageClosingTime
            ? formatClosingDays(averageClosingTime.averageClosingDays)
            : "Nao informado",
        ],
        [],
        ["Resumo da carteira", "Valor"],
        ["Pedidos", dashboardData.wonBudgets],
        ["Em negociacao", dashboardData.negotiationBudgets],
        ["Cancelados", dashboardData.lostBudgets],
        ["Conversao", formatPercentage(dashboardData.conversionRate)],
      ],
      sectionRowIndexes: [4, 12, 23],
      formatsByColumn: {
        1: "currency",
      },
      titleRowIndex: 0,
    },
    {
      headerStyle: "executive",
      mergeRanges: ["A1:C1", "A5:C5", "A18:C18"],
      name: "Analise Comercial",
      rows: [
        ["Analise comercial automatica"],
        ["Escopo", scopeLabel],
        ["Gerado em", new Date()],
        [],
        ["Rankings e destaques"],
        [
          "Maior valor bruto",
          topSalespersonByValue?.label ?? "Nao informado",
          topSalespersonByValue?.grossValue ?? null,
        ],
        [
          "Maior quantidade de orcamentos",
          topSalespersonByBudgetCount?.label ?? "Nao informado",
          topSalespersonByBudgetCount?.budgetCount ?? null,
        ],
        [
          "Melhor conversao",
          topSalespersonByConversion?.label ?? "Nao informado",
          topSalespersonByConversion !== undefined
            ? topSalespersonByConversion.conversionRate / 100
            : null,
        ],
        [
          "Maior ticket medio",
          topSalespersonByAverageTicket?.label ?? "Nao informado",
          topSalespersonByAverageTicket?.averageTicket ?? null,
        ],
        [
          "Construtora com maior valor",
          topConstructionCompany?.label ?? "Nao informado",
          topConstructionCompany?.grossValue ?? null,
        ],
        [
          "Obra com maior valor",
          topProject?.label ?? "Nao informado",
          topProject?.grossValue ?? null,
        ],
        [
          "Principal motivo de perda",
          topLossReason?.label ?? "Nao informado",
          topLossReason?.grossValue ?? null,
        ],
        [
          "Tempo medio de fechamento",
          averageClosingTime?.label ?? "Nao informado",
          averageClosingTime?.averageClosingDays ?? null,
        ],
        [
          "Atividade comercial mais recente",
          mostRecentSalesperson?.label ?? "Nao informado",
          mostRecentSalesperson
            ? toSpreadsheetDateOrFallback(
                mostRecentSalesperson.lastActivityAt,
                "Nao informada",
              )
            : "Nao informada",
        ],
        [
          "Maior carteira parada",
          mostStalledSalesperson?.label ?? "Nao informado",
          mostStalledSalesperson?.stalledBudgetCount ?? null,
        ],
        [
          "Orcamento mais antigo sem atividade",
          mostStalledBudget?.budgetNumber ?? "Nao informado",
          mostStalledBudget?.stalledDays ?? null,
        ],
        [],
        ["Comparacao mensal"],
        currentMonth && previousMonth
          ? [
              "Periodo comparado",
              `${currentMonth.monthLabel} vs ${previousMonth.monthLabel}`,
            ]
          : ["Periodo comparado", "Base insuficiente"],
        [
          "Orcamentos no mes atual",
          currentMonth?.budgetCount ?? null,
          currentMonth?.monthLabel ?? "Nao informado",
        ],
        [
          "Valor bruto no mes atual",
          currentMonth?.grossValue ?? null,
          currentMonth?.monthLabel ?? "Nao informado",
        ],
        [
          "Conversao no mes atual",
          currentMonth ? currentConversionRate / 100 : null,
          currentMonth?.monthLabel ?? "Nao informado",
        ],
        [
          "Variacao de orcamentos",
          budgetDelta,
          budgetDelta === null
            ? "Base insuficiente"
            : formatSignedInteger(budgetDelta),
        ],
        [
          "Variacao de valor bruto",
          grossValueDelta,
          grossValueDelta === null
            ? "Base insuficiente"
            : formatSignedCurrency(grossValueDelta),
        ],
        [
          "Variacao de conversao",
          conversionDelta === null ? null : conversionDelta / 100,
          conversionDelta === null
            ? "Base insuficiente"
            : formatSignedPercentagePoints(conversionDelta),
        ],
      ],
      sectionRowIndexes: [4, 17],
      titleRowIndex: 0,
    },
    {
      autoFilter: true,
      formatsByColumn: {
        0: "integer",
        3: "currency",
        4: "currency",
        5: "date",
      },
      headerRowIndex: 0,
      headerStyle: "table",
      name: "Top por Valor",
      rows: [
        [
          "Posicao",
          "Vendedor",
          "Orcamentos",
          "Valor bruto",
          "Ticket medio",
          "Ultima atividade comercial",
        ],
        ...dashboardData.topSalespeopleByValue.map((item, index) => [
          index + 1,
          item.label,
          item.budgetCount,
          item.grossValue,
          item.averageTicket,
          toSpreadsheetDateOrFallback(item.lastActivityAt, "Nao informada"),
        ]),
      ],
    },
    {
      autoFilter: true,
      formatsByColumn: {
        0: "integer",
        3: "currency",
        4: "percent",
        5: "date",
      },
      headerRowIndex: 0,
      headerStyle: "table",
      name: "Top por Quantidade",
      rows: [
        [
          "Posicao",
          "Vendedor",
          "Orcamentos",
          "Valor bruto",
          "Conversao",
          "Ultima atividade comercial",
        ],
        ...dashboardData.topSalespeopleByBudgetCount.map((item, index) => [
          index + 1,
          item.label,
          item.budgetCount,
          item.grossValue,
          item.conversionRate / 100,
          toSpreadsheetDateOrFallback(item.lastActivityAt, "Nao informada"),
        ]),
      ],
    },
    {
      autoFilter: true,
      formatsByColumn: {
        0: "integer",
        2: "percent",
        5: "currency",
      },
      headerRowIndex: 0,
      headerStyle: "table",
      name: "Top Conversao",
      rows: [
        [
          "Posicao",
          "Vendedor",
          "Conversao",
          "Pedidos",
          "Orcamentos",
          "Ticket medio",
        ],
        ...dashboardData.topSalespeopleByConversion.map((item, index) => [
          index + 1,
          item.label,
          item.conversionRate / 100,
          item.wonBudgetCount,
          item.budgetCount,
          item.averageTicket,
        ]),
      ],
    },
    {
      autoFilter: true,
      formatsByColumn: {
        0: "integer",
        2: "currency",
        4: "currency",
        5: "percent",
      },
      headerRowIndex: 0,
      headerStyle: "table",
      name: "Top Ticket Medio",
      rows: [
        [
          "Posicao",
          "Vendedor",
          "Ticket medio",
          "Orcamentos",
          "Valor bruto",
          "Conversao",
        ],
        ...dashboardData.topSalespeopleByAverageTicket.map((item, index) => [
          index + 1,
          item.label,
          item.averageTicket,
          item.budgetCount,
          item.grossValue,
          item.conversionRate / 100,
        ]),
      ],
    },
    {
      autoFilter: true,
      formatsByColumn: {
        2: "currency",
        4: "date",
      },
      headerRowIndex: 0,
      headerStyle: "table",
      name: "Carteira Negociacao",
      rows: [
        [
          "Vendedor",
          "Em aberto",
          "Valor em negociacao",
          "Orcamentos parados",
          "Ultima atividade comercial",
        ],
        ...dashboardData.negotiationPipeline.map((item) => [
          item.label,
          item.negotiationBudgetCount,
          item.negotiationGrossValue,
          item.stalledBudgetCount,
          toSpreadsheetDateOrFallback(item.lastActivityAt, "Nao informada"),
        ]),
      ],
    },
    {
      autoFilter: true,
      formatsByColumn: {
        2: "currency",
        3: "date",
      },
      headerRowIndex: 0,
      headerStyle: "table",
      name: "Ult Atividade",
      rows: [
        ["Vendedor", "Orcamentos", "Valor bruto", "Ultima atividade comercial"],
        ...dashboardData.recentSalespeople.map((item) => [
          item.label,
          item.budgetCount,
          item.grossValue,
          toSpreadsheetDateOrFallback(item.lastActivityAt, "Nao informada"),
        ]),
      ],
    },
    {
      autoFilter: true,
      formatsByColumn: {
        5: "currency",
        6: "integer",
        7: "date",
      },
      headerRowIndex: 0,
      headerStyle: "table",
      name: "Orcamentos Parados",
      rows: [
        [
          "Orcamento",
          "Vendedor",
          "Obra",
          "Construtora",
          "Status",
          "Valor bruto",
          "Dias parados",
          "Ultima atividade comercial",
        ],
        ...dashboardData.staleBudgets.map((item) => [
          item.budgetNumber,
          item.salespersonLabel,
          item.projectLabel,
          item.constructionCompanyLabel,
          item.statusLabel,
          item.grossValue,
          item.stalledDays,
          toSpreadsheetDateOrFallback(item.lastActivityAt, "Nao informada"),
        ]),
      ],
    },
    {
      autoFilter: true,
      formatsByColumn: {
        2: "currency",
        4: "currency",
      },
      headerRowIndex: 0,
      headerStyle: "table",
      name: "Evolucao Mensal",
      rows: [
        ["Mes", "Orcamentos", "Valor bruto", "Pedidos", "Valor convertido"],
        ...dashboardData.monthlyEvolution.map((item) => [
          item.monthLabel,
          item.budgetCount,
          item.grossValue,
          item.wonBudgetCount,
          item.wonGrossValue,
        ]),
      ],
    },
    {
      autoFilter: true,
      formatsByColumn: {
        2: "currency",
        3: "percent",
        4: "percent",
        5: "date",
      },
      headerRowIndex: 0,
      headerStyle: "table",
      name: "Top Construtoras",
      rows: [
        [
          "Construtora",
          "Orcamentos",
          "Valor bruto",
          "Conversao",
          "Conversao por valor",
          "Ultima atividade comercial",
        ],
        ...dashboardData.topConstructionCompanies.map((item) => [
          item.label,
          item.budgetCount,
          item.grossValue,
          item.conversionRate / 100,
          item.valueConversionRate / 100,
          toSpreadsheetDateOrFallback(item.lastActivityAt, "Nao informada"),
        ]),
      ],
    },
    {
      autoFilter: true,
      formatsByColumn: {
        2: "currency",
        3: "percent",
        4: "percent",
        5: "date",
      },
      headerRowIndex: 0,
      headerStyle: "table",
      name: "Top Obras",
      rows: [
        [
          "Obra",
          "Orcamentos",
          "Valor bruto",
          "Conversao",
          "Conversao por valor",
          "Ultima atividade comercial",
        ],
        ...dashboardData.topProjects.map((item) => [
          item.label,
          item.budgetCount,
          item.grossValue,
          item.conversionRate / 100,
          item.valueConversionRate / 100,
          toSpreadsheetDateOrFallback(item.lastActivityAt, "Nao informada"),
        ]),
      ],
    },
    {
      autoFilter: true,
      formatsByColumn: {
        2: "currency",
        3: "currency",
      },
      headerRowIndex: 0,
      headerStyle: "table",
      name: "Motivos Perda",
      rows: [
        ["Motivo de perda", "Cancelados", "Valor perdido", "Ticket medio"],
        ...dashboardData.topLossReasons.map((item) => [
          item.label,
          item.lostBudgetCount,
          item.grossValue,
          item.averageTicket,
        ]),
      ],
    },
    {
      autoFilter: true,
      formatsByColumn: {
        2: "currency",
        3: "integer",
      },
      headerRowIndex: 0,
      headerStyle: "table",
      name: "Tempo Fechamento",
      rows: [
        ["Recorte", "Tempo medio", "Valor bruto", "Orcamentos fechados"],
        ...dashboardData.averageClosingTimes.map((item) => [
          item.label,
          formatClosingDays(item.averageClosingDays),
          item.grossValue,
          item.budgetCount,
        ]),
      ],
    },
    {
      autoFilter: true,
      formatsByColumn: {
        5: "percent",
      },
      headerRowIndex: 0,
      headerStyle: "table",
      name: "Funil Vendedor",
      rows: [
        [
          "Vendedor",
          "Total de orcamentos",
          "Em negociacao",
          "Pedidos",
          "Cancelados",
          "Conversao",
        ],
        ...dashboardData.salespersonFunnel.map((item) => [
          item.label,
          item.totalBudgets,
          item.negotiationBudgets,
          item.wonBudgets,
          item.lostBudgets,
          item.conversionRate / 100,
        ]),
      ],
    },
    {
      headerStyle: "executive",
      mergeRanges: ["A1:C1", "A5:C5", "A15:B15"],
      name: "Resumo Tecnico",
      rows: [
        ["Visao tecnica por orcamentista"],
        ["Escopo", scopeLabel],
        ["Gerado em", new Date()],
        [],
        ["Indicadores tecnicos"],
        ["Indicador", "Valor", "Descricao"],
        [
          "Orcamentistas ativos",
          dashboardData.technicalOverview.summary.activeEstimators,
          "Quantidade de orcamentistas com producao no recorte",
        ],
        [
          "Cobertura tecnica",
          dashboardData.technicalOverview.summary.coverageRate / 100,
          "Percentual de orcamentos com orcamentista atribuido",
        ],
        [
          "Orcamentos com orcamentista",
          dashboardData.technicalOverview.summary.budgetsWithEstimator,
          "Quantidade com responsabilidade tecnica definida",
        ],
        [
          "Orcamentos sem orcamentista",
          dashboardData.technicalOverview.summary.budgetsWithoutEstimator,
          "Quantidade ainda sem atribuicao tecnica",
        ],
        [
          "Valor tecnico monitorado",
          dashboardData.technicalOverview.summary.totalGrossValue,
          "Valor bruto dos orcamentos com orcamentista atribuido",
        ],
        [
          "Conversao tecnica",
          dashboardData.technicalOverview.summary.conversionRate / 100,
          "Percentual de pedidos dentro da carteira tecnica atribuida",
        ],
        [],
        ["Destaques tecnicos", "Valor"],
        [
          "Top orcamentista por valor",
          topEstimatorByValue?.label ?? "Nao informado",
        ],
        [
          "Top orcamentista por quantidade",
          topEstimatorByBudgetCount?.label ?? "Nao informado",
        ],
        [
          "Top ticket medio tecnico",
          topEstimatorByAverageTicket?.label ?? "Nao informado",
        ],
        [
          "Ultima atividade tecnica",
          mostRecentEstimator?.label ?? "Nao informado",
        ],
      ],
      sectionRowIndexes: [4, 14],
      formatsByColumn: {
        1: "currency",
        2: "percent",
      },
      titleRowIndex: 0,
    },
    {
      autoFilter: true,
      formatsByColumn: {
        0: "integer",
        3: "currency",
        4: "currency",
        5: "date",
      },
      headerRowIndex: 0,
      headerStyle: "table",
      name: "Top Orcam Valor",
      rows: [
        [
          "Posicao",
          "Orcamentista",
          "Orcamentos",
          "Valor bruto",
          "Ticket medio",
          "Ultima atividade",
        ],
        ...dashboardData.technicalOverview.topEstimatorsByValue.map(
          (item, index) => [
            index + 1,
            item.label,
            item.budgetCount,
            item.grossValue,
            item.averageTicket,
            toSpreadsheetDateOrFallback(item.lastActivityAt, "Nao informada"),
          ],
        ),
      ],
    },
    {
      autoFilter: true,
      formatsByColumn: {
        0: "integer",
        3: "currency",
        4: "percent",
        5: "date",
      },
      headerRowIndex: 0,
      headerStyle: "table",
      name: "Top Orcam Quant",
      rows: [
        [
          "Posicao",
          "Orcamentista",
          "Orcamentos",
          "Valor bruto",
          "Conversao",
          "Ultima atividade",
        ],
        ...dashboardData.technicalOverview.topEstimatorsByBudgetCount.map(
          (item, index) => [
            index + 1,
            item.label,
            item.budgetCount,
            item.grossValue,
            item.conversionRate / 100,
            toSpreadsheetDateOrFallback(item.lastActivityAt, "Nao informada"),
          ],
        ),
      ],
    },
    {
      autoFilter: true,
      formatsByColumn: {
        2: "currency",
        3: "currency",
        4: "date",
      },
      headerRowIndex: 0,
      headerStyle: "table",
      name: "Ult Ativ Tecnica",
      rows: [
        [
          "Orcamentista",
          "Orcamentos",
          "Valor bruto",
          "Valor em negociacao",
          "Ultima atividade",
        ],
        ...dashboardData.technicalOverview.recentEstimators.map((item) => [
          item.label,
          item.budgetCount,
          item.grossValue,
          item.negotiationGrossValue,
          toSpreadsheetDateOrFallback(item.lastActivityAt, "Nao informada"),
        ]),
      ],
    },
  ];
}

export function DashboardPage() {
  const { user } = useAuth();
  const navigate = useNavigate();
  const isEstimatorUser =
    user?.role === "user" && user.user_kind === "estimator";
  const [sourceCompany, setSourceCompany] =
    useState<DashboardCompanyFilter>("");
  const [selectedYear, setSelectedYear] = useState("");
  const [selectedMonth, setSelectedMonth] = useState("");
  const [selectedSalespersonId, setSelectedSalespersonId] = useState("");

  const dashboardFilters = useMemo<DashboardSalespeopleFilters>(
    () => ({
      sourceCompany,
      salespersonId: selectedSalespersonId,
      year: selectedYear,
      month: selectedMonth,
    }),
    [selectedMonth, selectedSalespersonId, selectedYear, sourceCompany],
  );

  const budgetCatalogsQuery = useQuery({
    queryKey: ["dashboard", "budget-catalogs"],
    queryFn: getBudgetCatalogsRequest,
    enabled: !isEstimatorUser,
    gcTime: 1000 * 60 * 15,
    refetchOnWindowFocus: false,
    staleTime: 1000 * 60 * 5,
  });

  const dashboardQuery = useQuery({
    queryKey: ["dashboard", "salespeople", dashboardFilters],
    queryFn: () => getSalespeopleDashboardRequest(dashboardFilters),
    enabled: !isEstimatorUser,
    gcTime: 1000 * 60 * 10,
    refetchOnWindowFocus: false,
    retry: (failureCount, error) => {
      if (error instanceof AxiosError) {
        const statusCode = error.response?.status;
        if (statusCode === 401 || statusCode === 403) {
          return false;
        }
      }

      return failureCount < 1;
    },
    staleTime: 1000 * 60 * 2,
  });

  if (isEstimatorUser) {
    return (
      <Box sx={{ display: "flex", flexDirection: "column", gap: 3 }}>
        <PageHeader
          description="O dashboard comercial fica disponivel apenas para administradores e perfis comerciais."
          title="Dashboard"
        />
        <Alert severity="info">
          O perfil orcamentista nao participa do dashboard comercial. Utilize a
          tela de orcamentos para operar no seu escopo tecnico.
        </Alert>
      </Box>
    );
  }

  const availableYears = useMemo(() => {
    return dashboardQuery.data?.availableYears ?? [];
  }, [dashboardQuery.data?.availableYears]);

  useEffect(() => {
    if (selectedYear && !availableYears.includes(selectedYear)) {
      setSelectedYear("");
      setSelectedMonth("");
    }
  }, [availableYears, selectedYear]);

  const salespersonOptions = useMemo<BudgetCatalogItem[]>(() => {
    return [...(budgetCatalogsQuery.data?.salespeople ?? [])].sort((a, b) =>
      a.name.localeCompare(b.name),
    );
  }, [budgetCatalogsQuery.data?.salespeople]);

  const selectedSalespersonLabel = useMemo(() => {
    return (
      salespersonOptions.find(
        (item) => String(item.id) === selectedSalespersonId,
      )?.name ?? ""
    );
  }, [selectedSalespersonId, salespersonOptions]);

  const salespersonIdByName = useMemo(() => {
    return salespersonOptions.reduce<Map<string, string>>(
      (currentMap, item) => {
        currentMap.set(item.name, String(item.id));
        return currentMap;
      },
      new Map<string, string>(),
    );
  }, [salespersonOptions]);

  const budgetItems = useMemo(() => {
    return dashboardQuery.data;
  }, [dashboardQuery.data]);

  const statusIdByNormalizedName = useMemo(() => {
    return (budgetCatalogsQuery.data?.statuses ?? []).reduce<
      Map<string, string>
    >((currentMap, item) => {
      currentMap.set(normalizeLookupKey(item.name), String(item.id));
      return currentMap;
    }, new Map<string, string>());
  }, [budgetCatalogsQuery.data?.statuses]);

  const dashboardData = useMemo(() => {
    const summary = budgetItems?.summary;
    const technicalOverview = budgetItems?.technicalOverview;
    const monthlyEvolution = budgetItems?.monthlyEvolution ?? [];
    const maxMonthlyGrossValue = monthlyEvolution.reduce(
      (currentMax, item) => Math.max(currentMax, item.grossValue),
      0,
    );

    const metricCards: DashboardMetricCard[] = [
      {
        key: "active-salespeople",
        label: "Vendedores ativos",
        value: String(summary?.activeSalespeople ?? 0),
        helper: "Quantidade de vendedores com orcamentos no recorte atual",
        icon: InsightsRoundedIcon,
      },
      {
        key: "total-budgets",
        label: "Orcamentos monitorados",
        value: String(summary?.totalBudgets ?? 0),
        helper: "Volume total de orcamentos no recorte atual",
        icon: DescriptionRoundedIcon,
      },
      {
        key: "gross-value",
        label: "Valor bruto total",
        value: formatCompactCurrency(summary?.totalGrossValue ?? 0),
        helper: "Soma do valor bruto da carteira comercial",
        icon: AttachMoneyRoundedIcon,
      },
      {
        key: "ticket-medio",
        label: "Ticket medio",
        value: formatCompactCurrency(summary?.averageTicket ?? 0),
        helper: "Media de valor por orcamento do periodo",
        icon: TrendingUpRoundedIcon,
      },
      {
        key: "negotiation-gross-value",
        label: "Valor em negociacao",
        value: formatCompactCurrency(summary?.totalNegotiationGrossValue ?? 0),
        helper: `${summary?.negotiationBudgets ?? 0} orcamento(s) ainda em carteira`,
        icon: AttachMoneyRoundedIcon,
      },
      {
        key: "pedido-conversion",
        label: "Conversao em pedido",
        value: formatPercentage(summary?.conversionRate ?? 0),
        helper: `${summary?.wonBudgets ?? 0} pedido(s) em ${summary?.totalBudgets ?? 0} orcamento(s)`,
        icon: AssignmentTurnedInRoundedIcon,
      },
      {
        key: "value-conversion",
        label: "Conversao por valor",
        value: formatPercentage(summary?.valueConversionRate ?? 0),
        helper: `${formatCompactCurrency(summary?.totalGrossValue ?? 0)} em valor bruto no periodo`,
        icon: AttachMoneyRoundedIcon,
      },
      {
        key: "stalled-budgets",
        label: "Orcamentos parados",
        value: String(summary?.stalledBudgetsCount ?? 0),
        helper: "Oportunidades em negociacao sem atividade ha 7 dias ou mais",
        icon: DescriptionRoundedIcon,
      },
    ];
    const technicalMetricCards: DashboardMetricCard[] = [
      {
        key: "active-estimators",
        label: "Orcamentistas ativos",
        value: String(technicalOverview?.summary.activeEstimators ?? 0),
        helper: "Quantidade de orcamentistas com producao no recorte atual",
        icon: InsightsRoundedIcon,
      },
      {
        key: "coverage-rate",
        label: "Cobertura tecnica",
        value: formatPercentage(technicalOverview?.summary.coverageRate ?? 0),
        helper: `${technicalOverview?.summary.budgetsWithEstimator ?? 0} orcamento(s) com orcamentista definido`,
        icon: AssignmentTurnedInRoundedIcon,
      },
      {
        key: "technical-gross-value",
        label: "Valor tecnico monitorado",
        value: formatCompactCurrency(
          technicalOverview?.summary.totalGrossValue ?? 0,
        ),
        helper:
          "Valor bruto dos orcamentos com responsabilidade tecnica atribuida",
        icon: AttachMoneyRoundedIcon,
      },
      {
        key: "technical-stalled-budgets",
        label: "Orcamentos tecnicos parados",
        value: String(technicalOverview?.summary.stalledBudgetsCount ?? 0),
        helper:
          "Orcamentos em negociacao com orcamentista e sem atividade ha 7 dias ou mais",
        icon: DescriptionRoundedIcon,
      },
    ];

    return {
      averageClosingTimes: budgetItems?.averageClosingTimes ?? [],
      conversionRate: summary?.conversionRate ?? 0,
      lostBudgets: summary?.lostBudgets ?? 0,
      maxMonthlyGrossValue,
      metricCards,
      monthlyEvolution,
      negotiationBudgets: summary?.negotiationBudgets ?? 0,
      negotiationPipeline: budgetItems?.negotiationPipeline ?? [],
      recentSalespeople: budgetItems?.recentSalespeople ?? [],
      salespersonFunnel: budgetItems?.salespersonFunnel ?? [],
      staleBudgets: budgetItems?.staleBudgets ?? [],
      topConstructionCompanies: budgetItems?.topConstructionCompanies ?? [],
      topLossReasons: budgetItems?.topLossReasons ?? [],
      topProjects: budgetItems?.topProjects ?? [],
      topSalespeopleByAverageTicket:
        budgetItems?.topSalespeopleByAverageTicket ?? [],
      topSalespeopleByBudgetCount:
        budgetItems?.topSalespeopleByBudgetCount ?? [],
      topSalespeopleByConversion: budgetItems?.topSalespeopleByConversion ?? [],
      topSalespeopleByValue: budgetItems?.topSalespeopleByValue ?? [],
      technicalMetricCards,
      technicalOverview: budgetItems?.technicalOverview ?? {
        recentEstimators: [],
        summary: {
          activeEstimators: 0,
          averageTicket: 0,
          budgetsWithEstimator: 0,
          budgetsWithoutEstimator: 0,
          conversionRate: 0,
          coverageRate: 0,
          lostBudgets: 0,
          negotiationBudgets: 0,
          stalledBudgetsCount: 0,
          totalGrossValue: 0,
          totalNegotiationGrossValue: 0,
          wonBudgets: 0,
        },
        topEstimatorsByAverageTicket: [],
        topEstimatorsByBudgetCount: [],
        topEstimatorsByValue: [],
      },
      valueConversionRate: summary?.valueConversionRate ?? 0,
      wonBudgets: summary?.wonBudgets ?? 0,
    };
  }, [budgetItems]);

  const dashboardInsights = useMemo(() => {
    const monthlyEvolution = dashboardData.monthlyEvolution;
    if (monthlyEvolution.length < 2) {
      return {
        monthComparison: null,
      };
    }

    const currentMonth = monthlyEvolution[monthlyEvolution.length - 1];
    const previousMonth = monthlyEvolution[monthlyEvolution.length - 2];
    const currentConversionRate =
      currentMonth.budgetCount === 0
        ? 0
        : (currentMonth.wonBudgetCount / currentMonth.budgetCount) * 100;
    const previousConversionRate =
      previousMonth.budgetCount === 0
        ? 0
        : (previousMonth.wonBudgetCount / previousMonth.budgetCount) * 100;

    return {
      monthComparison: {
        currentMonthLabel: formatMonthKeyLabel(currentMonth.monthKey),
        previousMonthLabel: formatMonthKeyLabel(previousMonth.monthKey),
        currentMonth,
        previousMonth,
        budgetDelta: currentMonth.budgetCount - previousMonth.budgetCount,
        grossValueDelta: currentMonth.grossValue - previousMonth.grossValue,
        conversionDelta: currentConversionRate - previousConversionRate,
      },
    };
  }, [dashboardData.monthlyEvolution]);

  const dashboardErrorMessage = useMemo(() => {
    return getDashboardErrorMessage(dashboardQuery.error);
  }, [dashboardQuery.error]);

  const isDashboardRefreshing =
    dashboardQuery.isFetching && !dashboardQuery.isLoading;

  const handleOpenBudgetList = ({
    budgetNumber,
    salespersonId,
    statusId,
  }: Partial<BudgetListNavigationOptions> = {}) => {
    const searchParams = buildBudgetListSearchParams({
      budgetNumber,
      month: selectedMonth,
      salespersonId: salespersonId ?? selectedSalespersonId,
      sourceCompany,
      statusId,
      year: selectedYear,
    });

    navigate(`/budgets?${searchParams.toString()}`);
  };

  const handleExportDashboardCsv = async () => {
    const csvLines = [
      createCsvLine([
        "Escopo",
        buildDashboardScopeLabel(
          sourceCompany,
          selectedYear,
          selectedMonth,
          selectedSalespersonLabel,
        ),
      ]),
      "",
      createCsvLine([
        "Resumo",
        "Vendedores ativos",
        "Orcamentos monitorados",
        "Valor bruto total",
        "Ticket medio",
        "Valor em negociacao",
        "Conversao",
        "Conversao por valor",
        "Orcamentos parados",
      ]),
      createCsvLine([
        "Resumo",
        dashboardData.metricCards[0]?.value ?? "0",
        dashboardData.metricCards[1]?.value ?? "0",
        dashboardData.metricCards[2]?.value ?? "R$ 0,00",
        dashboardData.metricCards[3]?.value ?? "R$ 0,00",
        dashboardData.metricCards[4]?.value ?? "R$ 0,00",
        dashboardData.metricCards[5]?.value ?? "0,0%",
        dashboardData.metricCards[6]?.value ?? "0,0%",
        dashboardData.metricCards[7]?.value ?? "0",
      ]),
      "",
      createCsvLine([
        "Top por valor",
        "Vendedor",
        "Orcamentos",
        "Valor bruto",
        "Ticket medio",
        "Ultima atividade comercial",
      ]),
      ...dashboardData.topSalespeopleByValue.map((item) =>
        createCsvLine([
          "Top por valor",
          item.label,
          item.budgetCount,
          currencyFormatter.format(item.grossValue),
          currencyFormatter.format(item.averageTicket),
          formatDateOrFallback(item.lastActivityAt, "Nao informada"),
        ]),
      ),
      "",
      createCsvLine([
        "Carteira em negociacao",
        "Vendedor",
        "Em aberto",
        "Valor",
        "Parados",
        "Ultima atividade comercial",
      ]),
      ...dashboardData.negotiationPipeline.map((item) =>
        createCsvLine([
          "Carteira em negociacao",
          item.label,
          item.negotiationBudgetCount,
          currencyFormatter.format(item.negotiationGrossValue),
          item.stalledBudgetCount,
          formatDateOrFallback(item.lastActivityAt, "Nao informada"),
        ]),
      ),
      "",
      createCsvLine([
        "Orcamentos parados",
        "Orcamento",
        "Vendedor",
        "Obra",
        "Construtora",
        "Valor",
        "Dias parados",
      ]),
      ...dashboardData.staleBudgets.map((item) =>
        createCsvLine([
          "Orcamentos parados",
          item.budgetNumber,
          item.salespersonLabel,
          item.projectLabel,
          item.constructionCompanyLabel,
          currencyFormatter.format(item.grossValue),
          item.stalledDays,
        ]),
      ),
      "",
      createCsvLine([
        "Evolucao mensal",
        "Mes",
        "Orcamentos",
        "Valor bruto",
        "Pedidos",
        "Valor convertido",
      ]),
      ...dashboardData.monthlyEvolution.map((item) =>
        createCsvLine([
          "Evolucao mensal",
          item.monthLabel,
          item.budgetCount,
          currencyFormatter.format(item.grossValue),
          item.wonBudgetCount,
          currencyFormatter.format(item.wonGrossValue),
        ]),
      ),
      "",
      createCsvLine([
        "Top construtoras",
        "Construtora",
        "Orcamentos",
        "Valor bruto",
        "Conversao",
        "Conversao por valor",
      ]),
      ...dashboardData.topConstructionCompanies.map((item) =>
        createCsvLine([
          "Top construtoras",
          item.label,
          item.budgetCount,
          currencyFormatter.format(item.grossValue),
          formatPercentage(item.conversionRate),
          formatPercentage(item.valueConversionRate),
        ]),
      ),
      "",
      createCsvLine([
        "Top obras",
        "Obra",
        "Orcamentos",
        "Valor bruto",
        "Conversao",
        "Conversao por valor",
      ]),
      ...dashboardData.topProjects.map((item) =>
        createCsvLine([
          "Top obras",
          item.label,
          item.budgetCount,
          currencyFormatter.format(item.grossValue),
          formatPercentage(item.conversionRate),
          formatPercentage(item.valueConversionRate),
        ]),
      ),
      "",
      createCsvLine([
        "Motivos de perda",
        "Motivo",
        "Cancelados",
        "Valor perdido",
        "Ticket medio",
      ]),
      ...dashboardData.topLossReasons.map((item) =>
        createCsvLine([
          "Motivos de perda",
          item.label,
          item.lostBudgetCount,
          currencyFormatter.format(item.grossValue),
          currencyFormatter.format(item.averageTicket),
        ]),
      ),
      "",
      createCsvLine([
        "Tempo medio de fechamento",
        "Recorte",
        "Tempo medio",
        "Valor bruto",
        "Orcamentos fechados",
      ]),
      ...dashboardData.averageClosingTimes.map((item) =>
        createCsvLine([
          "Tempo medio de fechamento",
          item.label,
          formatClosingDays(item.averageClosingDays),
          currencyFormatter.format(item.grossValue),
          item.budgetCount,
        ]),
      ),
      "",
      createCsvLine([
        "Resumo tecnico",
        "Orcamentistas ativos",
        "Cobertura tecnica",
        "Com orcamentista",
        "Sem orcamentista",
        "Valor tecnico",
        "Ticket medio",
        "Valor em negociacao",
        "Conversao tecnica",
        "Parados",
      ]),
      createCsvLine([
        "Resumo tecnico",
        dashboardData.technicalOverview.summary.activeEstimators,
        formatPercentage(dashboardData.technicalOverview.summary.coverageRate),
        dashboardData.technicalOverview.summary.budgetsWithEstimator,
        dashboardData.technicalOverview.summary.budgetsWithoutEstimator,
        currencyFormatter.format(
          dashboardData.technicalOverview.summary.totalGrossValue,
        ),
        currencyFormatter.format(
          dashboardData.technicalOverview.summary.averageTicket,
        ),
        currencyFormatter.format(
          dashboardData.technicalOverview.summary.totalNegotiationGrossValue,
        ),
        formatPercentage(
          dashboardData.technicalOverview.summary.conversionRate,
        ),
        dashboardData.technicalOverview.summary.stalledBudgetsCount,
      ]),
      "",
      createCsvLine([
        "Top orcamentistas por valor",
        "Orcamentista",
        "Orcamentos",
        "Valor bruto",
        "Ticket medio",
        "Ultima atividade",
      ]),
      ...dashboardData.technicalOverview.topEstimatorsByValue.map((item) =>
        createCsvLine([
          "Top orcamentistas por valor",
          item.label,
          item.budgetCount,
          currencyFormatter.format(item.grossValue),
          currencyFormatter.format(item.averageTicket),
          formatDateOrFallback(item.lastActivityAt, "Nao informada"),
        ]),
      ),
      "",
      createCsvLine([
        "Top orcamentistas por quantidade",
        "Orcamentista",
        "Orcamentos",
        "Valor bruto",
        "Conversao",
        "Ultima atividade",
      ]),
      ...dashboardData.technicalOverview.topEstimatorsByBudgetCount.map(
        (item) =>
          createCsvLine([
            "Top orcamentistas por quantidade",
            item.label,
            item.budgetCount,
            currencyFormatter.format(item.grossValue),
            formatPercentage(item.conversionRate),
            formatDateOrFallback(item.lastActivityAt, "Nao informada"),
          ]),
      ),
      "",
      createCsvLine([
        "Ultima atividade tecnica",
        "Orcamentista",
        "Orcamentos",
        "Valor bruto",
        "Valor em negociacao",
        "Ultima atividade",
      ]),
      ...dashboardData.technicalOverview.recentEstimators.map((item) =>
        createCsvLine([
          "Ultima atividade tecnica",
          item.label,
          item.budgetCount,
          currencyFormatter.format(item.grossValue),
          currencyFormatter.format(item.negotiationGrossValue),
          formatDateOrFallback(item.lastActivityAt, "Nao informada"),
        ]),
      ),
    ].join("\n");

    const csvBlob = new Blob([`\uFEFF${csvLines}`], {
      type: "text/csv;charset=utf-8;",
    });
    const downloadUrl = window.URL.createObjectURL(csvBlob);
    const downloadLink = document.createElement("a");
    const fileScope = [
      sourceCompany || "todas-empresas",
      selectedYear || "todos-anos",
      selectedMonth || "todos-meses",
      selectedSalespersonLabel || "todos-vendedores",
    ]
      .join("-")
      .replaceAll(" ", "-")
      .toLowerCase();
    const fileName = `dashboard-gerencial-${fileScope}.csv`;
    const saveWithPickerSucceeded = await saveBlobWithPicker(
      csvBlob,
      fileName,
      "text/csv",
      ".csv",
    );
    if (saveWithPickerSucceeded) {
      return;
    }

    downloadBlob(csvBlob, fileName);
  };

  const handleExportDashboardXlsx = async () => {
    const XLSX = await import("xlsx-js-style");
    const workbook = XLSX.utils.book_new();
    const fileScope = [
      sourceCompany || "todas-empresas",
      selectedYear || "todos-anos",
      selectedMonth || "todos-meses",
      selectedSalespersonLabel || "todos-vendedores",
    ]
      .join("-")
      .replaceAll(" ", "-")
      .toLowerCase();
    const sheets = buildDashboardWorkbookSheets({
      dashboardData,
      selectedMonth,
      selectedSalespersonLabel,
      selectedYear,
      sourceCompany,
    });

    for (const sheet of sheets) {
      const worksheet = XLSX.utils.aoa_to_sheet(sheet.rows, {
        cellDates: true,
      });
      await applyWorksheetFormatting(XLSX, worksheet, sheet);
      XLSX.utils.book_append_sheet(workbook, worksheet, sheet.name);
    }

    const workbookArray = XLSX.write(workbook, {
      bookType: "xlsx",
      cellStyles: true,
      type: "array",
    });
    const xlsxBlob = new Blob([workbookArray], {
      type: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
    });
    const fileName = `dashboard-gerencial-${fileScope}.xlsx`;
    const saveWithPickerSucceeded = await saveBlobWithPicker(
      xlsxBlob,
      fileName,
      "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
      ".xlsx",
    );
    if (saveWithPickerSucceeded) {
      return;
    }

    downloadBlob(xlsxBlob, fileName);
  };

  return (
    <Box sx={{ display: "flex", flexDirection: "column", gap: 3 }}>
      <PageHeader
        action={
          <Box
            sx={{
              display: "grid",
              gap: 1.5,
              gridTemplateColumns: {
                lg: "repeat(4, minmax(180px, 220px))",
                sm: "repeat(2, minmax(180px, 220px))",
                xs: "minmax(0, 1fr)",
              },
              width: { md: "auto", xs: "100%" },
            }}
          >
            <TextField
              label="Empresa"
              onChange={(event) =>
                setSourceCompany(event.target.value as DashboardCompanyFilter)
              }
              select
              size="small"
              value={sourceCompany}
            >
              <MenuItem value="">Todas as empresas</MenuItem>
              <MenuItem value="Rocktec">ROCKTEC</MenuItem>
              <MenuItem value="Trox">TROX</MenuItem>
            </TextField>
            <TextField
              label="Ano"
              onChange={(event) => setSelectedYear(event.target.value)}
              select
              size="small"
              value={selectedYear}
            >
              <MenuItem value="">Todos os anos</MenuItem>
              {availableYears.map((year) => (
                <MenuItem key={year} value={year}>
                  {year}
                </MenuItem>
              ))}
            </TextField>
            <TextField
              label="Mes"
              onChange={(event) => setSelectedMonth(event.target.value)}
              select
              size="small"
              value={selectedMonth}
            >
              <MenuItem value="">Todos os meses</MenuItem>
              {monthOptions.map((month) => (
                <MenuItem key={month.value} value={month.value}>
                  {month.label}
                </MenuItem>
              ))}
            </TextField>
            <TextField
              label="Vendedor"
              onChange={(event) => setSelectedSalespersonId(event.target.value)}
              select
              size="small"
              value={selectedSalespersonId}
            >
              <MenuItem value="">Todos os vendedores</MenuItem>
              {salespersonOptions.map((salesperson) => (
                <MenuItem key={salesperson.id} value={String(salesperson.id)}>
                  {salesperson.name}
                </MenuItem>
              ))}
            </TextField>
          </Box>
        }
        description="Painel administrativo com leitura comercial por vendedor e visao tecnica separada por orcamentista."
        title="Dashboard administrativo"
      />

      {dashboardQuery.isLoading || isDashboardRefreshing ? (
        <LinearProgress />
      ) : null}

      {dashboardQuery.isError ? (
        <Alert
          action={
            <Button
              color="inherit"
              onClick={() => void dashboardQuery.refetch()}
              size="small"
            >
              Tentar novamente
            </Button>
          }
          severity="error"
          variant="outlined"
        >
          {dashboardErrorMessage}
        </Alert>
      ) : null}

      {!dashboardQuery.isLoading && !dashboardQuery.isError ? (
        <Alert severity="info" variant="outlined">
          {`Os indicadores abaixo refletem o recorte de ${buildDashboardScopeLabel(
            sourceCompany,
            selectedYear,
            selectedMonth,
            selectedSalespersonLabel,
          )}.`}
          {isDashboardRefreshing ? " Atualizando dados em segundo plano." : ""}
        </Alert>
      ) : null}

      {!dashboardQuery.isLoading && !dashboardQuery.isError ? (
        <SectionCard
          description="Atalhos para aprofundar a analise do dashboard e exportar o recorte atual."
          title="Acoes rapidas"
        >
          <Box
            sx={{
              display: "flex",
              flexWrap: "wrap",
              gap: 1.5,
            }}
          >
            <Button
              onClick={() => handleOpenBudgetList()}
              startIcon={<OpenInNewRoundedIcon />}
              variant="contained"
            >
              Abrir lista filtrada
            </Button>
            <Button
              onClick={() =>
                handleOpenBudgetList({
                  statusId:
                    statusIdByNormalizedName.get(
                      normalizeLookupKey("Em Negociacao"),
                    ) ?? "",
                })
              }
              startIcon={<OpenInNewRoundedIcon />}
              variant="outlined"
            >
              Ver em negociacao
            </Button>
            <Button
              onClick={() =>
                handleOpenBudgetList({
                  statusId:
                    statusIdByNormalizedName.get(
                      normalizeLookupKey("Pedido"),
                    ) ?? "",
                })
              }
              startIcon={<OpenInNewRoundedIcon />}
              variant="outlined"
            >
              Ver pedidos
            </Button>
            <Button
              onClick={() =>
                handleOpenBudgetList({
                  statusId:
                    statusIdByNormalizedName.get(
                      normalizeLookupKey("Cancelado"),
                    ) ?? "",
                })
              }
              startIcon={<OpenInNewRoundedIcon />}
              variant="outlined"
            >
              Ver cancelados
            </Button>
            <Button
              onClick={() => void handleExportDashboardCsv()}
              startIcon={<DownloadRoundedIcon />}
              variant="outlined"
            >
              Exportar CSV
            </Button>
            <Button
              onClick={() => void handleExportDashboardXlsx()}
              startIcon={<DownloadRoundedIcon />}
              variant="outlined"
            >
              Exportar XLSX
            </Button>
          </Box>
        </SectionCard>
      ) : null}

      <Box
        sx={{
          display: "grid",
          gap: 3,
          gridTemplateColumns: {
            lg: "repeat(3, minmax(0, 1fr))",
            md: "repeat(2, minmax(0, 1fr))",
            xs: "minmax(0, 1fr)",
          },
        }}
      >
        {(dashboardData?.metricCards ?? []).map((metric) => (
          <Box key={metric.label}>
            <SectionCard>
              <Box sx={{ display: "flex", flexDirection: "column", gap: 1.5 }}>
                <Chip
                  color="primary"
                  icon={<metric.icon />}
                  label="Resumo"
                  sx={{ alignSelf: "flex-start" }}
                  variant="outlined"
                />
                <Typography color="text.secondary" variant="body2">
                  {metric.label}
                </Typography>
                <Typography variant="h3">{metric.value}</Typography>
                <Typography color="text.secondary" variant="body2">
                  {metric.helper}
                </Typography>
              </Box>
            </SectionCard>
          </Box>
        ))}
      </Box>

      {!dashboardQuery.isLoading && !dashboardQuery.isError ? (
        <Alert severity="info" variant="outlined">
          A leitura comercial e a leitura tecnica aparecem separadas para evitar
          mistura entre vendedor e orcamentista nos indicadores gerenciais e nos
          arquivos exportados.
        </Alert>
      ) : null}

      <Box
        sx={{
          display: "grid",
          gap: 3,
          gridTemplateColumns: {
            lg: "repeat(4, minmax(0, 1fr))",
            md: "repeat(2, minmax(0, 1fr))",
            xs: "minmax(0, 1fr)",
          },
        }}
      >
        {(dashboardData?.technicalMetricCards ?? []).map((metric) => (
          <Box key={metric.label}>
            <SectionCard>
              <Box sx={{ display: "flex", flexDirection: "column", gap: 1.5 }}>
                <Chip
                  color="secondary"
                  icon={<metric.icon />}
                  label="Tecnico"
                  sx={{ alignSelf: "flex-start" }}
                  variant="outlined"
                />
                <Typography color="text.secondary" variant="body2">
                  {metric.label}
                </Typography>
                <Typography variant="h3">{metric.value}</Typography>
                <Typography color="text.secondary" variant="body2">
                  {metric.helper}
                </Typography>
              </Box>
            </SectionCard>
          </Box>
        ))}
      </Box>

      <Box
        sx={{
          display: "grid",
          gap: 3,
          gridTemplateColumns: {
            lg: "repeat(3, minmax(0, 1fr))",
            xs: "minmax(0, 1fr)",
          },
        }}
      >
        <SectionCard
          description="Ranking tecnico por valor bruto dos orcamentos atribuidos a cada orcamentista."
          title="Top orcamentistas por valor"
        >
          {(dashboardData.technicalOverview.topEstimatorsByValue ?? [])
            .length ? (
            <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
              {dashboardData.technicalOverview.topEstimatorsByValue.map(
                (item, index) => (
                  <Box key={item.label}>
                    <Box
                      sx={{
                        alignItems: "center",
                        display: "flex",
                        justifyContent: "space-between",
                      }}
                    >
                      <Typography sx={{ fontWeight: 700 }} variant="body2">
                        {`${index + 1}. ${item.label}`}
                      </Typography>
                      <Chip
                        label={`${item.budgetCount} orc.`}
                        size="small"
                        variant="outlined"
                      />
                    </Box>
                    <Box
                      sx={{
                        alignItems: "center",
                        display: "flex",
                        flexWrap: "wrap",
                        gap: 1,
                        mt: 0.5,
                      }}
                    >
                      <Chip
                        label={currencyFormatter.format(item.grossValue)}
                        size="small"
                        variant="outlined"
                      />
                      <Chip
                        label={`Ticket ${formatCompactCurrency(item.averageTicket)}`}
                        size="small"
                        variant="outlined"
                      />
                      <Chip
                        label={`Ultima atividade ${formatDateOrFallback(item.lastActivityAt, "Nao informada")}`}
                        size="small"
                        variant="outlined"
                      />
                    </Box>
                  </Box>
                ),
              )}
            </Box>
          ) : (
            <Alert severity="info" variant="outlined">
              Nenhum orcamentista encontrado no recorte atual.
            </Alert>
          )}
        </SectionCard>

        <SectionCard
          description="Orcamentistas com maior volume de orcamentos atribuidos no periodo."
          title="Top orcamentistas por quantidade"
        >
          {(dashboardData.technicalOverview.topEstimatorsByBudgetCount ?? [])
            .length ? (
            <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
              {dashboardData.technicalOverview.topEstimatorsByBudgetCount.map(
                (item, index) => (
                  <Box key={item.label}>
                    <Box
                      sx={{
                        alignItems: "center",
                        display: "flex",
                        justifyContent: "space-between",
                      }}
                    >
                      <Typography sx={{ fontWeight: 700 }} variant="body2">
                        {`${index + 1}. ${item.label}`}
                      </Typography>
                      <Chip
                        label={`${item.budgetCount} orc.`}
                        size="small"
                        variant="outlined"
                      />
                    </Box>
                    <Box
                      sx={{
                        alignItems: "center",
                        display: "flex",
                        flexWrap: "wrap",
                        gap: 1,
                        mt: 0.5,
                      }}
                    >
                      <Chip
                        label={currencyFormatter.format(item.grossValue)}
                        size="small"
                        variant="outlined"
                      />
                      <Chip
                        color="success"
                        label={`Conversao ${formatPercentage(item.conversionRate)}`}
                        size="small"
                        variant="outlined"
                      />
                      <Chip
                        label={`Negociacao ${currencyFormatter.format(item.negotiationGrossValue)}`}
                        size="small"
                        variant="outlined"
                      />
                    </Box>
                  </Box>
                ),
              )}
            </Box>
          ) : (
            <Alert severity="info" variant="outlined">
              Nenhum orcamentista encontrado no recorte atual.
            </Alert>
          )}
        </SectionCard>

        <SectionCard
          description="Ultima atividade registrada para cada orcamentista dentro do recorte selecionado."
          title="Ultima atividade tecnica"
        >
          {(dashboardData.technicalOverview.recentEstimators ?? []).length ? (
            <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
              {dashboardData.technicalOverview.recentEstimators.map((item) => (
                <Box
                  key={item.label}
                  sx={{
                    alignItems: "center",
                    display: "flex",
                    justifyContent: "space-between",
                  }}
                >
                  <Box>
                    <Typography sx={{ fontWeight: 700 }} variant="body2">
                      {item.label}
                    </Typography>
                    <Typography color="text.secondary" variant="caption">
                      {`${item.budgetCount} orcamento(s) · ${currencyFormatter.format(item.grossValue)}`}
                    </Typography>
                  </Box>
                  <Chip
                    icon={<InsightsRoundedIcon />}
                    label={formatDateOrFallback(
                      item.lastActivityAt,
                      "Nao informada",
                    )}
                    size="small"
                    variant="outlined"
                  />
                </Box>
              ))}
            </Box>
          ) : (
            <Alert severity="info" variant="outlined">
              Nenhuma atividade tecnica encontrada para o filtro atual.
            </Alert>
          )}
        </SectionCard>
      </Box>

      <Box
        sx={{
          display: "grid",
          gap: 3,
          gridTemplateColumns: {
            lg: "minmax(0, 1.25fr) minmax(0, 0.75fr)",
            xs: "minmax(0, 1fr)",
          },
        }}
      >
        <SectionCard
          description="Top 10 vendedores ordenados pelo valor bruto total do recorte atual."
          title="Top 10 por valor"
        >
          {(dashboardData?.topSalespeopleByValue ?? []).length ? (
            <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
              {dashboardData.topSalespeopleByValue.map((item, index) => (
                <Box
                  key={item.label}
                  sx={{
                    display: "flex",
                    flexDirection: "column",
                    gap: 0.5,
                  }}
                >
                  <Box
                    sx={{
                      alignItems: "center",
                      display: "flex",
                      justifyContent: "space-between",
                    }}
                  >
                    <Typography sx={{ fontWeight: 700 }} variant="body2">
                      {`${index + 1}. ${item.label}`}
                    </Typography>
                    <Box sx={{ display: "flex", gap: 1 }}>
                      <Chip
                        label={`${item.budgetCount} orc.`}
                        size="small"
                        variant="outlined"
                      />
                      <Button
                        onClick={() =>
                          handleOpenBudgetList({
                            salespersonId:
                              salespersonIdByName.get(item.label) ?? undefined,
                          })
                        }
                        size="small"
                        startIcon={<OpenInNewRoundedIcon />}
                        variant="text"
                      >
                        Abrir
                      </Button>
                    </Box>
                  </Box>
                  <Box
                    sx={{
                      alignItems: "center",
                      display: "flex",
                      flexWrap: "wrap",
                      gap: 1,
                    }}
                  >
                    <Chip
                      label={currencyFormatter.format(item.grossValue)}
                      size="small"
                      variant="outlined"
                    />
                    <Chip
                      label={`Ticket ${formatCompactCurrency(item.averageTicket)}`}
                      size="small"
                      variant="outlined"
                    />
                    <Chip
                      label={`Ultima atividade comercial ${formatDateOrFallback(item.lastActivityAt, "Nao informada")}`}
                      size="small"
                      variant="outlined"
                    />
                  </Box>
                </Box>
              ))}
            </Box>
          ) : (
            <Alert severity="info" variant="outlined">
              Nenhum orcamento encontrado para o filtro atual.
            </Alert>
          )}
        </SectionCard>

        <SectionCard
          description="Top 10 vendedores ordenados pelo volume de orcamentos."
          title="Top 10 por quantidade"
        >
          {(dashboardData?.topSalespeopleByBudgetCount ?? []).length ? (
            <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
              {dashboardData.topSalespeopleByBudgetCount.map((item, index) => (
                <Box key={item.label}>
                  <Box
                    sx={{
                      alignItems: "center",
                      display: "flex",
                      justifyContent: "space-between",
                    }}
                  >
                    <Typography sx={{ fontWeight: 700 }} variant="body2">
                      {`${index + 1}. ${item.label}`}
                    </Typography>
                    <Chip
                      label={`${item.budgetCount} orc.`}
                      size="small"
                      variant="outlined"
                    />
                  </Box>
                  <Box
                    sx={{
                      alignItems: "center",
                      display: "flex",
                      flexWrap: "wrap",
                      gap: 1,
                      mt: 0.5,
                    }}
                  >
                    <Chip
                      label={currencyFormatter.format(item.grossValue)}
                      size="small"
                      variant="outlined"
                    />
                    <Chip
                      color="success"
                      label={`Conversao ${formatPercentage(item.conversionRate)}`}
                      size="small"
                      variant="outlined"
                    />
                    <Chip
                      label={`Ultima atividade comercial ${formatDateOrFallback(item.lastActivityAt, "Nao informada")}`}
                      size="small"
                      variant="outlined"
                    />
                  </Box>
                </Box>
              ))}
            </Box>
          ) : (
            <Alert severity="info" variant="outlined">
              Nenhum orcamento encontrado para o filtro atual.
            </Alert>
          )}
        </SectionCard>
      </Box>

      <Box
        sx={{
          display: "grid",
          gap: 3,
          gridTemplateColumns: {
            lg: "repeat(2, minmax(0, 1fr))",
            xs: "minmax(0, 1fr)",
          },
        }}
      >
        <SectionCard
          description="Construtoras com maior volume financeiro no recorte atual, destacando conversao por quantidade e por valor."
          title="Top construtoras"
        >
          {(dashboardData?.topConstructionCompanies ?? []).length ? (
            <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
              {dashboardData.topConstructionCompanies.map((item, index) => (
                <Box key={item.label}>
                  <Box
                    sx={{
                      alignItems: "center",
                      display: "flex",
                      justifyContent: "space-between",
                    }}
                  >
                    <Typography sx={{ fontWeight: 700 }} variant="body2">
                      {`${index + 1}. ${item.label}`}
                    </Typography>
                    <Chip
                      label={`${item.budgetCount} orc.`}
                      size="small"
                      variant="outlined"
                    />
                  </Box>
                  <Box
                    sx={{
                      alignItems: "center",
                      display: "flex",
                      flexWrap: "wrap",
                      gap: 1,
                      mt: 0.5,
                    }}
                  >
                    <Chip
                      label={currencyFormatter.format(item.grossValue)}
                      size="small"
                      variant="outlined"
                    />
                    <Chip
                      color="success"
                      label={`Conv. ${formatPercentage(item.conversionRate)}`}
                      size="small"
                      variant="outlined"
                    />
                    <Chip
                      color="primary"
                      label={`Valor ${formatPercentage(item.valueConversionRate)}`}
                      size="small"
                      variant="outlined"
                    />
                    <Chip
                      label={`Ultima atividade comercial ${formatDateOrFallback(item.lastActivityAt, "Nao informada")}`}
                      size="small"
                      variant="outlined"
                    />
                  </Box>
                </Box>
              ))}
            </Box>
          ) : (
            <Alert severity="info" variant="outlined">
              Nenhuma construtora encontrada para o filtro atual.
            </Alert>
          )}
        </SectionCard>

        <SectionCard
          description="Obras com maior valor bruto no periodo, ajudando a identificar onde esta a melhor concentracao comercial."
          title="Top obras"
        >
          {(dashboardData?.topProjects ?? []).length ? (
            <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
              {dashboardData.topProjects.map((item, index) => (
                <Box key={item.label}>
                  <Box
                    sx={{
                      alignItems: "center",
                      display: "flex",
                      justifyContent: "space-between",
                    }}
                  >
                    <Typography sx={{ fontWeight: 700 }} variant="body2">
                      {`${index + 1}. ${item.label}`}
                    </Typography>
                    <Chip
                      label={`${item.budgetCount} orc.`}
                      size="small"
                      variant="outlined"
                    />
                  </Box>
                  <Box
                    sx={{
                      alignItems: "center",
                      display: "flex",
                      flexWrap: "wrap",
                      gap: 1,
                      mt: 0.5,
                    }}
                  >
                    <Chip
                      label={currencyFormatter.format(item.grossValue)}
                      size="small"
                      variant="outlined"
                    />
                    <Chip
                      color="success"
                      label={`Conv. ${formatPercentage(item.conversionRate)}`}
                      size="small"
                      variant="outlined"
                    />
                    <Chip
                      color="primary"
                      label={`Valor ${formatPercentage(item.valueConversionRate)}`}
                      size="small"
                      variant="outlined"
                    />
                    <Chip
                      label={`Ultima atividade comercial ${formatDateOrFallback(item.lastActivityAt, "Nao informada")}`}
                      size="small"
                      variant="outlined"
                    />
                  </Box>
                </Box>
              ))}
            </Box>
          ) : (
            <Alert severity="info" variant="outlined">
              Nenhuma obra encontrada para o filtro atual.
            </Alert>
          )}
        </SectionCard>

        <SectionCard
          description="Principais motivos de perda por impacto financeiro para orientar as acoes corretivas do time comercial."
          title="Motivos de perda"
        >
          {(dashboardData?.topLossReasons ?? []).length ? (
            <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
              {dashboardData.topLossReasons.map((item, index) => (
                <Box key={item.label}>
                  <Box
                    sx={{
                      alignItems: "center",
                      display: "flex",
                      justifyContent: "space-between",
                    }}
                  >
                    <Typography sx={{ fontWeight: 700 }} variant="body2">
                      {`${index + 1}. ${item.label}`}
                    </Typography>
                    <Chip
                      label={`${item.lostBudgetCount} perda(s)`}
                      size="small"
                      variant="outlined"
                    />
                  </Box>
                  <Box
                    sx={{
                      alignItems: "center",
                      display: "flex",
                      flexWrap: "wrap",
                      gap: 1,
                      mt: 0.5,
                    }}
                  >
                    <Chip
                      label={`Valor ${currencyFormatter.format(item.grossValue)}`}
                      size="small"
                      variant="outlined"
                    />
                    <Chip
                      label={`Ticket ${formatCompactCurrency(item.averageTicket)}`}
                      size="small"
                      variant="outlined"
                    />
                  </Box>
                </Box>
              ))}
            </Box>
          ) : (
            <Alert severity="info" variant="outlined">
              Nenhum motivo de perda encontrado para o filtro atual.
            </Alert>
          )}
        </SectionCard>

        <SectionCard
          description="Tempo medio entre o envio e o fechamento dos orcamentos finalizados no recorte atual."
          title="Tempo medio de fechamento"
        >
          {(dashboardData?.averageClosingTimes ?? []).length ? (
            <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
              {dashboardData.averageClosingTimes.map((item) => (
                <Box
                  key={item.label}
                  sx={{
                    border: (theme) => `1px solid ${theme.palette.divider}`,
                    borderRadius: 2,
                    p: 2,
                  }}
                >
                  <Box
                    sx={{
                      alignItems: "center",
                      display: "flex",
                      justifyContent: "space-between",
                    }}
                  >
                    <Typography sx={{ fontWeight: 700 }} variant="body2">
                      {item.label}
                    </Typography>
                    <Chip
                      label={`${item.budgetCount} fechado(s)`}
                      size="small"
                      variant="outlined"
                    />
                  </Box>
                  <Typography sx={{ mt: 1 }} variant="h5">
                    {formatClosingDays(item.averageClosingDays)}
                  </Typography>
                  <Typography color="text.secondary" variant="body2">
                    {`Valor bruto relacionado: ${currencyFormatter.format(item.grossValue)}`}
                  </Typography>
                </Box>
              ))}
            </Box>
          ) : (
            <Alert severity="info" variant="outlined">
              Nenhum fechamento encontrado para calcular o tempo medio no filtro
              atual.
            </Alert>
          )}
        </SectionCard>
      </Box>

      <Box
        sx={{
          display: "grid",
          gap: 3,
          gridTemplateColumns: {
            lg: "repeat(3, minmax(0, 1fr))",
            xs: "minmax(0, 1fr)",
          },
        }}
      >
        <SectionCard
          description="Comparativo entre os dois meses mais recentes do recorte atual para identificar aceleracao ou perda de ritmo."
          title="Tendencia mensal"
        >
          {dashboardInsights.monthComparison !== null ? (
            <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
              <Alert severity="info" variant="outlined">
                {`Comparando ${dashboardInsights.monthComparison.currentMonthLabel} com ${dashboardInsights.monthComparison.previousMonthLabel}.`}
              </Alert>
              <Box
                sx={{
                  display: "grid",
                  gap: 1.5,
                  gridTemplateColumns: {
                    md: "repeat(3, minmax(0, 1fr))",
                    xs: "minmax(0, 1fr)",
                  },
                }}
              >
                <Box
                  sx={{
                    border: (theme) => `1px solid ${theme.palette.divider}`,
                    borderRadius: 2,
                    p: 2,
                  }}
                >
                  <Typography color="text.secondary" variant="body2">
                    Orcamentos
                  </Typography>
                  <Typography sx={{ mt: 0.5 }} variant="h5">
                    {dashboardInsights.monthComparison.currentMonth.budgetCount}
                  </Typography>
                  <Chip
                    color={getDeltaColor(
                      dashboardInsights.monthComparison.budgetDelta,
                    )}
                    label={`${formatSignedInteger(
                      dashboardInsights.monthComparison.budgetDelta,
                    )} vs anterior`}
                    size="small"
                    sx={{ mt: 1 }}
                    variant="outlined"
                  />
                </Box>
                <Box
                  sx={{
                    border: (theme) => `1px solid ${theme.palette.divider}`,
                    borderRadius: 2,
                    p: 2,
                  }}
                >
                  <Typography color="text.secondary" variant="body2">
                    Valor bruto
                  </Typography>
                  <Typography sx={{ mt: 0.5 }} variant="h5">
                    {formatCompactCurrency(
                      dashboardInsights.monthComparison.currentMonth.grossValue,
                    )}
                  </Typography>
                  <Chip
                    color={getDeltaColor(
                      dashboardInsights.monthComparison.grossValueDelta,
                    )}
                    label={`${formatSignedCurrency(
                      dashboardInsights.monthComparison.grossValueDelta,
                    )} vs anterior`}
                    size="small"
                    sx={{ mt: 1 }}
                    variant="outlined"
                  />
                </Box>
                <Box
                  sx={{
                    border: (theme) => `1px solid ${theme.palette.divider}`,
                    borderRadius: 2,
                    p: 2,
                  }}
                >
                  <Typography color="text.secondary" variant="body2">
                    Conversao
                  </Typography>
                  <Typography sx={{ mt: 0.5 }} variant="h5">
                    {formatPercentage(
                      dashboardInsights.monthComparison.currentMonth
                        .budgetCount === 0
                        ? 0
                        : (dashboardInsights.monthComparison.currentMonth
                            .wonBudgetCount /
                            dashboardInsights.monthComparison.currentMonth
                              .budgetCount) *
                            100,
                    )}
                  </Typography>
                  <Chip
                    color={getDeltaColor(
                      dashboardInsights.monthComparison.conversionDelta,
                    )}
                    label={`${formatSignedPercentagePoints(
                      dashboardInsights.monthComparison.conversionDelta,
                    )} vs anterior`}
                    size="small"
                    sx={{ mt: 1 }}
                    variant="outlined"
                  />
                </Box>
              </Box>
            </Box>
          ) : (
            <Alert severity="info" variant="outlined">
              Sao necessarios pelo menos dois meses no recorte atual para montar
              o comparativo de tendencia.
            </Alert>
          )}
        </SectionCard>

        <SectionCard
          description="Vendedores com melhor conversao no periodo, priorizando quem tem pelo menos dois orcamentos no recorte."
          title="Top conversao"
        >
          {(dashboardData?.topSalespeopleByConversion ?? []).length ? (
            <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
              {dashboardData.topSalespeopleByConversion.map((item, index) => (
                <Box key={item.label}>
                  <Box
                    sx={{
                      alignItems: "center",
                      display: "flex",
                      justifyContent: "space-between",
                    }}
                  >
                    <Typography sx={{ fontWeight: 700 }} variant="body2">
                      {`${index + 1}. ${item.label}`}
                    </Typography>
                    <Chip
                      color="success"
                      label={formatPercentage(item.conversionRate)}
                      size="small"
                      variant="outlined"
                    />
                  </Box>
                  <Box
                    sx={{
                      alignItems: "center",
                      display: "flex",
                      flexWrap: "wrap",
                      gap: 1,
                      mt: 0.5,
                    }}
                  >
                    <Chip
                      label={`${item.wonBudgetCount} pedido(s)`}
                      size="small"
                      variant="outlined"
                    />
                    <Chip
                      label={`${item.budgetCount} orc.`}
                      size="small"
                      variant="outlined"
                    />
                    <Chip
                      label={`Ticket ${formatCompactCurrency(item.averageTicket)}`}
                      size="small"
                      variant="outlined"
                    />
                  </Box>
                </Box>
              ))}
            </Box>
          ) : (
            <Alert severity="info" variant="outlined">
              Nenhum vendedor elegivel para o ranking de conversao no filtro
              atual.
            </Alert>
          )}
        </SectionCard>

        <SectionCard
          description="Vendedores com maior ticket medio no periodo, considerando base minima para reduzir distorcao."
          title="Top ticket medio"
        >
          {(dashboardData?.topSalespeopleByAverageTicket ?? []).length ? (
            <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
              {dashboardData.topSalespeopleByAverageTicket.map(
                (item, index) => (
                  <Box key={item.label}>
                    <Box
                      sx={{
                        alignItems: "center",
                        display: "flex",
                        justifyContent: "space-between",
                      }}
                    >
                      <Typography sx={{ fontWeight: 700 }} variant="body2">
                        {`${index + 1}. ${item.label}`}
                      </Typography>
                      <Chip
                        color="primary"
                        label={formatCompactCurrency(item.averageTicket)}
                        size="small"
                        variant="outlined"
                      />
                    </Box>
                    <Box
                      sx={{
                        alignItems: "center",
                        display: "flex",
                        flexWrap: "wrap",
                        gap: 1,
                        mt: 0.5,
                      }}
                    >
                      <Chip
                        label={`${item.budgetCount} orc.`}
                        size="small"
                        variant="outlined"
                      />
                      <Chip
                        label={currencyFormatter.format(item.grossValue)}
                        size="small"
                        variant="outlined"
                      />
                      <Chip
                        label={`Conversao ${formatPercentage(item.conversionRate)}`}
                        size="small"
                        variant="outlined"
                      />
                    </Box>
                  </Box>
                ),
              )}
            </Box>
          ) : (
            <Alert severity="info" variant="outlined">
              Nenhum vendedor elegivel para o ranking de ticket medio no filtro
              atual.
            </Alert>
          )}
        </SectionCard>
      </Box>

      <Box
        sx={{
          display: "grid",
          gap: 3,
          gridTemplateColumns: {
            lg: "repeat(3, minmax(0, 1fr))",
            xs: "minmax(0, 1fr)",
          },
        }}
      >
        <SectionCard
          description="Carteira ainda em negociacao, com foco no valor e no volume por vendedor."
          title="Carteira em negociacao"
        >
          {(dashboardData?.negotiationPipeline ?? []).length ? (
            <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
              {dashboardData.negotiationPipeline.map((item) => (
                <Box key={item.label}>
                  <Box
                    sx={{
                      alignItems: "center",
                      display: "flex",
                      justifyContent: "space-between",
                    }}
                  >
                    <Typography sx={{ fontWeight: 700 }} variant="body2">
                      {item.label}
                    </Typography>
                    <Box sx={{ display: "flex", gap: 1 }}>
                      <Chip
                        color="warning"
                        label={`${item.negotiationBudgetCount} em aberto`}
                        size="small"
                        variant="outlined"
                      />
                      <Button
                        onClick={() =>
                          handleOpenBudgetList({
                            salespersonId:
                              salespersonIdByName.get(item.label) ?? undefined,
                            statusId:
                              statusIdByNormalizedName.get(
                                normalizeLookupKey("Em Negociacao"),
                              ) ?? undefined,
                          })
                        }
                        size="small"
                        startIcon={<OpenInNewRoundedIcon />}
                        variant="text"
                      >
                        Abrir
                      </Button>
                    </Box>
                  </Box>
                  <Box
                    sx={{
                      alignItems: "center",
                      display: "flex",
                      flexWrap: "wrap",
                      gap: 1,
                      mt: 0.5,
                    }}
                  >
                    <Chip
                      label={currencyFormatter.format(
                        item.negotiationGrossValue,
                      )}
                      size="small"
                      variant="outlined"
                    />
                    <Chip
                      label={`${item.stalledBudgetCount} parado(s)`}
                      size="small"
                      variant="outlined"
                    />
                    <Chip
                      label={`Ultima atividade comercial ${formatDateOrFallback(item.lastActivityAt, "Nao informada")}`}
                      size="small"
                      variant="outlined"
                    />
                  </Box>
                </Box>
              ))}
            </Box>
          ) : (
            <Alert severity="info" variant="outlined">
              Nenhuma carteira em negociacao encontrada para o filtro atual.
            </Alert>
          )}
        </SectionCard>

        <SectionCard
          description="Comparativo do funil principal por vendedor com foco em negociacao, pedidos e perdas."
          title="Funil por vendedor"
        >
          {(dashboardData?.salespersonFunnel ?? []).length ? (
            <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
              {dashboardData.salespersonFunnel.map((item) => {
                const wonWidth =
                  item.totalBudgets === 0
                    ? 0
                    : (item.wonBudgets / item.totalBudgets) * 100;
                const negotiationWidth =
                  item.totalBudgets === 0
                    ? 0
                    : (item.negotiationBudgets / item.totalBudgets) * 100;
                const lostWidth =
                  item.totalBudgets === 0
                    ? 0
                    : (item.lostBudgets / item.totalBudgets) * 100;

                return (
                  <Box key={item.label}>
                    <Box
                      sx={{
                        alignItems: "center",
                        display: "flex",
                        justifyContent: "space-between",
                      }}
                    >
                      <Typography sx={{ fontWeight: 700 }} variant="body2">
                        {item.label}
                      </Typography>
                      <Chip
                        label={`${item.totalBudgets} orc.`}
                        size="small"
                        variant="outlined"
                      />
                    </Box>
                    <Box
                      sx={{
                        bgcolor: "action.hover",
                        borderRadius: 999,
                        display: "flex",
                        height: 10,
                        mt: 1,
                        overflow: "hidden",
                      }}
                    >
                      <Box
                        sx={{
                          bgcolor: "success.main",
                          width: `${wonWidth}%`,
                        }}
                      />
                      <Box
                        sx={{
                          bgcolor: "warning.main",
                          width: `${negotiationWidth}%`,
                        }}
                      />
                      <Box
                        sx={{
                          bgcolor: "text.disabled",
                          width: `${lostWidth}%`,
                        }}
                      />
                    </Box>
                    <Box
                      sx={{
                        alignItems: "center",
                        display: "flex",
                        flexWrap: "wrap",
                        gap: 1,
                        mt: 1,
                      }}
                    >
                      <Chip
                        color="success"
                        label={`Pedido ${item.wonBudgets}`}
                        size="small"
                        variant="outlined"
                      />
                      <Chip
                        color="warning"
                        label={`Negociacao ${item.negotiationBudgets}`}
                        size="small"
                        variant="outlined"
                      />
                      <Chip
                        label={`Cancelado ${item.lostBudgets}`}
                        size="small"
                        variant="outlined"
                      />
                      <Chip
                        label={`Conversao ${formatPercentage(item.conversionRate)}`}
                        size="small"
                        variant="outlined"
                      />
                    </Box>
                  </Box>
                );
              })}
            </Box>
          ) : (
            <Alert severity="info" variant="outlined">
              Nenhum dado de funil encontrado para o filtro atual.
            </Alert>
          )}
        </SectionCard>

        <SectionCard
          description="Leitura rapida da ultima atividade comercial registrada por vendedor dentro do recorte."
          title="Ultima atividade comercial"
        >
          {(dashboardData?.recentSalespeople ?? []).length ? (
            <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
              {dashboardData.recentSalespeople.map((item) => (
                <Box
                  key={item.label}
                  sx={{
                    alignItems: "center",
                    display: "flex",
                    justifyContent: "space-between",
                  }}
                >
                  <Box>
                    <Typography sx={{ fontWeight: 700 }} variant="body2">
                      {item.label}
                    </Typography>
                    <Typography color="text.secondary" variant="caption">
                      {`${item.budgetCount} orcamento(s) · ${currencyFormatter.format(item.grossValue)}`}
                    </Typography>
                  </Box>
                  <Chip
                    icon={<InsightsRoundedIcon />}
                    label={formatDateOrFallback(
                      item.lastActivityAt,
                      "Nao informada",
                    )}
                    size="small"
                    variant="outlined"
                  />
                </Box>
              ))}
            </Box>
          ) : (
            <Alert severity="info" variant="outlined">
              Nenhuma atividade comercial encontrada para o filtro atual.
            </Alert>
          )}
        </SectionCard>

        <SectionCard
          description="Oportunidades em negociacao sem atividade comercial recente, priorizadas pelos casos mais antigos."
          title="Orcamentos parados"
        >
          {(dashboardData?.staleBudgets ?? []).length ? (
            <Box sx={{ display: "flex", flexDirection: "column", gap: 1.5 }}>
              {dashboardData.staleBudgets.map((item, index) => (
                <Box key={item.id}>
                  {index > 0 ? <Divider sx={{ mb: 1.5 }} /> : null}
                  <Box
                    sx={{
                      display: "flex",
                      flexDirection: "column",
                      gap: 0.75,
                    }}
                  >
                    <Box
                      sx={{
                        alignItems: "center",
                        display: "flex",
                        justifyContent: "space-between",
                      }}
                    >
                      <Typography sx={{ fontWeight: 700 }} variant="body2">
                        {item.budgetNumber}
                      </Typography>
                      <Box sx={{ display: "flex", gap: 1 }}>
                        <Chip
                          color="warning"
                          label={`${item.stalledDays} dia(s)`}
                          size="small"
                          variant="outlined"
                        />
                        <Button
                          onClick={() =>
                            handleOpenBudgetList({
                              budgetNumber: item.budgetNumber,
                            })
                          }
                          size="small"
                          startIcon={<OpenInNewRoundedIcon />}
                          variant="text"
                        >
                          Abrir
                        </Button>
                      </Box>
                    </Box>
                    <Typography color="text.secondary" variant="caption">
                      {`${item.salespersonLabel} · ${item.projectLabel}`}
                    </Typography>
                    <Typography color="text.secondary" variant="caption">
                      {`${item.constructionCompanyLabel} · ${item.statusLabel}`}
                    </Typography>
                    <Box
                      sx={{
                        alignItems: "center",
                        display: "flex",
                        flexWrap: "wrap",
                        gap: 1,
                      }}
                    >
                      <Chip
                        label={currencyFormatter.format(item.grossValue)}
                        size="small"
                        variant="outlined"
                      />
                      <Chip
                        label={`Ultima atividade comercial ${formatDateOrFallback(item.lastActivityAt, "Nao informada")}`}
                        size="small"
                        variant="outlined"
                      />
                    </Box>
                  </Box>
                </Box>
              ))}
            </Box>
          ) : (
            <Alert severity="success" variant="outlined">
              Nenhum orcamento parado encontrado para o filtro atual.
            </Alert>
          )}
        </SectionCard>
      </Box>

      <Box
        sx={{
          display: "grid",
          gap: 3,
          gridTemplateColumns: {
            lg: "minmax(0, 1.4fr) minmax(0, 0.6fr)",
            xs: "minmax(0, 1fr)",
          },
        }}
      >
        <SectionCard
          description="Evolucao do volume orcado e do valor convertido ao longo do tempo dentro do recorte atual."
          title="Evolucao mensal"
        >
          {(dashboardData?.monthlyEvolution ?? []).length ? (
            <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
              {dashboardData.monthlyEvolution.map((item) => {
                const progressValue =
                  (dashboardData.maxMonthlyGrossValue ?? 0) === 0
                    ? 0
                    : (item.grossValue / dashboardData.maxMonthlyGrossValue) *
                      100;

                return (
                  <Box key={item.monthKey}>
                    <Box
                      sx={{
                        alignItems: "center",
                        display: "flex",
                        justifyContent: "space-between",
                      }}
                    >
                      <Typography sx={{ fontWeight: 700 }} variant="body2">
                        {item.monthLabel}
                      </Typography>
                      <Chip
                        label={`${item.budgetCount} orc.`}
                        size="small"
                        variant="outlined"
                      />
                    </Box>
                    <LinearProgress
                      sx={{ borderRadius: 999, height: 8, mt: 1 }}
                      value={progressValue}
                      variant="determinate"
                    />
                    <Box
                      sx={{
                        alignItems: "center",
                        display: "flex",
                        flexWrap: "wrap",
                        gap: 1,
                        mt: 1,
                      }}
                    >
                      <Chip
                        label={currencyFormatter.format(item.grossValue)}
                        size="small"
                        variant="outlined"
                      />
                      <Chip
                        color="success"
                        label={`Pedido ${currencyFormatter.format(item.wonGrossValue)}`}
                        size="small"
                        variant="outlined"
                      />
                      <Chip
                        label={`${item.wonBudgetCount} pedido(s)`}
                        size="small"
                        variant="outlined"
                      />
                    </Box>
                  </Box>
                );
              })}
            </Box>
          ) : (
            <Alert severity="info" variant="outlined">
              Nenhuma evolucao mensal encontrada para o filtro atual.
            </Alert>
          )}
        </SectionCard>

        <SectionCard
          description="Leitura consolidada do funil principal dentro do periodo selecionado."
          title="Resumo da carteira"
        >
          <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
            {[
              {
                color: "success.main",
                label: "Pedidos",
                value: dashboardData?.wonBudgets ?? 0,
              },
              {
                color: "warning.main",
                label: "Em negociacao",
                value: dashboardData?.negotiationBudgets ?? 0,
              },
              {
                color: "text.secondary",
                label: "Cancelados",
                value: dashboardData?.lostBudgets ?? 0,
              },
              {
                color: "primary.main",
                label: "Conversao",
                value: formatPercentage(dashboardData?.conversionRate ?? 0),
              },
            ].map((item) => (
              <Box
                key={item.label}
                sx={{
                  alignItems: "center",
                  display: "flex",
                  justifyContent: "space-between",
                }}
              >
                <Typography color="text.secondary" variant="body2">
                  {item.label}
                </Typography>
                <Chip
                  label={String(item.value)}
                  size="small"
                  sx={{
                    bgcolor: item.color,
                    color: "common.white",
                    fontWeight: 700,
                  }}
                />
              </Box>
            ))}
          </Box>
        </SectionCard>
      </Box>
    </Box>
  );
}
