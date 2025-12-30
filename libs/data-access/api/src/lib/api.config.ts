import { InjectionToken } from '@angular/core';

// @REVIEW: Simplified API config - Single Source of Truth
// All URLs are relative - the SSR server/nginx proxies to backend
// Backend URL is configured via API_URL environment variable in server.ts

export interface ApiConfig {
  readonly baseUrl: string;
  readonly wsUrl: string;
  readonly grpcUrl: string;
}

export const API_CONFIG = new InjectionToken<ApiConfig>('API_CONFIG');

/**
 * Resolves WebSocket URL from current browser location
 * Returns empty string during SSR (WebSocket is client-side only)
 */
function resolveWsUrl(): string {
  if (typeof window !== 'undefined') {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    return `${protocol}//${window.location.host}`;
  }
  return '';
}

/**
 * Default API configuration using relative URLs
 * - /api/* requests are proxied by SSR server to API_URL
 * - /ws/* requests are proxied for WebSocket
 * - /grpc/* requests are proxied for gRPC-Web
 */
export const API_CONFIG_VALUE: ApiConfig = {
  baseUrl: '', // Relative URL - proxy handles routing
  wsUrl: resolveWsUrl(), // Derived from current host
  grpcUrl: '', // Relative URL - proxy handles routing
};

// @NOT-NEED: Legacy exports - kept for backward compatibility
export const defaultApiConfig = API_CONFIG_VALUE;
