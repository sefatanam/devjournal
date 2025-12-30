import { Injectable, inject, computed } from '@angular/core';
import { Observable } from 'rxjs';
import { ProtocolService } from './protocol.service';
import { JournalApiService } from './journal-api.service';
import type {
  JournalEntry,
  CreateJournalRequest,
  UpdateJournalRequest,
  JournalFilter,
  PaginatedResponse,
} from '@devjournal/shared-models';

/**
 * @REVIEW - Unified Journal API service
 * Currently using REST only - gRPC disabled due to version mismatch
 */
@Injectable({ providedIn: 'root' })
export class JournalUnifiedService {
  private readonly protocolService = inject(ProtocolService);
  private readonly restApi = inject(JournalApiService);

  /**
   * Current protocol being used
   */
  readonly protocol = computed(() => this.protocolService.protocol());

  /**
   * Creates a new journal entry
   */
  create(request: CreateJournalRequest): Observable<JournalEntry> {
    return this.restApi.create(request);
  }

  /**
   * Gets a journal entry by ID
   */
  getById(id: string): Observable<JournalEntry> {
    return this.restApi.getById(id);
  }

  /**
   * Lists journal entries with pagination
   */
  list(
    page: number,
    pageSize: number,
    filter: JournalFilter = {}
  ): Observable<PaginatedResponse<JournalEntry>> {
    return this.restApi.list(page, pageSize, filter);
  }

  /**
   * Updates a journal entry
   */
  update(id: string, request: UpdateJournalRequest): Observable<JournalEntry> {
    return this.restApi.update(id, request);
  }

  /**
   * Deletes a journal entry
   */
  delete(id: string): Observable<void> {
    return this.restApi.delete(id);
  }

  /**
   * Searches journal entries
   */
  search(
    query: string,
    limit = 20,
    offset = 0
  ): Observable<PaginatedResponse<JournalEntry>> {
    return this.restApi.search(query, limit, offset);
  }
}
