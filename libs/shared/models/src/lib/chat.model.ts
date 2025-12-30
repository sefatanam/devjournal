export interface StudyGroup {
  id: string;
  name: string;
  description: string;
  isPublic: boolean;
  maxMembers: number;
  createdBy: string;
  createdAt: Date;
  updatedAt: Date;
}

export interface StudyGroupMember {
  groupId: string;
  userId: string;
  displayName: string;
  joinedAt: Date;
}

export interface CreateGroupRequest {
  name: string;
  description: string;
  isPublic: boolean;
  maxMembers: number;
}

export interface ChatMessage {
  id: string;
  roomId: string;
  userId: string;
  userDisplayName: string;
  content: string;
  timestamp: Date;
  type: ChatMessageType;
}

export type ChatMessageType = 'message' | 'join' | 'leave' | 'system';

export interface ChatRoom {
  id: string;
  name: string;
  participants: ChatParticipant[];
  messages: ChatMessage[];
}

export interface ChatParticipant {
  userId: string;
  displayName: string;
  joinedAt: Date;
}

export interface SendMessageRequest {
  content: string;
}

export interface WebSocketMessage<T = unknown> {
  type: string;
  payload: T;
  timestamp: Date;
}
