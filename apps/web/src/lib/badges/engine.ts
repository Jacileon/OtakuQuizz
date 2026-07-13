'use server';

// ============================================================
// MOTEUR D'ATTRIBUTION DES BADGES
// ============================================================

import { createClient } from '@/lib/supabase/server';
import { Badge } from '@/types';

export async function checkAllBadges(userId: string): Promise<Badge[]> {
  const supabase = createClient();

  const { data: newBadges } = await supabase
    .rpc('check_and_award_badges', { target_user_id: userId });

  if (newBadges && newBadges.length > 0) {
    // Attribuer les badges
    for (const badge of newBadges) {
      await supabase.from('user_badges').insert({
        user_id: userId,
        badge_id: badge.badge_id,
      });
    }
  }

  return (newBadges || []) as Badge[];
}

export async function awardBadge(userId: string, badgeSlug: string): Promise<void> {
  const supabase = createClient();

  const { data: badge } = await supabase
    .from('badges')
    .select('id')
    .eq('slug', badgeSlug)
    .single();

  if (!badge) return;

  await supabase.from('user_badges').insert({
    user_id: userId,
    badge_id: badge.id,
  });
}

