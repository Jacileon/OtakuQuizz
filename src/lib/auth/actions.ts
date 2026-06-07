'use server';

// ============================================================
// SERVER ACTIONS - AUTHENTIFICATION
// ============================================================

import { redirect } from '../../../node_modules/next/navigation';
import { revalidatePath } from '../../../node_modules/next/cache';
import { createClient, createAdminClient } from '@/lib/supabase/server';
import { UserProfile } from '@/types';

export async function signInWithGoogle() {
  const supabase = createClient();
  const { data, error } = await supabase.auth.signInWithOAuth({
    provider: 'google',
    options: {
      redirectTo: `${process.env.NEXT_PUBLIC_APP_URL}/auth/callback`,
    },
  });

  if (error) {
    throw new Error(`Erreur OAuth: ${error.message}`);
  }

  if (data.url) {
    redirect(data.url);
  }
}

export async function signOut() {
  const supabase = createClient();
  await supabase.auth.signOut();
  revalidatePath('/', 'layout');
  redirect('/');
}

export async function getCurrentUser() {
  const supabase = createClient();

  // Essayer getUser() d'abord (appel API)
  const { data: { user }, error } = await supabase.auth.getUser();
  if (user) {
    console.log('🟢 getCurrentUser: via getUser()', user.id);
    return user;
  }

  // Fallback: getSession() depuis les cookies
  const { data: { session } } = await supabase.auth.getSession();
  if (session?.user) {
    console.log('🟢 getCurrentUser: via getSession()', session.user.id);
    return session.user;
  }

  console.error('🔴 getCurrentUser: aucun moyen de récupérer l\'utilisateur');
  return null;
}

export async function getCurrentProfile(): Promise<UserProfile | null> {
  const supabase = createClient();
  const { data: { session }, error: authError } = await supabase.auth.getSession();

  if (authError || !session?.user) {
    return null;
  }

  const user = session.user;

  const { data: profile, error: profileError } = await supabase
    .from('user_profiles')
    .select('*')
    .eq('id', user.id)
    .single();

  if (profileError || !profile) {
    return null;
  }

  return profile as UserProfile;
}

export async function isAdmin(): Promise<boolean> {
  const supabase = createClient();
  const { data: { user } } = await supabase.auth.getUser();

  if (!user) return false;

  return user.user_metadata?.role === 'admin' || user.app_metadata?.role === 'admin';
}

export async function requireAuth() {
  const user = await getCurrentUser();
  if (!user) {
    redirect('/login');
  }
  return user;
}

export async function requireAdmin() {
  const user = await getCurrentUser();
  if (!user) {
    redirect('/login');
  }

  const isUserAdmin = await isAdmin();
  if (!isUserAdmin) {
    redirect('/dashboard');
  }

  return user;
}

export async function updateProfile(formData: {
  username?: string;
  bio?: string;
  country?: string;
  phone?: string;
  favorite_anime?: string;
}) {
  const supabase = createClient();
  const { data: { session }, error: authError } = await supabase.auth.getSession();

  if (authError || !session?.user) {
    return { success: false, error: 'Non authentifié' };
  }

  const user = session.user;

  if (formData.username) {
    const { data: existing } = await supabase
      .from('user_profiles')
      .select('id')
      .eq('username', formData.username)
      .neq('id', user.id)
      .maybeSingle();

    if (existing) {
      return { success: false, error: 'Ce username est déjà pris' };
    }

    if (formData.username.length < 3 || formData.username.length > 30) {
      return { success: false, error: 'Le username doit avoir entre 3 et 30 caractères' };
    }
    if (!/^[a-zA-Z0-9_]+$/.test(formData.username)) {
      return { success: false, error: 'Le username ne peut contenir que des lettres, chiffres et underscores' };
    }
  }

  const updates: any = {};
  if (formData.username !== undefined) updates.username = formData.username;
  if (formData.bio !== undefined) updates.bio = formData.bio || null;
  if (formData.country !== undefined) updates.country = formData.country || null;
  if (formData.phone !== undefined) updates.phone = formData.phone || null;
  if (formData.favorite_anime !== undefined) updates.favorite_anime = formData.favorite_anime || null;

  const { error } = await supabase
    .from('user_profiles')
    .update(updates)
    .eq('id', user.id);

  if (error) {
    console.error('Erreur update profile:', error);
    return { success: false, error: 'Erreur lors de la mise à jour' };
  }

  revalidatePath('/profile');
  revalidatePath('/profile/edit');
  return { success: true, error: null };
}

