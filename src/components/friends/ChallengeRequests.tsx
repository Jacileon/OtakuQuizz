'use client';

import { useState, useEffect } from 'react';
import { Card, CardContent } from '@/components/ui/card';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Swords, Loader2, Zap, Clock, Check, X } from 'lucide-react';
import { useAuth } from '@/components/providers/AuthProvider';
import { createClient } from '@/lib/supabase/client';
import { toast } from '@/lib/hooks/useToast';
import { formatDistanceToNow } from 'date-fns';
import { fr } from 'date-fns/locale';
import Link from 'next/link';

export function ChallengeRequests() {
  const { user } = useAuth();
  const [invitations, setInvitations] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchInvitations = async () => {
    if (!user) return;
    const supabase = createClient();
    const { data } = await supabase
      .from('challenge_invitations')
      .select('*, inviter:inviter_id(*), session:session_id(*, quiz:quiz_id(title))')
      .eq('invitee_id', user.id)
      .eq('status', 'pending')
      .order('created_at', { ascending: false });
    
    setInvitations(data || []);
    setLoading(false);
  };

  useEffect(() => {
    if (user) {
      fetchInvitations();
    } else {
      setLoading(false);
    }
  }, [user]);

  const handleAccept = async (invitationId: string) => {
    try {
      const supabase = createClient();
      await supabase
        .from('challenge_invitations')
        .update({ status: 'accepted' })
        .eq('id', invitationId);
      
      toast({ title: 'Défi accepté !' });
      fetchInvitations();
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

  if (loading) {
    return (
      <div className="flex justify-center py-8">
        <Loader2 className="h-6 w-6 animate-spin text-primary" />
      </div>
    );
  }

  if (invitations.length === 0) {
    return (
      <Card>
        <CardContent className="flex flex-col items-center justify-center py-12">
          <Swords className="h-12 w-12 text-muted-foreground mb-4" />
          <p className="text-muted-foreground">Aucune demande de défi</p>
        </CardContent>
      </Card>
    );
  }

  return (
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
  );
}