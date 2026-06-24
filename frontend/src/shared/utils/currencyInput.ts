const brlCurrencyFormatter = new Intl.NumberFormat("pt-BR", {
  style: "currency",
  currency: "BRL",
});

function parseCurrencyNumber(value: string) {
  if (!value.trim()) {
    return null;
  }

  const parsedValue = Number(value);
  if (!Number.isFinite(parsedValue) || parsedValue < 0) {
    return null;
  }

  return parsedValue;
}

export function formatCurrencyInputValue(value: string) {
  const parsedValue = parseCurrencyNumber(value);
  if (parsedValue === null) {
    return "";
  }

  return brlCurrencyFormatter.format(parsedValue);
}

export function parseCurrencyInputToNumericString(value: string) {
  const digitsOnlyValue = value.replace(/\D/g, "");
  if (digitsOnlyValue.length === 0) {
    return "";
  }

  const parsedValue = Number(digitsOnlyValue) / 100;
  if (!Number.isFinite(parsedValue) || parsedValue < 0) {
    return "";
  }

  return String(parsedValue);
}
