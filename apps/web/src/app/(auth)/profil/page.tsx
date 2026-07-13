// ============================================================
// PAGE PROFIL PERSONNEL - Historique, Badges, Quiz créés
// ============================================================

import { redirect } from 'next/navigation';
import Link from 'next/link';
import { createClient } from '@/lib/supabase/server';
import { getCurrentUser } from '@/lib/auth/actions';
import { Card, CardContent } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Button } from '@/components/ui/button';
import { BadgeGrid } from '@/components/profile/BadgeGrid';
import { ProfileHeader } from '@/components/profile/ProfileHeader';
import { getUserStats, getUserBadges, getUserQuizzes, getUserCollections } from '@/lib/queries/social';
import { cn } from '@/lib/utils';
import {
  BarChart3,
  Award,
  BookOpen,
  History,
  Edit2,
  Play,
  Trophy,
  Clock,
  CheckCircle2,
  Gamepad2,
  Target,
  ChevronRight,
} from 'lucide-react';
import { QuizActions } from '@/components/quiz/QuizActions';

export const metadata = {
  title: 'Mon Profil | Otaku Quiz Africa',
};

export default async function ProfilPage() {
  const user = await getCurrentUser();
  if (!user) redirect('/login');

  const supabase = createClient();

  // Récupérer le profil utilisateur
  const { data: profile } = await supabase
    .from('user_profiles')
    .select('*')
    .eq('id', user.id)
    .single();

  if (!profile) redirect('/dashboard');

  // Récupérer les données en parallèle
  const [stats, badges, quizzes, collections] = await Promise.all([
    getUserStats(user.id),
    getUserBadges(user.id),
    getUserQuizzes(user.id),
    getUserCollections(user.id),
  ]);

  // Récupérer l'historique des parties jouées (completed_at non null)
  const { data: gameHistory, error: historyError } = await supabase
    .rpc('get_user_game_history', { p_user_id: user.id });

  console.log('Game history count:', gameHistory?.length, 'error:', historyError);
  console.log('User ID:', user.id);

  return (
    <div className="p-4 lg:p-8 pb-24 md:pb-8">
      <div className="max-w-4xl mx-auto space-y-8">
        {/* Header */}
        <ProfileHeader profile={profile} stats={stats} isOwnProfile={true} />

        <Tabs defaultValue="history" className="w-full">
          <TabsList className="w-full justify-start overflow-x-auto">
            <TabsTrigger value="history" className="gap-2">
              <History className="h-4 w-4" /> Historique ({gameHistory?.length || 0})
            </TabsTrigger>
            <TabsTrigger value="badges" className="gap-2">
              <Award className="h-4 w-4" /> Badges ({badges.length})
            </TabsTrigger>
            <TabsTrigger value="my-quizzes" className="gap-2">
              <BookOpen className="h-4 w-4" /> Mes Quiz ({quizzes.length})
            </TabsTrigger>
            <TabsTrigger value="stats" className="gap-2">
              <BarChart3 className="h-4 w-4" /> Stats
            </TabsTrigger>
          </TabsList>

          {/* ===== HISTORIQUE DES PARTIES ===== */}
          <TabsContent value="history" className="space-y-4">
            {gameHistory && gameHistory.length > 0 ? (
              <div className="space-y-3">
                {gameHistory.map((session: any) => (
                  <Link key={session.id} href={`/quiz/${session.quiz_id}/results/${session.id}`}>
                    <Card className="border-dark-border bg-dark-card hover:border-brand/30 transition-all cursor-pointer">
                      <CardContent className="p-4">
                        <div className="flex items-start gap-4">
                          {/* Thumbnail */}
                          <div className="w-16 h-16 rounded-lg bg-dark-surface overflow-hidden shrink-0">
                            {session.quiz_thumbnail ? (
                              <img
                                src={session.quiz_thumbnail}
                                alt={session.quiz_title}
                                className="w-full h-full object-cover"
                              />
                            ) : (
                              <div className="w-full h-full flex items-center justify-center">
                                <Gamepad2 className="h-6 w-6 text-muted-foreground" />
                              </div>
                            )}
                          </div>

                          {/* Info */}
                          <div className="flex-1 min-w-0">
                            <h3 className="font-medium truncate">{session.quiz_title || 'Quiz supprimé'}</h3>
                            <div className="flex flex-wrap items-center gap-2 mt-1 text-xs text-muted-foreground">
                              <span className="flex items-center gap-1">
                                <Clock className="h-3 w-3" />
                                {new Date(session.completed_at).toLocaleDateString('fr-FR', {
                                  day: 'numeric',
                                  month: 'short',
                                  year: 'numeric',
                                  hour: '2-digit',
                                  minute: '2-digit',
                                })}
                              </span>
                              {session.quiz_series && (
                                <span className="px-2 py-0.5 rounded-full bg-dark-surface text-xs">
                                  {session.quiz_series}
                                </span>
                              )}
                            </div>
                          </div>

                          {/* Score */}
                          <div className="text-right shrink-0">
                            <div className="flex items-center gap-1 justify-end">
                              {session.is_perfect && <Trophy className="h-4 w-4 text-yellow-500" />}
                              <span className="text-lg font-bold text-brand">{session.score}</span>
                            </div>
                            <div className="text-xs text-muted-foreground">
                              {session.correct_count}/{session.total_questions}
                            </div>
                            <div className="text-xs text-muted-foreground">
                              {session.accuracy_rate}%
                            </div>
                          </div>

                          <ChevronRight className="h-5 w-5 text-muted-foreground shrink-0" />
                        </div>
                      </CardContent>
                    </Card>
                  </Link>
                ))}
              </div>
            ) : (
              <Card className="border-dark-border bg-dark-card">
                <CardContent className="p-8 text-center">
                  <Gamepad2 className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                  <p className="text-muted-foreground mb-4">Tu n&apos;as pas encore joué de quiz</p>
                  <Link href="/explore">
                    <Button>
                      <Play className="h-4 w-4 mr-2" />
                      Explorer les quiz
                    </Button>
                  </Link>
                </CardContent>
              </Card>
            )}
          </TabsContent>

          {/* ===== BADGES ===== */}
          <TabsContent value="badges">
            <BadgeGrid badges={badges} />
          </TabsContent>

          {/* ===== MES QUIZ CRÉÉS ===== */}
          <TabsContent value="my-quizzes" className="space-y-4">
            {quizzes.length > 0 ? (
              <div className="grid sm:grid-cols-2 gap-4">
                {quizzes.map((quiz: any) => (
                  <Card key={quiz.id} className="border-dark-border bg-dark-card">
                    <CardContent className="p-4">
                      <div className="flex items-start gap-3">
                        {/* Thumbnail */}
                        <div className="w-14 h-14 rounded-lg bg-dark-surface overflow-hidden shrink-0">
                          {quiz.thumbnail_url ? (
                            <img
                              src={quiz.thumbnail_url}
                              alt={quiz.title}
                              className="w-full h-full object-cover"
                            />
                          ) : (
                            <div className="w-full h-full flex items-center justify-center">
                              <BookOpen className="h-5 w-5 text-muted-foreground" />
                            </div>
                          )}
                        </div>

                        {/* Info */}
                        <div className="flex-1 min-w-0">
                          <h3 className="font-medium truncate">{quiz.title}</h3>
                          <div className="flex items-center gap-2 mt-1 text-xs text-muted-foreground">
                            <span>{quiz.question_count} questions</span>
                            <span>•</span>
                            <span>{quiz.play_count} joué(s)</span>
                          </div>
                          <div className="flex items-center gap-2 mt-2">
                            <span className={`text-xs px-2 py-0.5 rounded-full ${
                              quiz.quiz_type === 'official' 
                                ? 'bg-brand/10 text-brand' 
                                : 'bg-dark-surface text-muted-foreground'
                            }`}>
                              {quiz.quiz_type}
                            </span>
                            {quiz.series && (
                              <span className="text-xs text-muted-foreground">{quiz.series}</span>
                            )}
                          </div>
                        </div>
                      </div>

                      {/* Actions */}
                      <div className="flex items-center justify-between mt-4">
                        <QuizActions quizId={quiz.id} />
                        <Link href={`/quiz/${quiz.id}/play`}>
                          <Button variant="default" size="sm" className="gap-2">
                            <Play className="h-3 w-3" />
                            Jouer
                          </Button>
                        </Link>
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>
            ) : (
              <Card className="border-dark-border bg-dark-card">
                <CardContent className="p-8 text-center">
                  <BookOpen className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                  <p className="text-muted-foreground mb-4">Tu n&apos;as pas encore créé de quiz</p>
                  <Link href="/quiz/create">
                    <Button>
                      <Edit2 className="h-4 w-4 mr-2" />
                      Créer un quiz
                    </Button>
                  </Link>
                </CardContent>
              </Card>
            )}
          </TabsContent>

          {/* ===== STATISTIQUES ===== */}
          <TabsContent value="stats" className="space-y-6">
            <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
              <StatBox 
                icon={<Gamepad2 className="h-5 w-5" />} 
                label="Quiz joués" 
                value={stats?.quizzes_played || 0} 
                color="brand" 
              />
              <StatBox 
                icon={<BookOpen className="h-5 w-5" />} 
                label="Quiz créés" 
                value={stats?.quizzes_created || 0} 
                color="accent" 
              />
              <StatBox 
                icon={<Target className="h-5 w-5" />} 
                label="Précision" 
                value={`${stats?.accuracy_rate || 0}%`} 
                color="green" 
              />
              <StatBox 
                icon={<Trophy className="h-5 w-5" />} 
                label="Meilleur score" 
                value={stats?.best_score_ever || 0} 
                color="purple" 
              />
            </div>

            {/* Résumé rapide */}
            <Card className="border-dark-border bg-dark-card">
              <CardContent className="p-6">
                <h3 className="font-display text-lg tracking-wider mb-4">RÉSUMÉ</h3>
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div className="flex items-center gap-2">
                    <CheckCircle2 className="h-4 w-4 text-green-400" />
                    <span className="text-muted-foreground">Réponses correctes :</span>
                    <span className="font-medium">{stats?.total_correct_answers || 0}</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <History className="h-4 w-4 text-blue-400" />
                    <span className="text-muted-foreground">Total réponses :</span>
                    <span className="font-medium">{stats?.total_answers || 0}</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <Award className="h-4 w-4 text-yellow-400" />
                    <span className="text-muted-foreground">Badges :</span>
                    <span className="font-medium">{badges.length}</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <Trophy className="h-4 w-4 text-purple-400" />
                    <span className="text-muted-foreground">Collections :</span>
                    <span className="font-medium">{collections.length}</span>
                  </div>
                </div>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
}

// Composant StatBox
function StatBox({ icon, label, value, color }: { icon: React.ReactNode; label: string; value: string | number; color: string }) {
  const colorMap: Record<string, string> = {
    brand: 'text-brand bg-brand/10',
    accent: 'text-accent bg-accent/10',
    green: 'text-green-400 bg-green-400/10',
    purple: 'text-purple-400 bg-purple-400/10',
  };

  return (
    <Card className="border-dark-border bg-dark-card">
      <CardContent className="p-4 text-center">
        <div className={`h-10 w-10 rounded-lg flex items-center justify-center mx-auto mb-2 ${colorMap[color]}`}>
          {icon}
        </div>
        <div className="font-display text-2xl">{value}</div>
        <div className="text-xs text-muted-foreground">{label}</div>
      </CardContent>
    </Card>
  );
}