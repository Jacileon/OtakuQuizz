"use client";

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { Home, Compass, Plus, Trophy, User, Users } from 'lucide-react';
import { cn } from '@/lib/utils';
import { useAuth } from '@/components/providers/AuthProvider';

export function MobileNav() {
  const pathname = usePathname();
  const { profile } = useAuth();

  const mobileLinks = [
    { href: '/dashboard', icon: Home, label: 'Accueil' },
    { href: '/explore', icon: Compass, label: 'Explorer' },
    { href: '/friends', icon: Users, label: 'Amis' },
    { href: '/quiz/create', icon: Plus, label: 'Créer', isAction: true },
    { href: '/leaderboard', icon: Trophy, label: 'Classement' },
    { href: profile?.username ? `/profile/${profile.username}` : '/dashboard', icon: User, label: 'Profil' },
  ];

  return (
    <nav className="fixed bottom-0 left-0 right-0 z-50 bg-dark/95 backdrop-blur border-t border-dark-border md:hidden">
      <div className="flex items-center justify-around h-16">
        {mobileLinks.map((link) => (
          <Link
            key={link.label}
            href={link.href}
            className={cn(
              'flex flex-col items-center gap-0.5 py-2 px-3 rounded-lg transition-colors',
              link.isAction
                ? 'bg-brand text-white -mt-4 h-14 w-14 rounded-full items-center justify-center shadow-lg shadow-brand/30'
                : pathname === link.href
                ? 'text-brand'
                : 'text-muted-foreground hover:text-white'
            )}
          >
            <link.icon className={cn('h-5 w-5', link.isAction && 'h-6 w-6')} />
            {!link.isAction && <span className="text-[10px]">{link.label}</span>}
          </Link>
        ))}
      </div>
    </nav>
  );
}
