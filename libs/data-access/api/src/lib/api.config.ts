import { InjectionToken } from '@angular/core';

export interface ApiConfig {
  baseUrl: string;
  wsUrl: string;
  grpcUrl: string;
}

export const API_CONFIG = new InjectionToken<ApiConfig>('API_CONFIG');

export const defaultApiConfig: ApiConfig = {
  baseUrl: 'http://localhost:8080',
  wsUrl: 'ws://localhost:8080',
  grpcUrl: 'http://localhost:8081',
};
