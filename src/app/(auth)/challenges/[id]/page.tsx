'use client';

import { useState, useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Loader2, Swords, Users, Zap, Clock, Trophy, ArrowLeft } from 'lucide-react';
import { useChallengeSession } from '@/lib/hooks/useChallenges';
import { useAuth } from '@/components/providers/AuthProvider';
import { useFriends } from '@/lib/hooks/useFriends';
import { inviteToChallenge, startChallenge } from '@/lib/actions/challenges';
import { toast } from '@/lib/hooks/useToast';
import { formatDistanceToNow } from 'date-fns';
import { fr } from 'date-fns/locale';
import { cn } from '@/lib/utils';

export default function ChallengePage() {
  const params = useParams();
  const router = useRouter();
  const sessionId = params.id as string;
  const { user } = useAuth();
  const { session, loading, refetch } = useChallengeSession(sessionId);
  const { friends } = useFriends();
  const [starting, setStarting] = useState(false);
  const [inviting, setInviting] = useState<string | null>(null);
  const [showInvite, setShowInvite] = useState(false);

  useEffect(() => {
    const interval = setInterval(refetch, 5000);
    return () => clearInterval(interval);
  }, [refetch]);

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-dark">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  if (!session) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-dark p-4">
        <Card className="max-w-md w-full">
          <CardContent className="p-8 text-center">
            <Swords className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
            <h2 className="text-xl font-semibold mb-2">Défi introuvable</h2>
            <p className="text-muted-foreground mb-4">Ce défi n'existe pas ou vous n'y avez pas accès.</p>
            <Button onClick={() => router.push('/friends')}>Retour</Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  const participants = session.participants || [];
  const invitations = session.invitations || [];
  const isCreator = user?.id === session.creator_id;
  const acceptedCount = participants.filter((p: any) => p.status === 'accepted').length;
  const canStart = isCreator && acceptedCount >= 2 && session.status === 'waiting';

  const handleInvite = async (friendId: string) => {
    try {
      setInviting(friendId);
      await inviteToChallenge(sessionId, friendId);
      toast({ title: 'Invitation envoyée' });
      refetch();
    } catch (error: any) {
      toast({ title: 'Erreur', description: error.message, variant: 'destructive' });
    } finally {
      setInviting(null);
    }
  };

  const handleStart = async () => {
    try {
      setStarting(true);
      await startChallenge(sessionId);
      toast({ title: 'Défi lancé !' });
      refetch();
    } catch (error: any) {
      toast({ title: 'Erreur', description: error.message, variant: 'destructive' });
    } finally {
      setStarting(false);
    }
  };

  return (
    <div className="p-4 lg:p-8 pb-24 md:pb-8">
      <div className="max-w-2xl mx-auto">
        <Button variant="ghost" onClick={() => router.push('/friends')} className="gap-2 mb-6">
          <ArrowLeft className="h-4 w-4" />
          Retour
        </Button>

        <Card className="mb-6 border-purple-500/30">
          <CardHeader>
            <CardTitle className="flex items-center gap-3">
              <Swords className="h-6 w-6 text-purple-500" />
              Défi en cours
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="p-4 rounded-lg bg-dark-surface">
              <p className="text-sm text-muted-foreground">Quiz</p>
              <p className="font-medium">{session.quiz?.title || 'Quiz'}</p>
            </div>

            <div className="flex items-center justify-between p-4 rounded-lg bg-dark-surface">
              <div className="flex items-center gap-2">
                <Zap className="h-5 w-5 text-yellow-500" />
                <span className="font-medium">Pool XP total</span>
              </div>
              <span className="text-2xl font-bold text-yellow-500">{session.total_xp_pool || 0} XP</span>
            </div>

            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <Clock className="h-4 w-4" />
              <span>
                {session.status === 'waiting' && `Expire ${formatDistanceToNow(new Date(session.invite_expires_at), { addSuffix: true, locale: fr })}`}
                {session.status === 'ready' && 'Prêt à lancer'}
                {session.status === 'playing' && 'En cours'}
                {session.status === 'completed' && 'Terminé'}
              </span>
            </div>

            <Badge variant={session.status === 'waiting' ? 'secondary' : session.status === 'completed' ? 'default' : 'outline'}>
              {session.status === 'waiting' ? 'En attente' : session.status === 'ready' ? 'Prêt' : session.status === 'playing' ? 'En cours' : 'Terminé'}
            </Badge>
          </CardContent>
        </Card>

        <Card className="mb-6">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Users className="h-5 w-5" />
              Participants ({acceptedCount})
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {participants.map((participant: any) => (
              <div
                key={participant.id}
                className={cn(
                  'flex items-center gap-3 p-3 rounded-lg border',
                  participant.status === 'accepted' ? 'border-green-500/30 bg-green-500/5' : 'border-muted'
                )}
              >
                <Avatar>
                  <AvatarImage src={participant.user?.avatar_url || undefined} />
                  <AvatarFallback>{participant.user?.username?.[0]?.toUpperCase()}</AvatarFallback>
                </Avatar>
                <div className="flex-1">
                  <p className="font-medium">{participant.user?.nickname || participant.user?.username}</p>
                  <div className="flex items-center gap-2 text-xs text-muted-foreground">
                    <Zap className="h-3 w-3 text-yellow-500" />
                    <span>Mise: {participant.xp_bet} XP</span>
                  </div>
                </div>
                <Badge variant={participant.status === 'accepted' ? 'default' : 'secondary'}>
                  {participant.status === 'accepted' ? 'Prêt' : participant.status}
                </Badge>
              </div>
            ))}

            {invitations.filter((i: any) => i.status === 'pending').map((invitation: any) => (
              <div
                key={invitation.id}
                className="flex items-center gap-3 p-3 rounded-lg border border-dashed"
              >
                <Avatar>
                  <AvatarImage src={invitation.invitee?.avatar_url || undefined} />
                  <AvatarFallback>{invitation.invitee?.username?.[0]?.toUpperCase()}</AvatarFallback>
                </Avatar>
                <div className="flex-1">
                  <p className="font-medium">{invitation.invitee?.nickname || invitation.invitee?.username}</p>
                  <p className="text-xs text-muted-foreground">Invitation envoyée</p>
                </div>
                <Badge variant="outline">En attente</Badge>
              </div>
            ))}
          </CardContent>
        </Card>

        {isCreator && session.status === 'waiting' && (
          <div className="space-y-4">
            {showInvite ? (
              <Card>
                <CardHeader>
                  <CardTitle className="text-base">Inviter des amis</CardTitle>
                </CardHeader>
                <CardContent className="space-y-2">
                  {friends.length === 0 ? (
                    <p className="text-sm text-muted-foreground">Aucun ami à inviter</p>
                  ) : (
                    friends.map((f) => {
                      const isAlreadyInvited = participants.some((p: any) => p.user_id === f.friend.id) ||
                        invitations.some((i: any) => i.invitee_id === f.friend.id);
                      
                      return (
                        <div key={f.id} className="flex items-center gap-3 p-2 rounded-lg hover:bg-accent/50">
                          <Avatar className="h-8 w-8">
                            <AvatarImage src={f.friend.avatar_url || undefined} />
                            <AvatarFallback>{f.friend.username[0].toUpperCase()}</AvatarFallback>
                          </Avatar>
                          <span className="flex-1 text-sm font-medium">{f.friend.nickname || f.friend.username}</span>
                          {!isAlreadyInvited && (
                            <Button
                              size="sm"
                              variant="outline"
                              onClick={() => handleInvite(f.friend.id)}
                              disabled={inviting === f.friend.id}
                            >
                              {inviting === f.friend.id ? <Loader2 className="h-3 w-3 animate-spin" /> : 'Inviter'}
                            </Button>
                          )}
                        </div>
                      );
                    })
                  )}
                  <Button variant="ghost" size="sm" onClick={() => setShowInvite(false)} className="w-full">
                    Fermer
                  </Button>
                </CardContent>
              </Card>
            ) : (
              <Button variant="outline" onClick={() => setShowInvite(true)} className="w-full gap-2">
                <Users className="h-4 w-4" />
                Inviter des amis
              </Button>
            )}

            <Button
              onClick={handleStart}
              disabled={!canStart || starting}
              className="w-full gap-2"
              size="lg"
            >
              {starting ? <Loader2 className="h-5 w-5 animate-spin" /> : <Swords className="h-5 w-5" />}
              {canStart ? 'Lancer le défi' : `En attente de ${2 - acceptedCount} joueur(s)`}
            </Button>
          </div>
        )}

        {session.status === 'completed' && (
          <Card>
            <CardContent className="p-8 text-center">
              <Trophy className="h-12 w-12 text-yellow-500 mx-auto mb-4" />
              <h2 className="text-xl font-semibold mb-2">Défi terminé !</h2>
              <p className="text-muted-foreground">
                {session.winner_id === user?.id ? 'Vous avez gagné !' : 'Consultez les résultats'}
              </p>
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  );
}