'use client';

import { useState } from 'react';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Badge } from '@/components/ui/badge';
import { Loader2, Send, ArrowLeft, Plus, MessageSquare, X } from 'lucide-react';
import { useAdminConversations, useAdminChat } from '@/lib/hooks/useChat';
import { createAdminConversation } from '@/lib/actions/admin-chat';
import { toast } from '@/lib/hooks/useToast';
import { formatDistanceToNow } from 'date-fns';
import { fr } from 'date-fns/locale';

export function SupportFloatingButton() {
  const [open, setOpen] = useState(false);

  return (
    <>
      <button
        onClick={() => setOpen(true)}
        className="fixed bottom-24 right-6 z-50 h-14 w-14 rounded-full bg-gradient-to-br from-purple-500 to-pink-500 text-white shadow-lg shadow-purple-500/30 hover:shadow-purple-500/50 transition-all hover:scale-110 flex items-center justify-center group"
        title="Contacter le support"
      >
        <span className="text-2xl group-hover:animate-bounce">🎧</span>
      </button>

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <span className="text-xl">🎧</span>
              Support Otaku Quiz
            </DialogTitle>
          </DialogHeader>
          <SupportPanel onClose={() => setOpen(false)} />
        </DialogContent>
      </Dialog>
    </>
  );
}

function SupportPanel({ onClose }: { onClose: () => void }) {
  const { conversations, loading } = useAdminConversations();
  const [selectedConvId, setSelectedConvId] = useState<string | null>(null);
  const [showNew, setShowNew] = useState(false);

  if (selectedConvId) {
    return <SupportChat conversationId={selectedConvId} onBack={() => setSelectedConvId(null)} />;
  }

  if (showNew) {
    return <NewTicket onCreated={(id) => { setSelectedConvId(id); setShowNew(false); }} onBack={() => setShowNew(false)} />;
  }

  return (
    <div className="space-y-4 max-h-[400px] overflow-y-auto">
      <Button onClick={() => setShowNew(true)} className="w-full gap-2">
        <Plus className="h-4 w-4" />
        Nouveau ticket
      </Button>

      {loading ? (
        <div className="flex justify-center py-8">
          <Loader2 className="h-6 w-6 animate-spin" />
        </div>
      ) : conversations.length === 0 ? (
        <div className="text-center py-8 text-muted-foreground">
          <p className="text-4xl mb-2">📭</p>
          <p>Aucun ticket</p>
        </div>
      ) : (
        <div className="space-y-2">
          {conversations.map((conv: any) => (
            <div
              key={conv.id}
              onClick={() => setSelectedConvId(conv.id)}
              className="flex items-center gap-3 p-3 rounded-lg border cursor-pointer hover:bg-accent/50 transition-colors"
            >
              <div className="h-10 w-10 rounded-full bg-purple-500/10 flex items-center justify-center">
                <MessageSquare className="h-5 w-5 text-purple-500" />
              </div>
              <div className="flex-1 min-w-0">
                <div className="flex items-center justify-between">
                  <p className="font-medium text-sm truncate">{conv.subject}</p>
                  <Badge variant={conv.status === 'open' ? 'default' : 'secondary'} className="text-xs">
                    {conv.status === 'open' ? 'Ouvert' : 'Fermé'}
                  </Badge>
                </div>
                {conv.last_message && (
                  <p className="text-xs text-muted-foreground truncate">{conv.last_message.content}</p>
                )}
              </div>
              {conv.unread_count > 0 && (
                <span className="h-5 w-5 rounded-full bg-red-500 text-white text-xs flex items-center justify-center">
                  {conv.unread_count}
                </span>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

function NewTicket({ onCreated, onBack }: { onCreated: (id: string) => void; onBack: () => void }) {
  const [subject, setSubject] = useState('');
  const [message, setMessage] = useState('');
  const [loading, setLoading] = useState(false);

  const handleCreate = async () => {
    if (!subject.trim() || !message.trim()) return;
    try {
      setLoading(true);
      const id = await createAdminConversation(subject);
      const { sendAdminMessage } = await import('@/lib/actions/admin-chat');
      await sendAdminMessage(id, message);
      toast({ title: 'Ticket créé', description: 'Un admin va vous répondre' });
      onCreated(id);
    } catch (error: any) {
      toast({ title: 'Erreur', description: error.message, variant: 'destructive' });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="space-y-4">
      <Button variant="ghost" size="sm" onClick={onBack} className="gap-1">
        <ArrowLeft className="h-4 w-4" /> Retour
      </Button>
      <Input
        placeholder="Sujet du ticket..."
        value={subject}
        onChange={(e) => setSubject(e.target.value)}
      />
      <Textarea
        placeholder="Décrivez votre problème..."
        value={message}
        onChange={(e) => setMessage(e.target.value)}
        rows={4}
      />
      <Button onClick={handleCreate} disabled={!subject.trim() || !message.trim() || loading} className="w-full">
        {loading ? <Loader2 className="h-4 w-4 animate-spin mr-2" /> : <Send className="h-4 w-4 mr-2" />}
        Envoyer
      </Button>
    </div>
  );
}

function SupportChat({ conversationId, onBack }: { conversationId: string; onBack: () => void }) {
  const { messages, loading, sending, send } = useAdminChat(conversationId);
  const [input, setInput] = useState('');

  const handleSend = async () => {
    if (!input.trim() || sending) return;
    const content = input;
    setInput('');
    await send(content);
  };

  return (
    <div className="flex flex-col h-[400px]">
      <div className="flex items-center gap-2 pb-3 border-b">
        <Button variant="ghost" size="sm" onClick={onBack}>
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <span className="font-medium">Support</span>
      </div>

      <div className="flex-1 overflow-y-auto p-3 space-y-3">
        {loading ? (
          <div className="flex justify-center py-8">
            <Loader2 className="h-6 w-6 animate-spin" />
          </div>
        ) : messages.length === 0 ? (
          <div className="text-center py-8 text-muted-foreground">
            <p className="text-3xl mb-2">💬</p>
            <p>Décrivez votre problème</p>
          </div>
        ) : (
          messages.map((msg: any) => <MessageBubble key={msg.id} message={msg} />)
        )}
      </div>

      <div className="flex gap-2 pt-3 border-t">
        <Input
          placeholder="Votre message..."
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && handleSend()}
          disabled={sending}
        />
        <Button onClick={handleSend} disabled={!input.trim() || sending} size="sm">
          {sending ? <Loader2 className="h-4 w-4 animate-spin" /> : <Send className="h-4 w-4" />}
        </Button>
      </div>
    </div>
  );
}

function MessageBubble({ message }: { message: any }) {
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
      <div className={`max-w-[80%] rounded-xl px-3 py-2 ${isOwn ? 'bg-purple-500 text-white' : 'bg-muted'}`}>
        <p className="text-sm">{message.content}</p>
        <p className={`text-[10px] mt-1 ${isOwn ? 'text-white/70' : 'text-muted-foreground'}`}>
          {formatDistanceToNow(new Date(message.created_at), { addSuffix: true, locale: fr })}
        </p>
      </div>
    </div>
  );
}