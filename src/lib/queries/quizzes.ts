'use server';

// ============================================================
// REQUÊTES QUIZZES - Supabase Server
// ============================================================

import { createClient } from '@/lib/supabase/server';
import { Quiz, PaginatedResponse, SearchParams } from '@/types';

export async function searchQuizzes(params: SearchParams): Promise<PaginatedResponse<Quiz>> {
  const supabase = createClient();
  const page = params.page || 1;
  const perPage = params.perPage || 12;
  const offset = (page - 1) * perPage;

  let query = supabase
    .from('quizzes')
    .select('*, creator:creator_id(username, avatar_url)', { count: 'exact' })
    .eq('is_visible', true);

  if (params.query) {
    query = query.or(`title.ilike.%${params.query}%,series.ilike.%${params.query}%`);
  }
  if (params.category) {
    query = query.eq('category', params.category);
  }
  if (params.subcategory) {
    query = query.eq('subcategory', params.subcategory);
  }
  if (params.series) {
    query = query.eq('series', params.series);
  }

  // Tri
  switch (params.sortBy) {
    case 'popular':
      query = query.order('play_count', { ascending: false });
      break;
    case 'recent':
      query = query.order('created_at', { ascending: false });
      break;
    case 'rated':
      query = query.order('total_reports', { ascending: true });
      break;
    default:
      query = query.order('play_count', { ascending: false });
  }

  const { data, count, error } = await query.range(offset, offset + perPage - 1);

  if (error) {
    return { data: [], count: 0, page, per_page: perPage, total_pages: 0 };
  }

  const totalPages = count ? Math.ceil(count / perPage) : 0;

  return {
    data: (data || []) as Quiz[],
    count: count || 0,
    page,
    per_page: perPage,
    total_pages: totalPages,
  };
}

export async function getQuizzesByCategory(
  category: string,
  subcategory?: string,
  page: number = 1
): Promise<PaginatedResponse<Quiz>> {
  return searchQuizzes({ category, subcategory, page });
}

export async function getQuizzesBySeries(series: string, page: number = 1): Promise<PaginatedResponse<Quiz>> {
  return searchQuizzes({ series, page });
}

export async function getPopularQuizzes(limit: number = 6): Promise<Quiz[]> {
  const supabase = createClient();
  const { data } = await supabase
    .from('quizzes')
    .select('*, creator:creator_id(username, avatar_url)')
    .eq('is_visible', true)
    .order('play_count', { ascending: false })
    .limit(limit);
  return (data || []) as Quiz[];
}

export async function getRecentQuizzes(limit: number = 6): Promise<Quiz[]> {
  const supabase = createClient();
  const { data } = await supabase
    .from('quizzes')
    .select('*, creator:creator_id(username, avatar_url)')
    .eq('is_visible', true)
    .order('created_at', { ascending: false })
    .limit(limit);
  return (data || []) as Quiz[];
}

export async function getQuizById(id: string): Promise<Quiz | null> {
  const supabase = createClient();
  const { data } = await supabase
    .from('quizzes')
    .select('*, creator:creator_id(username, avatar_url)')
    .eq('id', id)
    .eq('is_visible', true)
    .single();
  return data as Quiz | null;
}

export async function getQuizForPlay(id: string): Promise<any | null> {
  const supabase = createClient();
  const { data: quiz } = await supabase
    .from('quizzes')
    .select('*')
    .eq('id', id)
    .eq('is_visible', true)
    .single();

  if (!quiz) return null;

  const { data: questions } = await supabase
    .from('questions')
    .select(`
      id, quiz_id, question_text, question_type, media_url, media_public_id,
      time_limit_seconds, order_index,
      answers:answers(id, question_id, answer_text, order_index)
    `)
    .eq('quiz_id', id)
    .order('order_index', { ascending: true });

  return { ...quiz, questions: questions || [] };
}

export async function getAllSeries(): Promise<string[]> {
  const supabase = createClient();
  const { data } = await supabase
    .from('quizzes')
    .select('series')
    .eq('is_visible', true)
    .order('series');

  const series = Array.from(new Set((data || []).map((q: any) => q.series)));
  return series;
}

