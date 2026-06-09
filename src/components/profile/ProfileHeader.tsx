'use client';

import { UserProfile } from '@/types';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Card, CardContent } from '@/components/ui/card';
import { RankBadge } from '@/components/ui/RankBadge';
import { Button } from '@/components/ui/button';
import { MapPin, Heart, Edit, Clock } from 'lucide-react';
import { getInitials, getDisplayName } from '@/lib/utils';
import Link from '../../../node_modules/next/link';

interface ProfileHeaderProps {
  profile: UserProfile;
  stats: any;
  isOwnProfile: boolean;
}

export function ProfileHeader({ profile, stats, isOwnProfile }: ProfileHeaderProps) {
  return (
    <Card className="border-dark-border bg-dark-card overflow-hidden">
      <div className="h-32 bg-gradient-to-r from-brand/20 to-accent/20" />
      <CardContent className="p-6 relative">
        <div className="flex flex-col md:flex-row items-start md:items-end gap-4 -mt-16">
          <Avatar className="h-24 w-24 border-4 border-dark-card" style={{ borderColor: getRankColor(profile.rank) }}>
            <AvatarImage src={profile.avatar_url || undefined} />
            <AvatarFallback className="text-2xl bg-dark-surface">
              {getInitials(getDisplayName(profile))}
            </AvatarFallback>
          </Avatar>

          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-3 flex-wrap">
              <h1 className="font-display text-2xl tracking-wider">{getDisplayName(profile)}</h1>
              <RankBadge rank={profile.rank} size="sm" />
            </div>

            <div className="flex items-center gap-4 mt-2 text-sm text-muted-foreground">
              {profile.country && (
                <span className="flex items-center gap-1">
                  <MapPin className="h-3.5 w-3.5" /> {profile.country}
                </span>
              )}
              {profile.favorite_anime && (
                <span className="flex items-center gap-1">
                  <Heart className="h-3.5 w-3.5" /> {profile.favorite_anime}
                </span>
              )}
              <span className="flex items-center gap-1">
                <Clock className="h-3.5 w-3.5" /> {stats?.quizzes_played || 0} quiz joués
              </span>
            </div>

            {profile.bio && (
              <p className="mt-2 text-sm">{profile.bio}</p>
            )}
          </div>

          {isOwnProfile && (
            <Link href="/profile/edit">
              <Button variant="outline" className="gap-2">
                <Edit className="h-4 w-4" /> Modifier
              </Button>
            </Link>
          )}
        </div>
      </CardContent>
    </Card>
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
