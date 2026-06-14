import { api } from "../../../lib/axios/api";
import type {
  BudgetCatalogItem,
  BudgetCatalogsResult,
  BudgetApiItem,
  BudgetCreatePayload,
  BudgetDetailItem,
  BudgetListApiResponse,
  BudgetListFilters,
  BudgetListResult,
} from "../types/budget";

function mapBudgetListItem(item: BudgetApiItem) {
  return {
    id: item.id,
    budgetNumber: item.budget_number,
    yearBudget: item.year_budget,
    revision: item.revision,
    sentAt: item.sent_at,
    grossValue: item.gross_value,
    commissionValue: item.commission_value,
    areaM2: item.area_m2,
    statusId: item.status_id,
    priorityId: item.priority_id ?? null,
    installerId: item.installer_id ?? null,
    projectId: item.project_id ?? null,
    salespersonId: item.salesperson_id ?? null,
    contactId: item.contact_id ?? null,
    lossReasonId: item.loss_reason_id ?? null,
    designerName: item.designer_name,
    competitorName: item.competitor_name,
    competitorPrice: item.competitor_price ?? null,
    specificationDetails: item.specification_details,
    currentFollowUp: item.current_follow_up,
    createdAt: item.created_at,
    updatedAt: item.updated_at,
  };
}

type NamedCatalogApiItem = {
  id: number;
  name: string;
};

type CreateBudgetApiPayload = {
  budget_number: string;
  year_budget: number;
  revision: number;
  sent_at: string;
  gross_value: number;
  commission_value: number;
  area_m2: number;
  status_id: number;
  priority_id: number | null;
  installer_id: number | null;
  project_id: number | null;
  salesperson_id: number | null;
  contact_id: number | null;
  loss_reason_id: number | null;
  competitor_name: string;
  competitor_price: number | null;
  designer_name: string;
  specification_details: string;
  current_follow_up: string;
};

type CreateBudgetApiResponse = {
  id: number;
};

function mapNamedCatalogItem(item: NamedCatalogApiItem): BudgetCatalogItem {
  return {
    id: item.id,
    name: item.name,
  };
}

function mapCreateBudgetPayload(
  payload: BudgetCreatePayload,
): CreateBudgetApiPayload {
  return {
    area_m2: payload.areaM2,
    budget_number: payload.budgetNumber,
    commission_value: payload.commissionValue,
    competitor_name: payload.competitorName,
    competitor_price: payload.competitorPrice,
    contact_id: payload.contactId,
    current_follow_up: payload.currentFollowUp,
    designer_name: payload.designerName,
    gross_value: payload.grossValue,
    installer_id: payload.installerId,
    loss_reason_id: payload.lossReasonId,
    priority_id: payload.priorityId,
    project_id: payload.projectId,
    revision: payload.revision,
    salesperson_id: payload.salespersonId,
    sent_at: payload.sentAt,
    specification_details: payload.specificationDetails,
    status_id: payload.statusId,
    year_budget: payload.yearBudget,
  };
}

export async function getBudgetListRequest(
  filters: BudgetListFilters,
): Promise<BudgetListResult> {
  const response = await api.get<BudgetListApiResponse>("/budgets", {
    params: {
      budget_number: filters.budgetNumber || undefined,
      year_budget: filters.yearBudget || undefined,
      status_id: filters.statusId || undefined,
      installer_id: filters.installerId || undefined,
      salesperson_id: filters.salespersonId || undefined,
      page: filters.page,
      page_size: filters.pageSize,
      sort_by: filters.sortBy,
      sort_order: filters.sortOrder,
    },
  });

  return {
    items: response.data.items.map(mapBudgetListItem),
    page: response.data.page,
    pageSize: response.data.page_size,
    total: response.data.total,
  };
}

export async function getBudgetCatalogsRequest(): Promise<BudgetCatalogsResult> {
  const [
    statusesResponse,
    prioritiesResponse,
    installersResponse,
    projectsResponse,
    salespeopleResponse,
    contactsResponse,
    lossReasonsResponse,
  ] = await Promise.all([
    api.get<NamedCatalogApiItem[]>("/budget-statuses"),
    api.get<NamedCatalogApiItem[]>("/priorities"),
    api.get<NamedCatalogApiItem[]>("/installers"),
    api.get<NamedCatalogApiItem[]>("/projects"),
    api.get<NamedCatalogApiItem[]>("/salespeople"),
    api.get<NamedCatalogApiItem[]>("/contacts"),
    api.get<NamedCatalogApiItem[]>("/loss-reasons"),
  ]);

  return {
    statuses: statusesResponse.data.map(mapNamedCatalogItem),
    priorities: prioritiesResponse.data.map(mapNamedCatalogItem),
    installers: installersResponse.data.map(mapNamedCatalogItem),
    projects: projectsResponse.data.map(mapNamedCatalogItem),
    salespeople: salespeopleResponse.data.map(mapNamedCatalogItem),
    contacts: contactsResponse.data.map(mapNamedCatalogItem),
    lossReasons: lossReasonsResponse.data.map(mapNamedCatalogItem),
  };
}

export async function getBudgetByIdRequest(
  budgetId: number,
): Promise<BudgetDetailItem> {
  const response = await api.get<BudgetApiItem>(`/budgets/${budgetId}`);

  return mapBudgetListItem(response.data);
}

export async function createBudgetRequest(
  payload: BudgetCreatePayload,
): Promise<number> {
  const response = await api.post<CreateBudgetApiResponse>(
    "/budgets",
    mapCreateBudgetPayload(payload),
  );

  return response.data.id;
}

export async function updateBudgetRequest(
  budgetId: number,
  payload: BudgetCreatePayload,
): Promise<void> {
  await api.put(`/budgets/${budgetId}`, mapCreateBudgetPayload(payload));
}

export async function deleteBudgetRequest(budgetId: number): Promise<void> {
  await api.delete(`/budgets/${budgetId}`);
}
