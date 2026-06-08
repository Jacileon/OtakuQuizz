'use server';

import { createClient } from '@/lib/supabase/server';
import { ChallengeSession, ChallengeParticipant, ChallengeInvitation } from '@/types';
import { revalidatePath } from 'next/cache';

export async function createChallengeSession(quizId: string, xpBet: number): Promise<string> {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) throw new Error('Non authentifié');

  const { data: profile } = await supabase
    .from('user_profiles')
    .select('xp')
    .eq('id', user.id)
    .single();

  if (!profile || profile.xp < xpBet) {
    throw new Error('Solde XP insuffisant');
  }

  const { data: sessions } = await supabase
    .from('challenge_sessions')
    .select('id')
    .eq('quiz_id', quizId);

  const sessionIds = sessions?.map(s => s.id) || [];

  const { count } = sessionIds.length > 0 ? await supabase
    .from('challenge_participants')
    .select('*', { count: 'exact', head: true })
    .eq('user_id', user.id)
    .in('session_id', sessionIds) : { count: 0 };

  if (count && count >= 3) {
    throw new Error('Limite de 3 participations atteinte pour ce quiz');
  }

  const expiresAt = new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString();

  const { data: session, error } = await supabase
    .from('challenge_sessions')
    .insert({
      quiz_id: quizId,
      creator_id: user.id,
      invite_expires_at: expiresAt,
      status: 'waiting',
    })
    .select('id')
    .single();

  if (error) throw new Error('Erreur création défi');

  await supabase.from('challenge_participants').insert({
    session_id: session.id,
    user_id: user.id,
    xp_bet: xpBet,
    status: 'accepted',
  });

  revalidatePath('/challenges');
  return session.id;
}

export async function inviteToChallenge(sessionId: string, friendId: string): Promise<void> {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) throw new Error('Non authentifié');

  const { data: session } = await supabase
    .from('challenge_sessions')
    .select('id, creator_id, status, invite_expires_at')
    .eq('id', sessionId)
    .single();

  if (!session || session.creator_id !== user.id) throw new Error('Non autorisé');
  if (session.status !== 'waiting') throw new Error('Défi non disponible');

  const { data: existing } = await supabase
    .from('challenge_invitations')
    .select('id')
    .eq('session_id', sessionId)
    .eq('invitee_id', friendId)
    .single();

  if (existing) throw new Error('Invitation déjà envoyée');

  const { error } = await supabase.from('challenge_invitations').insert({
    session_id: sessionId,
    inviter_id: user.id,
    invitee_id: friendId,
    expires_at: session.invite_expires_at,
  });

  if (error) throw new Error('Erreur envoi invitation');

  await supabase.from('notifications').insert({
    user_id: friendId,
    type: 'friend_request',
    title: 'Défi reçu',
    message: `Vous avez été défié ! XP mis en jeu: ${0}`,
    data: { session_id: sessionId, inviter_id: user.id },
  });

  revalidatePath('/challenges');
}

export async function acceptChallengeInvitation(invitationId: string, xpBet: number): Promise<void> {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) throw new Error('Non authentifié');

  const { data: invitation } = await supabase
    .from('challenge_invitations')
    .select('*, session:session_id(*)')
    .eq('id', invitationId)
    .eq('invitee_id', user.id)
    .single();

  if (!invitation) throw new Error('Invitation introuvable');
  if (invitation.status !== 'pending') throw new Error('Invitation non disponible');

  const { data: profile } = await supabase
    .from('user_profiles')
    .select('xp')
    .eq('id', user.id)
    .single();

  if (!profile || profile.xp < xpBet) {
    throw new Error('Solde XP insuffisant');
  }

  await supabase
    .from('challenge_invitations')
    .update({ status: 'accepted' })
    .eq('id', invitationId);

  await supabase.from('challenge_participants').insert({
    session_id: invitation.session_id,
    user_id: user.id,
    xp_bet: xpBet,
    status: 'accepted',
  });

  const { count } = await supabase
    .from('challenge_participants')
    .select('*', { count: 'exact', head: true })
    .eq('session_id', invitation.session_id)
    .eq('status', 'accepted');

  if (count && count >= 2) {
    await supabase
      .from('challenge_sessions')
      .update({ status: 'ready' })
      .eq('id', invitation.session_id);
  }

  revalidatePath('/challenges');
}

export async function refuseChallengeInvitation(invitationId: string): Promise<void> {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) throw new Error('Non authentifié');

  const { data: invitation } = await supabase
    .from('challenge_invitations')
    .select('id, inviter_id')
    .eq('id', invitationId)
    .eq('invitee_id', user.id)
    .single();

  if (!invitation) throw new Error('Invitation introuvable');

  await supabase
    .from('challenge_invitations')
    .update({ status: 'refused' })
    .eq('id', invitationId);

  await supabase.from('notifications').insert({
    user_id: invitation.inviter_id,
    type: 'friend_request',
    title: 'Défi refusé',
    message: 'Votre invitation a été refusée',
    data: { invitation_id: invitationId },
  });

  revalidatePath('/challenges');
}

export async function getMyChallenges(): Promise<ChallengeSession[]> {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) return [];

  const { data: participations } = await supabase
    .from('challenge_participants')
    .select('session_id')
    .eq('user_id', user.id);

  if (!participations || participations.length === 0) return [];

  const sessionIds = participations.map(p => p.session_id);

  const { data: sessions } = await supabase
    .from('challenge_sessions')
    .select('*, quiz:quiz_id(*), participants:challenge_participants(*, user:user_id(*))')
    .in('id', sessionIds)
    .order('created_at', { ascending: false });

  return (sessions as any[]) || [];
}

export async function getChallengeSession(sessionId: string): Promise<ChallengeSession | null> {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) return null;

  const { data: session } = await supabase
    .from('challenge_sessions')
    .select('*, quiz:quiz_id(*), creator:creator_id(*), participants:challenge_participants(*, user:user_id(*)), invitations:challenge_invitations(*, invitee:invitee_id(*))')
    .eq('id', sessionId)
    .single();

  if (!session) return null;

  const isParticipant = (session as any).participants?.some((p: any) => p.user_id === user.id);
  if (!isParticipant && (session as any).creator_id !== user.id) return null;

  return session as any;
}

export async function getChallengeParticipationCount(quizId: string): Promise<number> {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) return 0;

  const { data: sessions } = await supabase
    .from('challenge_sessions')
    .select('id')
    .eq('quiz_id', quizId);

  const sessionIds = sessions?.map(s => s.id) || [];

  const { count } = sessionIds.length > 0 ? await supabase
    .from('challenge_participants')
    .select('*', { count: 'exact', head: true })
    .eq('user_id', user.id)
    .in('session_id', sessionIds) : { count: 0 };

  return count || 0;
}

export async function startChallenge(sessionId: string): Promise<void> {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) throw new Error('Non authentifié');

  const { data: session } = await supabase
    .from('challenge_sessions')
    .select('id, creator_id, status')
    .eq('id', sessionId)
    .single();

  if (!session || session.creator_id !== user.id) throw new Error('Non autorisé');
  if (session.status !== 'ready') throw new Error('Défi non prêt');

  await supabase
    .from('challenge_sessions')
    .update({ status: 'playing', started_at: new Date().toISOString() })
    .eq('id', sessionId);

  await supabase
    .from('challenge_participants')
    .update({ status: 'playing' })
    .eq('session_id', sessionId)
    .eq('status', 'accepted');

  revalidatePath('/challenges');
}

export async function completeChallenge(sessionId: string): Promise<void> {
  const supabase = await createClient();

  const { data: session } = await supabase
    .from('challenge_sessions')
    .select('id, total_xp_pool')
    .eq('id', sessionId)
    .single();

  if (!session) return;

  const { data: participants } = await supabase
    .from('challenge_participants')
    .select('*')
    .eq('session_id', sessionId)
    .eq('status', 'playing')
    .order('score', { ascending: false });

  if (!participants || participants.length === 0) return;

  const maxScore = participants[0].score;
  const winners = participants.filter(p => p.score === maxScore);
  const xpPerWinner = Math.floor(session.total_xp_pool / winners.length);

  for (const winner of winners) {
    await supabase
      .from('challenge_participants')
      .update({ 
        status: 'done', 
        completed_at: new Date().toISOString(),
        xp_won: xpPerWinner 
      })
      .eq('id', winner.id);

    await supabase
      .from('user_profiles')
      .update({ xp: supabase.rpc('increment_xp', { amount: xpPerWinner }) })
      .eq('id', winner.user_id);

    await supabase.from('xp_ledger').insert({
      user_id: winner.user_id,
      amount: xpPerWinner,
      type: 'won',
      reference_type: 'challenge',
      reference_id: sessionId,
    });
  }

  for (const loser of participants.filter(p => p.score < maxScore)) {
    await supabase
      .from('challenge_participants')
      .update({ 
        status: 'done', 
        completed_at: new Date().toISOString(),
        xp_lost: loser.xp_bet 
      })
      .eq('id', loser.id);
  }

  await supabase
    .from('challenge_sessions')
    .update({ 
      status: 'completed', 
      completed_at: new Date().toISOString(),
      winner_id: winners[0].user_id 
    })
    .eq('id', sessionId);

  revalidatePath('/challenges');
}