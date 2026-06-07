'use server';

import { createClient } from '@/lib/supabase/server';
import { AdminConversation, AdminMessage, UserProfile } from '@/types';
import { revalidatePath } from 'next/cache';

export async function createAdminConversation(subject: string): Promise<string> {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) throw new Error('Non authentifié');

  const { data: existing } = await supabase
    .from('admin_conversations')
    .select('id')
    .eq('user_id', user.id)
    .eq('status', 'open')
    .single();

  if (existing) return existing.id;

  const { data: created, error } = await supabase
    .from('admin_conversations')
    .insert({ user_id: user.id, subject })
    .select('id')
    .single();

  if (error) throw new Error('Erreur création conversation');
  return created.id;
}

export async function getMyAdminConversations(): Promise<(AdminConversation & { last_message?: AdminMessage; unread_count: number })[]> {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) return [];

  const { data: conversations } = await supabase
    .from('admin_conversations')
    .select('*')
    .eq('user_id', user.id)
    .order('last_message_at', { ascending: false, nullsFirst: false });

  if (!conversations) return [];

  const result = [];
  for (const conv of conversations) {
    const { data: lastMessage } = await supabase
      .from('admin_messages')
      .select('*')
      .eq('conversation_id', conv.id)
      .order('created_at', { ascending: false })
      .limit(1)
      .single();

    const { count: unreadCount } = await supabase
      .from('admin_messages')
      .select('*', { count: 'exact', head: true })
      .eq('conversation_id', conv.id)
      .eq('is_read', false)
      .neq('sender_id', user.id);

    result.push({
      ...conv,
      last_message: lastMessage as AdminMessage | undefined,
      unread_count: unreadCount || 0,
    });
  }

  return result as any;
}

export async function getAllAdminConversations(): Promise<(AdminConversation & { user: UserProfile; last_message?: AdminMessage; unread_count: number })[]> {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) throw new Error('Non authentifié');

  const { data: profile } = await supabase
    .from('user_profiles')
    .select('is_admin')
    .eq('id', user.id)
    .single();

  if (!profile?.is_admin) throw new Error('Non autorisé');

  const { data: conversations } = await supabase
    .from('admin_conversations')
    .select('*')
    .order('last_message_at', { ascending: false, nullsFirst: false });

  if (!conversations) return [];

  const result = [];
  for (const conv of conversations) {
    const { data: userData } = await supabase
      .from('user_profiles')
      .select('*')
      .eq('id', conv.user_id)
      .single();

    const { data: lastMessage } = await supabase
      .from('admin_messages')
      .select('*')
      .eq('conversation_id', conv.id)
      .order('created_at', { ascending: false })
      .limit(1)
      .single();

    const { count: unreadCount } = await supabase
      .from('admin_messages')
      .select('*', { count: 'exact', head: true })
      .eq('conversation_id', conv.id)
      .eq('is_read', false)
      .neq('sender_id', user.id);

    result.push({
      ...conv,
      user: userData as UserProfile,
      last_message: lastMessage as AdminMessage | undefined,
      unread_count: unreadCount || 0,
    });
  }

  return result as any;
}

export async function getAdminMessages(conversationId: string, limit = 50): Promise<AdminMessage[]> {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) return [];

  const { data: conversation } = await supabase
    .from('admin_conversations')
    .select('id, user_id')
    .eq('id', conversationId)
    .single();

  if (!conversation) return [];

  const { data: profile } = await supabase
    .from('user_profiles')
    .select('is_admin')
    .eq('id', user.id)
    .single();

  if (conversation.user_id !== user.id && !profile?.is_admin) return [];

  const { data: messages } = await supabase
    .from('admin_messages')
    .select('*')
    .eq('conversation_id', conversationId)
    .order('created_at', { ascending: true })
    .limit(limit);

  await supabase
    .from('admin_messages')
    .update({ is_read: true })
    .eq('conversation_id', conversationId)
    .neq('sender_id', user.id)
    .eq('is_read', false);

  return (messages as AdminMessage[]) || [];
}

export async function sendAdminMessage(conversationId: string, content: string): Promise<AdminMessage> {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) throw new Error('Non authentifié');

  const { data: conversation } = await supabase
    .from('admin_conversations')
    .select('id, user_id')
    .eq('id', conversationId)
    .single();

  if (!conversation) throw new Error('Conversation introuvable');

  const { data: profile } = await supabase
    .from('user_profiles')
    .select('is_admin')
    .eq('id', user.id)
    .single();

  if (conversation.user_id !== user.id && !profile?.is_admin) throw new Error('Non autorisé');

  if (profile?.is_admin && !conversation.admin_id) {
    await supabase
      .from('admin_conversations')
      .update({ admin_id: user.id, status: 'assigned' })
      .eq('id', conversationId);
  }

  const { data: message, error } = await supabase
    .from('admin_messages')
    .insert({ conversation_id: conversationId, sender_id: user.id, content })
    .select('*')
    .single();

  if (error) throw new Error('Erreur envoi message');
  revalidatePath('/friends');
  revalidatePath('/admin/support');
  return message as AdminMessage;
}

export async function closeAdminConversation(conversationId: string): Promise<void> {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) throw new Error('Non authentifié');

  const { data: profile } = await supabase
    .from('user_profiles')
    .select('is_admin')
    .eq('id', user.id)
    .single();

  if (!profile?.is_admin) throw new Error('Non autorisé');

  await supabase
    .from('admin_conversations')
    .update({ status: 'closed' })
    .eq('id', conversationId);
}