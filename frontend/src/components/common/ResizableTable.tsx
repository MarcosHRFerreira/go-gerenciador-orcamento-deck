import { Box, TableCell } from "@mui/material";
import type { SxProps, Theme } from "@mui/material/styles";
import { useCallback, useEffect, useMemo, useState } from "react";
import type { MouseEvent, ReactNode } from "react";

const defaultMinColumnWidth = 96;

export type ResizableColumnDefinition = {
  key: string;
  width: number;
  minWidth?: number;
  maxWidth?: number;
};

type ColumnWidthMap = Record<string, number>;

type ResizableTableHeadCellProps = {
  align?: "center" | "inherit" | "justify" | "left" | "right";
  children: ReactNode;
  onResizeStart: (event: MouseEvent<HTMLSpanElement>) => void;
  sx?: SxProps<Theme>;
  width: number;
};

function readStoredWidths(
  storageKey: string,
  defaultWidths: ColumnWidthMap,
): ColumnWidthMap {
  if (typeof window === "undefined") {
    return defaultWidths;
  }

  const rawValue = window.localStorage.getItem(storageKey);
  if (!rawValue) {
    return defaultWidths;
  }

  try {
    const parsedValue = JSON.parse(rawValue) as Record<string, unknown>;

    return Object.entries(defaultWidths).reduce<ColumnWidthMap>(
      (currentMap, [columnKey, defaultWidth]) => {
        const storedWidth = parsedValue[columnKey];
        currentMap[columnKey] =
          typeof storedWidth === "number" && Number.isFinite(storedWidth)
            ? storedWidth
            : defaultWidth;

        return currentMap;
      },
      {},
    );
  } catch {
    return defaultWidths;
  }
}

export function useResizableTableColumns(
  storageKey: string,
  columns: ResizableColumnDefinition[],
) {
  const defaultWidths = useMemo(
    () =>
      columns.reduce<ColumnWidthMap>((currentMap, column) => {
        currentMap[column.key] = column.width;

        return currentMap;
      }, {}),
    [columns],
  );
  const columnConfigMap = useMemo(
    () =>
      columns.reduce<Record<string, ResizableColumnDefinition>>(
        (currentMap, column) => {
          currentMap[column.key] = column;

          return currentMap;
        },
        {},
      ),
    [columns],
  );
  const [columnWidths, setColumnWidths] = useState<ColumnWidthMap>(() =>
    readStoredWidths(storageKey, defaultWidths),
  );

  useEffect(() => {
    setColumnWidths((currentWidths) => {
      const nextWidths = Object.entries(defaultWidths).reduce<ColumnWidthMap>(
        (currentMap, [columnKey, defaultWidth]) => {
          currentMap[columnKey] = currentWidths[columnKey] ?? defaultWidth;

          return currentMap;
        },
        {},
      );

      const currentKeys = Object.keys(currentWidths);
      const nextKeys = Object.keys(nextWidths);
      const hasSameShape =
        currentKeys.length === nextKeys.length &&
        nextKeys.every((key) => currentWidths[key] === nextWidths[key]);

      return hasSameShape ? currentWidths : nextWidths;
    });
  }, [defaultWidths]);

  useEffect(() => {
    if (typeof window === "undefined") {
      return;
    }

    window.localStorage.setItem(storageKey, JSON.stringify(columnWidths));
  }, [columnWidths, storageKey]);

  const getColumnWidth = useCallback(
    (columnKey: string) => columnWidths[columnKey] ?? defaultWidths[columnKey],
    [columnWidths, defaultWidths],
  );

  const createResizeHandler = useCallback(
    (columnKey: string) => (event: MouseEvent<HTMLSpanElement>) => {
      event.preventDefault();
      event.stopPropagation();

      const columnConfig = columnConfigMap[columnKey];
      if (!columnConfig) {
        return;
      }

      const startX = event.clientX;
      const startWidth = getColumnWidth(columnKey);
      const minWidth = columnConfig.minWidth ?? defaultMinColumnWidth;
      const maxWidth = columnConfig.maxWidth;
      const previousCursor = document.body.style.cursor;
      const previousUserSelect = document.body.style.userSelect;

      document.body.style.cursor = "col-resize";
      document.body.style.userSelect = "none";

      const handleMouseMove = (moveEvent: globalThis.MouseEvent) => {
        const deltaX = moveEvent.clientX - startX;
        const nextWidth = startWidth + deltaX;
        const clampedWidth =
          maxWidth === undefined
            ? Math.max(minWidth, nextWidth)
            : Math.min(maxWidth, Math.max(minWidth, nextWidth));

        setColumnWidths((currentWidths) => ({
          ...currentWidths,
          [columnKey]: clampedWidth,
        }));
      };

      const handleMouseUp = () => {
        document.body.style.cursor = previousCursor;
        document.body.style.userSelect = previousUserSelect;
        window.removeEventListener("mousemove", handleMouseMove);
        window.removeEventListener("mouseup", handleMouseUp);
      };

      window.addEventListener("mousemove", handleMouseMove);
      window.addEventListener("mouseup", handleMouseUp);
    },
    [columnConfigMap, getColumnWidth],
  );

  return {
    createResizeHandler,
    getColumnWidth,
  };
}

export function ResizableTableHeadCell({
  align,
  children,
  onResizeStart,
  sx,
  width,
}: ResizableTableHeadCellProps) {
  return (
    <TableCell
      align={align}
      sx={{
        ...sx,
        maxWidth: width,
        minWidth: width,
        position: "relative",
        width,
      }}
    >
      {children}
      <Box
        component="span"
        onMouseDown={onResizeStart}
        sx={{
          bgcolor: "transparent",
          cursor: "col-resize",
          height: "100%",
          position: "absolute",
          right: 0,
          top: 0,
          transition: "background-color 0.2s ease",
          userSelect: "none",
          width: 10,
          "&:hover": {
            bgcolor: "rgba(37, 99, 235, 0.18)",
          },
        }}
      />
    </TableCell>
  );
}
