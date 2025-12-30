import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { map } from 'rxjs/operators';
import {
  StudyGroup,
  StudyGroupMember,
  CreateGroupRequest,
} from '@devjournal/shared-models';
import { API_CONFIG } from './api.config';

@Injectable({ providedIn: 'root' })
export class StudyGroupApiService {
  private readonly http = inject(HttpClient);
  private readonly config = inject(API_CONFIG);

  private get baseUrl(): string {
    return `${this.config.baseUrl}/api/groups`;
  }

  list(): Observable<StudyGroup[]> {
    return this.http.get<StudyGroup[]>(this.baseUrl).pipe(
      map((groups) =>
        groups.map((g) => ({
          ...g,
          createdAt: new Date(g.createdAt),
          updatedAt: new Date(g.updatedAt),
        }))
      )
    );
  }

  listPublic(): Observable<StudyGroup[]> {
    return this.http.get<{ data: StudyGroup[]; total: number }>(`${this.baseUrl}/discover`).pipe(
      map((response) =>
        (response.data ?? []).map((g) => ({
          ...g,
          createdAt: new Date(g.createdAt),
          updatedAt: new Date(g.updatedAt),
        }))
      )
    );
  }

  getById(id: string): Observable<StudyGroup> {
    return this.http.get<StudyGroup>(`${this.baseUrl}/${id}`).pipe(
      map((g) => ({
        ...g,
        createdAt: new Date(g.createdAt),
        updatedAt: new Date(g.updatedAt),
      }))
    );
  }

  create(request: CreateGroupRequest): Observable<StudyGroup> {
    return this.http.post<StudyGroup>(this.baseUrl, request).pipe(
      map((g) => ({
        ...g,
        createdAt: new Date(g.createdAt),
        updatedAt: new Date(g.updatedAt),
      }))
    );
  }

  join(groupId: string): Observable<void> {
    return this.http.post<void>(`${this.baseUrl}/${groupId}/join`, {});
  }

  leave(groupId: string): Observable<void> {
    return this.http.post<void>(`${this.baseUrl}/${groupId}/leave`, {});
  }

  getMembers(groupId: string): Observable<StudyGroupMember[]> {
    return this.http
      .get<StudyGroupMember[]>(`${this.baseUrl}/${groupId}/members`)
      .pipe(
        map((members) =>
          members.map((m) => ({
            ...m,
            joinedAt: new Date(m.joinedAt),
          }))
        )
      );
  }

  delete(id: string): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/${id}`);
  }
}
