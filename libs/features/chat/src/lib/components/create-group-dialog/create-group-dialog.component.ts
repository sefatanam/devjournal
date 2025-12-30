import { Component, inject, output } from '@angular/core';
import { CommonModule } from '@angular/common';
import {
  FormBuilder,
  FormGroup,
  ReactiveFormsModule,
  Validators,
} from '@angular/forms';
import { ChatStore } from '../../store/chat.store';
import { CreateGroupRequest } from '@devjournal/shared-models';

@Component({
  selector: 'lib-create-group-dialog',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule],
  template: `
    <div class="dialog-backdrop" (click)="close.emit()">
      <div class="dialog" (click)="$event.stopPropagation()">
        <header class="dialog-header">
          <h2>Create Study Group</h2>
          <button class="btn-close" (click)="close.emit()">&times;</button>
        </header>

        <form [formGroup]="form" (ngSubmit)="onSubmit()">
          <div class="form-group">
            <label for="name">Group Name</label>
            <input
              id="name"
              type="text"
              formControlName="name"
              placeholder="e.g., Angular Signals Study Group"
            />
            @if (form.get('name')?.invalid && form.get('name')?.touched) {
              <span class="error">Name is required (3-50 characters)</span>
            }
          </div>

          <div class="form-group">
            <label for="description">Description</label>
            <textarea
              id="description"
              formControlName="description"
              placeholder="What will your group focus on?"
              rows="3"
            ></textarea>
            @if (form.get('description')?.invalid && form.get('description')?.touched) {
              <span class="error">Description is required</span>
            }
          </div>

          <div class="form-row">
            <div class="form-group">
              <label for="maxMembers">Max Members</label>
              <input
                id="maxMembers"
                type="number"
                formControlName="maxMembers"
                min="2"
                max="100"
              />
            </div>

            <div class="form-group checkbox-group">
              <label>
                <input type="checkbox" formControlName="isPublic" />
                <span>Public Group</span>
              </label>
              <span class="hint">Anyone can discover and join public groups</span>
            </div>
          </div>

          <footer class="dialog-footer">
            <button type="button" class="btn-cancel" (click)="close.emit()">
              Cancel
            </button>
            <button
              type="submit"
              class="btn-create"
              [disabled]="form.invalid || store.isLoading()"
            >
              @if (store.isLoading()) {
                Creating...
              } @else {
                Create Group
              }
            </button>
          </footer>
        </form>
      </div>
    </div>
  `,
  styleUrl: './create-group-dialog.component.scss',
})
export class CreateGroupDialogComponent {
  readonly close = output<void>();
  readonly created = output<void>();

  readonly store = inject(ChatStore);
  private readonly fb = inject(FormBuilder);

  form: FormGroup = this.fb.group({
    name: ['', [Validators.required, Validators.minLength(3), Validators.maxLength(50)]],
    description: ['', [Validators.required]],
    isPublic: [true],
    maxMembers: [20, [Validators.min(2), Validators.max(100)]],
  });

  onSubmit(): void {
    if (this.form.valid) {
      const request: CreateGroupRequest = this.form.value;
      this.store.createGroup(request);
      this.created.emit();
      this.close.emit();
    }
  }
}
