import { redirect } from 'next/navigation';
import Link from 'next/link';
import { createClient } from '@/lib/supabase/server';
import { getCurrentProfile } from '@/lib/auth/actions';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Badge } from '@/components/ui/badge';
import { Play, Clock, Users, Trophy, Edit, ArrowLeft, Gamepad2, Medal, Swords } from 'lucide-react';

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

  const { data: leaderboard } = await supabase
    .rpc('get_quiz_leaderboard', { quiz_id: params.id });

  return (
    <div className="p-4 lg:p-8 pb-24 md:pb-8">
      <div className="max-w-4xl mx-auto">
        <div className="mb-6">
          <Link href="/explore" className="inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-white mb-4 transition-colors">
            <ArrowLeft className="h-4 w-4" />
            Retour à l'exploration
          </Link>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Quiz Card */}
          <div className="lg:col-span-2">
            <Card className="border-dark-border bg-dark-card overflow-hidden">
              <div className="h-48 bg-gradient-to-br from-brand/20 to-accent/20 flex items-center justify-center relative">
                {quiz.thumbnail_url ? (
                  <img src={quiz.thumbnail_url} alt={quiz.title} className="h-full w-full object-cover" />
                ) : (
                  <Gamepad2 className="h-20 w-20 text-white/20" />
                )}
                {quiz.quiz_type === 'official' && (
                  <div className="absolute top-4 right-4 px-3 py-1 bg-brand rounded-full text-xs font-bold">
                    OFFICIEL
                  </div>
                )}
              </div>

              <CardContent className="p-6 space-y-6">
                <div>
                  <h1 className="font-display text-3xl tracking-wider mb-2">{quiz.title}</h1>
                  {quiz.description && (
                    <p className="text-muted-foreground mb-4">{quiz.description}</p>
                  )}
                  {quiz.creator && (
                    <div className="flex items-center gap-2 text-sm">
                      <Avatar className="h-6 w-6">
                        <AvatarImage src={(quiz.creator as any).avatar_url || undefined} />
                        <AvatarFallback>{(quiz.creator as any).username?.[0]?.toUpperCase()}</AvatarFallback>
                      </Avatar>
                      <span className="text-muted-foreground">Créé par</span>
                      <span className="font-medium">{(quiz.creator as any).username}</span>
                    </div>
                  )}
                </div>

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
                    <div className="text-sm font-medium">
                      {quiz.duration_seconds ? `${quiz.duration_seconds}s` : '30s'}
                    </div>
                    <div className="text-xs text-muted-foreground">
                      {quiz.duration_mode === 'global' ? 'Durée totale' : 'Par question'}
                    </div>
                  </div>
                  <div className="p-3 rounded-lg bg-dark-surface text-center">
                    <div className="text-sm font-medium">{quiz.series}</div>
                    <div className="text-xs text-muted-foreground">Série</div>
                  </div>
                </div>

                <div className="flex flex-col sm:flex-row gap-3">
                  <Link href={`/quiz/${quiz.id}/play`} className="flex-1">
                    <Button size="lg" className="w-full gap-2">
                      <Play className="h-5 w-5" />
                      Jouer maintenant
                    </Button>
                  </Link>
                  <Link href={`/challenges/create/${quiz.id}`} className="flex-1">
                    <Button size="lg" variant="outline" className="w-full gap-2">
                      <Swords className="h-5 w-5" />
                      Défier vos amis
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

          {/* Leaderboard */}
          <div>
            <Card className="border-dark-border bg-dark-card">
              <CardHeader>
                <CardTitle className="flex items-center gap-2 text-lg">
                  <Trophy className="h-5 w-5 text-yellow-500" />
                  Classement
                </CardTitle>
              </CardHeader>
              <CardContent>
                {leaderboard && leaderboard.length > 0 ? (
                  <div className="space-y-3">
                    {leaderboard.slice(0, 10).map((entry: any, index: number) => (
                      <div
                        key={entry.user_id}
                        className="flex items-center gap-3 p-2 rounded-lg hover:bg-dark-surface transition-colors"
                      >
                        <div className="w-8 h-8 flex items-center justify-center">
                          {index === 0 ? (
                            <Medal className="h-6 w-6 text-yellow-500" />
                          ) : index === 1 ? (
                            <Medal className="h-6 w-6 text-gray-400" />
                          ) : index === 2 ? (
                            <Medal className="h-6 w-6 text-amber-600" />
                          ) : (
                            <span className="text-sm text-muted-foreground">#{index + 1}</span>
                          )}
                        </div>
                        <Avatar className="h-8 w-8">
                          <AvatarImage src={entry.avatar_url || undefined} />
                          <AvatarFallback>{entry.username?.[0]?.toUpperCase()}</AvatarFallback>
                        </Avatar>
                        <div className="flex-1 min-w-0">
                          <p className="text-sm font-medium truncate">{entry.username}</p>
                          <div className="flex items-center gap-2">
                            <Badge variant="secondary" className="text-[10px]">{entry.user_rank}</Badge>
                            <span className="text-xs text-muted-foreground">{entry.xp_earned} XP</span>
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                ) : (
                  <div className="text-center py-8 text-muted-foreground">
                    <Trophy className="h-8 w-8 mx-auto mb-2 opacity-50" />
                    <p className="text-sm">Aucun score</p>
                    <p className="text-xs">Soyez le premier à jouer !</p>
                  </div>
                )}
              </CardContent>
            </Card>
          </div>
        </div>
      </div>
    </div>
  );
}