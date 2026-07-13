'use server';

import { createClient } from '@/lib/supabase/server';

export async function checkAndUpdateStreak(): Promise<{ 
  streak: number; 
  isNew: boolean; 
  xpEarned: number;
  message: string;
}> {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) return { streak: 0, isNew: false, xpEarned: 0, message: '' };

  const { data: profile } = await supabase
    .from('user_profiles')
    .select('current_streak, longest_streak, last_login_date')
    .eq('id', user.id)
    .single();

  if (!profile) return { streak: 0, isNew: false, xpEarned: 0, message: '' };

  const today = new Date().toISOString().split('T')[0];
  const lastLogin = profile.last_login_date;

  // Déjà connecté aujourd'hui
  if (lastLogin === today) {
    return { 
      streak: profile.current_streak, 
      isNew: false, 
      xpEarned: 0,
      message: '' 
    };
  }

  const yesterday = new Date();
  yesterday.setDate(yesterday.getDate() - 1);
  const yesterdayStr = yesterday.toISOString().split('T')[0];

  let newStreak = 1;
  let xpEarned = 2;
  let message = 'Nouveau streak démarré ! +2 XP';

  // Connexion consécutive (hier)
  if (lastLogin === yesterdayStr) {
    newStreak = profile.current_streak + 1;
    message = `🔥 Connexion jour ${newStreak} ! +2 XP`;
  }

  const newLongestStreak = Math.max(profile.longest_streak, newStreak);

  // Mettre à jour le profil
  await supabase
    .from('user_profiles')
    .update({
      current_streak: newStreak,
      longest_streak: newLongestStreak,
      last_login_date: today,
    })
    .eq('id', user.id);

  // Ajouter l'XP
  await supabase.rpc('increment_user_xp', {
    user_id: user.id,
    amount: xpEarned,
  });

  // Enregistrer la transaction XP
  await supabase.from('xp_transactions').insert({
    user_id: user.id,
    source: 'streak',
    amount: xpEarned,
  });

  return { streak: newStreak, isNew: true, xpEarned, message };
}