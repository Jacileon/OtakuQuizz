// ============================================================
// AUTH LAYOUT - Pages authentifiées
// ============================================================

import { Navbar } from '@/components/layout/Navbar';
import { MobileNav } from '@/components/layout/MobileNav';
import { Sidebar } from '@/components/layout/Sidebar';
import { ProfileCheck } from '@/components/providers/ProfileCheck';

export default function AuthLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <ProfileCheck>
      <div className="min-h-screen bg-dark">
        <Navbar />
        <div className="flex">
          <aside className="hidden lg:block w-72 shrink-0">
            <Sidebar />
          </aside>
          <main className="flex-1 min-h-[calc(100vh-64px)] lg:min-h-[calc(100vh-64px)]">
            {children}
          </main>
        </div>
        <MobileNav />
      </div>
    </ProfileCheck>
  );
}