import Link from '../../../../../node_modules/next/link';
import { redirect } from '../../../../../node_modules/next/navigation';
import { getCurrentUser } from '@/lib/auth/actions';
import {
  getUserProfile,
  getUserStats,
  getUserBadges,
  getUserQuizzes,
  getUserCollections,
} from '@/lib/queries/social';
import { ProfileHeader } from '@/components/profile/ProfileHeader';
import { BadgeGrid } from '@/components/profile/BadgeGrid';
import { Card, CardContent } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { RankBadge } from '@/components/ui/RankBadge';
import { Progress } from '@/components/ui/progress';
import { Trophy, Gamepad2, Target, Award, BookOpen, BarChart3 } from 'lucide-react';
import { cn } from '@/lib/utils';

export default async function ProfilePage({ params }: { params: Promise<{ username: string }> }) {
  const { username } = await params;
  const currentUser = await getCurrentUser();
  const profile = await getUserProfile(username);

  if (!profile) redirect('/dashboard');

  const isOwnProfile = currentUser?.id === profile.id;

  const [stats, badges, quizzes, collections] = await Promise.all([
    getUserStats(profile.id),
    getUserBadges(profile.id),
    getUserQuizzes(profile.id),
    getUserCollections(profile.id),
  ]);

  return (
    <div className="p-4 lg:p-8 pb-24 md:pb-8">
      <div className="max-w-4xl mx-auto space-y-8">
        <ProfileHeader profile={profile} stats={stats} isOwnProfile={isOwnProfile} />

        <Tabs defaultValue="stats" className="w-full">
          <TabsList className="w-full justify-start">
            <TabsTrigger value="stats" className="gap-2">
              <BarChart3 className="h-4 w-4" /> Statistiques
            </TabsTrigger>
            <TabsTrigger value="badges" className="gap-2">
              <Award className="h-4 w-4" /> Badges ({badges.length})
            </TabsTrigger>
            <TabsTrigger value="quizzes" className="gap-2">
              <BookOpen className="h-4 w-4" /> Quiz ({quizzes.length})
            </TabsTrigger>
            <TabsTrigger value="collections" className="gap-2">
              <Trophy className="h-4 w-4" /> Collections
            </TabsTrigger>
          </TabsList>

          <TabsContent value="stats" className="space-y-6">
            <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
              <StatBox icon={<Gamepad2 className="h-5 w-5" />} label="Quiz joués" value={stats?.quizzes_played || 0} color="brand" />
              <StatBox icon={<BookOpen className="h-5 w-5" />} label="Quiz créés" value={stats?.quizzes_created || 0} color="accent" />
              <StatBox icon={<Target className="h-5 w-5" />} label="Précision" value={`${stats?.accuracy_rate || 0}%`} color="green" />
              <StatBox icon={<Trophy className="h-5 w-5" />} label="Meilleur score" value={stats?.best_score_ever || 0} color="purple" />
            </div>

            <Card className="border-dark-border bg-dark-card">
              <CardContent className="p-6">
                <h3 className="font-display text-lg tracking-wider mb-4">PROGRESSION</h3>
                <div className="space-y-4">
                  <div className="flex justify-between text-sm">
                    <span>XP Total</span>
                    <span className="font-medium">{profile.xp.toLocaleString()}</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span>Niveau</span>
                    <span className="font-medium">{profile.level}</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span>Rang</span>
                    <RankBadge rank={profile.rank} size="sm" />
                  </div>
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="badges">
            <BadgeGrid badges={badges} />
          </TabsContent>

          <TabsContent value="quizzes">
            <div className="grid sm:grid-cols-2 gap-4">
              {quizzes.map((quiz: any) => (
                <Link key={quiz.id} href={`/quiz/${quiz.id}/play`}>
                  <Card className="border-dark-border bg-dark-card hover:border-brand/30 transition-colors">
                    <CardContent className="p-4">
                      <h3 className="font-medium truncate">{quiz.title}</h3>
                      <div className="flex items-center gap-3 text-xs text-muted-foreground mt-2">
                        <span>{quiz.question_count} questions</span>
                        <span>{quiz.play_count} plays</span>
                      </div>
                    </CardContent>
                  </Card>
                </Link>
              ))}
              {quizzes.length === 0 && (
                <p className="text-muted-foreground col-span-2 text-center py-8">Aucun quiz créé</p>
              )}
            </div>
          </TabsContent>

          <TabsContent value="collections">
            <div className="space-y-4">
              {collections.map((col: any) => (
                <Card key={col.series} className="border-dark-border bg-dark-card">
                  <CardContent className="p-4 space-y-3">
                    <div className="flex items-center justify-between">
                      <h3 className="font-medium">{col.series}</h3>
                      <span className="text-sm text-muted-foreground">
                        {col.completed_quizzes}/{col.total_quizzes}
                      </span>
                    </div>
                    <Progress value={col.progress_percent} className="h-2" />
                    <div className="flex justify-between text-xs text-muted-foreground">
                      <span>{col.progress_percent}% complété</span>
                      {col.best_score && <span>Meilleur score: {col.best_score}</span>}
                    </div>
                  </CardContent>
                </Card>
              ))}
              {collections.length === 0 && (
                <p className="text-muted-foreground text-center py-8">Aucune collection commencée</p>
              )}
            </div>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
}

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
        <div className={cn('h-10 w-10 rounded-lg flex items-center justify-center mx-auto mb-2', colorMap[color])}>
          {icon}
        </div>
        <div className="font-display text-2xl">{value}</div>
        <div className="text-xs text-muted-foreground">{label}</div>
      </CardContent>
    </Card>
  );
}
