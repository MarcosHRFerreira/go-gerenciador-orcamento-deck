export type EstimatorApiItem = {
  id: number;
  code: string;
  name: string;
  email: string;
  phone: string;
  active: boolean;
  notes: string;
  user_id?: number | null;
  user_name?: string | null;
  created_at: string;
  updated_at: string;
};

export type EstimatorItem = {
  id: number;
  code: string;
  name: string;
  email: string;
  phone: string;
  active: boolean;
  notes: string;
  userId: number | null;
  userName: string | null;
  createdAt: string;
  updatedAt: string;
};

export type EstimatorListFilters = {
  search: string;
  status: "all" | "active" | "inactive";
  link: "all" | "linked" | "unlinked";
};

export type CreateEstimatorPayload = {
  code: string;
  name: string;
  email: string;
  phone: string;
  notes: string;
  userId: number | null;
};

export type UpdateEstimatorPayload = {
  code: string;
  name: string;
  email: string;
  phone: string;
  active: boolean;
  notes: string;
  userId: number | null;
};

export type CreateEstimatorResponse = {
  id: number;
};
