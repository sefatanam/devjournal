import { Component, inject, OnInit, output } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ChatStore } from '../../store/chat.store';
import { StudyGroup } from '@devjournal/shared-models';
import { AuthStore } from '@devjournal/feature-auth';

@Component({
  selector: 'lib-chat-room-list',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="chat-room-list">
      <header class="list-header">
        <h2>Study Groups</h2>
        <button class="btn-create" (click)="createGroup.emit()">
          + New Group
        </button>
      </header>

      @if (store.isLoading()) {
        <div class="loading">Loading groups...</div>
      }

      @if (store.error()) {
        <div class="error">{{ store.error() }}</div>
      }

      <div class="groups-container">
        <section class="groups-section">
          <h3>My Groups</h3>
          @if (store.myGroups().length === 0) {
            <p class="empty-state">No groups yet. Join or create one!</p>
          } @else {
            <ul class="group-list">
              @for (group of store.myGroups(); track group.id) {
                <li class="group-item" >
                  <button class="group-item-btn" (click)="selectRoom.emit(group)">
                  <div class="group-info">
                    <span class="group-name">{{ group.name }}</span>
                    <span class="group-description">{{ group.description }}</span>
                  </div>
                  <div class="group-actions">
                    @if (group.isPublic) {
                      <span class="badge public">Public</span>
                    } @else {
                      <span class="badge private">Private</span>
                    }
                    @if (isOwner(group)) {
                      <button
                        class="btn-delete"
                        (click)="deleteGroup(group.id); $event.stopPropagation()"
                        title="Delete group"
                      >
                        🗑️
                      </button>
                    }
                  </div>
                  </button>
                </li>
              }
            </ul>
          }
        </section>

        <section class="groups-section">
          <h3>Discover Public Groups</h3>
          <button class="btn-discover" (click)="loadPublicGroups()">
            Browse Public Groups
          </button>
          @if (store.publicGroups().length > 0) {
            <ul class="group-list">
              @for (group of store.publicGroups(); track group.id) {
                <li class="group-item public-group">
                  <div class="group-info">
                    <span class="group-name">{{ group.name }}</span>
                    <span class="group-description">{{ group.description }}</span>
                  </div>
                  <button
                    class="btn-join"
                    (click)="joinGroup(group.id); $event.stopPropagation()"
                  >
                    Join
                  </button>
                </li>
              }
            </ul>
          }
        </section>
      </div>
    </div>
  `,
  styleUrl: './chat-room-list.component.scss',
})
export class ChatRoomListComponent implements OnInit {
  readonly store = inject(ChatStore);
  private readonly authStore = inject(AuthStore);

  readonly selectRoom = output<StudyGroup>();
  readonly createGroup = output<void>();

  ngOnInit(): void {
    this.store.loadMyGroups();
  }

  loadPublicGroups(): void {
    this.store.loadPublicGroups();
  }

  joinGroup(groupId: string): void {
    this.store.joinGroup(groupId);
  }

  isOwner(group: StudyGroup): boolean {
    const userId = this.authStore.user()?.id;
    return userId === group.createdBy;
  }

  deleteGroup(groupId: string): void {
    if (confirm('Are you sure you want to delete this group? This action cannot be undone.')) {
      this.store.deleteGroup(groupId);
    }
  }
}
