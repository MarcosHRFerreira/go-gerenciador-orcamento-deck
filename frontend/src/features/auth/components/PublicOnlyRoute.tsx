import { Navigate, Outlet } from "react-router-dom";
import { AuthLoadingScreen } from "./AuthLoadingScreen";
import { useAuth } from "../hooks/useAuth";

export function PublicOnlyRoute() {
  const { isAuthenticated, isLoading, user } = useAuth();

  if (isAuthenticated && (isLoading || !user)) {
    return <AuthLoadingScreen />;
  }

  if (isAuthenticated) {
    const redirectPath = user?.must_change_password
      ? "/change-password"
      : "/budgets";

    return <Navigate replace to={redirectPath} />;
  }

  return <Outlet />;
}
