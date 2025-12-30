export interface Snippet {
  id: string;
  userId: string;
  title: string;
  description: string;
  code: string;
  language: string;
  tags: string[];
  metadata: Record<string, unknown>;
  isPublic: boolean;
  viewsCount: number;
  createdAt: Date;
  updatedAt: Date;
}

export interface CreateSnippetRequest {
  title: string;
  description: string;
  code: string;
  language: string;
  tags: string[];
  metadata?: Record<string, unknown>;
  isPublic?: boolean;
}

export interface UpdateSnippetRequest {
  title?: string;
  description?: string;
  code?: string;
  language?: string;
  tags?: string[];
  metadata?: Record<string, unknown>;
  isPublic?: boolean;
}

export interface SnippetFilter {
  language?: string;
  tags?: string[];
  search?: string;
}
