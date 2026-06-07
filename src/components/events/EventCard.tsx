'use client';

import { useEffect, useState } from 'react';
import Link from '../../../node_modules/next/link';
import { Quiz } from '@/types';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { EventCountdown } from './EventCountdown';
import { Zap, Users, Trophy, Calendar } from 'lucide-react';
import { cn, getCountdown } from '@/lib/utils';

interface EventCardProps {
  event: Quiz;
  status: 'live' | 'upcoming' | 'past';
}

export function EventCard({ event, status }: EventCardProps) {
  const isLive = status === 'live';
  const isUpcoming = status === 'upcoming';
  const isPast = status === 'past';

  return (
    <Card className={cn(
      'border-dark-border transition-all',
      isLive ? 'bg-brand/5 border-brand/30' : 'bg-dark-card/50'
    )}>
      <CardContent className="p-4 space-y-3">
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <div className="flex items-center gap-2 mb-1">
              <Badge variant={isLive ? 'default' : isUpcoming ? 'secondary' : 'outline'} className="gap-1">
                {isLive && <Zap className="h-3 w-3" />}
                {isLive ? 'LIVE' : isUpcoming ? 'À VENIR' : 'TERMINÉ'}
              </Badge>
              <span className="text-xs text-muted-foreground">{event.series}</span>
            </div>
            <h3 className="font-medium">{event.title}</h3>
            <p className="text-sm text-muted-foreground line-clamp-2 mt-1">{event.description}</p>
          </div>
        </div>

        <div className="flex items-center gap-4 text-xs text-muted-foreground">
          <span className="flex items-center gap-1">
            <Users className="h-3.5 w-3.5" /> {event.play_count} participants
          </span>
          <span className="flex items-center gap-1">
            <Trophy className="h-3.5 w-3.5" /> {event.question_count} questions
          </span>
        </div>

        {isLive && (
          <EventCountdown endDate={event.event_end_at!} />
        )}
        {isUpcoming && (
          <EventCountdown endDate={event.event_start_at!} isStart />
        )}

        <Link href={isPast ? `/leaderboard/quiz/${event.id}` : `/quiz/${event.id}/play`}>
          <Button 
            className="w-full" 
            variant={isLive ? 'default' : isUpcoming ? 'secondary' : 'outline'}
            disabled={isUpcoming}
          >
            {isLive ? 'Participer maintenant' : isUpcoming ? 'Bientôt disponible' : 'Voir les résultats'}
          </Button>
        </Link>
      </CardContent>
    </Card>
  );
}

