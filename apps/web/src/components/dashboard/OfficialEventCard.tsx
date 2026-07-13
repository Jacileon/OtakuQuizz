'use client';

import { useEffect, useState } from 'react';
import Link from '../../../node_modules/next/link';
import { Quiz } from '@/types';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Zap, Users, Clock } from 'lucide-react';
import { getCountdown, isQuizOfficialActive } from '@/lib/utils';

interface OfficialEventCardProps {
  event: Quiz;
  status: 'live' | 'upcoming';
}

export function OfficialEventCard({ event, status }: OfficialEventCardProps) {
  const [countdown, setCountdown] = useState(getCountdown(event.event_end_at || event.event_start_at || ''));
  const isLive = status === 'live';

  useEffect(() => {
    if (!isLive) return;
    const timer = setInterval(() => {
      setCountdown(getCountdown(event.event_end_at || ''));
    }, 1000);
    return () => clearInterval(timer);
  }, [event.event_end_at, isLive]);

  return (
    <Card className={cn(
      'border-dark-border transition-all hover:-translate-y-0.5',
      isLive ? 'bg-brand/5 border-brand/30' : 'bg-dark-card/50'
    )}>
      <CardContent className="p-4 space-y-3">
        <div className="flex items-center justify-between">
          <Badge variant={isLive ? 'default' : 'secondary'} className="gap-1">
            <Zap className="h-3 w-3" />
            {isLive ? 'EN COURS' : 'À VENIR'}
          </Badge>
          <div className="flex items-center gap-1 text-xs text-muted-foreground">
            <Users className="h-3 w-3" />
            {event.play_count} participants
          </div>
        </div>

        <h3 className="font-medium line-clamp-1">{event.title}</h3>
        <p className="text-xs text-muted-foreground line-clamp-2">{event.description}</p>

        <div className="flex items-center gap-2 text-sm">
          <Clock className="h-4 w-4 text-accent" />
          {isLive ? (
            <span className="font-mono text-brand">
              {String(countdown.hours).padStart(2, '0')}:
              {String(countdown.minutes).padStart(2, '0')}:
              {String(countdown.seconds).padStart(2, '0')}
            </span>
          ) : (
            <span className="text-muted-foreground">
              Dans {countdown.days}j {countdown.hours}h
            </span>
          )}
        </div>

        <Link href={`/quiz/${event.id}/play`}>
          <Button className="w-full" variant={isLive ? 'default' : 'secondary'} disabled={!isLive}>
            {isLive ? 'Participer maintenant' : 'Bientôt disponible'}
          </Button>
        </Link>
      </CardContent>
    </Card>
  );
}

import { cn } from '@/lib/utils';

