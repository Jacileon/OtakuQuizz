'use client';

import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Loader2, ArrowLeft, Send, MessageSquare, Headphones, X, CheckCircle } from 'lucide-react';
import { useAllAdminConversations, useAdminChat } from '@/lib/hooks/useChat';
import { closeAdminConversation } from '@/lib/actions/admin-chat';
import { toast } from '@/lib/hooks/useToast';
import { formatDistanceToNow } from 'date-fns';
import { fr } from 'date-fns/locale';

export default function AdminTicketsPage() {
  const [selectedConvId, setSelectedConvId] = useState<string | null>(null);

  if (selectedConvId) {
    return (
      <div className="container max-w-2xl mx-auto py-8 px-4">
        <TicketChat conversationId={selectedConvId} onBack={() => setSelectedConvId(null)} />
      </div>
    );
  }

  return (
    <div className="container max-w-4xl mx-auto py-8 px-4">
      <div className="mb-8">
        <h1 className="text-3xl font-bold flex items-center gap-3">
          <Headphones className="h-8 w-8 text-primary" />
          Tickets Support
        </h1>
        <p className="text-muted-foreground mt-2">
          Gérez les demandes des utilisateurs
        </p>
      </div>

      <Tabs defaultValue="open" className="space-y-6">
        <TabsList className="grid w-full grid-cols-2">
          <TabsTrigger value="open" className="flex items-center gap-2">
            <MessageSquare className="h-4 w-4" />
            Ouverts
          </TabsTrigger>
          <TabsTrigger value="closed" className="flex items-center gap-2">
            <CheckCircle className="h-4 w-4" />
            Fermés
          </TabsTrigger>
        </TabsList>

        <TabsContent value="open">
          <TicketList status="open" onSelect={setSelectedConvId} />
        </TabsContent>

        <TabsContent value="closed">
          <TicketList status="closed" onSelect={setSelectedConvId} />
        </TabsContent>
      </Tabs>
    </div>
  );
}

function TicketList({ status, onSelect }: { status: 'open' | 'closed'; onSelect: (id: string) => void }) {
  const { conversations, loading } = useAllAdminConversations();

  if (loading) {
    return (
      <div className="flex justify-center py-8">
        <Loader2 className="h-6 w-6 animate-spin text-primary" />
      </div>
    );
  }

  const filtered = conversations.filter((c: any) => 
    status === 'open' ? c.status !== 'closed' : c.status === 'closed'
  );

  if (filtered.length === 0) {
    return (
      <Card>
        <CardContent className="flex flex-col items-center justify-center py-12">
          <MessageSquare className="h-12 w-12 text-muted-foreground mb-4" />
          <p className="text-muted-foreground">
            {status === 'open' ? 'Aucun ticket ouvert' : 'Aucun ticket fermé'}
          </p>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-2">
      {filtered.map((conv: any) => (
        <Card
          key={conv.id}
          className="cursor-pointer hover:bg-accent/50 transition-colors"
          onClick={() => onSelect(conv.id)}
        >
          <CardContent className="flex items-center gap-3 p-4">
            <Avatar>
              <AvatarImage src={conv.user?.avatar_url || undefined} />
              <AvatarFallback>{conv.user?.username?.[0]?.toUpperCase() || '?'}</AvatarFallback>
            </Avatar>
            <div className="flex-1 min-w-0">
              <div className="flex items-center justify-between">
                <p className="font-semibold">{conv.user?.username || 'Utilisateur'}</p>
                <div className="flex items-center gap-2">
                  <span className="text-xs text-muted-foreground">
                    {conv.last_message_at && formatDistanceToNow(new Date(conv.last_message_at), { addSuffix: true, locale: fr })}
                  </span>
                  <Badge variant={conv.status === 'open' ? 'default' : 'secondary'}>
                    {conv.status === 'open' ? 'Ouvert' : 'Fermé'}
                  </Badge>
                </div>
              </div>
              <p className="text-sm text-muted-foreground truncate">{conv.subject}</p>
              {conv.last_message && (
                <p className="text-xs text-muted-foreground truncate mt-1">{conv.last_message.content}</p>
              )}
            </div>
            {conv.unread_count > 0 && (
              <span className="h-5 w-5 rounded-full bg-red-500 text-white text-xs flex items-center justify-center">
                {conv.unread_count}
              </span>
            )}
          </CardContent>
        </Card>
      ))}
    </div>
  );
}

function TicketChat({ conversationId, onBack }: { conversationId: string; onBack: () => void }) {
  const { messages, loading, sending, send, close } = useAdminChat(conversationId);
  const [input, setInput] = useState('');
  const [closing, setClosing] = useState(false);

  const handleSend = async () => {
    if (!input.trim() || sending) return;
    const content = input;
    setInput('');
    await send(content);
  };

  const handleClose = async () => {
    try {
      setClosing(true);
      await closeAdminConversation(conversationId);
      toast({ title: 'Ticket fermé' });
      onBack();
    } catch (error: any) {
      toast({ title: 'Erreur', description: error.message, variant: 'destructive' });
    } finally {
      setClosing(false);
    }
  };

  return (
    <Card className="flex flex-col h-[600px]">
      <CardHeader className="flex flex-row items-center gap-3 py-3 border-b">
        <Button variant="ghost" size="sm" onClick={onBack}>
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <CardTitle className="text-base flex-1">Ticket Support</CardTitle>
        <Button variant="destructive" size="sm" onClick={handleClose} disabled={closing}>
          {closing ? <Loader2 className="h-4 w-4 animate-spin mr-1" /> : <X className="h-4 w-4 mr-1" />}
          Fermer
        </Button>
      </CardHeader>

      <CardContent className="flex-1 overflow-y-auto p-4 space-y-4">
        {loading ? (
          <div className="flex justify-center py-8">
            <Loader2 className="h-6 w-6 animate-spin" />
          </div>
        ) : messages.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-12 text-muted-foreground">
            <Headphones className="h-8 w-8 mb-2" />
            <p>Aucun message</p>
          </div>
        ) : (
          messages.map((msg: any) => <AdminMessageBubble key={msg.id} message={msg} />)
        )}
      </CardContent>

      <div className="p-4 border-t">
        <div className="flex gap-2">
          <Input
            placeholder="Répondre..."
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleSend()}
            disabled={sending}
          />
          <Button onClick={handleSend} disabled={!input.trim() || sending}>
            {sending ? <Loader2 className="h-4 w-4 animate-spin" /> : <Send className="h-4 w-4" />}
          </Button>
        </div>
      </div>
    </Card>
  );
}

function AdminMessageBubble({ message }: { message: any }) {
  const [isOwn, setIsOwn] = useState(false);

  if (typeof window !== 'undefined') {
    import('@/lib/supabase/client').then(({ createClient }) => {
      const supabase = createClient();
      supabase.auth.getUser().then(({ data }) => {
        setIsOwn(data.user?.id === message.sender_id);
      });
    });
  }

  return (
    <div className={`flex ${isOwn ? 'justify-end' : 'justify-start'}`}>
      <div className={`max-w-[70%] rounded-lg px-4 py-2 ${isOwn ? 'bg-primary text-primary-foreground' : 'bg-muted'}`}>
        <p className="text-sm">{message.content}</p>
        <p className={`text-xs mt-1 ${isOwn ? 'text-primary-foreground/70' : 'text-muted-foreground'}`}>
          {formatDistanceToNow(new Date(message.created_at), { addSuffix: true, locale: fr })}
        </p>
      </div>
    </div>
  );
}