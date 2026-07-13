'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Bell } from 'lucide-react';
import { cn } from '@/lib/utils';

export function NotificationBell() {
  const [unreadCount] = useState(0); // À connecter avec Supabase Realtime

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon" className="relative">
          <Bell className="h-5 w-5" />
          {unreadCount > 0 && (
            <span className="absolute -top-1 -right-1 h-4 w-4 rounded-full bg-brand text-[10px] font-bold flex items-center justify-center text-white">
              {unreadCount}
            </span>
          )}
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-72">
        <div className="p-4 text-center text-sm text-muted-foreground">
          <Bell className="h-8 w-8 mx-auto mb-2 opacity-50" />
          Aucune notification
        </div>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

