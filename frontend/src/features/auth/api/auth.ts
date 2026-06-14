import { api } from "../../../lib/axios/api";
import type {
  AuthUser,
  ChangePasswordPayload,
  ChangePasswordResponse,
  LoginPayload,
  LoginResponse,
} from "../types/auth";

export async function loginRequest(payload: LoginPayload) {
  const response = await api.post<LoginResponse>("/auth/login", payload);

  return {
    token: response.data.token,
    refreshToken: response.data.refresh_token,
  };
}

export async function getCurrentUserRequest() {
  const response = await api.get<AuthUser>("/users/me");
  return response.data;
}

export async function changePasswordRequest(payload: ChangePasswordPayload) {
  const response = await api.patch<ChangePasswordResponse>(
    "/auth/change-password",
    payload,
  );

  return {
    token: response.data.token,
    refreshToken: response.data.refresh_token,
  };
}
