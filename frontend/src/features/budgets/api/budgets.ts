import { api } from "../../../lib/axios/api";
import type {
  BudgetCatalogItem,
  BudgetCatalogsResult,
  BudgetApiItem,
  BudgetImportExecutionResult,
  BudgetImportPreviewOptions,
  BudgetImportPreviewResult,
  BudgetCreatePayload,
  BudgetDetailItem,
  ExecuteBudgetImportPayload,
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
    projectName: item.project_name ?? null,
    salespersonName: item.salesperson_name ?? null,
    contactName: item.contact_name ?? null,
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

type BudgetImportPreviewApiOptions = {
  duplicate_strategy: "ignore" | "update";
  create_missing_catalogs: boolean;
  use_default_not_informed: boolean;
};

type BudgetImportPreviewApiMessage = {
  code: string;
  message: string;
};

type BudgetImportPreviewApiSummary = {
  rows_read: number;
  rows_valid: number;
  rows_with_warning: number;
  rows_with_error: number;
  rows_empty_ignored: number;
  new_budgets: number;
  existing_budgets: number;
};

type BudgetImportCatalogActionsApi = {
  budget_statuses_to_create: number;
  priorities_to_create: number;
  installers_to_create: number;
  projects_to_create: number;
  project_types_to_create: number;
  salespeople_to_create: number;
  contacts_to_create: number;
  loss_reasons_to_create: number;
};

type BudgetImportPreviewRowApi = {
  row_number: number;
  budget_number: string;
  status: string;
  action: string;
  messages: string[];
};

type BudgetImportPreviewApiResponse = {
  preview_id: string;
  file_name: string;
  sheet_name: string;
  header_row: number;
  options: BudgetImportPreviewApiOptions;
  summary: BudgetImportPreviewApiSummary;
  catalog_actions: BudgetImportCatalogActionsApi;
  warnings: BudgetImportPreviewApiMessage[];
  errors: BudgetImportPreviewApiMessage[];
  sample_rows: BudgetImportPreviewRowApi[];
  inconsistency_rows: BudgetImportPreviewRowApi[];
};

type BudgetImportExecutionApiSummary = {
  rows_processed: number;
  budgets_created: number;
  budgets_updated: number;
  budgets_ignored: number;
  rows_failed: number;
  catalogs_created: number;
};

type BudgetImportExecutionApiResponse = {
  import_id: string;
  preview_id: string;
  status: string;
  started_at: string;
  finished_at: string;
  summary: BudgetImportExecutionApiSummary;
  result: {
    message: string;
  };
};

function mapNamedCatalogItem(item: NamedCatalogApiItem): BudgetCatalogItem {
  return {
    id: item.id,
    name: item.name,
  };
}

function mapBudgetImportPreviewOptions(
  options: BudgetImportPreviewApiOptions,
): BudgetImportPreviewOptions {
  return {
    duplicateStrategy: options.duplicate_strategy,
    createMissingCatalogs: options.create_missing_catalogs,
    useDefaultNotInformed: options.use_default_not_informed,
  };
}

function mapBudgetImportPreviewResult(
  response: BudgetImportPreviewApiResponse,
): BudgetImportPreviewResult {
  const sampleRows = response.sample_rows ?? [];
  const inconsistencyRows = response.inconsistency_rows ?? [];

  return {
    previewId: response.preview_id,
    fileName: response.file_name,
    sheetName: response.sheet_name,
    headerRow: response.header_row,
    options: mapBudgetImportPreviewOptions(response.options),
    summary: {
      rowsRead: response.summary.rows_read,
      rowsValid: response.summary.rows_valid,
      rowsWithWarning: response.summary.rows_with_warning,
      rowsWithError: response.summary.rows_with_error,
      rowsEmptyIgnored: response.summary.rows_empty_ignored,
      newBudgets: response.summary.new_budgets,
      existingBudgets: response.summary.existing_budgets,
    },
    catalogActions: {
      budgetStatusesToCreate:
        response.catalog_actions.budget_statuses_to_create,
      prioritiesToCreate: response.catalog_actions.priorities_to_create,
      installersToCreate: response.catalog_actions.installers_to_create,
      projectsToCreate: response.catalog_actions.projects_to_create,
      projectTypesToCreate: response.catalog_actions.project_types_to_create,
      salespeopleToCreate: response.catalog_actions.salespeople_to_create,
      contactsToCreate: response.catalog_actions.contacts_to_create,
      lossReasonsToCreate: response.catalog_actions.loss_reasons_to_create,
    },
    warnings: response.warnings ?? [],
    errors: response.errors ?? [],
    sampleRows: sampleRows.map((item) => ({
      rowNumber: item.row_number,
      budgetNumber: item.budget_number,
      status: item.status,
      action: item.action,
      messages: item.messages,
    })),
    inconsistencyRows: inconsistencyRows.map((item) => ({
      rowNumber: item.row_number,
      budgetNumber: item.budget_number,
      status: item.status,
      action: item.action,
      messages: item.messages,
    })),
  };
}

function mapBudgetImportExecutionResult(
  response: BudgetImportExecutionApiResponse,
): BudgetImportExecutionResult {
  return {
    importId: response.import_id,
    previewId: response.preview_id,
    status: response.status,
    startedAt: response.started_at,
    finishedAt: response.finished_at,
    summary: {
      rowsProcessed: response.summary.rows_processed,
      budgetsCreated: response.summary.budgets_created,
      budgetsUpdated: response.summary.budgets_updated,
      budgetsIgnored: response.summary.budgets_ignored,
      rowsFailed: response.summary.rows_failed,
      catalogsCreated: response.summary.catalogs_created,
    },
    result: response.result,
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

export async function getBudgetListCatalogsRequest(): Promise<BudgetCatalogsResult> {
  const [
    statusesResponse,
    prioritiesResponse,
    installersResponse,
    lossReasonsResponse,
  ] = await Promise.all([
    api.get<NamedCatalogApiItem[]>("/budget-statuses"),
    api.get<NamedCatalogApiItem[]>("/priorities"),
    api.get<NamedCatalogApiItem[]>("/installers"),
    api.get<NamedCatalogApiItem[]>("/loss-reasons"),
  ]);

  return {
    statuses: statusesResponse.data.map(mapNamedCatalogItem),
    priorities: prioritiesResponse.data.map(mapNamedCatalogItem),
    installers: installersResponse.data.map(mapNamedCatalogItem),
    projects: [],
    salespeople: [],
    contacts: [],
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

export async function previewBudgetImportRequest(
  file: File,
  options: BudgetImportPreviewOptions,
): Promise<BudgetImportPreviewResult> {
  const formData = new FormData();
  formData.append("file", file);
  formData.append("duplicate_strategy", options.duplicateStrategy);
  formData.append(
    "create_missing_catalogs",
    String(options.createMissingCatalogs),
  );
  formData.append(
    "use_default_not_informed",
    String(options.useDefaultNotInformed),
  );

  const response = await api.post<BudgetImportPreviewApiResponse>(
    "/budget-imports/preview",
    formData,
  );

  return mapBudgetImportPreviewResult(response.data);
}

export async function executeBudgetImportRequest(
  payload: ExecuteBudgetImportPayload,
): Promise<BudgetImportExecutionResult> {
  const response = await api.post<BudgetImportExecutionApiResponse>(
    "/budget-imports",
    {
      preview_id: payload.previewId,
    },
  );

  return mapBudgetImportExecutionResult(response.data);
}
