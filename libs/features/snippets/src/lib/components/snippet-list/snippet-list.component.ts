import {
  ChangeDetectionStrategy,
  Component,
  effect,
  inject,
  OnInit,
  signal,
} from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { SnippetStore, SUPPORTED_LANGUAGES } from '../../store/snippet.store';

@Component({
  selector: 'lib-snippet-list',
  standalone: true,
  imports: [CommonModule, RouterLink, FormsModule],
  templateUrl: './snippet-list.component.html',
  styleUrl: './snippet-list.component.scss',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class SnippetListComponent implements OnInit {
  protected readonly store = inject(SnippetStore);

  protected readonly searchQuery = signal('');
  protected readonly selectedLanguage = signal<string>('');

  protected readonly languages = SUPPORTED_LANGUAGES;

  constructor() {
    effect(() => {
      const search = this.searchQuery();
      const language = this.selectedLanguage();

      this.store.setFilter({
        search: search || undefined,
        language: language || undefined,
      });
    });
  }

  ngOnInit(): void {
    this.store.loadSnippets();
  }

  protected onSearch(query: string): void {
    this.searchQuery.set(query);
  }

  protected onLanguageFilter(language: string): void {
    this.selectedLanguage.set(language);
  }

  protected onPageChange(page: number): void {
    this.store.setPage(page);
    this.store.loadSnippets();
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
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  }

  protected truncateCode(code: string, maxLines = 5): string {
    const lines = code.split('\n');
    if (lines.length <= maxLines) return code;
    return lines.slice(0, maxLines).join('\n') + '\n...';
  }
}
