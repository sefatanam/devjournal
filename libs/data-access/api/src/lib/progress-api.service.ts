import { HttpClient } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { LearningProgress, ProgressSummary } from '@devjournal/shared-models';
import { API_CONFIG, defaultApiConfig } from './api.config';

export interface WeeklyProgressResponse {
  progress: LearningProgress[];
  period: 'weekly';
}

export interface MonthlyProgressResponse {
  progress: LearningProgress[];
  period: 'monthly';
}

export interface StreakResponse {
  currentStreak: number;
}

// @REVIEW - Phase 7: Progress API Service
@Injectable({ providedIn: 'root' })
export class ProgressApiService {
  private readonly http = inject(HttpClient);
  private readonly config = inject(API_CONFIG, { optional: true }) ?? defaultApiConfig;

  private readonly baseUrl = `${this.config.baseUrl}/api/progress`;

  getSummary(): Observable<ProgressSummary> {
    return this.http.get<ProgressSummary>(`${this.baseUrl}/summary`);
  }

  getToday(): Observable<LearningProgress> {
    return this.http.get<LearningProgress>(`${this.baseUrl}/today`);
  }

  getWeekly(): Observable<WeeklyProgressResponse> {
    return this.http.get<WeeklyProgressResponse>(`${this.baseUrl}/weekly`);
  }

  getMonthly(): Observable<MonthlyProgressResponse> {
    return this.http.get<MonthlyProgressResponse>(`${this.baseUrl}/monthly`);
  }

  getStreak(): Observable<StreakResponse> {
    return this.http.get<StreakResponse>(`${this.baseUrl}/streak`);
  }
}
