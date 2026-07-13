'use server';

import { createClient } from '@/lib/supabase/server';
import { Announcement } from '@/types';
import { revalidatePath } from 'next/cache';

export async function createAnnouncement(data: {
  title: string;
  description?: string;
  image_url?: string;
  quiz_id?: string;
  type: 'quiz' | 'event' | 'news';
  starts_at?: string;
  ends_at?: string;
}): Promise<string> {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) throw new Error('Non authentifié');

  const { data: profile } = await supabase
    .from('user_profiles')
    .select('is_admin')
    .eq('id', user.id)
    .single();

  if (!profile?.is_admin) throw new Error('Non autorisé');

  const { data: announcement, error } = await supabase
    .from('announcements')
    .insert({
      title: data.title,
      description: data.description,
      image_url: data.image_url,
      quiz_id: data.quiz_id || null,
      type: data.type,
      status: 'active',
      starts_at: data.starts_at || new Date().toISOString(),
      ends_at: data.ends_at || null,
      created_by: user.id,
    })
    .select('id')
    .single();

  if (error) throw new Error('Erreur création annonce');
  
  revalidatePath('/events');
  revalidatePath('/admin/announcements');
  return announcement.id;
}

export async function getActiveAnnouncements(): Promise<Announcement[]> {
  const supabase = await createClient();

  const { data } = await supabase
    .from('announcements')
    .select('*, quiz:quiz_id(*)')
    .eq('status', 'active')
    .or(`ends_at.is.null,ends_at.gt.${new Date().toISOString()}`)
    .order('created_at', { ascending: false });

  return (data as Announcement[]) || [];
}

export async function getAllAnnouncements(): Promise<Announcement[]> {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) return [];

  const { data: profile } = await supabase
    .from('user_profiles')
    .select('is_admin')
    .eq('id', user.id)
    .single();

  if (!profile?.is_admin) return [];

  const { data } = await supabase
    .from('announcements')
    .select('*, quiz:quiz_id(*)')
    .order('created_at', { ascending: false });

  return (data as Announcement[]) || [];
}

export async function updateAnnouncement(id: string, data: {
  title?: string;
  description?: string;
  image_url?: string;
  quiz_id?: string;
  status?: string;
  ends_at?: string;
}): Promise<void> {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) throw new Error('Non authentifié');

  const { data: profile } = await supabase
    .from('user_profiles')
    .select('is_admin')
    .eq('id', user.id)
    .single();

  if (!profile?.is_admin) throw new Error('Non autorisé');

  const updates: any = {};
  if (data.title) updates.title = data.title;
  if (data.description !== undefined) updates.description = data.description;
  if (data.image_url !== undefined) updates.image_url = data.image_url;
  if (data.quiz_id !== undefined) updates.quiz_id = data.quiz_id || null;
  if (data.status) updates.status = data.status;
  if (data.ends_at !== undefined) updates.ends_at = data.ends_at || null;

  const { error } = await supabase
    .from('announcements')
    .update(updates)
    .eq('id', id);

  if (error) throw new Error('Erreur mise à jour annonce');
  
  revalidatePath('/events');
  revalidatePath('/admin/announcements');
}

export async function deleteAnnouncement(id: string): Promise<void> {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  if (!user) throw new Error('Non authentifié');

  const { data: profile } = await supabase
    .from('user_profiles')
    .select('is_admin')
    .eq('id', user.id)
    .single();

  if (!profile?.is_admin) throw new Error('Non autorisé');

  const { error } = await supabase
    .from('announcements')
    .delete()
    .eq('id', id);

  if (error) throw new Error('Erreur suppression annonce');
  
  revalidatePath('/events');
  revalidatePath('/admin/announcements');
}