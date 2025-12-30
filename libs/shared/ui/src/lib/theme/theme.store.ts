import { computed, inject, PLATFORM_ID, signal } from '@angular/core';
import { isPlatformBrowser } from '@angular/common';
import { signalStore, withState, withMethods, withComputed, withHooks } from '@ngrx/signals';

export type Theme = 'light' | 'dark' | 'system';

interface ThemeState {
  theme: Theme;
  resolvedTheme: 'light' | 'dark';
}

const THEME_KEY = 'devjournal_theme';

function getSystemTheme(): 'light' | 'dark' {
  if (typeof window === 'undefined') return 'dark';
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}

function getStoredTheme(): Theme {
  if (typeof window === 'undefined') return 'system';
  const stored = localStorage.getItem(THEME_KEY);
  if (stored === 'light' || stored === 'dark' || stored === 'system') {
    return stored;
  }
  return 'system';
}

function resolveTheme(theme: Theme): 'light' | 'dark' {
  if (theme === 'system') {
    return getSystemTheme();
  }
  return theme;
}

function applyTheme(resolvedTheme: 'light' | 'dark'): void {
  if (typeof document === 'undefined') return;
  document.documentElement.setAttribute('data-theme', resolvedTheme);
}

export const ThemeStore = signalStore(
  { providedIn: 'root' },
  withState<ThemeState>(() => {
    const theme = getStoredTheme();
    const resolvedTheme = resolveTheme(theme);
    return { theme, resolvedTheme };
  }),
  withComputed((store) => ({
    isDark: computed(() => store.resolvedTheme() === 'dark'),
    isLight: computed(() => store.resolvedTheme() === 'light'),
    isSystem: computed(() => store.theme() === 'system'),
  })),
  withMethods((store, platformId = inject(PLATFORM_ID)) => ({
    setTheme(theme: Theme): void {
      if (!isPlatformBrowser(platformId)) return;
      
      const resolvedTheme = resolveTheme(theme);
      localStorage.setItem(THEME_KEY, theme);
      applyTheme(resolvedTheme);
      
      // Update state using patchState from ngrx/signals
      (store as any).theme.set(theme);
      (store as any).resolvedTheme.set(resolvedTheme);
    },
    
    toggle(): void {
      if (!isPlatformBrowser(platformId)) return;
      
      const currentResolved = store.resolvedTheme();
      const newTheme: Theme = currentResolved === 'dark' ? 'light' : 'dark';
      this.setTheme(newTheme);
    },
    
    initializeTheme(): void {
      if (!isPlatformBrowser(platformId)) return;
      
      // Apply the initial theme
      applyTheme(store.resolvedTheme());
      
      // Listen for system theme changes
      const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
      mediaQuery.addEventListener('change', (e) => {
        if (store.theme() === 'system') {
          const newResolved = e.matches ? 'dark' : 'light';
          applyTheme(newResolved);
          (store as any).resolvedTheme.set(newResolved);
        }
      });
    },
  })),
  withHooks({
    onInit(store) {
      store.initializeTheme();
    },
  })
);
