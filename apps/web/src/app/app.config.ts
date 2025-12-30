import {
  ApplicationConfig,
  provideBrowserGlobalErrorListeners,
} from '@angular/core';
import {
  provideRouter,
  withComponentInputBinding,
  withViewTransitions,
} from '@angular/router';
import {
  provideHttpClient,
  withInterceptors,
  withFetch,
} from '@angular/common/http';
import {
  provideClientHydration,
  withEventReplay,
} from '@angular/platform-browser';
import { appRoutes } from './app.routes';
import { authInterceptor } from '@devjournal/feature-auth';
import { API_CONFIG, API_CONFIG_VALUE } from '@devjournal/data-access-api';

// @REVIEW: Simplified app config - uses relative URLs
// All API routing is handled by SSR server proxy (reads API_URL from env)
export const appConfig: ApplicationConfig = {
  providers: [
    provideClientHydration(withEventReplay()),
    provideBrowserGlobalErrorListeners(),
    provideRouter(
      appRoutes,
      withComponentInputBinding(),
      withViewTransitions(),
    ),
    provideHttpClient(withFetch(), withInterceptors([authInterceptor])),
    {
      provide: API_CONFIG,
      useValue: API_CONFIG_VALUE,
    },
  ],
};
