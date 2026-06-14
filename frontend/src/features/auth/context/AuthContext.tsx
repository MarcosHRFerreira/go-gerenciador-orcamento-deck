import { useQuery, useQueryClient } from "@tanstack/react-query";
import {
  useCallback,
  useEffect,
  useMemo,
  useState,
  type PropsWithChildren,
} from "react";
import { setUnauthorizedHandler } from "../../../lib/axios/api";
import {
  clearStoredSession,
  getStoredSession,
  setStoredSession,
} from "../../../lib/storage/sessionStorage";
import { getCurrentUserRequest, loginRequest } from "../api/auth";
import type { AuthSession, LoginPayload } from "../types/auth";
import { AuthContext, type AuthContextValue } from "./auth-context";

const currentUserQueryKey = ["auth", "current-user"];

export function AuthProvider({ children }: PropsWithChildren) {
  const queryClient = useQueryClient();
  const [session, setSession] = useState<AuthSession | null>(() =>
    getStoredSession(),
  );

  const clearAuthState = useCallback(() => {
    clearStoredSession();
    setSession(null);
    queryClient.removeQueries({ queryKey: currentUserQueryKey });
  }, [queryClient]);

  useEffect(() => {
    setUnauthorizedHandler(clearAuthState);
  }, [clearAuthState]);

  const currentUserQuery = useQuery({
    queryKey: currentUserQueryKey,
    queryFn: getCurrentUserRequest,
    enabled: Boolean(session?.token),
    retry: false,
  });

  const login = useCallback(
    async (payload: LoginPayload) => {
      const nextSession = await loginRequest(payload);

      setStoredSession(nextSession);
      setSession(nextSession);

      await queryClient.invalidateQueries({ queryKey: currentUserQueryKey });
      await queryClient.refetchQueries({
        queryKey: currentUserQueryKey,
        type: "active",
      });
    },
    [queryClient],
  );

  const logout = useCallback(() => {
    clearAuthState();
  }, [clearAuthState]);

  const value = useMemo<AuthContextValue>(
    () => ({
      isAuthenticated: Boolean(session?.token),
      isLoading: Boolean(session?.token) && currentUserQuery.isLoading,
      session,
      user: currentUserQuery.data ?? null,
      login,
      logout,
    }),
    [currentUserQuery.data, currentUserQuery.isLoading, login, logout, session],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
