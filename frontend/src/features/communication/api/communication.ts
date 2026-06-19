import { api } from "../../../lib/axios/api";
import type {
  ConversationApiItem,
  ConversationItem,
  ConversationMessageApiItem,
  ConversationMessageItem,
  ConversationParticipantApiItem,
  ConversationParticipantItem,
  ConversationProjectApiItem,
  ConversationProjectItem,
  ConversationUnreadCountResponse,
  CreateConversationPayload,
  CreateConversationResponse,
  CreateNoticePayload,
  CreateNoticeResponse,
  NoticeApiItem,
  NoticeItem,
  NoticeStatusFilter,
  NoticeUnreadCountResponse,
  SendConversationMessagePayload,
  SendConversationMessageResponse,
} from "../types/communication";

type CreateConversationApiPayload = {
  participant_user_id: number;
  project_id?: number | null;
  initial_message: string;
};

type CreateConversationApiResponse = {
  id: number;
};

type CreateNoticeApiPayload = {
  title: string;
  body: string;
  scope_type: "all" | "users";
  priority: "info" | "warning" | "critical";
  pinned: boolean;
  expires_at?: string;
  recipient_user_ids: number[];
};

type CreateNoticeApiResponse = {
  id: number;
};

type NoticeUnreadCountApiResponse = {
  count: number;
};

type ConversationUnreadCountApiResponse = {
  count: number;
};

type SendConversationMessageApiPayload = {
  body: string;
};

type SendConversationMessageApiResponse = {
  id: number;
};

function mapNoticeItem(item: NoticeApiItem): NoticeItem {
  return {
    id: item.id,
    title: item.title,
    body: item.body,
    scopeType: item.scope_type,
    priority: item.priority,
    pinned: item.pinned,
    expiresAt: item.expires_at ?? null,
    createdByUserId: item.created_by_user_id,
    createdByUserName: item.created_by_user_name,
    readAt: item.read_at ?? null,
    createdAt: item.created_at,
    updatedAt: item.updated_at,
  };
}

function mapConversationParticipant(
  item: ConversationParticipantApiItem,
): ConversationParticipantItem {
  return {
    id: item.id,
    name: item.name,
    username: item.username,
    role: item.role,
  };
}

function mapConversationProject(
  item: ConversationProjectApiItem,
): ConversationProjectItem {
  return {
    id: item.id,
    code: item.code,
    name: item.name,
  };
}

function mapConversationItem(item: ConversationApiItem): ConversationItem {
  return {
    id: item.id,
    type: item.type,
    updatedAt: item.updated_at,
    project: item.project ? mapConversationProject(item.project) : null,
    participant: mapConversationParticipant(item.participant),
    lastMessageId: item.last_message_id ?? null,
    lastMessageBody: item.last_message_body ?? null,
    lastMessageAt: item.last_message_at ?? null,
    lastMessageSender: item.last_message_sender ?? null,
    unreadCount: item.unread_count,
  };
}

function mapConversationMessageItem(
  item: ConversationMessageApiItem,
): ConversationMessageItem {
  return {
    id: item.id,
    conversationId: item.conversation_id,
    sender: mapConversationParticipant(item.sender),
    body: item.body,
    createdAt: item.created_at,
    updatedAt: item.updated_at,
  };
}

function mapCreateConversationPayload(
  payload: CreateConversationPayload,
): CreateConversationApiPayload {
  return {
    participant_user_id: payload.participantUserId,
    project_id: payload.projectId ?? null,
    initial_message: payload.initialMessage,
  };
}

function mapCreateNoticePayload(
  payload: CreateNoticePayload,
): CreateNoticeApiPayload {
  return {
    title: payload.title,
    body: payload.body,
    scope_type: payload.scopeType,
    priority: payload.priority,
    pinned: payload.pinned,
    expires_at: payload.expiresAt ?? undefined,
    recipient_user_ids: payload.recipientUserIds,
  };
}

function mapSendConversationMessagePayload(
  payload: SendConversationMessagePayload,
): SendConversationMessageApiPayload {
  return {
    body: payload.body,
  };
}

export async function listNoticesRequest(
  status: NoticeStatusFilter,
): Promise<NoticeItem[]> {
  const response = await api.get<NoticeApiItem[]>("/notices", {
    params: {
      status: status === "all" ? undefined : status,
    },
  });

  return response.data.map(mapNoticeItem);
}

export async function getNoticeUnreadCountRequest(): Promise<number> {
  const response = await api.get<NoticeUnreadCountApiResponse>(
    "/notices/unread-count",
  );

  return response.data.count;
}

export async function listConversationsRequest(): Promise<ConversationItem[]> {
  const response = await api.get<ConversationApiItem[]>("/conversations");

  return response.data.map(mapConversationItem);
}

export async function listConversationMessagesRequest(
  conversationId: number,
): Promise<ConversationMessageItem[]> {
  const response = await api.get<ConversationMessageApiItem[]>(
    `/conversations/${conversationId}/messages`,
  );

  return response.data.map(mapConversationMessageItem);
}

export async function getConversationUnreadCountRequest(): Promise<number> {
  const response = await api.get<ConversationUnreadCountApiResponse>(
    "/conversations/unread-count",
  );

  return response.data.count;
}

export async function listConversationAvailableUsersRequest(): Promise<
  ConversationParticipantItem[]
> {
  const response = await api.get<ConversationParticipantApiItem[]>(
    "/conversations/available-users",
  );

  return response.data.map(mapConversationParticipant);
}

export async function createConversationRequest(
  payload: CreateConversationPayload,
): Promise<CreateConversationResponse> {
  const response = await api.post<CreateConversationApiResponse>(
    "/conversations",
    mapCreateConversationPayload(payload),
  );

  return {
    id: response.data.id,
  };
}

export async function createNoticeRequest(
  payload: CreateNoticePayload,
): Promise<CreateNoticeResponse> {
  const response = await api.post<CreateNoticeApiResponse>(
    "/notices",
    mapCreateNoticePayload(payload),
  );

  return {
    id: response.data.id,
  };
}

export async function markNoticeReadRequest(noticeId: number): Promise<void> {
  await api.patch(`/notices/${noticeId}/read`);
}

export async function sendConversationMessageRequest(
  conversationId: number,
  payload: SendConversationMessagePayload,
): Promise<SendConversationMessageResponse> {
  const response = await api.post<SendConversationMessageApiResponse>(
    `/conversations/${conversationId}/messages`,
    mapSendConversationMessagePayload(payload),
  );

  return {
    id: response.data.id,
  };
}

export async function markConversationReadRequest(
  conversationId: number,
): Promise<void> {
  await api.patch(`/conversations/${conversationId}/read`);
}

export const communicationPollingIntervals = {
  unreadCountMs: 5000,
  conversationsMs: 10000,
  messagesMs: 5000,
} as const;

export const communicationQueryKeys = {
  notices: (status: NoticeStatusFilter) => ["notices", status] as const,
  noticeUnreadCount: () => ["notice-unread-count"] as const,
  conversations: () => ["conversations"] as const,
  conversationMessages: (conversationId: number | null) =>
    ["conversations", conversationId, "messages"] as const,
  conversationUnreadCount: () => ["conversation-unread-count"] as const,
  availableConversationUsers: () => ["conversation-available-users"] as const,
};

export type { ConversationUnreadCountResponse, NoticeUnreadCountResponse };
