import { api } from "../../../lib/axios/api";
import type {
  ProjectApiItem,
  ProjectItem,
  ProjectPayload,
  ProjectTypeCatalogItem,
} from "../types/project";

type NamedProjectCatalogApiItem = {
  id: number;
  name: string;
};

function mapProjectItem(item: ProjectApiItem): ProjectItem {
  return {
    id: item.id,
    code: item.code,
    name: item.name,
    projectTypeId: item.project_type_id ?? null,
    city: item.city,
    state: item.state,
    notes: item.notes,
    createdAt: item.created_at,
    updatedAt: item.updated_at,
  };
}

type ProjectApiPayload = {
  code: string;
  name: string;
  project_type_id: number | null;
  city: string;
  state: string;
  notes: string;
};

type CreateProjectApiResponse = {
  id: number;
};

type NextProjectCodeApiResponse = {
  code: string;
};

function mapProjectTypeCatalogItem(
  item: NamedProjectCatalogApiItem,
): ProjectTypeCatalogItem {
  return {
    id: item.id,
    name: item.name,
  };
}

function mapProjectPayload(payload: ProjectPayload): ProjectApiPayload {
  return {
    code: payload.code.trim(),
    name: payload.name.trim(),
    project_type_id: payload.projectTypeId,
    city: payload.city.trim(),
    state: payload.state.trim(),
    notes: payload.notes.trim(),
  };
}

export async function listProjectsRequest(): Promise<ProjectItem[]> {
  const response = await api.get<ProjectApiItem[]>("/projects");

  return response.data.map(mapProjectItem);
}

export async function getProjectByIdRequest(
  projectId: number,
): Promise<ProjectItem> {
  const response = await api.get<ProjectApiItem>(`/projects/${projectId}`);

  return mapProjectItem(response.data);
}

export async function createProjectRequest(
  payload: ProjectPayload,
): Promise<number> {
  const response = await api.post<CreateProjectApiResponse>(
    "/projects",
    mapProjectPayload(payload),
  );

  return response.data.id;
}

export async function getNextProjectCodeRequest(): Promise<string> {
  const response = await api.get<NextProjectCodeApiResponse>("/projects/next-code");

  return response.data.code;
}


export async function updateProjectRequest(
  projectId: number,
  payload: ProjectPayload,
): Promise<void> {
  await api.put(`/projects/${projectId}`, mapProjectPayload(payload));
}

export async function deleteProjectRequest(projectId: number): Promise<void> {
  await api.delete(`/projects/${projectId}`);
}

export async function listProjectTypesRequest(): Promise<
  ProjectTypeCatalogItem[]
> {
  const response =
    await api.get<NamedProjectCatalogApiItem[]>("/project-types");

  return response.data.map(mapProjectTypeCatalogItem);
}
