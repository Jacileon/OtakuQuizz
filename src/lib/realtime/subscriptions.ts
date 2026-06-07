'use client';

// ============================================================
// SUBSCRIPTIONS REALTIME - Supabase
// ============================================================

import { getBrowserClient } from '@/lib/supabase/client';
import { RealtimeChannel } from '@supabase/supabase-js';

export function subscribeToLeaderboard(
  quizId: string,
  callback: (payload: any) => void
): RealtimeChannel {
  const supabase = getBrowserClient();

  return supabase
    .channel(`leaderboard:${quizId}`)
    .on(
      'postgres_changes',
      {
        event: 'INSERT',
        schema: 'public',
        table: 'game_sessions',
        filter: `quiz_id=eq.${quizId}`,
      },
      callback
    )
    .subscribe();
}

export function subscribeToUserXP(
  userId: string,
  callback: (payload: any) => void
): RealtimeChannel {
  const supabase = getBrowserClient();

  return supabase
    .channel(`user_xp:${userId}`)
    .on(
      'postgres_changes',
      {
        event: 'UPDATE',
        schema: 'public',
        table: 'user_profiles',
        filter: `id=eq.${userId}`,
      },
      callback
    )
    .subscribe();
}

export function unsubscribeAll(channels: RealtimeChannel[]): void {
  channels.forEach((channel) => {
    channel.unsubscribe();
  });
}

