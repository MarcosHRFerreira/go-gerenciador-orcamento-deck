export type ProjectApiItem = {
  id: number;
  code: string;
  name: string;
  project_type_id?: number | null;
  city: string;
  state: string;
  notes: string;
  created_at: string;
  updated_at: string;
};

export type ProjectItem = {
  id: number;
  code: string;
  name: string;
  projectTypeId: number | null;
  city: string;
  state: string;
  notes: string;
  createdAt: string;
  updatedAt: string;
};

export type ProjectTypeCatalogItem = {
  id: number;
  name: string;
};

export type ProjectPayload = {
  code: string;
  name: string;
  projectTypeId: number | null;
  city: string;
  state: string;
  notes: string;
};
