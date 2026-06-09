'use client';

import { useState } from 'react';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { UserSearch } from '@/components/friends/UserSearch';
import { FriendList } from '@/components/friends/FriendList';
import { FriendRequests } from '@/components/friends/FriendRequests';
import { ChallengeRequests } from '@/components/friends/ChallengeRequests';
import { NotificationsList } from '@/components/friends/NotificationsList';
import { ChatList, ChatWindow } from '@/components/chat/ChatList';
import { Users, Search, Bell, MessageCircle, Swords, ArrowLeft } from 'lucide-react';
import { useUnreadMessages } from '@/lib/hooks/useChat';
import { SupportFloatingButton } from '@/components/support/SupportFloatingButton';
import { Button } from '@/components/ui/button';

export default function FriendsPage() {
  const [chatFriendId, setChatFriendId] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState('friends');
  const unreadCount = useUnreadMessages();

  const handleOpenChat = (friendId: string) => {
    setChatFriendId(friendId);
    setActiveTab('chat');
  };

  const handleBackFromChat = () => {
    setChatFriendId(null);
    setActiveTab('friends');
  };

  return (
    <div className="container max-w-2xl mx-auto py-8 px-4">
      <div className="mb-8">
        <h1 className="text-3xl font-bold flex items-center gap-3">
          <Users className="h-8 w-8 text-primary" />
          Amis
          {unreadCount > 0 && (
            <span className="h-6 w-6 rounded-full bg-red-500 text-white text-sm flex items-center justify-center">
              {unreadCount}
            </span>
          )}
        </h1>
        <p className="text-muted-foreground mt-2">
          Gérez vos amis et discutez avec eux
        </p>
      </div>

      <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-6">
        <TabsList className="grid w-full grid-cols-5">
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
          <TabsTrigger value="challenges" className="flex items-center gap-2">
            <Swords className="h-4 w-4" />
            Défis
          </TabsTrigger>
          <TabsTrigger value="chat" className="flex items-center gap-2 relative">
            <MessageCircle className="h-4 w-4" />
            Chat
            {unreadCount > 0 && (
              <span className="absolute -top-1 -right-1 h-4 w-4 rounded-full bg-red-500 text-white text-[10px] flex items-center justify-center">
                {unreadCount}
              </span>
            )}
          </TabsTrigger>
        </TabsList>

        <TabsContent value="friends">
          <FriendList onOpenChat={handleOpenChat} />
        </TabsContent>

        <TabsContent value="search">
          <UserSearch />
        </TabsContent>

        <TabsContent value="requests">
          <FriendRequests />
        </TabsContent>

        <TabsContent value="challenges">
          <ChallengeRequests />
        </TabsContent>

        <TabsContent value="chat">
          {chatFriendId ? (
            <ChatWindow friendId={chatFriendId} onBack={handleBackFromChat} />
          ) : (
            <ChatList onOpenChat={handleOpenChat} />
          )}
        </TabsContent>
      </Tabs>

      <SupportFloatingButton />
    </div>
  );
}