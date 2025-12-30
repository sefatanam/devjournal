import { Injectable, inject, OnDestroy } from '@angular/core';
import { webSocket, WebSocketSubject } from 'rxjs/webSocket';
import {
  Observable,
  Subject,
  BehaviorSubject,
  timer,
  EMPTY,
} from 'rxjs';
import {
  catchError,
  filter,
  map,
  retry,
  takeUntil,
} from 'rxjs/operators';
import { ChatMessage } from '@devjournal/shared-models';
import { API_CONFIG } from './api.config';

export type ConnectionStatus = 'connecting' | 'connected' | 'disconnected' | 'error';

export interface WsMessage {
  type: string;
  content?: string;
  userId?: string;
  userDisplayName?: string;
  timestamp?: string;
  id?: string;
  roomId?: string;
}

const TOKEN_KEY = 'devjournal_token';

@Injectable({ providedIn: 'root' })
export class WebSocketService implements OnDestroy {
  private readonly config = inject(API_CONFIG);

  private socket$: WebSocketSubject<WsMessage> | null = null;
  private readonly destroy$ = new Subject<void>();
  private readonly reconnectDelay = 3000;
  private currentRoom: string | null = null;

  private readonly connectionStatus$ = new BehaviorSubject<ConnectionStatus>('disconnected');
  private readonly messages$ = new Subject<ChatMessage>();

  readonly status$ = this.connectionStatus$.asObservable();
  readonly isConnected$ = this.connectionStatus$.pipe(
    map((status) => status === 'connected')
  );

  connect(roomId: string): Observable<ChatMessage> {
    if (this.socket$ && this.currentRoom === roomId) {
      return this.messages$.asObservable();
    }

    this.disconnect();
    this.currentRoom = roomId;
    this.connectionStatus$.next('connecting');

    const token = this.getToken();
    if (!token) {
      this.connectionStatus$.next('error');
      return EMPTY;
    }

    // Handle both absolute URLs (dev) and relative URLs (prod via nginx)
    let wsUrl: string;
    if (this.config.baseUrl) {
      wsUrl = this.config.baseUrl.replace(/^http/, 'ws');
    } else if (typeof window !== 'undefined') {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      wsUrl = `${protocol}//${window.location.host}`;
    } else {
      wsUrl = 'ws://localhost:8080';
    }
    const url = `${wsUrl}/ws/chat/${roomId}?token=${token}`;

    this.socket$ = webSocket<WsMessage>({
      url,
      openObserver: {
        next: () => {
          this.connectionStatus$.next('connected');
        },
      },
      closeObserver: {
        next: () => {
          this.connectionStatus$.next('disconnected');
        },
      },
    });

    this.socket$
      .pipe(
        takeUntil(this.destroy$),
        retry({
          delay: (error, retryCount) => {
            if (retryCount >= 5) {
              this.connectionStatus$.next('error');
              return EMPTY;
            }
            this.connectionStatus$.next('connecting');
            return timer(this.reconnectDelay * retryCount);
          },
        }),
        map((msg) => this.mapToMessage(msg, roomId)),
        filter((msg): msg is ChatMessage => msg !== null),
        catchError((error) => {
          console.error('WebSocket error:', error);
          this.connectionStatus$.next('error');
          return EMPTY;
        })
      )
      .subscribe((message) => {
        this.messages$.next(message);
      });

    return this.messages$.asObservable();
  }

  sendMessage(content: string): void {
    if (this.socket$ && this.connectionStatus$.value === 'connected') {
      this.socket$.next({
        type: 'message',
        content,
      });
    }
  }

  disconnect(): void {
    if (this.socket$) {
      this.socket$.complete();
      this.socket$ = null;
    }
    this.currentRoom = null;
    this.connectionStatus$.next('disconnected');
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
    this.disconnect();
  }

  private mapToMessage(msg: WsMessage, roomId: string): ChatMessage | null {
    if (!msg.type) {
      return null;
    }

    const messageType = this.mapMessageType(msg.type);

    return {
      id: msg.id || crypto.randomUUID(),
      roomId: msg.roomId || roomId,
      userId: msg.userId || '',
      userDisplayName: msg.userDisplayName || 'Unknown',
      content: msg.content || '',
      timestamp: msg.timestamp ? new Date(msg.timestamp) : new Date(),
      type: messageType,
    };
  }

  private mapMessageType(
    type: string
  ): 'message' | 'join' | 'leave' | 'system' {
    switch (type) {
      case 'message':
        return 'message';
      case 'join':
        return 'join';
      case 'leave':
        return 'leave';
      default:
        return 'system';
    }
  }

  private getToken(): string | null {
    if (typeof window === 'undefined') {
      return null;
    }
    return localStorage.getItem(TOKEN_KEY);
  }
}
