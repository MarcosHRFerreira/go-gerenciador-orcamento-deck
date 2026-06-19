export type UserRole = "admin" | "user";
export type UserKind = "salesperson" | "estimator";

export type UserApiItem = {
  id: number;
  name: string;
  email: string;
  username: string;
  role: UserRole;
  user_kind?: UserKind | null;
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
  userKind: UserKind | null;
  active: boolean;
  mustChangePassword: boolean;
  createdAt: string;
  updatedAt: string;
};

export type UserListFilters = {
  search: string;
  role: "all" | UserRole;
  userKind: "all" | UserKind;
  status: "all" | "active" | "inactive";
};

export type CreateUserPayload = {
  name: string;
  email: string;
  username: string;
  password: string;
  passwordConfirm: string;
  role: UserRole;
  userKind?: UserKind;
};

export type CreateUserResponse = {
  id: number;
};

export type UpdateUserPayload = {
  name: string;
  email: string;
  username: string;
  role: UserRole;
  userKind?: UserKind;
};

export type UpdateUserRolePayload = {
  role: UserRole;
  userKind?: UserKind;
};

export type UpdateUserActivePayload = {
  active: boolean;
};

export type ResetUserPasswordPayload = {
  password: string;
  passwordConfirm: string;
};
