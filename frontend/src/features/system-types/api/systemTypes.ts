import { api } from "../../../lib/axios/api";
import type {
  CreateSystemTypePayload,
  SystemTypeApiItem,
  SystemTypeItem,
  UpdateSystemTypePayload,
} from "../types/systemType";

function mapSystemTypeItem(item: SystemTypeApiItem): SystemTypeItem {
  return {
    id: item.id,
    code: item.code,
    name: item.name,
    description: item.description,
    createdAt: item.created_at,
    updatedAt: item.updated_at,
  };
}

export async function listSystemTypesRequest(): Promise<SystemTypeItem[]> {
  const response = await api.get<SystemTypeApiItem[]>("/system-types");

  return response.data.map(mapSystemTypeItem);
}

export async function createSystemTypeRequest(
  payload: CreateSystemTypePayload,
): Promise<void> {
  await api.post("/system-types", payload);
}

export async function updateSystemTypeRequest(
  systemTypeId: number,
  payload: UpdateSystemTypePayload,
): Promise<void> {
  await api.put(`/system-types/${systemTypeId}`, payload);
}

export async function deleteSystemTypeRequest(
  systemTypeId: number,
): Promise<void> {
  await api.delete(`/system-types/${systemTypeId}`);
}
