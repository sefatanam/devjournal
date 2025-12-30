import { ChangeDetectionStrategy, Component, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { ProgressStore } from '../../store/progress.store';

// @REVIEW - Phase 7: Progress Dashboard Component
@Component({
  selector: 'devjournal-progress-dashboard',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [CommonModule, RouterLink],
  templateUrl: './progress-dashboard.component.html',
  styleUrl: './progress-dashboard.component.scss',
})
export class ProgressDashboardComponent implements OnInit {
  readonly store = inject(ProgressStore);
  readonly Math = Math;

  ngOnInit(): void {
    this.store.loadAll();
  }
}
