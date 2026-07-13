'use client';

import { useEffect, useState } from 'react';
import { useRouter, usePathname } from 'next/navigation';
import { useAuth } from '@/components/providers/AuthProvider';
import { createClient } from '@/lib/supabase/client';

export function ProfileCheck({ children }: { children: React.ReactNode }) {
  const { user, loading: authLoading } = useAuth();
  const router = useRouter();
  const pathname = usePathname();
  const [ready, setReady] = useState(false);

  useEffect(() => {
    // Si l'auth charge encore, attendre
    if (authLoading) return;

    // Si pas d'utilisateur, laisser passer
    if (!user) {
      setReady(true);
      return;
    }

    // Ne pas vérifier sur la page de complétion
    if (pathname === '/complete-profile') {
      setReady(true);
      return;
    }

    // Vérifier le profil une seule fois
    const checkProfile = async () => {
      try {
        const supabase = createClient();
        const { data: profile } = await supabase
          .from('user_profiles')
          .select('nickname')
          .eq('id', user.id)
          .single();

        if (!profile?.nickname) {
          router.replace('/complete-profile');
        } else {
          setReady(true);
        }
      } catch {
        setReady(true);
      }
    };

    checkProfile();
  }, [user?.id, authLoading]); // Dépendances stables

  // Afficher le contenu immédiatement si l'auth n'est pas chargée
  if (authLoading) {
    return <>{children}</>;
  }

  return <>{children}</>;
}