'use server';

import { createClient } from '@/lib/supabase/server';
import { Rank } from '@/types';
import { revalidatePath } from 'next/cache';

export async function canUserCreateQuiz(): Promise<{ allowed: boolean; reason?: string }> {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) return { allowed: false, reason: 'Non authentifié' };

  // Vérifier si admin
  const { data: profile } = await supabase
    .from('user_profiles')
    .select('is_admin, can_create_quiz, rank')
    .eq('id', user.id)
    .single();

  if (!profile) return { allowed: false, reason: 'Profil introuvable' };

  // Admin = toujours autorisé
  if (profile.is_admin) return { allowed: true };

  // Autorisation individuelle
  if (profile.can_create_quiz) return { allowed: true };

  // Vérifier les rangs autorisés
  const { data: config } = await supabase
    .from('app_config')
    .select('value')
    .eq('key', 'quiz_creation_allowed_ranks')
    .single();

  const allowedRanks: string[] = config?.value || ['C', 'E', 'S'];

  if (allowedRanks.includes(profile.rank)) {
    return { allowed: true };
  }

  return {
    allowed: false,
    reason: `Vous devez atteindre le rang ${allowedRanks[0]} pour créer un quiz.`
  };
}

export async function getQuizCreationConfig(): Promise<{ allowedRanks: string[] }> {
  const supabase = await createClient();
  
  const { data: config } = await supabase
    .from('app_config')
    .select('value')
    .eq('key', 'quiz_creation_allowed_ranks')
    .single();

  return { allowedRanks: config?.value || ['C', 'E', 'S'] };
}

export async function updateQuizCreationConfig(ranks: string[]): Promise<void> {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) throw new Error('Non authentifié');

  const { data: profile } = await supabase
    .from('user_profiles')
    .select('is_admin')
    .eq('id', user.id)
    .single();

  if (!profile?.is_admin) throw new Error('Non autorisé');

  await supabase
    .from('app_config')
    .update({ value: ranks, updated_at: new Date().toISOString(), updated_by: user.id })
    .eq('key', 'quiz_creation_allowed_ranks');

  revalidatePath('/admin/settings');
}

export async function toggleUserCanCreateQuiz(userId: string, canCreate: boolean): Promise<void> {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) throw new Error('Non authentifié');

  const { data: profile } = await supabase
    .from('user_profiles')
    .select('is_admin')
    .eq('id', user.id)
    .single();

  if (!profile?.is_admin) throw new Error('Non autorisé');

  await supabase
    .from('user_profiles')
    .update({ can_create_quiz: canCreate })
    .eq('id', userId);

  revalidatePath('/admin/users');
}