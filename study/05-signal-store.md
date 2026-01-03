# Angular Signal Store

## Overview

DevJournal uses **NgRx Signal Store** (`@ngrx/signals`) for state management. This is a modern, signal-based approach that replaces traditional NgRx Store with a simpler, more intuitive API.

## Why Signal Store?

| Feature | Traditional NgRx | Signal Store |
|---------|-----------------|--------------|
| Boilerplate | Actions, Reducers, Effects, Selectors | Single store file |
| Reactivity | RxJS Observables | Angular Signals |
| Learning Curve | Steep | Gentle |
| Type Safety | Good | Excellent |
| Performance | Good | Better (fine-grained) |
| DevTools | Redux DevTools | Planned |

## Signal Store Anatomy

```typescript
import {
  signalStore,
  withState,
  withComputed,
  withMethods,
  patchState,
} from '@ngrx/signals';
import { withEntities, setAllEntities, addEntity, updateEntity, removeEntity } from '@ngrx/signals/entities';
import { rxMethod } from '@ngrx/signals/rxjs-interop';
import { computed, inject } from '@angular/core';

// Define initial state type
interface JournalState {
  isLoading: boolean;
  error: string | null;
  filter: JournalFilter;
  selectedId: string | null;
}

// Create the store
export const JournalStore = signalStore(
  { providedIn: 'root' },  // Singleton store

  // 1. Initial state
  withState<JournalState>({
    isLoading: false,
    error: null,
    filter: { page: 1, pageSize: 10 },
    selectedId: null,
  }),

  // 2. Entity collection (like an array with built-in CRUD)
  withEntities<JournalEntry>(),

  // 3. Computed values (derived state)
  withComputed((store) => ({
    selectedEntry: computed(() => {
      const id = store.selectedId();
      return store.entityMap()[id] ?? null;
    }),
    totalPages: computed(() => {
      // Assuming total is stored somewhere
      return Math.ceil(store.entities().length / store.filter().pageSize);
    }),
    hasEntries: computed(() => store.entities().length > 0),
  })),

  // 4. Methods (actions + effects combined)
  withMethods((store, apiService = inject(JournalApiService)) => ({
    // Sync method
    setFilter(filter: Partial<JournalFilter>) {
      patchState(store, { filter: { ...store.filter(), ...filter } });
    },

    // Async method using rxMethod
    loadEntries: rxMethod<void>(
      pipe(
        tap(() => patchState(store, { isLoading: true, error: null })),
        switchMap(() =>
          apiService.list(store.filter()).pipe(
            tapResponse({
              next: (response) => {
                patchState(store, setAllEntities(response.data));
                patchState(store, { isLoading: false });
              },
              error: (error) => {
                patchState(store, { error: 'Failed to load entries', isLoading: false });
              },
            })
          )
        )
      )
    ),
  })),
);
```

## Journal Store Implementation

```typescript
// libs/features/journal/src/lib/store/journal.store.ts

import { computed, inject } from '@angular/core';
import {
  signalStore,
  withState,
  withComputed,
  withMethods,
  patchState,
} from '@ngrx/signals';
import {
  withEntities,
  setAllEntities,
  addEntity,
  updateEntity,
  removeEntity,
  selectId,
} from '@ngrx/signals/entities';
import { rxMethod } from '@ngrx/signals/rxjs-interop';
import { pipe, switchMap, tap } from 'rxjs';
import { tapResponse } from '@ngrx/operators';
import { JournalApiService } from '@devjournal/data-access-api';
import {
  JournalEntry,
  JournalFilter,
  JournalMood,
  CreateJournalRequest,
  UpdateJournalRequest,
} from '@devjournal/shared-models';

// State interface
interface JournalState {
  isLoading: boolean;
  error: string | null;
  filter: JournalFilter;
  selectedId: string | null;
  total: number;
}

// Initial state
const initialState: JournalState = {
  isLoading: false,
  error: null,
  filter: {
    page: 1,
    pageSize: 10,
    mood: undefined,
    search: undefined,
  },
  selectedId: null,
  total: 0,
};

export const JournalStore = signalStore(
  { providedIn: 'root' },

  // Base state
  withState(initialState),

  // Entity collection for journal entries
  withEntities<JournalEntry>(),

  // Computed signals (derived state)
  withComputed((store) => ({
    // Get currently selected entry
    selectedEntry: computed(() => {
      const id = store.selectedId();
      if (!id) return null;
      return store.entityMap()[id] ?? null;
    }),

    // Calculate total pages
    totalPages: computed(() => {
      const total = store.total();
      const pageSize = store.filter().pageSize;
      return Math.ceil(total / pageSize);
    }),

    // Check if there are any entries
    hasEntries: computed(() => store.entities().length > 0),

    // Group entries by mood
    entriesByMood: computed(() => {
      const entries = store.entities();
      return entries.reduce((acc, entry) => {
        const mood = entry.mood || 'unspecified';
        if (!acc[mood]) acc[mood] = [];
        acc[mood].push(entry);
        return acc;
      }, {} as Record<string, JournalEntry[]>);
    }),

    // Get current page info
    pagination: computed(() => ({
      page: store.filter().page,
      pageSize: store.filter().pageSize,
      total: store.total(),
      totalPages: Math.ceil(store.total() / store.filter().pageSize),
    })),
  })),

  // Methods (sync and async)
  withMethods((store, apiService = inject(JournalApiService)) => ({
    // ============ Sync Methods ============

    // Select an entry by ID
    selectEntry(id: string | null) {
      patchState(store, { selectedId: id });
    },

    // Update filter
    setFilter(filter: Partial<JournalFilter>) {
      patchState(store, {
        filter: { ...store.filter(), ...filter },
      });
    },

    // Set page
    setPage(page: number) {
      patchState(store, {
        filter: { ...store.filter(), page },
      });
    },

    // Set mood filter
    setMoodFilter(mood: JournalMood | undefined) {
      patchState(store, {
        filter: { ...store.filter(), mood, page: 1 },
      });
    },

    // Set search
    setSearch(search: string | undefined) {
      patchState(store, {
        filter: { ...store.filter(), search, page: 1 },
      });
    },

    // Clear error
    clearError() {
      patchState(store, { error: null });
    },

    // Reset state
    reset() {
      patchState(store, initialState);
    },

    // ============ Async Methods (rxMethod) ============

    // Load entries with current filter
    loadEntries: rxMethod<void>(
      pipe(
        tap(() => patchState(store, { isLoading: true, error: null })),
        switchMap(() =>
          apiService.list(store.filter()).pipe(
            tapResponse({
              next: (response) => {
                patchState(store, {
                  isLoading: false,
                  total: response.total,
                });
                patchState(store, setAllEntities(response.data));
              },
              error: (err: Error) => {
                patchState(store, {
                  isLoading: false,
                  error: err.message || 'Failed to load entries',
                });
              },
            })
          )
        )
      )
    ),

    // Load single entry by ID
    loadEntry: rxMethod<string>(
      pipe(
        tap(() => patchState(store, { isLoading: true, error: null })),
        switchMap((id) =>
          apiService.getById(id).pipe(
            tapResponse({
              next: (entry) => {
                patchState(store, { isLoading: false, selectedId: entry.id });
                // Add or update in entity collection
                const existing = store.entityMap()[entry.id];
                if (existing) {
                  patchState(store, updateEntity({ id: entry.id, changes: entry }));
                } else {
                  patchState(store, addEntity(entry));
                }
              },
              error: (err: Error) => {
                patchState(store, {
                  isLoading: false,
                  error: err.message || 'Failed to load entry',
                });
              },
            })
          )
        )
      )
    ),

    // Create new entry
    createEntry: rxMethod<CreateJournalRequest>(
      pipe(
        tap(() => patchState(store, { isLoading: true, error: null })),
        switchMap((request) =>
          apiService.create(request).pipe(
            tapResponse({
              next: (entry) => {
                patchState(store, { isLoading: false });
                patchState(store, addEntity(entry));
              },
              error: (err: Error) => {
                patchState(store, {
                  isLoading: false,
                  error: err.message || 'Failed to create entry',
                });
              },
            })
          )
        )
      )
    ),

    // Update existing entry
    updateEntry: rxMethod<{ id: string; data: UpdateJournalRequest }>(
      pipe(
        tap(() => patchState(store, { isLoading: true, error: null })),
        switchMap(({ id, data }) =>
          apiService.update(id, data).pipe(
            tapResponse({
              next: (entry) => {
                patchState(store, { isLoading: false });
                patchState(store, updateEntity({ id: entry.id, changes: entry }));
              },
              error: (err: Error) => {
                patchState(store, {
                  isLoading: false,
                  error: err.message || 'Failed to update entry',
                });
              },
            })
          )
        )
      )
    ),

    // Delete entry
    deleteEntry: rxMethod<string>(
      pipe(
        tap(() => patchState(store, { isLoading: true, error: null })),
        switchMap((id) =>
          apiService.delete(id).pipe(
            tapResponse({
              next: () => {
                patchState(store, { isLoading: false });
                patchState(store, removeEntity(id));
              },
              error: (err: Error) => {
                patchState(store, {
                  isLoading: false,
                  error: err.message || 'Failed to delete entry',
                });
              },
            })
          )
        )
      )
    ),
  }))
);
```

## Using the Store in Components

### List Component

```typescript
// libs/features/journal/src/lib/components/journal-list/journal-list.component.ts

import { Component, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { JournalStore } from '../../store/journal.store';
import { JournalMood } from '@devjournal/shared-models';

@Component({
  selector: 'lib-journal-list',
  standalone: true,
  imports: [CommonModule, RouterLink],
  template: `
    <div class="journal-list">
      <!-- Loading State -->
      @if (store.isLoading()) {
        <div class="loading">Loading entries...</div>
      }

      <!-- Error State -->
      @if (store.error()) {
        <div class="error">
          {{ store.error() }}
          <button (click)="store.clearError()">Dismiss</button>
        </div>
      }

      <!-- Filters -->
      <div class="filters">
        <input
          type="text"
          [value]="store.filter().search || ''"
          (input)="onSearchChange($event)"
          placeholder="Search entries..."
        />

        <select (change)="onMoodChange($event)">
          <option value="">All Moods</option>
          @for (mood of moods; track mood) {
            <option [value]="mood" [selected]="store.filter().mood === mood">
              {{ mood }}
            </option>
          }
        </select>
      </div>

      <!-- Entry List -->
      @if (store.hasEntries()) {
        <ul class="entries">
          @for (entry of store.entities(); track entry.id) {
            <li class="entry-card">
              <a [routerLink]="['/journal', entry.id]">
                <h3>{{ entry.title }}</h3>
                <p>{{ entry.content | slice:0:100 }}...</p>
                <div class="meta">
                  <span class="mood">{{ entry.mood }}</span>
                  <span class="date">{{ entry.createdAt | date:'short' }}</span>
                </div>
              </a>
            </li>
          }
        </ul>

        <!-- Pagination -->
        <div class="pagination">
          <button
            [disabled]="store.filter().page <= 1"
            (click)="previousPage()"
          >
            Previous
          </button>
          <span>
            Page {{ store.pagination().page }} of {{ store.pagination().totalPages }}
          </span>
          <button
            [disabled]="store.filter().page >= store.pagination().totalPages"
            (click)="nextPage()"
          >
            Next
          </button>
        </div>
      } @else if (!store.isLoading()) {
        <div class="empty-state">
          <p>No entries found. Start journaling!</p>
          <a routerLink="/journal/new" class="btn-create">Create Entry</a>
        </div>
      }
    </div>
  `,
})
export class JournalListComponent implements OnInit {
  readonly store = inject(JournalStore);

  readonly moods: JournalMood[] = [
    'productive',
    'learning',
    'struggling',
    'breakthrough',
    'tired',
  ];

  ngOnInit(): void {
    // Load entries when component initializes
    this.store.loadEntries();
  }

  onSearchChange(event: Event): void {
    const value = (event.target as HTMLInputElement).value;
    this.store.setSearch(value || undefined);
    this.store.loadEntries();
  }

  onMoodChange(event: Event): void {
    const value = (event.target as HTMLSelectElement).value;
    this.store.setMoodFilter(value as JournalMood || undefined);
    this.store.loadEntries();
  }

  previousPage(): void {
    this.store.setPage(this.store.filter().page - 1);
    this.store.loadEntries();
  }

  nextPage(): void {
    this.store.setPage(this.store.filter().page + 1);
    this.store.loadEntries();
  }
}
```

### Form Component

```typescript
// libs/features/journal/src/lib/components/journal-form/journal-form.component.ts

import { Component, inject, input, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import { JournalStore } from '../../store/journal.store';

@Component({
  selector: 'lib-journal-form',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule],
  template: `
    <form [formGroup]="form" (ngSubmit)="onSubmit()">
      <div class="form-group">
        <label for="title">Title</label>
        <input
          id="title"
          type="text"
          formControlName="title"
          placeholder="What did you work on today?"
        />
        @if (form.get('title')?.invalid && form.get('title')?.touched) {
          <span class="error">Title is required</span>
        }
      </div>

      <div class="form-group">
        <label for="content">Content</label>
        <textarea
          id="content"
          formControlName="content"
          rows="10"
          placeholder="Write about your day..."
        ></textarea>
      </div>

      <div class="form-group">
        <label>Mood</label>
        <div class="mood-selector">
          @for (mood of moods; track mood.value) {
            <button
              type="button"
              [class.selected]="form.get('mood')?.value === mood.value"
              (click)="form.patchValue({ mood: mood.value })"
            >
              {{ mood.emoji }} {{ mood.label }}
            </button>
          }
        </div>
      </div>

      <div class="form-group">
        <label for="tags">Tags (comma separated)</label>
        <input
          id="tags"
          type="text"
          formControlName="tagsInput"
          placeholder="angular, typescript, learning"
        />
      </div>

      <div class="form-actions">
        <button type="button" (click)="cancel()">Cancel</button>
        <button
          type="submit"
          [disabled]="form.invalid || store.isLoading()"
        >
          @if (store.isLoading()) {
            Saving...
          } @else if (isEditMode()) {
            Update Entry
          } @else {
            Create Entry
          }
        </button>
      </div>
    </form>
  `,
})
export class JournalFormComponent implements OnInit {
  // Input for edit mode (entry ID from route)
  readonly entryId = input<string>();

  readonly store = inject(JournalStore);
  private readonly fb = inject(FormBuilder);
  private readonly router = inject(Router);

  readonly moods = [
    { value: 'productive', label: 'Productive', emoji: '🚀' },
    { value: 'learning', label: 'Learning', emoji: '📚' },
    { value: 'struggling', label: 'Struggling', emoji: '😤' },
    { value: 'breakthrough', label: 'Breakthrough', emoji: '💡' },
    { value: 'tired', label: 'Tired', emoji: '😴' },
  ];

  form = this.fb.group({
    title: ['', [Validators.required, Validators.minLength(3)]],
    content: ['', [Validators.required]],
    mood: ['productive'],
    tagsInput: [''],
  });

  ngOnInit(): void {
    const id = this.entryId();
    if (id) {
      // Edit mode - load entry
      this.store.loadEntry(id);

      // Populate form when entry loads
      // Using effect would be cleaner, but this works for the example
      const entry = this.store.selectedEntry();
      if (entry) {
        this.form.patchValue({
          title: entry.title,
          content: entry.content,
          mood: entry.mood,
          tagsInput: entry.tags?.join(', ') || '',
        });
      }
    }
  }

  isEditMode(): boolean {
    return !!this.entryId();
  }

  onSubmit(): void {
    if (this.form.invalid) return;

    const formValue = this.form.value;
    const tags = formValue.tagsInput
      ?.split(',')
      .map((t) => t.trim())
      .filter((t) => t.length > 0) || [];

    const data = {
      title: formValue.title!,
      content: formValue.content!,
      mood: formValue.mood!,
      tags,
    };

    if (this.isEditMode()) {
      this.store.updateEntry({ id: this.entryId()!, data });
    } else {
      this.store.createEntry(data);
    }

    // Navigate back after save
    this.router.navigate(['/journal']);
  }

  cancel(): void {
    this.router.navigate(['/journal']);
  }
}
```

## Snippet Store (with Language Stats)

```typescript
// libs/features/snippets/src/lib/store/snippet.store.ts

import { computed, inject } from '@angular/core';
import {
  signalStore,
  withState,
  withComputed,
  withMethods,
  patchState,
} from '@ngrx/signals';
import { withEntities, setAllEntities, addEntity, updateEntity, removeEntity } from '@ngrx/signals/entities';
import { rxMethod } from '@ngrx/signals/rxjs-interop';
import { pipe, switchMap, tap } from 'rxjs';
import { tapResponse } from '@ngrx/operators';
import { SnippetApiService } from '@devjournal/data-access-api';
import { Snippet, SnippetFilter, LanguageStat } from '@devjournal/shared-models';

interface SnippetState {
  isLoading: boolean;
  error: string | null;
  filter: SnippetFilter;
  selectedId: string | null;
  total: number;
  languageStats: LanguageStat[];
}

const initialState: SnippetState = {
  isLoading: false,
  error: null,
  filter: { page: 1, pageSize: 12 },
  selectedId: null,
  total: 0,
  languageStats: [],
};

export const SnippetStore = signalStore(
  { providedIn: 'root' },

  withState(initialState),
  withEntities<Snippet>(),

  withComputed((store) => ({
    selectedSnippet: computed(() => {
      const id = store.selectedId();
      return id ? store.entityMap()[id] ?? null : null;
    }),

    // Get snippets by language
    snippetsByLanguage: computed(() => {
      const snippets = store.entities();
      return snippets.reduce((acc, snippet) => {
        const lang = snippet.language;
        if (!acc[lang]) acc[lang] = [];
        acc[lang].push(snippet);
        return acc;
      }, {} as Record<string, Snippet[]>);
    }),

    // Top languages
    topLanguages: computed(() => {
      return store.languageStats().slice(0, 5);
    }),

    // Total snippets count
    totalSnippets: computed(() => store.total()),
  })),

  withMethods((store, apiService = inject(SnippetApiService)) => ({
    setLanguageFilter(language: string | undefined) {
      patchState(store, {
        filter: { ...store.filter(), language, page: 1 },
      });
    },

    setTagFilter(tag: string | undefined) {
      patchState(store, {
        filter: { ...store.filter(), tag, page: 1 },
      });
    },

    loadSnippets: rxMethod<void>(
      pipe(
        tap(() => patchState(store, { isLoading: true, error: null })),
        switchMap(() =>
          apiService.list(store.filter()).pipe(
            tapResponse({
              next: (response) => {
                patchState(store, {
                  isLoading: false,
                  total: response.total,
                });
                patchState(store, setAllEntities(response.data));
              },
              error: (err: Error) => {
                patchState(store, {
                  isLoading: false,
                  error: err.message,
                });
              },
            })
          )
        )
      )
    ),

    loadLanguageStats: rxMethod<void>(
      pipe(
        switchMap(() =>
          apiService.getLanguageStats().pipe(
            tapResponse({
              next: (stats) => {
                patchState(store, { languageStats: stats });
              },
              error: () => {
                // Silent fail for stats
              },
            })
          )
        )
      )
    ),

    toggleFavorite: rxMethod<string>(
      pipe(
        switchMap((id) =>
          apiService.toggleFavorite(id).pipe(
            tapResponse({
              next: (isFavorite) => {
                patchState(store, updateEntity({
                  id,
                  changes: { isFavorite },
                }));
              },
              error: () => {},
            })
          )
        )
      )
    ),
  }))
);
```

## Auth Store

```typescript
// libs/features/auth/src/lib/auth.store.ts

import { computed, inject } from '@angular/core';
import {
  signalStore,
  withState,
  withComputed,
  withMethods,
  patchState,
} from '@ngrx/signals';
import { rxMethod } from '@ngrx/signals/rxjs-interop';
import { pipe, switchMap, tap } from 'rxjs';
import { tapResponse } from '@ngrx/operators';
import { Router } from '@angular/router';
import { AuthApiService } from '@devjournal/data-access-api';
import { User, AuthState } from '@devjournal/shared-models';

const initialState: AuthState = {
  user: null,
  token: null,
  isLoading: false,
  error: null,
};

export const AuthStore = signalStore(
  { providedIn: 'root' },

  withState(initialState),

  withComputed((store) => ({
    isAuthenticated: computed(() => !!store.token()),
    displayName: computed(() => store.user()?.displayName || 'User'),
    userInitials: computed(() => {
      const name = store.user()?.displayName || '';
      return name.split(' ').map(n => n[0]).join('').toUpperCase();
    }),
  })),

  withMethods((store, authService = inject(AuthApiService), router = inject(Router)) => ({
    // Initialize from localStorage
    init() {
      const token = localStorage.getItem('token');
      const userJson = localStorage.getItem('user');

      if (token && userJson) {
        try {
          const user = JSON.parse(userJson);
          patchState(store, { token, user });
        } catch {
          this.logout();
        }
      }
    },

    // Login
    login: rxMethod<{ email: string; password: string }>(
      pipe(
        tap(() => patchState(store, { isLoading: true, error: null })),
        switchMap(({ email, password }) =>
          authService.login(email, password).pipe(
            tapResponse({
              next: (response) => {
                // Save to localStorage
                localStorage.setItem('token', response.token);
                localStorage.setItem('user', JSON.stringify(response.user));

                patchState(store, {
                  token: response.token,
                  user: response.user,
                  isLoading: false,
                });

                router.navigate(['/dashboard']);
              },
              error: (err: Error) => {
                patchState(store, {
                  isLoading: false,
                  error: err.message || 'Login failed',
                });
              },
            })
          )
        )
      )
    ),

    // Register
    register: rxMethod<{ email: string; password: string; displayName: string }>(
      pipe(
        tap(() => patchState(store, { isLoading: true, error: null })),
        switchMap((data) =>
          authService.register(data).pipe(
            tapResponse({
              next: (response) => {
                localStorage.setItem('token', response.token);
                localStorage.setItem('user', JSON.stringify(response.user));

                patchState(store, {
                  token: response.token,
                  user: response.user,
                  isLoading: false,
                });

                router.navigate(['/dashboard']);
              },
              error: (err: Error) => {
                patchState(store, {
                  isLoading: false,
                  error: err.message || 'Registration failed',
                });
              },
            })
          )
        )
      )
    ),

    // Logout
    logout() {
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      patchState(store, initialState);
      router.navigate(['/login']);
    },

    clearError() {
      patchState(store, { error: null });
    },
  }))
);
```

## Key Concepts

### 1. `withState` - Base State
Defines the initial state shape and values.

### 2. `withEntities` - Entity Collections
Provides CRUD operations for arrays of items:
- `setAllEntities(items)` - Replace all
- `addEntity(item)` - Add one
- `updateEntity({ id, changes })` - Update one
- `removeEntity(id)` - Remove one
- `store.entities()` - Get array
- `store.entityMap()` - Get id-mapped object

### 3. `withComputed` - Derived State
Creates computed signals that automatically update when dependencies change.

### 4. `withMethods` - Actions
Defines synchronous and asynchronous methods. Use `rxMethod` for async operations with RxJS.

### 5. `patchState` - State Updates
Updates state immutably. Can be used with entity operations.

## Best Practices

1. **Single responsibility** - One store per feature domain
2. **Computed for derived state** - Don't store what you can compute
3. **rxMethod for async** - Handles subscriptions automatically
4. **tapResponse for errors** - Clean error handling pattern
5. **Initialize early** - Call init methods in APP_INITIALIZER or root components
6. **Type everything** - Full TypeScript support

## Next Steps

- [WebSocket Chat](./06-websocket-chat.md) - Real-time state with signals
- [REST API Implementation](./02-rest-api.md) - API services used by stores
