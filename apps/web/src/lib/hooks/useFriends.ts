'use client';

import { useState, useEffect, useCallback } from 'react';
import { Friendship, FriendshipStatus, UserProfile } from '@/types';
import { toast } from '@/lib/hooks/useToast';
import {
  sendFriendRequest,
  acceptFriendRequest,
  rejectFriendRequest,
  removeFriend,
  searchUsers,
  getFriends,
  getPendingRequests,
  getSentRequests,
  getFriendshipStatus,
  getRecentNotifications,
  markNotificationAsRead,
} from '@/lib/actions/friends';
import { Notification } from '@/types';

export function usePendingRequestsCount() {
  const [count, setCount] = useState(0);

  useEffect(() => {
    const fetchCount = async () => {
      try {
        const requests = await getPendingRequests();
        setCount(requests.length);
      } catch (error) {
        console.error('Erreur comptage demandes:', error);
      }
    };
    fetchCount();

    const interval = setInterval(fetchCount, 30000);
    return () => clearInterval(interval);
  }, []);

  return count;
}

export function useFriends() {
  const [friends, setFriends] = useState<(Friendship & { friend: UserProfile })[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchFriends = useCallback(async () => {
    try {
      setLoading(true);
      const data = await getFriends();
      setFriends(data);
    } catch (error) {
      console.error('Erreur chargement amis:', error);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchFriends();
  }, [fetchFriends]);

  const remove = useCallback(async (friendshipId: string) => {
    try {
      await removeFriend(friendshipId);
      setFriends(prev => prev.filter(f => f.id !== friendshipId));
      toast({ title: 'Ami supprimé', description: 'L\'ami a été retiré de votre liste' });
    } catch (error: any) {
      toast({ title: 'Erreur', description: error.message || 'Impossible de supprimer', variant: 'destructive' });
    }
  }, []);

  return { friends, loading, remove, refetch: fetchFriends };
}

export function useFriendRequests() {
  const [received, setReceived] = useState<(Friendship & { requester: UserProfile })[]>([]);
  const [sent, setSent] = useState<(Friendship & { addressee: UserProfile })[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchRequests = useCallback(async () => {
    try {
      setLoading(true);
      const [recv, snt] = await Promise.all([getPendingRequests(), getSentRequests()]);
      setReceived(recv);
      setSent(snt);
    } catch (error) {
      console.error('Erreur chargement demandes:', error);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchRequests();
  }, [fetchRequests]);

  const accept = useCallback(async (friendshipId: string) => {
    try {
      await acceptFriendRequest(friendshipId);
      setReceived(prev => prev.filter(r => r.id !== friendshipId));
      toast({ title: 'Demande acceptée', description: 'Vous êtes maintenant amis' });
    } catch (error: any) {
      toast({ title: 'Erreur', description: error.message || 'Impossible d\'accepter', variant: 'destructive' });
    }
  }, []);

  const reject = useCallback(async (friendshipId: string) => {
    try {
      await rejectFriendRequest(friendshipId);
      setReceived(prev => prev.filter(r => r.id !== friendshipId));
      toast({ title: 'Demande refusée' });
    } catch (error: any) {
      toast({ title: 'Erreur', description: error.message || 'Impossible de refuser', variant: 'destructive' });
    }
  }, []);

  return { received, sent, loading, accept, reject, refetch: fetchRequests };
}

export function useUserSearch() {
  const [results, setResults] = useState<UserProfile[]>([]);
  const [loading, setLoading] = useState(false);

  const search = useCallback(async (query: string) => {
    if (query.length < 2) {
      setResults([]);
      return;
    }
    try {
      setLoading(true);
      const data = await searchUsers(query);
      setResults(data);
    } catch (error) {
      console.error('Erreur recherche:', error);
    } finally {
      setLoading(false);
    }
  }, []);

  return { results, loading, search };
}

export function useFriendshipStatus(userId: string) {
  const [status, setStatus] = useState<FriendshipStatus | null>(null);
  const [friendshipId, setFriendshipId] = useState<string | null>(null);
  const [isRequester, setIsRequester] = useState(false);
  const [loading, setLoading] = useState(true);

  const fetchStatus = useCallback(async () => {
    try {
      setLoading(true);
      const result = await getFriendshipStatus(userId);
      setStatus(result.status);
      setFriendshipId(result.friendshipId);
      setIsRequester(result.isRequester);
    } catch (error) {
      console.error('Erreur statut amitié:', error);
    } finally {
      setLoading(false);
    }
  }, [userId]);

  useEffect(() => {
    fetchStatus();
  }, [fetchStatus]);

  const sendRequest = useCallback(async () => {
    try {
      await sendFriendRequest(userId);
      toast({ title: 'Demande envoyée', description: 'Votre demande d\'ami a été envoyée' });
      await fetchStatus();
    } catch (error: any) {
      toast({ title: 'Erreur', description: error.message || 'Impossible d\'envoyer la demande', variant: 'destructive' });
    }
  }, [userId, fetchStatus]);

  const accept = useCallback(async () => {
    if (!friendshipId) return;
    try {
      await acceptFriendRequest(friendshipId);
      toast({ title: 'Demande acceptée', description: 'Vous êtes maintenant amis' });
      await fetchStatus();
    } catch (error: any) {
      toast({ title: 'Erreur', description: error.message || 'Impossible d\'accepter', variant: 'destructive' });
    }
  }, [friendshipId, fetchStatus]);

  const reject = useCallback(async () => {
    if (!friendshipId) return;
    try {
      await rejectFriendRequest(friendshipId);
      toast({ title: 'Demande refusée' });
      await fetchStatus();
    } catch (error: any) {
      toast({ title: 'Erreur', description: error.message || 'Impossible de refuser', variant: 'destructive' });
    }
  }, [friendshipId, fetchStatus]);

  const remove = useCallback(async () => {
    if (!friendshipId) return;
    try {
      await removeFriend(friendshipId);
      toast({ title: 'Ami supprimé' });
      await fetchStatus();
    } catch (error: any) {
      toast({ title: 'Erreur', description: error.message || 'Impossible de supprimer', variant: 'destructive' });
    }
  }, [friendshipId, fetchStatus]);

  return { status, friendshipId, isRequester, loading, sendRequest, accept, reject, remove, refetch: fetchStatus };
}

export function useNotifications() {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchNotifications = useCallback(async () => {
    try {
      setLoading(true);
      const data = await getRecentNotifications();
      setNotifications(data);
    } catch (error) {
      console.error('Erreur chargement notifications:', error);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchNotifications();
  }, [fetchNotifications]);

  const markAsRead = useCallback(async (notificationId: string) => {
    try {
      await markNotificationAsRead(notificationId);
      setNotifications(prev =>
        prev.map(n => n.id === notificationId ? { ...n, is_read: true } : n)
      );
    } catch (error) {
      console.error('Erreur marquage notification:', error);
    }
  }, []);

  return { notifications, loading, markAsRead, refetch: fetchNotifications };
}