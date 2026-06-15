import { api } from "../../../lib/axios/api";
import type {
  CreateSalespersonPayload,
  SalespersonApiItem,
  SalespersonItem,
  UpdateSalespersonPayload,
} from "../types/salesperson";

function mapSalespersonItem(item: SalespersonApiItem): SalespersonItem {
  return {
    id: item.id,
    name: item.name,
    email: item.email,
    phone: item.phone,
    active: item.active,
    createdAt: item.created_at,
    updatedAt: item.updated_at,
  };
}

export async function listSalespeopleRequest(): Promise<SalespersonItem[]> {
  const response = await api.get<SalespersonApiItem[]>("/salespeople");

  return response.data.map(mapSalespersonItem);
}

export async function createSalespersonRequest(
  payload: CreateSalespersonPayload,
): Promise<void> {
  await api.post("/salespeople", payload);
}

export async function updateSalespersonRequest(
  salespersonId: number,
  payload: UpdateSalespersonPayload,
): Promise<void> {
  await api.put(`/salespeople/${salespersonId}`, payload);
}

export async function deleteSalespersonRequest(
  salespersonId: number,
): Promise<void> {
  await api.delete(`/salespeople/${salespersonId}`);
}
