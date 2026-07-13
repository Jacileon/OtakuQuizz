'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Badge } from '@/components/ui/badge';
import { Trophy, Medal, Crown, Award } from 'lucide-react';
import { OfficialLeaderboardEntry } from '@/types';
import { cn } from '@/lib/utils';

interface OfficialLeaderboardProps {
  entries: OfficialLeaderboardEntry[];
  rewards?: { title: string; rank_from: number; rank_to: number }[];
  showPodium?: boolean;
}

export function OfficialLeaderboard({ entries, rewards = [], showPodium = true }: OfficialLeaderboardProps) {
  const top3 = entries.slice(0, 3);
  const rest = entries.slice(3);

  const getRewardForRank = (rank: number) => {
    return rewards.find(r => rank >= r.rank_from && rank <= r.rank_to);
  };

  return (
    <div className="space-y-6">
      {showPodium && top3.length > 0 && (
        <div className="grid grid-cols-3 gap-4 items-end">
          {top3.length > 1 && (
            <PodiumPlace
              entry={top3[1]}
              place={2}
              reward={getRewardForRank(2)}
              className="order-1"
            />
          )}
          {top3.length > 0 && (
            <PodiumPlace
              entry={top3[0]}
              place={1}
              reward={getRewardForRank(1)}
              className="order-2"
            />
          )}
          {top3.length > 2 && (
            <PodiumPlace
              entry={top3[2]}
              place={3}
              reward={getRewardForRank(3)}
              className="order-3"
            />
          )}
        </div>
      )}

      {rest.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-lg">
              <Trophy className="h-5 w-5 text-yellow-500" />
              Classement complet
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              {rest.map((entry, index) => (
                <LeaderboardRow
                  key={entry.user_id}
                  entry={entry}
                  rank={index + 4}
                  reward={getRewardForRank(index + 4)}
                />
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}

function PodiumPlace({
  entry,
  place,
  reward,
  className,
}: {
  entry: OfficialLeaderboardEntry;
  place: number;
  reward?: { title: string };
  className?: string;
}) {
  const heights = {
    1: 'h-32',
    2: 'h-24',
    3: 'h-20',
  };

  const colors = {
    1: 'from-yellow-500 to-amber-500',
    2: 'from-gray-400 to-gray-500',
    3: 'from-amber-600 to-amber-700',
  };

  const icons = {
    1: <Crown className="h-8 w-8 text-yellow-300" />,
    2: <Medal className="h-7 w-7 text-gray-300" />,
    3: <Medal className="h-6 w-6 text-amber-400" />,
  };

  return (
    <div className={cn('flex flex-col items-center', className)}>
      <div className="relative mb-3">
        <Avatar className="h-16 w-16 border-4 border-yellow-500/50">
          <AvatarImage src={entry.user?.avatar_url || undefined} />
          <AvatarFallback>{entry.user?.username?.[0]?.toUpperCase()}</AvatarFallback>
        </Avatar>
        <div className="absolute -top-2 -right-2">
          {icons[place as keyof typeof icons]}
        </div>
      </div>

      <p className="font-medium text-sm text-center truncate max-w-full">{entry.user?.username}</p>
      <p className="text-2xl font-bold text-brand">{entry.score}</p>
      <p className="text-xs text-muted-foreground">points</p>

      <div className={cn(
        'w-full mt-2 rounded-t-lg bg-gradient-to-b flex items-center justify-center',
        colors[place as keyof typeof colors],
        heights[place as keyof typeof heights],
      )}>
        <span className="text-4xl font-bold text-white/80">#{place}</span>
      </div>

      {reward && (
        <div className="mt-2 p-2 rounded-lg bg-yellow-500/10 border border-yellow-500/30 text-center w-full">
          <Award className="h-4 w-4 text-yellow-500 mx-auto mb-1" />
          <p className="text-xs font-medium text-yellow-500">{reward.title}</p>
        </div>
      )}
    </div>
  );
}

function LeaderboardRow({
  entry,
  rank,
  reward,
}: {
  entry: OfficialLeaderboardEntry;
  rank: number;
  reward?: { title: string };
}) {
  return (
    <div className="flex items-center gap-3 p-3 rounded-lg hover:bg-accent/50 transition-colors">
      <div className="w-8 h-8 flex items-center justify-center">
        <span className="text-sm font-bold text-muted-foreground">#{rank}</span>
      </div>
      <Avatar className="h-10 w-10">
        <AvatarImage src={entry.user?.avatar_url || undefined} />
        <AvatarFallback>{entry.user?.username?.[0]?.toUpperCase()}</AvatarFallback>
      </Avatar>
      <div className="flex-1 min-w-0">
        <p className="font-medium truncate">{entry.user?.username}</p>
        <div className="flex items-center gap-2">
          <Badge variant="secondary" className="text-xs">{entry.user?.rank || 'F'}</Badge>
          {reward && (
            <span className="text-xs text-yellow-500 flex items-center gap-1">
              <Award className="h-3 w-3" />
              {reward.title}
            </span>
          )}
        </div>
      </div>
      <div className="text-right">
        <p className="font-bold text-lg">{entry.score}</p>
        <p className="text-xs text-muted-foreground">{entry.accuracy_rate}% précision</p>
      </div>
    </div>
  );
}