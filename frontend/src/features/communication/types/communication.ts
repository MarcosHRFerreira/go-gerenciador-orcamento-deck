import type { UserRole } from "../../users/types/user";

export type NoticePriority = "info" | "warning" | "critical";
export type NoticeScopeType = "all" | "users";
export type NoticeStatusFilter = "all" | "unread" | "read";

export type NoticeApiItem = {
  id: number;
  title: string;
  body: string;
  scope_type: NoticeScopeType;
  priority: NoticePriority;
  pinned: boolean;
  expires_at?: string | null;
  created_by_user_id: number;
  created_by_user_name: string;
  read_at?: string | null;
  created_at: string;
  updated_at: string;
};

export type NoticeItem = {
  id: number;
  title: string;
  body: string;
  scopeType: NoticeScopeType;
  priority: NoticePriority;
  pinned: boolean;
  expiresAt: string | null;
  createdByUserId: number;
  createdByUserName: string;
  readAt: string | null;
  createdAt: string;
  updatedAt: string;
};

export type CreateNoticePayload = {
  title: string;
  body: string;
  scopeType: NoticeScopeType;
  priority: NoticePriority;
  pinned: boolean;
  expiresAt: string | null;
  recipientUserIds: number[];
};

export type CreateNoticeResponse = {
  id: number;
};

export type NoticeUnreadCountResponse = {
  count: number;
};

export type ConversationType = "direct";

export type ConversationParticipantApiItem = {
  id: number;
  name: string;
  username: string;
  role: UserRole;
};

export type ConversationParticipantItem = {
  id: number;
  name: string;
  username: string;
  role: UserRole;
};

export type ConversationProjectApiItem = {
  id: number;
  code: string;
  name: string;
};

export type ConversationProjectItem = {
  id: number;
  code: string;
  name: string;
};

export type ConversationApiItem = {
  id: number;
  type: ConversationType;
  updated_at: string;
  project?: ConversationProjectApiItem | null;
  participant: ConversationParticipantApiItem;
  last_message_id?: number | null;
  last_message_body?: string | null;
  last_message_at?: string | null;
  last_message_sender?: number | null;
  unread_count: number;
};

export type ConversationItem = {
  id: number;
  type: ConversationType;
  updatedAt: string;
  project: ConversationProjectItem | null;
  participant: ConversationParticipantItem;
  lastMessageId: number | null;
  lastMessageBody: string | null;
  lastMessageAt: string | null;
  lastMessageSender: number | null;
  unreadCount: number;
};

export type ConversationMessageApiItem = {
  id: number;
  conversation_id: number;
  sender: ConversationParticipantApiItem;
  body: string;
  created_at: string;
  updated_at: string;
};

export type ConversationMessageItem = {
  id: number;
  conversationId: number;
  sender: ConversationParticipantItem;
  body: string;
  createdAt: string;
  updatedAt: string;
};

export type CreateConversationPayload = {
  participantUserId: number;
  projectId?: number | null;
  initialMessage: string;
};

export type CreateConversationResponse = {
  id: number;
};

export type SendConversationMessagePayload = {
  body: string;
};

export type SendConversationMessageResponse = {
  id: number;
};

export type ConversationUnreadCountResponse = {
  count: number;
};
