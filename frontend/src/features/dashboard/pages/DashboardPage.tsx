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
  Slider,
  TextField,
  Typography,
} from "@mui/material";
import { AxiosError, isAxiosError } from "axios";
import { useQuery } from "@tanstack/react-query";
import { useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  compactFilterFieldSx,
  FilterField,
} from "../../../components/common/FilterField";
import { PageHeader } from "../../../components/common/PageHeader";
import { SectionCard } from "../../../components/common/SectionCard";
import {
  formatCurrencyInputValue,
  parseCurrencyInputToNumericString,
} from "../../../shared/utils/currencyInput";
import { useAuth } from "../../auth/hooks/useAuth";
import { getBudgetCatalogsRequest } from "../../budgets/api/budgets";
import type { BudgetCatalogItem } from "../../budgets/types/budget";
import {
  getWonStatusPluralLabel,
  getWonStatusSingularLabel,
  isWonStatusLabel,
} from "../../budgets/utils/businessTerms";
import {
  getDashboardGrossValueRangeRequest,
  getSalespeopleDashboardRequest,
} from "../api/dashboard";
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
  grossValueMax?: string;
  grossValueMin?: string;
  installerId?: string;
  projectId?: string;
  projectName?: string;
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
  { value: "3", label: "Março" },
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

const dashboardFeedbackAlertSx = {
  borderRadius: 3,
  boxShadow: "0 14px 30px rgba(30, 58, 138, 0.08)",
  "& .MuiAlert-message": {
    fontWeight: 600,
    lineHeight: 1.65,
  },
} as const;

const dashboardInfoAlertSx = {
  borderRadius: 3,
  boxShadow: "0 14px 28px rgba(30, 58, 138, 0.07)",
  "& .MuiAlert-message": {
    lineHeight: 1.65,
  },
} as const;

const dashboardLoaderSx = {
  borderRadius: 999,
  height: 8,
  overflow: "hidden",
  "& .MuiLinearProgress-bar": {
    borderRadius: 999,
  },
} as const;

const dashboardValueRangeGridSx = {
  display: "grid",
  gap: 2,
  gridTemplateColumns: {
    md: "repeat(2, minmax(0, 1fr))",
    xs: "minmax(0, 1fr)",
  },
} as const;

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

function parseNonNegativeDecimalString(value: string | null) {
  if (value === null) {
    return "";
  }

  const parsedValue = Number(value);
  if (!Number.isFinite(parsedValue) || parsedValue < 0) {
    return "";
  }

  return String(parsedValue);
}

function parseDraftCurrencyValue(value: string) {
  if (!value.trim()) {
    return null;
  }

  const parsedValue = Number(value);
  if (!Number.isFinite(parsedValue) || parsedValue < 0) {
    return null;
  }

  return parsedValue;
}

function formatCurrencyRangeValue(value: string, fallback: number) {
  const parsedValue = parseDraftCurrencyValue(value);

  return currencyFormatter.format(parsedValue ?? fallback);
}

function normalizeGrossValueDraftRange(minValue: string, maxValue: string) {
  const parsedMin = parseDraftCurrencyValue(minValue);
  const parsedMax = parseDraftCurrencyValue(maxValue);

  if (parsedMin !== null && parsedMax !== null && parsedMin > parsedMax) {
    return {
      grossValueMin: String(parsedMax),
      grossValueMax: String(parsedMin),
    };
  }

  return {
    grossValueMin: parsedMin === null ? "" : String(parsedMin),
    grossValueMax: parsedMax === null ? "" : String(parsedMax),
  };
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
      return "Sua sessão expirou. Entre novamente para acessar o dashboard.";
    }
    if (statusCode === 403) {
      return (
        apiMessage ||
        "Você não possui permissão para acessar este dashboard administrativo."
      );
    }
    if (statusCode === 400) {
      return (
        apiMessage || "Os filtros informados para o dashboard são inválidos."
      );
    }

    return apiMessage || "Não foi possível carregar os dados do dashboard.";
  }

  return "Não foi possível carregar os dados do dashboard.";
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
  grossValueMax,
  grossValueMin,
  installerId,
  month,
  projectId,
  projectName,
  salespersonId,
  sourceCompany,
  statusId,
  year,
}: BudgetListNavigationOptions) {
  const searchParams = new URLSearchParams();

  if (budgetNumber) {
    searchParams.set("budgetNumber", budgetNumber);
  }
  if (installerId) {
    searchParams.set("installerId", installerId);
  }
  if (projectId) {
    searchParams.set("projectId", projectId);
  }
  if (projectName) {
    searchParams.set("projectName", projectName);
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
  if (grossValueMin) {
    searchParams.set("grossValueMin", grossValueMin);
  }
  if (grossValueMax) {
    searchParams.set("grossValueMax", grossValueMax);
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
  grossValueLabel: string,
) {
  const companyLabel = sourceCompany || "todas as empresas";
  const yearLabel = selectedYear || "todos os anos";
  const monthLabel =
    monthOptions.find((item) => item.value === selectedMonth)?.label ??
    "todos os meses";
  const salespersonScope = salespersonLabel || "todos os vendedores";
  const grossValueScope = grossValueLabel || "todos os valores";

  return `${companyLabel}, ${yearLabel}, ${monthLabel}, ${salespersonScope} e ${grossValueScope}`;
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
  selectedGrossValueLabel,
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
  selectedGrossValueLabel: string;
  selectedSalespersonLabel: string;
  selectedYear: string;
  sourceCompany: DashboardCompanyFilter;
}): SpreadsheetSheet[] {
  const scopeLabel = buildDashboardScopeLabel(
    sourceCompany,
    selectedYear,
    selectedMonth,
    selectedSalespersonLabel,
    selectedGrossValueLabel,
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
        ["Indicador", "Valor", "Descrição"],
        ...dashboardData.metricCards.map((metric) => [
          metric.label,
          metric.value,
          metric.helper,
        ]),
        [],
        ["Destaques"],
        [
          "Top vendedor por valor",
          topSalespersonByValue?.label ?? "Não informado",
        ],
        ["Maior valor bruto", topSalespersonByValue?.grossValue ?? null],
        [
          "Melhor conversão por valor",
          `${formatPercentage(dashboardData.valueConversionRate)}`,
        ],
        [
          "Última atividade comercial",
          topSalespersonByValue
            ? toSpreadsheetDateOrFallback(
                topSalespersonByValue.lastActivityAt,
                "Não informada",
              )
            : "Não informada",
        ],
        [
          "Vendedor com atividade mais recente",
          mostRecentSalesperson?.label ?? "Não informado",
        ],
        [
          "Orçamento mais parado",
          mostStalledBudget?.budgetNumber ?? "Não informado",
        ],
        [
          "Dias parados do caso mais antigo",
          mostStalledBudget?.stalledDays ?? null,
        ],
        [],
        ["Novos estratégicos"],
        ["Construtora líder", topConstructionCompany?.label ?? "Não informado"],
        ["Obra líder", topProject?.label ?? "Não informado"],
        ["Principal motivo de perda", topLossReason?.label ?? "Não informado"],
        [
          "Tempo médio de fechamento",
          averageClosingTime
            ? formatClosingDays(averageClosingTime.averageClosingDays)
            : "Não informado",
        ],
        [],
        ["Resumo da carteira", "Valor"],
        [getWonStatusPluralLabel(), dashboardData.wonBudgets],
        ["Em negociação", dashboardData.negotiationBudgets],
        ["Cancelados", dashboardData.lostBudgets],
        ["Conversão", formatPercentage(dashboardData.conversionRate)],
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
      name: "Análise Comercial",
      rows: [
        ["Análise comercial automática"],
        ["Escopo", scopeLabel],
        ["Gerado em", new Date()],
        [],
        ["Rankings e destaques"],
        [
          "Maior valor bruto",
          topSalespersonByValue?.label ?? "Não informado",
          topSalespersonByValue?.grossValue ?? null,
        ],
        [
          "Maior quantidade de orçamentos",
          topSalespersonByBudgetCount?.label ?? "Não informado",
          topSalespersonByBudgetCount?.budgetCount ?? null,
        ],
        [
          "Melhor conversão",
          topSalespersonByConversion?.label ?? "Não informado",
          topSalespersonByConversion !== undefined
            ? topSalespersonByConversion.conversionRate / 100
            : null,
        ],
        [
          "Maior ticket médio",
          topSalespersonByAverageTicket?.label ?? "Não informado",
          topSalespersonByAverageTicket?.averageTicket ?? null,
        ],
        [
          "Construtora com maior valor",
          topConstructionCompany?.label ?? "Não informado",
          topConstructionCompany?.grossValue ?? null,
        ],
        [
          "Obra com maior valor",
          topProject?.label ?? "Não informado",
          topProject?.grossValue ?? null,
        ],
        [
          "Principal motivo de perda",
          topLossReason?.label ?? "Não informado",
          topLossReason?.grossValue ?? null,
        ],
        [
          "Tempo médio de fechamento",
          averageClosingTime?.label ?? "Não informado",
          averageClosingTime?.averageClosingDays ?? null,
        ],
        [
          "Atividade comercial mais recente",
          mostRecentSalesperson?.label ?? "Não informado",
          mostRecentSalesperson
            ? toSpreadsheetDateOrFallback(
                mostRecentSalesperson.lastActivityAt,
                "Não informada",
              )
            : "Não informada",
        ],
        [
          "Maior carteira parada",
          mostStalledSalesperson?.label ?? "Não informado",
          mostStalledSalesperson?.stalledBudgetCount ?? null,
        ],
        [
          "Orçamento mais antigo sem atividade",
          mostStalledBudget?.budgetNumber ?? "Não informado",
          mostStalledBudget?.stalledDays ?? null,
        ],
        [],
        ["Comparação mensal"],
        currentMonth && previousMonth
          ? [
              "Período comparado",
              `${currentMonth.monthLabel} vs ${previousMonth.monthLabel}`,
            ]
          : ["Período comparado", "Base insuficiente"],
        [
          "Orçamentos no mês atual",
          currentMonth?.budgetCount ?? null,
          currentMonth?.monthLabel ?? "Não informado",
        ],
        [
          "Valor bruto no mês atual",
          currentMonth?.grossValue ?? null,
          currentMonth?.monthLabel ?? "Não informado",
        ],
        [
          "Conversão no mes atual",
          currentMonth ? currentConversionRate / 100 : null,
          currentMonth?.monthLabel ?? "Não informado",
        ],
        [
          "Variação de orçamentos",
          budgetDelta,
          budgetDelta === null
            ? "Base insuficiente"
            : formatSignedInteger(budgetDelta),
        ],
        [
          "Variação de valor bruto",
          grossValueDelta,
          grossValueDelta === null
            ? "Base insuficiente"
            : formatSignedCurrency(grossValueDelta),
        ],
        [
          "Variação de conversão",
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
          "Posição",
          "Vendedor",
          "Orçamentos",
          "Valor bruto",
          "Ticket médio",
          "Última atividade comercial",
        ],
        ...dashboardData.topSalespeopleByValue.map((item, index) => [
          index + 1,
          item.label,
          item.budgetCount,
          item.grossValue,
          item.averageTicket,
          toSpreadsheetDateOrFallback(item.lastActivityAt, "Não informada"),
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
          "Posição",
          "Vendedor",
          "Orçamentos",
          "Valor bruto",
          "Conversão",
          "Última atividade comercial",
        ],
        ...dashboardData.topSalespeopleByBudgetCount.map((item, index) => [
          index + 1,
          item.label,
          item.budgetCount,
          item.grossValue,
          item.conversionRate / 100,
          toSpreadsheetDateOrFallback(item.lastActivityAt, "Não informada"),
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
      name: "Top Conversão",
      rows: [
        [
          "Posição",
          "Vendedor",
          "Conversão",
          getWonStatusPluralLabel(),
          "Orçamentos",
          "Ticket médio",
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
          "Posição",
          "Vendedor",
          "Ticket médio",
          "Orçamentos",
          "Valor bruto",
          "Conversão",
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
      name: "Carteira em Negociação",
      rows: [
        [
          "Vendedor",
          "Em aberto",
          "Valor em negociação",
          "Orçamentos parados",
          "Última atividade comercial",
        ],
        ...dashboardData.negotiationPipeline.map((item) => [
          item.label,
          item.negotiationBudgetCount,
          item.negotiationGrossValue,
          item.stalledBudgetCount,
          toSpreadsheetDateOrFallback(item.lastActivityAt, "Não informada"),
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
      name: "Última Atividade",
      rows: [
        ["Vendedor", "Orçamentos", "Valor bruto", "Última atividade comercial"],
        ...dashboardData.recentSalespeople.map((item) => [
          item.label,
          item.budgetCount,
          item.grossValue,
          toSpreadsheetDateOrFallback(item.lastActivityAt, "Não informada"),
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
      name: "Orçamentos Parados",
      rows: [
        [
          "Orçamento",
          "Vendedor",
          "Obra",
          "Construtora",
          "Status",
          "Valor bruto",
          "Dias parados",
          "Última atividade comercial",
        ],
        ...dashboardData.staleBudgets.map((item) => [
          item.budgetNumber,
          item.salespersonLabel,
          item.projectLabel,
          item.constructionCompanyLabel,
          item.statusLabel,
          item.grossValue,
          item.stalledDays,
          toSpreadsheetDateOrFallback(item.lastActivityAt, "Não informada"),
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
      name: "Evolução Mensal",
      rows: [
        [
          "Mês",
          "Orçamentos",
          "Valor bruto",
          getWonStatusPluralLabel(),
          "Valor convertido",
        ],
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
          "Orçamentos",
          "Valor bruto",
          "Conversão",
          "Conversão por valor",
          "Última atividade comercial",
        ],
        ...dashboardData.topConstructionCompanies.map((item) => [
          item.label,
          item.budgetCount,
          item.grossValue,
          item.conversionRate / 100,
          item.valueConversionRate / 100,
          toSpreadsheetDateOrFallback(item.lastActivityAt, "Não informada"),
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
          "Orçamentos",
          "Valor bruto",
          "Conversão",
          "Conversão por valor",
          "Última atividade comercial",
        ],
        ...dashboardData.topProjects.map((item) => [
          item.label,
          item.budgetCount,
          item.grossValue,
          item.conversionRate / 100,
          item.valueConversionRate / 100,
          toSpreadsheetDateOrFallback(item.lastActivityAt, "Não informada"),
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
        ["Motivo de perda", "Cancelados", "Valor perdido", "Ticket médio"],
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
        ["Recorte", "Tempo médio", "Valor bruto", "Orçamentos fechados"],
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
          "Total de orçamentos",
          "Em negociação",
          getWonStatusPluralLabel(),
          "Cancelados",
          "Conversão",
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
      name: "Resumo Técnico",
      rows: [
        ["Visão técnica por orçamentista"],
        ["Escopo", scopeLabel],
        ["Gerado em", new Date()],
        [],
        ["Indicadores técnicos"],
        ["Indicador", "Valor", "Descrição"],
        [
          "Orçamentistas ativos",
          dashboardData.technicalOverview.summary.activeEstimators,
          "Quantidade de orçamentistas com produção no recorte",
        ],
        [
          "Cobertura técnica",
          dashboardData.technicalOverview.summary.coverageRate / 100,
          "Percentual de orçamentos com orçamentista atribuído",
        ],
        [
          "Orçamentos com orçamentista",
          dashboardData.technicalOverview.summary.budgetsWithEstimator,
          "Quantidade com responsabilidade técnica definida",
        ],
        [
          "Orçamentos sem orçamentista",
          dashboardData.technicalOverview.summary.budgetsWithoutEstimator,
          "Quantidade ainda sem atribuição técnica",
        ],
        [
          "Valor técnico monitorado",
          dashboardData.technicalOverview.summary.totalGrossValue,
          "Valor bruto dos orçamentos com orçamentista atribuído",
        ],
        [
          "Conversão técnica",
          dashboardData.technicalOverview.summary.conversionRate / 100,
          "Percentual de pedidos dentro da carteira técnica atribuída",
        ],
        [],
        ["Destaques técnicos", "Valor"],
        [
          "Top orçamentista por valor",
          topEstimatorByValue?.label ?? "Não informado",
        ],
        [
          "Top orçamentista por quantidade",
          topEstimatorByBudgetCount?.label ?? "Não informado",
        ],
        [
          "Top ticket médio técnico",
          topEstimatorByAverageTicket?.label ?? "Não informado",
        ],
        [
          "Última atividade técnica",
          mostRecentEstimator?.label ?? "Não informado",
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
      name: "Top Orçamentistas Valor",
      rows: [
        [
          "Posição",
          "Orçamentista",
          "Orçamentos",
          "Valor bruto",
          "Ticket médio",
          "Última atividade",
        ],
        ...dashboardData.technicalOverview.topEstimatorsByValue.map(
          (item, index) => [
            index + 1,
            item.label,
            item.budgetCount,
            item.grossValue,
            item.averageTicket,
            toSpreadsheetDateOrFallback(item.lastActivityAt, "Não informada"),
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
      name: "Top Orçamentistas Quantidade",
      rows: [
        [
          "Posição",
          "Orçamentista",
          "Orçamentos",
          "Valor bruto",
          "Conversão",
          "Última atividade",
        ],
        ...dashboardData.technicalOverview.topEstimatorsByBudgetCount.map(
          (item, index) => [
            index + 1,
            item.label,
            item.budgetCount,
            item.grossValue,
            item.conversionRate / 100,
            toSpreadsheetDateOrFallback(item.lastActivityAt, "Não informada"),
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
      name: "Última Atividade Técnica",
      rows: [
        [
          "Orçamentista",
          "Orçamentos",
          "Valor bruto",
          "Valor em negociação",
          "Última atividade",
        ],
        ...dashboardData.technicalOverview.recentEstimators.map((item) => [
          item.label,
          item.budgetCount,
          item.grossValue,
          item.negotiationGrossValue,
          toSpreadsheetDateOrFallback(item.lastActivityAt, "Não informada"),
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
  const [selectedInstallerId, setSelectedInstallerId] = useState("");
  const [selectedSalespersonId, setSelectedSalespersonId] = useState("");
  const [selectedStatusId, setSelectedStatusId] = useState("");
  const [selectedGrossValueMin, setSelectedGrossValueMin] = useState("");
  const [selectedGrossValueMax, setSelectedGrossValueMax] = useState("");
  const normalizedGrossValueRange = useMemo(
    () =>
      normalizeGrossValueDraftRange(
        selectedGrossValueMin,
        selectedGrossValueMax,
      ),
    [selectedGrossValueMax, selectedGrossValueMin],
  );

  const dashboardFilters = useMemo<DashboardSalespeopleFilters>(
    () => ({
      grossValueMax: normalizedGrossValueRange.grossValueMax,
      grossValueMin: normalizedGrossValueRange.grossValueMin,
      installerId: selectedInstallerId,
      sourceCompany,
      salespersonId: selectedSalespersonId,
      statusId: selectedStatusId,
      year: selectedYear,
      month: selectedMonth,
    }),
    [
      normalizedGrossValueRange.grossValueMax,
      normalizedGrossValueRange.grossValueMin,
      selectedInstallerId,
      selectedMonth,
      selectedSalespersonId,
      selectedStatusId,
      selectedYear,
      sourceCompany,
    ],
  );

  const budgetCatalogsQuery = useQuery({
    queryKey: ["dashboard", "budget-catalogs"],
    queryFn: getBudgetCatalogsRequest,
    enabled: !isEstimatorUser,
    gcTime: 1000 * 60 * 15,
    refetchOnWindowFocus: false,
    staleTime: 1000 * 60 * 5,
  });

  const grossValueRangeQuery = useQuery({
    queryKey: [
      "dashboard",
      "salespeople",
      "gross-value-range",
      dashboardFilters,
    ],
    queryFn: () => getDashboardGrossValueRangeRequest(dashboardFilters),
    enabled: !isEstimatorUser,
    refetchOnWindowFocus: false,
    staleTime: 1000 * 60 * 2,
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
          description="O dashboard comercial fica disponível apenas para administradores e perfis comerciais."
          title="Dashboard"
        />
        <Alert severity="info" sx={dashboardInfoAlertSx}>
          O perfil orçamentista não participa do dashboard comercial. Utilize a
          tela de orçamentos para operar no seu escopo técnico.
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

  const installerOptions = useMemo<BudgetCatalogItem[]>(() => {
    return [...(budgetCatalogsQuery.data?.installers ?? [])].sort((a, b) =>
      a.name.localeCompare(b.name),
    );
  }, [budgetCatalogsQuery.data?.installers]);

  const statusOptions = useMemo<BudgetCatalogItem[]>(() => {
    return [...(budgetCatalogsQuery.data?.statuses ?? [])].sort((a, b) =>
      a.name.localeCompare(b.name),
    );
  }, [budgetCatalogsQuery.data?.statuses]);

  const selectedSalespersonLabel = useMemo(() => {
    return (
      salespersonOptions.find(
        (item) => String(item.id) === selectedSalespersonId,
      )?.name ?? ""
    );
  }, [selectedSalespersonId, salespersonOptions]);

  const selectedInstallerLabel = useMemo(() => {
    return (
      installerOptions.find((item) => String(item.id) === selectedInstallerId)
        ?.name ?? ""
    );
  }, [installerOptions, selectedInstallerId]);

  const selectedMonthLabel = useMemo(() => {
    return (
      monthOptions.find((item) => item.value === selectedMonth)?.label ?? ""
    );
  }, [selectedMonth]);

  const selectedStatusLabel = useMemo(() => {
    return (
      statusOptions.find((item) => String(item.id) === selectedStatusId)
        ?.name ?? ""
    );
  }, [selectedStatusId, statusOptions]);

  const grossValueRange = useMemo(() => {
    const min = grossValueRangeQuery.data?.min ?? 0;
    const max = grossValueRangeQuery.data?.max ?? min;

    return {
      min,
      max: Math.max(min, max),
    };
  }, [grossValueRangeQuery.data?.max, grossValueRangeQuery.data?.min]);

  const isGrossValueSliderDisabled =
    grossValueRangeQuery.isLoading ||
    grossValueRange.max <= grossValueRange.min;

  const grossValueSliderStep = useMemo(() => {
    const span = grossValueRange.max - grossValueRange.min;
    if (span <= 0) {
      return 1;
    }
    if (span <= 1000) {
      return 1;
    }

    return Math.max(100, Math.round(span / 100));
  }, [grossValueRange.max, grossValueRange.min]);

  const grossValueSliderValue = useMemo<[number, number]>(() => {
    const resolvedMin =
      parseDraftCurrencyValue(normalizedGrossValueRange.grossValueMin) ??
      grossValueRange.min;
    const resolvedMax =
      parseDraftCurrencyValue(normalizedGrossValueRange.grossValueMax) ??
      grossValueRange.max;
    const nextMin = Math.min(
      Math.max(resolvedMin, grossValueRange.min),
      grossValueRange.max,
    );
    const nextMax = Math.max(
      Math.min(resolvedMax, grossValueRange.max),
      grossValueRange.min,
    );

    return nextMin <= nextMax ? [nextMin, nextMax] : [nextMax, nextMin];
  }, [
    grossValueRange.max,
    grossValueRange.min,
    normalizedGrossValueRange.grossValueMax,
    normalizedGrossValueRange.grossValueMin,
  ]);

  const selectedGrossValueLabel = useMemo(() => {
    if (
      dashboardFilters.grossValueMin.length === 0 &&
      dashboardFilters.grossValueMax.length === 0
    ) {
      return "todos os valores";
    }

    return `valor bruto de ${formatCurrencyRangeValue(
      dashboardFilters.grossValueMin,
      grossValueRange.min,
    )} até ${formatCurrencyRangeValue(
      dashboardFilters.grossValueMax,
      grossValueRange.max,
    )}`;
  }, [
    dashboardFilters.grossValueMax,
    dashboardFilters.grossValueMin,
    grossValueRange.max,
    grossValueRange.min,
  ]);

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
      if (isWonStatusLabel(item.name)) {
        currentMap.set(
          normalizeLookupKey(getWonStatusSingularLabel()),
          String(item.id),
        );
      }
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
        helper: "Quantidade de vendedores com orçamentos no recorte atual",
        icon: InsightsRoundedIcon,
      },
      {
        key: "total-budgets",
        label: "Orçamentos monitorados",
        value: String(summary?.totalBudgets ?? 0),
        helper: "Volume total de orçamentos no recorte atual",
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
        label: "Ticket médio",
        value: formatCompactCurrency(summary?.averageTicket ?? 0),
        helper: "Média de valor por orçamento do período",
        icon: TrendingUpRoundedIcon,
      },
      {
        key: "negotiation-gross-value",
        label: "Valor em negociação",
        value: formatCompactCurrency(summary?.totalNegotiationGrossValue ?? 0),
        helper: `${summary?.negotiationBudgets ?? 0} orçamento(s) ainda em carteira`,
        icon: AttachMoneyRoundedIcon,
      },
      {
        key: "pedido-conversion",
        label: "Conversão em pedido",
        value: formatPercentage(summary?.conversionRate ?? 0),
        helper: `${summary?.wonBudgets ?? 0} pedido(s) em ${summary?.totalBudgets ?? 0} orçamento(s)`,
        icon: AssignmentTurnedInRoundedIcon,
      },
      {
        key: "value-conversion",
        label: "Conversão por valor",
        value: formatPercentage(summary?.valueConversionRate ?? 0),
        helper: `${formatCompactCurrency(summary?.totalGrossValue ?? 0)} em valor bruto no período`,
        icon: AttachMoneyRoundedIcon,
      },
      {
        key: "stalled-budgets",
        label: "Orçamentos parados",
        value: String(summary?.stalledBudgetsCount ?? 0),
        helper: "Oportunidades em negociação sem atividade há 7 dias ou mais",
        icon: DescriptionRoundedIcon,
      },
    ];
    const technicalMetricCards: DashboardMetricCard[] = [
      {
        key: "active-estimators",
        label: "Orçamentistas ativos",
        value: String(technicalOverview?.summary.activeEstimators ?? 0),
        helper: "Quantidade de orçamentistas com produção no recorte atual",
        icon: InsightsRoundedIcon,
      },
      {
        key: "coverage-rate",
        label: "Cobertura técnica",
        value: formatPercentage(technicalOverview?.summary.coverageRate ?? 0),
        helper: `${technicalOverview?.summary.budgetsWithEstimator ?? 0} orçamento(s) com orçamentista definido`,
        icon: AssignmentTurnedInRoundedIcon,
      },
      {
        key: "technical-gross-value",
        label: "Valor técnico monitorado",
        value: formatCompactCurrency(
          technicalOverview?.summary.totalGrossValue ?? 0,
        ),
        helper:
          "Valor bruto dos orçamentos com responsabilidade técnica atribuída",
        icon: AttachMoneyRoundedIcon,
      },
      {
        key: "technical-stalled-budgets",
        label: "Orçamentos técnicos parados",
        value: String(technicalOverview?.summary.stalledBudgetsCount ?? 0),
        helper:
          "Orçamentos em negociação com orçamentista e sem atividade há 7 dias ou mais",
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

  const handleNormalizeGrossValueInputs = () => {
    const normalizedRange = normalizeGrossValueDraftRange(
      selectedGrossValueMin,
      selectedGrossValueMax,
    );

    setSelectedGrossValueMin(normalizedRange.grossValueMin);
    setSelectedGrossValueMax(normalizedRange.grossValueMax);
  };

  const handleGrossValueSliderChange = (
    _event: Event,
    value: number | number[],
  ) => {
    if (!Array.isArray(value)) {
      return;
    }

    const [nextMin, nextMax] = value;
    setSelectedGrossValueMin(String(Math.round(nextMin)));
    setSelectedGrossValueMax(String(Math.round(nextMax)));
  };

  const handleOpenBudgetList = ({
    budgetNumber,
    grossValueMax,
    grossValueMin,
    installerId,
    projectId,
    projectName,
    salespersonId,
    statusId,
  }: Partial<BudgetListNavigationOptions> = {}) => {
    const searchParams = buildBudgetListSearchParams({
      budgetNumber,
      grossValueMax: grossValueMax ?? dashboardFilters.grossValueMax,
      grossValueMin: grossValueMin ?? dashboardFilters.grossValueMin,
      installerId: installerId ?? selectedInstallerId,
      month: selectedMonth,
      projectId,
      projectName,
      salespersonId: salespersonId ?? selectedSalespersonId,
      sourceCompany,
      statusId: statusId ?? selectedStatusId,
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
          selectedGrossValueLabel,
        ),
      ]),
      "",
      createCsvLine([
        "Resumo",
        "Vendedores ativos",
        "Orçamentos monitorados",
        "Valor bruto total",
        "Ticket médio",
        "Valor em negociação",
        "Conversão",
        "Conversão por valor",
        "Orçamentos parados",
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
        "Orçamentos",
        "Valor bruto",
        "Ticket médio",
        "Última atividade comercial",
      ]),
      ...dashboardData.topSalespeopleByValue.map((item) =>
        createCsvLine([
          "Top por valor",
          item.label,
          item.budgetCount,
          currencyFormatter.format(item.grossValue),
          currencyFormatter.format(item.averageTicket),
          formatDateOrFallback(item.lastActivityAt, "Não informada"),
        ]),
      ),
      "",
      createCsvLine([
        "Carteira em negociação",
        "Vendedor",
        "Em aberto",
        "Valor",
        "Parados",
        "Última atividade comercial",
      ]),
      ...dashboardData.negotiationPipeline.map((item) =>
        createCsvLine([
          "Carteira em negociação",
          item.label,
          item.negotiationBudgetCount,
          currencyFormatter.format(item.negotiationGrossValue),
          item.stalledBudgetCount,
          formatDateOrFallback(item.lastActivityAt, "Não informada"),
        ]),
      ),
      "",
      createCsvLine([
        "Orçamentos parados",
        "Orçamento",
        "Vendedor",
        "Obra",
        "Construtora",
        "Valor",
        "Dias parados",
      ]),
      ...dashboardData.staleBudgets.map((item) =>
        createCsvLine([
          "Orçamentos parados",
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
        "Evolução mensal",
        "Mês",
        "Orçamentos",
        "Valor bruto",
        getWonStatusPluralLabel(),
        "Valor convertido",
      ]),
      ...dashboardData.monthlyEvolution.map((item) =>
        createCsvLine([
          "Evolução mensal",
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
        "Orçamentos",
        "Valor bruto",
        "Conversão",
        "Conversão por valor",
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
        "Orçamentos",
        "Valor bruto",
        "Conversão",
        "Conversão por valor",
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
        "Ticket médio",
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
        "Tempo médio de fechamento",
        "Recorte",
        "Tempo médio",
        "Valor bruto",
        "Orçamentos fechados",
      ]),
      ...dashboardData.averageClosingTimes.map((item) =>
        createCsvLine([
          "Tempo médio de fechamento",
          item.label,
          formatClosingDays(item.averageClosingDays),
          currencyFormatter.format(item.grossValue),
          item.budgetCount,
        ]),
      ),
      "",
      createCsvLine([
        "Resumo técnico",
        "Orçamentistas ativos",
        "Cobertura técnica",
        "Com orçamentista",
        "Sem orçamentista",
        "Valor técnico",
        "Ticket médio",
        "Valor em negociação",
        "Conversão técnica",
        "Parados",
      ]),
      createCsvLine([
        "Resumo técnico",
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
        "Top orçamentistas por valor",
        "Orçamentista",
        "Orçamentos",
        "Valor bruto",
        "Ticket médio",
        "Última atividade",
      ]),
      ...dashboardData.technicalOverview.topEstimatorsByValue.map((item) =>
        createCsvLine([
          "Top orçamentistas por valor",
          item.label,
          item.budgetCount,
          currencyFormatter.format(item.grossValue),
          currencyFormatter.format(item.averageTicket),
          formatDateOrFallback(item.lastActivityAt, "Não informada"),
        ]),
      ),
      "",
      createCsvLine([
        "Top orçamentistas por quantidade",
        "Orçamentista",
        "Orçamentos",
        "Valor bruto",
        "Conversão",
        "Última atividade",
      ]),
      ...dashboardData.technicalOverview.topEstimatorsByBudgetCount.map(
        (item) =>
          createCsvLine([
            "Top orçamentistas por quantidade",
            item.label,
            item.budgetCount,
            currencyFormatter.format(item.grossValue),
            formatPercentage(item.conversionRate),
            formatDateOrFallback(item.lastActivityAt, "Não informada"),
          ]),
      ),
      "",
      createCsvLine([
        "Última atividade técnica",
        "Orçamentista",
        "Orçamentos",
        "Valor bruto",
        "Valor em negociação",
        "Última atividade",
      ]),
      ...dashboardData.technicalOverview.recentEstimators.map((item) =>
        createCsvLine([
          "Última atividade técnica",
          item.label,
          item.budgetCount,
          currencyFormatter.format(item.grossValue),
          currencyFormatter.format(item.negotiationGrossValue),
          formatDateOrFallback(item.lastActivityAt, "Não informada"),
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
      dashboardFilters.grossValueMin || "min-livre",
      dashboardFilters.grossValueMax || "max-livre",
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
      dashboardFilters.grossValueMin || "min-livre",
      dashboardFilters.grossValueMax || "max-livre",
    ]
      .join("-")
      .replaceAll(" ", "-")
      .toLowerCase();
    const sheets = buildDashboardWorkbookSheets({
      dashboardData,
      selectedMonth,
      selectedGrossValueLabel,
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
        description="Painel administrativo com leitura comercial por vendedor e visão técnica separada por orçamentista."
        title="Dashboard administrativo"
      />

      <SectionCard
        description="Refine o recorte do dashboard sem apertar o cabeçalho da tela."
        title="Filtros do dashboard"
      >
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
          <FilterField label="Empresa">
            <TextField
              onChange={(event) =>
                setSourceCompany(event.target.value as DashboardCompanyFilter)
              }
              select
              size="small"
              sx={compactFilterFieldSx}
              value={sourceCompany}
            >
              <MenuItem value="">Todas as empresas</MenuItem>
              <MenuItem value="Rocktec">ROCKTEC</MenuItem>
              <MenuItem value="Trox">TROX</MenuItem>
            </TextField>
          </FilterField>
          <FilterField label="Ano">
            <TextField
              onChange={(event) => setSelectedYear(event.target.value)}
              select
              size="small"
              sx={compactFilterFieldSx}
              value={selectedYear}
            >
              <MenuItem value="">Todos os anos</MenuItem>
              {availableYears.map((year) => (
                <MenuItem key={year} value={year}>
                  {year}
                </MenuItem>
              ))}
            </TextField>
          </FilterField>
          <FilterField label="Mês">
            <TextField
              onChange={(event) => setSelectedMonth(event.target.value)}
              select
              size="small"
              sx={compactFilterFieldSx}
              value={selectedMonth}
            >
              <MenuItem value="">Todos os meses</MenuItem>
              {monthOptions.map((month) => (
                <MenuItem key={month.value} value={month.value}>
                  {month.label}
                </MenuItem>
              ))}
            </TextField>
          </FilterField>
          <FilterField label="Vendedor">
            <TextField
              onChange={(event) => setSelectedSalespersonId(event.target.value)}
              select
              size="small"
              sx={compactFilterFieldSx}
              value={selectedSalespersonId}
            >
              <MenuItem value="">Todos os vendedores</MenuItem>
              {salespersonOptions.map((salesperson) => (
                <MenuItem key={salesperson.id} value={String(salesperson.id)}>
                  {salesperson.name}
                </MenuItem>
              ))}
            </TextField>
          </FilterField>
          <FilterField label="Instalador">
            <TextField
              onChange={(event) => setSelectedInstallerId(event.target.value)}
              select
              size="small"
              sx={compactFilterFieldSx}
              value={selectedInstallerId}
            >
              <MenuItem value="">Todos os instaladores</MenuItem>
              {installerOptions.map((installer) => (
                <MenuItem key={installer.id} value={String(installer.id)}>
                  {installer.name}
                </MenuItem>
              ))}
            </TextField>
          </FilterField>
          <FilterField label="Status">
            <TextField
              onChange={(event) => setSelectedStatusId(event.target.value)}
              select
              size="small"
              sx={compactFilterFieldSx}
              value={selectedStatusId}
            >
              <MenuItem value="">Todos os status</MenuItem>
              {statusOptions.map((status) => (
                <MenuItem key={status.id} value={String(status.id)}>
                  {status.name}
                </MenuItem>
              ))}
            </TextField>
          </FilterField>
          <Box sx={dashboardValueRangeGridSx}>
            <FilterField label="Valor minimo">
              <TextField
                onBlur={handleNormalizeGrossValueInputs}
                onChange={(event) =>
                  setSelectedGrossValueMin(
                    parseCurrencyInputToNumericString(event.target.value),
                  )
                }
                placeholder={currencyFormatter.format(grossValueRange.min)}
                size="small"
                slotProps={{
                  htmlInput: {
                    inputMode: "numeric",
                  },
                }}
                sx={compactFilterFieldSx}
                value={formatCurrencyInputValue(selectedGrossValueMin)}
              />
            </FilterField>
            <FilterField label="Valor maximo">
              <TextField
                onBlur={handleNormalizeGrossValueInputs}
                onChange={(event) =>
                  setSelectedGrossValueMax(
                    parseCurrencyInputToNumericString(event.target.value),
                  )
                }
                placeholder={currencyFormatter.format(grossValueRange.max)}
                size="small"
                slotProps={{
                  htmlInput: {
                    inputMode: "numeric",
                  },
                }}
                sx={compactFilterFieldSx}
                value={formatCurrencyInputValue(selectedGrossValueMax)}
              />
            </FilterField>
          </Box>
          <Box
            sx={{
              gridColumn: "1 / -1",
              px: { md: 1, xs: 0 },
            }}
          >
            <Typography
              color="text.secondary"
              sx={{ display: "block", mb: 1, fontSize: "0.82rem" }}
              variant="caption"
            >
              Faixa de valor bruto
            </Typography>
            <Slider
              disableSwap
              disabled={isGrossValueSliderDisabled}
              marks={[
                { value: grossValueRange.min },
                { value: grossValueRange.max },
              ]}
              max={grossValueRange.max}
              min={grossValueRange.min}
              onChange={handleGrossValueSliderChange}
              step={grossValueSliderStep}
              value={grossValueSliderValue}
              valueLabelDisplay="auto"
              valueLabelFormat={(value) => currencyFormatter.format(value)}
            />
            <Typography color="text.secondary" variant="caption">
              {`${currencyFormatter.format(grossValueRange.min)} até ${currencyFormatter.format(grossValueRange.max)}`}
            </Typography>
          </Box>
        </Box>
        {grossValueRangeQuery.isError ? (
          <Alert severity="warning" sx={{ mt: 2 }} variant="outlined">
            Não foi possível carregar a faixa de valor bruto do dashboard.
          </Alert>
        ) : null}
      </SectionCard>

      {dashboardQuery.isLoading || isDashboardRefreshing ? (
        <LinearProgress sx={dashboardLoaderSx} />
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
          sx={dashboardFeedbackAlertSx}
          variant="outlined"
        >
          {dashboardErrorMessage}
        </Alert>
      ) : null}

      {!dashboardQuery.isLoading && !dashboardQuery.isError ? (
        <Alert severity="info" sx={dashboardInfoAlertSx} variant="outlined">
          {`Os indicadores abaixo refletem o recorte de ${buildDashboardScopeLabel(
            sourceCompany,
            selectedYear,
            selectedMonth,
            selectedSalespersonLabel,
            selectedGrossValueLabel,
          )}.`}
          {isDashboardRefreshing ? " Atualizando dados em segundo plano." : ""}
        </Alert>
      ) : null}

      {!dashboardQuery.isLoading && !dashboardQuery.isError ? (
        <SectionCard
          description="Atalhos para aprofundar a análise do dashboard e exportar o recorte atual com leitura mais executiva."
          title="Ações rápidas"
        >
          <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
            <Box sx={{ display: "flex", flexWrap: "wrap", gap: 1 }}>
              <Chip
                color="primary"
                label={`Empresa ${sourceCompany || "todas"}`}
                size="small"
                variant="outlined"
              />
              <Chip
                label={`Ano ${selectedYear || "todos"}`}
                size="small"
                variant="outlined"
              />
              <Chip
                label={`Mês ${selectedMonthLabel || "todos"}`}
                size="small"
                variant="outlined"
              />
              <Chip
                label={`Vendedor ${selectedSalespersonLabel || "todos"}`}
                size="small"
                variant="outlined"
              />
              <Chip
                label={`Instalador ${selectedInstallerLabel || "todos"}`}
                size="small"
                variant="outlined"
              />
              <Chip
                label={`Status ${selectedStatusLabel || "todos"}`}
                size="small"
                variant="outlined"
              />
              <Chip
                label={`Valor ${selectedGrossValueLabel}`}
                size="small"
                variant="outlined"
              />
            </Box>
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
                        normalizeLookupKey("Em Negociação"),
                      ) ?? "",
                  })
                }
                startIcon={<OpenInNewRoundedIcon />}
                variant="outlined"
              >
                Ver em negociação
              </Button>
              <Button
                onClick={() =>
                  handleOpenBudgetList({
                    statusId:
                      statusIdByNormalizedName.get(
                        normalizeLookupKey(getWonStatusSingularLabel()),
                      ) ?? "",
                  })
                }
                startIcon={<OpenInNewRoundedIcon />}
                variant="outlined"
              >
                Ver fechados
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
            <SectionCard
              sx={{
                background:
                  "radial-gradient(circle at top right, rgba(59,130,246,0.1) 0%, transparent 30%), linear-gradient(135deg, rgba(30,58,138,0.06) 0%, rgba(14,165,233,0.03) 100%)",
              }}
            >
              <Box sx={{ display: "flex", flexDirection: "column", gap: 1.75 }}>
                <Box
                  sx={{
                    alignItems: "center",
                    display: "flex",
                    gap: 1.25,
                    justifyContent: "space-between",
                  }}
                >
                  <Chip
                    color="primary"
                    icon={<metric.icon />}
                    label="Resumo executivo"
                    sx={{ alignSelf: "flex-start" }}
                    variant="outlined"
                  />
                  <Box
                    sx={{
                      alignItems: "center",
                      background:
                        "linear-gradient(135deg, rgba(30,58,138,0.12) 0%, rgba(14,165,233,0.12) 100%)",
                      border: "1px solid rgba(30,58,138,0.14)",
                      borderRadius: 3,
                      color: "primary.main",
                      display: "inline-flex",
                      height: 44,
                      justifyContent: "center",
                      width: 44,
                    }}
                  >
                    <metric.icon />
                  </Box>
                </Box>
                <Typography
                  color="text.secondary"
                  sx={{ fontWeight: 700 }}
                  variant="body2"
                >
                  {metric.label}
                </Typography>
                <Typography
                  sx={{
                    color: "var(--app-accent-text)",
                    fontWeight: 850,
                    letterSpacing: "-0.03em",
                  }}
                  variant="h3"
                >
                  {metric.value}
                </Typography>
                <Typography
                  color="text.secondary"
                  sx={{ lineHeight: 1.65 }}
                  variant="body2"
                >
                  {metric.helper}
                </Typography>
              </Box>
            </SectionCard>
          </Box>
        ))}
      </Box>

      {!dashboardQuery.isLoading && !dashboardQuery.isError ? (
        <Alert severity="info" sx={dashboardInfoAlertSx} variant="outlined">
          A leitura comercial e a leitura técnica aparecem separadas para evitar
          mistura entre vendedor e orçamentista nos indicadores gerenciais e nos
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
            <SectionCard
              sx={{
                background:
                  "radial-gradient(circle at top right, rgba(168,85,247,0.08) 0%, transparent 28%), linear-gradient(135deg, rgba(30,58,138,0.05) 0%, rgba(168,85,247,0.03) 100%)",
              }}
            >
              <Box sx={{ display: "flex", flexDirection: "column", gap: 1.75 }}>
                <Box
                  sx={{
                    alignItems: "center",
                    display: "flex",
                    gap: 1.25,
                    justifyContent: "space-between",
                  }}
                >
                  <Chip
                    color="secondary"
                    icon={<metric.icon />}
                    label="Leitura técnica"
                    sx={{ alignSelf: "flex-start" }}
                    variant="outlined"
                  />
                  <Box
                    sx={{
                      alignItems: "center",
                      background:
                        "linear-gradient(135deg, rgba(126,34,206,0.12) 0%, rgba(30,58,138,0.08) 100%)",
                      border: "1px solid rgba(126,34,206,0.14)",
                      borderRadius: 3,
                      color: "secondary.main",
                      display: "inline-flex",
                      height: 44,
                      justifyContent: "center",
                      width: 44,
                    }}
                  >
                    <metric.icon />
                  </Box>
                </Box>
                <Typography
                  color="text.secondary"
                  sx={{ fontWeight: 700 }}
                  variant="body2"
                >
                  {metric.label}
                </Typography>
                <Typography
                  sx={{
                    color: "var(--app-accent-text)",
                    fontWeight: 850,
                    letterSpacing: "-0.03em",
                  }}
                  variant="h3"
                >
                  {metric.value}
                </Typography>
                <Typography
                  color="text.secondary"
                  sx={{ lineHeight: 1.65 }}
                  variant="body2"
                >
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
          description="Ranking técnico por valor bruto dos orçamentos atribuídos a cada orçamentista."
          title="Top orçamentistas por valor"
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
                        label={`Última atividade ${formatDateOrFallback(item.lastActivityAt, "Não informada")}`}
                        size="small"
                        variant="outlined"
                      />
                    </Box>
                  </Box>
                ),
              )}
            </Box>
          ) : (
            <Alert severity="info" sx={dashboardInfoAlertSx} variant="outlined">
              Nenhum orçamentista encontrado no recorte atual.
            </Alert>
          )}
        </SectionCard>

        <SectionCard
          description="Orçamentistas com maior volume de orçamentos atribuídos no período."
          title="Top orçamentistas por quantidade"
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
                        label={`Conversão ${formatPercentage(item.conversionRate)}`}
                        size="small"
                        variant="outlined"
                      />
                      <Chip
                        label={`Negociação ${currencyFormatter.format(item.negotiationGrossValue)}`}
                        size="small"
                        variant="outlined"
                      />
                    </Box>
                  </Box>
                ),
              )}
            </Box>
          ) : (
            <Alert severity="info" sx={dashboardInfoAlertSx} variant="outlined">
              Nenhum orçamentista encontrado no recorte atual.
            </Alert>
          )}
        </SectionCard>

        <SectionCard
          description="Última atividade registrada para cada orçamentista dentro do recorte selecionado."
          title="Última atividade técnica"
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
                      {`${item.budgetCount} orçamento(s) · ${currencyFormatter.format(item.grossValue)}`}
                    </Typography>
                  </Box>
                  <Chip
                    icon={<InsightsRoundedIcon />}
                    label={formatDateOrFallback(
                      item.lastActivityAt,
                      "Não informada",
                    )}
                    size="small"
                    variant="outlined"
                  />
                </Box>
              ))}
            </Box>
          ) : (
            <Alert severity="info" sx={dashboardInfoAlertSx} variant="outlined">
              Nenhuma atividade técnica encontrada para o filtro atual.
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
                      label={`Última atividade comercial ${formatDateOrFallback(item.lastActivityAt, "Não informada")}`}
                      size="small"
                      variant="outlined"
                    />
                  </Box>
                </Box>
              ))}
            </Box>
          ) : (
            <Alert severity="info" sx={dashboardInfoAlertSx} variant="outlined">
              Nenhum orçamento encontrado para o filtro atual.
            </Alert>
          )}
        </SectionCard>

        <SectionCard
          description="Top 10 vendedores ordenados pelo volume de orçamentos."
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
                      label={`Conversão ${formatPercentage(item.conversionRate)}`}
                      size="small"
                      variant="outlined"
                    />
                    <Chip
                      label={`Última atividade comercial ${formatDateOrFallback(item.lastActivityAt, "Não informada")}`}
                      size="small"
                      variant="outlined"
                    />
                  </Box>
                </Box>
              ))}
            </Box>
          ) : (
            <Alert severity="info" sx={dashboardInfoAlertSx} variant="outlined">
              Nenhum orçamento encontrado para o filtro atual.
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
          description="Construtoras com maior volume financeiro no recorte atual, destacando conversão por quantidade e por valor."
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
                    <Button
                      onClick={() =>
                        handleOpenBudgetList({
                          projectId:
                            item.projectId !== null &&
                            item.projectId !== undefined
                              ? String(item.projectId)
                              : undefined,
                          projectName:
                            (item.projectId === null ||
                              item.projectId === undefined) &&
                            item.label.trim().toLowerCase() !==
                              "sem obra vinculada"
                              ? item.label
                              : undefined,
                        })
                      }
                      size="small"
                      sx={{
                        fontWeight: 700,
                        justifyContent: "flex-start",
                        px: 0,
                        textTransform: "none",
                      }}
                      variant="text"
                    >
                      {`${index + 1}. ${item.label}`}
                    </Button>
                    <Box sx={{ display: "flex", gap: 1 }}>
                      <Chip
                        label={`${item.budgetCount} orc.`}
                        size="small"
                        variant="outlined"
                      />
                      <Button
                        onClick={() =>
                          handleOpenBudgetList({
                            projectId:
                              item.projectId !== null &&
                              item.projectId !== undefined
                                ? String(item.projectId)
                                : undefined,
                            projectName:
                              (item.projectId === null ||
                                item.projectId === undefined) &&
                              item.label.trim().toLowerCase() !==
                                "sem obra vinculada"
                                ? item.label
                                : undefined,
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
                      label={`Última atividade comercial ${formatDateOrFallback(item.lastActivityAt, "Não informada")}`}
                      size="small"
                      variant="outlined"
                    />
                  </Box>
                </Box>
              ))}
            </Box>
          ) : (
            <Alert severity="info" sx={dashboardInfoAlertSx} variant="outlined">
              Nenhuma construtora encontrada para o filtro atual.
            </Alert>
          )}
        </SectionCard>

        <SectionCard
          description="Obras com maior valor bruto no período, ajudando a identificar onde está a melhor concentração comercial."
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
                    <Button
                      onClick={() =>
                        handleOpenBudgetList({
                          projectId:
                            item.projectId !== null &&
                            item.projectId !== undefined
                              ? String(item.projectId)
                              : undefined,
                        })
                      }
                      size="small"
                      sx={{
                        fontWeight: 700,
                        justifyContent: "flex-start",
                        px: 0,
                        textAlign: "left",
                        textTransform: "none",
                      }}
                      variant="text"
                    >
                      {`${index + 1}. ${item.label}`}
                    </Button>
                    <Box sx={{ display: "flex", gap: 1 }}>
                      <Chip
                        label={`${item.budgetCount} orc.`}
                        size="small"
                        variant="outlined"
                      />
                      <Button
                        onClick={() =>
                          handleOpenBudgetList({
                            projectId:
                              item.projectId !== null &&
                              item.projectId !== undefined
                                ? String(item.projectId)
                                : undefined,
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
                      label={`Última atividade comercial ${formatDateOrFallback(item.lastActivityAt, "Não informada")}`}
                      size="small"
                      variant="outlined"
                    />
                  </Box>
                </Box>
              ))}
            </Box>
          ) : (
            <Alert severity="info" sx={dashboardInfoAlertSx} variant="outlined">
              Nenhuma obra encontrada para o filtro atual.
            </Alert>
          )}
        </SectionCard>

        <SectionCard
          description="Principais motivos de perda por impacto financeiro para orientar as ações corretivas do time comercial."
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
            <Alert severity="info" sx={dashboardInfoAlertSx} variant="outlined">
              Nenhum motivo de perda encontrado para o filtro atual.
            </Alert>
          )}
        </SectionCard>

        <SectionCard
          description="Tempo médio entre o envio e o fechamento dos orçamentos finalizados no recorte atual."
          title="Tempo médio de fechamento"
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
            <Alert severity="info" sx={dashboardInfoAlertSx} variant="outlined">
              Nenhum fechamento encontrado para calcular o tempo médio no filtro
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
          description="Comparativo entre os dois meses mais recentes do recorte atual para identificar aceleração ou perda de ritmo."
          title="Tendência mensal"
        >
          {dashboardInsights.monthComparison !== null ? (
            <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
              <Alert
                severity="info"
                sx={dashboardInfoAlertSx}
                variant="outlined"
              >
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
                    Orçamentos
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
                    Conversão
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
            <Alert severity="info" sx={dashboardInfoAlertSx} variant="outlined">
              São necessários pelo menos dois meses no recorte atual para montar
              o comparativo de tendência.
            </Alert>
          )}
        </SectionCard>

        <SectionCard
          description="Vendedores com melhor conversão no período, priorizando quem tem pelo menos dois orçamentos no recorte."
          title="Top conversão"
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
            <Alert severity="info" sx={dashboardInfoAlertSx} variant="outlined">
              Nenhum vendedor elegível para o ranking de conversão no filtro
              atual.
            </Alert>
          )}
        </SectionCard>

        <SectionCard
          description="Vendedores com maior ticket médio no período, considerando base mínima para reduzir distorção."
          title="Top ticket médio"
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
                        label={`Conversão ${formatPercentage(item.conversionRate)}`}
                        size="small"
                        variant="outlined"
                      />
                    </Box>
                  </Box>
                ),
              )}
            </Box>
          ) : (
            <Alert severity="info" sx={dashboardInfoAlertSx} variant="outlined">
              Nenhum vendedor elegível para o ranking de conversão no filtro
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
          description="Carteira ainda em negociação, com foco no valor e no volume por vendedor."
          title="Carteira em negociação"
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
                                normalizeLookupKey("Em Negociação"),
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
                      label={`Última atividade comercial ${formatDateOrFallback(item.lastActivityAt, "Não informada")}`}
                      size="small"
                      variant="outlined"
                    />
                  </Box>
                </Box>
              ))}
            </Box>
          ) : (
            <Alert severity="info" sx={dashboardInfoAlertSx} variant="outlined">
              Nenhuma carteira em negociação encontrada para o filtro atual.
            </Alert>
          )}
        </SectionCard>

        <SectionCard
          description="Comparativo do funil principal por vendedor com foco em negociação, pedidos e perdas."
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
                        label={`${getWonStatusSingularLabel()} ${item.wonBudgets}`}
                        size="small"
                        variant="outlined"
                      />
                      <Chip
                        color="warning"
                        label={`Negociação ${item.negotiationBudgets}`}
                        size="small"
                        variant="outlined"
                      />
                      <Chip
                        label={`Cancelado ${item.lostBudgets}`}
                        size="small"
                        variant="outlined"
                      />
                      <Chip
                        label={`Conversão ${formatPercentage(item.conversionRate)}`}
                        size="small"
                        variant="outlined"
                      />
                    </Box>
                  </Box>
                );
              })}
            </Box>
          ) : (
            <Alert severity="info" sx={dashboardInfoAlertSx} variant="outlined">
              Nenhum dado de funil encontrado para o filtro atual.
            </Alert>
          )}
        </SectionCard>

        <SectionCard
          description="Leitura rápida da última atividade comercial registrada por vendedor dentro do recorte."
          title="Última atividade comercial"
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
                      {`${item.budgetCount} orçamento(s) · ${currencyFormatter.format(item.grossValue)}`}
                    </Typography>
                  </Box>
                  <Chip
                    icon={<InsightsRoundedIcon />}
                    label={formatDateOrFallback(
                      item.lastActivityAt,
                      "Não informada",
                    )}
                    size="small"
                    variant="outlined"
                  />
                </Box>
              ))}
            </Box>
          ) : (
            <Alert severity="info" sx={dashboardInfoAlertSx} variant="outlined">
              Nenhuma atividade comercial encontrada para o filtro atual.
            </Alert>
          )}
        </SectionCard>

        <SectionCard
          description="Oportunidades em negociação sem atividade comercial recente, priorizadas pelos casos mais antigos."
          title="Orçamentos parados"
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
                        label={`Última atividade comercial ${formatDateOrFallback(item.lastActivityAt, "Não informada")}`}
                        size="small"
                        variant="outlined"
                      />
                    </Box>
                  </Box>
                </Box>
              ))}
            </Box>
          ) : (
            <Alert
              severity="success"
              sx={dashboardFeedbackAlertSx}
              variant="outlined"
            >
              Nenhum orçamento parado encontrado para o filtro atual.
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
          description="Evolução do volume orçado e do valor convertido ao longo do tempo dentro do recorte atual."
          title="Evolução mensal"
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
                      sx={{ ...dashboardLoaderSx, mt: 1 }}
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
                        label={`${getWonStatusSingularLabel()} ${currencyFormatter.format(item.wonGrossValue)}`}
                        size="small"
                        variant="outlined"
                      />
                      <Chip
                        label={`${item.wonBudgetCount} fechado(s)`}
                        size="small"
                        variant="outlined"
                      />
                    </Box>
                  </Box>
                );
              })}
            </Box>
          ) : (
            <Alert severity="info" sx={dashboardInfoAlertSx} variant="outlined">
              Nenhuma evolução mensal encontrada para o filtro atual.
            </Alert>
          )}
        </SectionCard>

        <SectionCard
          description="Leitura consolidada do funil principal dentro do período selecionado."
          title="Resumo da carteira"
        >
          <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
            {[
              {
                color: "success.main",
                label: getWonStatusPluralLabel(),
                value: dashboardData?.wonBudgets ?? 0,
              },
              {
                color: "warning.main",
                label: "Em negociação",
                value: dashboardData?.negotiationBudgets ?? 0,
              },
              {
                color: "text.secondary",
                label: "Cancelados",
                value: dashboardData?.lostBudgets ?? 0,
              },
              {
                color: "primary.main",
                label: "Conversão",
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
