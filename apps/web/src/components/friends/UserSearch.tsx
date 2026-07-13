'use client';

import { useState, useEffect, useRef } from 'react';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Badge } from '@/components/ui/badge';
import { Search, UserPlus, Loader2 } from 'lucide-react';
import { useUserSearch, useFriendshipStatus } from '@/lib/hooks/useFriends';
import { UserProfile } from '@/types';
import Link from 'next/link';

export function UserSearch() {
  const [query, setQuery] = useState('');
  const { results, loading, search } = useUserSearch();
  const debounceRef = useRef<NodeJS.Timeout>();

  useEffect(() => {
    if (debounceRef.current) clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(() => {
      search(query);
    }, 300);
    return () => {
      if (debounceRef.current) clearTimeout(debounceRef.current);
    };
  }, [query, search]);

  return (
    <div className="space-y-4">
      <div className="relative">
        <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground h-4 w-4" />
        <Input
          placeholder="Rechercher un utilisateur..."
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          className="pl-10"
        />
      </div>

      {loading && (
        <div className="flex justify-center py-8">
          <Loader2 className="h-6 w-6 animate-spin text-primary" />
        </div>
      )}

      {!loading && results.length > 0 && (
        <div className="space-y-2">
          {results.map((user) => (
            <UserCard key={user.id} user={user} />
          ))}
        </div>
      )}

      {!loading && query.length >= 2 && results.length === 0 && (
        <p className="text-center text-muted-foreground py-8">
          Aucun utilisateur trouvé
        </p>
      )}
    </div>
  );
}

function UserCard({ user }: { user: UserProfile }) {
  const { status, loading, sendRequest } = useFriendshipStatus(user.id);

  return (
    <Card>
      <CardContent className="flex items-center justify-between p-4">
        <Link href={`/profile/${user.username}`} className="flex items-center gap-3 hover:opacity-80">
          <Avatar>
            <AvatarImage src={user.avatar_url || undefined} />
            <AvatarFallback>{user.username[0].toUpperCase()}</AvatarFallback>
          </Avatar>
          <div>
            <p className="font-semibold">{user.username}</p>
            <div className="flex items-center gap-2">
              <Badge variant="secondary" className="text-xs">{user.rank}</Badge>
              <span className="text-xs text-muted-foreground">Niv. {user.level}</span>
            </div>
          </div>
        </Link>

        {loading ? (
          <Loader2 className="h-4 w-4 animate-spin" />
        ) : status === null || status === 'rejected' ? (
          <Button size="sm" onClick={sendRequest}>
            <UserPlus className="h-4 w-4 mr-1" />
            {status === 'rejected' ? 'Renvoyer' : 'Ajouter'}
          </Button>
        ) : status === 'pending' ? (
          <Badge variant="outline">Demande envoyée</Badge>
        ) : status === 'accepted' ? (
          <Badge variant="default">Ami</Badge>
        ) : null}
      </CardContent>
    </Card>
  );
}