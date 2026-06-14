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
import {
  getCurrentUserRequest,
  loginRequest,
  logoutRequest,
  refreshSessionRequest,
} from "../api/auth";
import type { AuthSession, LoginPayload } from "../types/auth";
import { AuthContext, type AuthContextValue } from "./auth-context";

const currentUserQueryKey = ["auth", "current-user"];

export function AuthProvider({ children }: PropsWithChildren) {
  const queryClient = useQueryClient();
  const [session, setSession] = useState<AuthSession | null>(() =>
    getStoredSession(),
  );
  const [isBootstrapping, setIsBootstrapping] = useState(true);

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

  useEffect(() => {
    let isMounted = true;

    const bootstrapSession = async () => {
      try {
        const nextSession = await refreshSessionRequest();
        if (!isMounted) {
          return;
        }

        setStoredSession(nextSession);
        setSession(nextSession);
      } catch {
        if (!isMounted) {
          return;
        }

        clearStoredSession();
        setSession(null);
      } finally {
        if (isMounted) {
          setIsBootstrapping(false);
        }
      }
    };

    void bootstrapSession();

    return () => {
      isMounted = false;
    };
  }, []);

  const currentUserQuery = useQuery({
    queryKey: currentUserQueryKey,
    queryFn: getCurrentUserRequest,
    enabled: Boolean(session?.token) && !isBootstrapping,
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
    void (async () => {
      try {
        await logoutRequest();
      } finally {
        await clearAuthState();
      }
    })();
  }, [clearAuthState]);

  const value = useMemo<AuthContextValue>(
    () => ({
      isAuthenticated: Boolean(session?.token),
      isLoading:
        isBootstrapping ||
        (Boolean(session?.token) && currentUserQuery.isLoading),
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
      isBootstrapping,
      login,
      logout,
      refreshCurrentUser,
      replaceSession,
      session,
    ],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
