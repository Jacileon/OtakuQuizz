'use client';

import { useEffect, useState } from 'react';
import { useRouter, usePathname } from 'next/navigation';
import { useAuth } from '@/components/providers/AuthProvider';
import { createClient } from '@/lib/supabase/client';
import { Loader2 } from 'lucide-react';

export function ProfileCheck({ children }: { children: React.ReactNode }) {
  const { user, loading: authLoading } = useAuth();
  const router = useRouter();
  const pathname = usePathname();
  const [checking, setChecking] = useState(true);

  useEffect(() => {
    const checkProfile = async () => {
      if (authLoading) return;
      
      if (!user) {
        setChecking(false);
        return;
      }

      // Ne pas vérifier si on est déjà sur la page de complétion
      if (pathname === '/complete-profile') {
        setChecking(false);
        return;
      }

      const supabase = createClient();
      const { data: profile } = await supabase
        .from('user_profiles')
        .select('nickname')
        .eq('id', user.id)
        .single();

      if (!profile?.nickname) {
        router.push('/complete-profile');
        return;
      }

      setChecking(false);
    };

    checkProfile();
  }, [user, authLoading, pathname, router]);

  if (authLoading || checking) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-dark">
        <Loader2 className="h-8 w-8 animate-spin text-brand" />
      </div>
    );
  }

  return <>{children}</>;
}