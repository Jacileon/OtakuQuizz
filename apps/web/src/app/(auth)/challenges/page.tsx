'use client';

import { useMyChallenges } from '@/lib/hooks/useChallenges';
import { useAuth } from '@/components/providers/AuthProvider';
import { createClient } from '@/lib/supabase/client';
import { useState, useEffect } from 'react';
import { Card, CardContent } from '@/components/ui/card';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Swords, Loader2, Zap, Clock, Check, X, Users } from 'lucide-react';
import { toast } from '@/lib/hooks/useToast';
import { formatDistanceToNow } from 'date-fns';
import { fr } from 'date-fns/locale';
import Link from 'next/link';

export default function ChallengesPage() {
  const { user } = useAuth();
  const [invitations, setInvitations] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const { challenges, loading: loadingChallenges, refetch } = useMyChallenges();

  const fetchInvitations = async () => {
    if (!user) return;
    const supabase = createClient();
    const { data: invs } = await supabase
      .from('challenge_invitations')
      .select('*, inviter:inviter_id(*), session:session_id(*, quiz:quiz_id(title))')
      .eq('invitee_id', user.id)
      .eq('status', 'pending')
      .order('created_at', { ascending: false });
    setInvitations(invs || []);
    setLoading(false);
  };

  useEffect(() => {
    fetchInvitations();
  }, [user]);

  const handleAccept = async (invitationId: string) => {
    try {
      const supabase = createClient();
      const inv = invitations.find(i => i.id === invitationId);
      if (!inv) return;
      await supabase
        .from('challenge_invitations')
        .update({ status: 'accepted' })
        .eq('id', invitationId);
      await supabase.from('challenge_participants').insert({
        session_id: inv.session_id,
        user_id: user?.id,
        xp_bet: 0,
        status: 'accepted',
      });
      toast({ title: 'Défi accepté !' });
      fetchInvitations();
      refetch();
    } catch (error: any) {
      toast({ title: 'Erreur', description: error.message, variant: 'destructive' });
    }
  };

  const handleReject = async (invitationId: string) => {
    try {
      const supabase = createClient();
      await supabase
        .from('challenge_invitations')
        .update({ status: 'refused' })
        .eq('id', invitationId);
      toast({ title: 'Défi refusé' });
      fetchInvitations();
    } catch (error: any) {
      toast({ title: 'Erreur', description: error.message, variant: 'destructive' });
    }
  };

  if (loading || loadingChallenges) {
    return (
      <div className="flex items-center justify-center min-h-[60vh]">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  return (
    <div className="container max-w-2xl mx-auto py-8 px-4">
      <div className="mb-8">
        <h1 className="text-3xl font-bold flex items-center gap-3">
          <Swords className="h-8 w-8 text-purple-500" />
          Défis
        </h1>
        <p className="text-muted-foreground mt-2">
          Gérez vos défis et invitations
        </p>
      </div>

      <div className="space-y-6">
        {invitations.length > 0 && (
          <div>
            <h3 className="text-sm font-semibold text-muted-foreground mb-3">
              Demandes de défi reçues ({invitations.length})
            </h3>
            <div className="space-y-3">
              {invitations.map((inv) => (
                <Card key={inv.id} className="border-purple-500/20">
                  <CardContent className="p-4">
                    <div className="flex items-center gap-3">
                      <Avatar>
                        <AvatarImage src={inv.inviter?.avatar_url || undefined} />
                        <AvatarFallback>{inv.inviter?.username?.[0]?.toUpperCase()}</AvatarFallback>
                      </Avatar>
                      <div className="flex-1 min-w-0">
                        <p className="font-semibold">{inv.inviter?.nickname || inv.inviter?.username}</p>
                        <p className="text-sm text-muted-foreground truncate">
                          {inv.session?.quiz?.title || 'Quiz'}
                        </p>
                        <div className="flex items-center gap-2 mt-1">
                          <Clock className="h-3 w-3 text-muted-foreground" />
                          <span className="text-xs text-muted-foreground">
                            {formatDistanceToNow(new Date(inv.created_at), { addSuffix: true, locale: fr })}
                          </span>
                        </div>
                      </div>
                      <div className="flex gap-2">
                        <Button size="sm" onClick={() => handleAccept(inv.id)}>
                          <Check className="h-4 w-4 mr-1" />
                          Accepter
                        </Button>
                        <Button size="sm" variant="outline" onClick={() => handleReject(inv.id)}>
                          <X className="h-4 w-4" />
                        </Button>
                      </div>
                    </div>
                    <Link href={`/challenges/${inv.session_id}`}>
                      <Button variant="ghost" size="sm" className="w-full mt-3 gap-2">
                        <Swords className="h-4 w-4" />
                        Voir le défi
                      </Button>
                    </Link>
                  </CardContent>
                </Card>
              ))}
            </div>
          </div>
        )}

        {challenges.length > 0 && (
          <div>
            <h3 className="text-sm font-semibold text-muted-foreground mb-3">
              Mes défis ({challenges.length})
            </h3>
            <div className="space-y-3">
              {challenges.map((challenge: any) => (
                <Link key={challenge.id} href={`/challenges/${challenge.id}`}>
                  <Card className="hover:border-purple-500/30 transition-all cursor-pointer">
                    <CardContent className="p-4">
                      <div className="flex items-center gap-3">
                        <div className="h-10 w-10 rounded-lg bg-purple-500/10 flex items-center justify-center">
                          <Swords className="h-5 w-5 text-purple-500" />
                        </div>
                        <div className="flex-1 min-w-0">
                          <p className="font-medium truncate">{challenge.quiz?.title || 'Quiz'}</p>
                          <div className="flex items-center gap-2 mt-1">
                            <Users className="h-3 w-3 text-muted-foreground" />
                            <span className="text-xs text-muted-foreground">
                              {challenge.participants?.length || 0} participant(s)
                            </span>
                            <Zap className="h-3 w-3 text-yellow-500" />
                            <span className="text-xs text-muted-foreground">
                              {challenge.total_xp_pool || 0} XP
                            </span>
                          </div>
                        </div>
                        <Badge variant={
                          challenge.status === 'waiting' ? 'secondary' :
                          challenge.status === 'completed' ? 'default' : 'outline'
                        }>
                          {challenge.status === 'waiting' ? 'En attente' :
                           challenge.status === 'ready' ? 'Prêt' :
                           challenge.status === 'playing' ? 'En cours' :
                           challenge.status === 'completed' ? 'Terminé' : challenge.status}
                        </Badge>
                      </div>
                    </CardContent>
                  </Card>
                </Link>
              ))}
            </div>
          </div>
        )}

        {invitations.length === 0 && challenges.length === 0 && (
          <Card>
            <CardContent className="flex flex-col items-center justify-center py-12">
              <Swords className="h-12 w-12 text-muted-foreground mb-4" />
              <p className="text-muted-foreground">Aucun défi</p>
              <p className="text-sm text-muted-foreground">Créez un défi depuis la page d'un quiz</p>
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  );
}
