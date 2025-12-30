// Proto messages (types and schemas)
export {
  JournalEntry,
  JournalEntrySchema,
  CreateEntryRequest,
  CreateEntryRequestSchema,
  GetEntryRequest,
  GetEntryRequestSchema,
  ListEntriesRequest,
  ListEntriesRequestSchema,
  ListEntriesResponse,
  ListEntriesResponseSchema,
  UpdateEntryRequest,
  UpdateEntryRequestSchema,
  DeleteEntryRequest,
  DeleteEntryRequestSchema,
  DeleteEntryResponse,
  DeleteEntryResponseSchema,
  SearchEntriesRequest,
  SearchEntriesRequestSchema,
} from './lib/devjournal/v1/journal_pb';

export {
  Snippet,
  SnippetSchema,
  CreateSnippetRequest,
  CreateSnippetRequestSchema,
  GetSnippetRequest,
  GetSnippetRequestSchema,
  ListSnippetsRequest,
  ListSnippetsRequestSchema,
  ListSnippetsResponse,
  ListSnippetsResponseSchema,
  UpdateSnippetRequest,
  UpdateSnippetRequestSchema,
  DeleteSnippetRequest,
  DeleteSnippetRequestSchema,
  DeleteSnippetResponse,
  DeleteSnippetResponseSchema,
  SearchSnippetsRequest,
  SearchSnippetsRequestSchema,
} from './lib/devjournal/v1/snippet_pb';

export {
  User,
  UserSchema,
  RegisterRequest,
  RegisterRequestSchema,
  LoginRequest,
  LoginRequestSchema,
  AuthResponse,
  AuthResponseSchema,
  GetProfileRequest,
  GetProfileRequestSchema,
  UpdateProfileRequest,
  UpdateProfileRequestSchema,
  ValidateTokenRequest,
  ValidateTokenRequestSchema,
} from './lib/devjournal/v1/user_pb';

export {
  DailyProgress,
  DailyProgressSchema,
  ProgressSummary,
  ProgressSummarySchema,
  ProgressList,
  ProgressListSchema,
  StreakResponse,
  StreakResponseSchema,
  GetSummaryRequest,
  GetSummaryRequestSchema,
  GetTodayProgressRequest,
  GetTodayProgressRequestSchema,
  GetWeeklyProgressRequest,
  GetWeeklyProgressRequestSchema,
  GetMonthlyProgressRequest,
  GetMonthlyProgressRequestSchema,
  GetStreakRequest,
  GetStreakRequestSchema,
} from './lib/devjournal/v1/progress_pb';

// Connect RPC clients (services)
export { JournalService } from './lib/devjournal/v1/journal_connect';
export { SnippetService } from './lib/devjournal/v1/snippet_connect';
export { AuthService } from './lib/devjournal/v1/user_connect';
export { ProgressService } from './lib/devjournal/v1/progress_connect';
