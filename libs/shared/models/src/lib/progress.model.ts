export interface LearningProgress {
  id: string;
  userId: string;
  date: Date;
  entriesCount: number;
  snippetsCount: number;
  streakDays: number;
}

export interface ProgressSummary {
  totalEntries: number;
  totalSnippets: number;
  currentStreak: number;
  longestStreak: number;
  thisWeekEntries: number;
  thisMonthEntries: number;
}

export interface DailyProgress {
  date: Date;
  entriesCount: number;
  snippetsCount: number;
}
