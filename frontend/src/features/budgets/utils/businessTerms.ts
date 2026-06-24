function normalizeBusinessTerm(value: string | null | undefined) {
  return (value ?? "")
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "")
    .trim()
    .toLowerCase();
}

export function isWonStatusLabel(value: string | null | undefined) {
  const normalizedValue = normalizeBusinessTerm(value);

  return normalizedValue === "pedido" || normalizedValue === "fechado";
}

export function getBudgetStatusDisplayName(value: string | null | undefined) {
  const trimmedValue = (value ?? "").trim();

  if (!trimmedValue) {
    return trimmedValue;
  }

  return isWonStatusLabel(trimmedValue) ? "Fechado" : trimmedValue;
}

export function getWonStatusSingularLabel() {
  return "Fechado";
}

export function getWonStatusPluralLabel() {
  return "Fechados";
}

export function getFactorFieldLabel() {
  return "Fator";
}
