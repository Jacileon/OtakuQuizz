'use client';

import { cn } from '@/lib/utils';
import { Card } from '@/components/ui/card';

interface AnswerButtonProps {
  answer: { id: string; answer_text: string };
  isSelected: boolean;
  onSelect: () => void;
  disabled?: boolean;
}

export function AnswerButton({ answer, isSelected, onSelect, disabled = false }: AnswerButtonProps) {
  return (
    <Card
      onClick={disabled ? undefined : onSelect}
      className={cn(
        'border-2 transition-all duration-200',
        !disabled && 'cursor-pointer hover:-translate-y-0.5',
        disabled && 'opacity-50 cursor-not-allowed',
        isSelected
          ? 'border-brand bg-brand/10 shadow-lg shadow-brand/10'
          : 'border-dark-border bg-dark-card hover:border-brand/50 hover:bg-dark-surface'
      )}
    >
      <div className="p-4">
        <p className={cn(
          'font-medium',
          isSelected ? 'text-brand' : 'text-white'
        )}>
          {answer.answer_text}
        </p>
      </div>
    </Card>
  );
}
