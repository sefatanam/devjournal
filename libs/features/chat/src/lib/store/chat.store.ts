import { computed, inject } from '@angular/core';
import {
  patchState,
  signalStore,
  type,
  withComputed,
  withMethods,
  withState,
} from '@ngrx/signals';
import {
  addEntity,
  removeEntity,
  setAllEntities,
  withEntities,
  EntityId,
} from '@ngrx/signals/entities';
import { rxMethod } from '@ngrx/signals/rxjs-interop';
import { tapResponse } from '@ngrx/operators';
import { pipe, switchMap, tap, Subject } from 'rxjs';
import { takeUntil } from 'rxjs/operators';
import {
  ChatMessage,
  StudyGroup,
  StudyGroupMember,
  CreateGroupRequest,
} from '@devjournal/shared-models';
import {
  StudyGroupApiService,
  WebSocketService,
  ConnectionStatus,
} from '@devjournal/data-access-api';

interface ChatState {
  isLoading: boolean;
  error: string | null;
  currentRoomId: string | null;
  connectionStatus: ConnectionStatus;
  members: StudyGroupMember[];
}

const initialState: ChatState = {
  isLoading: false,
  error: null,
  currentRoomId: null,
  connectionStatus: 'disconnected',
  members: [],
};

export const ChatStore = signalStore(
  { providedIn: 'root' },
  withState(initialState),
  withEntities({ entity: type<StudyGroup>(), collection: 'group' }),
  withEntities({ entity: type<ChatMessage>(), collection: 'message' }),
  withComputed((store) => ({
    currentRoom: computed(() => {
      const id = store.currentRoomId();
      return id ? store.groupEntityMap()[id] : null;
    }),
    isConnected: computed(() => store.connectionStatus() === 'connected'),
    sortedMessages: computed(() =>
      [...store.messageEntities()].sort(
        (a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()
      )
    ),
    myGroups: computed(() => store.groupEntities()),
    publicGroups: computed(() =>
      store.groupEntities().filter((g) => g.isPublic)
    ),
  })),
  withMethods(
    (
      store,
      groupApi = inject(StudyGroupApiService),
      wsService = inject(WebSocketService)
    ) => {
      const disconnect$ = new Subject<void>();

      return {
        loadMyGroups: rxMethod<void>(
          pipe(
            tap(() => patchState(store, { isLoading: true, error: null })),
            switchMap(() =>
              groupApi.list().pipe(
                tapResponse({
                  next: (groups) => {
                    patchState(
                      store,
                      setAllEntities(groups, { collection: 'group' }),
                      { isLoading: false }
                    );
                  },
                  error: (error: Error) => {
                    patchState(store, {
                      isLoading: false,
                      error: error.message || 'Failed to load groups',
                    });
                  },
                })
              )
            )
          )
        ),

        loadPublicGroups: rxMethod<void>(
          pipe(
            tap(() => patchState(store, { isLoading: true, error: null })),
            switchMap(() =>
              groupApi.listPublic().pipe(
                tapResponse({
                  next: (groups) => {
                    patchState(
                      store,
                      setAllEntities(groups, { collection: 'group' }),
                      { isLoading: false }
                    );
                  },
                  error: (error: Error) => {
                    patchState(store, {
                      isLoading: false,
                      error: error.message || 'Failed to load public groups',
                    });
                  },
                })
              )
            )
          )
        ),

        createGroup: rxMethod<CreateGroupRequest>(
          pipe(
            tap(() => patchState(store, { isLoading: true, error: null })),
            switchMap((request) =>
              groupApi.create(request).pipe(
                tapResponse({
                  next: (group) => {
                    patchState(
                      store,
                      addEntity(group, { collection: 'group' }),
                      { isLoading: false }
                    );
                  },
                  error: (error: Error) => {
                    patchState(store, {
                      isLoading: false,
                      error: error.message || 'Failed to create group',
                    });
                  },
                })
              )
            )
          )
        ),

        joinGroup: rxMethod<string>(
          pipe(
            tap(() => patchState(store, { isLoading: true, error: null })),
            switchMap((groupId) =>
              groupApi.join(groupId).pipe(
                tapResponse({
                  next: () => {
                    patchState(store, { isLoading: false });
                  },
                  error: (error: Error) => {
                    patchState(store, {
                      isLoading: false,
                      error: error.message || 'Failed to join group',
                    });
                  },
                })
              )
            )
          )
        ),

        leaveGroup: rxMethod<string>(
          pipe(
            tap(() => patchState(store, { isLoading: true, error: null })),
            switchMap((groupId) =>
              groupApi.leave(groupId).pipe(
                tapResponse({
                  next: () => {
                    patchState(
                      store,
                      removeEntity(groupId as EntityId, { collection: 'group' }),
                      { isLoading: false }
                    );
                  },
                  error: (error: Error) => {
                    patchState(store, {
                      isLoading: false,
                      error: error.message || 'Failed to leave group',
                    });
                  },
                })
              )
            )
          )
        ),

        deleteGroup: rxMethod<string>(
          pipe(
            tap(() => patchState(store, { isLoading: true, error: null })),
            switchMap((groupId) =>
              groupApi.delete(groupId).pipe(
                tapResponse({
                  next: () => {
                    patchState(
                      store,
                      removeEntity(groupId as EntityId, { collection: 'group' }),
                      { isLoading: false }
                    );
                  },
                  error: (error: Error) => {
                    patchState(store, {
                      isLoading: false,
                      error: error.message || 'Failed to delete group',
                    });
                  },
                })
              )
            )
          )
        ),

        loadMembers: rxMethod<string>(
          pipe(
            tap(() => patchState(store, { isLoading: true, error: null })),
            switchMap((groupId) =>
              groupApi.getMembers(groupId).pipe(
                tapResponse({
                  next: (members) => {
                    patchState(store, { members, isLoading: false });
                  },
                  error: (error: Error) => {
                    patchState(store, {
                      isLoading: false,
                      error: error.message || 'Failed to load members',
                    });
                  },
                })
              )
            )
          )
        ),

        connectToRoom(roomId: string): void {
          disconnect$.next();
          patchState(store, {
            currentRoomId: roomId,
            connectionStatus: 'connecting',
          });
          patchState(store, setAllEntities([] as ChatMessage[], { collection: 'message' }));

          wsService.status$.pipe(takeUntil(disconnect$)).subscribe((status) => {
            patchState(store, { connectionStatus: status });
          });

          wsService
            .connect(roomId)
            .pipe(takeUntil(disconnect$))
            .subscribe((message) => {
              patchState(store, addEntity(message, { collection: 'message' }));
            });
        },

        disconnectFromRoom(): void {
          disconnect$.next();
          wsService.disconnect();
          patchState(store, {
            currentRoomId: null,
            connectionStatus: 'disconnected',
            members: [],
          });
          patchState(store, setAllEntities([] as ChatMessage[], { collection: 'message' }));
        },

        sendMessage(content: string): void {
          if (store.connectionStatus() === 'connected' && content.trim()) {
            wsService.sendMessage(content.trim());
          }
        },

        clearError(): void {
          patchState(store, { error: null });
        },
      };
    }
  )
);
