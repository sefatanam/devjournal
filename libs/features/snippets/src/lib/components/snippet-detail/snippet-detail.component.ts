import {
  ChangeDetectionStrategy,
  Component,
  inject,
  input,
  OnInit,
  signal,
} from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { SnippetStore, SUPPORTED_LANGUAGES } from '../../store/snippet.store';
import { CodeEditorComponent } from '../code-editor/code-editor.component';

@Component({
  selector: 'lib-snippet-detail',
  standalone: true,
  imports: [CommonModule, RouterLink, CodeEditorComponent],
  templateUrl: './snippet-detail.component.html',
  styleUrl: './snippet-detail.component.scss',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class SnippetDetailComponent implements OnInit {
  protected readonly store = inject(SnippetStore);

  readonly id = input.required<string>();

  protected readonly showDeleteConfirm = signal(false);
  protected readonly copied = signal(false);

  protected readonly languages = SUPPORTED_LANGUAGES;

  ngOnInit(): void {
    this.store.loadSnippet(this.id());
  }

  protected getLanguageIcon(language: string): string {
    const lang = this.languages.find((l) => l.value === language);
    return lang?.icon || '📦';
  }

  protected getLanguageLabel(language: string): string {
    const lang = this.languages.find((l) => l.value === language);
    return lang?.label || language;
  }

  protected formatDate(date: Date): string {
    return new Date(date).toLocaleDateString('en-US', {
      weekday: 'long',
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  }

  protected formatTime(date: Date): string {
    return new Date(date).toLocaleTimeString('en-US', {
      hour: '2-digit',
      minute: '2-digit',
    });
  }

  protected async copyCode(): Promise<void> {
    const snippet = this.store.selectedSnippet();
    if (snippet?.code) {
      try {
        await navigator.clipboard.writeText(snippet.code);
        this.copied.set(true);
        setTimeout(() => this.copied.set(false), 2000);
      } catch (err) {
        console.error('Failed to copy code:', err);
      }
    }
  }

  protected confirmDelete(): void {
    this.showDeleteConfirm.set(true);
  }

  protected cancelDelete(): void {
    this.showDeleteConfirm.set(false);
  }

  protected deleteSnippet(): void {
    this.store.deleteSnippet(this.id());
  }
}
