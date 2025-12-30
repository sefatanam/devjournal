import { HttpClient, HttpParams } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import {
  CreateSnippetRequest,
  PaginatedResponse,
  Snippet,
  SnippetFilter,
  UpdateSnippetRequest,
} from '@devjournal/shared-models';
import { API_CONFIG, defaultApiConfig } from './api.config';

@Injectable({ providedIn: 'root' })
export class SnippetApiService {
  private readonly http = inject(HttpClient);
  private readonly config = inject(API_CONFIG, { optional: true }) ?? defaultApiConfig;

  private readonly baseUrl = `${this.config.baseUrl}/api/snippets`;

  list(
    page = 1,
    pageSize = 10,
    filter?: SnippetFilter
  ): Observable<PaginatedResponse<Snippet>> {
    let params = new HttpParams()
      .set('page', page.toString())
      .set('pageSize', pageSize.toString());

    if (filter?.language) {
      params = params.set('language', filter.language);
    }
    if (filter?.search) {
      params = params.set('search', filter.search);
    }
    if (filter?.tags?.length) {
      params = params.set('tags', filter.tags.join(','));
    }

    return this.http.get<PaginatedResponse<Snippet>>(this.baseUrl, { params });
  }

  getById(id: string): Observable<Snippet> {
    return this.http.get<Snippet>(`${this.baseUrl}/${id}`);
  }

  create(request: CreateSnippetRequest): Observable<Snippet> {
    return this.http.post<Snippet>(this.baseUrl, request);
  }

  update(id: string, request: UpdateSnippetRequest): Observable<Snippet> {
    return this.http.put<Snippet>(`${this.baseUrl}/${id}`, request);
  }

  delete(id: string): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/${id}`);
  }

  search(
    query: string,
    limit = 20,
    offset = 0
  ): Observable<PaginatedResponse<Snippet>> {
    const params = new HttpParams()
      .set('search', query)
      .set('limit', limit.toString())
      .set('offset', offset.toString());

    return this.http.get<PaginatedResponse<Snippet>>(this.baseUrl, { params });
  }

  getLanguageStats(): Observable<Record<string, number>> {
    return this.http.get<Record<string, number>>(`${this.baseUrl}/stats/languages`);
  }
}
