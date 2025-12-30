import { Injectable, inject, computed } from '@angular/core';
import { Observable } from 'rxjs';
import { ProtocolService } from './protocol.service';
import { SnippetApiService } from './snippet-api.service';
import type {
  Snippet,
  CreateSnippetRequest,
  UpdateSnippetRequest,
  SnippetFilter,
  PaginatedResponse,
} from '@devjournal/shared-models';

/**
 * @REVIEW - Unified Snippet API service
 * Currently using REST only - gRPC disabled due to version mismatch
 */
@Injectable({ providedIn: 'root' })
export class SnippetUnifiedService {
  private readonly protocolService = inject(ProtocolService);
  private readonly restApi = inject(SnippetApiService);

  /**
   * Current protocol being used
   */
  readonly protocol = computed(() => this.protocolService.protocol());

  /**
   * Creates a new snippet
   */
  create(request: CreateSnippetRequest): Observable<Snippet> {
    return this.restApi.create(request);
  }

  /**
   * Gets a snippet by ID
   */
  getById(id: string): Observable<Snippet> {
    return this.restApi.getById(id);
  }

  /**
   * Lists snippets with pagination
   */
  list(
    page: number,
    pageSize: number,
    filter: SnippetFilter = {}
  ): Observable<PaginatedResponse<Snippet>> {
    return this.restApi.list(page, pageSize, filter);
  }

  /**
   * Updates a snippet
   */
  update(id: string, request: UpdateSnippetRequest): Observable<Snippet> {
    return this.restApi.update(id, request);
  }

  /**
   * Deletes a snippet
   */
  delete(id: string): Observable<void> {
    return this.restApi.delete(id);
  }

  /**
   * Searches snippets
   */
  search(
    query: string,
    limit = 20,
    offset = 0
  ): Observable<PaginatedResponse<Snippet>> {
    return this.restApi.search(query, limit, offset);
  }

  /**
   * Gets language statistics
   */
  getLanguageStats(): Observable<Record<string, number>> {
    return this.restApi.getLanguageStats();
  }
}
