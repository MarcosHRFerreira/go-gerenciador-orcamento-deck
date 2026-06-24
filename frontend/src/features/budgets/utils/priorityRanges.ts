import type { BudgetCatalogItem, BudgetListItem } from "../types/budget";

const faixa0a50kLabel = "Faixa 0 a 50k";
const faixa50ka250kLabel = "Faixa 50k a 250k";
const faixaAcima250kLabel = "Faixa acima de 250k";

function normalizePriorityLabel(value: string) {
  return value
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "")
    .trim()
    .toLowerCase();
}

export function getPriorityDisplayLabel(value: string) {
  const normalizedValue = normalizePriorityLabel(value);

  if (
    normalizedValue === normalizePriorityLabel(faixa0a50kLabel) ||
    normalizedValue === "faixa 0 entre 50k" ||
    normalizedValue === "faixa 0 ate 50k"
  ) {
    return faixa0a50kLabel;
  }

  if (
    normalizedValue === normalizePriorityLabel(faixa50ka250kLabel) ||
    normalizedValue === "faixa 50k entre 250k" ||
    normalizedValue === "faixa 50k ate 250k"
  ) {
    return faixa50ka250kLabel;
  }

  if (
    normalizedValue === normalizePriorityLabel(faixaAcima250kLabel) ||
    normalizedValue === "faixa 250k maior"
  ) {
    return faixaAcima250kLabel;
  }

  return value;
}

export function getPriorityLabelByGrossValue(grossValue: number): string {
  if (grossValue <= 50000) {
    return faixa0a50kLabel;
  }

  if (grossValue <= 250000) {
    return faixa50ka250kLabel;
  }

  return faixaAcima250kLabel;
}

export function parseBudgetDecimalInput(value: string): number | null {
  const normalizedValue = value.trim();
  if (normalizedValue.length === 0) {
    return null;
  }

  const decimalValue = normalizedValue
    .replace(/\./g, "")
    .replace(",", ".")
    .replace(/\s+/g, "");
  const parsedValue = Number(decimalValue);

  return Number.isFinite(parsedValue) ? parsedValue : null;
}

export function resolvePriorityIdByGrossValue(
  grossValue: number,
  priorities: BudgetCatalogItem[],
): number | null {
  const expectedLabel = getPriorityLabelByGrossValue(grossValue);
  const matchedPriority = priorities.find(
    (priority) => getPriorityDisplayLabel(priority.name) === expectedLabel,
  );

  return matchedPriority?.id ?? null;
}

export function getBudgetPriorityLabel(
  budget: Pick<BudgetListItem, "grossValue">,
) {
  return getPriorityLabelByGrossValue(budget.grossValue);
}
