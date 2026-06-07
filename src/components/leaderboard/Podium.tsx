'use client';

import { LeaderboardEntry } from '@/types';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { RankBadge } from '@/components/ui/RankBadge';
import { Crown, Medal } from 'lucide-react';
import { getInitials } from '@/lib/utils';
import Link from '../../../node_modules/next/link';

interface PodiumProps {
  entries: LeaderboardEntry[];
}

export function Podium({ entries }: PodiumProps) {
  const [first, second, third] = [entries[0], entries[1], entries[2]];

  return (
    <div className="flex items-end justify-center gap-4 py-8">
      {/* 2nd */}
      {second && (
        <div className="flex flex-col items-center gap-2 animate-slide-up" style={{ animationDelay: '100ms' }}>
          <div className="relative">
            <Avatar className="h-16 w-16 border-2 border-gray-300">
              <AvatarImage src={second.avatar_url || undefined} />
              <AvatarFallback className="text-lg bg-dark-surface">{getInitials(second.username)}</AvatarFallback>
            </Avatar>
            <div className="absolute -top-2 -right-2 h-6 w-6 rounded-full bg-gray-300 flex items-center justify-center text-dark text-xs font-bold">
              2
            </div>
          </div>
          <Link href={`/profile/${second.username}`} className="font-medium text-sm hover:text-brand transition-colors">
            {second.username}
          </Link>
          <RankBadge rank={second.user_rank} size="sm" />
          <div className="font-display text-xl text-gray-300">{second.score?.toLocaleString()}</div>
          <div className="w-24 h-24 bg-dark-card rounded-t-lg border border-dark-border flex items-center justify-center">
            <Medal className="h-8 w-8 text-gray-300" />
          </div>
        </div>
      )}

      {/* 1st */}
      {first && (
        <div className="flex flex-col items-center gap-2 animate-slide-up">
          <div className="relative">
            <div className="absolute -top-6 left-1/2 -translate-x-1/2">
              <Crown className="h-6 w-6 text-yellow-400" />
            </div>
            <Avatar className="h-20 w-20 border-2 border-yellow-400 ring-4 ring-yellow-400/20">
              <AvatarImage src={first.avatar_url || undefined} />
              <AvatarFallback className="text-xl bg-dark-surface">{getInitials(first.username)}</AvatarFallback>
            </Avatar>
            <div className="absolute -top-2 -right-2 h-7 w-7 rounded-full bg-yellow-400 flex items-center justify-center text-dark text-xs font-bold">
              1
            </div>
          </div>
          <Link href={`/profile/${first.username}`} className="font-medium hover:text-brand transition-colors">
            {first.username}
          </Link>
          <RankBadge rank={first.user_rank} size="sm" />
          <div className="font-display text-2xl text-yellow-400">{first.score?.toLocaleString()}</div>
          <div className="w-28 h-32 bg-dark-card rounded-t-lg border border-dark-border flex items-center justify-center">
            <Crown className="h-10 w-10 text-yellow-400" />
          </div>
        </div>
      )}

      {/* 3rd */}
      {third && (
        <div className="flex flex-col items-center gap-2 animate-slide-up" style={{ animationDelay: '200ms' }}>
          <div className="relative">
            <Avatar className="h-16 w-16 border-2 border-amber-600">
              <AvatarImage src={third.avatar_url || undefined} />
              <AvatarFallback className="text-lg bg-dark-surface">{getInitials(third.username)}</AvatarFallback>
            </Avatar>
            <div className="absolute -top-2 -right-2 h-6 w-6 rounded-full bg-amber-600 flex items-center justify-center text-white text-xs font-bold">
              3
            </div>
          </div>
          <Link href={`/profile/${third.username}`} className="font-medium text-sm hover:text-brand transition-colors">
            {third.username}
          </Link>
          <RankBadge rank={third.user_rank} size="sm" />
          <div className="font-display text-xl text-amber-600">{third.score?.toLocaleString()}</div>
          <div className="w-24 h-16 bg-dark-card rounded-t-lg border border-dark-border flex items-center justify-center">
            <Medal className="h-8 w-8 text-amber-600" />
          </div>
        </div>
      )}
    </div>
  );
}

