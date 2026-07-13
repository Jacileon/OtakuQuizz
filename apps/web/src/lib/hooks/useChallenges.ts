'use client';

import { useState, useEffect, useCallback } from 'react';
import { ChallengeSession } from '@/types';
import { toast } from '@/lib/hooks/useToast';
import {
  createChallengeSession,
  getMyChallenges,
  getChallengeSession,
  getChallengeParticipationCount,
} from '@/lib/actions/challenges';

export function useMyChallenges() {
  const [challenges, setChallenges] = useState<ChallengeSession[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchChallenges = useCallback(async () => {
    try {
      setLoading(true);
      const data = await getMyChallenges();
      setChallenges(data);
    } catch (error) {
      console.error('Erreur chargement défis:', error);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchChallenges();
  }, [fetchChallenges]);

  return { challenges, loading, refetch: fetchChallenges };
}

export function useChallengeSession(sessionId: string | null) {
  const [session, setSession] = useState<ChallengeSession | null>(null);
  const [loading, setLoading] = useState(true);

  const fetchSession = useCallback(async () => {
    if (!sessionId) return;
    try {
      setLoading(true);
      const data = await getChallengeSession(sessionId);
      setSession(data);
    } catch (error) {
      console.error('Erreur chargement session:', error);
    } finally {
      setLoading(false);
    }
  }, [sessionId]);

  useEffect(() => {
    fetchSession();
  }, [fetchSession]);

  return { session, loading, refetch: fetchSession };
}

export function useChallengeParticipationCount(quizId: string) {
  const [count, setCount] = useState(0);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchCount = async () => {
      try {
        const c = await getChallengeParticipationCount(quizId);
        setCount(c);
      } catch (error) {
        console.error('Erreur comptage participations:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchCount();
  }, [quizId]);

  return { count, loading, remaining: Math.max(0, 3 - count) };
}