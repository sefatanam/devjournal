export interface JournalEntry {
  id: string;
  userId: string;
  title: string;
  content: string;
  mood: JournalMood;
  tags: string[];
  createdAt: Date;
  updatedAt: Date;
}

export type JournalMood =
  | 'excited'
  | 'happy'
  | 'neutral'
  | 'frustrated'
  | 'tired';

export interface CreateJournalRequest {
  title: string;
  content: string;
  mood: JournalMood;
  tags: string[];
}

export interface UpdateJournalRequest {
  title?: string;
  content?: string;
  mood?: JournalMood;
  tags?: string[];
}

export interface JournalFilter {
  mood?: JournalMood;
  tags?: string[];
  search?: string;
  startDate?: Date;
  endDate?: Date;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}
