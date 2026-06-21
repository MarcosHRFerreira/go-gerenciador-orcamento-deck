import DashboardRoundedIcon from "@mui/icons-material/DashboardRounded";
import DarkModeRoundedIcon from "@mui/icons-material/DarkModeRounded";
import DescriptionRoundedIcon from "@mui/icons-material/DescriptionRounded";
import ApartmentRoundedIcon from "@mui/icons-material/ApartmentRounded";
import CampaignRoundedIcon from "@mui/icons-material/CampaignRounded";
import ForumRoundedIcon from "@mui/icons-material/ForumRounded";
import ManageAccountsRoundedIcon from "@mui/icons-material/ManageAccountsRounded";
import EngineeringRoundedIcon from "@mui/icons-material/EngineeringRounded";
import LightModeRoundedIcon from "@mui/icons-material/LightModeRounded";
import LogoutRoundedIcon from "@mui/icons-material/LogoutRounded";
import MenuRoundedIcon from "@mui/icons-material/MenuRounded";
import PeopleAltRoundedIcon from "@mui/icons-material/PeopleAltRounded";
import SettingsInputComponentRoundedIcon from "@mui/icons-material/SettingsInputComponentRounded";
import UploadFileRoundedIcon from "@mui/icons-material/UploadFileRounded";
import {
  AppBar,
  Avatar,
  Badge,
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
import { useEffect, useMemo, useRef, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  Link as RouterLink,
  Outlet,
  useLocation,
  useNavigate,
} from "react-router-dom";
import type { ReactElement } from "react";
import { useAuth } from "../../features/auth/hooks/useAuth";
import {
  communicationPollingIntervals,
  communicationQueryKeys,
  getConversationUnreadCountRequest,
  getNoticeUnreadCountRequest,
} from "../../features/communication/api/communication";
import { useThemeMode } from "../../app/theme/useThemeMode";
import BrandLogo from "../common/BrandLogo";

const drawerWidth = 280;
const collapsedDrawerWidth = 92;
const desktopSidebarStorageKey = "app-shell-desktop-sidebar-collapsed";
const autoCollapseDelayMs = 15_000;
const notificationToneDurationSeconds = 0.12;
const notificationToneFrequencyHz = 880;
const notificationToneGain = 0.02;
const shellSidebarBackground =
  "linear-gradient(180deg, #0F172A 0%, #162456 45%, #1E3A8A 100%)";

type NavigationItem = {
  icon: ReactElement;
  label: string;
  requiresAdmin?: boolean;
  to: string;
};

const navigationItems: NavigationItem[] = [
  {
    icon: <DashboardRoundedIcon />,
    label: "Dashboard",
    requiresAdmin: true,
    to: "/dashboard",
  },
  { icon: <DescriptionRoundedIcon />, label: "Orçamentos", to: "/budgets" },
  {
    icon: <ApartmentRoundedIcon />,
    label: "Obras",
    to: "/projects",
  },
  {
    icon: <UploadFileRoundedIcon />,
    label: "Importação",
    requiresAdmin: true,
    to: "/budgets/import",
  },
  {
    icon: <CampaignRoundedIcon />,
    label: "Comunicação",
    to: "/communication",
  },
  {
    icon: <PeopleAltRoundedIcon />,
    label: "Vendedores",
    requiresAdmin: true,
    to: "/salespeople",
  },
  {
    icon: <EngineeringRoundedIcon />,
    label: "Orçamentistas",
    requiresAdmin: true,
    to: "/estimators",
  },
  {
    icon: <SettingsInputComponentRoundedIcon />,
    label: "Tipos de Sistema",
    requiresAdmin: true,
    to: "/system-types",
  },
  {
    icon: <ManageAccountsRoundedIcon />,
    label: "Usuários",
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
  const audioContextRef = useRef<AudioContext | null>(null);
  const hasUserInteractedRef = useRef(false);
  const previousUnreadConversationsCountRef = useRef<number | null>(null);
  const [desktopSidebarCollapsed, setDesktopSidebarCollapsed] = useState(() => {
    if (typeof window === "undefined") {
      return false;
    }

    return window.localStorage.getItem(desktopSidebarStorageKey) === "true";
  });
  const { logout, user } = useAuth();
  const unreadNoticesQuery = useQuery({
    queryKey: communicationQueryKeys.noticeUnreadCount(),
    queryFn: getNoticeUnreadCountRequest,
    enabled: Boolean(user),
    refetchInterval: communicationPollingIntervals.unreadCountMs,
    refetchIntervalInBackground: true,
    retry: false,
  });
  const unreadConversationsQuery = useQuery({
    queryKey: communicationQueryKeys.conversationUnreadCount(),
    queryFn: getConversationUnreadCountRequest,
    enabled: Boolean(user),
    refetchInterval: communicationPollingIntervals.unreadCountMs,
    refetchIntervalInBackground: true,
    retry: false,
  });
  const unreadNoticesCount = unreadNoticesQuery.data ?? 0;
  const unreadConversationsCount = unreadConversationsQuery.data ?? 0;
  const unreadCommunicationCount =
    unreadNoticesCount + unreadConversationsCount;
  const availableNavigationItems = useMemo(
    () =>
      navigationItems.filter((item) => {
        if (item.requiresAdmin && user?.role !== "admin") {
          return false;
        }

        if (item.to === "/dashboard" && user?.user_kind === "estimator") {
          return false;
        }

        return true;
      }),
    [user?.role, user?.user_kind],
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
  const userLabel = user?.name ?? "Usuário";
  const userRoleLabel =
    user?.role === "admin"
      ? "Perfil administrador"
      : user?.user_kind === "estimator"
        ? "Perfil orçamentista"
        : "Perfil comercial";
  const themeModeLabel = mode === "light" ? "Modo claro" : "Modo escuro";

  const handleLogout = () => {
    logout();
    navigate("/login", { replace: true });
  };

  const handleOpenCommunication = (tab: "notices" | "conversations") => {
    navigate(`/communication?tab=${tab}`);
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

  const ensureAudioContext = () => {
    if (typeof window === "undefined") {
      return null;
    }

    if (audioContextRef.current) {
      return audioContextRef.current;
    }

    const BrowserAudioContext = window.AudioContext;
    if (!BrowserAudioContext) {
      return null;
    }

    audioContextRef.current = new BrowserAudioContext();
    return audioContextRef.current;
  };

  const playMessageNotificationTone = async () => {
    if (typeof window === "undefined" || !hasUserInteractedRef.current) {
      return;
    }

    const audioContext = ensureAudioContext();
    if (!audioContext) {
      return;
    }

    if (audioContext.state === "suspended") {
      await audioContext.resume();
    }

    const oscillator = audioContext.createOscillator();
    const gainNode = audioContext.createGain();
    const startAt = audioContext.currentTime;
    const endAt = startAt + notificationToneDurationSeconds;

    oscillator.type = "sine";
    oscillator.frequency.setValueAtTime(notificationToneFrequencyHz, startAt);
    gainNode.gain.setValueAtTime(0.0001, startAt);
    gainNode.gain.exponentialRampToValueAtTime(
      notificationToneGain,
      startAt + 0.02,
    );
    gainNode.gain.exponentialRampToValueAtTime(0.0001, endAt);
    oscillator.connect(gainNode);
    gainNode.connect(audioContext.destination);
    oscillator.start(startAt);
    oscillator.stop(endAt);
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

  useEffect(() => {
    if (typeof window === "undefined") {
      return;
    }

    const activateAudio = () => {
      hasUserInteractedRef.current = true;
      void ensureAudioContext()?.resume();
    };

    window.addEventListener("pointerdown", activateAudio, { passive: true });
    window.addEventListener("keydown", activateAudio);

    return () => {
      window.removeEventListener("pointerdown", activateAudio);
      window.removeEventListener("keydown", activateAudio);
    };
  }, []);

  useEffect(() => {
    const previousCount = previousUnreadConversationsCountRef.current;
    previousUnreadConversationsCountRef.current = unreadConversationsCount;

    if (
      previousCount === null ||
      unreadConversationsCount <= previousCount ||
      !user
    ) {
      return;
    }

    void playMessageNotificationTone();
  }, [unreadConversationsCount, user]);

  useEffect(() => {
    return () => {
      audioContextRef.current?.close().catch(() => undefined);
      audioContextRef.current = null;
    };
  }, []);

  const renderDrawerContent = (collapsed: boolean) => (
    <Box
      sx={{
        background: shellSidebarBackground,
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
            Painel administrativo em azul escuro, claro e consistente.
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
                  borderRadius: 3,
                  justifyContent: collapsed ? "center" : "flex-start",
                  px: collapsed ? 1 : 1.5,
                  py: 1.1,
                  transition: "all 0.2s ease",
                  ...(isSelected
                    ? {
                        bgcolor: alpha("#FFFFFF", 0.14),
                        boxShadow: `0 10px 22px ${alpha("#020617", 0.22)}`,
                      }
                    : {}),
                  "&:hover": {
                    bgcolor: alpha("#FFFFFF", isSelected ? 0.18 : 0.1),
                  },
                }}
                to={item.to}
              >
                <ListItemIcon
                  sx={{
                    color: isSelected ? "#FFFFFF" : "#BFDBFE",
                    justifyContent: "center",
                    minWidth: collapsed ? 0 : 40,
                  }}
                >
                  {item.to === "/communication" &&
                  unreadCommunicationCount > 0 ? (
                    <Badge
                      badgeContent={unreadCommunicationCount}
                      color="error"
                    >
                      {item.icon}
                    </Badge>
                  ) : (
                    item.icon
                  )}
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
            borderRadius: 3,
            justifyContent: collapsed ? "center" : "flex-start",
            px: collapsed ? 1 : 1.5,
            py: 1.1,
            "&:hover": {
              bgcolor: alpha("#FFFFFF", 0.1),
            },
          }}
        >
          <ListItemIcon
            sx={{
              color: "#BFDBFE",
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
          borderColor: alpha(theme.palette.primary.main, 0.16),
          bgcolor:
            theme.palette.mode === "dark"
              ? alpha(theme.palette.background.paper, 0.88)
              : alpha(theme.palette.common.white, 0.88),
          backdropFilter: "blur(14px)",
          boxShadow: `0 12px 28px ${alpha(theme.palette.primary.main, 0.08)}`,
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
                <Typography
                  sx={{
                    color:
                      theme.palette.mode === "dark"
                        ? theme.palette.primary.light
                        : theme.palette.primary.dark,
                    fontWeight: 800,
                  }}
                  variant="h5"
                >
                  {currentPageLabel}
                </Typography>
                <Typography color="text.secondary" variant="body2">
                  Experiência administrativa clara, moderna e objetiva.
                </Typography>
              </Box>
            </Box>
            <Box sx={{ alignItems: "center", display: "flex", gap: 1.5 }}>
              <Tooltip title="Avisos">
                <Badge
                  badgeContent={unreadNoticesCount}
                  color="error"
                  max={99}
                  overlap="circular"
                >
                  <IconButton
                    aria-label="Abrir avisos"
                    color="inherit"
                    onClick={() => handleOpenCommunication("notices")}
                    sx={{
                      bgcolor:
                        theme.palette.mode === "dark"
                          ? alpha(theme.palette.warning.light, 0.24)
                          : alpha(theme.palette.warning.main, 0.16),
                      border: "1px solid",
                      borderColor:
                        theme.palette.mode === "dark"
                          ? alpha(theme.palette.warning.light, 0.58)
                          : alpha(theme.palette.warning.main, 0.38),
                      boxShadow:
                        theme.palette.mode === "dark"
                          ? `0 0 0 1px ${alpha(theme.palette.warning.light, 0.14)}`
                          : "none",
                      color:
                        theme.palette.mode === "dark"
                          ? theme.palette.warning.light
                          : theme.palette.warning.dark,
                      "&:hover": {
                        bgcolor:
                          theme.palette.mode === "dark"
                            ? alpha(theme.palette.warning.light, 0.34)
                            : alpha(theme.palette.warning.main, 0.26),
                      },
                    }}
                  >
                    <CampaignRoundedIcon />
                  </IconButton>
                </Badge>
              </Tooltip>
              <Tooltip title="Conversas">
                <Badge
                  badgeContent={unreadConversationsCount}
                  color="error"
                  max={99}
                  overlap="circular"
                >
                  <IconButton
                    aria-label="Abrir conversas"
                    color="inherit"
                    onClick={() => handleOpenCommunication("conversations")}
                    sx={{
                      bgcolor:
                        theme.palette.mode === "dark"
                          ? alpha(theme.palette.info.light, 0.24)
                          : alpha(theme.palette.info.main, 0.16),
                      border: "1px solid",
                      borderColor:
                        theme.palette.mode === "dark"
                          ? alpha(theme.palette.info.light, 0.58)
                          : alpha(theme.palette.info.main, 0.38),
                      boxShadow:
                        theme.palette.mode === "dark"
                          ? `0 0 0 1px ${alpha(theme.palette.info.light, 0.14)}`
                          : "none",
                      color:
                        theme.palette.mode === "dark"
                          ? theme.palette.info.light
                          : theme.palette.info.dark,
                      "&:hover": {
                        bgcolor:
                          theme.palette.mode === "dark"
                            ? alpha(theme.palette.info.light, 0.34)
                            : alpha(theme.palette.info.main, 0.26),
                      },
                    }}
                  >
                    <ForumRoundedIcon />
                  </IconButton>
                </Badge>
              </Tooltip>
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
          "& .MuiDrawer-paper": {
            borderRight: "none",
            width: drawerWidth,
          },
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
            background: shellSidebarBackground,
            borderRight: "none",
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
