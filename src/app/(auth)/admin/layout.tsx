// ============================================================
// LAYOUT ADMIN
// ============================================================

import { requireAdmin } from '@/lib/auth/actions';
import { redirect } from '../../../../node_modules/next/navigation';

export default async function AdminLayout({ children }: { children: React.ReactNode }) {
  try {
    await requireAdmin();
    return (
      <div className="min-h-screen bg-dark">
        <div className="border-b border-dark-border p-4">
          <h1 className="font-display text-xl tracking-wider text-brand">ADMINISTRATION</h1>
        </div>
        {children}
      </div>
    );
  } catch {
    redirect('/dashboard');
  }
}

