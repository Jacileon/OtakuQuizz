// ============================================================
// DASHBOARD UTILISATEUR
// ============================================================

import Link from '../../../../node_modules/next/link';
import { redirect } from '../../../../node_modules/next/navigation';
import { getCurrentProfile } from '@/lib/auth/actions';
import {
  getDashboardStats,
  getActiveOfficialQuizzes,
  getUpcomingOfficialQuizzes,
  getRecommendedQuizzes,
  getRecentActivity,
  getRecentBadges,
} from '@/lib/queries/dashboard';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { RankBadge } from '@/components/ui/RankBadge';
import { XPBar } from '@/components/ui/XPBar';
import { StatCard } from '@/components/dashboard/StatCard';
import { OfficialEventCard } from '@/components/dashboard/OfficialEventCard';
import { QuizCard } from '@/components/dashboard/QuizCard';
import { Gamepad2, Trophy, Target, Clock, Zap, ChevronRight, Award } from 'lucide-react';
import { formatTimeAgo } from '@/lib/utils';

export default async function DashboardPage() {
  const profile = await getCurrentProfile();
  if (!profile) redirect('/login');

  const stats = await getDashboardStats(profile.id);
  const activeEvents = await getActiveOfficialQuizzes();
  const upcomingEvents = await getUpcomingOfficialQuizzes();
  const recommended = await getRecommendedQuizzes(profile.id, profile.favorite_anime || 'Naruto');
  const recentActivity = await getRecentActivity(profile.id, 5);
  const recentBadges = await getRecentBadges(profile.id, 3);

  return (
    <div className="p-4 lg:p-8 space-y-8 pb-24 md:pb-8">
      {/* HEADER */}
      <div className="flex flex-col md:flex-row items-start md:items-center justify-between gap-4">
        <div>
          <h1 className="font-display text-3xl tracking-wider">
            Bonjour {profile.username} 👋
          </h1>
          <div className="flex items-center gap-3 mt-2">
            <RankBadge rank={profile.rank} size="md" showLabel />
            <span className="text-sm text-muted-foreground">
              Niveau {profile.level}
            </span>
          </div>
        </div>
        <Link href="/explore">
          <Button className="gap-2">
            <Gamepad2 className="h-4 w-4" />
            Jouer maintenant
          </Button>
        </Link>
      </div>

      {/* XP BAR */}
      <XPBar currentXP={profile.xp} rank={profile.rank} className="max-w-xl" />

      {/* STATS RAPIDES */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          icon={<Gamepad2 className="h-5 w-5" />}
          label="Quiz joués"
          value={(stats.totalPlayed || 0).toString()}
          color="brand"
        />
        <StatCard
          icon={<Trophy className="h-5 w-5" />}
          label="Meilleur score"
          value={stats.bestScore.toString()}
          color="accent"
        />
        <StatCard
          icon={<Target className="h-5 w-5" />}
          label="Précision"
          value={`${stats.accuracy}%`}
          color="green"
        />
        <StatCard
          icon={<Award className="h-5 w-5" />}
          label="Classement mensuel"
          value={stats.monthlyRank ? `#${stats.monthlyRank}` : '-'}
          color="purple"
        />
      </div>

      {/* ÉVÉNEMENTS OFFICIELS */}
      {(activeEvents.length > 0 || upcomingEvents.length > 0) && (
        <div>
          <h2 className="font-display text-xl tracking-wider mb-4 flex items-center gap-2">
            <Zap className="h-5 w-5 text-brand" />
            ÉVÉNEMENTS OFFICIELS
          </h2>
          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-4">
            {activeEvents.map((event) => (
              <OfficialEventCard key={event.id} event={event} status="live" />
            ))}
            {upcomingEvents.map((event) => (
              <OfficialEventCard key={event.id} event={event} status="upcoming" />
            ))}
          </div>
        </div>
      )}

      {/* QUIZ RECOMMANDÉS */}
      <div>
        <div className="flex items-center justify-between mb-4">
          <h2 className="font-display text-xl tracking-wider">
            RECOMMANDÉS POUR TOI
          </h2>
          <Link href="/explore">
            <Button variant="ghost" size="sm" className="gap-1">
              Voir plus <ChevronRight className="h-4 w-4" />
            </Button>
          </Link>
        </div>
        <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {recommended.map((quiz) => (
            <QuizCard key={quiz.id} quiz={quiz} />
          ))}
        </div>
      </div>

      {/* ACTIVITÉ RÉCENTE */}
      <div className="grid lg:grid-cols-2 gap-8">
        <div>
          <h2 className="font-display text-xl tracking-wider mb-4 flex items-center gap-2">
            <Clock className="h-5 w-5 text-accent" />
            ACTIVITÉ RÉCENTE
          </h2>
          <div className="space-y-3">
            {recentActivity.length > 0 ? (
              recentActivity.map((session) => (
                <Card key={session.id} className="border-dark-border bg-dark-card/50">
                  <CardContent className="p-4 flex items-center justify-between">
                    <div>
                      <div className="font-medium">{(session as any).quiz?.title || 'Quiz'}</div>
                      <div className="text-xs text-muted-foreground">
                        {formatTimeAgo(session.completed_at!)} • {(session as any).quiz?.series}
                      </div>
                    </div>
                    <div className="text-right">
                      <div className="font-display text-lg text-brand">{session.score}</div>
                      <div className="text-xs text-muted-foreground">{session.accuracy_rate}%</div>
                    </div>
                  </CardContent>
                </Card>
              ))
            ) : (
              <p className="text-muted-foreground text-sm">Aucune activité récente</p>
            )}
          </div>
        </div>

        {/* BADGES RÉCENTS */}
        <div>
          <h2 className="font-display text-xl tracking-wider mb-4 flex items-center gap-2">
            <Award className="h-5 w-5 text-yellow-400" />
            BADGES RÉCENTS
          </h2>
          <div className="space-y-3">
            {recentBadges.length > 0 ? (
              recentBadges.map((ub) => (
                <Card key={ub.id} className="border-dark-border bg-dark-card/50">
                  <CardContent className="p-4 flex items-center gap-4">
                    <div className="h-10 w-10 rounded-full bg-brand/10 flex items-center justify-center">
                      <Award className="h-5 w-5 text-brand" />
                    </div>
                    <div>
                      <div className="font-medium">{ub.badge?.name}</div>
                      <div className="text-xs text-muted-foreground">{ub.badge?.description}</div>
                    </div>
                  </CardContent>
                </Card>
              ))
            ) : (
              <p className="text-muted-foreground text-sm">Aucun badge obtenu encore</p>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

