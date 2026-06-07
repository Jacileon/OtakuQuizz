// ============================================================
// CLIENT SUPABASE CÔTÉ SERVEUR
// ============================================================

import { createServerClient } from '@supabase/ssr';
import { cookies } from 'next/headers';

export function createClient() {
  const cookieStore = cookies();

  return createServerClient(
    process.env.NEXT_PUBLIC_SUPABASE_URL!,
    process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY!,
    {
      cookies: {
        getAll() {
          const all = cookieStore.getAll();
          const authCookies = all.filter(c => c.name.includes('sb-') || c.name.includes('supabase'));
          console.log('🍪 [server:getAll]', all.map(c => c.name));
          authCookies.forEach(c => {
            console.log('🍪 cookie value preview:', c.name, 'len:', c.value?.length, 'start:', c.value?.substring(0, 30));
          });
          return all;
        },
        setAll(cookiesToSet) {
          console.log('🍪 [server:setAll]', cookiesToSet.map(c => ({ name: c.name, value: c.value?.slice(0, 20) + '...' })));
          try {
            cookiesToSet.forEach(({ name, value, options }) =>
              cookieStore.set(name, value, options)
            );
          } catch {
            // Ignoré en Server Components / lors du rendu
          }
        },
      },
    }
  );
}

export function createAdminClient() {
  return createServerClient(
    process.env.NEXT_PUBLIC_SUPABASE_URL!,
    process.env.SUPABASE_SERVICE_ROLE_KEY!,
    {
      cookies: {
        getAll() {
          return [];
        },
        setAll() {},
      },
    }
  );
}
