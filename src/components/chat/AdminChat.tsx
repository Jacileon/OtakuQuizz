'use client';

import { useState, useRef, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { MessageSquare, Send, Loader2, ArrowLeft, Headphones, X, Plus } from 'lucide-react';
import { useAdminConversations, useAllAdminConversations, useAdminChat } from '@/lib/hooks/useChat';
import { createAdminConversation } from '@/lib/actions/admin-chat';
import { AdminMessage, UserProfile } from '@/types';
import { formatDistanceToNow } from 'date-fns';
import { fr } from 'date-fns/locale';
import { cn } from '@/lib/utils';
import { toast } from '@/lib/hooks/useToast';

export function AdminChatButton({ onOpen }: { onOpen: () => void }) {
  return (
    <Button variant="outline" onClick={onOpen} className="gap-2">
      <Headphones className="h-4 w-4" />
      Contacter le support
    </Button>
  );
}

export function AdminChatList({ onOpenChat }: { onOpenChat: (conversationId: string) => void }) {
  const { conversations, loading } = useAdminConversations();

  if (loading) {
    return (
      <div className="flex justify-center py-8">
        <Loader2 className="h-6 w-6 animate-spin text-primary" />
      </div>
    );
  }

  if (conversations.length === 0) {
    return (
      <Card>
        <CardContent className="flex flex-col items-center justify-center py-12">
          <Headphones className="h-12 w-12 text-muted-foreground mb-4" />
          <p className="text-muted-foreground">Aucune conversation avec le support</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-2">
      {conversations.map((conv) => (
        <Card
          key={conv.id}
          className="cursor-pointer hover:bg-accent/50 transition-colors"
          onClick={() => onOpenChat(conv.id)}
        >
          <CardContent className="flex items-center gap-3 p-4">
            <div className="h-10 w-10 rounded-full bg-primary/10 flex items-center justify-center">
              <Headphones className="h-5 w-5 text-primary" />
            </div>
            <div className="flex-1 min-w-0">
              <div className="flex items-center justify-between">
                <p className="font-semibold">{conv.subject}</p>
                <Badge variant={conv.status === 'open' ? 'default' : conv.status === 'closed' ? 'secondary' : 'outline'}>
                  {conv.status === 'open' ? 'Ouvert' : conv.status === 'closed' ? 'Fermé' : 'Assigné'}
                </Badge>
              </div>
              {conv.last_message && (
                <p className="text-sm text-muted-foreground truncate">{conv.last_message.content}</p>
              )}
            </div>
            {conv.unread_count > 0 && (
              <Badge className="h-5 w-5 rounded-full p-0 flex items-center justify-center text-xs">
                {conv.unread_count}
              </Badge>
            )}
          </CardContent>
        </Card>
      ))}
    </div>
  );
}

export function AdminChatWindow({ conversationId, onBack, isAdmin = false }: { conversationId: string; onBack: () => void; isAdmin?: boolean }) {
  const { messages, loading, sending, send, close } = useAdminChat(conversationId);
  const [input, setInput] = useState('');
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLTextAreaElement>(null);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const handleSend = async () => {
    if (!input.trim() || sending) return;
    const content = input;
    setInput('');
    await send(content);
    inputRef.current?.focus();
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  return (
    <Card className="flex flex-col h-[600px]">
      <CardHeader className="flex flex-row items-center gap-3 py-3 border-b">
        <Button variant="ghost" size="sm" onClick={onBack}>
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <CardTitle className="text-base flex-1">Support Admin</CardTitle>
        {isAdmin && (
          <Button variant="destructive" size="sm" onClick={close}>
            <X className="h-4 w-4 mr-1" />
            Fermer
          </Button>
        )}
      </CardHeader>

      <CardContent className="flex-1 overflow-y-auto p-4 space-y-4">
        {loading ? (
          <div className="flex justify-center py-8">
            <Loader2 className="h-6 w-6 animate-spin text-primary" />
          </div>
        ) : messages.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-12 text-muted-foreground">
            <Headphones className="h-8 w-8 mb-2" />
            <p>Commencez la conversation</p>
            <p className="text-sm">Décrivez votre problème</p>
          </div>
        ) : (
          messages.map((msg) => <AdminMessageBubble key={msg.id} message={msg} />)
        )}
        <div ref={messagesEndRef} />
      </CardContent>

      <div className="p-4 border-t">
        <div className="flex gap-2">
          <Textarea
            ref={inputRef}
            placeholder="Écrire un message..."
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
            disabled={sending}
            rows={1}
            className="resize-none"
          />
          <Button onClick={handleSend} disabled={!input.trim() || sending}>
            {sending ? <Loader2 className="h-4 w-4 animate-spin" /> : <Send className="h-4 w-4" />}
          </Button>
        </div>
      </div>
    </Card>
  );
}

function AdminMessageBubble({ message }: { message: AdminMessage }) {
  const [isOwn, setIsOwn] = useState(false);

  useEffect(() => {
    const checkOwn = async () => {
      const { createClient } = await import('@/lib/supabase/client');
      const supabase = createClient();
      const { data: { user } } = await supabase.auth.getUser();
      setIsOwn(user?.id === message.sender_id);
    };
    checkOwn();
  }, [message.sender_id]);

  return (
    <div className={cn('flex', isOwn ? 'justify-end' : 'justify-start')}>
      <div
        className={cn(
          'max-w-[70%] rounded-lg px-4 py-2',
          isOwn ? 'bg-primary text-primary-foreground' : 'bg-muted'
        )}
      >
        <p className="text-sm">{message.content}</p>
        <p className={cn('text-xs mt-1', isOwn ? 'text-primary-foreground/70' : 'text-muted-foreground')}>
          {formatDistanceToNow(new Date(message.created_at), { addSuffix: true, locale: fr })}
        </p>
      </div>
    </div>
  );
}

export function NewAdminConversationDialog({ onCreated }: { onCreated: (id: string) => void }) {
  const [subject, setSubject] = useState('');
  const [loading, setLoading] = useState(false);

  const handleCreate = async () => {
    if (!subject.trim()) return;
    try {
      setLoading(true);
      const id = await createAdminConversation(subject);
      toast({ title: 'Conversation créée' });
      onCreated(id);
    } catch (error: any) {
      toast({ title: 'Erreur', description: error.message, variant: 'destructive' });
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card>
      <CardContent className="p-4 space-y-4">
        <p className="font-medium">Nouvelle demande de support</p>
        <Input
          placeholder="Sujet de votre demande..."
          value={subject}
          onChange={(e) => setSubject(e.target.value)}
        />
        <Button onClick={handleCreate} disabled={!subject.trim() || loading} className="w-full">
          {loading ? <Loader2 className="h-4 w-4 animate-spin mr-2" /> : <Plus className="h-4 w-4 mr-2" />}
          Créer la conversation
        </Button>
      </CardContent>
    </Card>
  );
}

export function AdminSupportList({ onOpenChat }: { onOpenChat: (conversationId: string) => void }) {
  const { conversations, loading } = useAllAdminConversations();

  if (loading) {
    return (
      <div className="flex justify-center py-8">
        <Loader2 className="h-6 w-6 animate-spin text-primary" />
      </div>
    );
  }

  return (
    <div className="space-y-2">
      {conversations.map((conv) => (
        <Card
          key={conv.id}
          className="cursor-pointer hover:bg-accent/50 transition-colors"
          onClick={() => onOpenChat(conv.id)}
        >
          <CardContent className="flex items-center gap-3 p-4">
            <Avatar>
              <AvatarImage src={conv.user.avatar_url || undefined} />
              <AvatarFallback>{conv.user.username[0].toUpperCase()}</AvatarFallback>
            </Avatar>
            <div className="flex-1 min-w-0">
              <div className="flex items-center justify-between">
                <p className="font-semibold">{conv.user.username}</p>
                <Badge variant={conv.status === 'open' ? 'default' : conv.status === 'closed' ? 'secondary' : 'outline'}>
                  {conv.status === 'open' ? 'Ouvert' : conv.status === 'closed' ? 'Fermé' : 'Assigné'}
                </Badge>
              </div>
              <p className="text-sm text-muted-foreground">{conv.subject}</p>
              {conv.last_message && (
                <p className="text-xs text-muted-foreground truncate mt-1">{conv.last_message.content}</p>
              )}
            </div>
            {conv.unread_count > 0 && (
              <Badge className="h-5 w-5 rounded-full p-0 flex items-center justify-center text-xs">
                {conv.unread_count}
              </Badge>
            )}
          </CardContent>
        </Card>
      ))}
    </div>
  );
}