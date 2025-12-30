import { Injectable, inject } from '@angular/core';
import { Observable, from, map } from 'rxjs';
import { create } from '@bufbuild/protobuf';
import {
  Snippet as ProtoSnippet,
  CreateSnippetRequestSchema,
  GetSnippetRequestSchema,
  ListSnippetsRequestSchema,
  UpdateSnippetRequestSchema,
  DeleteSnippetRequestSchema,
  SearchSnippetsRequestSchema,
  GetLanguageStatsRequestSchema,
} from '@devjournal/shared-proto';
import { GrpcClientService } from './grpc-client.service';
import type {
  Snippet,
  CreateSnippetRequest,
  UpdateSnippetRequest,
  PaginatedResponse,
} from '@devjournal/shared-models';

/**
 * @REVIEW - Snippet API service using gRPC/Connect RPC
 * Provides Observable-based methods compatible with Signal Store
 */
@Injectable({ providedIn: 'root' })
export class SnippetGrpcService {
  private readonly grpcClient = inject(GrpcClientService);

  /**
   * Creates a new snippet
   */
  create(request: CreateSnippetRequest): Observable<Snippet> {
    const req = create(CreateSnippetRequestSchema, {
      title: request.title,
      description: request.description,
      code: request.code,
      language: request.language,
      tags: request.tags,
      isPublic: request.isPublic,
    });

    return from(this.grpcClient.snippetClient.createSnippet(req)).pipe(
      map((snippet) => this.mapProtoToSnippet(snippet))
    );
  }

  /**
   * Gets a snippet by ID
   */
  getById(id: string): Observable<Snippet> {
    const req = create(GetSnippetRequestSchema, { id });

    return from(this.grpcClient.snippetClient.getSnippet(req)).pipe(
      map((snippet) => this.mapProtoToSnippet(snippet))
    );
  }

  /**
   * Lists snippets with pagination and optional filters
   */
  list(
    limit = 20,
    offset = 0,
    language?: string,
    tags?: string[]
  ): Observable<PaginatedResponse<Snippet>> {
    const req = create(ListSnippetsRequestSchema, {
      limit,
      offset,
      language: language || '',
      tags: tags || [],
    });

    return from(this.grpcClient.snippetClient.listSnippets(req)).pipe(
      map((response) => ({
        items: response.snippets.map((s) => this.mapProtoToSnippet(s)),
        total: Number(response.totalCount),
        page: Math.floor(offset / limit) + 1,
        pageSize: limit,
      }))
    );
  }

  /**
   * Updates a snippet
   */
  update(id: string, request: UpdateSnippetRequest): Observable<Snippet> {
    const req = create(UpdateSnippetRequestSchema, {
      id,
      title: request.title,
      description: request.description,
      code: request.code,
      language: request.language,
      tags: request.tags,
      isPublic: request.isPublic,
    });

    return from(this.grpcClient.snippetClient.updateSnippet(req)).pipe(
      map((snippet) => this.mapProtoToSnippet(snippet))
    );
  }

  /**
   * Deletes a snippet
   */
  delete(id: string): Observable<boolean> {
    const req = create(DeleteSnippetRequestSchema, { id });

    return from(this.grpcClient.snippetClient.deleteSnippet(req)).pipe(
      map((response) => response.success)
    );
  }

  /**
   * Searches snippets
   */
  search(
    query: string,
    limit = 20,
    offset = 0
  ): Observable<PaginatedResponse<Snippet>> {
    const req = create(SearchSnippetsRequestSchema, {
      query,
      limit,
      offset,
    });

    return from(this.grpcClient.snippetClient.searchSnippets(req)).pipe(
      map((response) => ({
        items: response.snippets.map((s) => this.mapProtoToSnippet(s)),
        total: Number(response.totalCount),
        page: Math.floor(offset / limit) + 1,
        pageSize: limit,
      }))
    );
  }

  /**
   * Gets language statistics
   */
  getLanguageStats(): Observable<Record<string, number>> {
    const req = create(GetLanguageStatsRequestSchema, {});

    return from(this.grpcClient.snippetClient.getLanguageStats(req)).pipe(
      map((response) => {
        const stats: Record<string, number> = {};
        for (const [key, value] of Object.entries(response.languageCounts)) {
          stats[key] = Number(value);
        }
        return stats;
      })
    );
  }

  /**
   * Maps a proto Snippet to the domain model
   */
  private mapProtoToSnippet(proto: ProtoSnippet): Snippet {
    return {
      id: proto.id,
      userId: proto.userId,
      title: proto.title,
      description: proto.description,
      code: proto.code,
      language: proto.language,
      tags: proto.tags,
      isPublic: proto.isPublic,
      viewsCount: proto.viewsCount,
      createdAt: proto.createdAt?.toDate().toISOString() || new Date().toISOString(),
      updatedAt: proto.updatedAt?.toDate().toISOString() || new Date().toISOString(),
    };
  }
}
