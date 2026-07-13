// ============================================================
// PAGE BADGES
// ============================================================

import { getCurrentUser } from '@/lib/auth/actions';
import { getUserBadges } from '@/lib/queries/social';
import { createClient } from '@/lib/supabase/server';
import { Card, CardContent } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Award, Lock, Star } from 'lucide-react';
import { redirect } from '../../../../node_modules/next/navigation';
import { cn } from '@/lib/utils';

export default async function BadgesPage() {
  const user = await getCurrentUser();
  if (!user) redirect('/login');

  const supabase = createClient();

  const [userBadges, allBadges] = await Promise.all([
    getUserBadges(user.id),
    supabase.from('badges').select('*').order('condition_value', { ascending: true }),
  ]);

  const earnedBadgeIds = new Set(userBadges.map((ub: any) => ub.badge_id));
  const earned = allBadges.data?.filter((b: any) => earnedBadgeIds.has(b.id)) || [];
  const locked = allBadges.data?.filter((b: any) => !earnedBadgeIds.has(b.id)) || [];
  const rare = earned.filter((b: any) => b.is_rare);

  return (
    <div className="p-4 lg:p-8 pb-24 md:pb-8">
      <div className="max-w-4xl mx-auto space-y-6">
        <h1 className="font-display text-3xl tracking-wider flex items-center gap-3">
          <Award className="h-8 w-8 text-brand" />
          BADGES
        </h1>

        <Tabs defaultValue="earned" className="w-full">
          <TabsList className="w-full justify-start">
            <TabsTrigger value="earned" className="gap-2">
              <Award className="h-4 w-4" /> Obtenus ({earned.length})
            </TabsTrigger>
            <TabsTrigger value="locked" className="gap-2">
              <Lock className="h-4 w-4" /> À débloquer ({locked.length})
            </TabsTrigger>
            <TabsTrigger value="rare" className="gap-2">
              <Star className="h-4 w-4" /> Rares ({rare.length})
            </TabsTrigger>
          </TabsList>

          <TabsContent value="earned" className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-4">
            {earned.map((badge: any) => (
              <BadgeCard key={badge.id} badge={badge} earned />
            ))}
            {earned.length === 0 && <p className="col-span-full text-center text-muted-foreground py-8">Aucun badge obtenu</p>}
          </TabsContent>

          <TabsContent value="locked" className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-4">
            {locked.map((badge: any) => (
              <BadgeCard key={badge.id} badge={badge} earned={false} />
            ))}
          </TabsContent>

          <TabsContent value="rare" className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-4">
            {rare.map((badge: any) => (
              <BadgeCard key={badge.id} badge={badge} earned />
            ))}
            {rare.length === 0 && <p className="col-span-full text-center text-muted-foreground py-8">Aucun badge rare</p>}
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
}

function BadgeCard({ badge, earned }: { badge: any; earned: boolean }) {
  return (
    <Card className={cn(
      'border-dark-border transition-all',
      earned ? 'bg-dark-card hover:border-brand/30' : 'bg-dark-surface/50 opacity-60'
    )}>
      <CardContent className="p-4 text-center space-y-2">
        <div className={cn(
          'h-12 w-12 rounded-full flex items-center justify-center mx-auto',
          earned ? (badge.is_rare ? 'bg-yellow-500/10' : 'bg-brand/10') : 'bg-dark-surface'
        )}>
          <Award className={cn(
            'h-6 w-6',
            earned ? (badge.is_rare ? 'text-yellow-400' : 'text-brand') : 'text-muted-foreground'
          )} />
        </div>
        <div className="font-medium text-sm">{badge.name}</div>
        <div className="text-xs text-muted-foreground">{badge.description}</div>
        {badge.is_rare && earned && <div className="text-xs text-yellow-400">★ RARE</div>}
      </CardContent>
    </Card>
  );
}

