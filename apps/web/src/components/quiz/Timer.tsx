'use client';

import { useEffect, useState, useCallback } from 'react';
import { cn } from '@/lib/utils';
import { Clock, Timer as TimerIcon } from 'lucide-react';

interface TimerProps {
  duration: number;
  onTimeUp: () => void;
  isActive: boolean;
  mode?: 'per_question' | 'global';
  totalQuestions?: number;
  currentQuestion?: number;
}

export function Timer({ duration, onTimeUp, isActive, mode = 'per_question', totalQuestions, currentQuestion }: TimerProps) {
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

  const formatTime = (seconds: number) => {
    if (mode === 'global') {
      const mins = Math.floor(seconds / 60);
      const secs = seconds % 60;
      return `${String(mins).padStart(2, '0')}:${String(secs).padStart(2, '0')}`;
    }
    return String(seconds).padStart(2, '0');
  };

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          {mode === 'global' ? (
            <>
              <Clock className="h-4 w-4" />
              <span>Temps restant</span>
            </>
          ) : (
            <>
              <TimerIcon className="h-4 w-4" />
              <span>Question {currentQuestion}/{totalQuestions}</span>
            </>
          )}
        </div>
        <div className={cn(
          'font-display text-2xl tabular-nums transition-colors',
          isLow ? 'text-red-500 animate-timer-pulse' : isMedium ? 'text-orange-400' : 'text-white'
        )}>
          {formatTime(timeLeft)}
        </div>
      </div>
      <div className="h-2 bg-dark-surface rounded-full overflow-hidden">
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