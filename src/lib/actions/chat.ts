'use server';

import { createClient } from '@/lib/supabase/server';
import { Conversation, Message, UserProfile } from '@/types';
import { revalidatePath } from 'next/cache';

export async function getOrCreateConversation(friendId: string): Promise<string> {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) throw new Error('Non authentifié');

  const [smallerId, biggerId] = [user.id, friendId].sort() as [string, string];

  const { data: existing } = await supabase
    .from('conversations')
    .select('id')
    .eq('user1_id', smallerId)
    .eq('user2_id', biggerId)
    .single();

  if (existing) return existing.id;

  const { data: created, error } = await supabase
    .from('conversations')
    .insert({ user1_id: smallerId, user2_id: biggerId })
    .select('id')
    .single();

  if (error) throw new Error('Erreur création conversation');
  return created.id;
}

export async function getConversations(): Promise<(Conversation & { other_user: UserProfile; last_message?: Message; unread_count: number })[]> {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) return [];

  const { data: conversations } = await supabase
    .from('conversations')
    .select('*')
    .or(`user1_id.eq.${user.id},user2_id.eq.${user.id}`)
    .order('last_message_at', { ascending: false, nullsFirst: false });

  if (!conversations) return [];

  const result = [];
  for (const conv of conversations) {
    const otherUserId = conv.user1_id === user.id ? conv.user2_id : conv.user1_id;

    const { data: otherUser } = await supabase
      .from('user_profiles')
      .select('*')
      .eq('id', otherUserId)
      .single();

    const { data: lastMessage } = await supabase
      .from('messages')
      .select('*')
      .eq('conversation_id', conv.id)
      .order('created_at', { ascending: false })
      .limit(1)
      .single();

    const { count: unreadCount } = await supabase
      .from('messages')
      .select('*', { count: 'exact', head: true })
      .eq('conversation_id', conv.id)
      .eq('is_read', false)
      .neq('sender_id', user.id);

    result.push({
      ...conv,
      other_user: otherUser as UserProfile,
      last_message: lastMessage as Message | undefined,
      unread_count: unreadCount || 0,
    });
  }

  return result as any;
}

export async function getMessages(conversationId: string, limit = 50): Promise<Message[]> {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) return [];

  const { data: conversation } = await supabase
    .from('conversations')
    .select('id')
    .eq('id', conversationId)
    .or(`user1_id.eq.${user.id},user2_id.eq.${user.id}`)
    .single();

  if (!conversation) return [];

  const { data: messages } = await supabase
    .from('messages')
    .select('*')
    .eq('conversation_id', conversationId)
    .order('created_at', { ascending: true })
    .limit(limit);

  await supabase
    .from('messages')
    .update({ is_read: true })
    .eq('conversation_id', conversationId)
    .neq('sender_id', user.id)
    .eq('is_read', false);

  return (messages as Message[]) || [];
}

export async function sendMessage(conversationId: string, content: string): Promise<Message> {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) throw new Error('Non authentifié');

  const { data: conversation } = await supabase
    .from('conversations')
    .select('id')
    .eq('id', conversationId)
    .or(`user1_id.eq.${user.id},user2_id.eq.${user.id}`)
    .single();

  if (!conversation) throw new Error('Conversation introuvable');

  const { data: message, error } = await supabase
    .from('messages')
    .insert({ conversation_id: conversationId, sender_id: user.id, content })
    .select('*')
    .single();

  if (error) throw new Error('Erreur envoi message');
  revalidatePath('/friends');
  return message as Message;
}

export async function getUnreadMessagesCount(): Promise<number> {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) return 0;

  const { data: conversations } = await supabase
    .from('conversations')
    .select('id')
    .or(`user1_id.eq.${user.id},user2_id.eq.${user.id}`);

  if (!conversations || conversations.length === 0) return 0;

  const convIds = conversations.map(c => c.id);

  const { count } = await supabase
    .from('messages')
    .select('*', { count: 'exact', head: true })
    .in('conversation_id', convIds)
    .eq('is_read', false)
    .neq('sender_id', user.id);

  return count || 0;
}

export async function deleteMessages(conversationId: string, messageIds: string[]): Promise<void> {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) throw new Error('Non authentifié');

  const { data: conversation } = await supabase
    .from('conversations')
    .select('id')
    .eq('id', conversationId)
    .or(`user1_id.eq.${user.id},user2_id.eq.${user.id}`)
    .single();

  if (!conversation) throw new Error('Conversation introuvable');

  const { error } = await supabase
    .from('messages')
    .delete()
    .in('id', messageIds)
    .eq('conversation_id', conversationId);

  if (error) throw new Error('Erreur suppression messages');
  revalidatePath('/friends');
}

export async function deleteConversation(conversationId: string): Promise<void> {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) throw new Error('Non authentifié');

  const { data: conversation } = await supabase
    .from('conversations')
    .select('id')
    .eq('id', conversationId)
    .or(`user1_id.eq.${user.id},user2_id.eq.${user.id}`)
    .single();

  if (!conversation) throw new Error('Conversation introuvable');

  await supabase
    .from('messages')
    .delete()
    .eq('conversation_id', conversationId);

  const { error } = await supabase
    .from('conversations')
    .delete()
    .eq('id', conversationId);

  if (error) throw new Error('Erreur suppression conversation');
  revalidatePath('/friends');
}