import {
  ChangeDetectionStrategy,
  Component,
  computed,
  effect,
  inject,
  input,
  OnInit,
  signal,
} from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router, RouterLink } from '@angular/router';
import {
  FormBuilder,
  FormGroup,
  ReactiveFormsModule,
  Validators,
} from '@angular/forms';
import { JournalStore } from '../../store/journal.store';
import { JournalMood } from '@devjournal/shared-models';

@Component({
  selector: 'lib-journal-form',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, RouterLink],
  templateUrl: './journal-form.component.html',
  styleUrl: './journal-form.component.scss',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class JournalFormComponent implements OnInit {
  private readonly fb = inject(FormBuilder);
  private readonly router = inject(Router);
  protected readonly store = inject(JournalStore);

  readonly entryId = input<string | undefined>(undefined, { alias: 'id' });

  protected readonly isEditMode = computed(() => !!this.entryId());
  protected readonly pageTitle = computed(() =>
    this.isEditMode() ? 'Edit Entry' : 'New Entry'
  );

  protected readonly tagInput = signal('');

  protected readonly moods: { value: JournalMood; label: string; emoji: string }[] = [
    { value: 'excited', label: 'Excited', emoji: '🤩' },
    { value: 'happy', label: 'Happy', emoji: '😊' },
    { value: 'neutral', label: 'Neutral', emoji: '😐' },
    { value: 'frustrated', label: 'Frustrated', emoji: '😤' },
    { value: 'tired', label: 'Tired', emoji: '😴' },
  ];

  protected readonly form: FormGroup = this.fb.group({
    title: ['', [Validators.required, Validators.minLength(3), Validators.maxLength(255)]],
    content: ['', [Validators.required, Validators.minLength(10)]],
    mood: ['neutral' as JournalMood, [Validators.required]],
    tags: [[] as string[]],
  });

  constructor() {
    effect(() => {
      const entry = this.store.selectedEntry();
      if (entry && this.isEditMode()) {
        this.form.patchValue({
          title: entry.title,
          content: entry.content,
          mood: entry.mood,
          tags: [...entry.tags],
        });
      }
    });
  }

  ngOnInit(): void {
    const id = this.entryId();
    if (id) {
      this.store.loadEntry(id);
    }
  }

  protected onSubmit(): void {
    if (this.form.valid) {
      const { title, content, mood, tags } = this.form.value;
      const request = { title, content, mood, tags };

      const id = this.entryId();
      if (id) {
        this.store.updateEntry({ id, request });
      } else {
        this.store.createEntry(request);
      }
    } else {
      this.form.markAllAsTouched();
    }
  }

  protected onCancel(): void {
    this.router.navigate(['/journal']);
  }

  protected addTag(): void {
    const tag = this.tagInput().trim().toLowerCase();
    if (tag && !this.currentTags.includes(tag)) {
      this.form.patchValue({
        tags: [...this.currentTags, tag],
      });
    }
    this.tagInput.set('');
  }

  protected removeTag(tagToRemove: string): void {
    this.form.patchValue({
      tags: this.currentTags.filter((t) => t !== tagToRemove),
    });
  }

  protected onTagKeydown(event: KeyboardEvent): void {
    if (event.key === 'Enter' || event.key === ',') {
      event.preventDefault();
      this.addTag();
    }
  }

  protected get currentTags(): string[] {
    return this.form.get('tags')?.value || [];
  }

  protected get titleError(): string {
    const control = this.form.get('title');
    if (control?.hasError('required')) return 'Title is required';
    if (control?.hasError('minlength')) return 'Title must be at least 3 characters';
    if (control?.hasError('maxlength')) return 'Title cannot exceed 255 characters';
    return '';
  }

  protected get contentError(): string {
    const control = this.form.get('content');
    if (control?.hasError('required')) return 'Content is required';
    if (control?.hasError('minlength')) return 'Content must be at least 10 characters';
    return '';
  }
}
