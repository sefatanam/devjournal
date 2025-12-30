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
  CreateSnippetRequest,
  Snippet,
  SnippetFilter,
  UpdateSnippetRequest,
} from '@devjournal/shared-models';
import { SnippetUnifiedService } from '@devjournal/data-access-api';

interface SnippetState {
  isLoading: boolean;
  error: string | null;
  selectedId: string | null;
  filter: SnippetFilter;
  currentPage: number;
  pageSize: number;
  totalItems: number;
}

const initialState: SnippetState = {
  isLoading: false,
  error: null,
  selectedId: null,
  filter: {},
  currentPage: 1,
  pageSize: 12,
  totalItems: 0,
};

export const SUPPORTED_LANGUAGES = [
  { value: 'typescript', label: 'TypeScript', icon: '🔷' },
  { value: 'javascript', label: 'JavaScript', icon: '🟨' },
  { value: 'go', label: 'Go', icon: '🐹' },
  { value: 'python', label: 'Python', icon: '🐍' },
  { value: 'rust', label: 'Rust', icon: '🦀' },
  { value: 'java', label: 'Java', icon: '☕' },
  { value: 'csharp', label: 'C#', icon: '🟣' },
  { value: 'html', label: 'HTML', icon: '🌐' },
  { value: 'css', label: 'CSS', icon: '🎨' },
  { value: 'scss', label: 'SCSS', icon: '🎨' },
  { value: 'sql', label: 'SQL', icon: '🗃️' },
  { value: 'bash', label: 'Bash', icon: '💻' },
  { value: 'json', label: 'JSON', icon: '📋' },
  { value: 'yaml', label: 'YAML', icon: '📄' },
  { value: 'markdown', label: 'Markdown', icon: '📝' },
  { value: 'other', label: 'Other', icon: '📦' },
] as const;

export type SupportedLanguage = (typeof SUPPORTED_LANGUAGES)[number]['value'];

export const SnippetStore = signalStore(
  { providedIn: 'root' },
  withState(initialState),
  withEntities<Snippet>(),
  withComputed((store) => ({
    selectedSnippet: computed(() => {
      const id = store.selectedId();
      return id ? store.entityMap()[id] : null;
    }),
    totalPages: computed(() => Math.ceil(store.totalItems() / store.pageSize())),
    hasSnippets: computed(() => store.ids().length > 0),
    snippetsByLanguage: computed(() => {
      const snippets = store.entities();
      const byLanguage: Record<string, Snippet[]> = {};
      snippets.forEach((snippet) => {
        const lang = snippet.language || 'other';
        if (!byLanguage[lang]) {
          byLanguage[lang] = [];
        }
        byLanguage[lang].push(snippet);
      });
      return byLanguage;
    }),
    uniqueTags: computed(() => {
      const snippets = store.entities();
      const tags = new Set<string>();
      snippets.forEach((snippet) => {
        snippet.tags.forEach((tag) => tags.add(tag));
      });
      return Array.from(tags).sort();
    }),
  })),
  withMethods(
    (store, snippetApi = inject(SnippetUnifiedService), router = inject(Router)) => ({
      loadSnippets: rxMethod<void>(
        pipe(
          tap(() => patchState(store, { isLoading: true, error: null })),
          switchMap(() =>
            snippetApi
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
                      error: error.message || 'Failed to load snippets',
                    });
                  },
                })
              )
          )
        )
      ),

      loadSnippet: rxMethod<string>(
        pipe(
          tap(() => patchState(store, { isLoading: true, error: null })),
          switchMap((id) =>
            snippetApi.getById(id).pipe(
              tapResponse({
                next: (snippet) => {
                  patchState(store, addEntity(snippet), {
                    isLoading: false,
                    selectedId: snippet.id,
                  });
                },
                error: (error: Error) => {
                  patchState(store, {
                    isLoading: false,
                    error: error.message || 'Failed to load snippet',
                  });
                },
              })
            )
          )
        )
      ),

      createSnippet: rxMethod<CreateSnippetRequest>(
        pipe(
          tap(() => patchState(store, { isLoading: true, error: null })),
          switchMap((request) =>
            snippetApi.create(request).pipe(
              tapResponse({
                next: (snippet) => {
                  patchState(store, addEntity(snippet), {
                    isLoading: false,
                    totalItems: store.totalItems() + 1,
                  });
                  router.navigate(['/snippets', snippet.id]);
                },
                error: (error: Error) => {
                  patchState(store, {
                    isLoading: false,
                    error: error.message || 'Failed to create snippet',
                  });
                },
              })
            )
          )
        )
      ),

      updateSnippet: rxMethod<{ id: string; request: UpdateSnippetRequest }>(
        pipe(
          tap(() => patchState(store, { isLoading: true, error: null })),
          switchMap(({ id, request }) =>
            snippetApi.update(id, request).pipe(
              tapResponse({
                next: (snippet) => {
                  patchState(
                    store,
                    updateEntity({ id: snippet.id, changes: snippet }),
                    { isLoading: false }
                  );
                  router.navigate(['/snippets', snippet.id]);
                },
                error: (error: Error) => {
                  patchState(store, {
                    isLoading: false,
                    error: error.message || 'Failed to update snippet',
                  });
                },
              })
            )
          )
        )
      ),

      deleteSnippet: rxMethod<string>(
        pipe(
          tap(() => patchState(store, { isLoading: true, error: null })),
          switchMap((id) =>
            snippetApi.delete(id).pipe(
              tapResponse({
                next: () => {
                  patchState(store, removeEntity(id), {
                    isLoading: false,
                    totalItems: store.totalItems() - 1,
                    selectedId: store.selectedId() === id ? null : store.selectedId(),
                  });
                  router.navigate(['/snippets']);
                },
                error: (error: Error) => {
                  patchState(store, {
                    isLoading: false,
                    error: error.message || 'Failed to delete snippet',
                  });
                },
              })
            )
          )
        )
      ),

      selectSnippet(id: string | null): void {
        patchState(store, { selectedId: id });
      },

      setFilter(filter: SnippetFilter): void {
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
