// Production environment - uses nginx proxy
export const environment = {
  production: true,
  apiUrl: '', // Empty = relative URLs, nginx proxies /api/ to backend
  wsUrl: '', // Empty = uses current host for WebSocket
  grpcUrl: '/grpc', // nginx proxies /grpc/ to Envoy
};
