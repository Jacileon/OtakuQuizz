'use client';

import { useState, useEffect, useCallback, useRef } from 'react';
import { Conversation, Message, AdminConversation, AdminMessage, UserProfile } from '@/types';
import { toast } from '@/lib/hooks/useToast';
import {
  getOrCreateConversation,
  getConversations,
  getMessages,
  sendMessage,
  getUnreadMessagesCount,
} from '@/lib/actions/chat';
import {
  createAdminConversation,
  getMyAdminConversations,
  getAllAdminConversations,
  getAdminMessages,
  sendAdminMessage,
  closeAdminConversation,
} from '@/lib/actions/admin-chat';

export function useConversations() {
  const [conversations, setConversations] = useState<(Conversation & { other_user: UserProfile; last_message?: Message; unread_count: number })[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchConversations = useCallback(async () => {
    try {
      setLoading(true);
      const data = await getConversations();
      setConversations(data);
    } catch (error) {
      console.error('Erreur chargement conversations:', error);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchConversations();
  }, [fetchConversations]);

  return { conversations, loading, refetch: fetchConversations };
}

export function useChat(friendId: string | null) {
  const [conversationId, setConversationId] = useState<string | null>(null);
  const [messages, setMessages] = useState<Message[]>([]);
  const [loading, setLoading] = useState(true);
  const [sending, setSending] = useState(false);
  const intervalRef = useRef<NodeJS.Timeout>();

  useEffect(() => {
    if (!friendId) return;

    const init = async () => {
      try {
        const convId = await getOrCreateConversation(friendId);
        setConversationId(convId);
        const msgs = await getMessages(convId);
        setMessages(msgs);
      } catch (error) {
        console.error('Erreur init chat:', error);
      } finally {
        setLoading(false);
      }
    };

    init();
  }, [friendId]);

  useEffect(() => {
    if (!conversationId) return;

    intervalRef.current = setInterval(async () => {
      try {
        const msgs = await getMessages(conversationId);
        setMessages(msgs);
      } catch (error) {
        console.error('Erreur refresh messages:', error);
      }
    }, 3000);

    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current);
    };
  }, [conversationId]);

  const send = useCallback(async (content: string) => {
    if (!conversationId || !content.trim()) return;
    try {
      setSending(true);
      const newMessage = await sendMessage(conversationId, content);
      setMessages(prev => [...prev, newMessage]);
    } catch (error: any) {
      toast({ title: 'Erreur', description: error.message || 'Impossible d\'envoyer', variant: 'destructive' });
    } finally {
      setSending(false);
    }
  }, [conversationId]);

  const refetch = useCallback(async () => {
    if (!conversationId) return;
    try {
      const msgs = await getMessages(conversationId);
      setMessages(msgs);
    } catch (error) {
      console.error('Erreur refresh messages:', error);
    }
  }, [conversationId]);

  return { conversationId, messages, loading, sending, send, refetch };
}

export function useUnreadMessages() {
  const [count, setCount] = useState(0);

  useEffect(() => {
    const fetchCount = async () => {
      try {
        const c = await getUnreadMessagesCount();
        setCount(c);
      } catch (error) {
        console.error('Erreur comptage messages:', error);
      }
    };

    fetchCount();
    const interval = setInterval(fetchCount, 15000);
    return () => clearInterval(interval);
  }, []);

  return count;
}

export function useAdminConversations() {
  const [conversations, setConversations] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchConversations = useCallback(async () => {
    try {
      setLoading(true);
      const data = await getMyAdminConversations();
      setConversations(data);
    } catch (error) {
      console.error('Erreur chargement conversations admin:', error);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchConversations();
  }, [fetchConversations]);

  return { conversations, loading, refetch: fetchConversations };
}

export function useAllAdminConversations() {
  const [conversations, setConversations] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchConversations = useCallback(async () => {
    try {
      setLoading(true);
      const data = await getAllAdminConversations();
      setConversations(data);
    } catch (error) {
      console.error('Erreur chargement conversations admin:', error);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchConversations();
  }, [fetchConversations]);

  return { conversations, loading, refetch: fetchConversations };
}

export function useAdminChat(conversationId: string | null) {
  const [messages, setMessages] = useState<AdminMessage[]>([]);
  const [loading, setLoading] = useState(true);
  const [sending, setSending] = useState(false);
  const intervalRef = useRef<NodeJS.Timeout>();

  useEffect(() => {
    if (!conversationId) return;

    const init = async () => {
      try {
        const msgs = await getAdminMessages(conversationId);
        setMessages(msgs);
      } catch (error) {
        console.error('Erreur init chat admin:', error);
      } finally {
        setLoading(false);
      }
    };

    init();
  }, [conversationId]);

  useEffect(() => {
    if (!conversationId) return;

    intervalRef.current = setInterval(async () => {
      try {
        const msgs = await getAdminMessages(conversationId);
        setMessages(msgs);
      } catch (error) {
        console.error('Erreur refresh messages admin:', error);
      }
    }, 3000);

    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current);
    };
  }, [conversationId]);

  const send = useCallback(async (content: string) => {
    if (!conversationId || !content.trim()) return;
    try {
      setSending(true);
      const newMessage = await sendAdminMessage(conversationId, content);
      setMessages(prev => [...prev, newMessage]);
    } catch (error: any) {
      toast({ title: 'Erreur', description: error.message || 'Impossible d\'envoyer', variant: 'destructive' });
    } finally {
      setSending(false);
    }
  }, [conversationId]);

  const close = useCallback(async () => {
    if (!conversationId) return;
    try {
      await closeAdminConversation(conversationId);
      toast({ title: 'Conversation fermée' });
    } catch (error: any) {
      toast({ title: 'Erreur', description: error.message, variant: 'destructive' });
    }
  }, [conversationId]);

  return { messages, loading, sending, send, close };
}