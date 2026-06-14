import TrendingUpRoundedIcon from "@mui/icons-material/TrendingUpRounded";
import { Box, Chip, Typography } from "@mui/material";
import { PageHeader } from "../../../components/common/PageHeader";
import { SectionCard } from "../../../components/common/SectionCard";

const metrics = [
  {
    label: "Orcamentos ativos",
    value: "148",
    helper: "Visao geral do pipeline comercial",
  },
  {
    label: "Follow-ups pendentes",
    value: "23",
    helper: "Demandam retorno nas proximas 24h",
  },
  {
    label: "Valor em negociacao",
    value: "R$ 2,4 mi",
    helper: "Total consolidado do momento",
  },
];

export function DashboardPage() {
  return (
    <Box sx={{ display: "flex", flexDirection: "column", gap: 3 }}>
      <PageHeader
        description="Base inicial do painel administrativo com identidade visual moderna, limpa e profissional."
        title="Dashboard"
      />

      <Box
        sx={{
          display: "grid",
          gap: 3,
          gridTemplateColumns: {
            lg: "repeat(3, minmax(0, 1fr))",
            md: "repeat(2, minmax(0, 1fr))",
            xs: "minmax(0, 1fr)",
          },
        }}
      >
        {metrics.map((metric) => (
          <Box key={metric.label}>
            <SectionCard>
              <Box sx={{ display: "flex", flexDirection: "column", gap: 1.5 }}>
                <Chip
                  color="primary"
                  icon={<TrendingUpRoundedIcon />}
                  label="Resumo"
                  sx={{ alignSelf: "flex-start" }}
                  variant="outlined"
                />
                <Typography color="text.secondary" variant="body2">
                  {metric.label}
                </Typography>
                <Typography variant="h3">{metric.value}</Typography>
                <Typography color="text.secondary" variant="body2">
                  {metric.helper}
                </Typography>
              </Box>
            </SectionCard>
          </Box>
        ))}
      </Box>

      <SectionCard
        description="Esse bloco pode virar uma composicao com atalhos, indicadores e ultimas movimentacoes."
        title="Direcao do MVP"
      >
        <Box
          sx={{
            border: "1px dashed",
            borderColor: "divider",
            borderRadius: 4,
            p: 3,
          }}
        >
          <Typography color="text.secondary" variant="body1">
            O proximo passo funcional e integrar autenticacao real, dados do
            usuario logado e a listagem de orcamentos consumindo a API Go.
          </Typography>
        </Box>
      </SectionCard>
    </Box>
  );
}
