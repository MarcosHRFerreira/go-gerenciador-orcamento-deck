export type AuthSession = {
  token: string;
};

export type LoginPayload = {
  email: string;
  password: string;
};

export type LoginResponse = {
  token: string;
};

export type ChangePasswordPayload = {
  current_password: string;
  new_password: string;
  new_password_confirm: string;
};

export type ChangePasswordResponse = {
  token: string;
};

export type RefreshTokenResponse = {
  token: string;
};

export type AuthUser = {
  id: number;
  name: string;
  email: string;
  username: string;
  role: "admin" | "user";
  user_kind?: "salesperson" | "estimator" | null;
  active: boolean;
  must_change_password: boolean;
  created_at: string;
  updated_at: string;
};
