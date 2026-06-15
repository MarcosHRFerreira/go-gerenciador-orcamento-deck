import { useQueryClient } from "@tanstack/react-query";
import { useQuery } from "@tanstack/react-query";
import { useNavigate, useSearchParams } from "react-router-dom";
import { createBudgetRequest } from "../api/budgets";
import { BudgetForm } from "../components/BudgetForm";
import { createDefaultBudgetFormValues } from "../components/budgetFormValues";
import type { BudgetCreatePayload } from "../types/budget";
import { getProjectByIdRequest } from "../../projects/api/projects";

function parseOptionalProjectId(value: string | null) {
  if (value === null || value.trim() === "") {
    return null;
  }

  const parsedValue = Number(value);

  if (!Number.isInteger(parsedValue) || parsedValue <= 0) {
    return null;
  }

  return parsedValue;
}

export function BudgetCreatePage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [searchParams] = useSearchParams();
  const projectId = parseOptionalProjectId(searchParams.get("projectId"));
  const returnToProject = searchParams.get("returnTo") === "project";
  const projectQuery = useQuery({
    enabled: projectId !== null,
    queryFn: () => getProjectByIdRequest(projectId as number),
    queryKey: ["project", projectId],
  });

  const navigateBack = () => {
    if (projectId !== null && returnToProject) {
      navigate(`/projects/${projectId}`);
      return;
    }

    navigate("/budgets");
  };

  const handleSubmit = async (payload: BudgetCreatePayload) => {
    await createBudgetRequest(payload);
    await queryClient.invalidateQueries({ queryKey: ["budgets"] });
    if (projectId !== null) {
      await queryClient.invalidateQueries({ queryKey: ["project-budgets", projectId] });
    }
    navigateBack();
  };

  const initialValues = (() => {
    const defaultValues = createDefaultBudgetFormValues();
    if (projectId === null) {
      return defaultValues;
    }

    return {
      ...defaultValues,
      projectId: String(projectId),
    };
  })();

  const initialDataError =
    searchParams.get("projectId") !== null && projectId === null
      ? "Projeto informado para o novo orcamento e invalido."
      : projectQuery.isError
        ? "Nao foi possivel carregar o projeto informado para o novo orcamento."
        : null;

  return (
    <BudgetForm
      initialDataError={initialDataError}
      initialValues={initialValues}
      lockedProjectId={projectId}
      lockedProjectLabel={projectQuery.data?.name ?? null}
      mode="create"
      onCancel={navigateBack}
      onSubmit={handleSubmit}
      submitLabel="Salvar orçamento"
      subtitle={
        projectId !== null
          ? `Preencha os dados para cadastrar um novo orcamento ja vinculado ao projeto ${projectQuery.data?.name ?? `#${projectId}`}.`
          : "Preencha os dados do orçamento para cadastrar um novo registro na base."
      }
      title={projectId !== null ? "Novo orçamento do projeto" : "Novo orçamento"}
    />
  );
}
