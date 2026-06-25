import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useMemo } from "react";
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

  const initialValues = useMemo(
    () =>
      budgetQuery.data === undefined
        ? createDefaultBudgetFormValues()
        : mapBudgetDetailToFormValues(budgetQuery.data),
    [budgetQuery.data],
  );
  const formInstanceKey =
    budgetQuery.data === undefined
      ? "budget-edit-loading"
      : `budget-edit-${budgetQuery.data.id}-${budgetQuery.data.updatedAt}`;

  return (
    <BudgetForm
      key={formInstanceKey}
      currentContactId={budgetQuery.data?.contactId ?? null}
      currentContactLabel={budgetQuery.data?.contactName ?? null}
      currentEstimatorId={budgetQuery.data?.estimatorId ?? null}
      currentEstimatorLabel={budgetQuery.data?.estimatorName ?? null}
      currentInstallerId={budgetQuery.data?.installerId ?? null}
      currentInstallerLabel={budgetQuery.data?.installerName ?? null}
      currentLossReasonId={budgetQuery.data?.lossReasonId ?? null}
      currentLossReasonLabel={budgetQuery.data?.lossReasonName ?? null}
      currentProductLineId={budgetQuery.data?.productLineId ?? null}
      currentProductLineLabel={budgetQuery.data?.productLineName ?? null}
      currentProjectId={budgetQuery.data?.projectId ?? null}
      currentProjectLabel={budgetQuery.data?.projectName ?? null}
      currentSalespersonId={budgetQuery.data?.salespersonId ?? null}
      currentSalespersonLabel={budgetQuery.data?.salespersonName ?? null}
      currentSystemTypeId={budgetQuery.data?.systemTypeId ?? null}
      currentSystemTypeLabel={budgetQuery.data?.systemTypeName ?? null}
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
