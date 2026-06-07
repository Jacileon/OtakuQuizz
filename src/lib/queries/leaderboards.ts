'use server';

// ============================================================
// REQUÊTES CLASSEMENTS - Supabase Server
// ============================================================

import { createClient } from '@/lib/supabase/server';
import { LeaderboardEntry, PaginatedResponse } from '@/types';
import { LEADERBOARD_PAGE_SIZE } from '@/lib/constants';

export async function getGlobalLeaderboard(page: number = 1, perPage: number = LEADERBOARD_PAGE_SIZE): Promise<LeaderboardEntry[]> {
  const supabase = createClient();
  const { data } = await supabase
    .rpc('get_global_leaderboard', { limit_count: perPage, offset_count: (page - 1) * perPage });
  return (data || []) as LeaderboardEntry[];
}

export async function getMonthlyLeaderboard(yearMonth: string, page: number = 1): Promise<LeaderboardEntry[]> {
  const supabase = createClient();
  const { data } = await supabase
    .rpc('get_monthly_leaderboard', { year_month: yearMonth, limit_count: LEADERBOARD_PAGE_SIZE });
  return (data || []) as LeaderboardEntry[];
}

export async function getWeeklyLeaderboard(page: number = 1): Promise<LeaderboardEntry[]> {
  // Utilise les 7 derniers jours
  const supabase = createClient();
  const weekAgo = new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString();

  const { data } = await supabase
    .from('game_sessions')
    .select('user_id, score, user:user_id(username, avatar_url, rank)')
    .gte('completed_at', weekAgo)
    .order('score', { ascending: false })
    .limit(50);

  return (data || []).map((entry: any, index: number) => ({
    rank: index + 1,
    user_id: entry.user_id,
    username: entry.user?.username || 'Unknown',
    avatar_url: entry.user?.avatar_url,
    user_rank: entry.user?.rank || 'F',
    score: entry.score,
  })) as LeaderboardEntry[];
}

export async function getSeriesLeaderboard(series: string, page: number = 1): Promise<LeaderboardEntry[]> {
  const supabase = createClient();
  const { data } = await supabase
    .rpc('get_series_leaderboard', { series_name: series, limit_count: 50 });
  return (data || []) as LeaderboardEntry[];
}

export async function getQuizLeaderboard(quizId: string, page: number = 1): Promise<LeaderboardEntry[]> {
  const supabase = createClient();
  const { data } = await supabase
    .rpc('get_quiz_leaderboard', { quiz_id: quizId });
  return (data || []) as LeaderboardEntry[];
}

export async function getUserRankInLeaderboard(userId: string, type: string): Promise<number> {
  const supabase = createClient();

  if (type === 'global') {
    const { data } = await supabase
      .from('user_profiles')
      .select('id')
      .order('xp', { ascending: false });

    if (!data) return 0;
    const index = data.findIndex((u) => u.id === userId);
    return index >= 0 ? index + 1 : 0;
  }

  return 0;
}

