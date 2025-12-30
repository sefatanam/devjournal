import { Injectable, inject, effect } from '@angular/core';
import { createClient, Client, Transport, Interceptor } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';
import {
  JournalService,
  SnippetService,
} from '@devjournal/shared-proto';
import { API_CONFIG } from './api.config';

/**
 * @REVIEW - gRPC client service using Connect RPC
 * Provides typed clients for all gRPC services
 */
@Injectable({ providedIn: 'root' })
export class GrpcClientService {
  private readonly config = inject(API_CONFIG);
  private transport: Transport | null = null;
  private token: string | null = null;

  constructor() {
    // Initialize token from localStorage
    this.token = localStorage.getItem('auth_token');
  }

  /**
   * Sets the authentication token for gRPC requests
   */
  setToken(token: string | null): void {
    this.token = token;
    // Reset transport to recreate with new token
    this.transport = null;
  }

  /**
   * Auth interceptor that adds the Bearer token to requests
   */
  private authInterceptor: Interceptor = (next) => async (req) => {
    // Always get the latest token from localStorage
    const currentToken = this.token || localStorage.getItem('auth_token');
    if (currentToken) {
      req.header.set('Authorization', `Bearer ${currentToken}`);
    }
    return next(req);
  };

  /**
   * Gets or creates the Connect transport with auth headers
   */
  private getTransport(): Transport {
    if (!this.transport) {
      this.transport = createConnectTransport({
        baseUrl: this.config.grpcUrl,
        interceptors: [this.authInterceptor],
      });
    }
    return this.transport;
  }

  /**
   * Gets a typed client for the JournalService
   */
  get journalClient(): Client<typeof JournalService> {
    return createClient(JournalService, this.getTransport());
  }

  /**
   * Gets a typed client for the SnippetService
   */
  get snippetClient(): Client<typeof SnippetService> {
    return createClient(SnippetService, this.getTransport());
  }
}
