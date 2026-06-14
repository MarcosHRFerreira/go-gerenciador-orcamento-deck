import type { BudgetDetailItem } from "../types/budget";

export type BudgetFormValues = {
  areaM2: string;
  budgetNumber: string;
  commissionValue: string;
  competitorName: string;
  competitorPrice: string;
  contactId: string;
  currentFollowUp: string;
  designerName: string;
  grossValue: string;
  installerId: string;
  lossReasonId: string;
  priorityId: string;
  projectId: string;
  revision: string;
  salespersonId: string;
  sentAt: string;
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
    competitorName: "",
    competitorPrice: "",
    contactId: "",
    currentFollowUp: "",
    designerName: "",
    grossValue: "",
    installerId: "",
    lossReasonId: "",
    priorityId: "",
    projectId: "",
    revision: "0",
    salespersonId: "",
    sentAt: toDateTimeLocalValue(new Date().toISOString()),
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
    competitorName: budget.competitorName,
    competitorPrice:
      budget.competitorPrice === null ? "" : String(budget.competitorPrice),
    contactId: budget.contactId === null ? "" : String(budget.contactId),
    currentFollowUp: budget.currentFollowUp,
    designerName: budget.designerName,
    grossValue: String(budget.grossValue),
    installerId: budget.installerId === null ? "" : String(budget.installerId),
    lossReasonId:
      budget.lossReasonId === null ? "" : String(budget.lossReasonId),
    priorityId: budget.priorityId === null ? "" : String(budget.priorityId),
    projectId: budget.projectId === null ? "" : String(budget.projectId),
    revision: String(budget.revision),
    salespersonId:
      budget.salespersonId === null ? "" : String(budget.salespersonId),
    sentAt: toDateTimeLocalValue(budget.sentAt),
    specificationDetails: budget.specificationDetails,
    statusId: String(budget.statusId),
    yearBudget: String(budget.yearBudget),
  };
}
