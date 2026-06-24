type SpreadsheetCellValue = string | number | Date | null;

type SpreadsheetCellFormat = "currency" | "date" | "integer";

type SpreadsheetColumn<TItem> = {
  header: string;
  format?: SpreadsheetCellFormat;
  value: (item: TItem, index: number) => SpreadsheetCellValue;
};

type ExportSheetToExcelOptions<TItem> = {
  columns: Array<SpreadsheetColumn<TItem>>;
  fileName: string;
  items: TItem[];
  sheetName: string;
};

const spreadsheetDateFormatter = new Intl.DateTimeFormat("pt-BR", {
  dateStyle: "short",
});

function getCellDisplayLength(value: SpreadsheetCellValue) {
  if (value === null) {
    return 0;
  }

  if (value instanceof Date) {
    return spreadsheetDateFormatter.format(value).length;
  }

  return String(value).length;
}

function buildWorksheetColumnWidths(rows: SpreadsheetCellValue[][]) {
  const maxColumns = rows.reduce(
    (currentMax, row) => Math.max(currentMax, row.length),
    0,
  );

  return Array.from({ length: maxColumns }, (_, columnIndex) => {
    const largestWidth = rows.reduce((currentMax, row) => {
      const cellValue = row[columnIndex] ?? null;

      return Math.max(currentMax, getCellDisplayLength(cellValue));
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

function getHeaderCellStyle(): import("xlsx-js-style").CellStyle {
  return {
    alignment: {
      horizontal: "center",
      vertical: "center",
      wrapText: true,
    },
    border: getSpreadsheetBorderStyle(),
    fill: {
      fgColor: { rgb: "FF4472C4" },
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

function downloadBlob(fileBlob: Blob, fileName: string) {
  const downloadUrl = window.URL.createObjectURL(fileBlob);
  const downloadLink = document.createElement("a");

  downloadLink.href = downloadUrl;
  downloadLink.download = fileName;
  document.body.appendChild(downloadLink);
  downloadLink.click();
  document.body.removeChild(downloadLink);
  window.URL.revokeObjectURL(downloadUrl);
}

async function applyWorksheetFormatting(
  XLSX: typeof import("xlsx-js-style"),
  worksheet: import("xlsx-js-style").WorkSheet,
  columns: Array<SpreadsheetColumn<unknown>>,
  rows: SpreadsheetCellValue[][],
) {
  const worksheetRange = worksheet["!ref"];
  if (!worksheetRange) {
    return;
  }

  worksheet["!cols"] = buildWorksheetColumnWidths(rows);
  worksheet["!rows"] = rows.map(() => ({ hpt: 22 }));

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

      cell.s = rowIndex === 0 ? getHeaderCellStyle() : bodyCellStyle;
    }
  }

  worksheet["!autofilter"] = {
    ref: XLSX.utils.encode_range({
      s: { c: decodedRange.s.c, r: 0 },
      e: { c: decodedRange.e.c, r: decodedRange.e.r },
    }),
  };

  columns.forEach((column, columnIndex) => {
    if (!column.format) {
      return;
    }

    for (let rowIndex = 1; rowIndex <= decodedRange.e.r; rowIndex += 1) {
      const cellAddress = XLSX.utils.encode_cell({
        c: columnIndex,
        r: rowIndex,
      });
      const cell = worksheet[cellAddress];
      if (!cell) {
        continue;
      }

      if (column.format === "date") {
        if (cell.t === "d" || cell.t === "n") {
          cell.z = getSpreadsheetNumberFormat(column.format);
        }
        continue;
      }

      if (cell.t === "n") {
        cell.z = getSpreadsheetNumberFormat(column.format);
      }
    }
  });
}

function sanitizeFileNamePart(value: string) {
  return value
    .trim()
    .replace(/\s+/g, "-")
    .replace(/[^a-zA-Z0-9-_]/g, "")
    .toLowerCase();
}

export async function exportSheetToExcel<TItem>({
  columns,
  fileName,
  items,
  sheetName,
}: ExportSheetToExcelOptions<TItem>) {
  const XLSX = await import("xlsx-js-style");
  const workbook = XLSX.utils.book_new();
  const rows: SpreadsheetCellValue[][] = [
    columns.map((column) => column.header),
    ...items.map((item, index) => columns.map((column) => column.value(item, index))),
  ];
  const worksheet = XLSX.utils.aoa_to_sheet(rows, {
    cellDates: true,
  });

  await applyWorksheetFormatting(
    XLSX,
    worksheet,
    columns as Array<SpreadsheetColumn<unknown>>,
    rows,
  );
  XLSX.utils.book_append_sheet(workbook, worksheet, sheetName);

  const workbookArray = XLSX.write(workbook, {
    bookType: "xlsx",
    cellStyles: true,
    type: "array",
  });
  const xlsxBlob = new Blob([workbookArray], {
    type: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
  });
  const normalizedFileName = sanitizeFileNamePart(fileName);

  downloadBlob(xlsxBlob, `${normalizedFileName || "exportacao"}.xlsx`);
}
