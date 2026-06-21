import { Alert, Box, LinearProgress } from "@mui/material";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useMemo } from "react";
import { useNavigate } from "react-router-dom";
import EstimatorForm, {
  defaultEstimatorFormValues,
} from "../components/EstimatorForm";
import {
  createEstimatorRequest,
  getNextEstimatorCodeRequest,
} from "../api/estimators";
import { listUsersRequest } from "../../users/api/users";
import { PageHeader } from "../../../components/common/PageHeader";
import { SectionCard } from "../../../components/common/SectionCard";
import type {
  CreateEstimatorPayload,
  UpdateEstimatorPayload,
} from "../types/estimator";

export default function EstimatorCreatePage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const nextCodeQuery = useQuery({
    queryFn: getNextEstimatorCodeRequest,
    queryKey: ["estimators", "next-code", "page"],
    staleTime: 0,
  });
  const usersQuery = useQuery({
    queryFn: listUsersRequest,
    queryKey: ["users", "estimator-links"],
    staleTime: 1000 * 60 * 5,
  });

  const initialValues = useMemo(
    () => ({
      ...defaultEstimatorFormValues,
      code: nextCodeQuery.data ?? "",
    }),
    [nextCodeQuery.data],
  );

  const handleSubmit = async (
    payload: CreateEstimatorPayload | UpdateEstimatorPayload,
  ) => {
    if ("active" in payload) {
      throw new Error("Payload inválido para criação de orçamentista.");
    }

    await createEstimatorRequest(payload);
    await queryClient.invalidateQueries({ queryKey: ["estimators"] });
    navigate("/estimators");
  };

  if (nextCodeQuery.isLoading || usersQuery.isLoading) {
    return (
      <Box sx={{ display: "flex", flexDirection: "column", gap: 3 }}>
        <PageHeader
          description="Carregando os dados necessários para o novo cadastro."
          title="Novo orçamentista"
        />
        <SectionCard
          description="Aguarde enquanto o próximo código e os vínculos disponíveis são preparados."
          title="Dados do orçamentista"
        >
          <LinearProgress />
        </SectionCard>
      </Box>
    );
  }

  if (nextCodeQuery.isError || usersQuery.isError) {
    return (
      <Box sx={{ display: "flex", flexDirection: "column", gap: 3 }}>
        <PageHeader
          description="Não foi possível carregar os dados necessários para o cadastro."
          title="Novo orçamentista"
        />
        <SectionCard
          description="Tente novamente a partir da listagem de orçamentistas."
          title="Falha ao carregar"
        >
          <Alert severity="error">
            Não foi possível preparar o cadastro do orçamentista.
          </Alert>
        </SectionCard>
      </Box>
    );
  }

  return (
    <EstimatorForm
      codeHelpText="Código sugerido automaticamente para o novo cadastro."
      codeReadOnly
      initialValues={initialValues}
      linkedUsers={usersQuery.data ?? []}
      mode="create"
      onCancel={() => navigate("/estimators")}
      onSubmit={handleSubmit}
      submitLabel="Cadastrar orçamentista"
      subtitle="Cadastre um novo orçamentista com código, contatos e vínculo operacional."
      title="Novo orçamentista"
    />
  );
}
