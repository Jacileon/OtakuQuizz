'use client';

import Link from '../../../node_modules/next/link';
import { Quiz } from '@/types';
import { Button } from '@/components/ui/button';
import { Zap, ChevronRight } from 'lucide-react';

interface LiveEventBannerProps {
  events: Quiz[];
}

export function LiveEventBanner({ events }: LiveEventBannerProps) {
  return (
    <div className="relative overflow-hidden rounded-lg bg-gradient-to-r from-brand/20 to-accent/20 border border-brand/30 p-4">
      <div className="absolute inset-0 bg-brand/5 animate-pulse" />
      <div className="relative flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="h-3 w-3 rounded-full bg-brand animate-pulse" />
          <div>
            <div className="font-medium text-brand">{events.length} ÉVÉNEMENT{events.length > 1 ? 'S' : ''} EN COURS</div>
            <div className="text-sm text-muted-foreground">
              {events[0]?.title}
            </div>
          </div>
        </div>
        <Link href={`/quiz/${events[0].id}/play`}>
          <Button size="sm" className="gap-1">
            Rejoindre <ChevronRight className="h-4 w-4" />
          </Button>
        </Link>
      </div>
    </div>
  );
}

