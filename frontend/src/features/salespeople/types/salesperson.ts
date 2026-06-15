export type SalespersonApiItem = {
  id: number;
  name: string;
  email: string;
  phone: string;
  active: boolean;
  created_at: string;
  updated_at: string;
};

export type SalespersonItem = {
  id: number;
  name: string;
  email: string;
  phone: string;
  active: boolean;
  createdAt: string;
  updatedAt: string;
};

export type SalespersonListFilters = {
  search: string;
  status: "all" | "active" | "inactive";
};

export type CreateSalespersonPayload = {
  name: string;
  email: string;
  phone: string;
};

export type UpdateSalespersonPayload = {
  name: string;
  email: string;
  phone: string;
  active: boolean;
};
