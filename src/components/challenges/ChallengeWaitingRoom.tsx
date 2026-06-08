'use client';

import { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Loader2, Users, Swords, Clock, Trophy, Zap, UserPlus, Copy, Check } from 'lucide-react';
import { useChallengeSession } from '@/lib/hooks/useChallenges';
import { startChallenge } from '@/lib/actions/challenges';
import { toast } from '@/lib/hooks/useToast';
import { ChallengeSession, ChallengeParticipant } from '@/types';
import { formatDistanceToNow } from 'date-fns';
import { fr } from 'date-fns/locale';
import { cn } from '@/lib/utils';
import { InviteFriendsModal } from './InviteFriendsModal';

export function ChallengeWaitingRoom({ sessionId, onStart }: { sessionId: string; onStart: () => void }) {
  const { session, loading, refetch } = useChallengeSession(sessionId);
  const [starting, setStarting] = useState(false);
  const [showInvite, setShowInvite] = useState(false);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    const interval = setInterval(refetch, 5000);
    return () => clearInterval(interval);
  }, [refetch]);

  if (loading || !session) {
    return (
      <div className="flex justify-center py-12">
        <Loader2 className="h-8 w-8 animate-spin" />
      </div>
    );
  }

  const participants = session.participants || [];
  const invitations = session.invitations || [];
  const acceptedCount = participants.filter(p => p.status === 'accepted').length;
  const canStart = acceptedCount >= 2;
  const isCreator = session.creator_id === participants.find(p => p.user_id === session.creator_id)?.user_id;

  const handleStart = async () => {
    try {
      setStarting(true);
      await startChallenge(sessionId);
      toast({ title: 'Défi lancé !' });
      onStart();
    } catch (error: any) {
      toast({ title: 'Erreur', description: error.message, variant: 'destructive' });
    } finally {
      setStarting(false);
    }
  };

  const copyInviteLink = () => {
    const link = `${window.location.origin}/challenge/${sessionId}`;
    navigator.clipboard.writeText(link);
    setCopied(true);
    toast({ title: 'Lien copié !' });
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="max-w-2xl mx-auto space-y-6">
      <Card className="border-purple-500/30 bg-gradient-to-br from-purple-500/5 to-pink-500/5">
        <CardHeader>
          <CardTitle className="flex items-center gap-3">
            <Swords className="h-6 w-6 text-purple-500" />
            Salle d'attente du défi
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="flex items-center justify-between p-4 rounded-lg bg-dark-surface">
            <div>
              <p className="text-sm text-muted-foreground">Quiz</p>
              <p className="font-medium">{session.quiz?.title}</p>
            </div>
            <div className="text-right">
              <p className="text-sm text-muted-foreground">Expire dans</p>
              <p className="font-medium flex items-center gap-1">
                <Clock className="h-4 w-4" />
                {formatDistanceToNow(new Date(session.invite_expires_at), { locale: fr })}
              </p>
            </div>
          </div>

          <div className="flex items-center justify-between p-4 rounded-lg bg-dark-surface">
            <div className="flex items-center gap-2">
              <Zap className="h-5 w-5 text-yellow-500" />
              <span className="font-medium">Pool XP total</span>
            </div>
            <span className="text-2xl font-bold text-yellow-500">{session.total_xp_pool} XP</span>
          </div>

          <div>
            <div className="flex items-center justify-between mb-3">
              <h3 className="font-medium flex items-center gap-2">
                <Users className="h-4 w-4" />
                Participants ({acceptedCount}/2 minimum)
              </h3>
              <Badge variant={canStart ? 'default' : 'secondary'}>
                {canStart ? 'Prêt' : 'En attente'}
              </Badge>
            </div>

            <div className="space-y-2">
              {participants.map((participant) => (
                <ParticipantCard key={participant.id} participant={participant} />
              ))}

              {invitations.filter(i => i.status === 'pending').map((invitation) => (
                <div
                  key={invitation.id}
                  className="flex items-center gap-3 p-3 rounded-lg border border-dashed border-muted-foreground/30"
                >
                  <Avatar className="h-10 w-10">
                    <AvatarImage src={invitation.invitee?.avatar_url || undefined} />
                    <AvatarFallback>{invitation.invitee?.username?.[0]?.toUpperCase()}</AvatarFallback>
                  </Avatar>
                  <div className="flex-1">
                    <p className="font-medium">{invitation.invitee?.username}</p>
                    <p className="text-xs text-muted-foreground">Invitation envoyée</p>
                  </div>
                  <Badge variant="outline">En attente</Badge>
                </div>
              ))}
            </div>
          </div>

          <div className="flex gap-3">
            <Button
              variant="outline"
              onClick={() => setShowInvite(true)}
              className="flex-1 gap-2"
            >
              <UserPlus className="h-4 w-4" />
              Inviter des amis
            </Button>
            <Button
              variant="outline"
              onClick={copyInviteLink}
              className="gap-2"
            >
              {copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
              {copied ? 'Copié' : 'Lien'}
            </Button>
          </div>

          {isCreator && (
            <Button
              onClick={handleStart}
              disabled={!canStart || starting}
              className="w-full gap-2"
              size="lg"
            >
              {starting ? (
                <Loader2 className="h-5 w-5 animate-spin" />
              ) : (
                <Swords className="h-5 w-5" />
              )}
              {canStart ? 'Lancer le défi' : `En attente de ${2 - acceptedCount} joueur(s)`}
            </Button>
          )}
        </CardContent>
      </Card>

      {showInvite && (
        <InviteFriendsModal
          sessionId={sessionId}
          onClose={() => { setShowInvite(false); refetch(); }}
        />
      )}
    </div>
  );
}

function ParticipantCard({ participant }: { participant: ChallengeParticipant }) {
  const user = participant.user;

  return (
    <div className={cn(
      'flex items-center gap-3 p-3 rounded-lg border',
      participant.status === 'accepted' ? 'border-green-500/30 bg-green-500/5' : 'border-muted'
    )}>
      <Avatar className="h-10 w-10">
        <AvatarImage src={user?.avatar_url || undefined} />
        <AvatarFallback>{user?.username?.[0]?.toUpperCase()}</AvatarFallback>
      </Avatar>
      <div className="flex-1">
        <p className="font-medium">{user?.username}</p>
        <div className="flex items-center gap-2 text-xs text-muted-foreground">
          <Zap className="h-3 w-3 text-yellow-500" />
          <span>Mise: {participant.xp_bet} XP</span>
        </div>
      </div>
      <Badge variant={participant.status === 'accepted' ? 'default' : 'secondary'}>
        {participant.status === 'accepted' ? 'Prêt' : participant.status}
      </Badge>
    </div>
  );
}