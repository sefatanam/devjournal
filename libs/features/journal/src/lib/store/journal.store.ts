import { computed, inject } from '@angular/core';
import { Router } from '@angular/router';
import {
  patchState,
  signalStore,
  withComputed,
  withMethods,
  withState,
} from '@ngrx/signals';
import {
  addEntity,
  removeEntity,
  setAllEntities,
  updateEntity,
  withEntities,
} from '@ngrx/signals/entities';
import { rxMethod } from '@ngrx/signals/rxjs-interop';
import { tapResponse } from '@ngrx/operators';
import { pipe, switchMap, tap } from 'rxjs';
import {
  CreateJournalRequest,
  JournalEntry,
  JournalFilter,
  JournalMood,
  UpdateJournalRequest,
} from '@devjournal/shared-models';
import { JournalUnifiedService } from '@devjournal/data-access-api';

interface JournalState {
  isLoading: boolean;
  error: string | null;
  selectedId: string | null;
  filter: JournalFilter;
  currentPage: number;
  pageSize: number;
  totalItems: number;
}

const initialState: JournalState = {
  isLoading: false,
  error: null,
  selectedId: null,
  filter: {},
  currentPage: 1,
  pageSize: 10,
  totalItems: 0,
};

export const JournalStore = signalStore(
  { providedIn: 'root' },
  withState(initialState),
  withEntities<JournalEntry>(),
  withComputed((store) => ({
    selectedEntry: computed(() => {
      const id = store.selectedId();
      return id ? store.entityMap()[id] : null;
    }),
    totalPages: computed(() => Math.ceil(store.totalItems() / store.pageSize())),
    hasEntries: computed(() => store.ids().length > 0),
    entriesByMood: computed(() => {
      const entries = store.entities();
      const byMood: Record<JournalMood, JournalEntry[]> = {
        excited: [],
        happy: [],
        neutral: [],
        frustrated: [],
        tired: [],
      };
      entries.forEach((entry) => {
        if (entry.mood && byMood[entry.mood]) {
          byMood[entry.mood].push(entry);
        }
      });
      return byMood;
    }),
  })),
  withMethods(
    (store, journalApi = inject(JournalUnifiedService), router = inject(Router)) => ({
      loadEntries: rxMethod<void>(
        pipe(
          tap(() => patchState(store, { isLoading: true, error: null })),
          switchMap(() =>
            journalApi
              .list(store.currentPage(), store.pageSize(), store.filter())
              .pipe(
                tapResponse({
                  next: (response) => {
                    patchState(
                      store,
                      setAllEntities(response.data),
                      {
                        isLoading: false,
                        totalItems: response.total,
                      }
                    );
                  },
                  error: (error: Error) => {
                    patchState(store, {
                      isLoading: false,
                      error: error.message || 'Failed to load entries',
                    });
                  },
                })
              )
          )
        )
      ),

      loadEntry: rxMethod<string>(
        pipe(
          tap(() => patchState(store, { isLoading: true, error: null })),
          switchMap((id) =>
            journalApi.getById(id).pipe(
              tapResponse({
                next: (entry) => {
                  patchState(store, addEntity(entry), {
                    isLoading: false,
                    selectedId: entry.id,
                  });
                },
                error: (error: Error) => {
                  patchState(store, {
                    isLoading: false,
                    error: error.message || 'Failed to load entry',
                  });
                },
              })
            )
          )
        )
      ),

      createEntry: rxMethod<CreateJournalRequest>(
        pipe(
          tap(() => patchState(store, { isLoading: true, error: null })),
          switchMap((request) =>
            journalApi.create(request).pipe(
              tapResponse({
                next: (entry) => {
                  patchState(store, addEntity(entry), {
                    isLoading: false,
                    totalItems: store.totalItems() + 1,
                  });
                  router.navigate(['/journal', entry.id]);
                },
                error: (error: Error) => {
                  patchState(store, {
                    isLoading: false,
                    error: error.message || 'Failed to create entry',
                  });
                },
              })
            )
          )
        )
      ),

      updateEntry: rxMethod<{ id: string; request: UpdateJournalRequest }>(
        pipe(
          tap(() => patchState(store, { isLoading: true, error: null })),
          switchMap(({ id, request }) =>
            journalApi.update(id, request).pipe(
              tapResponse({
                next: (entry) => {
                  patchState(
                    store,
                    updateEntity({ id: entry.id, changes: entry }),
                    { isLoading: false }
                  );
                  router.navigate(['/journal', entry.id]);
                },
                error: (error: Error) => {
                  patchState(store, {
                    isLoading: false,
                    error: error.message || 'Failed to update entry',
                  });
                },
              })
            )
          )
        )
      ),

      deleteEntry: rxMethod<string>(
        pipe(
          tap(() => patchState(store, { isLoading: true, error: null })),
          switchMap((id) =>
            journalApi.delete(id).pipe(
              tapResponse({
                next: () => {
                  patchState(store, removeEntity(id), {
                    isLoading: false,
                    totalItems: store.totalItems() - 1,
                    selectedId: store.selectedId() === id ? null : store.selectedId(),
                  });
                  router.navigate(['/journal']);
                },
                error: (error: Error) => {
                  patchState(store, {
                    isLoading: false,
                    error: error.message || 'Failed to delete entry',
                  });
                },
              })
            )
          )
        )
      ),

      selectEntry(id: string | null): void {
        patchState(store, { selectedId: id });
      },

      setFilter(filter: JournalFilter): void {
        patchState(store, { filter, currentPage: 1 });
      },

      setPage(page: number): void {
        patchState(store, { currentPage: page });
      },

      clearError(): void {
        patchState(store, { error: null });
      },
    })
  )
);
