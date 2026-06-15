export type ProjectApiItem = {
  id: number;
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
