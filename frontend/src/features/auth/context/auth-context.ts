import { createContext } from "react";
import type { AuthSession, AuthUser, LoginPayload } from "../types/auth";

export type AuthContextValue = {
  isAuthenticated: boolean;
  isLoading: boolean;
  session: AuthSession | null;
  user: AuthUser | null;
  login: (payload: LoginPayload) => Promise<void>;
  logout: () => void;
  refreshCurrentUser: () => Promise<void>;
  replaceSession: (session: AuthSession) => Promise<void>;
};

export const AuthContext = createContext<AuthContextValue | null>(null);
