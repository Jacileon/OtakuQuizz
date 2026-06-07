'use client';

import { useEffect, useState } from 'react';
import { Rank } from '@/types';
import { getCurrentRankProgress, getRankColor } from '@/lib/utils';
import { cn } from '@/lib/utils';

interface XPBarProps {
  currentXP: number;
  rank: Rank;
  showNumbers?: boolean;
  className?: string;
}

export function XPBar({ currentXP, rank, showNumbers = true, className }: XPBarProps) {
  const [animatedPercent, setAnimatedPercent] = useState(0);
  const progress = getCurrentRankProgress(currentXP);
  const color = getRankColor(rank);

  useEffect(() => {
    const timer = setTimeout(() => setAnimatedPercent(progress.percent), 100);
    return () => clearTimeout(timer);
  }, [progress.percent]);

  return (
    <div className={cn('w-full', className)}>
      {showNumbers && (
        <div className="flex justify-between text-xs text-muted-foreground mb-1">
          <span>XP: {currentXP.toLocaleString()}</span>
          <span>
            {progress.current.toLocaleString()} / {progress.next.toLocaleString()}
          </span>
        </div>
      )}
      <div className="relative h-2.5 w-full overflow-hidden rounded-full bg-dark-surface">
        <div
          className="h-full rounded-full transition-all duration-1000 ease-out"
          style={{
            width: `${animatedPercent}%`,
            backgroundColor: color,
            boxShadow: `0 0 8px ${color}50`,
          }}
        />
      </div>
    </div>
  );
}

