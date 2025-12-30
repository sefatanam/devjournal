import { RenderMode, ServerRoute } from '@angular/ssr';

export const serverRoutes: ServerRoute[] = [
  // Public routes - can be prerendered at build time
  {
    path: '',
    renderMode: RenderMode.Prerender,
  },
  {
    path: 'login',
    renderMode: RenderMode.Prerender,
  },
  {
    path: 'register',
    renderMode: RenderMode.Prerender,
  },
  // Authenticated routes - render on client (auth state only available in browser)
  {
    path: 'dashboard',
    renderMode: RenderMode.Client,
  },
  {
    path: 'journal',
    renderMode: RenderMode.Client,
  },
  {
    path: 'journal/:id',
    renderMode: RenderMode.Client,
  },
  {
    path: 'journal/:id/edit',
    renderMode: RenderMode.Client,
  },
  {
    path: 'journal/new',
    renderMode: RenderMode.Client,
  },
  {
    path: 'snippets',
    renderMode: RenderMode.Client,
  },
  {
    path: 'snippets/:id',
    renderMode: RenderMode.Client,
  },
  {
    path: 'snippets/:id/edit',
    renderMode: RenderMode.Client,
  },
  {
    path: 'snippets/new',
    renderMode: RenderMode.Client,
  },
  {
    path: 'progress',
    renderMode: RenderMode.Client,
  },
  {
    path: 'chat',
    renderMode: RenderMode.Client,
  },
  // Catch-all - render on client for safety
  {
    path: '**',
    renderMode: RenderMode.Client,
  },
];
