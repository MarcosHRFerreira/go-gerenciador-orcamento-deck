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
import {
  compactFilterFieldSx,
  FilterField,
  filterGroupSx,
  filterGroupTitleSx,
  filterSectionCardSx,
} from "../../../components/common/FilterField";
import { PageHeader } from "../../../components/common/PageHeader";
import {
  ResizableTableHeadCell,
  useResizableTableColumns,
} from "../../../components/common/ResizableTable";
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
  background: "linear-gradient(180deg, #1E3A8A 0%, #1D4ED8 100%)",
  borderBottomColor: "#1E40AF",
  borderBottomWidth: 2,
  color: "common.white",
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

const projectListColumnDefinitions = [
  { key: "code", width: 140, minWidth: 120 },
  { key: "name", width: 280, minWidth: 220 },
  { key: "type", width: 180, minWidth: 160 },
  { key: "city", width: 160, minWidth: 140 },
  { key: "state", width: 120, minWidth: 100 },
  { key: "updatedAt", width: 180, minWidth: 160 },
  { key: "actions", width: 260, minWidth: 220 },
] as const;

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
  const { createResizeHandler, getColumnWidth } = useResizableTableColumns(
    "project-list-columns:v1",
    [...projectListColumnDefinitions],
  );
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
        getMutationErrorMessage(error, "Não foi possível remover a obra."),
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
        description="Consulte, cadastre e mantenha o catálogo de obras usado nos orçamentos."
        title="Obras"
      />

      <SectionCard
        description="Busque pelo código ou pela descrição da obra."
        sx={filterSectionCardSx}
        title="Consulta"
      >
        <Box sx={filterGroupSx}>
          <Typography sx={filterGroupTitleSx} variant="subtitle2">
            Identificação
          </Typography>
          <FilterField label="Buscar obra">
            <TextField
              fullWidth
              onChange={(event) => setSearch(event.target.value)}
              placeholder="Ex: OBR-000001 ou Centro Empresarial"
              size="small"
              sx={compactFilterFieldSx}
              value={search}
            />
          </FilterField>
        </Box>
      </SectionCard>

      <SectionCard
        description="Cadastros disponíveis para associação e manutenção administrativa."
        title="Lista de obras"
      >
        {feedbackMessage ? (
          <Alert severity="success">{feedbackMessage}</Alert>
        ) : null}
        {feedbackError ? <Alert severity="error">{feedbackError}</Alert> : null}
        {projectsQuery.isError ? (
          <Alert severity="error">
            Não foi possível carregar as obras cadastradas.
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
                  <ResizableTableHeadCell
                    onResizeStart={createResizeHandler("code")}
                    sx={tableHeadCellSx}
                    width={getColumnWidth("code")}
                  >
                    Código
                  </ResizableTableHeadCell>
                  <ResizableTableHeadCell
                    onResizeStart={createResizeHandler("name")}
                    sx={tableHeadCellSx}
                    width={getColumnWidth("name")}
                  >
                    Descrição
                  </ResizableTableHeadCell>
                  <ResizableTableHeadCell
                    onResizeStart={createResizeHandler("type")}
                    sx={tableHeadCellSx}
                    width={getColumnWidth("type")}
                  >
                    Tipo
                  </ResizableTableHeadCell>
                  <ResizableTableHeadCell
                    onResizeStart={createResizeHandler("city")}
                    sx={tableHeadCellSx}
                    width={getColumnWidth("city")}
                  >
                    Cidade
                  </ResizableTableHeadCell>
                  <ResizableTableHeadCell
                    onResizeStart={createResizeHandler("state")}
                    sx={tableHeadCellSx}
                    width={getColumnWidth("state")}
                  >
                    Estado
                  </ResizableTableHeadCell>
                  <ResizableTableHeadCell
                    onResizeStart={createResizeHandler("updatedAt")}
                    sx={tableHeadCellSx}
                    width={getColumnWidth("updatedAt")}
                  >
                    Atualizado em
                  </ResizableTableHeadCell>
                  <ResizableTableHeadCell
                    onResizeStart={createResizeHandler("actions")}
                    sx={tableHeadCellSx}
                    width={getColumnWidth("actions")}
                  >
                    Ações
                  </ResizableTableHeadCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {filteredProjects.map((item) => (
                  <TableRow hover key={item.id}>
                    <TableCell
                      sx={{ ...tableDetailCellSx, width: getColumnWidth("code") }}
                    >
                      <Typography sx={{ fontWeight: 600 }} variant="body2">
                        {item.code}
                      </Typography>
                    </TableCell>
                    <TableCell
                      sx={{ ...tableDetailCellSx, width: getColumnWidth("name") }}
                    >
                      <Typography variant="body2">{item.name}</Typography>
                    </TableCell>
                    <TableCell
                      sx={{ ...tableDetailCellSx, width: getColumnWidth("type") }}
                    >
                      {item.projectTypeId === null
                        ? "Não informado"
                        : (projectTypeMap.get(item.projectTypeId) ??
                          `#${item.projectTypeId}`)}
                    </TableCell>
                    <TableCell
                      sx={{ ...tableDetailCellSx, width: getColumnWidth("city") }}
                    >
                      {item.city.trim() ? item.city : "Não informado"}
                    </TableCell>
                    <TableCell
                      sx={{ ...tableDetailCellSx, width: getColumnWidth("state") }}
                    >
                      {item.state.trim() ? item.state : "Não informado"}
                    </TableCell>
                    <TableCell
                      sx={{
                        ...tableDetailCellSx,
                        width: getColumnWidth("updatedAt"),
                      }}
                    >
                      {formatDateTime(item.updatedAt)}
                    </TableCell>
                    <TableCell
                      sx={{
                        ...tableDetailCellSx,
                        whiteSpace: "nowrap",
                        width: getColumnWidth("actions"),
                      }}
                    >
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
            Confirma a exclusão da obra{" "}
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
