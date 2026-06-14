export type UserRole = "admin" | "user";

export type UserApiItem = {
  id: number;
  name: string;
  email: string;
  username: string;
  role: UserRole;
  active: boolean;
  must_change_password: boolean;
  created_at: string;
  updated_at: string;
};

export type UserItem = {
  id: number;
  name: string;
  email: string;
  username: string;
  role: UserRole;
  active: boolean;
  mustChangePassword: boolean;
  createdAt: string;
  updatedAt: string;
};

export type UserListFilters = {
  search: string;
  role: "all" | UserRole;
  status: "all" | "active" | "inactive";
};

export type CreateUserPayload = {
  name: string;
  email: string;
  username: string;
  password: string;
  passwordConfirm: string;
  role: UserRole;
};

export type CreateUserResponse = {
  id: number;
};

export type UpdateUserRolePayload = {
  role: UserRole;
};

export type UpdateUserActivePayload = {
  active: boolean;
};

export type ResetUserPasswordPayload = {
  password: string;
  passwordConfirm: string;
};
