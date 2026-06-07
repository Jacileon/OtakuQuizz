'use server';

// ============================================================
// REQUÊTES ÉVÉNEMENTS - Supabase Server
// ============================================================

import { createClient } from '@/lib/supabase/server';
import { Quiz, PaginatedResponse } from '@/types';

export async function getActiveEvents(): Promise<Quiz[]> {
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

export async function getUpcomingEvents(): Promise<Quiz[]> {
  const supabase = createClient();
  const now = new Date().toISOString();

  const { data } = await supabase
    .from('quizzes')
    .select('*, creator:creator_id(username, avatar_url)')
    .eq('quiz_type', 'official')
    .eq('is_visible', true)
    .gt('event_start_at', now)
    .order('event_start_at', { ascending: true })
    .limit(5);

  return (data || []) as Quiz[];
}

export async function getPastEvents(page: number = 1): Promise<PaginatedResponse<Quiz>> {
  const supabase = createClient();
  const now = new Date().toISOString();
  const perPage = 10;
  const offset = (page - 1) * perPage;

  const { data, count } = await supabase
    .from('quizzes')
    .select('*, creator:creator_id(username, avatar_url)', { count: 'exact' })
    .eq('quiz_type', 'official')
    .lt('event_end_at', now)
    .order('event_end_at', { ascending: false })
    .range(offset, offset + perPage - 1);

  return {
    data: (data || []) as Quiz[],
    count: count || 0,
    page,
    per_page: perPage,
    total_pages: count ? Math.ceil(count / perPage) : 0,
  };
}

export async function canJoinEvent(quizId: string, userId: string): Promise<boolean> {
  const supabase = createClient();
  const { data: quiz } = await supabase
    .from('quizzes')
    .select('event_start_at, event_end_at')
    .eq('id', quizId)
    .single();

  if (!quiz || !quiz.event_start_at || !quiz.event_end_at) return false;

  const now = new Date();
  const start = new Date(quiz.event_start_at);
  const end = new Date(quiz.event_end_at);

  return now >= start && now <= end;
}

