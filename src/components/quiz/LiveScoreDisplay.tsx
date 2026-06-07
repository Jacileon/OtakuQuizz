'use client';

import { useEffect, useState } from '../../../node_modules/@types/react';
import { LeaderboardEntry } from '@/types';
import { subscribeToLeaderboard } from '@/lib/realtime/subscriptions';
import { Card, CardContent } from '@/components/ui/card';
import { Zap } from 'lucide-react';

interface LiveScoreDisplayProps {
  quizId: string;
}

export function LiveScoreDisplay({ quizId }: LiveScoreDisplayProps) {
  const [scores, setScores] = useState<LeaderboardEntry[]>([]);

  useEffect(() => {
    const channel = subscribeToLeaderboard(quizId, (payload) => {
      // Rafraîchir les scores
      console.log('Nouveau score:', payload);
    });

    return () => {
      channel.unsubscribe();
    };
  }, [quizId]);

  return (
    <Card className="border-brand/30 bg-brand/5">
      <CardContent className="p-4">
        <div className="flex items-center gap-2 mb-3">
          <Zap className="h-4 w-4 text-brand animate-pulse" />
          <span className="font-display text-sm tracking-wider">SCORES EN DIRECT</span>
        </div>
        <div className="space-y-2">
          {scores.slice(0, 5).map((entry) => (
            <div key={entry.user_id} className="flex items-center justify-between text-sm">
              <span>{entry.username}</span>
              <span className="font-medium text-brand">{entry.score}</span>
            </div>
          ))}
          {scores.length === 0 && (
            <p className="text-xs text-muted-foreground">En attente de joueurs...</p>
          )}
        </div>
      </CardContent>
    </Card>
  );
}

