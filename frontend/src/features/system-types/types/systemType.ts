export type SystemTypeApiItem = {
  id: number;
  code: string;
  name: string;
  description: string;
  created_at: string;
  updated_at: string;
};

export type SystemTypeItem = {
  id: number;
  code: string;
  name: string;
  description: string;
  createdAt: string;
  updatedAt: string;
};

export type SystemTypeListFilters = {
  search: string;
};

export type CreateSystemTypePayload = {
  code: string;
  name: string;
  description: string;
};

export type UpdateSystemTypePayload = {
  code: string;
  name: string;
  description: string;
};
