'use client';

import { createContext, useContext, useEffect, useState, ReactNode } from "react";
import { User } from '@supabase/supabase-js';
import { UserProfile } from '@/types';
import { getBrowserClient } from '@/lib/supabase/client';

interface AuthContextType {
  user: User | null;
  profile: UserProfile | null;
  loading: boolean;
  refreshProfile: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType>({
  user: null,
  profile: null,
  loading: true,
  refreshProfile: async () => {},
});

export function AuthProvider({ children }: { children: ReactNode }) {
  console.log('[AuthProvider] ⚡ RENDER');

  const [user, setUser] = useState<User | null>(null);
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [loading, setLoading] = useState(true);

  const refreshProfile = async () => {
    console.log('[AuthProvider] refreshProfile appelé, user:', user?.id ?? 'null');
    if (!user) return;
    const supabase = getBrowserClient();
    const { data, error } = await supabase
      .from('user_profiles')
      .select('*')
      .eq('id', user.id)
      .single();
    console.log('[AuthProvider] refreshProfile résultat →', { data: !!data, error });
    setProfile(data as UserProfile | null);
  };

  useEffect(() => {
    console.log('[AuthProvider] useEffect monté');
    const supabase = getBrowserClient();

    const initializeAuth = async () => {
      console.log('[AuthProvider] initializeAuth → START');

      let session = null;
      try {
        console.log('[AuthProvider] appel getSession...');
        const result = await supabase.auth.getSession();
        console.log('[AuthProvider] getSession résultat brut:', result);
        session = result?.data?.session ?? null;
        console.log('[AuthProvider] session trouvée:', !!session, '| userId:', session?.user?.id ?? 'aucun');
      } catch (e: any) {
        console.error('[AuthProvider] ❌ getSession EXCEPTION:', e?.message ?? e);
      }

      setUser(session?.user ?? null);
      console.log('[AuthProvider] setUser appelé avec:', session?.user?.id ?? 'null');

      if (session?.user) {
        console.log('[AuthProvider] fetch profile pour userId:', session.user.id);
        try {
          const { data, error } = await supabase
            .from('user_profiles')
            .select('*')
            .eq('id', session.user.id)
            .single();
          console.log('[AuthProvider] profile fetch résultat →', { data: !!data, error: error?.message ?? null });
          setProfile(data as UserProfile | null);
        } catch (e: any) {
          console.error('[AuthProvider] ❌ profile fetch EXCEPTION:', e?.message ?? e);
        }
      } else {
        console.log('[AuthProvider] pas de session → pas de fetch profile');
      }

      console.log('[AuthProvider] setLoading(false) → appelé');
      setLoading(false);
      console.log('[AuthProvider] initializeAuth → DONE');
    };

    initializeAuth();

    console.log('[AuthProvider] onAuthStateChange → abonnement');
    const { data: { subscription } } = supabase.auth.onAuthStateChange(
      async (event: string, session: any) => {
        console.log('[AuthProvider] onAuthStateChange event:', event, '| session:', !!session);
        setUser(session?.user ?? null);

        if (session?.user) {
          console.log('[AuthProvider] onAuthStateChange → fetch profile userId:', session.user.id);
          try {
            const { data, error } = await supabase
              .from('user_profiles')
              .select('*')
              .eq('id', session.user.id)
              .single();
            console.log('[AuthProvider] onAuthStateChange profile →', { data: !!data, error: error?.message ?? null });
            setProfile(data as UserProfile | null);
          } catch (e: any) {
            console.error('[AuthProvider] ❌ onAuthStateChange profile EXCEPTION:', e?.message ?? e);
          }
        } else {
          console.log('[AuthProvider] onAuthStateChange → pas de session, reset profile');
          setProfile(null);
        }
      }
    );

    return () => {
      console.log('[AuthProvider] useEffect cleanup → unsubscribe');
      subscription.unsubscribe();
    };
  }, []);

  console.log('[AuthProvider] state actuel →', { loading, userId: user?.id ?? 'null', hasProfile: !!profile });

  return (
    <AuthContext.Provider value={{ user, profile, loading, refreshProfile }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth doit être utilisé dans un AuthProvider');
  }
  return context;
}