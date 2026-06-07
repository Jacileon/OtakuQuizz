import { redirect } from 'next/navigation';
import Link from 'next/link';
import { createClient } from '@/lib/supabase/server';
import { getCurrentProfile } from '@/lib/auth/actions';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { RankBadge } from '@/components/ui/RankBadge';
import { Play, Clock, Users, Trophy, Edit, ArrowLeft, Gamepad2 } from 'lucide-react';

export default async function QuizDetailPage({
  params,
}: {
  params: { id: string };
}) {
  const supabase = createClient();
  const profile = await getCurrentProfile();

  const { data: quiz, error } = await supabase
    .from('quizzes')
    .select('*, creator:creator_id(username, avatar_url), questions(id)')
    .eq('id', params.id)
    .single();

  if (error || !quiz) {
    redirect('/explore');
  }

  const questionCount = quiz.questions?.length || quiz.question_count || 0;
  const isCreator = profile?.id === quiz.creator_id;

  return (
    <div className="p-4 lg:p-8 pb-24 md:pb-8">
      <div className="max-w-4xl mx-auto">
        {/* Header */}
        <div className="mb-6">
          <Link href="/explore" className="inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-white mb-4 transition-colors">
            <ArrowLeft className="h-4 w-4" />
            Retour à l'exploration
          </Link>
        </div>

        {/* Quiz Card */}
        <Card className="border-dark-border bg-dark-card overflow-hidden">
          {/* Banner */}
          <div className="h-48 bg-gradient-to-br from-brand/20 to-accent/20 flex items-center justify-center relative">
            <Gamepad2 className="h-20 w-20 text-white/20" />
            {quiz.quiz_type === 'official' && (
              <div className="absolute top-4 right-4 px-3 py-1 bg-brand rounded-full text-xs font-bold">
                OFFICIEL
              </div>
            )}
          </div>

          <CardContent className="p-6 space-y-6">
            {/* Title & Creator */}
            <div>
              <h1 className="font-display text-3xl tracking-wider mb-2">{quiz.title}</h1>
              {quiz.description && (
                <p className="text-muted-foreground mb-4">{quiz.description}</p>
              )}
              {quiz.creator && (
                <div className="flex items-center gap-2 text-sm">
                  <div className="h-6 w-6 rounded-full bg-dark-surface" />
                  <span className="text-muted-foreground">Créé par</span>
                  <span className="font-medium">{(quiz.creator as any).username}</span>
                </div>
              )}
            </div>

            {/* Stats */}
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
              <div className="p-3 rounded-lg bg-dark-surface text-center">
                <div className="font-display text-2xl text-brand">{questionCount}</div>
                <div className="text-xs text-muted-foreground">Questions</div>
              </div>
              <div className="p-3 rounded-lg bg-dark-surface text-center">
                <div className="font-display text-2xl text-accent">{quiz.play_count}</div>
                <div className="text-xs text-muted-foreground">Parties</div>
              </div>
              <div className="p-3 rounded-lg bg-dark-surface text-center">
                <div className="text-sm font-medium">{quiz.category}</div>
                <div className="text-xs text-muted-foreground">Catégorie</div>
              </div>
              <div className="p-3 rounded-lg bg-dark-surface text-center">
                <div className="text-sm font-medium">{quiz.series}</div>
                <div className="text-xs text-muted-foreground">Série</div>
              </div>
            </div>

            {/* Actions */}
            <div className="flex flex-col sm:flex-row gap-3">
              <Link href={`/quiz/${quiz.id}/play`} className="flex-1">
                <Button size="lg" className="w-full gap-2">
                  <Play className="h-5 w-5" />
                  Jouer maintenant
                </Button>
              </Link>
              {isCreator && (
                <Link href={`/quiz/${quiz.id}/edit`}>
                  <Button variant="outline" size="lg" className="gap-2">
                    <Edit className="h-5 w-5" />
                    Modifier
                  </Button>
                </Link>
              )}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}