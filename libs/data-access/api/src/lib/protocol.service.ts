import { Injectable, signal, computed } from '@angular/core';

/**
 * Available API protocols
 */
export type ApiProtocol = 'rest' | 'grpc';

/**
 * @REVIEW - Protocol toggle service
 * Allows switching between REST and gRPC API protocols
 */
@Injectable({ providedIn: 'root' })
export class ProtocolService {
  private readonly _protocol = signal<ApiProtocol>('rest');

  /**
   * Current API protocol
   */
  readonly protocol = this._protocol.asReadonly();

  /**
   * Whether gRPC is currently selected
   */
  readonly isGrpc = computed(() => this._protocol() === 'grpc');

  /**
   * Whether REST is currently selected
   */
  readonly isRest = computed(() => this._protocol() === 'rest');

  /**
   * Sets the API protocol
   */
  setProtocol(protocol: ApiProtocol): void {
    this._protocol.set(protocol);
    // Persist to localStorage
    localStorage.setItem('api_protocol', protocol);
  }

  /**
   * Toggles between REST and gRPC
   */
  toggle(): void {
    this.setProtocol(this._protocol() === 'rest' ? 'grpc' : 'rest');
  }

  /**
   * Initializes the protocol from localStorage
   */
  initialize(): void {
    const stored = localStorage.getItem('api_protocol') as ApiProtocol | null;
    if (stored && (stored === 'rest' || stored === 'grpc')) {
      this._protocol.set(stored);
    }
  }
}
