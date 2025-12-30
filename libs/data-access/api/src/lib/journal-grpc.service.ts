import { Injectable, inject } from '@angular/core';
import { Observable, from, map } from 'rxjs';
import { create } from '@bufbuild/protobuf';
import {
  JournalEntry as ProtoJournalEntry,
  CreateEntryRequestSchema,
  GetEntryRequestSchema,
  ListEntriesRequestSchema,
  UpdateEntryRequestSchema,
  DeleteEntryRequestSchema,
  SearchEntriesRequestSchema,
} from '@devjournal/shared-proto';
import { GrpcClientService } from './grpc-client.service';
import type {
  JournalEntry,
  CreateJournalRequest,
  UpdateJournalRequest,
  PaginatedResponse,
} from '@devjournal/shared-models';

/**
 * @REVIEW - Journal API service using gRPC/Connect RPC
 * Provides Observable-based methods compatible with Signal Store
 */
@Injectable({ providedIn: 'root' })
export class JournalGrpcService {
  private readonly grpcClient = inject(GrpcClientService);

  /**
   * Creates a new journal entry
   */
  create(request: CreateJournalRequest): Observable<JournalEntry> {
    const req = create(CreateEntryRequestSchema, {
      title: request.title,
      content: request.content,
      mood: request.mood,
      tags: request.tags,
    });

    return from(this.grpcClient.journalClient.createEntry(req)).pipe(
      map((entry) => this.mapProtoToJournalEntry(entry))
    );
  }

  /**
   * Gets a journal entry by ID
   */
  getById(id: string): Observable<JournalEntry> {
    const req = create(GetEntryRequestSchema, { id });

    return from(this.grpcClient.journalClient.getEntry(req)).pipe(
      map((entry) => this.mapProtoToJournalEntry(entry))
    );
  }

  /**
   * Lists journal entries with pagination
   */
  list(
    limit = 20,
    offset = 0,
    mood?: string
  ): Observable<PaginatedResponse<JournalEntry>> {
    const req = create(ListEntriesRequestSchema, {
      limit,
      offset,
      mood: mood || '',
    });

    return from(this.grpcClient.journalClient.listEntries(req)).pipe(
      map((response) => ({
        items: response.entries.map((e) => this.mapProtoToJournalEntry(e)),
        total: response.totalCount,
        page: Math.floor(offset / limit) + 1,
        pageSize: limit,
      }))
    );
  }

  /**
   * Updates a journal entry
   */
  update(id: string, request: UpdateJournalRequest): Observable<JournalEntry> {
    const req = create(UpdateEntryRequestSchema, {
      id,
      title: request.title,
      content: request.content,
      mood: request.mood,
      tags: request.tags,
    });

    return from(this.grpcClient.journalClient.updateEntry(req)).pipe(
      map((entry) => this.mapProtoToJournalEntry(entry))
    );
  }

  /**
   * Deletes a journal entry
   */
  delete(id: string): Observable<boolean> {
    const req = create(DeleteEntryRequestSchema, { id });

    return from(this.grpcClient.journalClient.deleteEntry(req)).pipe(
      map((response) => response.success)
    );
  }

  /**
   * Searches journal entries
   */
  search(
    query: string,
    limit = 20,
    offset = 0
  ): Observable<PaginatedResponse<JournalEntry>> {
    const req = create(SearchEntriesRequestSchema, {
      query,
      limit,
      offset,
    });

    return from(this.grpcClient.journalClient.searchEntries(req)).pipe(
      map((response) => ({
        items: response.entries.map((e) => this.mapProtoToJournalEntry(e)),
        total: response.totalCount,
        page: Math.floor(offset / limit) + 1,
        pageSize: limit,
      }))
    );
  }

  /**
   * Maps a proto JournalEntry to the domain model
   */
  private mapProtoToJournalEntry(proto: ProtoJournalEntry): JournalEntry {
    return {
      id: proto.id,
      userId: proto.userId,
      title: proto.title,
      content: proto.content,
      mood: proto.mood as JournalEntry['mood'],
      tags: proto.tags,
      createdAt: proto.createdAt?.toDate().toISOString() || new Date().toISOString(),
      updatedAt: proto.updatedAt?.toDate().toISOString() || new Date().toISOString(),
    };
  }
}
