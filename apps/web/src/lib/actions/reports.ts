'use server';

// ============================================================
// SERVER ACTIONS - SIGNALEMENTS
// ============================================================

import { createClient } from '@/lib/supabase/server';
import { getCurrentUser } from '@/lib/auth/actions';
import { ApiResponse, ReportReason } from '@/types';

export async function reportQuiz(
  quizId: string,
  reason: ReportReason,
  description?: string
): Promise<ApiResponse<void>> {
  try {
    const user = await getCurrentUser();
    if (!user) return { data: null, error: 'Non authentifié', success: false };

    const supabase = createClient();

    // Vérifier si déjà signalé
    const { data: existing } = await supabase
      .from('reports')
      .select('id')
      .eq('reporter_id', user.id)
      .eq('quiz_id', quizId)
      .single();

    if (existing) {
      return { data: null, error: 'Tu as déjà signalé ce quiz', success: false };
    }

    const { error } = await supabase.from('reports').insert({
      reporter_id: user.id,
      quiz_id: quizId,
      reason,
      description: description || null,
    });

    if (error) {
      return { data: null, error: 'Erreur lors du signalement', success: false };
    }

    return { data: null, error: null, success: true };
  } catch (error) {
    return { data: null, error: 'Erreur serveur', success: false };
  }
}

export async function reportUser(
  userId: string,
  reason: string,
  description?: string
): Promise<void> {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) throw new Error('Non authentifié');

  // Vérifier si déjà signalé
  const { data: existing } = await supabase
    .from('user_reports')
    .select('id')
    .eq('reporter_id', user.id)
    .eq('reported_user_id', userId)
    .single();

  if (existing) {
    throw new Error('Vous avez déjà signalé cet utilisateur');
  }

  const { error } = await supabase.from('user_reports').insert({
    reporter_id: user.id,
    reported_user_id: userId,
    reason,
    description: description || null,
  });

  if (error) throw new Error('Erreur lors du signalement');
}

