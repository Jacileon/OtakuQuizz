import { UserBadge } from '@/types';
import { Card, CardContent } from '@/components/ui/card';
import { Award, Lock } from 'lucide-react';
import { cn } from '@/lib/utils';

interface BadgeGridProps {
  badges: UserBadge[];
}

export function BadgeGrid({ badges }: BadgeGridProps) {
  return (
    <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-4">
      {badges.map((ub) => (
        <Card key={ub.id} className="border-dark-border bg-dark-card hover:border-brand/30 transition-colors">
          <CardContent className="p-4 text-center space-y-2">
            <div className={cn(
              'h-12 w-12 rounded-full flex items-center justify-center mx-auto',
              ub.badge?.is_rare ? 'bg-yellow-500/10' : 'bg-brand/10'
            )}>
              <Award className={cn(
                'h-6 w-6',
                ub.badge?.is_rare ? 'text-yellow-400' : 'text-brand'
              )} />
            </div>
            <div className="font-medium text-sm">{ub.badge?.name}</div>
            <div className="text-xs text-muted-foreground">{ub.badge?.description}</div>
            {ub.badge?.is_rare && (
              <div className="text-xs text-yellow-400 font-medium">★ RARE</div>
            )}
          </CardContent>
        </Card>
      ))}
      {badges.length === 0 && (
        <div className="col-span-full text-center py-8">
          <Lock className="h-8 w-8 mx-auto mb-2 text-muted-foreground" />
          <p className="text-muted-foreground">Aucun badge obtenu encore</p>
        </div>
      )}
    </div>
  );
}

