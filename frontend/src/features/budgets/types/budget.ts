﻿﻿export type BudgetSortBy =
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
  priorityId: string;
  installerId: string;
  systemTypeId: string;
  projectCode: string;
  projectName: string;
  salespersonId: string;
  estimatorId: string;
  sentAtFrom: string;
  sentAtTo: string;
  grossValueMin: string;
  grossValueMax: string;
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
  delivery_date?: string | null;
  gross_value: number;
  commission_value: number;
  area_m2: number;
  status_id: number;
  priority_id?: number | null;
  installer_id?: number | null;
  product_line_id?: number | null;
  system_type_id?: number | null;
  project_id?: number | null;
  salesperson_id?: number | null;
  estimator_id?: number | null;
  contact_id?: number | null;
  loss_reason_id?: number | null;
  construction_company: string;
  competitor_name: string;
  competitor_price?: number | null;
  projetista_name: string;
  source_company: string;
  status_name?: string | null;
  priority_name?: string | null;
  installer_name?: string | null;
  product_line_code?: string | null;
  product_line_name?: string | null;
  system_type_code?: string | null;
  system_type_name?: string | null;
  project_code?: string | null;
  project_name?: string | null;
  salesperson_name?: string | null;
  estimator_name?: string | null;
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
  deliveryDate: string | null;
  grossValue: number;
  commissionValue: number;
  areaM2: number;
  statusId: number;
  priorityId: number | null;
  installerId: number | null;
  productLineId: number | null;
  systemTypeId: number | null;
  projectId: number | null;
  salespersonId: number | null;
  estimatorId: number | null;
  contactId: number | null;
  lossReasonId: number | null;
  constructionCompany: string;
  projetistaName: string;
  competitorName: string;
  competitorPrice: number | null;
  statusName: string | null;
  sourceCompany: string;
  priorityName: string | null;
  installerName: string | null;
  productLineCode: string | null;
  productLineName: string | null;
  systemTypeCode: string | null;
  systemTypeName: string | null;
  projectCode: string | null;
  projectName: string | null;
  salespersonName: string | null;
  estimatorName: string | null;
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

export type BudgetGrossValueRangeApiResponse = {
  min: number;
  max: number;
};

export type BudgetGrossValueRangeResult = {
  min: number;
  max: number;
};

export type BudgetDeliveryStatus =
  | "overdue"
  | "due_today"
  | "due_in_1_day"
  | "due_in_2_days"
  | "future"
  | "missing_delivery_date";

export type BudgetDeliveryMonitorFilters = {
  budgetNumber: string;
  projectName: string;
  salespersonId: string;
  statusId: string;
  deliveryDateFrom: string;
  deliveryDateTo: string;
  deliveryStatus: "" | BudgetDeliveryStatus;
  missingDeliveryDate: boolean;
  page: number;
  pageSize: number;
};

export type BudgetDeliveryMonitorApiItem = {
  id: number;
  budget_number: string;
  project_id?: number | null;
  project_code?: string | null;
  project_name?: string | null;
  construction_company: string;
  salesperson_id?: number | null;
  salesperson_name?: string | null;
  status_id: number;
  status_name?: string | null;
  delivery_date?: string | null;
  days_until_delivery?: number | null;
  delivery_status: BudgetDeliveryStatus;
  delivery_status_label: string;
  updated_at: string;
};

export type BudgetDeliveryMonitorItem = {
  id: number;
  budgetNumber: string;
  projectId: number | null;
  projectCode: string | null;
  projectName: string | null;
  constructionCompany: string;
  salespersonId: number | null;
  salespersonName: string | null;
  statusId: number;
  statusName: string | null;
  deliveryDate: string | null;
  daysUntilDelivery: number | null;
  deliveryStatus: BudgetDeliveryStatus;
  deliveryStatusLabel: string;
  updatedAt: string;
};

export type BudgetDeliveryMonitorSummary = {
  total: number;
  overdueCount: number;
  dueTodayCount: number;
  dueInUpTo2DaysCount: number;
  missingDeliveryCount: number;
  futureCount: number;
};

export type BudgetDeliveryMonitorApiResponse = {
  items: BudgetDeliveryMonitorApiItem[];
  summary: {
    total: number;
    overdue_count: number;
    due_today_count: number;
    due_in_up_to_2_days_count: number;
    missing_delivery_count: number;
    future_count: number;
  };
  page: number;
  page_size: number;
  total: number;
};

export type BudgetDeliveryMonitorResult = {
  items: BudgetDeliveryMonitorItem[];
  summary: BudgetDeliveryMonitorSummary;
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
  productLines: BudgetCatalogItem[];
  systemTypes: BudgetCatalogItem[];
  projects: BudgetCatalogItem[];
  salespeople: BudgetCatalogItem[];
  estimators: BudgetCatalogItem[];
  contacts: BudgetCatalogItem[];
  lossReasons: BudgetCatalogItem[];
};

export type BudgetCreatePayload = {
  budgetNumber: string;
  yearBudget: number;
  revision: number;
  sentAt: string;
  deliveryDate: string | null;
  grossValue: number;
  commissionValue: number;
  areaM2: number;
  statusId: number;
  priorityId: number | null;
  installerId: number | null;
  productLineId: number | null;
  systemTypeId: number | null;
  projectId: number | null;
  salespersonId: number | null;
  estimatorId: number | null;
  contactId: number | null;
  lossReasonId: number | null;
  constructionCompany: string;
  competitorName: string;
  competitorPrice: number | null;
  projetistaName: string;
  specificationDetails: string;
  currentFollowUp: string;
};

export type BudgetElectWinnerPayload = {
  notes: string;
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
  productLinesToCreate: number;
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
