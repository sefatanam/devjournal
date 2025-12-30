import { Route } from '@angular/router';
import { authGuard, guestGuard } from '@devjournal/feature-auth';

export const appRoutes: Route[] = [
  {
    path: '',
    pathMatch: 'full',
    canActivate: [guestGuard],
    loadComponent: () =>
      import('./pages/landing/landing.component').then((m) => m.LandingComponent),
  },
  {
    path: 'login',
    canActivate: [guestGuard],
    loadComponent: () =>
      import('@devjournal/feature-auth').then((m) => m.LoginComponent),
  },
  {
    path: 'register',
    canActivate: [guestGuard],
    loadComponent: () =>
      import('@devjournal/feature-auth').then((m) => m.RegisterComponent),
  },
  {
    path: 'dashboard',
    canActivate: [authGuard],
    loadComponent: () =>
      import('./pages/dashboard/dashboard.component').then(
        (m) => m.DashboardComponent
      ),
  },
  {
    path: 'journal',
    canActivate: [authGuard],
    children: [
      {
        path: '',
        loadComponent: () =>
          import('@devjournal/feature-journal').then((m) => m.JournalListComponent),
      },
      {
        path: 'new',
        loadComponent: () =>
          import('@devjournal/feature-journal').then((m) => m.JournalFormComponent),
      },
      {
        path: ':id',
        loadComponent: () =>
          import('@devjournal/feature-journal').then((m) => m.JournalDetailComponent),
      },
      {
        path: ':id/edit',
        loadComponent: () =>
          import('@devjournal/feature-journal').then((m) => m.JournalFormComponent),
      },
    ],
  },
  {
    path: 'snippets',
    canActivate: [authGuard],
    children: [
      {
        path: '',
        loadComponent: () =>
          import('@devjournal/feature-snippets').then((m) => m.SnippetListComponent),
      },
      {
        path: 'new',
        loadComponent: () =>
          import('@devjournal/feature-snippets').then((m) => m.SnippetFormComponent),
      },
      {
        path: ':id',
        loadComponent: () =>
          import('@devjournal/feature-snippets').then((m) => m.SnippetDetailComponent),
      },
      {
        path: ':id/edit',
        loadComponent: () =>
          import('@devjournal/feature-snippets').then((m) => m.SnippetFormComponent),
      },
    ],
  },
  {
    path: 'progress',
    canActivate: [authGuard],
    loadComponent: () =>
      import('@devjournal/features-progress').then((m) => m.ProgressDashboardComponent),
  },
  {
    path: 'chat',
    canActivate: [authGuard],
    loadComponent: () =>
      import('./pages/chat/chat-page.component').then((m) => m.ChatPageComponent),
  },
  {
    path: '**',
    redirectTo: 'dashboard',
  },
];
