import { Navigate, Outlet, useLocation } from "react-router-dom";
import { AuthLoadingScreen } from "./AuthLoadingScreen";
import { useAuth } from "../hooks/useAuth";

export function ProtectedRoute() {
  const location = useLocation();
  const { isAuthenticated, isLoading, user } = useAuth();

  if (!isAuthenticated) {
    return <Navigate replace state={{ from: location }} to="/login" />;
  }

  if (isLoading || !user) {
    return <AuthLoadingScreen />;
  }

  if (user.must_change_password && location.pathname !== "/change-password") {
    return (
      <Navigate replace state={{ from: location }} to="/change-password" />
    );
  }

  return <Outlet />;
}
