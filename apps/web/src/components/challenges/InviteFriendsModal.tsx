'use client';

import { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Loader2, Search, UserPlus, X, Check } from 'lucide-react';
import { useFriends } from '@/lib/hooks/useFriends';
import { inviteToChallenge } from '@/lib/actions/challenges';
import { toast } from '@/lib/hooks/useToast';
import { UserProfile } from '@/types';

export function InviteFriendsModal({ sessionId, onClose }: { sessionId: string; onClose: () => void }) {
  const { friends, loading } = useFriends();
  const [search, setSearch] = useState('');
  const [inviting, setInviting] = useState<string | null>(null);
  const [invited, setInvited] = useState<Set<string>>(new Set());

  const filteredFriends = friends.filter(f =>
    f.friend.username.toLowerCase().includes(search.toLowerCase())
  );

  const handleInvite = async (friendId: string) => {
    try {
      setInviting(friendId);
      await inviteToChallenge(sessionId, friendId);
      setInvited(prev => new Set(Array.from(prev).concat(friendId)));
      toast({ title: 'Invitation envoyée' });
    } catch (error: any) {
      toast({ title: 'Erreur', description: error.message, variant: 'destructive' });
    } finally {
      setInviting(null);
    }
  };

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between">
        <CardTitle className="flex items-center gap-2">
          <UserPlus className="h-5 w-5" />
          Inviter des amis
        </CardTitle>
        <Button variant="ghost" size="sm" onClick={onClose}>
          <X className="h-4 w-4" />
        </Button>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground h-4 w-4" />
          <Input
            placeholder="Rechercher un ami..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-10"
          />
        </div>

        {loading ? (
          <div className="flex justify-center py-8">
            <Loader2 className="h-6 w-6 animate-spin" />
          </div>
        ) : filteredFriends.length === 0 ? (
          <div className="text-center py-8 text-muted-foreground">
            <p>Aucun ami trouvé</p>
          </div>
        ) : (
          <div className="space-y-2 max-h-[300px] overflow-y-auto">
            {filteredFriends.map((friendship) => {
              const friend = friendship.friend;
              const isInvited = invited.has(friend.id);

              return (
                <div
                  key={friendship.id}
                  className="flex items-center gap-3 p-3 rounded-lg border hover:bg-accent/50 transition-colors"
                >
                  <Avatar>
                    <AvatarImage src={friend.avatar_url || undefined} />
                    <AvatarFallback>{friend.username[0].toUpperCase()}</AvatarFallback>
                  </Avatar>
                  <div className="flex-1 min-w-0">
                    <p className="font-medium truncate">{friend.username}</p>
                    <div className="flex items-center gap-2">
                      <Badge variant="secondary" className="text-xs">{friend.rank}</Badge>
                      <span className="text-xs text-muted-foreground">Niv. {friend.level}</span>
                    </div>
                  </div>
                  <Button
                    size="sm"
                    variant={isInvited ? "secondary" : "default"}
                    onClick={() => !isInvited && handleInvite(friend.id)}
                    disabled={isInvited || inviting === friend.id}
                  >
                    {inviting === friend.id ? (
                      <Loader2 className="h-4 w-4 animate-spin" />
                    ) : isInvited ? (
                      <>
                        <Check className="h-4 w-4 mr-1" />
                        Invité
                      </>
                    ) : (
                      <>
                        <UserPlus className="h-4 w-4 mr-1" />
                        Inviter
                      </>
                    )}
                  </Button>
                </div>
              );
            })}
          </div>
        )}
      </CardContent>
    </Card>
  );
}