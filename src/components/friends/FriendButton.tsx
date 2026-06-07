'use client';

import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { UserPlus, UserMinus, Check, X, Loader2 } from 'lucide-react';
import { useFriendshipStatus } from '@/lib/hooks/useFriends';

interface FriendButtonProps {
  userId: string;
}

export function FriendButton({ userId }: FriendButtonProps) {
  const { status, isRequester, loading, sendRequest, accept, reject, remove } = useFriendshipStatus(userId);

  if (loading) {
    return (
      <Button variant="outline" disabled>
        <Loader2 className="h-4 w-4 animate-spin" />
      </Button>
    );
  }

  if (status === null) {
    return (
      <Button onClick={sendRequest}>
        <UserPlus className="h-4 w-4 mr-2" />
        Ajouter en ami
      </Button>
    );
  }

  if (status === 'pending' && isRequester) {
    return (
      <Badge variant="outline" className="px-4 py-2">
        Demande envoyée
      </Badge>
    );
  }

  if (status === 'pending' && !isRequester) {
    return (
      <div className="flex gap-2">
        <Button size="sm" onClick={accept}>
          <Check className="h-4 w-4 mr-1" />
          Accepter
        </Button>
        <Button size="sm" variant="outline" onClick={reject}>
          <X className="h-4 w-4" />
        </Button>
      </div>
    );
  }

  if (status === 'accepted') {
    return (
      <Button variant="outline" onClick={remove}>
        <UserMinus className="h-4 w-4 mr-2" />
        Ami
      </Button>
    );
  }

  return null;
}