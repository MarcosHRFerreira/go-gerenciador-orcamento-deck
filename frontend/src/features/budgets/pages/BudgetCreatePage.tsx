import { useQueryClient } from "@tanstack/react-query";
import { useQuery } from "@tanstack/react-query";
import { useNavigate, useSearchParams } from "react-router-dom";
import { createBudgetRequest } from "../api/budgets";
import { BudgetForm } from "../components/BudgetForm";
import { createDefaultBudgetFormValues } from "../components/budgetFormValues";
import type { BudgetCreatePayload } from "../types/budget";
import { getProjectByIdRequest } from "../../projects/api/projects";
import { useAuth } from "../../auth/hooks/useAuth";

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
  const { user } = useAuth();
  const [searchParams] = useSearchParams();
  const projectId = parseOptionalProjectId(searchParams.get("projectId"));
  const returnToProject = searchParams.get("returnTo") === "project";
  const canCreateBudget =
    user?.role === "admin" ||
    (user?.role === "user" && user.user_kind === "estimator");
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
    if (!canCreateBudget) {
      return;
    }

    await createBudgetRequest(payload);
    await queryClient.invalidateQueries({ queryKey: ["budgets"] });
    if (projectId !== null) {
      await queryClient.invalidateQueries({
        queryKey: ["project-budgets", projectId],
      });
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

  const initialDataError = !canCreateBudget
    ? "Seu perfil nao possui permissao para criar orcamentos."
    : searchParams.get("projectId") !== null && projectId === null
      ? "Obra informada para o novo orcamento e invalida."
      : projectQuery.isError
        ? "Nao foi possivel carregar a obra informada para o novo orcamento."
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
          ? `Preencha os dados para cadastrar um novo orcamento ja vinculado a obra ${projectQuery.data?.name ?? `#${projectId}`}.`
          : "Preencha os dados do orçamento para cadastrar um novo registro na base."
      }
      title={projectId !== null ? "Novo orçamento da obra" : "Novo orçamento"}
    />
  );
}
