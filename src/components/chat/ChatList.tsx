'use client';

import { useState, useRef, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  MessageSquare,
  Send,
  Loader2,
  ArrowLeft,
  MessageCircle,
  MoreVertical,
  Trash2,
  Check,
  X,
  CheckSquare,
  Square,
} from 'lucide-react';
import { useConversations, useChat, useUnreadMessages } from '@/lib/hooks/useChat';
import { Message, UserProfile } from '@/types';
import { formatDistanceToNow } from 'date-fns';
import { fr } from 'date-fns/locale';
import { cn } from '@/lib/utils';
import { toast } from '@/lib/hooks/useToast';
import { deleteMessages, deleteConversation } from '@/lib/actions/chat';
import { createClient } from '@/lib/supabase/client';

export function ChatList({ onOpenChat }: { onOpenChat: (friendId: string) => void }) {
  const { conversations, loading } = useConversations();

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
          <MessageCircle className="h-12 w-12 text-muted-foreground mb-4" />
          <p className="text-muted-foreground">Aucune conversation</p>
          <p className="text-sm text-muted-foreground">Commencez à discuter avec vos amis</p>
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
          onClick={() => onOpenChat(conv.other_user.id)}
        >
          <CardContent className="flex items-center gap-3 p-4">
            <Avatar>
              <AvatarImage src={conv.other_user.avatar_url || undefined} />
              <AvatarFallback className="bg-primary/20 text-primary font-bold">
                {conv.other_user.username[0].toUpperCase()}
              </AvatarFallback>
            </Avatar>
            <div className="flex-1 min-w-0">
              <div className="flex items-center justify-between">
                <p className="font-semibold truncate">{conv.other_user.username}</p>
                {conv.last_message && (
                  <span className="text-xs text-muted-foreground">
                    {formatDistanceToNow(new Date(conv.last_message.created_at), { addSuffix: true, locale: fr })}
                  </span>
                )}
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

export function ChatWindow({ friendId, onBack }: { friendId: string; onBack: () => void }) {
  const { conversationId, messages, loading, sending, send, refetch } = useChat(friendId);
  const [input, setInput] = useState('');
  const [currentUserId, setCurrentUserId] = useState<string | null>(null);
  const [otherUser, setOtherUser] = useState<UserProfile | null>(null);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  // Mode sélection
  const [selectionMode, setSelectionMode] = useState(false);
  const [selectedMessages, setSelectedMessages] = useState<Set<string>>(new Set());

  // Dialogs
  const [showDeleteMessages, setShowDeleteMessages] = useState(false);
  const [showDeleteConversation, setShowDeleteConversation] = useState(false);

  useEffect(() => {
    const supabase = createClient();
    supabase.auth.getUser().then(({ data }) => {
      if (data.user) setCurrentUserId(data.user.id);
    });

    supabase
      .from('user_profiles')
      .select('*')
      .eq('id', friendId)
      .single()
      .then(({ data }) => {
        if (data) setOtherUser(data as UserProfile);
      });
  }, [friendId]);

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

  const toggleSelectionMode = () => {
    setSelectionMode(!selectionMode);
    setSelectedMessages(new Set());
  };

  const toggleMessageSelection = (messageId: string) => {
    setSelectedMessages(prev => {
      const newSet = new Set(prev);
      if (newSet.has(messageId)) {
        newSet.delete(messageId);
      } else {
        newSet.add(messageId);
      }
      return newSet;
    });
  };

  const selectAll = () => {
    const ownMessageIds = messages
      .filter(m => m.sender_id === currentUserId)
      .map(m => m.id);
    setSelectedMessages(new Set(ownMessageIds));
  };

  const handleDeleteMessages = async () => {
    if (!conversationId || selectedMessages.size === 0) return;
    try {
      await deleteMessages(conversationId, Array.from(selectedMessages));
      toast({ title: `${selectedMessages.size} message(s) supprimé(s)` });
      setSelectionMode(false);
      setSelectedMessages(new Set());
      refetch();
    } catch (error: any) {
      toast({ title: 'Erreur', description: error.message, variant: 'destructive' });
    }
    setShowDeleteMessages(false);
  };

  const handleDeleteConversation = async () => {
    if (!conversationId) return;
    try {
      await deleteConversation(conversationId);
      toast({ title: 'Conversation supprimée' });
      onBack();
    } catch (error: any) {
      toast({ title: 'Erreur', description: error.message, variant: 'destructive' });
    }
    setShowDeleteConversation(false);
  };

  return (
    <>
      <Card className="flex flex-col h-[600px]">
        <CardHeader className="flex flex-row items-center gap-3 py-3 border-b">
          <Button variant="ghost" size="sm" onClick={selectionMode ? toggleSelectionMode : onBack}>
            {selectionMode ? <X className="h-4 w-4" /> : <ArrowLeft className="h-4 w-4" />}
          </Button>

          {otherUser && (
            <Avatar className="h-8 w-8">
              <AvatarImage src={otherUser.avatar_url || undefined} />
              <AvatarFallback className="bg-primary/20 text-primary font-bold">
                {otherUser.username[0].toUpperCase()}
              </AvatarFallback>
            </Avatar>
          )}

          <CardTitle className="text-base flex-1">
            {selectionMode
              ? `${selectedMessages.size} sélectionné(s)`
              : otherUser?.username || 'Conversation'
            }
          </CardTitle>

          {selectionMode ? (
            <div className="flex gap-2">
              <Button variant="ghost" size="sm" onClick={selectAll}>
                <CheckSquare className="h-4 w-4" />
              </Button>
              <Button
                variant="destructive"
                size="sm"
                onClick={() => setShowDeleteMessages(true)}
                disabled={selectedMessages.size === 0}
              >
                <Trash2 className="h-4 w-4" />
              </Button>
            </div>
          ) : (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="sm">
                  <MoreVertical className="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem onClick={toggleSelectionMode}>
                  <CheckSquare className="h-4 w-4 mr-2" />
                  Sélectionner des messages
                </DropdownMenuItem>
                <DropdownMenuItem
                  onClick={() => setShowDeleteConversation(true)}
                  className="text-destructive focus:text-destructive"
                >
                  <Trash2 className="h-4 w-4 mr-2" />
                  Supprimer la discussion
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          )}
        </CardHeader>

        <CardContent className="flex-1 overflow-y-auto p-4 space-y-3">
          {loading ? (
            <div className="flex justify-center py-8">
              <Loader2 className="h-6 w-6 animate-spin text-primary" />
            </div>
          ) : messages.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-muted-foreground">
              <MessageSquare className="h-8 w-8 mb-2" />
              <p>Aucun message</p>
              <p className="text-sm">Commencez la conversation !</p>
            </div>
          ) : (
            messages.map((msg) => {
              const isOwn = msg.sender_id === currentUserId;
              return (
                <MessageBubble
                  key={msg.id}
                  message={msg}
                  isOwn={isOwn}
                  otherUser={otherUser}
                  selectionMode={selectionMode}
                  isSelected={selectedMessages.has(msg.id)}
                  onToggleSelect={() => toggleMessageSelection(msg.id)}
                  onLongPress={() => {
                    if (!selectionMode) {
                      setSelectionMode(true);
                      setSelectedMessages(new Set([msg.id]));
                    }
                  }}
                />
              );
            })
          )}
          <div ref={messagesEndRef} />
        </CardContent>

        {!selectionMode && (
          <div className="p-4 border-t">
            <div className="flex gap-2">
              <Input
                ref={inputRef}
                placeholder="Écrire un message..."
                value={input}
                onChange={(e) => setInput(e.target.value)}
                onKeyDown={handleKeyDown}
                disabled={sending}
              />
              <Button onClick={handleSend} disabled={!input.trim() || sending}>
                {sending ? <Loader2 className="h-4 w-4 animate-spin" /> : <Send className="h-4 w-4" />}
              </Button>
            </div>
          </div>
        )}
      </Card>

      {/* Dialog confirmation suppression messages */}
      <Dialog open={showDeleteMessages} onOpenChange={setShowDeleteMessages}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Supprimer {selectedMessages.size} message(s) ?</DialogTitle>
            <DialogDescription>
              Cette action est irréversible. Les messages seront supprimés définitivement.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowDeleteMessages(false)}>
              Annuler
            </Button>
            <Button variant="destructive" onClick={handleDeleteMessages}>
              Supprimer
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Dialog confirmation suppression conversation */}
      <Dialog open={showDeleteConversation} onOpenChange={setShowDeleteConversation}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Supprimer toute la discussion ?</DialogTitle>
            <DialogDescription>
              Cette action est irréversible. Tous les messages seront supprimés et la conversation disparaîtra de votre liste.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowDeleteConversation(false)}>
              Annuler
            </Button>
            <Button variant="destructive" onClick={handleDeleteConversation}>
              Supprimer la discussion
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}

function MessageBubble({
  message,
  isOwn,
  otherUser,
  selectionMode,
  isSelected,
  onToggleSelect,
  onLongPress,
}: {
  message: Message;
  isOwn: boolean;
  otherUser: UserProfile | null;
  selectionMode: boolean;
  isSelected: boolean;
  onToggleSelect: () => void;
  onLongPress: () => void;
}) {
  const pressTimer = useRef<NodeJS.Timeout>();

  const handleMouseDown = () => {
    pressTimer.current = setTimeout(() => {
      onLongPress();
    }, 500);
  };

  const handleMouseUp = () => {
    if (pressTimer.current) {
      clearTimeout(pressTimer.current);
    }
  };

  const handleClick = () => {
    if (selectionMode) {
      onToggleSelect();
    }
  };

  return (
    <div
      className={cn(
        'flex items-end gap-2 mb-2',
        isOwn ? 'flex-row-reverse' : 'flex-row'
      )}
      onMouseDown={handleMouseDown}
      onMouseUp={handleMouseUp}
      onMouseLeave={handleMouseUp}
      onTouchStart={handleMouseDown}
      onTouchEnd={handleMouseUp}
      onClick={handleClick}
    >
      {/* Avatar */}
      {!selectionMode && (
        <div className="shrink-0">
          {isOwn ? (
            <div className="w-8 h-8 rounded-full bg-blue-500 flex items-center justify-center text-white text-xs font-bold">
              Vous
            </div>
          ) : (
            <Avatar className="w-8 h-8">
              <AvatarImage src={otherUser?.avatar_url || undefined} />
              <AvatarFallback className="bg-gray-300 text-gray-700 font-bold text-xs">
                {otherUser?.username?.[0]?.toUpperCase() || '?'}
              </AvatarFallback>
            </Avatar>
          )}
        </div>
      )}

      {/* Case à cocher en mode sélection */}
      {selectionMode && (
        <div className="shrink-0 w-6 h-6 flex items-center justify-center">
          {isSelected ? (
            <div className="w-5 h-5 rounded bg-blue-500 flex items-center justify-center">
              <Check className="h-3 w-3 text-white" />
            </div>
          ) : (
            <div className="w-5 h-5 rounded border-2 border-gray-400" />
          )}
        </div>
      )}

      {/* Bulle de message */}
      <div
        className={cn(
          'max-w-[70%] px-4 py-2 rounded-2xl',
          isOwn
            ? 'bg-blue-500 text-white rounded-br-sm'
            : 'bg-gray-200 text-gray-900 rounded-bl-sm',
          selectionMode && 'cursor-pointer',
          selectionMode && isSelected && 'ring-2 ring-blue-500'
        )}
        style={{
          marginLeft: isOwn ? 0 : undefined,
          marginRight: isOwn ? 0 : undefined,
        }}
      >
        <p className="text-sm whitespace-pre-wrap break-words">{message.content}</p>
        <p className={cn(
          'text-[10px] mt-1',
          isOwn ? 'text-blue-100 text-right' : 'text-gray-500 text-right'
        )}>
          {formatDistanceToNow(new Date(message.created_at), { addSuffix: true, locale: fr })}
        </p>
      </div>
    </div>
  );
}