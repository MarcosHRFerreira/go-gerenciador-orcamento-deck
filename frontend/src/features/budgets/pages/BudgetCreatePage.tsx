import { useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "react-router-dom";
import { createBudgetRequest } from "../api/budgets";
import { BudgetForm } from "../components/BudgetForm";
import { createDefaultBudgetFormValues } from "../components/budgetFormValues";
import type { BudgetCreatePayload } from "../types/budget";

export function BudgetCreatePage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const handleSubmit = async (payload: BudgetCreatePayload) => {
    await createBudgetRequest(payload);
    await queryClient.invalidateQueries({ queryKey: ["budgets"] });
    navigate("/budgets");
  };

  return (
    <BudgetForm
      initialValues={createDefaultBudgetFormValues()}
      mode="create"
      onCancel={() => navigate("/budgets")}
      onSubmit={handleSubmit}
      submitLabel="Salvar orçamento"
      subtitle="Preencha os dados do orçamento para cadastrar um novo registro na base."
      title="Novo orçamento"
    />
  );
}
