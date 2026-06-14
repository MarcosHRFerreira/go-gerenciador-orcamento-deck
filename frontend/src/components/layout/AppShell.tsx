import DashboardRoundedIcon from "@mui/icons-material/DashboardRounded";
import DescriptionRoundedIcon from "@mui/icons-material/DescriptionRounded";
import LogoutRoundedIcon from "@mui/icons-material/LogoutRounded";
import MenuRoundedIcon from "@mui/icons-material/MenuRounded";
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
  Typography,
  useMediaQuery,
} from "@mui/material";
import { useTheme } from "@mui/material/styles";
import { useEffect, useMemo, useState } from "react";
import {
  Link as RouterLink,
  Outlet,
  useLocation,
  useNavigate,
} from "react-router-dom";
import { useAuth } from "../../features/auth/hooks/useAuth";

const drawerWidth = 280;
const collapsedDrawerWidth = 92;
const desktopSidebarStorageKey = "app-shell-desktop-sidebar-collapsed";
const autoCollapseDelayMs = 15_000;

const navigationItems = [
  { icon: <DashboardRoundedIcon />, label: "Dashboard", to: "/" },
  { icon: <DescriptionRoundedIcon />, label: "Orçamentos", to: "/budgets" },
];

export function AppShell() {
  const theme = useTheme();
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
  const currentDrawerWidth = desktopSidebarCollapsed
    ? collapsedDrawerWidth
    : drawerWidth;

  const currentPageLabel = useMemo(() => {
    const activeItem = navigationItems.find(
      (item) =>
        location.pathname === item.to ||
        location.pathname.startsWith(`${item.to}/`),
    );
    return activeItem?.label ?? "Painel";
  }, [location.pathname]);

  const userInitial = user?.name?.charAt(0).toUpperCase() ?? "A";
  const userLabel = user?.name ?? "Usuario";
  const userRoleLabel =
    user?.role === "admin" ? "Perfil administrador" : "Perfil usuario";

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
          gap: 0.5,
          px: collapsed ? 0.75 : 1.5,
          py: 1,
          transition: "padding 0.2s ease",
        }}
      >
        <Typography
          color="common.white"
          sx={{ textAlign: collapsed ? "center" : "left" }}
          variant="h5"
        >
          {collapsed ? "GO" : "Gestão de Orçamentos"}
        </Typography>
        {!collapsed ? (
          <Typography color="rgba(226, 232, 240, 0.75)" variant="body2">
            Painel administrativo moderno e limpo.
          </Typography>
        ) : null}
      </Box>
      <Divider sx={{ borderColor: "rgba(148, 163, 184, 0.16)", my: 2 }} />
      <List sx={{ flexGrow: 1 }}>
        {navigationItems.map((item) => {
          const isSelected =
            location.pathname === item.to ||
            location.pathname.startsWith(`${item.to}/`);

          return (
            <ListItemButton
              key={item.to}
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
          );
        })}
      </List>
      <Divider sx={{ borderColor: "rgba(148, 163, 184, 0.16)", my: 2 }} />
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
          bgcolor: "rgba(255,255,255,0.82)",
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
              <Box>
                <Typography variant="h5">{currentPageLabel}</Typography>
                <Typography color="text.secondary" variant="body2">
                  Experiencia administrativa clara, moderna e objetiva.
                </Typography>
              </Box>
            </Box>
            <Box sx={{ alignItems: "center", display: "flex", gap: 1.5 }}>
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
          p: { md: 4, xs: 2 },
          pt: { md: 13, xs: 12 },
          transition: "margin-left 0.2s ease",
        }}
      >
        <Outlet />
      </Box>
    </Box>
  );
}
