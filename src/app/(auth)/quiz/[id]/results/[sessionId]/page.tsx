// ============================================================
// PAGE DE RÉSULTATS
// ============================================================

import Link from '../../../../../../../node_modules/next/link';
import { redirect } from '../../../../../../../node_modules/next/navigation';
import { createClient } from '@/lib/supabase/server';
import { getCurrentUser } from '@/lib/auth/actions';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { RankBadge } from '@/components/ui/RankBadge';
import { Progress } from '@/components/ui/progress';
import { BadgeUnlockModal } from '@/components/notifications/BadgeUnlockModal';
import { Trophy, Target, Clock, RotateCcw, Share2, BarChart3, Star, Zap } from 'lucide-react';

import { formatDuration, cn } from '@/lib/utils';

export default async function ResultsPage({
  params,
}: {
  params: Promise<{ id: string; sessionId: string }>;
}) {
  const { id, sessionId } = await params;
  const user = await getCurrentUser();
  if (!user) redirect('/login');

  const supabase = createClient();

  const { data: session } = await supabase
    .from('game_sessions')
    .select('*, quiz:quiz_id(title, series)')
    .eq('id', sessionId)
    .eq('user_id', user.id)
    .single();

  if (!session || !session.completed_at) {
    redirect(`/quiz/${id}/play`);
  }

  const { data: playerAnswers } = await supabase
    .from('player_answers')
    .select(`
      *,
      question:question_id(question_text, question_type, media_url),
      answer:answer_id(answer_text),
      correct_answer:questions!inner(answers!inner(answer_text, is_correct))
    `)

    .eq('session_id', sessionId);
  // Récupérer les badges récemment débloqués
  const { data: recentBadges } = await supabase
    .from('user_badges')
    .select('*, badge:badge_id(*)')
    .eq('user_id', user.id)
    .gte('earned_at', new Date(Date.now() - 60000).toISOString())
    .order('earned_at', { ascending: false });

  const isPerfect = session.is_perfect;
  const accuracy = session.accuracy_rate;

  return (
    <div className="min-h-screen bg-dark p-4 lg:p-8">
      <div className="max-w-3xl mx-auto space-y-8">
        {/* Header */}
        <div className="text-center space-y-4">
          {isPerfect && (
            <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-yellow-500/10 border border-yellow-500/30 text-yellow-400">
              <Star className="h-5 w-5" />
              <span className="font-display text-lg tracking-wider">QUIZ PARFAIT</span>
              <Star className="h-5 w-5" />
            </div>
          )}
          <h1 className="font-display text-4xl tracking-wider">
            {(session as any).quiz?.title}
          </h1>
          <p className="text-muted-foreground">{(session as any).quiz?.series}</p>
        </div>

        {/* Score Card */}
        <Card className={cn(
          'border-2',
          isPerfect ? 'border-yellow-500/50 bg-yellow-500/5' : 'border-brand/30 bg-dark-card'
        )}>
          <CardContent className="p-8 text-center space-y-6">
            <div>
              <div className="font-display text-7xl text-brand mb-2">
                {session.score}
              </div>
              <div className="text-muted-foreground">points</div>
            </div>

            <div className="flex items-center justify-center gap-2 text-accent">
              <Zap className="h-5 w-5" />
              <span className="font-medium">+{session.score} XP</span>
            </div>

            <div className="grid grid-cols-3 gap-4 pt-4">
              <div className="text-center">
                <Target className="h-5 w-5 mx-auto mb-1 text-accent" />
                <div className="font-display text-xl">{accuracy}%</div>
                <div className="text-xs text-muted-foreground">Précision</div>
              </div>
              <div className="text-center">
                <Trophy className="h-5 w-5 mx-auto mb-1 text-brand" />
                <div className="font-display text-xl">
                  {session.correct_count}/{session.total_questions}
                </div>
                <div className="text-xs text-muted-foreground">Correct</div>
              </div>
              <div className="text-center">
                <Clock className="h-5 w-5 mx-auto mb-1 text-green-400" />
                <div className="font-display text-xl">
                  {formatDuration(session.time_taken_ms || 0)}
                </div>
                <div className="text-xs text-muted-foreground">Temps</div>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Réponses détaillées */}
        <div>
          <h2 className="font-display text-xl tracking-wider mb-4">DÉTAIL DES RÉPONSES</h2>
          <div className="space-y-3">
            {playerAnswers?.map((pa: any, index: number) => (
              <Card
                key={pa.id}
                className={cn(
                  'border-dark-border',
                  pa.is_correct ? 'bg-green-500/5 border-green-500/20' : 'bg-red-500/5 border-red-500/20'
                )}
              >
                <CardContent className="p-4">
                  <div className="flex items-start gap-3">
                    <div className={cn(
                      'h-6 w-6 rounded-full flex items-center justify-center text-xs font-bold shrink-0',
                      pa.is_correct ? 'bg-green-500 text-white' : 'bg-red-500 text-white'
                    )}>
                      {index + 1}
                    </div>
                    <div className="flex-1 space-y-2">
                      <p className="text-sm font-medium">{pa.question?.question_text}</p>
                      <div className="flex flex-wrap gap-2 text-xs">
                        <span className={cn(
                          'px-2 py-1 rounded',
                          pa.is_correct ? 'bg-green-500/20 text-green-400' : 'bg-red-500/20 text-red-400'
                        )}>
                          Ta réponse: {pa.answer?.answer_text || 'Non répondu'}
                        </span>
                        {!pa.is_correct && (
                          <span className="px-2 py-1 rounded bg-green-500/20 text-green-400">
                            Bonne réponse: {pa.correct_answer?.answers?.find((a: any) => a.is_correct)?.answer_text}
                          </span>
                        )}
                      </div>
                      <div className="text-xs text-muted-foreground">
                        +{pa.points_earned} pts • {pa.time_taken_ms}ms
                      </div>
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>

        {/* Actions */}
        <div className="flex flex-wrap gap-4 justify-center">
          <Link href={`/quiz/${id}/play`}>
            <Button className="gap-2">
              <RotateCcw className="h-4 w-4" /> Rejouer
            </Button>
          </Link>
          <Link href={`/leaderboard/quiz/${id}`}>
            <Button variant="outline" className="gap-2">
              <BarChart3 className="h-4 w-4" /> Classement
            </Button>
          </Link>
          <Button variant="outline" className="gap-2">
            <Share2 className="h-4 w-4" /> Partager
          </Button>
        </div>
      </div>

      {/* Modal badges */}
      {recentBadges && recentBadges.length > 0 && (
        <BadgeUnlockModal badges={recentBadges.map((ub: any) => ub.badge)} />
      )}
    </div>
  );
}
