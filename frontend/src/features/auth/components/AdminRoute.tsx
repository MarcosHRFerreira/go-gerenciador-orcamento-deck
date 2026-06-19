import { Alert, Box } from "@mui/material";
import { Navigate, Outlet, useLocation } from "react-router-dom";
import { useAuth } from "../hooks/useAuth";
import { AuthLoadingScreen } from "./AuthLoadingScreen";

export function AdminRoute() {
  const location = useLocation();
  const { isLoading, user } = useAuth();

  if (isLoading) {
    return <AuthLoadingScreen />;
  }

  if (!user) {
    return <Navigate replace state={{ from: location }} to="/login" />;
  }

  if (user.role !== "admin") {
    return (
      <Box sx={{ p: { md: 3, xs: 2 } }}>
        <Alert severity="error">
          Você não possui permissão para acessar esta área administrativa.
        </Alert>
      </Box>
    );
  }

  return <Outlet />;
}
