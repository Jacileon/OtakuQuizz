'use client';

import Link from '../../../node_modules/next/link';
import { useState } from 'react';
import { useAuth } from '@/components/providers/AuthProvider';
import { Button } from '@/components/ui/button';
import { RankBadge } from '@/components/ui/RankBadge';
import { XPBar } from '@/components/ui/XPBar';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Sheet, SheetContent, SheetTrigger } from '@/components/ui/sheet';
import {
  Menu,
  Home,
  Compass,
  Trophy,
  Calendar,
  Plus,
  LogOut,
  User,
  Settings,
  Sword,
  Users,
  Swords,
} from 'lucide-react';
import { signOut } from '@/lib/auth/actions';
import { getInitials, getDisplayName } from '@/lib/utils';
import { usePendingRequestsCount } from '@/lib/hooks/useFriends';

const navLinks = [
  { href: '/dashboard', label: 'Accueil', icon: Home },
  { href: '/explore', label: 'Explorer', icon: Compass },
  { href: '/friends', label: 'Amis', icon: Users },
  { href: '/challenges', label: 'Défis', icon: Swords },
  { href: '/events', label: 'Événements', icon: Calendar },
  { href: '/leaderboard', label: 'Classements', icon: Trophy },
];

function FriendsNavLink() {
  const pendingCount = usePendingRequestsCount();

  return (
    <Link
      href="/friends"
      className="flex items-center gap-2 px-3 py-2 text-sm text-muted-foreground hover:text-white transition-colors rounded-md hover:bg-dark-surface relative"
    >
      <Users className="h-4 w-4" />
      Amis
      {pendingCount > 0 && (
        <span className="absolute -top-1 -right-1 h-5 w-5 rounded-full bg-red-500 text-white text-xs flex items-center justify-center font-bold">
          {pendingCount}
        </span>
      )}
    </Link>
  );
}

export function Navbar() {
  const { user, profile } = useAuth();
  const [mobileOpen, setMobileOpen] = useState(false);
  const pendingCount = usePendingRequestsCount();

  return (
    <header className="sticky top-0 z-50 w-full border-b border-dark-border bg-dark/95 backdrop-blur supports-[backdrop-filter]:bg-dark/60">
      <div className="container flex h-16 items-center justify-between">
        {/* Logo */}
        <Link href="/dashboard" className="flex items-center gap-2 mr-6">
          <Sword className="h-7 w-7 text-brand" />
          <span className="font-display text-xl tracking-wider text-white hidden sm:inline">
            OTAKU QUIZ AFRICA
          </span>
          <span className="font-display text-lg tracking-wider text-white sm:hidden">
            OQA
          </span>
        </Link>

        {/* Desktop Nav */}
        <nav className="hidden md:flex items-center gap-1">
          {navLinks.map((link) => (
            link.href === '/friends' ? (
              <FriendsNavLink key={link.href} />
            ) : (
              <Link
                key={link.href}
                href={link.href}
                className="flex items-center gap-2 px-3 py-2 text-sm text-muted-foreground hover:text-white transition-colors rounded-md hover:bg-dark-surface"
              >
                <link.icon className="h-4 w-4" />
                {link.label}
              </Link>
            )
          ))}
        </nav>

        {/* Right Side */}
        <div className="flex items-center gap-3">
          {user ? (
            <>
              <Link href="/quiz/create">
                <Button size="sm" className="hidden sm:flex gap-2">
                  <Plus className="h-4 w-4" />
                  Créer un Quiz
                </Button>
              </Link>

              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <button className="flex items-center gap-2 rounded-full hover:bg-dark-surface p-1 pr-3 transition-colors">
                    <Avatar className="h-8 w-8 border-2" style={{ borderColor: profile ? getRankColor(profile.rank) : '#888' }}>
                      <AvatarImage src={profile?.avatar_url || undefined} />
                      <AvatarFallback className="text-xs bg-dark-surface">
                        {profile ? getInitials(getDisplayName(profile)) : '?'}
                      </AvatarFallback>
                    </Avatar>
                    <span className="hidden sm:inline text-sm font-medium">{profile ? getDisplayName(profile) : ''}</span>
                    {profile && <RankBadge rank={profile.rank} size="sm" />}
                  </button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" className="w-64">
                  <div className="px-3 py-2">
                    <p className="font-medium text-sm">{profile ? getDisplayName(profile) : ''}</p>
                    <p className="text-xs text-muted-foreground">{profile?.email}</p>
                    {profile && (
                      <div className="mt-2">
                        <XPBar currentXP={profile.xp} rank={profile.rank} showNumbers={false} />
                      </div>
                    )}
                  </div>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem asChild>
                    <Link href="/profil" className="flex items-center gap-2 cursor-pointer">
                      <User className="h-4 w-4" /> Mon Profil
                    </Link>
                  </DropdownMenuItem>
                  {profile?.username && (
                    <DropdownMenuItem asChild>
                      <Link href={`/profile/${profile.username}`} className="flex items-center gap-2 cursor-pointer">
                        <User className="h-4 w-4" /> Profil Public
                      </Link>
                    </DropdownMenuItem>
                  )}
                  <DropdownMenuItem asChild>
                    <Link href="/profile/edit" className="flex items-center gap-2 cursor-pointer">
                      <Settings className="h-4 w-4" /> Paramètres
                    </Link>
                  </DropdownMenuItem>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem
                    onClick={() => signOut()}
                    className="text-red-400 focus:text-red-400 cursor-pointer"
                  >
                    <LogOut className="h-4 w-4 mr-2" /> Déconnexion
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </>
          ) : (
            <Link href="/login">
              <Button>Connexion</Button>
            </Link>
          )}

          {/* Mobile Menu */}
          <Sheet open={mobileOpen} onOpenChange={setMobileOpen}>
            <SheetTrigger asChild className="md:hidden">
              <Button variant="ghost" size="icon">
                <Menu className="h-5 w-5" />
              </Button>
            </SheetTrigger>
            <SheetContent side="left" className="w-72">
              <div className="flex flex-col gap-4 mt-8">
                <Link href="/dashboard" className="flex items-center gap-2 text-lg font-display">
                  <Sword className="h-6 w-6 text-brand" /> OTAKU QUIZ AFRICA
                </Link>
                <div className="flex flex-col gap-2 mt-4">
                  {navLinks.map((link) => (
                    <Link
                      key={link.href}
                      href={link.href}
                      onClick={() => setMobileOpen(false)}
                      className="flex items-center gap-3 px-3 py-2.5 rounded-md hover:bg-dark-surface transition-colors relative"
                    >
                      <link.icon className="h-5 w-5 text-brand" />
                      {link.label}
                      {link.href === '/friends' && pendingCount > 0 && (
                        <span className="ml-auto h-5 w-5 rounded-full bg-red-500 text-white text-xs flex items-center justify-center font-bold">
                          {pendingCount}
                        </span>
                      )}
                    </Link>
                  ))}
                  <Link
                    href="/profil"
                    onClick={() => setMobileOpen(false)}
                    className="flex items-center gap-3 px-3 py-2.5 rounded-md hover:bg-dark-surface transition-colors"
                  >
                    <User className="h-5 w-5 text-brand" /> Mon Profil
                  </Link>
                  <Link
                    href="/quiz/create"
                    onClick={() => setMobileOpen(false)}
                    className="flex items-center gap-3 px-3 py-2.5 rounded-md hover:bg-dark-surface transition-colors"
                  >
                    <Plus className="h-5 w-5 text-brand" /> Créer un Quiz
                  </Link>
                </div>
              </div>
            </SheetContent>
          </Sheet>
        </div>
      </div>
    </header>
  );
}

function getRankColor(rank: string): string {
  const colors: Record<string, string> = {
    'F': '#888888', 'E': '#4CAF50', 'D': '#2196F3', 'C': '#9C27B0',
    'B': '#FF9800', 'A': '#F44336', 'S': '#FFD700', 'S+': '#FFA500',
    'SS': '#FF69B4', 'SSS': '#00FFFF', 'Légende': '#FF0080',
  };
  return colors[rank] || '#888888';
}

