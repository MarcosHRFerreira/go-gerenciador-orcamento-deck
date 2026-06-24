import type { BudgetDetailItem } from "../types/budget";

export type BudgetFormValues = {
  areaM2: string;
  budgetNumber: string;
  commissionValue: string;
  constructionCompany: string;
  competitorName: string;
  competitorPrice: string;
  contactId: string;
  currentFollowUp: string;
  deliveryDate: string;
  projetistaName: string;
  grossValue: string;
  installerId: string;
  lossReasonId: string;
  priorityId: string;
  productLineId: string;
  systemTypeId: string;
  projectId: string;
  revision: string;
  salespersonId: string;
  estimatorId: string;
  sentAt: string;
  sourceCompany: string;
  specificationDetails: string;
  statusId: string;
  yearBudget: string;
};

function toDateTimeLocalValue(value: string) {
  const date = new Date(value);
  const localDate = new Date(date.getTime() - date.getTimezoneOffset() * 60000);

  return localDate.toISOString().slice(0, 16);
}

export function createDefaultBudgetFormValues(): BudgetFormValues {
  return {
    areaM2: "0",
    budgetNumber: "",
    commissionValue: "0",
    constructionCompany: "",
    competitorName: "",
    competitorPrice: "",
    contactId: "",
    currentFollowUp: "",
    deliveryDate: "",
    projetistaName: "",
    grossValue: "",
    installerId: "",
    lossReasonId: "",
    priorityId: "",
    productLineId: "",
    systemTypeId: "",
    projectId: "",
    revision: "0",
    salespersonId: "",
    estimatorId: "",
    sentAt: toDateTimeLocalValue(new Date().toISOString()),
    sourceCompany: "",
    specificationDetails: "",
    statusId: "",
    yearBudget: String(new Date().getFullYear()),
  };
}

export function mapBudgetDetailToFormValues(
  budget: BudgetDetailItem,
): BudgetFormValues {
  return {
    areaM2: String(budget.areaM2),
    budgetNumber: budget.budgetNumber,
    commissionValue: String(budget.commissionValue),
    constructionCompany: budget.constructionCompany,
    competitorName: budget.competitorName,
    competitorPrice:
      budget.competitorPrice === null ? "" : String(budget.competitorPrice),
    contactId: budget.contactId === null ? "" : String(budget.contactId),
    currentFollowUp: budget.currentFollowUp,
    deliveryDate: budget.deliveryDate ?? "",
    projetistaName: budget.projetistaName,
    grossValue: String(budget.grossValue),
    installerId: budget.installerId === null ? "" : String(budget.installerId),
    lossReasonId:
      budget.lossReasonId === null ? "" : String(budget.lossReasonId),
    priorityId: budget.priorityId === null ? "" : String(budget.priorityId),
    productLineId:
      budget.productLineId === null ? "" : String(budget.productLineId),
    systemTypeId:
      budget.systemTypeId === null ? "" : String(budget.systemTypeId),
    projectId: budget.projectId === null ? "" : String(budget.projectId),
    revision: String(budget.revision),
    salespersonId:
      budget.salespersonId === null ? "" : String(budget.salespersonId),
    estimatorId: budget.estimatorId === null ? "" : String(budget.estimatorId),
    sentAt: toDateTimeLocalValue(budget.sentAt),
    sourceCompany: budget.sourceCompany,
    specificationDetails: budget.specificationDetails,
    statusId: String(budget.statusId),
    yearBudget: String(budget.yearBudget),
  };
}
