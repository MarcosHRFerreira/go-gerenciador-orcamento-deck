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

  const resetApplicationSession = useCallback(async () => {
    await queryClient.cancelQueries();
    queryClient.clear();
  }, [queryClient]);

  const refreshCurrentUser = useCallback(async () => {
    await queryClient.invalidateQueries({ queryKey: currentUserQueryKey });
    await queryClient.refetchQueries({
      queryKey: currentUserQueryKey,
      type: "active",
    });
  }, [queryClient]);

  const replaceSession = useCallback(
    async (nextSession: AuthSession) => {
      await resetApplicationSession();
      setStoredSession(nextSession);
      setSession(nextSession);
    },
    [resetApplicationSession],
  );

  const clearAuthState = useCallback(async () => {
    await resetApplicationSession();
    clearStoredSession();
    setSession(null);
  }, [resetApplicationSession]);

  useEffect(() => {
    setUnauthorizedHandler(() => {
      void clearAuthState();
    });
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
      await replaceSession(nextSession);
    },
    [replaceSession],
  );

  const logout = useCallback(() => {
    void clearAuthState();
  }, [clearAuthState]);

  const value = useMemo<AuthContextValue>(
    () => ({
      isAuthenticated: Boolean(session?.token),
      isLoading: Boolean(session?.token) && currentUserQuery.isLoading,
      session,
      user: currentUserQuery.data ?? null,
      login,
      logout,
      refreshCurrentUser,
      replaceSession,
    }),
    [
      currentUserQuery.data,
      currentUserQuery.isLoading,
      login,
      logout,
      refreshCurrentUser,
      replaceSession,
      session,
    ],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
