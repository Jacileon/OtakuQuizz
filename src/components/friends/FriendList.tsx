'use client';

import { useState } from 'react';
import { Card, CardContent } from '@/components/ui/card';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { UserMinus, Loader2, Users } from 'lucide-react';
import { useFriends } from '@/lib/hooks/useFriends';
import Link from 'next/link';

export function FriendList() {
  const { friends, loading, remove } = useFriends();
  const [deleteId, setDeleteId] = useState<string | null>(null);
  const [deleting, setDeleting] = useState(false);

  const handleDelete = async () => {
    if (!deleteId) return;
    setDeleting(true);
    await remove(deleteId);
    setDeleting(false);
    setDeleteId(null);
  };

  if (loading) {
    return (
      <div className="flex justify-center py-8">
        <Loader2 className="h-6 w-6 animate-spin text-primary" />
      </div>
    );
  }

  if (friends.length === 0) {
    return (
      <Card>
        <CardContent className="flex flex-col items-center justify-center py-12">
          <Users className="h-12 w-12 text-muted-foreground mb-4" />
          <p className="text-muted-foreground">Pas encore d'amis</p>
          <p className="text-sm text-muted-foreground">Recherchez des utilisateurs pour les ajouter</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <>
      <div className="space-y-2">
        {friends.map((friendship) => (
          <Card key={friendship.id}>
            <CardContent className="flex items-center justify-between p-4">
              <Link href={`/profile/${friendship.friend.username}`} className="flex items-center gap-3 hover:opacity-80">
                <Avatar>
                  <AvatarImage src={friendship.friend.avatar_url || undefined} />
                  <AvatarFallback>{friendship.friend.username[0].toUpperCase()}</AvatarFallback>
                </Avatar>
                <div>
                  <p className="font-semibold">{friendship.friend.username}</p>
                  <div className="flex items-center gap-2">
                    <Badge variant="secondary" className="text-xs">{friendship.friend.rank}</Badge>
                    <span className="text-xs text-muted-foreground">Niv. {friendship.friend.level}</span>
                  </div>
                </div>
              </Link>

              <Button
                variant="ghost"
                size="sm"
                onClick={() => setDeleteId(friendship.id)}
                className="text-destructive hover:text-destructive"
              >
                <UserMinus className="h-4 w-4" />
              </Button>
            </CardContent>
          </Card>
        ))}
      </div>

      <Dialog open={!!deleteId} onOpenChange={() => setDeleteId(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Supprimer cet ami ?</DialogTitle>
            <DialogDescription>
              Cette action est irréversible. L'ami sera retiré de votre liste.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteId(null)} disabled={deleting}>
              Annuler
            </Button>
            <Button variant="destructive" onClick={handleDelete} disabled={deleting}>
              {deleting ? <Loader2 className="h-4 w-4 animate-spin mr-2" /> : null}
              Supprimer
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}