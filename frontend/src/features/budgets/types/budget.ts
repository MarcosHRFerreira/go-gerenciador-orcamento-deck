export type BudgetSortBy =
  | "sent_at"
  | "gross_value"
  | "created_at"
  | "updated_at"
  | "year_budget"
  | "budget_number";

export type BudgetSortOrder = "asc" | "desc";

export type BudgetListFilters = {
  budgetNumber: string;
  sourceCompany: string;
  yearBudget: string;
  statusId: string;
  installerId: string;
  projectName: string;
  salespersonId: string;
  projectId?: string;
  page: number;
  pageSize: number;
  sortBy: BudgetSortBy;
  sortOrder: BudgetSortOrder;
};

export type BudgetApiItem = {
  id: number;
  budget_number: string;
  year_budget: number;
  revision: number;
  sent_at: string;
  gross_value: number;
  commission_value: number;
  area_m2: number;
  status_id: number;
  priority_id?: number | null;
  installer_id?: number | null;
  project_id?: number | null;
  salesperson_id?: number | null;
  contact_id?: number | null;
  loss_reason_id?: number | null;
  competitor_name: string;
  competitor_price?: number | null;
  designer_name: string;
  source_company: string;
  status_name?: string | null;
  priority_name?: string | null;
  installer_name?: string | null;
  project_name?: string | null;
  salesperson_name?: string | null;
  contact_name?: string | null;
  loss_reason_name?: string | null;
  specification_details: string;
  current_follow_up: string;
  created_at: string;
  updated_at: string;
};

export type BudgetListApiResponse = {
  items: BudgetApiItem[];
  page: number;
  page_size: number;
  total: number;
};

export type BudgetListItem = {
  id: number;
  budgetNumber: string;
  yearBudget: number;
  revision: number;
  sentAt: string;
  grossValue: number;
  commissionValue: number;
  areaM2: number;
  statusId: number;
  priorityId: number | null;
  installerId: number | null;
  projectId: number | null;
  salespersonId: number | null;
  contactId: number | null;
  lossReasonId: number | null;
  designerName: string;
  competitorName: string;
  competitorPrice: number | null;
  statusName: string | null;
  sourceCompany: string;
  priorityName: string | null;
  installerName: string | null;
  projectName: string | null;
  salespersonName: string | null;
  contactName: string | null;
  lossReasonName: string | null;
  specificationDetails: string;
  currentFollowUp: string;
  createdAt: string;
  updatedAt: string;
};

export type BudgetDetailItem = BudgetListItem;

export type BudgetListResult = {
  items: BudgetListItem[];
  page: number;
  pageSize: number;
  total: number;
};

export type BudgetStatusHistoryApiItem = {
  id: number;
  budget_id: number;
  from_status_id?: number | null;
  to_status_id: number;
  changed_by_user_id: number;
  notes: string;
  changed_at: string;
  created_at: string;
  updated_at: string;
};

export type BudgetStatusHistoryItem = {
  id: number;
  budgetId: number;
  fromStatusId: number | null;
  toStatusId: number;
  changedByUserId: number;
  notes: string;
  changedAt: string;
  createdAt: string;
  updatedAt: string;
};

export type BudgetCatalogItem = {
  id: number;
  name: string;
};

export type BudgetCatalogsResult = {
  statuses: BudgetCatalogItem[];
  priorities: BudgetCatalogItem[];
  installers: BudgetCatalogItem[];
  projects: BudgetCatalogItem[];
  salespeople: BudgetCatalogItem[];
  contacts: BudgetCatalogItem[];
  lossReasons: BudgetCatalogItem[];
};

export type BudgetCreatePayload = {
  budgetNumber: string;
  yearBudget: number;
  revision: number;
  sentAt: string;
  grossValue: number;
  commissionValue: number;
  areaM2: number;
  statusId: number;
  priorityId: number | null;
  installerId: number | null;
  projectId: number | null;
  salespersonId: number | null;
  contactId: number | null;
  lossReasonId: number | null;
  competitorName: string;
  competitorPrice: number | null;
  designerName: string;
  specificationDetails: string;
  currentFollowUp: string;
};

export type BudgetImportPreviewOptions = {
  duplicateStrategy: "ignore" | "update";
  createMissingCatalogs: boolean;
  useDefaultNotInformed: boolean;
};

export type BudgetImportPreviewMessage = {
  code: string;
  message: string;
};

export type BudgetImportPreviewSummary = {
  rowsRead: number;
  rowsValid: number;
  rowsWithWarning: number;
  rowsWithError: number;
  rowsEmptyIgnored: number;
  newBudgets: number;
  existingBudgets: number;
};

export type BudgetImportCatalogActions = {
  budgetStatusesToCreate: number;
  prioritiesToCreate: number;
  installersToCreate: number;
  projectsToCreate: number;
  projectTypesToCreate: number;
  salespeopleToCreate: number;
  contactsToCreate: number;
  lossReasonsToCreate: number;
};

export type BudgetImportPreviewRow = {
  rowNumber: number;
  budgetNumber: string;
  status: string;
  action: string;
  messages: string[];
};

export type BudgetImportPreviewLayoutInfo = {
  key: string;
  name: string;
  sourceCompany: string;
  description: string;
};

export type BudgetImportPreviewFieldGroup = {
  key: string;
  title: string;
  description: string;
  fields: string[];
};

export type BudgetImportPreviewGovernance = {
  duplicateScope: string;
  duplicatePolicy: string;
  missingValuePolicy: string;
  defaultCatalogs: string[];
  legacyMatchingScope: string;
};

export type BudgetImportPreviewResult = {
  previewId: string;
  fileName: string;
  sheetName: string;
  headerRow: number;
  layout: BudgetImportPreviewLayoutInfo;
  fieldGroups: BudgetImportPreviewFieldGroup[];
  governance: BudgetImportPreviewGovernance;
  options: BudgetImportPreviewOptions;
  summary: BudgetImportPreviewSummary;
  catalogActions: BudgetImportCatalogActions;
  warnings: BudgetImportPreviewMessage[];
  errors: BudgetImportPreviewMessage[];
  sampleRows: BudgetImportPreviewRow[];
  inconsistencyRows: BudgetImportPreviewRow[];
};

export type ExecuteBudgetImportPayload = {
  previewId: string;
};

export type BudgetImportExecutionSummary = {
  rowsExpected: number;
  rowsProcessed: number;
  budgetsCreated: number;
  budgetsUpdated: number;
  budgetsIgnored: number;
  rowsFailed: number;
  catalogsCreated: number;
};

export type BudgetImportExecutionResult = {
  importId: string;
  previewId: string;
  status: string;
  startedAt: string;
  finishedAt: string;
  summary: BudgetImportExecutionSummary;
  result: {
    message: string;
  };
};
