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

  const [user, setUser] = useState<User | null>(null);
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [loading, setLoading] = useState(true);

  const refreshProfile = async () => {
    if (!user) return;
    const supabase = getBrowserClient();
    const { data, error } = await supabase
      .from('user_profiles')
      .select('*')
      .eq('id', user.id)
      .single();
    setProfile(data as UserProfile | null);
  };

  useEffect(() => {
    const supabase = getBrowserClient();

    const initializeAuth = async () => {

      let session = null;
      try {
        const result = await supabase.auth.getSession();
        session = result?.data?.session ?? null;
      } catch (e: any) {
        console.error('[AuthProvider] ❌ getSession EXCEPTION:', e?.message ?? e);
      }

      setUser(session?.user ?? null);

      if (session?.user) {
        try {
          const { data, error } = await supabase
            .from('user_profiles')
            .select('*')
            .eq('id', session.user.id)
            .single();
          setProfile(data as UserProfile | null);
        } catch (e: any) {
          console.error('[AuthProvider] ❌ profile fetch EXCEPTION:', e?.message ?? e);
        }
      } else {
      }     
      setLoading(false);
    };

    initializeAuth();

    const { data: { subscription } } = supabase.auth.onAuthStateChange(
      async (event: string, session: any) => {
        setUser(session?.user ?? null);

        if (session?.user) {
          try {
            const { data, error } = await supabase
              .from('user_profiles')
              .select('*')
              .eq('id', session.user.id)
              .single();
            setProfile(data as UserProfile | null);
          } catch (e: any) {
            console.error('[AuthProvider] ❌ onAuthStateChange profile EXCEPTION:', e?.message ?? e);
          }
        } else {
          setProfile(null);
        }
      }
    );

    return () => {
      subscription.unsubscribe();
    };
  }, []);

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