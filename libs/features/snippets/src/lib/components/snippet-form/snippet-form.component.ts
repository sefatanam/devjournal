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
import { SnippetStore, SUPPORTED_LANGUAGES } from '../../store/snippet.store';
import { CodeEditorComponent } from '../code-editor/code-editor.component';

@Component({
  selector: 'lib-snippet-form',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, RouterLink, CodeEditorComponent],
  templateUrl: './snippet-form.component.html',
  styleUrl: './snippet-form.component.scss',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class SnippetFormComponent implements OnInit {
  private readonly fb = inject(FormBuilder);
  private readonly router = inject(Router);
  protected readonly store = inject(SnippetStore);

  readonly snippetId = input<string | undefined>(undefined, { alias: 'id' });

  protected readonly isEditMode = computed(() => !!this.snippetId());
  protected readonly pageTitle = computed(() =>
    this.isEditMode() ? 'Edit Snippet' : 'New Snippet'
  );

  protected readonly languages = SUPPORTED_LANGUAGES;
  protected readonly tagInput = signal('');
  protected readonly codeContent = signal('');

  protected readonly form: FormGroup = this.fb.group({
    title: ['', [Validators.required, Validators.minLength(3), Validators.maxLength(255)]],
    description: [''],
    language: ['typescript', [Validators.required]],
    code: ['', [Validators.required, Validators.minLength(1)]],
    tags: [[] as string[]],
  });

  constructor() {
    effect(() => {
      const snippet = this.store.selectedSnippet();
      if (snippet && this.isEditMode()) {
        this.form.patchValue({
          title: snippet.title,
          description: snippet.description,
          language: snippet.language,
          code: snippet.code,
          tags: [...snippet.tags],
        });
        this.codeContent.set(snippet.code);
      }
    });
  }

  ngOnInit(): void {
    const id = this.snippetId();
    if (id) {
      this.store.loadSnippet(id);
    }
  }

  protected onSubmit(): void {
    if (this.form.valid) {
      const { title, description, language, code, tags } = this.form.value;
      const request = { title, description, language, code, tags };

      const id = this.snippetId();
      if (id) {
        this.store.updateSnippet({ id, request });
      } else {
        this.store.createSnippet(request);
      }
    } else {
      this.form.markAllAsTouched();
    }
  }

  protected onCancel(): void {
    this.router.navigate(['/snippets']);
  }

  protected onCodeChange(code: string): void {
    this.form.patchValue({ code });
    this.codeContent.set(code);
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

  protected get selectedLanguage(): string {
    return this.form.get('language')?.value || 'typescript';
  }

  protected get titleError(): string {
    const control = this.form.get('title');
    if (control?.hasError('required')) return 'Title is required';
    if (control?.hasError('minlength')) return 'Title must be at least 3 characters';
    if (control?.hasError('maxlength')) return 'Title cannot exceed 255 characters';
    return '';
  }

  protected get codeError(): string {
    const control = this.form.get('code');
    if (control?.hasError('required')) return 'Code is required';
    return '';
  }
}
