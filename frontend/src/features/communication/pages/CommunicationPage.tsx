import AddRoundedIcon from "@mui/icons-material/AddRounded";
import CampaignRoundedIcon from "@mui/icons-material/CampaignRounded";
import ForumRoundedIcon from "@mui/icons-material/ForumRounded";
import MarkEmailReadRoundedIcon from "@mui/icons-material/MarkEmailReadRounded";
import MarkEmailUnreadRoundedIcon from "@mui/icons-material/MarkEmailUnreadRounded";
import NotificationsActiveRoundedIcon from "@mui/icons-material/NotificationsActiveRounded";
import PushPinRoundedIcon from "@mui/icons-material/PushPinRounded";
import SendRoundedIcon from "@mui/icons-material/SendRounded";
import {
  Alert,
  Autocomplete,
  Badge,
  Box,
  Button,
  Checkbox,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Divider,
  FormControlLabel,
  LinearProgress,
  MenuItem,
  Tab,
  Tabs,
  TextField,
  Typography,
} from "@mui/material";
import { alpha } from "@mui/material/styles";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { isAxiosError } from "axios";
import { useEffect, useMemo, useRef, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { PageHeader } from "../../../components/common/PageHeader";
import { SectionCard } from "../../../components/common/SectionCard";
import { useAuth } from "../../auth/hooks/useAuth";
import { getProjectByIdRequest } from "../../projects/api/projects";
import {
  communicationPollingIntervals,
  communicationQueryKeys,
  createConversationRequest,
  createNoticeRequest,
  getConversationUnreadCountRequest,
  getNoticeUnreadCountRequest,
  listConversationAvailableUsersRequest,
  listConversationMessagesRequest,
  listConversationsRequest,
  listNoticesRequest,
  markConversationReadRequest,
  markNoticeReadRequest,
  sendConversationMessageRequest,
} from "../api/communication";
import type {
  ConversationItem,
  ConversationMessageItem,
  ConversationParticipantItem,
  CreateConversationPayload,
  CreateNoticePayload,
  NoticeItem,
  NoticePriority,
  NoticeScopeType,
  NoticeStatusFilter,
  SendConversationMessagePayload,
} from "../types/communication";

const dateTimeFormatter = new Intl.DateTimeFormat("pt-BR", {
  dateStyle: "short",
  timeStyle: "short",
});
const communicationBlue = "#1E3A8A";
const communicationFeedbackAlertSx = {
  borderRadius: 3,
  boxShadow: "0 14px 30px rgba(30, 58, 138, 0.08)",
  "& .MuiAlert-message": {
    fontWeight: 600,
  },
} as const;

const communicationInfoAlertSx = {
  borderRadius: 3,
  boxShadow: "0 14px 28px rgba(30, 58, 138, 0.07)",
  "& .MuiAlert-message": {
    fontWeight: 500,
    lineHeight: 1.65,
  },
} as const;

const communicationLoaderSx = {
  borderRadius: 999,
  height: 8,
  overflow: "hidden",
  "& .MuiLinearProgress-bar": {
    borderRadius: 999,
  },
} as const;

type CreateNoticeFormState = {
  title: string;
  body: string;
  scopeType: NoticeScopeType;
  priority: NoticePriority;
  pinned: boolean;
  expiresAt: string;
  recipientUserIds: number[];
};

type CommunicationTab = "notices" | "conversations";

type CreateConversationFormState = {
  participantUserId: number | null;
  projectId: number | null;
  initialMessage: string;
};

const defaultCreateNoticeFormState: CreateNoticeFormState = {
  title: "",
  body: "",
  scopeType: "all",
  priority: "info",
  pinned: false,
  expiresAt: "",
  recipientUserIds: [],
};

const defaultCreateConversationFormState: CreateConversationFormState = {
  participantUserId: null,
  projectId: null,
  initialMessage: "",
};

function getErrorMessage(error: unknown, fallbackMessage: string) {
  if (isAxiosError<{ message?: string }>(error)) {
    return error.response?.data?.message ?? fallbackMessage;
  }

  return fallbackMessage;
}

function getPriorityLabel(priority: NoticePriority) {
  if (priority === "critical") {
    return "Crítico";
  }

  if (priority === "warning") {
    return "Atenção";
  }

  return "Informativo";
}

function getPriorityColor(priority: NoticePriority) {
  if (priority === "critical") {
    return "error" as const;
  }

  if (priority === "warning") {
    return "warning" as const;
  }

  return "info" as const;
}

function getPriorityAccentColor(priority: NoticePriority) {
  if (priority === "critical") {
    return "#DC2626";
  }

  if (priority === "warning") {
    return "#D97706";
  }

  return "#2563EB";
}

function getScopeLabel(scopeType: NoticeScopeType) {
  return scopeType === "all" ? "Geral" : "Direcionado";
}

function getRoleLabel(role: string) {
  return role === "admin" ? "Administrador" : "Usuário";
}

function formatDateTime(value: string | null) {
  if (!value) {
    return "Não informado";
  }

  return dateTimeFormatter.format(new Date(value));
}

function formatConversationPreview(value: string | null) {
  if (!value) {
    return "Nenhuma mensagem enviada ainda.";
  }

  const normalizedValue = value.trim();
  if (normalizedValue.length <= 96) {
    return normalizedValue;
  }

  return `${normalizedValue.slice(0, 93)}...`;
}

function toCreateNoticePayload(
  formState: CreateNoticeFormState,
): CreateNoticePayload {
  return {
    title: formState.title.trim(),
    body: formState.body.trim(),
    scopeType: formState.scopeType,
    priority: formState.priority,
    pinned: formState.pinned,
    expiresAt:
      formState.expiresAt.trim() === ""
        ? null
        : new Date(formState.expiresAt).toISOString(),
    recipientUserIds: formState.recipientUserIds,
  };
}

function toCreateConversationPayload(
  formState: CreateConversationFormState,
): CreateConversationPayload {
  return {
    participantUserId: formState.participantUserId ?? 0,
    projectId: formState.projectId,
    initialMessage: formState.initialMessage.trim(),
  };
}

function validateCreateNoticeForm(formState: CreateNoticeFormState) {
  if (formState.title.trim().length < 3) {
    return "Informe um título com pelo menos 3 caracteres.";
  }

  if (formState.body.trim().length < 3) {
    return "Informe uma mensagem com pelo menos 3 caracteres.";
  }

  if (
    formState.scopeType === "users" &&
    formState.recipientUserIds.length === 0
  ) {
    return "Selecione pelo menos um destinatário para o aviso direcionado.";
  }

  if (
    formState.expiresAt.trim() !== "" &&
    Number.isNaN(new Date(formState.expiresAt).getTime())
  ) {
    return "Informe uma data de expiração válida.";
  }

  return null;
}

function validateCreateConversationForm(
  formState: CreateConversationFormState,
) {
  if (!formState.participantUserId) {
    return "Selecione um destinatário para iniciar a conversa.";
  }

  if (formState.initialMessage.trim().length === 0) {
    return "Informe a primeira mensagem da conversa.";
  }

  return null;
}

export function CommunicationPage() {
  const queryClient = useQueryClient();
  const { user } = useAuth();
  const [searchParams, setSearchParams] = useSearchParams();
  const contextualProjectIdParam = searchParams.get("projectId");
  const contextualProjectId = contextualProjectIdParam
    ? Number(contextualProjectIdParam)
    : null;
  const hasContextualProject =
    contextualProjectId !== null &&
    Number.isInteger(contextualProjectId) &&
    contextualProjectId > 0;
  const requestedTab = searchParams.get("tab");
  const autoMarkedConversationKeyRef = useRef<string | null>(null);
  const autoMarkedNoticeIdsRef = useRef<Set<number>>(new Set());
  const activeTab: CommunicationTab =
    requestedTab === "conversations" ? "conversations" : "notices";
  const [statusFilter, setStatusFilter] = useState<NoticeStatusFilter>("all");
  const [activeConversationId, setActiveConversationId] = useState<
    number | null
  >(null);
  const [feedbackMessage, setFeedbackMessage] = useState<string | null>(null);
  const [feedbackError, setFeedbackError] = useState<string | null>(null);
  const [createNoticeDialogOpen, setCreateNoticeDialogOpen] = useState(false);
  const [createConversationDialogOpen, setCreateConversationDialogOpen] =
    useState(false);
  const [createFormState, setCreateFormState] = useState<CreateNoticeFormState>(
    defaultCreateNoticeFormState,
  );
  const [createConversationFormState, setCreateConversationFormState] =
    useState<CreateConversationFormState>(defaultCreateConversationFormState);
  const [conversationMessageDraft, setConversationMessageDraft] = useState("");

  const noticesQuery = useQuery({
    queryKey: communicationQueryKeys.notices(statusFilter),
    queryFn: async () => listNoticesRequest(statusFilter),
  });

  const noticeUnreadCountQuery = useQuery({
    queryKey: communicationQueryKeys.noticeUnreadCount(),
    queryFn: getNoticeUnreadCountRequest,
    enabled: Boolean(user),
  });

  const conversationsQuery = useQuery({
    queryKey: communicationQueryKeys.conversations(),
    queryFn: listConversationsRequest,
    enabled: Boolean(user),
    refetchInterval:
      user && activeTab === "conversations"
        ? communicationPollingIntervals.conversationsMs
        : false,
    refetchIntervalInBackground: true,
  });

  const conversationUnreadCountQuery = useQuery({
    queryKey: communicationQueryKeys.conversationUnreadCount(),
    queryFn: getConversationUnreadCountRequest,
    enabled: Boolean(user),
    refetchInterval: user ? communicationPollingIntervals.unreadCountMs : false,
    refetchIntervalInBackground: true,
  });

  const availableUsersQuery = useQuery({
    queryKey: communicationQueryKeys.availableConversationUsers(),
    queryFn: listConversationAvailableUsersRequest,
    enabled:
      Boolean(user) &&
      (activeTab === "conversations" ||
        createConversationDialogOpen ||
        createNoticeDialogOpen),
  });

  const contextualProjectQuery = useQuery({
    queryKey: ["communication-context-project", contextualProjectId],
    queryFn: async () => getProjectByIdRequest(contextualProjectId ?? 0),
    enabled: hasContextualProject,
  });

  const visibleConversations = useMemo(() => {
    const allConversations = conversationsQuery.data ?? [];
    if (!hasContextualProject) {
      return allConversations;
    }

    return allConversations.filter(
      (conversation) => conversation.project?.id === contextualProjectId,
    );
  }, [contextualProjectId, conversationsQuery.data, hasContextualProject]);

  const resolvedActiveConversationId = useMemo(() => {
    if (visibleConversations.length === 0) {
      return null;
    }

    const hasSelectedConversation = visibleConversations.some(
      (conversation) => conversation.id === activeConversationId,
    );

    if (hasSelectedConversation) {
      return activeConversationId;
    }

    return visibleConversations[0]?.id ?? null;
  }, [activeConversationId, visibleConversations]);
  const activeConversation = useMemo(
    () =>
      visibleConversations.find(
        (conversation) => conversation.id === resolvedActiveConversationId,
      ) ?? null,
    [resolvedActiveConversationId, visibleConversations],
  );

  const conversationMessagesQuery = useQuery({
    queryKey: communicationQueryKeys.conversationMessages(
      resolvedActiveConversationId,
    ),
    queryFn: async () =>
      listConversationMessagesRequest(resolvedActiveConversationId ?? 0),
    enabled: resolvedActiveConversationId !== null,
    refetchInterval:
      activeTab === "conversations" && resolvedActiveConversationId !== null
        ? communicationPollingIntervals.messagesMs
        : false,
    refetchIntervalInBackground: true,
  });

  const createNoticeMutation = useMutation({
    mutationFn: async (payload: CreateNoticePayload) =>
      createNoticeRequest(payload),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["notices"] });
      await queryClient.invalidateQueries({
        queryKey: communicationQueryKeys.noticeUnreadCount(),
      });
      setFeedbackError(null);
      setFeedbackMessage("Aviso publicado com sucesso.");
      setCreateNoticeDialogOpen(false);
      setCreateFormState(defaultCreateNoticeFormState);
    },
    onError: (error) => {
      setFeedbackMessage(null);
      setFeedbackError(
        getErrorMessage(error, "Não foi possível publicar o aviso."),
      );
    },
  });

  const markReadMutation = useMutation({
    mutationFn: async ({
      noticeId,
      silent,
    }: {
      noticeId: number;
      silent?: boolean;
    }) => {
      await markNoticeReadRequest(noticeId);

      return {
        noticeId,
        silent: Boolean(silent),
      };
    },
    onSuccess: async ({ silent }) => {
      await queryClient.invalidateQueries({ queryKey: ["notices"] });
      await queryClient.invalidateQueries({
        queryKey: communicationQueryKeys.noticeUnreadCount(),
      });
      if (!silent) {
        setFeedbackError(null);
        setFeedbackMessage("Aviso marcado como lido.");
      }
    },
    onError: (error, variables) => {
      autoMarkedNoticeIdsRef.current.delete(variables.noticeId);
      if (!variables.silent) {
        setFeedbackMessage(null);
        setFeedbackError(
          getErrorMessage(error, "Não foi possível atualizar o aviso."),
        );
      }
    },
  });

  const createConversationMutation = useMutation({
    mutationFn: async (payload: CreateConversationPayload) =>
      createConversationRequest(payload),
    onSuccess: async (response) => {
      await queryClient.invalidateQueries({
        queryKey: communicationQueryKeys.conversations(),
      });
      await queryClient.invalidateQueries({
        queryKey: communicationQueryKeys.conversationUnreadCount(),
      });
      const nextParams = new URLSearchParams(searchParams);
      nextParams.set("tab", "conversations");
      setSearchParams(nextParams, { replace: true });
      setActiveConversationId(response.id);
      setCreateConversationDialogOpen(false);
      setCreateConversationFormState(defaultCreateConversationFormState);
      setFeedbackError(null);
      setFeedbackMessage("Conversa iniciada com sucesso.");
    },
    onError: (error) => {
      setFeedbackMessage(null);
      setFeedbackError(
        getErrorMessage(error, "Não foi possível iniciar a conversa."),
      );
    },
  });

  const sendConversationMessageMutation = useMutation({
    mutationFn: async (payload: SendConversationMessagePayload) => {
      if (resolvedActiveConversationId === null) {
        throw new Error("Conversa não selecionada");
      }

      return sendConversationMessageRequest(
        resolvedActiveConversationId,
        payload,
      );
    },
    onSuccess: async () => {
      if (resolvedActiveConversationId !== null) {
        await queryClient.invalidateQueries({
          queryKey: communicationQueryKeys.conversationMessages(
            resolvedActiveConversationId,
          ),
        });
      }
      await queryClient.invalidateQueries({
        queryKey: communicationQueryKeys.conversations(),
      });
      await queryClient.invalidateQueries({
        queryKey: communicationQueryKeys.conversationUnreadCount(),
      });
      setConversationMessageDraft("");
      setFeedbackError(null);
      setFeedbackMessage("Mensagem enviada com sucesso.");
    },
    onError: (error) => {
      setFeedbackMessage(null);
      setFeedbackError(
        getErrorMessage(error, "Não foi possível enviar a mensagem."),
      );
    },
  });

  const markConversationReadMutation = useMutation({
    mutationFn: async (conversationId: number) =>
      markConversationReadRequest(conversationId),
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: communicationQueryKeys.conversations(),
      });
      await queryClient.invalidateQueries({
        queryKey: communicationQueryKeys.conversationUnreadCount(),
      });
    },
  });

  const availableRecipientOptions = useMemo(
    () => availableUsersQuery.data ?? [],
    [availableUsersQuery.data],
  );

  const selectedRecipientOptions = useMemo(
    () =>
      availableRecipientOptions.filter((item) =>
        createFormState.recipientUserIds.includes(item.id),
      ),
    [availableRecipientOptions, createFormState.recipientUserIds],
  );

  const selectedConversationRecipient = useMemo(
    () =>
      availableRecipientOptions.find(
        (item) => item.id === createConversationFormState.participantUserId,
      ) ?? null,
    [availableRecipientOptions, createConversationFormState.participantUserId],
  );

  const unreadNoticesCount = noticeUnreadCountQuery.data ?? 0;
  const unreadConversationsCount = conversationUnreadCountQuery.data ?? 0;
  const contextualProject = contextualProjectQuery.data ?? null;

  useEffect(() => {
    if (activeTab !== "notices" || !noticesQuery.isSuccess) {
      return;
    }

    const unreadNoticeIds = (noticesQuery.data ?? [])
      .filter((notice) => notice.readAt === null)
      .map((notice) => notice.id);

    unreadNoticeIds.forEach((noticeId) => {
      if (autoMarkedNoticeIdsRef.current.has(noticeId)) {
        return;
      }

      autoMarkedNoticeIdsRef.current.add(noticeId);
      markReadMutation.mutate({
        noticeId,
        silent: true,
      });
    });
  }, [activeTab, markReadMutation, noticesQuery.data, noticesQuery.isSuccess]);

  useEffect(() => {
    if (
      activeConversation === null ||
      activeConversation.unreadCount === 0 ||
      !conversationMessagesQuery.isSuccess
    ) {
      return;
    }

    const currentConversationKey = `${activeConversation.id}:${activeConversation.lastMessageId ?? 0}`;
    if (autoMarkedConversationKeyRef.current === currentConversationKey) {
      return;
    }

    autoMarkedConversationKeyRef.current = currentConversationKey;
    markConversationReadMutation.mutate(activeConversation.id);
  }, [
    activeConversation,
    conversationMessagesQuery.isSuccess,
    markConversationReadMutation,
  ]);

  const handleOpenCreateNoticeDialog = () => {
    setFeedbackError(null);
    setFeedbackMessage(null);
    setCreateNoticeDialogOpen(true);
  };

  const handleOpenCreateConversationDialog = () => {
    setFeedbackError(null);
    setFeedbackMessage(null);
    setCreateConversationFormState((currentState) => ({
      ...currentState,
      projectId: hasContextualProject ? contextualProjectId : null,
    }));
    setCreateConversationDialogOpen(true);
  };

  const handleCloseCreateNoticeDialog = () => {
    if (createNoticeMutation.isPending) {
      return;
    }

    setCreateNoticeDialogOpen(false);
    setCreateFormState(defaultCreateNoticeFormState);
  };

  const handleCloseCreateConversationDialog = () => {
    if (createConversationMutation.isPending) {
      return;
    }

    setCreateConversationDialogOpen(false);
    setCreateConversationFormState(defaultCreateConversationFormState);
  };

  const handleClearContextualProject = () => {
    const nextParams = new URLSearchParams(searchParams);
    nextParams.delete("projectId");
    nextParams.set("tab", "conversations");
    setSearchParams(nextParams, { replace: true });
  };

  const handleCreateFormChange = <Key extends keyof CreateNoticeFormState>(
    key: Key,
    value: CreateNoticeFormState[Key],
  ) => {
    setCreateFormState((currentState) => ({
      ...currentState,
      [key]: value,
    }));
  };

  const handleCreateConversationFormChange = <
    Key extends keyof CreateConversationFormState,
  >(
    key: Key,
    value: CreateConversationFormState[Key],
  ) => {
    setCreateConversationFormState((currentState) => ({
      ...currentState,
      [key]: value,
    }));
  };

  const handleSubmitCreateNotice = () => {
    const validationError = validateCreateNoticeForm(createFormState);
    if (validationError) {
      setFeedbackMessage(null);
      setFeedbackError(validationError);
      return;
    }

    setFeedbackError(null);
    setFeedbackMessage(null);
    createNoticeMutation.mutate(toCreateNoticePayload(createFormState));
  };

  const handleSubmitCreateConversation = () => {
    const validationError = validateCreateConversationForm(
      createConversationFormState,
    );
    if (validationError) {
      setFeedbackMessage(null);
      setFeedbackError(validationError);
      return;
    }

    setFeedbackError(null);
    setFeedbackMessage(null);
    createConversationMutation.mutate(
      toCreateConversationPayload(createConversationFormState),
    );
  };

  const handleSendConversationMessage = () => {
    if (resolvedActiveConversationId === null) {
      return;
    }

    if (conversationMessageDraft.trim().length === 0) {
      setFeedbackMessage(null);
      setFeedbackError("Informe uma mensagem antes de enviar.");
      return;
    }

    setFeedbackError(null);
    setFeedbackMessage(null);
    sendConversationMessageMutation.mutate({
      body: conversationMessageDraft.trim(),
    });
  };

  const currentAction =
    activeTab === "notices" ? (
      user?.role === "admin" ? (
        <Button
          onClick={handleOpenCreateNoticeDialog}
          startIcon={<AddRoundedIcon />}
          variant="contained"
        >
          Novo aviso
        </Button>
      ) : null
    ) : (
      <Button
        onClick={handleOpenCreateConversationDialog}
        startIcon={<AddRoundedIcon />}
        variant="contained"
      >
        {hasContextualProject ? "Nova conversa da obra" : "Nova conversa"}
      </Button>
    );

  return (
    <Box sx={{ display: "flex", flexDirection: "column", gap: 3 }}>
      <PageHeader
        title="Comunicação"
        description="Centralize avisos internos e conversas diretas com histórico entre os usuários do sistema."
        action={currentAction}
      />

      <Box
        sx={{
          display: "grid",
          gap: 2,
          gridTemplateColumns: {
            lg: "repeat(3, minmax(0, 1fr))",
            md: "repeat(2, minmax(0, 1fr))",
            xs: "minmax(0, 1fr)",
          },
        }}
      >
        <Box
          sx={{
            background: (theme) =>
              `linear-gradient(135deg, ${alpha(theme.palette.warning.main, 0.16)} 0%, ${alpha(theme.palette.error.main, 0.08)} 100%)`,
            border: "1px solid",
            borderColor: (theme) => alpha(theme.palette.warning.main, 0.24),
            borderRadius: 4,
            boxShadow: (theme) =>
              `0 16px 30px ${alpha(theme.palette.warning.main, 0.09)}`,
            p: 2.5,
          }}
        >
          <Box sx={{ display: "flex", flexDirection: "column", gap: 0.75 }}>
            <Typography
              sx={{ color: communicationBlue, fontWeight: 800 }}
              variant="body2"
            >
              Avisos monitorados
            </Typography>
            <Typography
              sx={{ color: communicationBlue, fontWeight: 850 }}
              variant="h4"
            >
              {noticesQuery.data?.length ?? 0}
            </Typography>
            <Typography color="text.secondary" variant="body2">
              {unreadNoticesCount > 0
                ? `${unreadNoticesCount} aviso(s) aguardando leitura.`
                : "Todos os avisos do recorte atual já foram lidos."}
            </Typography>
          </Box>
        </Box>
        <Box
          sx={{
            background: (theme) =>
              `linear-gradient(135deg, ${alpha(theme.palette.primary.main, 0.14)} 0%, ${alpha(theme.palette.info.main, 0.08)} 100%)`,
            border: "1px solid",
            borderColor: (theme) => alpha(theme.palette.primary.main, 0.2),
            borderRadius: 4,
            boxShadow: (theme) =>
              `0 16px 30px ${alpha(theme.palette.primary.main, 0.09)}`,
            p: 2.5,
          }}
        >
          <Box sx={{ display: "flex", flexDirection: "column", gap: 0.75 }}>
            <Typography
              sx={{ color: communicationBlue, fontWeight: 800 }}
              variant="body2"
            >
              Conversas ativas
            </Typography>
            <Typography
              sx={{ color: communicationBlue, fontWeight: 850 }}
              variant="h4"
            >
              {visibleConversations.length}
            </Typography>
            <Typography color="text.secondary" variant="body2">
              {unreadConversationsCount > 0
                ? `${unreadConversationsCount} conversa(s) com mensagem pendente.`
                : "Nenhuma conversa com pendência de leitura neste momento."}
            </Typography>
          </Box>
        </Box>
        <Box
          sx={{
            background: (theme) =>
              `linear-gradient(135deg, ${alpha(theme.palette.success.main, 0.13)} 0%, ${alpha(theme.palette.info.main, 0.06)} 100%)`,
            border: "1px solid",
            borderColor: (theme) => alpha(theme.palette.success.main, 0.18),
            borderRadius: 4,
            boxShadow: (theme) =>
              `0 16px 30px ${alpha(theme.palette.success.main, 0.08)}`,
            p: 2.5,
          }}
        >
          <Box sx={{ display: "flex", flexDirection: "column", gap: 0.75 }}>
            <Typography
              sx={{ color: communicationBlue, fontWeight: 800 }}
              variant="body2"
            >
              Contexto atual
            </Typography>
            <Typography
              sx={{ color: communicationBlue, fontWeight: 850 }}
              variant="h5"
            >
              {hasContextualProject && contextualProject
                ? `${contextualProject.code}`
                : "Central geral"}
            </Typography>
            <Typography color="text.secondary" variant="body2">
              {hasContextualProject && contextualProject
                ? `Mensagens filtradas para a obra ${contextualProject.name}.`
                : "Leitura consolidada de avisos e conversas da operação."}
            </Typography>
          </Box>
        </Box>
      </Box>

      {feedbackMessage ? (
        <Alert severity="success" sx={communicationFeedbackAlertSx}>
          {feedbackMessage}
        </Alert>
      ) : null}
      {feedbackError ? (
        <Alert severity="error" sx={communicationFeedbackAlertSx}>
          {feedbackError}
        </Alert>
      ) : null}

      <SectionCard>
        <Tabs
          sx={{
            backdropFilter: "blur(10px)",
            bgcolor: (theme) => alpha(theme.palette.primary.main, 0.05),
            border: "1px solid",
            borderColor: (theme) => alpha(theme.palette.primary.main, 0.1),
            borderRadius: 999,
            boxShadow: (theme) =>
              `0 14px 30px ${alpha(theme.palette.primary.main, 0.08)}`,
            p: 0.75,
            "& .MuiTabs-flexContainer": {
              gap: 1,
            },
            "& .MuiTabs-indicator": {
              display: "none",
            },
            "& .MuiTab-root": {
              color: "text.secondary",
              borderRadius: 999,
              fontSize: "1rem",
              fontWeight: 600,
              minHeight: 52,
              px: 2.5,
              textTransform: "none",
            },
            "& .MuiSvgIcon-root": {
              fontSize: 22,
            },
            "& .Mui-selected": {
              bgcolor: (theme) => alpha(theme.palette.primary.main, 0.12),
              boxShadow: (theme) =>
                `0 10px 20px ${alpha(theme.palette.primary.main, 0.12)}`,
              color: communicationBlue,
              fontWeight: 800,
            },
          }}
          onChange={(_, value: CommunicationTab) => {
            const nextParams = new URLSearchParams(searchParams);
            nextParams.set("tab", value);
            setSearchParams(nextParams, { replace: true });
          }}
          value={activeTab}
        >
          <Tab
            icon={
              unreadNoticesCount > 0 ? (
                <Badge
                  badgeContent={unreadNoticesCount}
                  color="error"
                  sx={{
                    "& .MuiBadge-badge": {
                      fontSize: "0.8rem",
                      fontWeight: 700,
                      minHeight: 22,
                      minWidth: 22,
                    },
                  }}
                >
                  <CampaignRoundedIcon fontSize="small" />
                </Badge>
              ) : (
                <CampaignRoundedIcon fontSize="small" />
              )
            }
            iconPosition="start"
            label="Avisos"
            value="notices"
          />
          <Tab
            icon={
              unreadConversationsCount > 0 ? (
                <Badge
                  badgeContent={unreadConversationsCount}
                  color="primary"
                  sx={{
                    "& .MuiBadge-badge": {
                      fontSize: "0.8rem",
                      fontWeight: 700,
                      minHeight: 22,
                      minWidth: 22,
                    },
                  }}
                >
                  <ForumRoundedIcon fontSize="small" />
                </Badge>
              ) : (
                <ForumRoundedIcon fontSize="small" />
              )
            }
            iconPosition="start"
            label="Conversas"
            value="conversations"
          />
        </Tabs>

        {activeTab === "notices" ? (
          <Box sx={{ display: "flex", flexDirection: "column", gap: 2.5 }}>
            <Box
              sx={{
                background: (theme) =>
                  `linear-gradient(135deg, ${alpha(theme.palette.warning.main, 0.18)} 0%, ${alpha(theme.palette.error.main, 0.1)} 100%)`,
                border: "1px solid",
                borderColor: (theme) => alpha(theme.palette.warning.main, 0.24),
                borderRadius: 3,
                display: "flex",
                gap: 2,
                justifyContent: "space-between",
                px: 3,
                py: 2.5,
                flexWrap: "wrap",
              }}
            >
              <Box sx={{ display: "flex", flexDirection: "column", gap: 0.5 }}>
                <Typography
                  sx={{
                    color: communicationBlue,
                    fontWeight: 800,
                    letterSpacing: 0.2,
                  }}
                  variant="h5"
                >
                  Central de avisos
                </Typography>
                <Typography
                  color="text.primary"
                  sx={{ lineHeight: 1.6 }}
                  variant="body1"
                >
                  Consulte os comunicados publicados e marque como lidos os
                  avisos recebidos.
                </Typography>
              </Box>
              <Box
                sx={{
                  alignItems: "center",
                  backgroundColor: (theme) =>
                    alpha(theme.palette.common.white, 0.44),
                  border: "1px solid",
                  borderColor: (theme) =>
                    alpha(theme.palette.warning.main, 0.18),
                  borderRadius: 999,
                  display: "flex",
                  gap: 1,
                  flexWrap: "wrap",
                  p: 0.75,
                }}
              >
                <Chip
                  color="warning"
                  icon={<CampaignRoundedIcon />}
                  label={`${noticesQuery.data?.length ?? 0} aviso(s)`}
                  size="medium"
                  sx={{ "& .MuiChip-label": { fontWeight: 700 } }}
                  variant="filled"
                />
                <Chip
                  color={unreadNoticesCount > 0 ? "error" : "default"}
                  icon={<NotificationsActiveRoundedIcon />}
                  label={`${unreadNoticesCount} não lido(s)`}
                  size="medium"
                  sx={{ "& .MuiChip-label": { fontWeight: 700 } }}
                  variant={unreadNoticesCount > 0 ? "filled" : "outlined"}
                />
              </Box>
            </Box>
            <Tabs
              sx={{
                "& .MuiTabs-flexContainer": {
                  gap: 1,
                },
                "& .MuiTabs-indicator": {
                  display: "none",
                },
                "& .MuiTab-root": {
                  borderRadius: 999,
                  fontSize: "0.95rem",
                  fontWeight: 600,
                  minHeight: 46,
                  px: 2.5,
                  textTransform: "none",
                },
                "& .Mui-selected": {
                  bgcolor: (theme) => alpha(theme.palette.warning.main, 0.14),
                  color: communicationBlue,
                  fontWeight: 800,
                },
              }}
              onChange={(_, value: NoticeStatusFilter) =>
                setStatusFilter(value)
              }
              value={statusFilter}
            >
              <Tab label="Todos" value="all" />
              <Tab
                label={`Não lidos${unreadNoticesCount > 0 ? ` (${unreadNoticesCount})` : ""}`}
                value="unread"
              />
              <Tab label="Lidos" value="read" />
            </Tabs>

            {noticesQuery.isLoading ? (
              <LinearProgress sx={communicationLoaderSx} />
            ) : null}

            {!noticesQuery.isLoading && noticesQuery.data?.length === 0 ? (
              <Alert severity="info" sx={communicationInfoAlertSx}>
                Nenhum aviso encontrado no momento.
              </Alert>
            ) : null}

            <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
              {(noticesQuery.data ?? []).map((notice) => (
                <NoticeCard
                  key={notice.id}
                  notice={notice}
                  onMarkAsRead={() =>
                    markReadMutation.mutate({
                      noticeId: notice.id,
                    })
                  }
                  readActionDisabled={markReadMutation.isPending}
                />
              ))}
            </Box>
          </Box>
        ) : (
          <Box sx={{ display: "flex", flexDirection: "column", gap: 2.5 }}>
            <Box
              sx={{
                background: (theme) =>
                  `linear-gradient(135deg, ${alpha(theme.palette.primary.main, 0.16)} 0%, ${alpha(theme.palette.info.main, 0.1)} 100%)`,
                border: "1px solid",
                borderColor: (theme) => alpha(theme.palette.primary.main, 0.22),
                borderRadius: 3,
                display: "flex",
                gap: 2,
                justifyContent: "space-between",
                px: 3,
                py: 2.5,
                flexWrap: "wrap",
              }}
            >
              <Box sx={{ display: "flex", flexDirection: "column", gap: 0.5 }}>
                <Typography
                  sx={{
                    color: communicationBlue,
                    fontWeight: 800,
                    letterSpacing: 0.2,
                  }}
                  variant="h5"
                >
                  Central de conversas
                </Typography>
                <Typography
                  color="text.primary"
                  sx={{ lineHeight: 1.6 }}
                  variant="body1"
                >
                  Abra conversas diretas, acompanhe o histórico de mensagens e
                  responda sem sair da central de comunicação.
                </Typography>
              </Box>
              <Box
                sx={{
                  alignItems: "center",
                  backgroundColor: (theme) =>
                    alpha(theme.palette.common.white, 0.42),
                  border: "1px solid",
                  borderColor: (theme) =>
                    alpha(theme.palette.primary.main, 0.16),
                  borderRadius: 999,
                  display: "flex",
                  gap: 1,
                  flexWrap: "wrap",
                  p: 0.75,
                }}
              >
                <Chip
                  color="info"
                  icon={<ForumRoundedIcon />}
                  label={`${visibleConversations.length} conversa(s)`}
                  size="medium"
                  sx={{ "& .MuiChip-label": { fontWeight: 700 } }}
                  variant="filled"
                />
                <Chip
                  color={unreadConversationsCount > 0 ? "primary" : "default"}
                  icon={<MarkEmailUnreadRoundedIcon />}
                  label={`${unreadConversationsCount} não lida(s)`}
                  size="medium"
                  sx={{ "& .MuiChip-label": { fontWeight: 700 } }}
                  variant={unreadConversationsCount > 0 ? "filled" : "outlined"}
                />
              </Box>
            </Box>

            {hasContextualProject ? (
              contextualProject ? (
                <Alert
                  action={
                    <Button onClick={handleClearContextualProject} size="small">
                      Ver todas
                    </Button>
                  }
                  severity="info"
                  sx={communicationInfoAlertSx}
                >
                  Conversas filtradas pela obra {contextualProject.code} -{" "}
                  {contextualProject.name}.
                </Alert>
              ) : contextualProjectQuery.isError ? (
                <Alert severity="warning" sx={communicationInfoAlertSx}>
                  Não foi possível carregar o contexto da obra solicitado.
                </Alert>
              ) : null
            ) : null}

            {(conversationsQuery.isLoading ||
              conversationMessagesQuery.isLoading) && (
              <LinearProgress sx={communicationLoaderSx} />
            )}

            {!conversationsQuery.isLoading &&
            visibleConversations.length === 0 ? (
              <Alert severity="info" sx={communicationInfoAlertSx}>
                {hasContextualProject
                  ? "Nenhuma conversa encontrada para esta obra."
                  : "Nenhuma conversa iniciada ainda."}
              </Alert>
            ) : null}

            <Box
              sx={{
                display: "grid",
                gap: 2,
                gridTemplateColumns: {
                  lg: "320px minmax(0, 1fr)",
                  xs: "1fr",
                },
                minHeight: 460,
              }}
            >
              <Box
                sx={{
                  background: (theme) =>
                    `linear-gradient(180deg, ${alpha(theme.palette.info.main, 0.05)} 0%, ${alpha(theme.palette.primary.main, 0.02)} 100%)`,
                  border: "1px solid",
                  borderColor: (theme) => alpha(theme.palette.info.main, 0.18),
                  borderRadius: 3,
                  boxShadow: (theme) =>
                    `0 16px 32px ${alpha(theme.palette.info.main, 0.08)}`,
                  display: "flex",
                  flexDirection: "column",
                  overflow: "hidden",
                }}
              >
                <Box
                  sx={{
                    alignItems: "center",
                    bgcolor: (theme) =>
                      theme.palette.mode === "dark"
                        ? alpha(theme.palette.info.light, 0.12)
                        : alpha(theme.palette.info.main, 0.08),
                    display: "flex",
                    justifyContent: "space-between",
                    px: 2.25,
                    py: 1.75,
                  }}
                >
                  <Typography sx={{ fontWeight: 700 }} variant="h6">
                    Recentes
                  </Typography>
                  {unreadConversationsCount > 0 ? (
                    <Chip
                      color="primary"
                      label={`${unreadConversationsCount} não lida(s)`}
                      size="medium"
                      sx={{ "& .MuiChip-label": { fontWeight: 700 } }}
                    />
                  ) : null}
                </Box>
                <Divider />
                <Box sx={{ display: "flex", flexDirection: "column" }}>
                  {visibleConversations.map((conversation) => (
                    <ConversationListItemCard
                      key={conversation.id}
                      conversation={conversation}
                      isSelected={
                        conversation.id === resolvedActiveConversationId
                      }
                      onClick={() => setActiveConversationId(conversation.id)}
                    />
                  ))}
                </Box>
              </Box>

              <Box
                sx={{
                  background: (theme) =>
                    `linear-gradient(180deg, ${alpha(theme.palette.primary.main, 0.045)} 0%, ${alpha(theme.palette.info.main, 0.02)} 100%)`,
                  border: "1px solid",
                  borderColor: (theme) =>
                    alpha(theme.palette.primary.main, 0.18),
                  borderRadius: 3,
                  boxShadow: (theme) =>
                    `0 18px 34px ${alpha(theme.palette.primary.main, 0.08)}`,
                  display: "flex",
                  flexDirection: "column",
                  minHeight: 460,
                  overflow: "hidden",
                }}
              >
                {activeConversation ? (
                  <>
                    <Box
                      sx={{
                        alignItems: "center",
                        bgcolor: (theme) =>
                          theme.palette.mode === "dark"
                            ? alpha(theme.palette.primary.light, 0.14)
                            : alpha(theme.palette.primary.main, 0.08),
                        display: "flex",
                        justifyContent: "space-between",
                        px: 2.5,
                        py: 2,
                      }}
                    >
                      <Box
                        sx={{
                          display: "flex",
                          flexDirection: "column",
                          gap: 0.5,
                        }}
                      >
                        <Typography
                          sx={{ color: communicationBlue, fontWeight: 800 }}
                          variant="h6"
                        >
                          {activeConversation.participant.name}
                        </Typography>
                        <Box sx={{ display: "flex", flexWrap: "wrap", gap: 1 }}>
                          <Chip
                            label={`@${activeConversation.participant.username}`}
                            size="small"
                          />
                          <Chip
                            label={getRoleLabel(
                              activeConversation.participant.role,
                            )}
                            size="small"
                          />
                          {activeConversation.project ? (
                            <Chip
                              label={`Obra ${activeConversation.project.code}`}
                              size="small"
                              variant="outlined"
                            />
                          ) : null}
                        </Box>
                      </Box>
                      <Typography color="text.secondary" variant="body2">
                        Atualizada em{" "}
                        {formatDateTime(
                          activeConversation.lastMessageAt ??
                            activeConversation.updatedAt,
                        )}
                      </Typography>
                    </Box>
                    <Divider />
                    <Box
                      sx={{
                        background:
                          "linear-gradient(180deg, rgba(37, 99, 235, 0.04) 0%, rgba(14, 165, 233, 0.02) 100%)",
                        display: "flex",
                        flex: 1,
                        flexDirection: "column",
                        gap: 1.5,
                        maxHeight: 420,
                        overflowY: "auto",
                        p: 2.5,
                      }}
                    >
                      {(conversationMessagesQuery.data ?? []).map((message) => (
                        <ConversationMessageBubble
                          key={message.id}
                          isOwnMessage={message.sender.id === user?.id}
                          message={message}
                        />
                      ))}
                      {conversationMessagesQuery.isSuccess &&
                      (conversationMessagesQuery.data?.length ?? 0) === 0 ? (
                        <Alert severity="info" sx={communicationInfoAlertSx}>
                          Nenhuma mensagem registrada nesta conversa.
                        </Alert>
                      ) : null}
                    </Box>
                    <Divider />
                    <Box
                      sx={{
                        alignItems: "flex-end",
                        bgcolor: (theme) =>
                          theme.palette.mode === "dark"
                            ? alpha(theme.palette.primary.light, 0.08)
                            : alpha(theme.palette.primary.main, 0.04),
                        display: "flex",
                        gap: 1.5,
                        p: 2.25,
                      }}
                    >
                      <Box
                        sx={{
                          display: "flex",
                          flex: 1,
                          flexDirection: "column",
                          gap: 0.75,
                        }}
                      >
                        <Typography
                          sx={{
                            color: communicationBlue,
                            fontSize: "0.82rem",
                            fontWeight: 700,
                          }}
                        >
                          Mensagem
                        </Typography>
                        <TextField
                          fullWidth
                          minRows={3}
                          multiline
                          onChange={(event) =>
                            setConversationMessageDraft(event.target.value)
                          }
                          placeholder="Escreva uma resposta clara e objetiva."
                          value={conversationMessageDraft}
                        />
                      </Box>
                      <Button
                        disabled={sendConversationMessageMutation.isPending}
                        onClick={handleSendConversationMessage}
                        startIcon={<SendRoundedIcon />}
                        variant="contained"
                      >
                        Enviar
                      </Button>
                    </Box>
                  </>
                ) : (
                  <Box
                    sx={{
                      alignItems: "center",
                      display: "flex",
                      flex: 1,
                      justifyContent: "center",
                      p: 3,
                    }}
                  >
                    <Alert severity="info" sx={communicationInfoAlertSx}>
                      Selecione uma conversa para visualizar o histórico.
                    </Alert>
                  </Box>
                )}
              </Box>
            </Box>
          </Box>
        )}
      </SectionCard>

      <Dialog
        fullWidth
        maxWidth="md"
        onClose={handleCloseCreateNoticeDialog}
        open={createNoticeDialogOpen}
      >
        <DialogTitle>Novo aviso</DialogTitle>
        <DialogContent
          sx={{ display: "flex", flexDirection: "column", gap: 2.5, pt: 2 }}
        >
          <TextField
            label="Título"
            onChange={(event) =>
              handleCreateFormChange("title", event.target.value)
            }
            value={createFormState.title}
          />
          <TextField
            label="Mensagem"
            minRows={5}
            multiline
            onChange={(event) =>
              handleCreateFormChange("body", event.target.value)
            }
            value={createFormState.body}
          />
          <Box
            sx={{
              display: "grid",
              gap: 2,
              gridTemplateColumns: {
                md: "repeat(2, minmax(0, 1fr))",
                xs: "1fr",
              },
            }}
          >
            <TextField
              label="Tipo de destinatário"
              onChange={(event) =>
                handleCreateFormChange(
                  "scopeType",
                  event.target.value as NoticeScopeType,
                )
              }
              select
              value={createFormState.scopeType}
            >
              <MenuItem value="all">Geral</MenuItem>
              <MenuItem value="users">Direcionado</MenuItem>
            </TextField>
            <TextField
              label="Prioridade"
              onChange={(event) =>
                handleCreateFormChange(
                  "priority",
                  event.target.value as NoticePriority,
                )
              }
              select
              value={createFormState.priority}
            >
              <MenuItem value="info">Informativo</MenuItem>
              <MenuItem value="warning">Atenção</MenuItem>
              <MenuItem value="critical">Crítico</MenuItem>
            </TextField>
          </Box>

          {createFormState.scopeType === "users" ? (
            <Autocomplete<ConversationParticipantItem, true, false, false>
              getOptionLabel={(option) => option.name}
              onChange={(_, value) =>
                handleCreateFormChange(
                  "recipientUserIds",
                  value.map((item) => item.id),
                )
              }
              options={availableRecipientOptions}
              renderInput={(params) => (
                <TextField
                  {...params}
                  label="Destinatários"
                  placeholder="Selecione um ou mais usuários"
                />
              )}
              value={selectedRecipientOptions}
            />
          ) : null}

          <TextField
            label="Expira em"
            onChange={(event) =>
              handleCreateFormChange("expiresAt", event.target.value)
            }
            slotProps={{ inputLabel: { shrink: true } }}
            type="datetime-local"
            value={createFormState.expiresAt}
          />

          <FormControlLabel
            control={
              <Checkbox
                checked={createFormState.pinned}
                onChange={(event) =>
                  handleCreateFormChange("pinned", event.target.checked)
                }
              />
            }
            label="Fixar este aviso no topo da lista"
          />
        </DialogContent>
        <DialogActions sx={{ px: 3, pb: 3 }}>
          <Button
            disabled={createNoticeMutation.isPending}
            onClick={handleCloseCreateNoticeDialog}
          >
            Cancelar
          </Button>
          <Button
            disabled={createNoticeMutation.isPending}
            onClick={handleSubmitCreateNotice}
            variant="contained"
          >
            Publicar aviso
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog
        fullWidth
        maxWidth="sm"
        onClose={handleCloseCreateConversationDialog}
        open={createConversationDialogOpen}
      >
        <DialogTitle>Nova conversa</DialogTitle>
        <DialogContent
          sx={{ display: "flex", flexDirection: "column", gap: 2.5, pt: 2 }}
        >
          <Autocomplete<ConversationParticipantItem, false, false, false>
            getOptionLabel={(option) => `${option.name} (@${option.username})`}
            loading={availableUsersQuery.isLoading}
            onChange={(_, value) =>
              handleCreateConversationFormChange(
                "participantUserId",
                value?.id ?? null,
              )
            }
            options={availableRecipientOptions}
            renderInput={(params) => (
              <TextField
                {...params}
                label="Destinatário"
                placeholder="Selecione um usuário"
              />
            )}
            value={selectedConversationRecipient}
          />
          {createConversationFormState.projectId && contextualProject ? (
            <Alert severity="info" sx={communicationInfoAlertSx}>
              Esta conversa será vinculada à obra {contextualProject.code} -{" "}
              {contextualProject.name}.
            </Alert>
          ) : null}
          {availableUsersQuery.isSuccess &&
          availableRecipientOptions.length === 0 ? (
            <Alert severity="info" sx={communicationInfoAlertSx}>
              Nenhum destinatário disponível para iniciar conversa no momento.
            </Alert>
          ) : null}
          <TextField
            label="Primeira mensagem"
            minRows={5}
            multiline
            onChange={(event) =>
              handleCreateConversationFormChange(
                "initialMessage",
                event.target.value,
              )
            }
            value={createConversationFormState.initialMessage}
          />
        </DialogContent>
        <DialogActions sx={{ px: 3, pb: 3 }}>
          <Button
            disabled={createConversationMutation.isPending}
            onClick={handleCloseCreateConversationDialog}
          >
            Cancelar
          </Button>
          <Button
            disabled={createConversationMutation.isPending}
            onClick={handleSubmitCreateConversation}
            variant="contained"
          >
            Iniciar conversa
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}

type NoticeCardProps = {
  notice: NoticeItem;
  onMarkAsRead: () => void;
  readActionDisabled: boolean;
};

function NoticeCard({
  notice,
  onMarkAsRead,
  readActionDisabled,
}: NoticeCardProps) {
  const accentColor = getPriorityAccentColor(notice.priority);

  return (
    <SectionCard
      sx={{
        background: `linear-gradient(135deg, ${accentColor}12 0%, transparent 38%)`,
        border: `1px solid ${accentColor}2E`,
        boxShadow: `0 10px 24px ${accentColor}12`,
      }}
    >
      <Box sx={{ display: "flex", flexDirection: "column", gap: 1.5 }}>
        <Box
          sx={{
            alignItems: { md: "center", xs: "flex-start" },
            display: "flex",
            flexDirection: { md: "row", xs: "column" },
            gap: 1,
            justifyContent: "space-between",
          }}
        >
          <Box sx={{ display: "flex", flexWrap: "wrap", gap: 1 }}>
            <Chip
              color={getPriorityColor(notice.priority)}
              label={getPriorityLabel(notice.priority)}
              size="medium"
              sx={{ "& .MuiChip-label": { fontWeight: 700 } }}
            />
            <Chip
              label={getScopeLabel(notice.scopeType)}
              size="medium"
              sx={{ "& .MuiChip-label": { fontWeight: 600 } }}
            />
            <Chip
              color={notice.readAt ? "default" : "primary"}
              label={notice.readAt ? "Lido" : "Não lido"}
              size="medium"
              sx={
                notice.readAt
                  ? { "& .MuiChip-label": { fontWeight: 600 } }
                  : {
                      "& .MuiChip-label": {
                        fontWeight: 700,
                      },
                      bgcolor: `${accentColor}18`,
                      color: accentColor,
                    }
              }
            />
            {notice.pinned ? (
              <Chip
                color="warning"
                icon={<PushPinRoundedIcon />}
                label="Fixado"
                size="medium"
                sx={{ "& .MuiChip-label": { fontWeight: 700 } }}
              />
            ) : null}
          </Box>
          {!notice.readAt ? (
            <Button
              disabled={readActionDisabled}
              onClick={onMarkAsRead}
              size="medium"
              startIcon={<MarkEmailReadRoundedIcon />}
              sx={{ fontWeight: 700 }}
              variant="outlined"
            >
              Marcar como lido
            </Button>
          ) : null}
        </Box>

        <Box
          sx={{
            alignItems: "center",
            display: "flex",
            gap: 1,
          }}
        >
          <Box
            sx={{
              alignItems: "center",
              bgcolor: `${accentColor}22`,
              borderRadius: 2,
              color: accentColor,
              display: "inline-flex",
              justifyContent: "center",
              p: 1,
            }}
          >
            <CampaignRoundedIcon fontSize="small" />
          </Box>
          <Typography sx={{ color: accentColor, fontWeight: 700 }} variant="h5">
            {notice.title}
          </Typography>
        </Box>

        <Typography
          sx={{ lineHeight: 1.75, whiteSpace: "pre-wrap" }}
          variant="body1"
        >
          {notice.body}
        </Typography>

        <Box
          sx={{
            color: "text.secondary",
            display: "grid",
            gap: 0.75,
            gridTemplateColumns: { md: "repeat(2, minmax(0, 1fr))", xs: "1fr" },
          }}
        >
          <Typography variant="body2">
            <strong>Publicado por:</strong> {notice.createdByUserName}
          </Typography>
          <Typography variant="body2">
            <strong>Enviado em:</strong> {formatDateTime(notice.createdAt)}
          </Typography>
          <Typography variant="body2">
            <strong>Lido em:</strong> {formatDateTime(notice.readAt)}
          </Typography>
          <Typography variant="body2">
            <strong>Expira em:</strong> {formatDateTime(notice.expiresAt)}
          </Typography>
        </Box>
      </Box>
    </SectionCard>
  );
}

type ConversationListItemCardProps = {
  conversation: ConversationItem;
  isSelected: boolean;
  onClick: () => void;
};

function ConversationListItemCard({
  conversation,
  isSelected,
  onClick,
}: ConversationListItemCardProps) {
  return (
    <Box
      onClick={onClick}
      sx={{
        bgcolor: (theme) =>
          isSelected
            ? theme.palette.mode === "dark"
              ? "rgba(59, 130, 246, 0.22)"
              : "rgba(37, 99, 235, 0.12)"
            : conversation.unreadCount > 0
              ? theme.palette.mode === "dark"
                ? alpha(theme.palette.info.light, 0.08)
                : alpha(theme.palette.info.main, 0.05)
              : "transparent",
        borderBottom: "1px solid",
        borderColor: (theme) =>
          isSelected
            ? theme.palette.mode === "dark"
              ? "rgba(96, 165, 250, 0.42)"
              : "rgba(37, 99, 235, 0.2)"
            : theme.palette.divider,
        borderLeft: "6px solid",
        borderLeftColor: (theme) =>
          isSelected
            ? theme.palette.mode === "dark"
              ? theme.palette.info.light
              : theme.palette.info.main
            : conversation.unreadCount > 0
              ? theme.palette.mode === "dark"
                ? alpha(theme.palette.info.light, 0.55)
                : alpha(theme.palette.info.main, 0.4)
              : "transparent",
        boxShadow: (theme) =>
          isSelected
            ? theme.palette.mode === "dark"
              ? "inset 0 0 0 1px rgba(96, 165, 250, 0.18), 0 6px 18px rgba(15, 23, 42, 0.18)"
              : "inset 0 0 0 1px rgba(37, 99, 235, 0.08), 0 4px 14px rgba(37, 99, 235, 0.08)"
            : "none",
        cursor: "pointer",
        minHeight: 104,
        px: 2.5,
        py: 1.8,
        transition:
          "background-color 0.2s ease, border-color 0.2s ease, box-shadow 0.2s ease",
        "&:hover": {
          bgcolor: (theme) =>
            isSelected
              ? theme.palette.mode === "dark"
                ? "rgba(59, 130, 246, 0.24)"
                : "rgba(37, 99, 235, 0.14)"
              : theme.palette.mode === "dark"
                ? alpha(theme.palette.info.light, 0.08)
                : alpha(theme.palette.info.main, 0.05),
        },
      }}
    >
      <Box
        sx={{
          alignItems: "center",
          display: "flex",
          justifyContent: "space-between",
          mb: 0.5,
        }}
      >
        <Typography
          color={isSelected ? "primary.main" : "text.primary"}
          noWrap
          sx={{
            fontSize: "1rem",
            fontWeight: isSelected || conversation.unreadCount > 0 ? 700 : 600,
          }}
          variant="subtitle2"
        >
          {conversation.participant.name}
        </Typography>
        {conversation.unreadCount > 0 ? (
          <Badge
            badgeContent={conversation.unreadCount}
            color="error"
            sx={{
              "& .MuiBadge-badge": {
                boxShadow: "0 0 0 2px rgba(255,255,255,0.9)",
                fontSize: "0.82rem",
                fontWeight: 700,
                minHeight: 22,
                minWidth: 22,
              },
            }}
          />
        ) : null}
      </Box>
      <Typography
        color={isSelected ? "text.primary" : "text.secondary"}
        noWrap
        sx={{
          fontSize: "0.98rem",
          lineHeight: 1.55,
          fontWeight: conversation.unreadCount > 0 ? 600 : 400,
        }}
        variant="body2"
      >
        {formatConversationPreview(conversation.lastMessageBody)}
      </Typography>
      <Box
        sx={{
          alignItems: "center",
          display: "flex",
          gap: 1,
          justifyContent: "space-between",
          mt: 1,
        }}
      >
        <Box sx={{ display: "flex", flexWrap: "wrap", gap: 0.75 }}>
          <Chip
            label={getRoleLabel(conversation.participant.role)}
            size="medium"
            sx={{ "& .MuiChip-label": { fontWeight: 600 } }}
          />
          {conversation.project ? (
            <Chip
              label={`Obra ${conversation.project.code}`}
              size="medium"
              sx={(theme) => ({
                "& .MuiChip-label": {
                  fontWeight: 700,
                },
                ...(isSelected
                  ? {
                      bgcolor:
                        theme.palette.mode === "dark"
                          ? "rgba(59, 130, 246, 0.18)"
                          : "rgba(37, 99, 235, 0.12)",
                      borderColor:
                        theme.palette.mode === "dark"
                          ? "rgba(96, 165, 250, 0.5)"
                          : "rgba(37, 99, 235, 0.28)",
                      color:
                        theme.palette.mode === "dark"
                          ? theme.palette.info.light
                          : theme.palette.info.dark,
                      fontWeight: 600,
                    }
                  : {}),
              })}
              variant="outlined"
            />
          ) : null}
        </Box>
        <Typography
          color={isSelected ? "primary.main" : "text.secondary"}
          sx={{
            fontSize: "0.82rem",
            fontWeight: conversation.unreadCount > 0 ? 700 : 500,
          }}
          variant="caption"
        >
          {formatDateTime(conversation.lastMessageAt ?? conversation.updatedAt)}
        </Typography>
      </Box>
    </Box>
  );
}

type ConversationMessageBubbleProps = {
  message: ConversationMessageItem;
  isOwnMessage: boolean;
};

function ConversationMessageBubble({
  message,
  isOwnMessage,
}: ConversationMessageBubbleProps) {
  return (
    <Box
      sx={{
        alignSelf: isOwnMessage ? "flex-end" : "flex-start",
        background: isOwnMessage
          ? "linear-gradient(135deg, #2563EB 0%, #0EA5E9 100%)"
          : "linear-gradient(135deg, rgba(251, 191, 36, 0.18) 0%, rgba(249, 115, 22, 0.12) 100%)",
        border: "1px solid",
        borderColor: isOwnMessage ? "transparent" : "rgba(245, 158, 11, 0.28)",
        borderRadius: 3,
        color: isOwnMessage ? "primary.contrastText" : "text.primary",
        boxShadow: isOwnMessage
          ? "0 10px 22px rgba(37, 99, 235, 0.22)"
          : "0 8px 18px rgba(245, 158, 11, 0.10)",
        maxWidth: "78%",
        px: 2,
        py: 1.5,
      }}
    >
      <Box
        sx={{
          alignItems: "center",
          display: "flex",
          gap: 1,
          justifyContent: "space-between",
          mb: 0.5,
        }}
      >
        <Box sx={{ alignItems: "center", display: "flex", gap: 0.75 }}>
          {isOwnMessage ? (
            <SendRoundedIcon fontSize="small" />
          ) : (
            <ForumRoundedIcon fontSize="small" />
          )}
          <Typography sx={{ fontWeight: 700 }} variant="body2">
            {message.sender.name}
          </Typography>
        </Box>
        <Typography
          color={isOwnMessage ? "inherit" : "text.secondary"}
          variant="caption"
        >
          {formatDateTime(message.createdAt)}
        </Typography>
      </Box>
      <Typography sx={{ whiteSpace: "pre-wrap" }} variant="body2">
        {message.body}
      </Typography>
    </Box>
  );
}
