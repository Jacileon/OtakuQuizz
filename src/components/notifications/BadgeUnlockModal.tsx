'use client';

import { useEffect, useState } from 'react';
import { Badge } from '@/types';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Award, Zap, X } from 'lucide-react';
import confetti from 'canvas-confetti';

interface BadgeUnlockModalProps {
  badges: Badge[];
}

export function BadgeUnlockModal({ badges }: BadgeUnlockModalProps) {
  const [open, setOpen] = useState(true);
  const [currentIndex, setCurrentIndex] = useState(0);

  useEffect(() => {
    if (open && badges.length > 0) {
      confetti({
        particleCount: 100,
        spread: 70,
        origin: { y: 0.6 },
        colors: ['#E63946', '#F4A261', '#FFD700', '#FF69B4'],
      });
    }
  }, [open, badges]);

  if (!badges.length) return null;

  const currentBadge = badges[currentIndex];

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogContent className="border-brand/30 bg-dark-card max-w-sm">
        <DialogHeader className="text-center">
          <DialogTitle className="font-display text-2xl tracking-wider">
            BADGE DÉBLOQUÉ !
          </DialogTitle>
        </DialogHeader>

        <div className="text-center space-y-4 py-4">
          <div className={cn(
            'h-20 w-20 rounded-full flex items-center justify-center mx-auto',
            currentBadge.is_rare ? 'bg-yellow-500/20 animate-pulse' : 'bg-brand/20'
          )}>
            <Award className={cn(
              'h-10 w-10',
              currentBadge.is_rare ? 'text-yellow-400' : 'text-brand'
            )} />
          </div>

          <div>
            <h3 className="font-display text-xl">{currentBadge.name}</h3>
            <p className="text-sm text-muted-foreground mt-1">{currentBadge.description}</p>
          </div>

          {currentBadge.is_rare && (
            <div className="text-yellow-400 text-sm font-medium">★ Badge Rare</div>
          )}

          <div className="flex items-center justify-center gap-2 text-accent">
            <Zap className="h-4 w-4" />
            <span>+XP Bonus</span>
          </div>
        </div>

        <div className="flex gap-2">
          {currentIndex < badges.length - 1 ? (
            <Button onClick={() => setCurrentIndex((i) => i + 1)} className="w-full">
              Suivant
            </Button>
          ) : (
            <Button onClick={() => setOpen(false)} className="w-full">
              Super !
            </Button>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}

import { cn } from '@/lib/utils';

