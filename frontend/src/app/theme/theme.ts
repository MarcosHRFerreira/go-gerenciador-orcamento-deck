import { alpha, createTheme } from "@mui/material/styles";
import type { PaletteMode } from "@mui/material";

const primaryMain = "#2563EB";
const lightSidebarBackground = "#0F172A";
const darkSidebarBackground = "#020617";

export function createAppTheme(mode: PaletteMode) {
  const isDarkMode = mode === "dark";
  const sidebarBackground = isDarkMode
    ? darkSidebarBackground
    : lightSidebarBackground;
  const backgroundDefault = isDarkMode ? "#0B1120" : "#F5F7FB";
  const backgroundPaper = isDarkMode ? "#111827" : "#FFFFFF";
  const dividerColor = isDarkMode ? "#243041" : "#E5E7EB";
  const textPrimary = isDarkMode ? "#F8FAFC" : "#111827";
  const textSecondary = isDarkMode ? "#94A3B8" : "#6B7280";
  const inputBackground = isDarkMode ? "#0F172A" : "#FFFFFF";
  const inputBorder = isDarkMode ? "#334155" : "#D7DCE5";
  const inputHoverBorder = isDarkMode ? "#475569" : "#B7C0CF";
  const paperShadow = isDarkMode
    ? "0 12px 34px rgba(2, 6, 23, 0.35)"
    : "0 8px 30px rgba(15, 23, 42, 0.04)";

  return createTheme({
    palette: {
      mode,
      primary: {
        main: primaryMain,
        dark: "#1D4ED8",
        light: "#DBEAFE",
      },
      success: {
        main: "#16A34A",
      },
      warning: {
        main: "#D97706",
      },
      error: {
        main: "#DC2626",
      },
      info: {
        main: "#0284C7",
      },
      text: {
        primary: textPrimary,
        secondary: textSecondary,
      },
      background: {
        default: backgroundDefault,
        paper: backgroundPaper,
      },
      divider: dividerColor,
    },
    shape: {
      borderRadius: 14,
    },
    typography: {
      fontFamily: 'Inter, system-ui, "Segoe UI", Arial, sans-serif',
      h3: {
        fontSize: "1.75rem",
        fontWeight: 700,
        lineHeight: 1.2,
      },
      h4: {
        fontSize: "1.25rem",
        fontWeight: 600,
      },
      h5: {
        fontSize: "1rem",
        fontWeight: 600,
      },
      subtitle1: {
        fontSize: "1rem",
        fontWeight: 500,
      },
      body1: {
        fontSize: "0.95rem",
        lineHeight: 1.6,
      },
      body2: {
        fontSize: "0.875rem",
        lineHeight: 1.6,
      },
      button: {
        fontWeight: 600,
        textTransform: "none",
      },
    },
    components: {
      MuiCssBaseline: {
        styleOverrides: {
          body: {
            backgroundColor: backgroundDefault,
          },
        },
      },
      MuiPaper: {
        styleOverrides: {
          root: {
            backgroundImage: "none",
            border: `1px solid ${dividerColor}`,
            boxShadow: paperShadow,
          },
        },
      },
      MuiButton: {
        defaultProps: {
          disableElevation: true,
        },
        styleOverrides: {
          root: {
            minHeight: 42,
            borderRadius: 12,
            paddingInline: 18,
          },
        },
      },
      MuiOutlinedInput: {
        styleOverrides: {
          root: {
            backgroundColor: inputBackground,
            borderRadius: 12,
            "& fieldset": {
              borderColor: inputBorder,
            },
            "&:hover fieldset": {
              borderColor: inputHoverBorder,
            },
            "&.Mui-focused fieldset": {
              borderColor: primaryMain,
              boxShadow: `0 0 0 4px ${alpha(primaryMain, 0.08)}`,
            },
          },
        },
      },
      MuiDrawer: {
        styleOverrides: {
          paper: {
            backgroundColor: sidebarBackground,
            border: "none",
            boxShadow: "none",
            color: "#E5E7EB",
          },
        },
      },
      MuiListItemButton: {
        styleOverrides: {
          root: {
            borderRadius: 12,
            marginBottom: 4,
            "&.Mui-selected": {
              backgroundColor: alpha("#FFFFFF", 0.12),
            },
            "&.Mui-selected:hover": {
              backgroundColor: alpha("#FFFFFF", 0.16),
            },
          },
        },
      },
      MuiChip: {
        styleOverrides: {
          root: {
            borderRadius: 10,
            fontWeight: 600,
          },
        },
      },
    },
  });
}
