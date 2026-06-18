import AddRoundedIcon from "@mui/icons-material/AddRounded";
import DeleteOutlineRoundedIcon from "@mui/icons-material/DeleteOutlineRounded";
import EditRoundedIcon from "@mui/icons-material/EditRounded";
import VisibilityRoundedIcon from "@mui/icons-material/VisibilityRounded";
import {
  Alert,
  Box,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TextField,
  Typography,
} from "@mui/material";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { isAxiosError } from "axios";
import { useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { PageHeader } from "../../../components/common/PageHeader";
import { SectionCard } from "../../../components/common/SectionCard";
import {
  deleteProjectRequest,
  listProjectsRequest,
  listProjectTypesRequest,
} from "../api/projects";
import type { ProjectItem, ProjectTypeCatalogItem } from "../types/project";

const dateTimeFormatter = new Intl.DateTimeFormat("pt-BR", {
  dateStyle: "short",
  timeStyle: "short",
});

const tableHeadCellSx = {
  backgroundColor: "rgba(37, 99, 235, 0.08)",
  borderBottomColor: "primary.main",
  borderBottomWidth: 2,
  color: "text.primary",
  fontSize: "0.75rem",
  fontWeight: 700,
  letterSpacing: "0.04em",
  py: 1.5,
  textTransform: "uppercase",
  whiteSpace: "nowrap",
};

const tableDetailCellSx = {
  color: "text.secondary",
  fontSize: "0.82rem",
  lineHeight: 1.45,
  py: 1.25,
  verticalAlign: "top",
};

function normalizeText(value: string) {
  return value
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "")
    .trim()
    .toLowerCase();
}

function createNameMap<T extends { id: number; name: string }>(items: T[]) {
  return new Map(items.map((item) => [item.id, item.name]));
}

function formatDateTime(value: string) {
  return dateTimeFormatter.format(new Date(value));
}

function getMutationErrorMessage(error: unknown, fallbackMessage: string) {
  if (isAxiosError<{ message?: string }>(error)) {
    return error.response?.data?.message ?? fallbackMessage;
  }

  return fallbackMessage;
}

export default function ProjectListPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [search, setSearch] = useState("");
  const [feedbackMessage, setFeedbackMessage] = useState<string | null>(null);
  const [feedbackError, setFeedbackError] = useState<string | null>(null);
  const [pendingDelete, setPendingDelete] = useState<ProjectItem | null>(null);

  const projectsQuery = useQuery({
    queryKey: ["projects"],
    queryFn: listProjectsRequest,
  });
  const projectTypesQuery = useQuery({
    queryKey: ["project-types"],
    queryFn: listProjectTypesRequest,
  });

  const deleteMutation = useMutation({
    mutationFn: deleteProjectRequest,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["projects"] });
      setFeedbackError(null);
      setFeedbackMessage("Obra removida com sucesso.");
      setPendingDelete(null);
    },
    onError: (error) => {
      setFeedbackMessage(null);
      setFeedbackError(
        getMutationErrorMessage(error, "Nao foi possivel remover a obra."),
      );
    },
  });

  const projectTypeMap = useMemo(
    () => createNameMap<ProjectTypeCatalogItem>(projectTypesQuery.data ?? []),
    [projectTypesQuery.data],
  );
  const filteredProjects = useMemo(() => {
    const normalizedSearch = normalizeText(search);

    return (projectsQuery.data ?? []).filter((item) => {
      if (normalizedSearch === "") {
        return true;
      }

      return (
        normalizeText(item.code).includes(normalizedSearch) ||
        normalizeText(item.name).includes(normalizedSearch)
      );
    });
  }, [projectsQuery.data, search]);

  const handleConfirmDelete = async () => {
    if (pendingDelete === null) {
      return;
    }

    await deleteMutation.mutateAsync(pendingDelete.id);
  };

  return (
    <Box sx={{ display: "flex", flexDirection: "column", gap: 3 }}>
      <PageHeader
        action={
          <Button
            onClick={() => navigate("/projects/new")}
            startIcon={<AddRoundedIcon />}
            variant="contained"
          >
            Nova obra
          </Button>
        }
        description="Consulte, cadastre e mantenha o catalogo de obras usado nos orcamentos."
        title="Obras"
      />

      <SectionCard
        description="Busque pelo codigo ou pela descricao da obra."
        title="Consulta"
      >
        <TextField
          fullWidth
          label="Buscar obra"
          onChange={(event) => setSearch(event.target.value)}
          placeholder="Ex: OBR-000001 ou Centro Empresarial"
          size="small"
          value={search}
        />
      </SectionCard>

      <SectionCard
        description="Cadastros disponiveis para associacao e manutencao administrativa."
        title="Lista de obras"
      >
        {feedbackMessage ? (
          <Alert severity="success">{feedbackMessage}</Alert>
        ) : null}
        {feedbackError ? <Alert severity="error">{feedbackError}</Alert> : null}
        {projectsQuery.isError ? (
          <Alert severity="error">
            Nao foi possivel carregar as obras cadastradas.
          </Alert>
        ) : null}

        {projectsQuery.isSuccess && filteredProjects.length === 0 ? (
          <Alert severity="info" variant="outlined">
            Nenhuma obra encontrada para o filtro informado.
          </Alert>
        ) : null}

        {projectsQuery.isSuccess && filteredProjects.length > 0 ? (
          <TableContainer>
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell sx={tableHeadCellSx}>Codigo</TableCell>
                  <TableCell sx={tableHeadCellSx}>Descricao</TableCell>
                  <TableCell sx={tableHeadCellSx}>Tipo</TableCell>
                  <TableCell sx={tableHeadCellSx}>Cidade</TableCell>
                  <TableCell sx={tableHeadCellSx}>Estado</TableCell>
                  <TableCell sx={tableHeadCellSx}>Atualizado em</TableCell>
                  <TableCell sx={tableHeadCellSx}>Acoes</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {filteredProjects.map((item) => (
                  <TableRow hover key={item.id}>
                    <TableCell sx={tableDetailCellSx}>
                      <Typography sx={{ fontWeight: 600 }} variant="body2">
                        {item.code}
                      </Typography>
                    </TableCell>
                    <TableCell sx={tableDetailCellSx}>
                      <Typography variant="body2">{item.name}</Typography>
                    </TableCell>
                    <TableCell sx={tableDetailCellSx}>
                      {item.projectTypeId === null
                        ? "Nao informado"
                        : (projectTypeMap.get(item.projectTypeId) ??
                          `#${item.projectTypeId}`)}
                    </TableCell>
                    <TableCell sx={tableDetailCellSx}>
                      {item.city.trim() ? item.city : "Nao informado"}
                    </TableCell>
                    <TableCell sx={tableDetailCellSx}>
                      {item.state.trim() ? item.state : "Nao informado"}
                    </TableCell>
                    <TableCell sx={tableDetailCellSx}>
                      {formatDateTime(item.updatedAt)}
                    </TableCell>
                    <TableCell sx={tableDetailCellSx}>
                      <Button
                        onClick={() => navigate(`/projects/${item.id}`)}
                        size="small"
                        startIcon={<VisibilityRoundedIcon />}
                        sx={{ minWidth: "auto", px: 0.75, py: 0.25 }}
                        variant="text"
                      >
                        Abrir
                      </Button>
                      <Button
                        onClick={() => navigate(`/projects/${item.id}/edit`)}
                        size="small"
                        startIcon={<EditRoundedIcon />}
                        sx={{ minWidth: "auto", px: 0.75, py: 0.25 }}
                        variant="text"
                      >
                        Editar
                      </Button>
                      <Button
                        color="error"
                        onClick={() => setPendingDelete(item)}
                        size="small"
                        startIcon={<DeleteOutlineRoundedIcon />}
                        sx={{ minWidth: "auto", px: 0.75, py: 0.25 }}
                        variant="text"
                      >
                        Excluir
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        ) : null}
      </SectionCard>

      <Dialog
        onClose={() => setPendingDelete(null)}
        open={pendingDelete !== null}
      >
        <DialogTitle>Excluir obra</DialogTitle>
        <DialogContent>
          <DialogContentText>
            Confirma a exclusao da obra{" "}
            <strong>{pendingDelete?.code ?? ""}</strong>?
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button
            disabled={deleteMutation.isPending}
            onClick={() => setPendingDelete(null)}
            variant="outlined"
          >
            Cancelar
          </Button>
          <Button
            color="error"
            disabled={deleteMutation.isPending}
            onClick={handleConfirmDelete}
            variant="contained"
          >
            {deleteMutation.isPending ? "Excluindo..." : "Excluir"}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
