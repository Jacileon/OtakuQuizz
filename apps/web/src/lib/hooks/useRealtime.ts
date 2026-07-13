'use client';

import { useEffect, useRef } from '../../../node_modules/@types/react';
import { RealtimeChannel } from '@supabase/supabase-js';
import { subscribeToUserXP, unsubscribeAll } from '@/lib/realtime/subscriptions';

export function useRealtime(userId: string | null) {
  const channelsRef = useRef<RealtimeChannel[]>([]);

  useEffect(() => {
    if (!userId) return;

    const xpChannel = subscribeToUserXP(userId, (payload) => {
      console.log("Mise à jour XP:", payload);
    });

    channelsRef.current = [xpChannel];

    return () => {
      unsubscribeAll(channelsRef.current);
    };
  }, [userId]);
}

