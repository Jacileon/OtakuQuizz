'use client';

import { useEffect, useState } from 'react';
import { cn } from '@/lib/utils';

interface TimerProps {
  duration: number;
  onTimeUp: () => void;
  isActive: boolean;
}

export function Timer({ duration, onTimeUp, isActive }: TimerProps) {
  const [timeLeft, setTimeLeft] = useState(duration);

  useEffect(() => {
    if (!isActive) return;
    setTimeLeft(duration);

    const interval = setInterval(() => {
      setTimeLeft((prev) => {
        if (prev <= 1) {
          clearInterval(interval);
          onTimeUp();
          return 0;
        }
        return prev - 1;
      });
    }, 1000);

    return () => clearInterval(interval);
  }, [duration, isActive, onTimeUp]);

  const percentage = (timeLeft / duration) * 100;
  const isLow = timeLeft <= 5;
  const isMedium = timeLeft <= 10 && timeLeft > 5;

  return (
    <div className="flex items-center gap-4">
      <div className={cn(
        'font-display text-3xl tabular-nums transition-colors',
        isLow ? 'text-red-500 animate-timer-pulse' : isMedium ? 'text-orange-400' : 'text-white'
      )}>
        {String(timeLeft).padStart(2, '0')}
      </div>
      <div className="flex-1 h-3 bg-dark-surface rounded-full overflow-hidden">
        <div
          className={cn(
            'h-full rounded-full transition-all duration-1000 ease-linear',
            isLow ? 'bg-red-500' : isMedium ? 'bg-orange-400' : 'bg-brand'
          )}
          style={{ width: `${percentage}%` }}
        />
      </div>
    </div>
  );
}

