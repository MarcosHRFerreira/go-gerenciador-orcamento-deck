import { Alert, Box, LinearProgress } from "@mui/material";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useMemo } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { PageHeader } from "../../../components/common/PageHeader";
import { SectionCard } from "../../../components/common/SectionCard";
import EstimatorForm, {
  mapEstimatorToInitialValues,
} from "../components/EstimatorForm";
import { listUsersRequest } from "../../users/api/users";
import {
  listEstimatorsRequest,
  updateEstimatorRequest,
} from "../api/estimators";
import type {
  CreateEstimatorPayload,
  UpdateEstimatorPayload,
} from "../types/estimator";

function getEstimatorId(value: string | undefined) {
  const parsedValue = Number(value);

  if (!Number.isInteger(parsedValue) || parsedValue <= 0) {
    return null;
  }

  return parsedValue;
}

export default function EstimatorEditRoutePage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { estimatorId: estimatorIdParam } = useParams();
  const estimatorId = getEstimatorId(estimatorIdParam);
  const estimatorsQuery = useQuery({
    enabled: estimatorId !== null,
    queryFn: listEstimatorsRequest,
    queryKey: ["estimators"],
  });
  const usersQuery = useQuery({
    queryFn: listUsersRequest,
    queryKey: ["users", "estimator-links"],
    staleTime: 1000 * 60 * 5,
  });
  const currentEstimator = useMemo(() => {
    if (estimatorId === null) {
      return null;
    }

    return (
      estimatorsQuery.data?.find((item) => item.id === estimatorId) ?? null
    );
  }, [estimatorId, estimatorsQuery.data]);
  const initialValues = useMemo(
    () => mapEstimatorToInitialValues(currentEstimator),
    [currentEstimator],
  );

  const handleSubmit = async (
    payload: CreateEstimatorPayload | UpdateEstimatorPayload,
  ) => {
    if (currentEstimator === null) {
      throw new Error("Não foi possível identificar o orçamentista.");
    }

    if (!("active" in payload)) {
      throw new Error("Payload inválido para edição de orçamentista.");
    }

    await updateEstimatorRequest(currentEstimator.id, payload);
    await queryClient.invalidateQueries({ queryKey: ["estimators"] });
    navigate("/estimators");
  };

  if (estimatorId === null) {
    return (
      <Box sx={{ display: "flex", flexDirection: "column", gap: 3 }}>
        <PageHeader
          description="O identificador informado para edição não é válido."
          title="Editar orçamentista"
        />
        <SectionCard
          description="Revise a navegação e tente novamente pela listagem de orçamentistas."
          title="Orçamentista inválido"
        >
          <Alert severity="error">
            Não foi possível identificar o orçamentista.
          </Alert>
        </SectionCard>
      </Box>
    );
  }

  if (estimatorsQuery.isLoading || usersQuery.isLoading) {
    return (
      <Box sx={{ display: "flex", flexDirection: "column", gap: 3 }}>
        <PageHeader
          description="Carregando os dados cadastrais para edição."
          title="Editar orçamentista"
        />
        <SectionCard
          description="Aguarde enquanto os dados do orçamentista e os vínculos disponíveis são carregados."
          title="Dados do orçamentista"
        >
          <LinearProgress />
        </SectionCard>
      </Box>
    );
  }

  if (
    estimatorsQuery.isError ||
    usersQuery.isError ||
    currentEstimator === null
  ) {
    return (
      <Box sx={{ display: "flex", flexDirection: "column", gap: 3 }}>
        <PageHeader
          description="Não foi possível carregar o orçamentista solicitado."
          title="Editar orçamentista"
        />
        <SectionCard
          description="Volte para a listagem e selecione novamente o orçamentista."
          title="Falha ao carregar"
        >
          <Alert severity="error">
            Orçamentista não encontrado para edição.
          </Alert>
        </SectionCard>
      </Box>
    );
  }

  return (
    <EstimatorForm
      codeHelpText="O código é protegido na edição para preservar a rastreabilidade do cadastro."
      codeReadOnly
      currentEstimator={currentEstimator}
      initialValues={initialValues}
      linkedUsers={usersQuery.data ?? []}
      mode="edit"
      onCancel={() => navigate("/estimators")}
      onSubmit={handleSubmit}
      submitLabel="Salvar alterações"
      subtitle="Atualize os dados cadastrais e o vínculo operacional do orçamentista selecionado."
      title="Editar orçamentista"
    />
  );
}
