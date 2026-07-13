import { redirect } from 'next/navigation';
import { createClient } from '@/lib/supabase/server';
import { OfficialLeaderboard } from '@/components/leaderboard/OfficialLeaderboard';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Trophy, Shield, Clock, Calendar, ArrowLeft, Play } from 'lucide-react';
import Link from 'next/link';

export default async function OfficialQuizLeaderboardPage({
  params,
}: {
  params: { id: string };
}) {
  const supabase = createClient();

  const { data: quiz, error } = await supabase
    .from('quizzes')
    .select('*')
    .eq('id', params.id)
    .eq('quiz_type', 'official')
    .single();

  if (error || !quiz || !quiz.leaderboard_public) {
    redirect('/explore');
  }

  const { data: leaderboard } = await supabase
    .rpc('get_quiz_leaderboard', { quiz_id: params.id });

  const { data: rewards } = await supabase
    .from('quiz_rewards')
    .select('*')
    .eq('quiz_id', params.id)
    .order('rank_from', { ascending: true });

  const isArchived = quiz.status === 'archived';
  const isActive = quiz.status === 'active';

  return (
    <div className="p-4 lg:p-8 pb-24 md:pb-8">
      <div className="max-w-4xl mx-auto">
        <div className="mb-6">
          <Link href="/explore" className="inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-white mb-4 transition-colors">
            <ArrowLeft className="h-4 w-4" />
            Retour à l'exploration
          </Link>
        </div>

        <Card className="mb-8 border-yellow-500/30 bg-gradient-to-br from-yellow-500/5 to-amber-500/5">
          <CardHeader>
            <div className="flex items-start justify-between">
              <div>
                <div className="flex items-center gap-3 mb-2">
                  <Shield className="h-6 w-6 text-yellow-500" />
                  <Badge variant="default" className="bg-gradient-to-r from-yellow-500 to-amber-500">
                    Quiz Officiel
                  </Badge>
                  {isArchived && (
                    <Badge variant="secondary">Archivé</Badge>
                  )}
                </div>
                <CardTitle className="text-2xl">{quiz.title}</CardTitle>
                {quiz.description && (
                  <p className="text-muted-foreground mt-2">{quiz.description}</p>
                )}
              </div>
              {isActive && (
                <Link href={`/quiz/${quiz.id}/play`}>
                  <Button className="gap-2">
                    <Play className="h-4 w-4" />
                    Jouer
                  </Button>
                </Link>
              )}
            </div>
          </CardHeader>
          <CardContent>
            <div className="flex flex-wrap gap-4 text-sm text-muted-foreground">
              {quiz.starts_at && (
                <span className="flex items-center gap-1">
                  <Calendar className="h-4 w-4" />
                  Début: {new Date(quiz.starts_at).toLocaleDateString('fr-FR')}
                </span>
              )}
              {quiz.ends_at && (
                <span className="flex items-center gap-1">
                  <Clock className="h-4 w-4" />
                  Fin: {new Date(quiz.ends_at).toLocaleDateString('fr-FR')}
                </span>
              )}
              <span>{quiz.question_count || 0} questions</span>
              <span>{quiz.play_count || 0} parties</span>
            </div>
          </CardContent>
        </Card>

        <div className="mb-6">
          <h2 className="font-display text-2xl tracking-wider flex items-center gap-3">
            <Trophy className="h-8 w-8 text-yellow-500" />
            CLASSEMENT OFFICIEL
          </h2>
        </div>

        {leaderboard && leaderboard.length > 0 ? (
          <OfficialLeaderboard
            entries={leaderboard}
            rewards={rewards || []}
            showPodium={true}
          />
        ) : (
          <Card>
            <CardContent className="flex flex-col items-center justify-center py-16">
              <Trophy className="h-16 w-16 text-muted-foreground mb-4 opacity-50" />
              <h3 className="text-lg font-semibold mb-2">Aucun score</h3>
              <p className="text-muted-foreground text-center">
                Soyez le premier à jouer à ce quiz officiel !
              </p>
              {isActive && (
                <Link href={`/quiz/${quiz.id}/play`} className="mt-4">
                  <Button className="gap-2">
                    <Play className="h-4 w-4" />
                    Jouer maintenant
                  </Button>
                </Link>
              )}
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  );
}