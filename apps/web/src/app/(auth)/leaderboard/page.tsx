// ============================================================
// PAGE CLASSEMENTS
// ============================================================

import { Suspense } from '../../../../node_modules/@types/react';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { LeaderboardTable } from '@/components/leaderboard/LeaderboardTable';
import { Podium } from '@/components/leaderboard/Podium';
import { Skeleton } from '@/components/ui/skeleton';
import { Trophy, Calendar, BarChart3, Flame } from 'lucide-react';
import { getGlobalLeaderboard, getMonthlyLeaderboard, getWeeklyLeaderboard } from '@/lib/queries/leaderboards';
import { getCurrentUser } from '@/lib/auth/actions';

export default async function LeaderboardPage() {
  const currentUser = await getCurrentUser();
  const yearMonth = `${new Date().getFullYear()}-${String(new Date().getMonth() + 1).padStart(2, '0')}`;

  const [global, monthly, weekly] = await Promise.all([
    getGlobalLeaderboard(1),
    getMonthlyLeaderboard(yearMonth, 1),
    getWeeklyLeaderboard(1),
  ]);

  return (
    <div className="p-4 lg:p-8 pb-24 md:pb-8">
      <div className="max-w-4xl mx-auto space-y-8">
        <h1 className="font-display text-3xl tracking-wider flex items-center gap-3">
          <Trophy className="h-8 w-8 text-brand" />
          CLASSEMENTS
        </h1>

        <Tabs defaultValue="global" className="w-full">
          <TabsList className="w-full justify-start flex-wrap">
            <TabsTrigger value="global" className="gap-2">
              <BarChart3 className="h-4 w-4" /> Global
            </TabsTrigger>
            <TabsTrigger value="monthly" className="gap-2">
              <Calendar className="h-4 w-4" /> Mensuel
            </TabsTrigger>
            <TabsTrigger value="weekly" className="gap-2">
              <Flame className="h-4 w-4" /> Hebdo
            </TabsTrigger>
          </TabsList>

          <TabsContent value="global" className="space-y-6">
            {global.length > 0 && <Podium entries={global.slice(0, 3)} />}
            <LeaderboardTable 
              entries={global} 
              currentUserId={currentUser?.id} 
              type="global"
            />
          </TabsContent>

          <TabsContent value="monthly" className="space-y-6">
            <div className="flex items-center justify-between">
              <h2 className="font-display text-lg tracking-wider">{yearMonth}</h2>
              <div className="text-sm text-muted-foreground">
                Réinitialisation le 1er du mois
              </div>
            </div>
            {monthly.length > 0 && <Podium entries={monthly.slice(0, 3)} />}
            <LeaderboardTable 
              entries={monthly} 
              currentUserId={currentUser?.id}
              type="monthly"
            />
          </TabsContent>

          <TabsContent value="weekly" className="space-y-6">
            {weekly.length > 0 && <Podium entries={weekly.slice(0, 3)} />}
            <LeaderboardTable 
              entries={weekly} 
              currentUserId={currentUser?.id}
              type="weekly"
            />
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
}

