'use server';

import { createClient } from '@/lib/supabase/server';
import { Quiz, QuizReward, OfficialLeaderboardEntry } from '@/types';
import { revalidatePath } from 'next/cache';

export async function createOfficialQuiz(data: {
  title: string;
  description: string;
  category: string;
  subcategory: string;
  series: string;
  starts_at: string;
  ends_at: string;
  duration_seconds: number;
  duration_mode: 'global' | 'per_question';
  rewards: QuizReward[];
}): Promise<string> {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) throw new Error('Non authentifié');

  const { data: profile } = await supabase
    .from('user_profiles')
    .select('is_admin')
    .eq('id', user.id)
    .single();

  if (!profile?.is_admin) throw new Error('Non autorisé');

  const now = new Date();
  const startsAt = new Date(data.starts_at);
  const status = startsAt > now ? 'scheduled' : 'active';

  const { data: quiz, error } = await supabase
    .from('quizzes')
    .insert({
      creator_id: user.id,
      title: data.title,
      description: data.description,
      category: data.category,
      subcategory: data.subcategory,
      series: data.series,
      quiz_type: 'official',
      status: status,
      starts_at: data.starts_at,
      ends_at: data.ends_at,
      duration_seconds: data.duration_seconds,
      duration_mode: data.duration_mode,
      leaderboard_public: true,
      is_visible: true,
    })
    .select('id')
    .single();

  if (error) throw new Error('Erreur création quiz');

  if (data.rewards && data.rewards.length > 0) {
    const rewardsToInsert = data.rewards.map(r => ({
      quiz_id: quiz.id,
      title: r.title,
      description: r.description,
      url: r.url,
      rank_from: r.rank_from,
      rank_to: r.rank_to,
    }));

    await supabase.from('quiz_rewards').insert(rewardsToInsert);
  }

  revalidatePath('/admin/official-quizzes');
  return quiz.id;
}

export async function updateOfficialQuiz(quizId: string, data: {
  title?: string;
  description?: string;
  category?: string;
  subcategory?: string;
  series?: string;
  starts_at?: string;
  ends_at?: string;
  duration_seconds?: number;
  duration_mode?: 'global' | 'per_question';
  status?: string;
  rewards?: QuizReward[];
}): Promise<void> {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) throw new Error('Non authentifié');

  const { data: profile } = await supabase
    .from('user_profiles')
    .select('is_admin')
    .eq('id', user.id)
    .single();

  if (!profile?.is_admin) throw new Error('Non autorisé');

  const updateData: any = {};
  if (data.title) updateData.title = data.title;
  if (data.description) updateData.description = data.description;
  if (data.category) updateData.category = data.category;
  if (data.subcategory) updateData.subcategory = data.subcategory;
  if (data.series) updateData.series = data.series;
  if (data.starts_at) updateData.starts_at = data.starts_at;
  if (data.ends_at) updateData.ends_at = data.ends_at;
  if (data.duration_seconds) updateData.duration_seconds = data.duration_seconds;
  if (data.duration_mode) updateData.duration_mode = data.duration_mode;
  if (data.status) updateData.status = data.status;

  const { error } = await supabase
    .from('quizzes')
    .update(updateData)
    .eq('id', quizId)
    .eq('quiz_type', 'official');

  if (error) throw new Error('Erreur mise à jour quiz');

  if (data.rewards) {
    await supabase.from('quiz_rewards').delete().eq('quiz_id', quizId);
    
    if (data.rewards.length > 0) {
      const rewardsToInsert = data.rewards.map(r => ({
        quiz_id: quizId,
        title: r.title,
        description: r.description,
        url: r.url,
        rank_from: r.rank_from,
        rank_to: r.rank_to,
      }));

      await supabase.from('quiz_rewards').insert(rewardsToInsert);
    }
  }

  revalidatePath('/admin/official-quizzes');
  revalidatePath(`/quiz/${quizId}`);
}

export async function getOfficialQuizzes(): Promise<Quiz[]> {
  const supabase = await createClient();

  const { data } = await supabase
    .from('quizzes')
    .select('*')
    .eq('quiz_type', 'official')
    .neq('status', 'deleted')
    .order('starts_at', { ascending: false });

  return (data as Quiz[]) || [];
}

export async function getActiveOfficialQuizzes(): Promise<Quiz[]> {
  const supabase = await createClient();

  const { data } = await supabase
    .from('quizzes')
    .select('*')
    .eq('quiz_type', 'official')
    .in('status', ['active', 'scheduled'])
    .eq('is_visible', true)
    .order('starts_at', { ascending: true });

  return (data as Quiz[]) || [];
}

export async function getOfficialQuizDetails(quizId: string): Promise<Quiz & { rewards: QuizReward[] } | null> {
  const supabase = await createClient();

  const { data: quiz } = await supabase
    .from('quizzes')
    .select('*')
    .eq('id', quizId)
    .eq('quiz_type', 'official')
    .single();

  if (!quiz) return null;

  const { data: rewards } = await supabase
    .from('quiz_rewards')
    .select('*')
    .eq('quiz_id', quizId)
    .order('rank_from', { ascending: true });

  return {
    ...quiz,
    rewards: (rewards as QuizReward[]) || [],
  } as any;
}

export async function getOfficialLeaderboard(quizId: string): Promise<OfficialLeaderboardEntry[]> {
  const supabase = await createClient();

  const { data } = await supabase
    .rpc('get_quiz_leaderboard', { quiz_id: quizId });

  return (data as OfficialLeaderboardEntry[]) || [];
}

export async function archiveExpiredOfficialQuizzes(): Promise<void> {
  const supabase = await createClient();
  
  await supabase.rpc('auto_archive_expired_official_quizzes');
}

export async function submitOfficialQuizResult(data: {
  quizId: string;
  sessionId: string;
  score: number;
  correctCount: number;
  totalQuestions: number;
  accuracyRate: number;
  timeTakenMs: number;
}): Promise<void> {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) throw new Error('Non authentifié');

  const { data: quiz } = await supabase
    .from('quizzes')
    .select('id, leaderboard_public')
    .eq('id', data.quizId)
    .eq('quiz_type', 'official')
    .single();

  if (!quiz || !quiz.leaderboard_public) return;

  await supabase.from('official_leaderboard').insert({
    quiz_id: data.quizId,
    user_id: user.id,
    session_id: data.sessionId,
    score: data.score,
    accuracy_rate: data.accuracyRate,
    time_taken_ms: data.timeTakenMs,
  });
}