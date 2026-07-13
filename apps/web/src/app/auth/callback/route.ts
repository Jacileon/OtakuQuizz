// ============================================================
// ROUTE CALLBACK OAUTH GOOGLE
// ============================================================

import { NextResponse } from 'next/server';
import { createClient } from '@/lib/supabase/server';

export async function GET(request: Request) {
  const { searchParams, origin } = new URL(request.url);
  const code = searchParams.get('code');
  const next = searchParams.get('next') ?? '/dashboard';

  if (!code) {
    return NextResponse.redirect(`${origin}/login?error=no_code`);
  }

  const supabase = createClient();
  const { error } = await supabase.auth.exchangeCodeForSession(code);

  if (error) {
    return NextResponse.redirect(`${origin}/login?error=auth_failed`);
  }

  // Vérifier/créer le profil utilisateur
  const { data: { user } } = await supabase.auth.getUser();

  if (user) {
    const { data: existingProfile } = await supabase
      .from('user_profiles')
      .select('id')
      .eq('id', user.id)
      .single();

    if (!existingProfile) {
      // Créer le profil pour le premier login
      const username = user.user_metadata?.full_name?.replace(/\s+/g, '_').toLowerCase() 
        || user.email?.split('@')[0] 
        || 'user_' + Math.random().toString(36).slice(2, 8);

      await supabase.from('user_profiles').insert({
        id: user.id,
        email: user.email!,
        username: username.slice(0, 30),
        avatar_url: user.user_metadata?.avatar_url || null,
        xp: 0,
        level: 1,
        rank: 'F',
        is_premium: false,
      });

      // Créer les stats initiales
      await supabase.from('user_stats').insert({
        user_id: user.id,
        quizzes_played: 0,
        quizzes_created: 0,
        total_correct_answers: 0,
        total_answers: 0,
        accuracy_rate: 0,
        best_score_ever: 0,
      });
    }
  }

  return NextResponse.redirect(`${origin}${next}`);
}
