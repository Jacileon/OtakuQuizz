'use client';

import { useEffect, useState } from 'react';
import { getCountdown } from '@/lib/utils';
import { Clock } from 'lucide-react';

interface EventCountdownProps {
  endDate: string;
  isStart?: boolean;
}

export function EventCountdown({ endDate, isStart = false }: EventCountdownProps) {
  const [countdown, setCountdown] = useState(getCountdown(endDate));

  useEffect(() => {
    const timer = setInterval(() => {
      setCountdown(getCountdown(endDate));
    }, 1000);
    return () => clearInterval(timer);
  }, [endDate]);

  if (countdown.total <= 0) {
    return (
      <div className="text-sm text-brand font-medium">
        {isStart ? 'Commencé !' : 'Terminé !'}
      </div>
    );
  }

  return (
    <div className="flex items-center gap-2 text-sm">
      <Clock className="h-4 w-4 text-accent" />
      <span className="text-muted-foreground">{isStart ? 'Dans' : 'Reste'}:</span>
      <span className="font-mono font-medium">
        {String(countdown.days).padStart(2, '0')}:
        {String(countdown.hours).padStart(2, '0')}:
        {String(countdown.minutes).padStart(2, '0')}:
        {String(countdown.seconds).padStart(2, '0')}
      </span>
    </div>
  );
}

