'use client';

// ============================================================
// REALTIME PROVIDER
// ============================================================

import { useEffect } from '../../../node_modules/@types/react';
import { useAuth } from '@/components/providers/AuthProvider';
import { useRealtime } from '@/lib/hooks/useRealtime';

export function RealtimeProvider({ children }: { children: React.ReactNode }) {
  const { user } = useAuth();
  useRealtime(user?.id || null);

  return <>{children}</>;
}

