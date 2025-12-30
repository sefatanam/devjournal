import { computed, inject } from '@angular/core';
import {
  patchState,
  signalStore,
  withComputed,
  withMethods,
  withState,
} from '@ngrx/signals';
import { rxMethod } from '@ngrx/signals/rxjs-interop';
import { tapResponse } from '@ngrx/operators';
import { pipe, switchMap, tap, forkJoin } from 'rxjs';
import { LearningProgress, ProgressSummary } from '@devjournal/shared-models';
import { ProgressApiService } from '@devjournal/data-access-api';

// @REVIEW - Phase 7: Progress Signal Store
interface ProgressState {
  isLoading: boolean;
  error: string | null;
  summary: ProgressSummary | null;
  todayProgress: LearningProgress | null;
  weeklyProgress: LearningProgress[];
  monthlyProgress: LearningProgress[];
  currentStreak: number;
}

const initialState: ProgressState = {
  isLoading: false,
  error: null,
  summary: null,
  todayProgress: null,
  weeklyProgress: [],
  monthlyProgress: [],
  currentStreak: 0,
};

export const ProgressStore = signalStore(
  { providedIn: 'root' },
  withState(initialState),
  withComputed((store) => ({
    // Total activity for today
    todayTotal: computed(() => {
      const today = store.todayProgress();
      if (!today) return 0;
      return today.entriesCount + today.snippetsCount;
    }),

    // Weekly totals
    weeklyTotals: computed(() => {
      const weekly = store.weeklyProgress() ?? [];
      return {
        entries: weekly.reduce((sum, p) => sum + (p.entriesCount ?? 0), 0),
        snippets: weekly.reduce((sum, p) => sum + (p.snippetsCount ?? 0), 0),
        activeDays: weekly.filter(p => (p.entriesCount ?? 0) > 0 || (p.snippetsCount ?? 0) > 0).length,
      };
    }),

    // Monthly totals
    monthlyTotals: computed(() => {
      const monthly = store.monthlyProgress() ?? [];
      return {
        entries: monthly.reduce((sum, p) => sum + (p.entriesCount ?? 0), 0),
        snippets: monthly.reduce((sum, p) => sum + (p.snippetsCount ?? 0), 0),
        activeDays: monthly.filter(p => (p.entriesCount ?? 0) > 0 || (p.snippetsCount ?? 0) > 0).length,
      };
    }),

    // Activity chart data (last 7 days)
    weeklyChartData: computed(() => {
      const weekly = store.weeklyProgress() ?? [];
      return weekly.map(p => ({
        date: new Date(p.date),
        entries: p.entriesCount ?? 0,
        snippets: p.snippetsCount ?? 0,
        total: (p.entriesCount ?? 0) + (p.snippetsCount ?? 0),
      }));
    }),

    // Is the user active today?
    isActiveToday: computed(() => {
      const today = store.todayProgress();
      return today ? (today.entriesCount > 0 || today.snippetsCount > 0) : false;
    }),

    // Has data been loaded?
    hasData: computed(() => store.summary() !== null),
  })),
  withMethods((store, progressApi = inject(ProgressApiService)) => ({
    // Load all progress data at once
    loadAll: rxMethod<void>(
      pipe(
        tap(() => patchState(store, { isLoading: true, error: null })),
        switchMap(() =>
          forkJoin({
            summary: progressApi.getSummary(),
            today: progressApi.getToday(),
            weekly: progressApi.getWeekly(),
            streak: progressApi.getStreak(),
          }).pipe(
            tapResponse({
              next: ({ summary, today, weekly, streak }) => {
                patchState(store, {
                  summary,
                  todayProgress: today,
                  weeklyProgress: weekly?.progress ?? [],
                  currentStreak: streak?.currentStreak ?? 0,
                  isLoading: false,
                });
              },
              error: (error: Error) => {
                patchState(store, {
                  isLoading: false,
                  error: error.message || 'Failed to load progress data',
                });
              },
            })
          )
        )
      )
    ),

    // Load summary only
    loadSummary: rxMethod<void>(
      pipe(
        tap(() => patchState(store, { isLoading: true, error: null })),
        switchMap(() =>
          progressApi.getSummary().pipe(
            tapResponse({
              next: (summary) => {
                patchState(store, { summary, isLoading: false });
              },
              error: (error: Error) => {
                patchState(store, {
                  isLoading: false,
                  error: error.message || 'Failed to load summary',
                });
              },
            })
          )
        )
      )
    ),

    // Load today's progress
    loadToday: rxMethod<void>(
      pipe(
        tap(() => patchState(store, { isLoading: true, error: null })),
        switchMap(() =>
          progressApi.getToday().pipe(
            tapResponse({
              next: (todayProgress) => {
                patchState(store, { todayProgress, isLoading: false });
              },
              error: (error: Error) => {
                patchState(store, {
                  isLoading: false,
                  error: error.message || 'Failed to load today\'s progress',
                });
              },
            })
          )
        )
      )
    ),

    // Load weekly progress
    loadWeekly: rxMethod<void>(
      pipe(
        tap(() => patchState(store, { isLoading: true, error: null })),
        switchMap(() =>
          progressApi.getWeekly().pipe(
            tapResponse({
              next: (response) => {
                patchState(store, {
                  weeklyProgress: response?.progress ?? [],
                  isLoading: false,
                });
              },
              error: (error: Error) => {
                patchState(store, {
                  isLoading: false,
                  error: error.message || 'Failed to load weekly progress',
                });
              },
            })
          )
        )
      )
    ),

    // Load monthly progress
    loadMonthly: rxMethod<void>(
      pipe(
        tap(() => patchState(store, { isLoading: true, error: null })),
        switchMap(() =>
          progressApi.getMonthly().pipe(
            tapResponse({
              next: (response) => {
                patchState(store, {
                  monthlyProgress: response?.progress ?? [],
                  isLoading: false,
                });
              },
              error: (error: Error) => {
                patchState(store, {
                  isLoading: false,
                  error: error.message || 'Failed to load monthly progress',
                });
              },
            })
          )
        )
      )
    ),

    // Load current streak
    loadStreak: rxMethod<void>(
      pipe(
        tap(() => patchState(store, { isLoading: true, error: null })),
        switchMap(() =>
          progressApi.getStreak().pipe(
            tapResponse({
              next: (response) => {
                patchState(store, {
                  currentStreak: response?.currentStreak ?? 0,
                  isLoading: false,
                });
              },
              error: (error: Error) => {
                patchState(store, {
                  isLoading: false,
                  error: error.message || 'Failed to load streak',
                });
              },
            })
          )
        )
      )
    ),

    // Refresh all data (e.g., after creating a new entry)
    refresh(): void {
      this.loadAll();
    },

    clearError(): void {
      patchState(store, { error: null });
    },
  }))
);
