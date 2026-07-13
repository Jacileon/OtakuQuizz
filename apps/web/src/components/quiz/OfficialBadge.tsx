import { Badge } from '@/components/ui/badge';
import { Shield, Trophy, Swords } from 'lucide-react';
import { cn } from '@/lib/utils';

interface OfficialBadgeProps {
  type: 'official' | 'challenge';
  size?: 'sm' | 'md' | 'lg';
  className?: string;
}

export function OfficialBadge({ type, size = 'sm', className }: OfficialBadgeProps) {
  if (type === 'official') {
    return (
      <Badge
        variant="default"
        className={cn(
          'bg-gradient-to-r from-yellow-500 to-amber-500 text-white border-0 gap-1',
          size === 'sm' && 'text-xs px-2 py-0.5',
          size === 'md' && 'text-sm px-3 py-1',
          size === 'lg' && 'text-base px-4 py-1.5',
          className
        )}
      >
        <Shield className={cn(
          size === 'sm' && 'h-3 w-3',
          size === 'md' && 'h-4 w-4',
          size === 'lg' && 'h-5 w-5',
        )} />
        Officiel
      </Badge>
    );
  }

  return (
    <Badge
      variant="default"
      className={cn(
        'bg-gradient-to-r from-purple-500 to-pink-500 text-white border-0 gap-1',
        size === 'sm' && 'text-xs px-2 py-0.5',
        size === 'md' && 'text-sm px-3 py-1',
        size === 'lg' && 'text-base px-4 py-1.5',
        className
      )}
    >
      <Swords className={cn(
        size === 'sm' && 'h-3 w-3',
        size === 'md' && 'h-4 w-4',
        size === 'lg' && 'h-5 w-5',
      )} />
      Défi
    </Badge>
  );
}

export function OfficialBanner({ className }: { className?: string }) {
  return (
    <div className={cn(
      'absolute top-2 right-2 z-10',
      className
    )}>
      <OfficialBadge type="official" size="sm" />
    </div>
  );
}

export function ChallengeBanner({ className }: { className?: string }) {
  return (
    <div className={cn(
      'absolute top-2 right-2 z-10',
      className
    )}>
      <OfficialBadge type="challenge" size="sm" />
    </div>
  );
}