import { Box, Typography } from "@mui/material";
import { alpha } from "@mui/material/styles";
import type { SxProps, Theme } from "@mui/material/styles";
import type { PropsWithChildren } from "react";

export const compactFilterFieldSx = {
  width: "100%",
  "& .MuiFormHelperText-root": {
    marginLeft: 0,
    marginTop: 0.8,
  },
  "& .MuiOutlinedInput-root": {
    backgroundColor: "rgba(255, 255, 255, 0.74)",
    borderRadius: 3,
    boxShadow: "0 10px 22px rgba(30, 58, 138, 0.06)",
    transition:
      "box-shadow 0.2s ease, transform 0.2s ease, border-color 0.2s ease",
    "& fieldset": {
      borderColor: "rgba(30, 58, 138, 0.16)",
    },
    "&:hover": {
      boxShadow: "0 14px 28px rgba(30, 58, 138, 0.1)",
      transform: "translateY(-1px)",
    },
    "&:hover fieldset": {
      borderColor: "rgba(30, 58, 138, 0.28)",
    },
    "&.Mui-focused": {
      boxShadow: "0 16px 30px rgba(30, 58, 138, 0.14)",
    },
    "&.Mui-focused fieldset": {
      borderColor: (theme: Theme) =>
        theme.palette.mode === "dark"
          ? theme.palette.primary.light
          : theme.palette.primary.dark,
      borderWidth: "1px",
    },
  },
} as const;

export const filterFieldContainerSx = {
  display: "grid",
  gap: 0.75,
  minWidth: 0,
} as const;

export const filterFieldLabelSx = {
  color: (theme: Theme) =>
    theme.palette.mode === "dark"
      ? theme.palette.primary.light
      : theme.palette.primary.dark,
  fontSize: "0.78rem",
  fontWeight: 800,
  letterSpacing: "0.04em",
  lineHeight: 1.2,
  textTransform: "uppercase",
} as const;

export const filterGroupSx: SxProps<Theme> = {
  background: (theme) =>
    theme.palette.mode === "dark"
      ? `radial-gradient(circle at top right, ${alpha(theme.palette.primary.light, 0.14)} 0%, transparent 26%), linear-gradient(135deg, ${alpha(theme.palette.primary.light, 0.12)} 0%, ${alpha(theme.palette.info.light, 0.05)} 100%)`
      : `radial-gradient(circle at top right, ${alpha(theme.palette.info.main, 0.11)} 0%, transparent 26%), linear-gradient(135deg, ${alpha(theme.palette.primary.main, 0.07)} 0%, ${alpha(theme.palette.info.main, 0.035)} 100%)`,
  border: "1px solid",
  borderColor: (theme) => alpha(theme.palette.primary.main, 0.16),
  borderRadius: 4,
  boxShadow: (theme) =>
    `0 16px 32px ${alpha(theme.palette.primary.main, 0.08)}`,
  display: "grid",
  gap: 2,
  minWidth: 0,
  overflow: "hidden",
  p: 2.25,
  position: "relative",
  "&::before": {
    background: (theme) =>
      `linear-gradient(90deg, ${alpha(theme.palette.primary.main, 0.88)} 0%, ${alpha(theme.palette.info.main, 0.34)} 100%)`,
    content: '""',
    height: 3,
    left: 0,
    position: "absolute",
    right: 0,
    top: 0,
  },
};

export const filterGroupTitleSx = {
  color: (theme: Theme) =>
    theme.palette.mode === "dark"
      ? theme.palette.primary.light
      : theme.palette.primary.dark,
  fontWeight: 850,
  letterSpacing: "-0.01em",
} as const;

export const filterSectionCardSx: SxProps<Theme> = {
  background: (theme) =>
    `linear-gradient(135deg, ${alpha(theme.palette.primary.main, 0.07)} 0%, ${alpha(theme.palette.info.main, 0.035)} 100%)`,
  border: "1px solid",
  borderColor: (theme) => alpha(theme.palette.primary.main, 0.16),
  boxShadow: (theme) =>
    `0 12px 24px ${alpha(theme.palette.primary.main, 0.07)}`,
  "& .MuiTypography-h5": {
    color: (theme: Theme) =>
      theme.palette.mode === "dark"
        ? theme.palette.primary.light
        : theme.palette.primary.dark,
    fontWeight: 800,
  },
  "& .MuiTypography-body2": {
    color: "text.primary",
  },
};

type FilterFieldProps = PropsWithChildren<{
  label: string;
}>;

export function FilterField({ children, label }: FilterFieldProps) {
  return (
    <Box sx={filterFieldContainerSx}>
      <Typography sx={filterFieldLabelSx}>{label}</Typography>
      {children}
    </Box>
  );
}
