export type AuthSession = {
  token: string;
  refreshToken: string;
};

export type LoginPayload = {
  email: string;
  password: string;
};

export type LoginResponse = {
  token: string;
  refresh_token: string;
};

export type RefreshTokenPayload = {
  refresh_token: string;
};

export type RefreshTokenResponse = {
  token: string;
  refresh_token: string;
};

export type AuthUser = {
  id: number;
  name: string;
  email: string;
  username: string;
  role: 'admin' | 'user';
  active: boolean;
  created_at: string;
  updated_at: string;
};
