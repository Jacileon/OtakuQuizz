// ============================================================
// LAYOUT ADMIN
// ============================================================

import { requireAdmin } from '@/lib/auth/actions';
import { redirect } from '../../../../node_modules/next/navigation';
import Link from 'next/link';
import { Shield, FileText, Headphones } from 'lucide-react';

export default async function AdminLayout({ children }: { children: React.ReactNode }) {
  try {
    await requireAdmin();
    return (
      <div className="min-h-screen bg-dark">
        <div className="border-b border-dark-border p-4">
          <div className="container flex items-center justify-between">
            <h1 className="font-display text-xl tracking-wider text-brand flex items-center gap-2">
              <Shield className="h-6 w-6" />
              ADMINISTRATION
            </h1>
            <nav className="flex items-center gap-4">
              <Link href="/admin/reports" className="flex items-center gap-2 text-sm text-muted-foreground hover:text-white transition-colors">
                <FileText className="h-4 w-4" />
                Signalements
              </Link>
              <Link href="/admin/tickets" className="flex items-center gap-2 text-sm text-muted-foreground hover:text-white transition-colors">
                <Headphones className="h-4 w-4" />
                Tickets
              </Link>
            </nav>
          </div>
        </div>
        {children}
      </div>
    );
  } catch {
    redirect('/dashboard');
  }
}