import { api } from "../../../lib/axios/api";
import type {
  ProjectApiItem,
  ProjectItem,
  ProjectTypeCatalogItem,
} from "../types/project";

type NamedProjectCatalogApiItem = {
  id: number;
  name: string;
};

function mapProjectItem(item: ProjectApiItem): ProjectItem {
  return {
    id: item.id,
    name: item.name,
    projectTypeId: item.project_type_id ?? null,
    city: item.city,
    state: item.state,
    notes: item.notes,
    createdAt: item.created_at,
    updatedAt: item.updated_at,
  };
}

function mapProjectTypeCatalogItem(
  item: NamedProjectCatalogApiItem,
): ProjectTypeCatalogItem {
  return {
    id: item.id,
    name: item.name,
  };
}

export async function getProjectByIdRequest(projectId: number): Promise<ProjectItem> {
  const response = await api.get<ProjectApiItem>(`/projects/${projectId}`);

  return mapProjectItem(response.data);
}

export async function listProjectTypesRequest(): Promise<ProjectTypeCatalogItem[]> {
  const response = await api.get<NamedProjectCatalogApiItem[]>("/project-types");

  return response.data.map(mapProjectTypeCatalogItem);
}
