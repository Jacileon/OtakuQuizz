'use client';

import Link from '../../../node_modules/next/link';
import { Skeleton } from '@/components/ui/skeleton';
import { TrendingUp, Flame, HelpCircle } from 'lucide-react';
import { useEffect, useState } from 'react';
import { getBrowserClient } from '@/lib/supabase/client';
import { LeaderboardEntry } from '@/types';
import { RankBadge } from '@/components/ui/RankBadge';

export function Sidebar() {
  const [weeklyTop, setWeeklyTop] = useState<LeaderboardEntry[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchWeekly = async () => {
      const supabase = getBrowserClient();
      const { data } = await supabase
        .rpc('get_global_leaderboard', { limit_count: 5 });
      if (data) setWeeklyTop(data as LeaderboardEntry[]);
      setLoading(false);
    };
    fetchWeekly();
  }, []);

  return (
    <div className="sticky top-16 h-[calc(100vh-64px)] overflow-y-auto border-r border-dark-border p-4 space-y-6">
      {/* Mini Leaderboard */}
      <div>
        <h3 className="font-display text-lg tracking-wider mb-3 flex items-center gap-2">
          <TrendingUp className="h-4 w-4 text-brand" />
          TOP 5 HEBDO
        </h3>
        <div className="space-y-2">
          {loading ? (
            Array.from({ length: 5 }).map((_, i) => (
              <Skeleton key={i} className="h-10 w-full" />
            ))
          ) : (
            weeklyTop.map((entry, i) => (
              <div
                key={entry.user_id}
                className="flex items-center gap-2 p-2 rounded-md hover:bg-dark-surface transition-colors"
              >
                <span className={
                  i === 0 ? 'text-yellow-400 font-bold' :
                  i === 1 ? 'text-gray-300 font-bold' :
                  i === 2 ? 'text-amber-600 font-bold' :
                  'text-muted-foreground'
                }>
                  #{i + 1}
                </span>
                <span className="text-sm truncate flex-1">{entry.username}</span>
                <RankBadge rank={entry.user_rank} size="sm" />
              </div>
            ))
          )}
        </div>
      </div>

      {/* Popular Categories */}
      <div>
        <h3 className="font-display text-lg tracking-wider mb-3 flex items-center gap-2">
          <Flame className="h-4 w-4 text-accent" />
          CATÉGORIES
        </h3>
        <div className="flex flex-wrap gap-2">
          {['Shonen', 'Isekai', 'Openings', 'Ghibli', 'Seinen'].map((cat) => (
            <Link
              key={cat}
              href={`/explore?category=${cat}`}
              className="px-2 py-1 text-xs rounded-md bg-dark-surface hover:bg-dark-card border border-dark-border transition-colors"
            >
              {cat}
            </Link>
          ))}
        </div>
      </div>

      {/* FAQ Link */}
      <div>
        <Link
          href="/faq"
          className="flex items-center gap-2 p-3 rounded-lg hover:bg-dark-surface transition-colors"
        >
          <HelpCircle className="h-5 w-5 text-muted-foreground" />
          <span className="text-sm">FAQ - Aide</span>
        </Link>
      </div>
    </div>
  );
}

