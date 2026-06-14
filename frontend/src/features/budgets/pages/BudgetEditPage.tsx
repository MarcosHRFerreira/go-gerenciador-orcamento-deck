import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useNavigate, useParams } from "react-router-dom";
import { getBudgetByIdRequest, updateBudgetRequest } from "../api/budgets";
import { BudgetForm } from "../components/BudgetForm";
import {
  createDefaultBudgetFormValues,
  mapBudgetDetailToFormValues,
} from "../components/budgetFormValues";
import type { BudgetCreatePayload } from "../types/budget";

function getBudgetId(value: string | undefined) {
  const parsedValue = Number(value);

  if (!Number.isInteger(parsedValue) || parsedValue <= 0) {
    return null;
  }

  return parsedValue;
}

export function BudgetEditPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { budgetId: budgetIdParam } = useParams();
  const budgetId = getBudgetId(budgetIdParam);
  const budgetQuery = useQuery({
    enabled: budgetId !== null,
    queryFn: () => getBudgetByIdRequest(budgetId as number),
    queryKey: ["budget", budgetId],
  });

  const handleSubmit = async (payload: BudgetCreatePayload) => {
    if (budgetId === null) {
      return;
    }

    await updateBudgetRequest(budgetId, payload);
    await queryClient.invalidateQueries({ queryKey: ["budgets"] });
    await queryClient.invalidateQueries({ queryKey: ["budget", budgetId] });
    navigate("/budgets");
  };

  const initialDataError =
    budgetId === null
      ? "ID do orçamento inválido."
      : budgetQuery.isError
        ? "Não foi possível carregar os dados do orçamento."
        : null;

  const initialValues =
    budgetQuery.data === undefined
      ? createDefaultBudgetFormValues()
      : mapBudgetDetailToFormValues(budgetQuery.data);

  return (
    <BudgetForm
      initialDataError={initialDataError}
      initialValues={initialValues}
      isInitialDataLoading={budgetQuery.isLoading}
      mode="edit"
      onCancel={() => navigate("/budgets")}
      onSubmit={handleSubmit}
      submitLabel="Salvar alterações"
      subtitle="Atualize os dados do orçamento e salve as alterações na base."
      title="Editar orçamento"
    />
  );
}
