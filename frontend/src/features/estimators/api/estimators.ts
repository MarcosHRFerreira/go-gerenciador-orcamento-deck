import { api } from "../../../lib/axios/api";
import type {
  CreateEstimatorPayload,
  CreateEstimatorResponse,
  EstimatorApiItem,
  EstimatorItem,
  UpdateEstimatorPayload,
} from "../types/estimator";

type CreateEstimatorApiPayload = {
  code: string;
  name: string;
  email: string;
  phone: string;
  notes: string;
  user_id: number | null;
};

type UpdateEstimatorApiPayload = CreateEstimatorApiPayload & {
  active: boolean;
};

type CreateEstimatorApiResponse = {
  id: number;
};

type NextEstimatorCodeApiResponse = {
  code: string;
};

function mapEstimatorItem(item: EstimatorApiItem): EstimatorItem {
  return {
    id: item.id,
    code: item.code,
    name: item.name,
    email: item.email,
    phone: item.phone,
    active: item.active,
    notes: item.notes,
    userId: item.user_id ?? null,
    userName: item.user_name ?? null,
    createdAt: item.created_at,
    updatedAt: item.updated_at,
  };
}

function mapCreateEstimatorPayload(
  payload: CreateEstimatorPayload,
): CreateEstimatorApiPayload {
  return {
    code: payload.code,
    name: payload.name,
    email: payload.email,
    phone: payload.phone,
    notes: payload.notes,
    user_id: payload.userId,
  };
}

function mapUpdateEstimatorPayload(
  payload: UpdateEstimatorPayload,
): UpdateEstimatorApiPayload {
  return {
    ...mapCreateEstimatorPayload(payload),
    active: payload.active,
  };
}

export async function listEstimatorsRequest(): Promise<EstimatorItem[]> {
  const response = await api.get<EstimatorApiItem[]>("/estimators");

  return response.data.map(mapEstimatorItem);
}

export async function getNextEstimatorCodeRequest(): Promise<string> {
  const response = await api.get<NextEstimatorCodeApiResponse>(
    "/estimators/next-code",
  );

  return response.data.code;
}

export async function createEstimatorRequest(
  payload: CreateEstimatorPayload,
): Promise<CreateEstimatorResponse> {
  const response = await api.post<CreateEstimatorApiResponse>(
    "/estimators",
    mapCreateEstimatorPayload(payload),
  );

  return {
    id: response.data.id,
  };
}

export async function updateEstimatorRequest(
  estimatorId: number,
  payload: UpdateEstimatorPayload,
): Promise<void> {
  await api.put(`/estimators/${estimatorId}`, mapUpdateEstimatorPayload(payload));
}

export async function deleteEstimatorRequest(
  estimatorId: number,
): Promise<void> {
  await api.delete(`/estimators/${estimatorId}`);
}
