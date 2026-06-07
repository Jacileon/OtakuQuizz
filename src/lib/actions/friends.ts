'use server';

import { createClient } from '@/lib/supabase/server';
import { Friendship, FriendshipStatus, Notification, UserProfile } from '@/types';
import { revalidatePath } from 'next/cache';

export async function sendFriendRequest(addresseeId: string) {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) throw new Error('Non authentifié');
  if (user.id === addresseeId) throw new Error('Impossible de s\'ajouter soi-même');

  const { data: existing } = await supabase
    .from('friendships')
    .select('id, status')
    .or(`and(requester_id.eq.${user.id},addressee_id.eq.${addresseeId}),and(requester_id.eq.${addresseeId},addressee_id.eq.${user.id})`)
    .single();

  if (existing) {
    if (existing.status === 'accepted') throw new Error('Déjà amis');
    if (existing.status === 'pending') throw new Error('Demande déjà envoyée');
    if (existing.status === 'rejected') {
      const { error } = await supabase
        .from('friendships')
        .update({ status: 'pending', requester_id: user.id, addressee_id: addresseeId })
        .eq('id', existing.id);
      if (error) throw new Error('Erreur lors de l\'envoi');
      revalidatePath('/friends');
      return { success: true };
    }
  }

  const { error } = await supabase
    .from('friendships')
    .insert({ requester_id: user.id, addressee_id: addresseeId, status: 'pending' });

  if (error) throw new Error('Erreur lors de l\'envoi');
  revalidatePath('/friends');
  return { success: true };
}

export async function acceptFriendRequest(friendshipId: string) {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) throw new Error('Non authentifié');

  const { data: friendship, error: fetchError } = await supabase
    .from('friendships')
    .select('id, requester_id, addressee_id')
    .eq('id', friendshipId)
    .single();

  if (fetchError || !friendship) throw new Error('Demande introuvable');
  if (friendship.addressee_id !== user.id) throw new Error('Non autorisé');

  const { error } = await supabase
    .from('friendships')
    .update({ status: 'accepted' })
    .eq('id', friendshipId);

  if (error) throw new Error('Erreur lors de l\'acceptation');

  await supabase.from('notifications').insert({
    user_id: friendship.requester_id,
    type: 'friend_request',
    title: 'Demande acceptée',
    message: `Votre demande d'ami a été acceptée`,
    data: { friendship_id: friendshipId, user_id: user.id },
  });

  revalidatePath('/friends');
  return { success: true };
}

export async function rejectFriendRequest(friendshipId: string) {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) throw new Error('Non authentifié');

  const { data: friendship, error: fetchError } = await supabase
    .from('friendships')
    .select('id, requester_id, addressee_id')
    .eq('id', friendshipId)
    .single();

  if (fetchError || !friendship) throw new Error('Demande introuvable');
  if (friendship.addressee_id !== user.id) throw new Error('Non autorisé');

  const { error } = await supabase
    .from('friendships')
    .update({ status: 'rejected' })
    .eq('id', friendshipId);

  if (error) throw new Error('Erreur lors du refus');

  await supabase.from('notifications').insert({
    user_id: friendship.requester_id,
    type: 'friend_request',
    title: 'Demande refusée',
    message: `Votre demande d'ami a été refusée`,
    data: { friendship_id: friendshipId, user_id: user.id },
  });

  revalidatePath('/friends');
  return { success: true };
}

export async function removeFriend(friendshipId: string) {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) throw new Error('Non authentifié');

  const { data: friendship, error: fetchError } = await supabase
    .from('friendships')
    .select('id, requester_id, addressee_id')
    .eq('id', friendshipId)
    .single();

  if (fetchError || !friendship) throw new Error('Amitié introuvable');
  if (friendship.requester_id !== user.id && friendship.addressee_id !== user.id) {
    throw new Error('Non autorisé');
  }

  const { error } = await supabase
    .from('friendships')
    .delete()
    .eq('id', friendshipId);

  if (error) throw new Error('Erreur lors de la suppression');
  revalidatePath('/friends');
  return { success: true };
}

export async function searchUsers(query: string): Promise<UserProfile[]> {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) return [];

  const { data } = await supabase
    .from('user_profiles')
    .select('*')
    .neq('id', user.id)
    .ilike('username', `%${query}%`)
    .limit(20);

  return (data as UserProfile[]) || [];
}

export async function getFriends(): Promise<(Friendship & { friend: UserProfile })[]> {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) return [];

  const { data } = await supabase
    .from('friendships')
    .select('*, requester:requester_id(*), addressee:addressee_id(*)')
    .or(`requester_id.eq.${user.id},addressee_id.eq.${user.id}`)
    .eq('status', 'accepted')
    .order('updated_at', { ascending: false });

  return (data || []).map((f: any) => ({
    ...f,
    friend: f.requester_id === user.id ? f.addressee : f.requester,
  }));
}

export async function getPendingRequests(): Promise<(Friendship & { requester: UserProfile })[]> {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) return [];

  const { data } = await supabase
    .from('friendships')
    .select('*, requester:requester_id(*)')
    .eq('addressee_id', user.id)
    .eq('status', 'pending')
    .order('created_at', { ascending: false });

  return (data as any[]) || [];
}

export async function getSentRequests(): Promise<(Friendship & { addressee: UserProfile })[]> {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) return [];

  const { data } = await supabase
    .from('friendships')
    .select('*, addressee:addressee_id(*)')
    .eq('requester_id', user.id)
    .eq('status', 'pending')
    .order('created_at', { ascending: false });

  return (data as any[]) || [];
}

export async function getFriendshipStatus(userId: string): Promise<{ status: FriendshipStatus | null; friendshipId: string | null; isRequester: boolean }> {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) return { status: null, friendshipId: null, isRequester: false };

  const { data } = await supabase
    .from('friendships')
    .select('id, status, requester_id')
    .or(`and(requester_id.eq.${user.id},addressee_id.eq.${userId}),and(requester_id.eq.${userId},addressee_id.eq.${user.id})`)
    .single();

  if (!data) return { status: null, friendshipId: null, isRequester: false };

  return {
    status: data.status as FriendshipStatus,
    friendshipId: data.id,
    isRequester: data.requester_id === user.id,
  };
}

export async function getRecentNotifications(): Promise<Notification[]> {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) return [];

  const { data } = await supabase
    .from('notifications')
    .select('*')
    .eq('user_id', user.id)
    .order('created_at', { ascending: false })
    .limit(20);

  return (data as Notification[]) || [];
}

export async function markNotificationAsRead(notificationId: string) {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) throw new Error('Non authentifié');

  const { error } = await supabase
    .from('notifications')
    .update({ is_read: true })
    .eq('id', notificationId)
    .eq('user_id', user.id);

  if (error) throw new Error('Erreur lors de la mise à jour');
  return { success: true };
}