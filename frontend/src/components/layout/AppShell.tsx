import DashboardRoundedIcon from "@mui/icons-material/DashboardRounded";
import DarkModeRoundedIcon from "@mui/icons-material/DarkModeRounded";
import DescriptionRoundedIcon from "@mui/icons-material/DescriptionRounded";
import ApartmentRoundedIcon from "@mui/icons-material/ApartmentRounded";
import ManageAccountsRoundedIcon from "@mui/icons-material/ManageAccountsRounded";
import LightModeRoundedIcon from "@mui/icons-material/LightModeRounded";
import LogoutRoundedIcon from "@mui/icons-material/LogoutRounded";
import MenuRoundedIcon from "@mui/icons-material/MenuRounded";
import PeopleAltRoundedIcon from "@mui/icons-material/PeopleAltRounded";
import UploadFileRoundedIcon from "@mui/icons-material/UploadFileRounded";
import {
  AppBar,
  Avatar,
  Box,
  Divider,
  Drawer,
  IconButton,
  List,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  Toolbar,
  Tooltip,
  Typography,
  useMediaQuery,
} from "@mui/material";
import { alpha, useTheme } from "@mui/material/styles";
import { useEffect, useMemo, useState } from "react";
import {
  Link as RouterLink,
  Outlet,
  useLocation,
  useNavigate,
} from "react-router-dom";
import type { ReactElement } from "react";
import { useAuth } from "../../features/auth/hooks/useAuth";
import { useThemeMode } from "../../app/theme/useThemeMode";
import BrandLogo from "../common/BrandLogo";

const drawerWidth = 280;
const collapsedDrawerWidth = 92;
const desktopSidebarStorageKey = "app-shell-desktop-sidebar-collapsed";
const autoCollapseDelayMs = 15_000;

type NavigationItem = {
  icon: ReactElement;
  label: string;
  requiresAdmin?: boolean;
  to: string;
};

const navigationItems: NavigationItem[] = [
  { icon: <DashboardRoundedIcon />, label: "Dashboard", to: "/dashboard" },
  { icon: <DescriptionRoundedIcon />, label: "Orçamentos", to: "/budgets" },
  {
    icon: <ApartmentRoundedIcon />,
    label: "Obras",
    requiresAdmin: true,
    to: "/projects",
  },
  {
    icon: <UploadFileRoundedIcon />,
    label: "Importação",
    to: "/budgets/import",
  },
  {
    icon: <PeopleAltRoundedIcon />,
    label: "Vendedores",
    requiresAdmin: true,
    to: "/salespeople",
  },
  {
    icon: <ManageAccountsRoundedIcon />,
    label: "Usuarios",
    requiresAdmin: true,
    to: "/users",
  },
];

export function AppShell() {
  const theme = useTheme();
  const { mode, toggleThemeMode } = useThemeMode();
  const isDesktop = useMediaQuery(theme.breakpoints.up("lg"));
  const navigate = useNavigate();
  const location = useLocation();
  const [mobileOpen, setMobileOpen] = useState(false);
  const [autoCollapseArmed, setAutoCollapseArmed] = useState(false);
  const [desktopSidebarCollapsed, setDesktopSidebarCollapsed] = useState(() => {
    if (typeof window === "undefined") {
      return false;
    }

    return window.localStorage.getItem(desktopSidebarStorageKey) === "true";
  });
  const { logout, user } = useAuth();
  const availableNavigationItems = useMemo(
    () =>
      navigationItems.filter(
        (item) => !item.requiresAdmin || user?.role === "admin",
      ),
    [user?.role],
  );
  const currentDrawerWidth = desktopSidebarCollapsed
    ? collapsedDrawerWidth
    : drawerWidth;

  const currentPageLabel = useMemo(() => {
    const activeItem = availableNavigationItems.find(
      (item) =>
        location.pathname === item.to ||
        location.pathname.startsWith(`${item.to}/`),
    );
    return activeItem?.label ?? "Painel";
  }, [availableNavigationItems, location.pathname]);

  const userInitial = user?.name?.charAt(0).toUpperCase() ?? "A";
  const userLabel = user?.name ?? "Usuario";
  const userRoleLabel =
    user?.role === "admin" ? "Perfil administrador" : "Perfil usuario";
  const themeModeLabel = mode === "light" ? "Modo claro" : "Modo escuro";

  const handleLogout = () => {
    logout();
    navigate("/login", { replace: true });
  };

  const handleToggleSidebar = () => {
    if (!isDesktop) {
      setMobileOpen(true);
      return;
    }

    const nextCollapsedState = !desktopSidebarCollapsed;

    setDesktopSidebarCollapsed(nextCollapsedState);
    setAutoCollapseArmed(!nextCollapsedState);
  };

  useEffect(() => {
    window.localStorage.setItem(
      desktopSidebarStorageKey,
      String(desktopSidebarCollapsed),
    );
  }, [desktopSidebarCollapsed]);

  useEffect(() => {
    if (!isDesktop || desktopSidebarCollapsed || !autoCollapseArmed) {
      return;
    }

    const timeoutId = window.setTimeout(() => {
      setDesktopSidebarCollapsed(true);
      setAutoCollapseArmed(false);
    }, autoCollapseDelayMs);

    return () => {
      window.clearTimeout(timeoutId);
    };
  }, [autoCollapseArmed, desktopSidebarCollapsed, isDesktop]);

  const renderDrawerContent = (collapsed: boolean) => (
    <Box
      sx={{
        display: "flex",
        flexDirection: "column",
        height: "100%",
        px: collapsed ? 1.25 : 2,
        py: 3,
        transition: "padding 0.2s ease",
      }}
    >
      <Box
        sx={{
          display: "flex",
          flexDirection: "column",
          gap: 1,
          px: collapsed ? 0.75 : 1.5,
          py: 1,
          transition: "padding 0.2s ease",
        }}
      >
        {collapsed ? (
          <Typography
            color="common.white"
            sx={{ textAlign: "center" }}
            variant="h5"
          >
            GO
          </Typography>
        ) : (
          <Typography color="common.white" variant="h5">
            Gestão de Orçamentos
          </Typography>
        )}
        {!collapsed ? (
          <Typography color="rgba(226, 232, 240, 0.75)" variant="body2">
            Painel administrativo moderno e limpo.
          </Typography>
        ) : null}
      </Box>
      <Divider sx={{ borderColor: "rgba(148, 163, 184, 0.16)", my: 2 }} />
      <List sx={{ flexGrow: 1 }}>
        {availableNavigationItems.map((item) => {
          const isSelected =
            location.pathname === item.to ||
            location.pathname.startsWith(`${item.to}/`);

          return (
            <Tooltip
              arrow
              disableHoverListener={!collapsed}
              key={item.to}
              placement="right"
              title={item.label}
            >
              <ListItemButton
                component={RouterLink}
                onClick={() => setMobileOpen(false)}
                selected={isSelected}
                sx={{
                  color: "#E2E8F0",
                  justifyContent: collapsed ? "center" : "flex-start",
                  px: collapsed ? 1 : 1.5,
                }}
                to={item.to}
              >
                <ListItemIcon
                  sx={{
                    color: "inherit",
                    justifyContent: "center",
                    minWidth: collapsed ? 0 : 40,
                  }}
                >
                  {item.icon}
                </ListItemIcon>
                {!collapsed ? <ListItemText primary={item.label} /> : null}
              </ListItemButton>
            </Tooltip>
          );
        })}
      </List>
      <Divider sx={{ borderColor: "rgba(148, 163, 184, 0.16)", my: 2 }} />
      <Tooltip
        arrow
        disableHoverListener={!collapsed}
        placement="right"
        title="Sair"
      >
        <ListItemButton
          onClick={handleLogout}
          sx={{
            color: "#E2E8F0",
            justifyContent: collapsed ? "center" : "flex-start",
            px: collapsed ? 1 : 1.5,
          }}
        >
          <ListItemIcon
            sx={{
              color: "inherit",
              justifyContent: "center",
              minWidth: collapsed ? 0 : 40,
            }}
          >
            <LogoutRoundedIcon />
          </ListItemIcon>
          {!collapsed ? <ListItemText primary="Sair" /> : null}
        </ListItemButton>
      </Tooltip>
    </Box>
  );

  return (
    <Box sx={{ display: "flex", minHeight: "100vh" }}>
      <AppBar
        color="inherit"
        elevation={0}
        position="fixed"
        sx={{
          borderBottom: "1px solid",
          borderColor: "divider",
          bgcolor: alpha(theme.palette.background.paper, 0.82),
          backdropFilter: "blur(14px)",
          ml: { lg: `${currentDrawerWidth}px` },
          transition: "margin-left 0.2s ease, width 0.2s ease",
          width: { lg: `calc(100% - ${currentDrawerWidth}px)` },
        }}
      >
        <Toolbar sx={{ minHeight: 76 }}>
          <Box
            sx={{
              alignItems: "center",
              display: "flex",
              justifyContent: "space-between",
              width: "100%",
            }}
          >
            <Box sx={{ alignItems: "center", display: "flex", gap: 1.5 }}>
              <IconButton onClick={handleToggleSidebar}>
                <MenuRoundedIcon />
              </IconButton>
              <BrandLogo
                imageSx={{ width: { sm: 88, xs: 72 } }}
                wrapperSx={{
                  border: "1px solid",
                  borderColor: "divider",
                  display: { sm: "inline-flex", xs: "none" },
                  p: 0.5,
                }}
              />
              <Box>
                <Typography variant="h5">{currentPageLabel}</Typography>
                <Typography color="text.secondary" variant="body2">
                  Experiencia administrativa clara, moderna e objetiva.
                </Typography>
              </Box>
            </Box>
            <Box sx={{ alignItems: "center", display: "flex", gap: 1.5 }}>
              <Tooltip title={themeModeLabel}>
                <IconButton
                  aria-label={themeModeLabel}
                  color="inherit"
                  onClick={toggleThemeMode}
                  sx={{
                    bgcolor: alpha(theme.palette.primary.main, 0.08),
                    border: "1px solid",
                    borderColor: "divider",
                    "&:hover": {
                      bgcolor: alpha(theme.palette.primary.main, 0.16),
                    },
                  }}
                >
                  {mode === "light" ? (
                    <DarkModeRoundedIcon />
                  ) : (
                    <LightModeRoundedIcon />
                  )}
                </IconButton>
              </Tooltip>
              <Avatar
                sx={{
                  bgcolor: "primary.light",
                  color: "primary.main",
                  fontWeight: 700,
                }}
              >
                {userInitial}
              </Avatar>
              <Box sx={{ display: { sm: "block", xs: "none" } }}>
                <Typography variant="subtitle1">{userLabel}</Typography>
                <Typography color="text.secondary" variant="body2">
                  {userRoleLabel}
                </Typography>
              </Box>
            </Box>
          </Box>
        </Toolbar>
      </AppBar>

      <Drawer
        onClose={() => setMobileOpen(false)}
        open={mobileOpen}
        sx={{
          display: { lg: "none", xs: "block" },
          "& .MuiDrawer-paper": { width: drawerWidth },
        }}
        variant="temporary"
      >
        {renderDrawerContent(false)}
      </Drawer>

      <Drawer
        open
        sx={{
          boxSizing: "border-box",
          display: { lg: "block", xs: "none" },
          flexShrink: 0,
          whiteSpace: "nowrap",
          width: currentDrawerWidth,
          "& .MuiDrawer-paper": {
            boxSizing: "border-box",
            overflowX: "hidden",
            transition: "width 0.2s ease",
            width: currentDrawerWidth,
          },
        }}
        variant="permanent"
      >
        {renderDrawerContent(desktopSidebarCollapsed)}
      </Drawer>

      <Box
        component="main"
        sx={{
          flexGrow: 1,
          ml: { lg: `${currentDrawerWidth}px` },
          minWidth: 0,
          pb: { md: 4, xs: 2 },
          pl: { lg: 0.5, md: 1.5, xs: 2 },
          pr: { lg: 3, md: 3, xs: 2 },
          pt: { md: 13, xs: 12 },
          transition: "margin-left 0.2s ease",
        }}
      >
        <Outlet />
      </Box>
    </Box>
  );
}
