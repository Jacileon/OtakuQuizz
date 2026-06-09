'use server';

import { createClient } from '@/lib/supabase/server';

export async function checkProfileComplete(): Promise<{ complete: boolean; missing?: string }> {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) return { complete: false, missing: 'Non authentifié' };

  const { data: profile } = await supabase
    .from('user_profiles')
    .select('nickname, phone')
    .eq('id', user.id)
    .single();

  if (!profile) return { complete: false, missing: 'Profil introuvable' };

  if (!profile.nickname) {
    return { complete: false, missing: 'Veuillez compléter votre surnom (nickname)' };
  }

  if (!profile.phone) {
    return { complete: false, missing: 'Veuillez compléter votre numéro de téléphone' };
  }

  return { complete: true };
}