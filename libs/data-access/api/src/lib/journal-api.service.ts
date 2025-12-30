import { HttpClient, HttpParams } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import {
  CreateJournalRequest,
  JournalEntry,
  JournalFilter,
  PaginatedResponse,
  UpdateJournalRequest,
} from '@devjournal/shared-models';
import { API_CONFIG, defaultApiConfig } from './api.config';

@Injectable({ providedIn: 'root' })
export class JournalApiService {
  private readonly http = inject(HttpClient);
  private readonly config = inject(API_CONFIG, { optional: true }) ?? defaultApiConfig;

  private readonly baseUrl = `${this.config.baseUrl}/api/entries`;

  list(
    page = 1,
    pageSize = 10,
    filter?: JournalFilter
  ): Observable<PaginatedResponse<JournalEntry>> {
    let params = new HttpParams()
      .set('page', page.toString())
      .set('pageSize', pageSize.toString());

    if (filter?.mood) {
      params = params.set('mood', filter.mood);
    }
    if (filter?.search) {
      params = params.set('search', filter.search);
    }
    if (filter?.tags?.length) {
      params = params.set('tags', filter.tags.join(','));
    }

    return this.http.get<PaginatedResponse<JournalEntry>>(this.baseUrl, { params });
  }

  getById(id: string): Observable<JournalEntry> {
    return this.http.get<JournalEntry>(`${this.baseUrl}/${id}`);
  }

  create(request: CreateJournalRequest): Observable<JournalEntry> {
    return this.http.post<JournalEntry>(this.baseUrl, request);
  }

  update(id: string, request: UpdateJournalRequest): Observable<JournalEntry> {
    return this.http.put<JournalEntry>(`${this.baseUrl}/${id}`, request);
  }

  delete(id: string): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/${id}`);
  }

  search(
    query: string,
    limit = 20,
    offset = 0
  ): Observable<PaginatedResponse<JournalEntry>> {
    const params = new HttpParams()
      .set('search', query)
      .set('limit', limit.toString())
      .set('offset', offset.toString());

    return this.http.get<PaginatedResponse<JournalEntry>>(this.baseUrl, { params });
  }
}
