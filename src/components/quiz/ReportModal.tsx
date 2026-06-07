'use client';

import { useState } from '../../../node_modules/@types/react';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { reportQuiz } from '@/lib/actions/reports';
import { toast } from '@/lib/hooks/useToast';
import { AlertTriangle, Flag } from 'lucide-react';

const reasons = [
  { value: 'wrong_answer', label: 'Mauvaise réponse' },
  { value: 'incorrect_content', label: 'Contenu incorrect' },
  { value: 'spam', label: 'Spam' },
  { value: 'plagiarism', label: 'Plagiat' },
  { value: 'inappropriate', label: 'Contenu inapproprié' },
];

interface ReportModalProps {
  quizId: string;
  open: boolean;
  onClose: () => void;
}

export function ReportModal({ quizId, open, onClose }: ReportModalProps) {
  const [selectedReason, setSelectedReason] = useState('');
  const [description, setDescription] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async () => {
    if (!selectedReason) return;
    setIsSubmitting(true);

    const result = await reportQuiz(quizId, selectedReason as any, description);

    if (result.success) {
      toast({ title: 'Signalement envoyé', description: 'Merci pour ta contribution !' });
      onClose();
    } else {
      toast({ title: 'Erreur', description: result.error || 'Erreur', variant: 'destructive' });
    }

    setIsSubmitting(false);
  };

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="border-dark-border bg-dark-card">
        <DialogHeader>
          <DialogTitle className="font-display text-xl tracking-wider flex items-center gap-2">
            <Flag className="h-5 w-5 text-brand" />
            SIGNALER UN QUIZ
          </DialogTitle>
        </DialogHeader>

        <div className="space-y-4 py-4">
          <div className="space-y-2">
            <span className="text-sm font-medium">Raison</span>
            {reasons.map((reason) => (
              <button
                key={reason.value}
                onClick={() => setSelectedReason(reason.value)}
                className={cn(
                  'w-full p-3 rounded-lg border text-left text-sm transition-colors',
                  selectedReason === reason.value
                    ? 'border-brand bg-brand/10 text-brand'
                    : 'border-dark-border bg-dark-surface hover:border-brand/50'
                )}
              >
                {reason.label}
              </button>
            ))}
          </div>

          <div>
            <span className="text-sm font-medium">Description (optionnel)</span>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Décris le problème..."
              maxLength={300}
              className="w-full p-3 rounded-lg bg-dark-surface border border-dark-border text-white placeholder:text-muted-foreground focus:border-brand focus:outline-none resize-none mt-2"
              rows={3}
            />
            <div className="text-xs text-muted-foreground mt-1">{description.length}/300</div>
          </div>

          <Button onClick={handleSubmit} disabled={!selectedReason || isSubmitting} className="w-full gap-2">
            <AlertTriangle className="h-4 w-4" />
            {isSubmitting ? 'Envoi...' : 'Envoyer le signalement'}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}

import { cn } from '@/lib/utils';

