// ============================================================
// UTILITIES OTAKU QUIZ AFRICA
// ============================================================

import { type ClassValue, clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';
import { Rank, RankConfig } from '@/types';
import { RANKS, RANK_COLORS } from './constants';

export function cn(...inputs: ClassValue[]): string {
  return twMerge(clsx(inputs));
}

export function getRankFromXP(xp: number): Rank {
  for (let i = RANKS.length - 1; i >= 0; i--) {
    if (xp >= RANKS[i].minXP) {
      return RANKS[i].rank;
    }
  }
  return 'F';
}

export function getRankConfig(rank: Rank): RankConfig | undefined {
  return RANKS.find((r) => r.rank === rank);
}

export function getRankColor(rank: string): string {
  return RANK_COLORS[rank] || '#888888';
}

export function getNextRankXP(currentRank: Rank): number {
  const currentIndex = RANKS.findIndex((r) => r.rank === currentRank);
  if (currentIndex === -1 || currentIndex >= RANKS.length - 1) return 0;
  return RANKS[currentIndex + 1].minXP;
}

export function getCurrentRankProgress(xp: number): { current: number; next: number; percent: number } {
  const rank = getRankFromXP(xp);
  const rankConfig = getRankConfig(rank);
  if (!rankConfig) return { current: 0, next: 100, percent: 0 };

  const nextRankXP = getNextRankXP(rank);
  if (nextRankXP === 0) return { current: xp, next: xp, percent: 100 };

  const progress = xp - rankConfig.minXP;
  const total = nextRankXP - rankConfig.minXP;
  const percent = Math.min(100, Math.max(0, (progress / total) * 100));

  return {
    current: progress,
    next: total,
    percent,
  };
}

export function calculateScore(
  correct: boolean,
  timeMs: number,
  maxTimeMs: number,
  streak: number
): number {
  if (!correct) return 0;

  const basePoints = 100;
  const timeRatio = Math.max(0, 1 - timeMs / maxTimeMs);
  const speedBonus = Math.round(timeRatio * 50);
  const streakBonus = streak * 10;

  return basePoints + speedBonus + streakBonus;
}

export function formatXP(xp: number): string {
  if (xp >= 1000000) return `${(xp / 1000000).toFixed(1)}M`;
  if (xp >= 1000) return `${(xp / 1000).toFixed(1)}K`;
  return xp.toString();
}

export function formatTimeAgo(date: Date | string): string {
  const now = new Date();
  const target = typeof date === 'string' ? new Date(date) : date;
  const diffMs = now.getTime() - target.getTime();
  const diffSec = Math.floor(diffMs / 1000);
  const diffMin = Math.floor(diffSec / 60);
  const diffHour = Math.floor(diffMin / 60);
  const diffDay = Math.floor(diffHour / 24);
  const diffWeek = Math.floor(diffDay / 7);
  const diffMonth = Math.floor(diffDay / 30);
  const diffYear = Math.floor(diffDay / 365);

  if (diffSec < 60) return "À l'instant";
  if (diffMin < 60) return `Il y a ${diffMin} min`;
  if (diffHour < 24) return `Il y a ${diffHour}h`;
  if (diffDay < 7) return `Il y a ${diffDay} j`;
  if (diffWeek < 4) return `Il y a ${diffWeek} sem`;
  if (diffMonth < 12) return `Il y a ${diffMonth} mois`;
  return `Il y a ${diffYear} an${diffYear > 1 ? 's' : ''}`;
}

export function formatDuration(ms: number): string {
  const seconds = Math.floor(ms / 1000);
  const minutes = Math.floor(seconds / 60);
  const remainingSeconds = seconds % 60;

  if (minutes > 0) {
    return `${minutes}m ${remainingSeconds.toString().padStart(2, '0')}s`;
  }
  return `${remainingSeconds}s`;
}

export function formatNumber(num: number): string {
  return new Intl.NumberFormat('fr-FR').format(num);
}

export function shuffleArray<T>(array: T[]): T[] {
  const shuffled = [...array];
  for (let i = shuffled.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [shuffled[i], shuffled[j]] = [shuffled[j], shuffled[i]];
  }
  return shuffled;
}

export function slugify(text: string): string {
  return text
    .toLowerCase()
    .normalize('NFD')
    .replace(/[̀-ͯ]/g, '')
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '');
}

export function validateUsername(username: string): { valid: boolean; error?: string } {
  if (username.length < 3) return { valid: false, error: 'Minimum 3 caractères' };
  if (username.length > 30) return { valid: false, error: 'Maximum 30 caractères' };
  if (!/^[a-zA-Z0-9_]+$/.test(username)) return { valid: false, error: 'Lettres, chiffres et underscore uniquement' };
  return { valid: true };
}

export function isQuizOfficialActive(quiz: { event_start_at: string | null; event_end_at: string | null }): boolean {
  if (!quiz.event_start_at || !quiz.event_end_at) return false;
  const now = new Date();
  const start = new Date(quiz.event_start_at);
  const end = new Date(quiz.event_end_at);
  return now >= start && now <= end;
}

export function isQuizOfficialUpcoming(quiz: { event_start_at: string | null }): boolean {
  if (!quiz.event_start_at) return false;
  return new Date() < new Date(quiz.event_start_at);
}

export function getCountdown(targetDate: string | Date): { days: number; hours: number; minutes: number; seconds: number; total: number } {
  const now = new Date().getTime();
  const target = new Date(targetDate).getTime();
  const diff = Math.max(0, target - now);

  return {
    days: Math.floor(diff / (1000 * 60 * 60 * 24)),
    hours: Math.floor((diff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60)),
    minutes: Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60)),
    seconds: Math.floor((diff % (1000 * 60)) / 1000),
    total: diff,
  };
}

export function getInitials(name: string): string {
  return name
    .split(' ')
    .map((n) => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2);
}

export function getDisplayName(profile: { nickname?: string | null; username: string }): string {
  return profile.nickname || profile.username;
}

