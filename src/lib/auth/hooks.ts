'use client';

import { useEffect, useState } from 'react';
import { useRouter } from '../../../node_modules/next/navigation';
import { User } from '@supabase/supabase-js';
import { UserProfile } from '@/types';
import { getBrowserClient } from '@/lib/supabase/client';

export function useUser() {
  const [user, setUser] = useState<User | null>(null);
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [loading, setLoading] = useState(true);
  const [debugInfo, setDebugInfo] = useState<string>('');

  useEffect(() => {
    const supabase = getBrowserClient();
    console.log('[useUser] Hook initialized');

    const getUser = async () => {
      try {
        console.log('[useUser] Getting session...');
        const { data: { session }, error: sessionError } = await supabase.auth.getSession();
        
        console.log('[useUser] Session result:', {
          hasSession: !!session,
          hasUser: !!session?.user,
          error: sessionError?.message
        });

        if (session?.user) {
          console.log('[useUser] User found from session:', session.user.id);
          setUser(session.user);
          
          const { data, error: profileError } = await supabase
            .from('user_profiles')
            .select('*')
            .eq('id', session.user.id)
            .single();
          
          if (profileError) {
            console.error('[useUser] Profile error:', profileError);
          }
          setProfile(data as UserProfile | null);
        } else {
          console.log('[useUser] No session, trying getUser...');
          const { data: { user }, error: userError } = await supabase.auth.getUser();
          
          console.log('[useUser] getUser result:', {
            hasUser: !!user,
            error: userError?.message
          });

          if (user) {
            setUser(user);
            const { data } = await supabase
              .from('user_profiles')
              .select('*')
              .eq('id', user.id)
              .single();
            setProfile(data as UserProfile | null);
          } else {
            setDebugInfo('No session found. Please login again.');
          }
        }
      } catch (error) {
        console.error('[useUser] Error:', error);
        setDebugInfo(`Error: ${error}`);
      } finally {
        setLoading(false);
      }
    };

    getUser();

    const { data: { subscription } } = supabase.auth.onAuthStateChange(
      async (event: string, session: any) => {
        console.log('[useUser] Auth state changed:', event, session?.user?.id);
        if (event === 'SIGNED_IN' && session?.user) {
          setUser(session.user);
          const { data } = await supabase
            .from('user_profiles')
            .select('*')
            .eq('id', session.user.id)
            .single();
          setProfile(data as UserProfile | null);
        } else if (event === 'SIGNED_OUT') {
          setUser(null);
          setProfile(null);
        }
      }
    );

    return () => subscription.unsubscribe();
  }, []);

  return { user, profile, loading, debugInfo };
}

export function useRequireAuth() {
  const { user, loading } = useUser();
  const router = useRouter();

  useEffect(() => {
    if (!loading && !user) {
      router.push('/login');
    }
  }, [user, loading, router]);

  return { user, loading };
}

export function useIsAdmin() {
  const { user, loading } = useUser();
  const [isAdmin, setIsAdmin] = useState(false);

  useEffect(() => {
    if (user) {
      setIsAdmin(user.user_metadata?.role === 'admin' || user.app_metadata?.role === 'admin');
    }
  }, [user]);

  return { isAdmin, loading };
}
