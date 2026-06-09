'use client';

import { useEffect, useState } from 'react';
import { Progress } from '@/components/ui/progress';
import { Badge } from '@/components/ui/badge';
import { RankBadge } from '@/components/ui/RankBadge';
import { Trophy, ArrowRight } from 'lucide-react';
import { createClient } from '@/lib/supabase/client';
import { cn } from '@/lib/utils';
import { Rank } from '@/types';

interface RankProgressBarProps {
  currentXP: number;
  currentRank: string;
  className?: string;
}

interface RankConfigEntry {
  rank_label: string;
  xp_required: number;
  display_order: number;
}

export function RankProgressBar({ currentXP, currentRank, className }: RankProgressBarProps) {
  const [rankConfig, setRankConfig] = useState<RankConfigEntry[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchRankConfig = async () => {
      const supabase = createClient();
      const { data } = await supabase
        .from('rank_config')
        .select('*')
        .order('display_order', { ascending: true });
      
      if (data) setRankConfig(data as RankConfigEntry[]);
      setLoading(false);
    };

    fetchRankConfig();
  }, []);

  if (loading || rankConfig.length === 0) {
    return null;
  }

  // Trouver le rang actuel et le suivant
  const currentRankIndex = rankConfig.findIndex(r => r.rank_label === currentRank);
  const currentRankConf = rankConfig[currentRankIndex];
  const nextRankConf = rankConfig[currentRankIndex + 1];

  // Si rang maximum
  if (!nextRankConf) {
    return (
      <div className={cn('p-4 rounded-lg bg-gradient-to-r from-yellow-500/10 to-amber-500/10 border border-yellow-500/20', className)}>
        <div className="flex items-center gap-3">
          <RankBadge rank={currentRank as Rank} size="lg" />
          <div>
            <p className="font-semibold">Rang maximum atteint 🏆</p>
            <p className="text-sm text-muted-foreground">
              {currentXP.toLocaleString()} XP
            </p>
          </div>
        </div>
      </div>
    );
  }

  // Calculer la progression
  const xpInCurrentRank = currentXP - currentRankConf.xp_required;
  const xpNeededForNextRank = nextRankConf.xp_required - currentRankConf.xp_required;
  const progressPercent = Math.min(100, Math.max(0, (xpInCurrentRank / xpNeededForNextRank) * 100));

  return (
    <div className={cn('p-4 rounded-lg bg-dark-surface border border-dark-border', className)}>
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-2">
          <RankBadge rank={currentRank as Rank} size="sm" />
          <span className="font-medium">Rang {currentRank}</span>
        </div>
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <ArrowRight className="h-4 w-4" />
          <RankBadge rank={nextRankConf.rank_label as Rank} size="sm" />
          <span>Rang {nextRankConf.rank_label}</span>
        </div>
      </div>

      <Progress value={progressPercent} className="h-3 mb-2" />

      <div className="flex items-center justify-between text-sm">
        <span className="text-muted-foreground">
          {currentXP.toLocaleString()} XP
        </span>
        <span className="text-muted-foreground">
          {nextRankConf.xp_required.toLocaleString()} XP
        </span>
      </div>

      <p className="text-xs text-muted-foreground mt-2 text-center">
        Encore {(nextRankConf.xp_required - currentXP).toLocaleString()} XP pour le rang {nextRankConf.rank_label}
      </p>
    </div>
  );
}