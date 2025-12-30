import {
  Component,
  inject,
  OnDestroy,
  input,
  effect,
  ElementRef,
  ViewChild,
  AfterViewChecked,
} from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ChatStore } from '../../store/chat.store';

@Component({
  selector: 'lib-chat-room',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="chat-room">
      <header class="room-header">
        <button class="btn-back" (click)="goBack()">
          &larr; Back
        </button>
        <div class="room-info">
          <h2>{{ store.currentRoom()?.name || 'Chat Room' }}</h2>
          <span class="connection-status" [class]="store.connectionStatus()">
            {{ store.connectionStatus() }}
          </span>
        </div>
        <div class="room-members">
          <span class="member-count">
            {{ store.members().length }} members
          </span>
        </div>
      </header>

      <div class="messages-container" #messagesContainer>
        @if (store.sortedMessages().length === 0) {
          <div class="empty-messages">
            <p>No messages yet. Start the conversation!</p>
          </div>
        } @else {
          @for (message of store.sortedMessages(); track message.id) {
            <div
              class="message"
              [class.system]="message.type !== 'message'"
              [class.own]="isOwnMessage(message.userId)"
            >
              @if (message.type === 'message') {
                <div class="message-header">
                  <span class="message-author">{{ message.userDisplayName }}</span>
                  <span class="message-time">
                    {{ formatTime(message.timestamp) }}
                  </span>
                </div>
                <div class="message-content">{{ message.content }}</div>
              } @else {
                <div class="system-message">
                  @switch (message.type) {
                    @case ('join') {
                      <span>{{ message.userDisplayName }} joined the chat</span>
                    }
                    @case ('leave') {
                      <span>{{ message.userDisplayName }} left the chat</span>
                    }
                    @default {
                      <span>{{ message.content }}</span>
                    }
                  }
                </div>
              }
            </div>
          }
        }
      </div>

      <footer class="message-input-container">
        @if (!store.isConnected()) {
          <div class="connecting-overlay">
            @if (store.connectionStatus() === 'connecting') {
              <span>Connecting...</span>
            } @else if (store.connectionStatus() === 'error') {
              <span>Connection error. Please try again.</span>
              <button class="btn-retry" (click)="reconnect()">Retry</button>
            } @else {
              <span>Disconnected</span>
            }
          </div>
        }
        <input
          type="text"
          [(ngModel)]="messageText"
          (keydown.enter)="sendMessage()"
          placeholder="Type a message..."
          [disabled]="!store.isConnected()"
          class="message-input"
        />
        <button
          class="btn-send"
          (click)="sendMessage()"
          [disabled]="!store.isConnected() || !messageText.trim()"
        >
          Send
        </button>
      </footer>
    </div>
  `,
  styleUrl: './chat-room.component.scss',
})
export class ChatRoomComponent implements OnDestroy, AfterViewChecked {
  readonly roomId = input.required<string>();
  readonly store = inject(ChatStore);

  @ViewChild('messagesContainer') messagesContainer!: ElementRef;

  messageText = '';
  private currentUserId = ''; // @REVIEW - should come from AuthStore
  private shouldScrollToBottom = false;

  constructor() {
    effect(() => {
      const id = this.roomId();
      if (id) {
        this.store.connectToRoom(id);
        this.store.loadMembers(id);
      }
    });
  }

  ngAfterViewChecked(): void {
    if (this.shouldScrollToBottom) {
      this.scrollToBottom();
      this.shouldScrollToBottom = false;
    }
  }

  ngOnDestroy(): void {
    this.store.disconnectFromRoom();
  }

  sendMessage(): void {
    if (this.messageText.trim() && this.store.isConnected()) {
      this.store.sendMessage(this.messageText);
      this.messageText = '';
      this.shouldScrollToBottom = true;
    }
  }

  goBack(): void {
    this.store.disconnectFromRoom();
    // Navigation should be handled by parent or router
  }

  reconnect(): void {
    const id = this.roomId();
    if (id) {
      this.store.connectToRoom(id);
    }
  }

  isOwnMessage(userId: string): boolean {
    return userId === this.currentUserId;
  }

  formatTime(date: Date): string {
    return new Date(date).toLocaleTimeString([], {
      hour: '2-digit',
      minute: '2-digit',
    });
  }

  private scrollToBottom(): void {
    if (this.messagesContainer) {
      const el = this.messagesContainer.nativeElement;
      el.scrollTop = el.scrollHeight;
    }
  }
}
