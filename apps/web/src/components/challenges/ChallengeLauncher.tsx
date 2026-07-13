'use client';

import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Loader2, Swords, Zap, Users, AlertTriangle } from 'lucide-react';
import { createChallengeSession } from '@/lib/actions/challenges';
import { useChallengeParticipationCount } from '@/lib/hooks/useChallenges';
import { ChallengeWaitingRoom } from '@/components/challenges/ChallengeWaitingRoom';
import { toast } from '@/lib/hooks/useToast';
import { Quiz } from '@/types';

export function ChallengeLauncher({ quiz }: { quiz: Quiz }) {
  const [sessionId, setSessionId] = useState<string | null>(null);
  const [xpBet, setXpBet] = useState(100);
  const [creating, setCreating] = useState(false);
  const { count, remaining, loading: loadingCount } = useChallengeParticipationCount(quiz.id);

  const handleCreate = async () => {
    try {
      setCreating(true);
      const id = await createChallengeSession(quiz.id, xpBet);
      setSessionId(id);
      toast({ title: 'Défi créé !', description: 'Invitez vos amis à rejoindre' });
    } catch (error: any) {
      toast({ title: 'Erreur', description: error.message, variant: 'destructive' });
    } finally {
      setCreating(false);
    }
  };

  if (sessionId) {
    return <ChallengeWaitingRoom sessionId={sessionId} onStart={() => {}} />;
  }

  if (loadingCount) {
    return (
      <div className="flex justify-center py-12">
        <Loader2 className="h-8 w-8 animate-spin" />
      </div>
    );
  }

  if (remaining === 0) {
    return (
      <Card className="border-red-500/30">
        <CardContent className="p-8 text-center">
          <AlertTriangle className="h-12 w-12 text-red-500 mx-auto mb-4" />
          <h3 className="text-lg font-semibold mb-2">Limite atteinte</h3>
          <p className="text-muted-foreground">
            Vous avez déjà participé à 3 défis sur ce quiz.
          </p>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="border-purple-500/30 bg-gradient-to-br from-purple-500/5 to-pink-500/5">
      <CardHeader>
        <CardTitle className="flex items-center gap-3">
          <Swords className="h-6 w-6 text-purple-500" />
          Créer un défi
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-6">
        <div className="p-4 rounded-lg bg-dark-surface">
          <h3 className="font-medium mb-1">{quiz.title}</h3>
          <p className="text-sm text-muted-foreground">{quiz.question_count} questions</p>
        </div>

        <div className="flex items-center gap-2 p-3 rounded-lg bg-yellow-500/10 border border-yellow-500/30">
          <Users className="h-5 w-5 text-yellow-500" />
          <span className="text-sm">
            <strong>{remaining}</strong> défi(s) restant(s) sur ce quiz
          </span>
        </div>

        <div className="space-y-2">
          <label className="text-sm font-medium flex items-center gap-2">
            <Zap className="h-4 w-4 text-yellow-500" />
            Mise XP
          </label>
          <Input
            type="number"
            value={xpBet}
            onChange={(e) => setXpBet(parseInt(e.target.value) || 0)}
            min={10}
            max={10000}
            placeholder="Montant XP à miser"
          />
          <p className="text-xs text-muted-foreground">
            Montant d'XP que vous misez sur ce défi (minimum 10)
          </p>
        </div>

        <Button
          onClick={handleCreate}
          disabled={creating || xpBet < 10}
          className="w-full gap-2"
          size="lg"
        >
          {creating ? (
            <Loader2 className="h-5 w-5 animate-spin" />
          ) : (
            <Swords className="h-5 w-5" />
          )}
          Créer le défi
        </Button>
      </CardContent>
    </Card>
  );
}