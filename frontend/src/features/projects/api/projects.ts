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

function reportProjectsDebugEvent(
  hypothesisId: string,
  location: string,
  message: string,
  data: Record<string, unknown>,
) {
  // #region debug-point D:projects-api-report
  fetch("http://127.0.0.1:7777/event", {
    body: JSON.stringify({
      data,
      hypothesisId,
      location,
      msg: message,
      runId: "pre-fix",
      sessionId: "projects-user-load",
      ts: Date.now(),
    }),
    headers: {
      "Content-Type": "application/json",
    },
    method: "POST",
  }).catch(() => undefined);
  // #endregion
}

export async function listProjectsRequest(): Promise<ProjectItem[]> {
  // #region debug-point D:projects-api-entry
  reportProjectsDebugEvent(
    "D",
    "projects/api.ts:listProjectsRequest:entry",
    "[DEBUG] frontend requesting /projects",
    {
      baseURL: api.defaults.baseURL ?? null,
    },
  );
  // #endregion
  try {
    const response = await api.get<ProjectApiItem[]>("/projects");

    // #region debug-point E:projects-api-success
    reportProjectsDebugEvent(
      "E",
      "projects/api.ts:listProjectsRequest:success",
      "[DEBUG] frontend received /projects success",
      {
        itemCount: response.data.length,
        status: response.status,
      },
    );
    // #endregion

    return response.data.map(mapProjectItem);
  } catch (error) {
    const typedError = error as {
      message?: string;
      response?: { data?: unknown; status?: number };
    };

    // #region debug-point E:projects-api-error
    reportProjectsDebugEvent(
      "E",
      "projects/api.ts:listProjectsRequest:error",
      "[DEBUG] frontend received /projects error",
      {
        errorMessage: typedError.message ?? null,
        responseData: typedError.response?.data ?? null,
        status: typedError.response?.status ?? null,
      },
    );
    // #endregion
    throw error;
  }
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
  const response = await api.get<NextProjectCodeApiResponse>(
    "/projects/next-code",
  );

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
  try {
    const response =
      await api.get<NamedProjectCatalogApiItem[]>("/project-types");

    reportProjectsDebugEvent(
      "C",
      "projects/api.ts:listProjectTypesRequest:success",
      "[DEBUG] frontend received /project-types success",
      {
        itemCount: response.data.length,
        status: response.status,
      },
    );

    return response.data.map(mapProjectTypeCatalogItem);
  } catch (error) {
    const typedError = error as {
      message?: string;
      response?: { data?: unknown; status?: number };
    };

    reportProjectsDebugEvent(
      "C",
      "projects/api.ts:listProjectTypesRequest:error",
      "[DEBUG] frontend received /project-types error",
      {
        errorMessage: typedError.message ?? null,
        responseData: typedError.response?.data ?? null,
        status: typedError.response?.status ?? null,
      },
    );
    throw error;
  }
}
