// ============================================================
// CONSTANTS OTAKU QUIZ AFRICA
// ============================================================

import { RankConfig } from '@/types';

export const RANKS: RankConfig[] = [
  { rank: 'F', minXP: 0, maxXP: 99, color: '#888888', bgColor: 'bg-rank-f', label: 'Novice' },
  { rank: 'E', minXP: 100, maxXP: 499, color: '#4CAF50', bgColor: 'bg-rank-e', label: 'Apprenti' },
  { rank: 'D', minXP: 500, maxXP: 1499, color: '#2196F3', bgColor: 'bg-rank-d', label: 'Initié' },
  { rank: 'C', minXP: 1500, maxXP: 2999, color: '#9C27B0', bgColor: 'bg-rank-c', label: 'Connaisseur' },
  { rank: 'B', minXP: 3000, maxXP: 5999, color: '#FF9800', bgColor: 'bg-rank-b', label: 'Expert' },
  { rank: 'A', minXP: 6000, maxXP: 9999, color: '#F44336', bgColor: 'bg-rank-a', label: 'Maître' },
  { rank: 'S', minXP: 10000, maxXP: 14999, color: '#FFD700', bgColor: 'bg-rank-s', label: 'Elite' },
  { rank: 'S+', minXP: 15000, maxXP: 24999, color: '#FFA500', bgColor: 'bg-rank-s-plus', label: 'Héros' },
  { rank: 'SS', minXP: 25000, maxXP: 39999, color: '#FF69B4', bgColor: 'bg-rank-ss', label: 'Légende' },
  { rank: 'SSS', minXP: 40000, maxXP: 59999, color: '#00FFFF', bgColor: 'bg-rank-sss', label: 'Mythique' },
  { rank: 'Légende', minXP: 60000, maxXP: null, color: '#FF0080', bgColor: 'bg-rank-legend', label: 'Immortel' },
];

export const BADGE_TYPES = {
  FIRST_QUIZ: 'first_quiz',
  QUIZ_10: 'quiz_10',
  QUIZ_100: 'quiz_100',
  QUIZ_500: 'quiz_500',
  QUIZ_1000: 'quiz_1000',
  PERFECT_QUIZ: 'perfect_quiz',
  PERFECT_5: 'perfect_5',
  ACCURACY_90: 'accuracy_90',
  FIRST_CREATION: 'first_creation',
  CREATOR_10: 'creator_10',
  POPULAR_CREATOR: 'popular_creator',
  ELITE_CREATOR: 'elite_creator',
  TOP10_MONTHLY: 'top10_monthly',
  MONTHLY_CHAMPION: 'monthly_champion',
} as const;

export const QUESTION_TYPES = {
  TEXT: 'text',
  TRUE_FALSE: 'true_false',
  IMAGE: 'image',
  IMAGE_SHADOW: 'image_shadow',
  GIF: 'gif',
  AUDIO: 'audio',
} as const;

export const REPORT_REASONS = {
  WRONG_ANSWER: 'wrong_answer',
  INCORRECT_CONTENT: 'incorrect_content',
  SPAM: 'spam',
  PLAGIARISM: 'plagiarism',
  INAPPROPRIATE: 'inappropriate',
} as const;

export const CATEGORY_LIST = [
  'Anime',
  'Manga',
  'Openings',
  'Films',
  'Personnages',
  'OST',
  'Studio',
  'Seiyuu',
];

export const SUBCATEGORY_LIST: Record<string, string[]> = {
  'Anime': ['Shonen', 'Shojo', 'Seinen', 'Isekai', 'Mecha', 'Slice of Life', 'Thriller', 'Fantasy'],
  'Manga': ['Shonen', 'Shojo', 'Seinen', 'Josei', 'Webtoon', 'One-shot'],
  'Openings': ['Rock', 'Pop', 'Jazz', 'Classique', 'Rap', 'Indie'],
  'Films': ['Ghibli', 'Shinkai', 'Hosoda', 'Autres'],
  'Personnages': ['Héros', 'Vilains', 'Side Characters', 'Animaux'],
  'OST': ['Battle', 'Emotion', 'Comédie', 'Suspense'],
  'Studio': ['MAPPA', 'Ufotable', 'Wit Studio', 'A-1 Pictures', 'Bones', 'Madhouse'],
  'Seiyuu': ['Masculin', 'Féminin', 'Légendaire'],
};

export const SCORE_BASE_POINTS = 100;
export const SCORE_MAX_SPEED_BONUS = 50;
export const SCORE_STREAK_BONUS = 10;
export const QUIZ_MIN_QUESTIONS = 5;
export const REPORT_THRESHOLD = 5;
export const MAX_QUIZ_TIME_SECONDS = 60;
export const MIN_QUIZ_TIME_SECONDS = 10;
export const DEFAULT_QUIZ_TIME_SECONDS = 30;
export const MAX_ANSWERS_PER_QUESTION = 6;
export const MIN_ANSWERS_PER_QUESTION = 2;
export const LEADERBOARD_PAGE_SIZE = 50;
export const EXPLORE_PAGE_SIZE = 12;

export const RANK_COLORS: Record<string, string> = {
  'F': '#888888',
  'E': '#4CAF50',
  'D': '#2196F3',
  'C': '#9C27B0',
  'B': '#FF9800',
  'A': '#F44336',
  'S': '#FFD700',
  'S+': '#FFA500',
  'SS': '#FF69B4',
  'SSS': '#00FFFF',
  'Légende': '#FF0080',
};

export const RANK_LABELS: Record<string, string> = {
  'F': 'Novice',
  'E': 'Apprenti',
  'D': 'Initié',
  'C': 'Connaisseur',
  'B': 'Expert',
  'A': 'Maître',
  'S': 'Elite',
  'S+': 'Héros',
  'SS': 'Légende',
  'SSS': 'Mythique',
  'Légende': 'Immortel',
};

