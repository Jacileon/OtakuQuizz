'use client';

import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Bell, Check, UserPlus, UserMinus, X } from 'lucide-react';
import { useNotifications } from '@/lib/hooks/useFriends';
import { Notification } from '@/types';
import { formatDistanceToNow } from 'date-fns';
import { fr } from 'date-fns/locale';

export function NotificationsList() {
  const { notifications, loading, markAsRead } = useNotifications();

  if (loading) {
    return <div className="text-center py-4 text-muted-foreground">Chargement...</div>;
  }

  if (notifications.length === 0) {
    return (
      <Card>
        <CardContent className="flex flex-col items-center justify-center py-8">
          <Bell className="h-8 w-8 text-muted-foreground mb-2" />
          <p className="text-muted-foreground text-sm">Aucune notification</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-2">
      {notifications.map((notification) => (
        <NotificationCard
          key={notification.id}
          notification={notification}
          onMarkAsRead={markAsRead}
        />
      ))}
    </div>
  );
}

function NotificationCard({
  notification,
  onMarkAsRead,
}: {
  notification: Notification;
  onMarkAsRead: (id: string) => void;
}) {
  const getIcon = () => {
    switch (notification.type) {
      case 'friend_request':
        return notification.title.includes('acceptée') ? (
          <Check className="h-4 w-4 text-green-500" />
        ) : (
          <X className="h-4 w-4 text-red-500" />
        );
      default:
        return <Bell className="h-4 w-4" />;
    }
  };

  const timeAgo = formatDistanceToNow(new Date(notification.created_at), {
    addSuffix: true,
    locale: fr,
  });

  return (
    <Card className={notification.is_read ? 'opacity-60' : ''}>
      <CardContent className="flex items-start gap-3 p-4">
        <div className="mt-0.5">{getIcon()}</div>
        <div className="flex-1 min-w-0">
          <p className="font-medium text-sm">{notification.title}</p>
          <p className="text-sm text-muted-foreground">{notification.message}</p>
          <p className="text-xs text-muted-foreground mt-1">{timeAgo}</p>
        </div>
        {!notification.is_read && (
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onMarkAsRead(notification.id)}
            className="shrink-0"
          >
            <Check className="h-4 w-4" />
          </Button>
        )}
      </CardContent>
    </Card>
  );
}