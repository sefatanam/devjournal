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
import { JournalStore } from '../../store/journal.store';
import { JournalMood } from '@devjournal/shared-models';

@Component({
  selector: 'lib-journal-detail',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './journal-detail.component.html',
  styleUrl: './journal-detail.component.scss',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class JournalDetailComponent implements OnInit {
  protected readonly store = inject(JournalStore);

  readonly id = input.required<string>();

  protected readonly showDeleteConfirm = signal(false);

  ngOnInit(): void {
    this.store.loadEntry(this.id());
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

  protected getMoodLabel(mood: JournalMood): string {
    const labelMap: Record<JournalMood, string> = {
      excited: 'Excited',
      happy: 'Happy',
      neutral: 'Neutral',
      frustrated: 'Frustrated',
      tired: 'Tired',
    };
    return labelMap[mood] || 'Neutral';
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

  protected confirmDelete(): void {
    this.showDeleteConfirm.set(true);
  }

  protected cancelDelete(): void {
    this.showDeleteConfirm.set(false);
  }

  protected deleteEntry(): void {
    this.store.deleteEntry(this.id());
  }
}
