import {
  AngularNodeAppEngine,
  createNodeRequestHandler,
  isMainModule,
  writeResponseToNodeResponse,
} from '@angular/ssr/node';
import express from 'express';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { createProxyMiddleware } from 'http-proxy-middleware';

// @REVIEW: SSR Server - Single Source of Truth for API routing
// =============================================================
// Environment Variables (from .env or Railway):
//   - PORT: Server port (default: 4000)
//   - API_URL: Backend API URL (default: http://localhost:8080)
// =============================================================

const serverDistFolder = dirname(fileURLToPath(import.meta.url));
const browserDistFolder = resolve(serverDistFolder, '../browser');

const app = express();
const angularApp = new AngularNodeAppEngine();

// =============================================================
// API Configuration - Read from environment
// =============================================================
const API_URL = process.env['API_URL'] || 'http://localhost:8080';

console.log(`[SSR] API proxy target: ${API_URL}`);

// =============================================================
// Proxy Configuration - Routes all /api/* and /ws/* to backend
// =============================================================

// @REVIEW: Using pathFilter approach to preserve full path
// Express app.use('/api', ...) strips the /api prefix before passing to middleware
// Using pathFilter with app.use() on root keeps the full path

const apiProxy = createProxyMiddleware({
  target: API_URL,
  changeOrigin: true,
  pathFilter: '/api/**',
});

const wsProxy = createProxyMiddleware({
  target: API_URL,
  changeOrigin: true,
  ws: true,
  pathFilter: '/ws/**',
});

const GRPC_URL = process.env['GRPC_URL'] || API_URL.replace(':8080', ':8081');
const grpcProxy = createProxyMiddleware({
  target: GRPC_URL,
  changeOrigin: true,
  pathFilter: '/grpc/**',
});

// Apply proxies before static files
app.use(apiProxy);
app.use(wsProxy);
app.use(grpcProxy);

// =============================================================
// Static Files & Angular SSR
// =============================================================

/**
 * Serve static files from /browser
 */
app.use(
  express.static(browserDistFolder, {
    maxAge: '1y',
    index: false,
    redirect: false,
  }),
);

/**
 * Handle all other requests by rendering the Angular application.
 */
app.use('/**', (req, res, next) => {
  angularApp
    .handle(req)
    .then((response) =>
      response ? writeResponseToNodeResponse(response, res) : next(),
    )
    .catch(next);
});

// =============================================================
// Server Startup
// =============================================================

/**
 * Start the server if this module is the main entry point.
 * PORT is set by Railway automatically, defaults to 4000 for local dev.
 */
if (isMainModule(import.meta.url) || process.env['pm_id']) {
  const port = process.env['PORT'] || 4000;
  app.listen(port, () => {
    console.log(`[SSR] Server listening on http://localhost:${port}`);
    console.log(`[SSR] API requests proxied to ${API_URL}`);
  });
}

/**
 * Request handler used by the Angular CLI (for dev-server and during build).
 */
export const reqHandler = createNodeRequestHandler(app);
