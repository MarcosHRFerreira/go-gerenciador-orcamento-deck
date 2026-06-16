import { api } from "../../../lib/axios/api";
import type {
  CreateUserPayload,
  CreateUserResponse,
  ResetUserPasswordPayload,
  UpdateUserPayload,
  UpdateUserActivePayload,
  UpdateUserRolePayload,
  UserApiItem,
  UserItem,
} from "../types/user";

type CreateUserApiPayload = {
  name: string;
  email: string;
  username: string;
  password: string;
  password_confirm: string;
  role: "admin" | "user";
};

type CreateUserApiResponse = {
  id: number;
};

type UpdateUserApiPayload = {
  name: string;
  email: string;
  username: string;
  role: "admin" | "user";
};

function mapUserItem(item: UserApiItem): UserItem {
  return {
    id: item.id,
    name: item.name,
    email: item.email,
    username: item.username,
    role: item.role,
    active: item.active,
    mustChangePassword: item.must_change_password,
    createdAt: item.created_at,
    updatedAt: item.updated_at,
  };
}

function mapCreateUserPayload(
  payload: CreateUserPayload,
): CreateUserApiPayload {
  return {
    name: payload.name,
    email: payload.email,
    username: payload.username,
    password: payload.password,
    password_confirm: payload.passwordConfirm,
    role: payload.role,
  };
}

function mapUpdateUserPayload(
  payload: UpdateUserPayload,
): UpdateUserApiPayload {
  return {
    email: payload.email,
    name: payload.name,
    role: payload.role,
    username: payload.username,
  };
}

export async function listUsersRequest(): Promise<UserItem[]> {
  const response = await api.get<UserApiItem[]>("/users");

  return response.data.map(mapUserItem);
}

export async function createUserRequest(
  payload: CreateUserPayload,
): Promise<CreateUserResponse> {
  const response = await api.post<CreateUserApiResponse>(
    "/users",
    mapCreateUserPayload(payload),
  );

  return {
    id: response.data.id,
  };
}

export async function updateUserRequest(
  userId: number,
  payload: UpdateUserPayload,
): Promise<void> {
  await api.put(`/users/${userId}`, mapUpdateUserPayload(payload));
}

export async function updateUserRoleRequest(
  userId: number,
  payload: UpdateUserRolePayload,
): Promise<void> {
  await api.patch(`/users/${userId}/role`, payload);
}

export async function updateUserActiveRequest(
  userId: number,
  payload: UpdateUserActivePayload,
): Promise<void> {
  await api.patch(`/users/${userId}/active`, payload);
}

export async function resetUserPasswordRequest(
  userId: number,
  payload: ResetUserPasswordPayload,
): Promise<void> {
  await api.patch(`/users/${userId}/reset-password`, {
    password: payload.password,
    password_confirm: payload.passwordConfirm,
  });
}
