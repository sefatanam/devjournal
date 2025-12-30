import { computed, inject } from '@angular/core';
import { Router } from '@angular/router';
import {
  patchState,
  signalStore,
  withComputed,
  withMethods,
  withState,
} from '@ngrx/signals';
import { rxMethod } from '@ngrx/signals/rxjs-interop';
import { tapResponse } from '@ngrx/operators';
import { pipe, switchMap, tap } from 'rxjs';
import {
  AuthState,
  LoginRequest,
  RegisterRequest,
  User,
} from '@devjournal/shared-models';
import { AuthApiService } from '@devjournal/data-access-api';

const TOKEN_KEY = 'devjournal_token';
const USER_KEY = 'devjournal_user';

const initialState: AuthState = {
  user: null,
  token: null,
  isAuthenticated: false,
  isLoading: false,
  error: null,
};

function loadStoredAuth(): Partial<AuthState> {
  if (typeof window === 'undefined') {
    return {};
  }

  const token = localStorage.getItem(TOKEN_KEY);
  const userJson = localStorage.getItem(USER_KEY);

  if (token && userJson) {
    try {
      const user = JSON.parse(userJson) as User;
      return {
        user,
        token,
        isAuthenticated: true,
      };
    } catch {
      localStorage.removeItem(TOKEN_KEY);
      localStorage.removeItem(USER_KEY);
    }
  }

  return {};
}

export const AuthStore = signalStore(
  { providedIn: 'root' },
  withState(() => ({
    ...initialState,
    ...loadStoredAuth(),
  })),
  withComputed((store) => ({
    userDisplayName: computed(() => store.user()?.displayName ?? 'Guest'),
    userEmail: computed(() => store.user()?.email ?? ''),
  })),
  withMethods((store, authApi = inject(AuthApiService), router = inject(Router)) => ({
    login: rxMethod<LoginRequest>(
      pipe(
        tap(() => patchState(store, { isLoading: true, error: null })),
        switchMap((credentials) =>
          authApi.login(credentials).pipe(
            tapResponse({
              next: (response) => {
                localStorage.setItem(TOKEN_KEY, response.token);
                localStorage.setItem(USER_KEY, JSON.stringify(response.user));
                patchState(store, {
                  user: response.user,
                  token: response.token,
                  isAuthenticated: true,
                  isLoading: false,
                  error: null,
                });
                router.navigate(['/dashboard']);
              },
              error: (error: Error) => {
                patchState(store, {
                  isLoading: false,
                  error: error.message || 'Login failed',
                });
              },
            })
          )
        )
      )
    ),

    register: rxMethod<RegisterRequest>(
      pipe(
        tap(() => patchState(store, { isLoading: true, error: null })),
        switchMap((data) =>
          authApi.register(data).pipe(
            tapResponse({
              next: (response) => {
                localStorage.setItem(TOKEN_KEY, response.token);
                localStorage.setItem(USER_KEY, JSON.stringify(response.user));
                patchState(store, {
                  user: response.user,
                  token: response.token,
                  isAuthenticated: true,
                  isLoading: false,
                  error: null,
                });
                router.navigate(['/dashboard']);
              },
              error: (error: Error) => {
                patchState(store, {
                  isLoading: false,
                  error: error.message || 'Registration failed',
                });
              },
            })
          )
        )
      )
    ),

    logout(): void {
      localStorage.removeItem(TOKEN_KEY);
      localStorage.removeItem(USER_KEY);
      patchState(store, initialState);
      router.navigate(['/login']);
    },

    clearError(): void {
      patchState(store, { error: null });
    },

    getToken(): string | null {
      return store.token();
    },
  }))
);
