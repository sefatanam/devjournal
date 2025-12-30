import {
  ApplicationConfig,
  provideBrowserGlobalErrorListeners,
} from '@angular/core';
import { provideRouter, withComponentInputBinding, withViewTransitions } from '@angular/router';
import { provideHttpClient, withInterceptors, withFetch } from '@angular/common/http';
import {
  provideClientHydration,
  withEventReplay,
} from '@angular/platform-browser';
import { appRoutes } from './app.routes';
import { authInterceptor } from '@devjournal/feature-auth';
import { API_CONFIG } from '@devjournal/data-access-api';
import { environment } from '../environments/environment';

// Compute WebSocket URL based on environment
function getWsUrl(): string {
  if (environment.wsUrl) {
    return environment.wsUrl;
  }
  // For production, derive WS URL from current location
  if (typeof window !== 'undefined') {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    return `${protocol}//${window.location.host}`;
  }
  return '';
}

export const appConfig: ApplicationConfig = {
  providers: [
    provideClientHydration(withEventReplay()),
    provideBrowserGlobalErrorListeners(),
    provideRouter(appRoutes, withComponentInputBinding(), withViewTransitions()),
    provideHttpClient(
      withFetch(),
      withInterceptors([authInterceptor])
    ),
    {
      provide: API_CONFIG,
      useValue: {
        baseUrl: environment.apiUrl,
        wsUrl: getWsUrl(),
        grpcUrl: environment.grpcUrl,
      },
    },
  ],
};
