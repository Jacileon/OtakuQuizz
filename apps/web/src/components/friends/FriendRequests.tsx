'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Check, X, Loader2, Bell } from 'lucide-react';
import { useFriendRequests } from '@/lib/hooks/useFriends';
import Link from 'next/link';

export function FriendRequests() {
  const { received, sent, loading, accept, reject } = useFriendRequests();

  if (loading) {
    return (
      <div className="flex justify-center py-8">
        <Loader2 className="h-6 w-6 animate-spin text-primary" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {received.length > 0 && (
        <div>
          <h3 className="text-sm font-semibold text-muted-foreground mb-3">
            Demandes reçues ({received.length})
          </h3>
          <div className="space-y-2">
            {received.map((request) => (
              <Card key={request.id}>
                <CardContent className="flex items-center justify-between p-4">
                  <Link href={`/profile/${request.requester.username}`} className="flex items-center gap-3 hover:opacity-80">
                    <Avatar>
                      <AvatarImage src={request.requester.avatar_url || undefined} />
                      <AvatarFallback>{request.requester.username[0].toUpperCase()}</AvatarFallback>
                    </Avatar>
                    <div>
                      <p className="font-semibold">{request.requester.username}</p>
                      <div className="flex items-center gap-2">
                        <Badge variant="secondary" className="text-xs">{request.requester.rank}</Badge>
                        <span className="text-xs text-muted-foreground">Niv. {request.requester.level}</span>
                      </div>
                    </div>
                  </Link>

                  <div className="flex gap-2">
                    <Button size="sm" onClick={() => accept(request.id)}>
                      <Check className="h-4 w-4 mr-1" />
                      Accepter
                    </Button>
                    <Button size="sm" variant="outline" onClick={() => reject(request.id)}>
                      <X className="h-4 w-4" />
                    </Button>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      )}

      {sent.length > 0 && (
        <div>
          <h3 className="text-sm font-semibold text-muted-foreground mb-3">
            Demandes envoyées ({sent.length})
          </h3>
          <div className="space-y-2">
            {sent.map((request) => (
              <Card key={request.id}>
                <CardContent className="flex items-center justify-between p-4">
                  <Link href={`/profile/${request.addressee.username}`} className="flex items-center gap-3 hover:opacity-80">
                    <Avatar>
                      <AvatarImage src={request.addressee.avatar_url || undefined} />
                      <AvatarFallback>{request.addressee.username[0].toUpperCase()}</AvatarFallback>
                    </Avatar>
                    <div>
                      <p className="font-semibold">{request.addressee.username}</p>
                      <div className="flex items-center gap-2">
                        <Badge variant="secondary" className="text-xs">{request.addressee.rank}</Badge>
                        <span className="text-xs text-muted-foreground">Niv. {request.addressee.level}</span>
                      </div>
                    </div>
                  </Link>

                  <Badge variant="outline">En attente</Badge>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      )}

      {received.length === 0 && sent.length === 0 && (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <Bell className="h-12 w-12 text-muted-foreground mb-4" />
            <p className="text-muted-foreground">Aucune demande d'ami</p>
          </CardContent>
        </Card>
      )}
    </div>
  );
}