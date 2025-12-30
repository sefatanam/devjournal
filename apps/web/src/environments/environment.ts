// Development environment
// Points directly to Go backend - SSR server also proxies for fallback
export const environment = {
  production: false,
  apiUrl: 'http://localhost:8080', // Direct connection to Go API
  wsUrl: 'ws://localhost:8080', // Direct WebSocket connection
  grpcUrl: 'http://localhost:8081',
};
