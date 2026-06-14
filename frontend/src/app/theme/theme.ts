import { alpha, createTheme } from '@mui/material/styles';

const primaryMain = '#2563EB';
const sidebarBackground = '#0F172A';

export const appTheme = createTheme({
  palette: {
    mode: 'light',
    primary: {
      main: primaryMain,
      dark: '#1D4ED8',
      light: '#DBEAFE',
    },
    success: {
      main: '#16A34A',
    },
    warning: {
      main: '#D97706',
    },
    error: {
      main: '#DC2626',
    },
    info: {
      main: '#0284C7',
    },
    text: {
      primary: '#111827',
      secondary: '#6B7280',
    },
    background: {
      default: '#F5F7FB',
      paper: '#FFFFFF',
    },
    divider: '#E5E7EB',
  },
  shape: {
    borderRadius: 14,
  },
  typography: {
    fontFamily: 'Inter, system-ui, "Segoe UI", Arial, sans-serif',
    h3: {
      fontSize: '1.75rem',
      fontWeight: 700,
      lineHeight: 1.2,
    },
    h4: {
      fontSize: '1.25rem',
      fontWeight: 600,
    },
    h5: {
      fontSize: '1rem',
      fontWeight: 600,
    },
    subtitle1: {
      fontSize: '1rem',
      fontWeight: 500,
    },
    body1: {
      fontSize: '0.95rem',
      lineHeight: 1.6,
    },
    body2: {
      fontSize: '0.875rem',
      lineHeight: 1.6,
    },
    button: {
      fontWeight: 600,
      textTransform: 'none',
    },
  },
  components: {
    MuiCssBaseline: {
      styleOverrides: {
        body: {
          backgroundColor: '#F5F7FB',
        },
      },
    },
    MuiPaper: {
      styleOverrides: {
        root: {
          backgroundImage: 'none',
          border: '1px solid #E5E7EB',
          boxShadow: '0 8px 30px rgba(15, 23, 42, 0.04)',
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
          borderRadius: 12,
          backgroundColor: '#FFFFFF',
          '& fieldset': {
            borderColor: '#D7DCE5',
          },
          '&:hover fieldset': {
            borderColor: '#B7C0CF',
          },
          '&.Mui-focused fieldset': {
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
          color: '#E5E7EB',
          border: 'none',
          boxShadow: 'none',
        },
      },
    },
    MuiListItemButton: {
      styleOverrides: {
        root: {
          borderRadius: 12,
          marginBottom: 4,
          '&.Mui-selected': {
            backgroundColor: alpha('#FFFFFF', 0.12),
          },
          '&.Mui-selected:hover': {
            backgroundColor: alpha('#FFFFFF', 0.16),
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
