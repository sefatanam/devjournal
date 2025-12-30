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
import { JournalStore } from '../../store/journal.store';
import { JournalMood } from '@devjournal/shared-models';

@Component({
  selector: 'lib-journal-list',
  standalone: true,
  imports: [CommonModule, RouterLink, FormsModule],
  templateUrl: './journal-list.component.html',
  styleUrl: './journal-list.component.scss',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class JournalListComponent implements OnInit {
  protected readonly store = inject(JournalStore);

  protected readonly searchQuery = signal('');
  protected readonly selectedMood = signal<JournalMood | ''>('');

  protected readonly moods: { value: JournalMood | ''; label: string; emoji: string }[] = [
    { value: '', label: 'All Moods', emoji: '🎭' },
    { value: 'excited', label: 'Excited', emoji: '🤩' },
    { value: 'happy', label: 'Happy', emoji: '😊' },
    { value: 'neutral', label: 'Neutral', emoji: '😐' },
    { value: 'frustrated', label: 'Frustrated', emoji: '😤' },
    { value: 'tired', label: 'Tired', emoji: '😴' },
  ];

  constructor() {
    effect(() => {
      const search = this.searchQuery();
      const mood = this.selectedMood();

      this.store.setFilter({
        search: search || undefined,
        mood: mood || undefined,
      });
    });
  }

  ngOnInit(): void {
    this.store.loadEntries();
  }

  protected onSearch(query: string): void {
    this.searchQuery.set(query);
  }

  protected onMoodFilter(mood: JournalMood | ''): void {
    this.selectedMood.set(mood);
  }

  protected onPageChange(page: number): void {
    this.store.setPage(page);
    this.store.loadEntries();
  }

  protected getMoodEmoji(mood: JournalMood): string {
    const moodMap: Record<JournalMood, string> = {
      excited: '🤩',
      happy: '😊',
      neutral: '😐',
      frustrated: '😤',
      tired: '😴',
    };
    return moodMap[mood] || '😐';
  }

  protected formatDate(date: Date): string {
    return new Date(date).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  }

  protected truncateContent(content: string, maxLength = 150): string {
    if (content.length <= maxLength) return content;
    return content.substring(0, maxLength).trim() + '...';
  }
}
