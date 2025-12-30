export interface ApiError {
  message: string;
  code: string;
  details?: Record<string, string>;
}

export interface ApiResponse<T> {
  data: T;
  message?: string;
}

export interface PaginationParams {
  page: number;
  pageSize: number;
}

export interface SortParams {
  sortBy: string;
  sortOrder: 'asc' | 'desc';
}

export interface ListParams extends PaginationParams, Partial<SortParams> {
  search?: string;
}
