import { createClient } from '@/lib/supabase/client';
import { UserProfile } from '@/types';

export async function getUserProfile(username: string): Promise<UserProfile | null> {
  const supabase = createClient();
  const { data, error } = await supabase
    .from('user_profiles')
    .select('*')
    .eq('username', username)
    .single();

  if (error) return null;
  return data as UserProfile | null;
}

export async function getUserStats(userId: string) {
  const supabase = createClient();
  const { data } = await supabase
    .from('user_stats')
    .select('*')
    .eq('user_id', userId)
    .single();
  return data;
}

export async function getUserBadges(userId: string) {
  const supabase = createClient();
  const { data } = await supabase
    .from('user_badges')
    .select('*, badge:badge_id(*)')
    .eq('user_id', userId)
    .order('earned_at', { ascending: false });
  return data || [];
}

export async function getUserQuizzes(userId: string) {
  const supabase = createClient();
  const { data } = await supabase
    .from('quizzes')
    .select('*')
    .eq('creator_id', userId)
    .eq('is_visible', true)
    .order('play_count', { ascending: false });
  return data || [];
}

export async function getUserCollections(userId: string) {
  const supabase = createClient();

  const { data: sessions } = await supabase
    .from('game_sessions')
    .select('quiz:quiz_id(series)')
    .eq('user_id', userId)
    .not('completed_at', 'is', null);

  const seriesPlayed = Array.from(new Set((sessions || []).map((s: any) => s.quiz?.series).filter(Boolean)));

  const collections = [];
  for (const series of seriesPlayed) {
    const { count: total } = await supabase
      .from('quizzes')
      .select('*', { count: 'exact', head: true })
      .eq('series', series)
      .eq('is_visible', true);

    const { count: completed } = await supabase
      .from('game_sessions')
      .select('quiz:quiz_id!inner(*)', { count: 'exact', head: true })
      .eq('user_id', userId)
      .eq('quiz.series', series)
      .not('completed_at', 'is', null);

    const { data: best } = await supabase
      .from('game_sessions')
      .select('score')
      .eq('user_id', userId)
      .not('completed_at', 'is', null)
      .order('score', { ascending: false })
      .limit(1);

    collections.push({
      series,
      total_quizzes: total || 0,
      completed_quizzes: completed || 0,
      progress_percent: total ? Math.round(((completed || 0) / total) * 100) : 0,
      best_score: best?.[0]?.score || null,
    });
  }

  return collections;
}


