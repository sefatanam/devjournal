import { Component, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import {
  ChatRoomListComponent,
  ChatRoomComponent,
  CreateGroupDialogComponent,
  ChatStore,
} from '@devjournal/features-chat';
import { StudyGroup } from '@devjournal/shared-models';

@Component({
  selector: 'app-chat-page',
  standalone: true,
  imports: [
    CommonModule,
    ChatRoomListComponent,
    ChatRoomComponent,
    CreateGroupDialogComponent,
  ],
  template: `
    <div class="chat-page">
      @if (!selectedRoom()) {
        <lib-chat-room-list
          (selectRoom)="onSelectRoom($event)"
          (createGroup)="showCreateDialog.set(true)"
        />
      } @else {
        <lib-chat-room
          [roomId]="selectedRoom()!.id"
        />
        <button class="btn-back-float" (click)="onBackToList()">
          ← Back to Groups
        </button>
      }

      @if (showCreateDialog()) {
        <lib-create-group-dialog
          (created)="onGroupCreated()"
          (close)="showCreateDialog.set(false)"
        />
      }
    </div>
  `,
  styleUrl: './chat-page.component.scss',
})
export class ChatPageComponent {
  private readonly store = inject(ChatStore);

  readonly selectedRoom = signal<StudyGroup | null>(null);
  readonly showCreateDialog = signal(false);

  onSelectRoom(group: StudyGroup): void {
    this.selectedRoom.set(group);
  }

  onBackToList(): void {
    this.store.disconnectFromRoom();
    this.selectedRoom.set(null);
  }

  onGroupCreated(): void {
    this.showCreateDialog.set(false);
    // Refresh the groups list
    this.store.loadMyGroups();
  }
}
