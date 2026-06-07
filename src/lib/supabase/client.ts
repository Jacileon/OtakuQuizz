// ============================================================
// CLIENT SUPABASE CÔTÉ NAVIGATEUR (Singleton)
// ============================================================

import { createBrowserClient } from '@supabase/ssr';

let clientInstance: ReturnType<typeof createBrowserClient> | null = null;

export function createClient() {
  return createBrowserClient(
    process.env.NEXT_PUBLIC_SUPABASE_URL!,
    process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY!
  );
}

// Singleton - réutiliser le même client pour garder la session
export function getBrowserClient() {
  if (!clientInstance) {
    clientInstance = createClient();
  }
  return clientInstance;
}
