'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Flag, Loader2 } from 'lucide-react';
import { toast } from '@/lib/hooks/useToast';
import { reportUser } from '@/lib/actions/reports';

interface ReportUserDialogProps {
  userId: string;
  username: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function ReportUserDialog({ userId, username, open, onOpenChange }: ReportUserDialogProps) {
  const [reason, setReason] = useState('');
  const [description, setDescription] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async () => {
    if (!reason.trim()) {
      toast({ title: 'Erreur', description: 'Veuillez sélectionner une raison', variant: 'destructive' });
      return;
    }

    try {
      setLoading(true);
      await reportUser(userId, reason, description);
      toast({ title: 'Signalement envoyé', description: 'Votre signalement a été transmis aux administrateurs' });
      onOpenChange(false);
      setReason('');
      setDescription('');
    } catch (error: any) {
      toast({ title: 'Erreur', description: error.message, variant: 'destructive' });
    } finally {
      setLoading(false);
    }
  };

  const reasons = [
    { value: 'spam', label: 'Spam' },
    { value: 'harassment', label: 'Harcèlement' },
    { value: 'inappropriate', label: 'Contenu inapproprié' },
    { value: 'cheating', label: 'Triche' },
    { value: 'other', label: 'Autre' },
  ];

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Flag className="h-5 w-5 text-destructive" />
            Signaler {username}
          </DialogTitle>
          <DialogDescription>
            Votre signalement sera examiné par les administrateurs.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          <div>
            <label className="text-sm font-medium mb-2 block">Raison du signalement</label>
            <div className="grid grid-cols-2 gap-2">
              {reasons.map((r) => (
                <button
                  key={r.value}
                  onClick={() => setReason(r.value)}
                  className={`p-2 text-sm rounded-lg border transition-colors ${
                    reason === r.value
                      ? 'border-destructive bg-destructive/10 text-destructive'
                      : 'border-muted hover:border-muted-foreground/50'
                  }`}
                >
                  {r.label}
                </button>
              ))}
            </div>
          </div>

          <div>
            <label className="text-sm font-medium mb-1 block">Description (optionnel)</label>
            <Textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Décrivez le problème..."
              rows={3}
            />
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Annuler
          </Button>
          <Button
            variant="destructive"
            onClick={handleSubmit}
            disabled={loading || !reason}
          >
            {loading ? <Loader2 className="h-4 w-4 animate-spin mr-2" /> : null}
            Signaler
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}