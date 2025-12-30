// API Configuration
export * from './lib/api.config';

// REST API Services
export * from './lib/auth-api.service';
export * from './lib/journal-api.service';
export * from './lib/snippet-api.service';
export * from './lib/progress-api.service';
export * from './lib/studygroup-api.service';

// WebSocket
export * from './lib/websocket.service';

// @REVIEW - gRPC Services temporarily disabled due to connect-es version mismatch
// export * from './lib/grpc-client.service';
// export * from './lib/journal-grpc.service';
// export * from './lib/snippet-grpc.service';

// Protocol Toggle
export * from './lib/protocol.service';

// Unified Services (REST-only for now)
export * from './lib/journal-unified.service';
export * from './lib/snippet-unified.service';
