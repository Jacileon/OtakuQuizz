'use client';

import { Rank } from '@/types';
import { getRankColor, getRankConfig } from '@/lib/utils';
import { cn } from '@/lib/utils';

interface RankBadgeProps {
  rank: Rank;
  size?: 'sm' | 'md' | 'lg';
  showLabel?: boolean;
  className?: string;
}

const sizeClasses = {
  sm: 'text-xs px-2 py-0.5',
  md: 'text-sm px-3 py-1',
  lg: 'text-base px-4 py-1.5',
};

export function RankBadge({ rank, size = 'md', showLabel = false, className }: RankBadgeProps) {
  const config = getRankConfig(rank);
  const color = getRankColor(rank);
  const isHighRank = ['S', 'S+', 'SS', 'SSS', 'Légende'].includes(rank);

  return (
    <span
      className={cn(
        'inline-flex items-center gap-1 rounded-md font-display font-bold tracking-wider uppercase',
        sizeClasses[size],
        isHighRank && 'animate-shimmer',
        className
      )}
      style={{
        backgroundColor: `${color}20`,
        color: color,
        border: `1px solid ${color}40`,
        boxShadow: isHighRank ? `0 0 10px ${color}30` : 'none',
      }}
    >
      {rank}
      {showLabel && config && (
        <span className="font-sans font-normal opacity-70 text-[0.85em]">
          {config.label}
        </span>
      )}
    </span>
  );
}

