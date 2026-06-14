import axios, { AxiosError } from "axios";
import type { InternalAxiosRequestConfig } from "axios";
import type {
  AuthSession,
  RefreshTokenResponse,
} from "../../features/auth/types/auth";
import {
  clearStoredSession,
  getStoredSession,
  setStoredSession,
} from "../storage/sessionStorage";

type RetryableRequestConfig = InternalAxiosRequestConfig & {
  _retry?: boolean;
};

const baseURL = import.meta.env.VITE_API_URL?.trim() || "http://localhost:8080";

const authApi = axios.create({ baseURL });

export const api = axios.create({ baseURL });

let refreshSessionPromise: Promise<AuthSession | null> | null = null;
let unauthorizedHandler: (() => void) | null = null;

export function setUnauthorizedHandler(handler: () => void) {
  unauthorizedHandler = handler;
}

async function requestRefreshSession() {
  const currentSession = getStoredSession();

  if (!currentSession) {
    return null;
  }

  try {
    const response = await authApi.post<RefreshTokenResponse>(
      "/auth/refresh",
      { refresh_token: currentSession.refreshToken },
      {
        headers: {
          Authorization: `Bearer ${currentSession.token}`,
        },
      },
    );

    const refreshedSession: AuthSession = {
      token: response.data.token,
      refreshToken: response.data.refresh_token,
    };

    setStoredSession(refreshedSession);
    return refreshedSession;
  } catch {
    clearStoredSession();
    unauthorizedHandler?.();
    return null;
  }
}

api.interceptors.request.use((config) => {
  const nextConfig = { ...config };
  const currentSession = getStoredSession();

  if (currentSession?.token) {
    nextConfig.headers = nextConfig.headers ?? {};
    nextConfig.headers.Authorization = `Bearer ${currentSession.token}`;
  }

  return nextConfig;
});

api.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const originalRequest = error.config as RetryableRequestConfig | undefined;
    const requestUrl = originalRequest?.url ?? "";

    if (
      error.response?.status !== 401 ||
      !originalRequest ||
      originalRequest._retry ||
      requestUrl.includes("/auth/login") ||
      requestUrl.includes("/auth/refresh")
    ) {
      return Promise.reject(error);
    }

    originalRequest._retry = true;

    if (!refreshSessionPromise) {
      refreshSessionPromise = requestRefreshSession().finally(() => {
        refreshSessionPromise = null;
      });
    }

    const refreshedSession = await refreshSessionPromise;

    if (!refreshedSession) {
      return Promise.reject(error);
    }

    originalRequest.headers = originalRequest.headers ?? {};
    originalRequest.headers.Authorization = `Bearer ${refreshedSession.token}`;

    return api(originalRequest);
  },
);
