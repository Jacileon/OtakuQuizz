'use client';

import { useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { Edit2, Archive, Trash2, Loader2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { deleteQuiz, archiveQuiz } from '@/lib/actions/quiz';
import { toast } from '@/lib/hooks/useToast';

type QuizActionsProps = {
  quizId: string;
};

export function QuizActions({ quizId }: QuizActionsProps) {
  const router = useRouter();
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const [showArchiveDialog, setShowArchiveDialog] = useState(false);
  const [archiving, setArchiving] = useState(false);
  const [deleting, setDeleting] = useState(false);

  const handleArchive = async () => {
    setArchiving(true);
    const result = await archiveQuiz(quizId);
    setArchiving(false);
    setShowArchiveDialog(false);
    if (result.success) {
      toast({ title: 'Quiz archivé' });
      router.refresh();
    } else {
      toast({ title: 'Erreur', description: result.error, variant: 'destructive' });
    }
  };

  const handleDelete = async () => {
    setDeleting(true);
    const result = await deleteQuiz(quizId);
    setDeleting(false);
    setShowDeleteDialog(false);
    if (result.success) {
      toast({ title: 'Quiz supprimé' });
      router.refresh();
    } else {
      toast({ title: 'Erreur', description: result.error, variant: 'destructive' });
    }
  };

  return (
    <>
      <div className="flex gap-1">
        <Link href={`/quiz/${quizId}/edit`}>
          <Button variant="ghost" size="icon" className="h-8 w-8">
            <Edit2 className="h-3.5 w-3.5" />
          </Button>
        </Link>
        <Button
          variant="ghost"
          size="icon"
          className="h-8 w-8"
          onClick={() => setShowArchiveDialog(true)}
        >
          <Archive className="h-3.5 w-3.5" />
        </Button>
        <Button
          variant="ghost"
          size="icon"
          className="h-8 w-8 text-red-400 hover:text-red-300 hover:bg-red-400/10"
          onClick={() => setShowDeleteDialog(true)}
        >
          <Trash2 className="h-3.5 w-3.5" />
        </Button>
      </div>

      <Dialog open={showArchiveDialog} onOpenChange={setShowArchiveDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Archiver ce quiz ?</DialogTitle>
            <DialogDescription>
              Le quiz ne sera plus visible par les autres joueurs. Tu pourras le réactiver plus tard.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowArchiveDialog(false)} disabled={archiving}>
              Annuler
            </Button>
            <Button variant="secondary" onClick={handleArchive} disabled={archiving}>
              {archiving && <Loader2 className="h-4 w-4 mr-2 animate-spin" />}
              Archiver
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Supprimer ce quiz ?</DialogTitle>
            <DialogDescription>
              Cette action est irréversible. Le quiz et toutes ses données seront définitivement supprimés.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowDeleteDialog(false)} disabled={deleting}>
              Annuler
            </Button>
            <Button variant="destructive" onClick={handleDelete} disabled={deleting}>
              {deleting && <Loader2 className="h-4 w-4 mr-2 animate-spin" />}
              Supprimer
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
