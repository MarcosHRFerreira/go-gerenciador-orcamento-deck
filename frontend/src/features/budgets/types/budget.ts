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
  yearBudget: string;
  statusId: string;
  installerId: string;
  salespersonId: string;
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
