'use client';

import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { UserSearch } from '@/components/friends/UserSearch';
import { FriendList } from '@/components/friends/FriendList';
import { FriendRequests } from '@/components/friends/FriendRequests';
import { NotificationsList } from '@/components/friends/NotificationsList';
import { Users, Search, Bell, MessageSquare } from 'lucide-react';

export default function FriendsPage() {
  return (
    <div className="container max-w-2xl mx-auto py-8 px-4">
      <div className="mb-8">
        <h1 className="text-3xl font-bold flex items-center gap-3">
          <Users className="h-8 w-8 text-primary" />
          Amis
        </h1>
        <p className="text-muted-foreground mt-2">
          Gérez vos amis et découvrez de nouveaux joueurs
        </p>
      </div>

      <Tabs defaultValue="friends" className="space-y-6">
        <TabsList className="grid w-full grid-cols-4">
          <TabsTrigger value="friends" className="flex items-center gap-2">
            <Users className="h-4 w-4" />
            Amis
          </TabsTrigger>
          <TabsTrigger value="search" className="flex items-center gap-2">
            <Search className="h-4 w-4" />
            Rechercher
          </TabsTrigger>
          <TabsTrigger value="requests" className="flex items-center gap-2">
            <Bell className="h-4 w-4" />
            Demandes
          </TabsTrigger>
          <TabsTrigger value="notifications" className="flex items-center gap-2">
            <MessageSquare className="h-4 w-4" />
            Notifications
          </TabsTrigger>
        </TabsList>

        <TabsContent value="friends">
          <FriendList />
        </TabsContent>

        <TabsContent value="search">
          <UserSearch />
        </TabsContent>

        <TabsContent value="requests">
          <FriendRequests />
        </TabsContent>

        <TabsContent value="notifications">
          <NotificationsList />
        </TabsContent>
      </Tabs>
    </div>
  );
}