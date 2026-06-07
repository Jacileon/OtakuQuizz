// ============================================================
// REQUÊTES DASHBOARD - Supabase Server
// ============================================================

'use server';

import { createClient } from '@/lib/supabase/server';
import { Quiz, UserBadge, GameSession, UserProfile } from '@/types';

export async function getDashboardStats(userId: string) {
  const supabase = createClient();

  const { data: stats } = await supabase
    .from('user_stats')
    .select('*')
    .eq('user_id', userId)
    .single();

  const { data: recentSessions } = await supabase
    .from('game_sessions')
    .select('*')
    .eq('user_id', userId)
    .not('completed_at', 'is', null)
    .gte('completed_at', new Date(new Date().getFullYear(), new Date().getMonth(), 1).toISOString())
    .order('completed_at', { ascending: false });

  const monthlyQuizzes = recentSessions?.length || 0;
  const bestScore = stats?.best_score_ever || 0;
  const accuracy = stats?.accuracy_rate || 0;

  const { data: monthlyRank } = await supabase
    .from('leaderboard_monthly')
    .select('rank_position')
    .eq('user_id', userId)
    .eq('month_year', `${new Date().getFullYear()}-${String(new Date().getMonth() + 1).padStart(2, '0')}`)
    .single();

  return {
    monthlyQuizzes,
    bestScore,
    accuracy,
    monthlyRank: monthlyRank?.rank_position || null,
  };
}

export async function getActiveOfficialQuizzes(): Promise<Quiz[]> {
  const supabase = createClient();
  const now = new Date().toISOString();

  const { data } = await supabase
    .from('quizzes')
    .select('*, creator:creator_id(username, avatar_url)')
    .eq('quiz_type', 'official')
    .eq('is_visible', true)
    .lte('event_start_at', now)
    .gte('event_end_at', now)
    .order('event_start_at', { ascending: true });

  return (data || []) as Quiz[];
}

export async function getUpcomingOfficialQuizzes(): Promise<Quiz[]> {
  const supabase = createClient();
  const now = new Date().toISOString();

  const { data } = await supabase
    .from('quizzes')
    .select('*, creator:creator_id(username, avatar_url)')
    .eq('quiz_type', 'official')
    .eq('is_visible', true)
    .gt('event_start_at', now)
    .order('event_start_at', { ascending: true })
    .limit(3);

  return (data || []) as Quiz[];
}

export async function getRecommendedQuizzes(userId: string, series: string): Promise<Quiz[]> {
  const supabase = createClient();

  const { data } = await supabase
    .from('quizzes')
    .select('*, creator:creator_id(username, avatar_url)')
    .eq('is_visible', true)
    .eq('series', series)
    .neq('creator_id', userId)
    .limit(6);

  if (!data || data.length < 3) {
    const { data: fallback } = await supabase
      .from('quizzes')
      .select('*, creator:creator_id(username, avatar_url)')
      .eq('is_visible', true)
      .order('play_count', { ascending: false })
      .limit(6);
    return (fallback || []) as Quiz[];
  }

  return (data || []) as Quiz[];
}

export async function getRecentActivity(userId: string, limit: number = 5): Promise<GameSession[]> {
  const supabase = createClient();

  const { data } = await supabase
    .from('game_sessions')
    .select('*, quiz:quiz_id(title, series, thumbnail_url)')
    .eq('user_id', userId)
    .not('completed_at', 'is', null)
    .order('completed_at', { ascending: false })
    .limit(limit);

  return (data || []) as GameSession[];
}

export async function getRecentBadges(userId: string, limit: number = 3): Promise<UserBadge[]> {
  const supabase = createClient();

  const { data } = await supabase
    .from('user_badges')
    .select('*, badge:badge_id(*)')
    .eq('user_id', userId)
    .order('earned_at', { ascending: false })
    .limit(limit);

  return (data || []) as UserBadge[];
}

