'use client';

import { LeaderboardEntry } from '@/types';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { RankBadge } from '@/components/ui/RankBadge';
import { cn, getInitials } from '@/lib/utils';
import Link from '../../../node_modules/next/link';

interface LeaderboardTableProps {
  entries: LeaderboardEntry[];
  currentUserId?: string;
  type: string;
}

export function LeaderboardTable({ entries, currentUserId, type }: LeaderboardTableProps) {
  return (
    <div className="border border-dark-border rounded-lg overflow-hidden">
      <div className="grid grid-cols-12 gap-2 p-3 text-xs text-muted-foreground uppercase tracking-wider bg-dark-surface/50">
        <div className="col-span-1">Rang</div>
        <div className="col-span-5">Joueur</div>
        <div className="col-span-2 text-center">Rang</div>
        <div className="col-span-2 text-center">Score</div>
        <div className="col-span-2 text-center">Quiz</div>
      </div>

      <div className="divide-y divide-dark-border">
        {entries.map((entry) => (
          <div
            key={entry.user_id}
            className={cn(
              'grid grid-cols-12 gap-2 p-3 items-center transition-colors hover:bg-dark-surface/30',
              entry.user_id === currentUserId && 'bg-brand/5 border-l-2 border-l-brand'
            )}
          >
            <div className="col-span-1">
              {entry.rank <= 3 ? (
                <span className={cn(
                  'font-display text-lg',
                  entry.rank === 1 ? 'text-yellow-400' :
                  entry.rank === 2 ? 'text-gray-300' :
                  'text-amber-600'
                )}>
                  {entry.rank === 1 ? '🥇' : entry.rank === 2 ? '🥈' : '🥉'}
                </span>
              ) : (
                <span className="text-muted-foreground">#{entry.rank}</span>
              )}
            </div>

            <div className="col-span-5 flex items-center gap-3 min-w-0">
              <Link href={`/profile/${entry.username}`}>
                <Avatar className="h-8 w-8 shrink-0">
                  <AvatarImage src={entry.avatar_url || undefined} />
                  <AvatarFallback className="text-xs bg-dark-surface">
                    {getInitials(entry.username)}
                  </AvatarFallback>
                </Avatar>
              </Link>
              <Link href={`/profile/${entry.username}`} className="truncate hover:text-brand transition-colors">
                {entry.username}
              </Link>
            </div>

            <div className="col-span-2 text-center">
              <RankBadge rank={entry.user_rank} size="sm" />
            </div>

            <div className="col-span-2 text-center font-display text-brand">
              {entry.score?.toLocaleString()}
            </div>

            <div className="col-span-2 text-center text-sm text-muted-foreground">
              {entry.quiz_count || '-'}
            </div>
          </div>
        ))}

        {entries.length === 0 && (
          <div className="p-8 text-center text-muted-foreground">
            Aucune donnée disponible
          </div>
        )}
      </div>
    </div>
  );
}

