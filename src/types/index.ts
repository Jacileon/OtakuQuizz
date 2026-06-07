// ============================================================
// TYPES OTAKU QUIZ AFRICA - TypeScript Strict
// ============================================================

// === RANGS ===
export type Rank = 'F' | 'E' | 'D' | 'C' | 'B' | 'A' | 'S' | 'S+' | 'SS' | 'SSS' | 'Légende';

export type RankConfig = {
  rank: Rank;
  minXP: number;
  maxXP: number | null;
  color: string;
  bgColor: string;
  label: string;
};

// === UTILISATEUR ===
export type UserProfile = {
  id: string;
  email: string;
  username: string;
  avatar_url: string | null;
  bio: string | null;
  country: string | null;
  phone: string | null;
  favorite_anime: string | null;
  xp: number;
  level: number;
  rank: Rank;
  is_premium: boolean;
  created_at: string;
  updated_at: string;
};

export type UserProfileInsert = Omit<UserProfile, 'created_at' | 'updated_at'>;
export type UserProfileUpdate = Partial<Omit<UserProfile, 'id' | 'created_at'>>;

export type UserStats = {
  user_id: string;
  quizzes_played: number;
  quizzes_created: number;
  total_correct_answers: number;
  total_answers: number;
  accuracy_rate: number;
  best_score_ever: number;
  monthly_rank: number | null;
  global_rank: number | null;
  updated_at: string;
};

// === QUIZ ===
export type QuizType = 'community' | 'official' | 'private';
export type QuizStatus = 'draft' | 'published' | 'hidden' | 'deleted';
export type QuestionType = 'text' | 'true_false' | 'image' | 'gif' | 'audio' | 'character_guess' | 'impostor';

// Type pour les questions character_guess et impostor
export type CharacterGuessItem = {
  image_url?: string;  // URL de l'image (optionnel)
  answer: string;      // Le nom du personnage/anime
  clue?: string;       // Indice textuel pour le mode "text"
};

export type CharacterGuessData = {
  characters: CharacterGuessItem[]; // 4 personnages/anime
  mode?: 'image' | 'text'; // Mode d'affichage
};

// Type pour "Devine" - 4 indices (texte ou image) + 1 réponse
export type GuessClue = {
  type: 'text' | 'image';
  content: string; // Texte OU URL de l'image
};

export type GuessData = {
  clues: GuessClue[]; // 4 indices mixtes
  answer: string;      // La réponse à deviner
};

// Type pour "Trouve l'intrus" - 4 éléments + 1 intrus
export type FindOddItem = {
  type: 'text' | 'image';
  content: string; // Texte OU URL de l'image
};

export type FindOddData = {
  items: FindOddItem[]; // 4 éléments (3 similaires + 1 intrus)
  odd_index: number;     // Index de l'intrus (0-3)
};

export type Quiz = {
  id: string;
  creator_id: string;
  title: string;
  description: string | null;
  thumbnail_url: string | null;
  thumbnail_public_id: string | null;
  category: string;
  subcategory: string;
  series: string;
  quiz_type: QuizType;
  status: QuizStatus;
  question_count: number;
  play_count: number;
  total_reports: number;
  is_visible: boolean;
  event_start_at: string | null;
  event_end_at: string | null;
  created_at: string;
  updated_at: string;
  creator?: UserProfile;
};

export type QuizInsert = Omit<Quiz, 'id' | 'created_at' | 'updated_at' | 'play_count' | 'total_reports'>;
export type QuizUpdate = Partial<Omit<Quiz, 'id' | 'creator_id' | 'created_at'>>;

export type QuizWithQuestions = Quiz & { questions: Question[] };

// === QUESTIONS ===
export type Question = {
  id: string;
  quiz_id: string;
  question_text: string;
  question_type: QuestionType;
  media_url: string | null;
  media_public_id: string | null;
  time_limit_seconds: number;
  order_index: number;
  created_at: string;
  answers: Answer[];
  character_guess_data?: CharacterGuessData;
  character_guess_mode?: 'image' | 'text';
  find_odd_data?: FindOddData;
};

export type QuestionInsert = Omit<Question, 'id' | 'created_at'>;

export type Answer = {
  id: string;
  question_id: string;
  answer_text: string;
  is_correct: boolean;
  order_index: number;
  created_at: string;
};

export type AnswerInsert = Omit<Answer, 'id' | 'created_at'>;

// Answer without is_correct (for client-side during quiz)
export type AnswerClient = Omit<Answer, 'is_correct'>;

export type QuestionClient = Omit<Question, 'answers'> & {
  answers: AnswerClient[];
};

// === SESSIONS DE JEU ===
export type GameSession = {
  id: string;
  user_id: string;
  quiz_id: string;
  started_at: string;
  completed_at: string | null;
  score: number;
  correct_count: number;
  total_questions: number;
  accuracy_rate: number;
  is_perfect: boolean;
  time_taken_ms: number | null;
  created_at: string;
};

export type PlayerAnswer = {
  id: string;
  session_id: string;
  question_id: string;
  answer_id: string | null;
  is_correct: boolean;
  time_taken_ms: number;
  points_earned: number;
  created_at: string;
};

export type PlayerAnswerDraft = {
  questionId: string;
  answerId: string | null;
  timeMs: number;
};

// === CLASSEMENTS ===
export type LeaderboardType = 'global' | 'monthly' | 'weekly' | 'quiz' | 'series';

export type LeaderboardEntry = {
  rank: number;
  user_id: string;
  username: string;
  avatar_url: string | null;
  user_rank: Rank;
  score: number;
  xp?: number;
  quiz_count?: number;
  accuracy_rate?: number;
  time_taken_ms?: number;
};

// === SOCIAL ===
export type FriendshipStatus = 'pending' | 'accepted' | 'rejected';

export type Friendship = {
  id: string;
  requester_id: string;
  addressee_id: string;
  status: FriendshipStatus;
  created_at: string;
  updated_at: string;
  requester?: UserProfile;
  addressee?: UserProfile;
};

// === BADGES ===
export type Badge = {
  id: string;
  slug: string;
  name: string;
  description: string;
  icon_url: string | null;
  condition_type: string;
  condition_value: number;
  is_rare: boolean;
  created_at: string;
};

export type UserBadge = {
  id: string;
  user_id: string;
  badge_id: string;
  earned_at: string;
  badge?: Badge;
};

// === SIGNALEMENTS ===
export type ReportReason = 'wrong_answer' | 'incorrect_content' | 'spam' | 'plagiarism' | 'inappropriate';
export type ReportStatus = 'pending' | 'reviewed' | 'resolved' | 'dismissed';

export type Report = {
  id: string;
  reporter_id: string;
  quiz_id: string;
  reason: ReportReason;
  description: string | null;
  status: ReportStatus;
  created_at: string;
};

// === COLLECTIONS ===
export type SeriesCollection = {
  series: string;
  total_quizzes: number;
  completed_quizzes: number;
  progress_percent: number;
  best_score: number | null;
};

// === API RESPONSES ===
export type ApiResponse<T> = {
  data: T | null;
  error: string | null;
  success: boolean;
};

export type PaginatedResponse<T> = {
  data: T[];
  count: number;
  page: number;
  per_page: number;
  total_pages: number;
};

// === SEARCH & FILTERS ===
export type SearchParams = {
  query?: string;
  category?: string;
  subcategory?: string;
  series?: string;
  sortBy?: 'popular' | 'recent' | 'rated' | 'series';
  page?: number;
  perPage?: number;
};

// === QUIZ CREATION ===
export type QuizCreateInput = {
  title: string;
  description?: string;
  category: string;
  subcategory: string;
  series: string;
  thumbnail_url?: string;
  thumbnail_public_id?: string;
  questions: QuestionCreateInput[];
};

export type QuestionCreateInput = {
  question_text: string;
  question_type: QuestionType;
  media_url?: string;
  media_public_id?: string;
  time_limit_seconds: number;
  answers: AnswerCreateInput[];
  character_guess_data?: CharacterGuessData; // Pour le type character_guess
  character_guess_mode?: 'image' | 'text'; // Mode pour character_guess et impostor
  guess_data?: GuessData; // Pour le type devine: 4 indices + 1 réponse
  find_odd_data?: FindOddData; // Pour le type impostor: 4 éléments + 1 intrus
};

export type AnswerCreateInput = {
  answer_text: string;
  is_correct: boolean;
};

export type QuizUpdateInput = Partial<QuizCreateInput>;

// === SUBMIT ===
export type QuizSubmitInput = {
  sessionId: string;
  answers: PlayerAnswerDraft[];
};

export type QuizSubmitResult = {
  score: number;
  correctCount: number;
  totalQuestions: number;
  accuracyRate: number;
  isPerfect: boolean;
  xpEarned: number;
  newBadges?: Badge[];
};

// === NOTIFICATIONS ===
export type NotificationType = 'friend_request' | 'badge_unlocked' | 'quiz_completed' | 'event_starting' | 'rank_up';

export type Notification = {
  id: string;
  user_id: string;
  type: NotificationType;
  title: string;
  message: string;
  data: Record<string, unknown> | null;
  is_read: boolean;
  created_at: string;
};

// === ADMIN ===
export type AdminAction = 'restore' | 'delete' | 'dismiss';

export type ModerationStats = {
  total_reports: number;
  pending_reports: number;
  resolved_reports: number;
  hidden_quizzes: number;
  banned_users: number;
};

// === CHAT ===
export type Conversation = {
  id: string;
  user1_id: string;
  user2_id: string;
  last_message_at: string | null;
  created_at: string;
  other_user?: UserProfile;
  last_message?: Message;
  unread_count?: number;
};

export type Message = {
  id: string;
  conversation_id: string;
  sender_id: string;
  content: string;
  is_read: boolean;
  created_at: string;
  sender?: UserProfile;
};

// === ADMIN CHAT ===
export type AdminConversation = {
  id: string;
  user_id: string;
  admin_id: string | null;
  subject: string;
  status: 'open' | 'assigned' | 'closed';
  last_message_at: string | null;
  created_at: string;
  user?: UserProfile;
  admin?: UserProfile;
  last_message?: AdminMessage;
  unread_count?: number;
};

export type AdminMessage = {
  id: string;
  conversation_id: string;
  sender_id: string;
  content: string;
  is_read: boolean;
  created_at: string;
  sender?: UserProfile;
};

