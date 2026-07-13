'use client';

import { useState, useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Badge } from '@/components/ui/badge';
import { Loader2, Swords, Users, Zap, ArrowLeft, UserPlus } from 'lucide-react';
import { useFriends } from '@/lib/hooks/useFriends';
import { createChallengeSession, inviteToChallenge } from '@/lib/actions/challenges';
import { toast } from '@/lib/hooks/useToast';
import { checkChallengeParticipationLimit } from '@/lib/actions/challenges';
import { UserProfile } from '@/types';

export default function CreateChallengePage() {
  const params = useParams();
  const router = useRouter();
  const quizId = params.id as string;
  const { friends, loading: loadingFriends } = useFriends();
  
  const [loading, setLoading] = useState(false);
  const [xpBet, setXpBet] = useState(100);
  const [selectedFriends, setSelectedFriends] = useState<Set<string>>(new Set());
  const [quizTitle, setQuizTitle] = useState('');

  useEffect(() => {
    const fetchQuiz = async () => {
      const { createClient } = await import('@/lib/supabase/client');
      const supabase = createClient();
      const { data: quiz } = await supabase
        .from('quizzes')
        .select('title')
        .eq('id', quizId)
        .single();
      if (quiz) setQuizTitle(quiz.title);
    };
    fetchQuiz();
  }, [quizId]);

  const toggleFriend = (friendId: string) => {
    setSelectedFriends(prev => {
      const newSet = new Set(prev);
      if (newSet.has(friendId)) {
        newSet.delete(friendId);
      } else {
        newSet.add(friendId);
      }
      return newSet;
    });
  };

  const handleCreate = async () => {
    if (selectedFriends.size === 0) {
      toast({ title: 'Erreur', description: 'Sélectionnez au moins un ami', variant: 'destructive' });
      return;
    }

    try {
      setLoading(true);

      // Vérifier la limite de participation pour chaque ami sélectionné
      const friendsArray = Array.from(selectedFriends);
      for (const friendId of friendsArray) {
        const result = await checkChallengeParticipationLimit(quizId, friendId);
        if (!result.allowed) {
          toast({ 
            title: 'Limite atteinte', 
            description: result.message,
            variant: 'destructive' 
          });
          setLoading(false);
          return;
        }
      }

      // Créer la session de défi
      const sessionId = await createChallengeSession(quizId, xpBet);

      // Inviter chaque ami sélectionné
      for (const friendId of friendsArray) {
        await inviteToChallenge(sessionId, friendId);
      }

      toast({ title: 'Défi créé !', description: 'Invitations envoyées à vos amis' });
      router.push(`/challenges/${sessionId}`);
    } catch (error: any) {
      toast({ title: 'Erreur', description: error.message, variant: 'destructive' });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="p-4 lg:p-8 pb-24 md:pb-8">
      <div className="max-w-2xl mx-auto">
        <Button variant="ghost" onClick={() => router.back()} className="gap-2 mb-6">
          <ArrowLeft className="h-4 w-4" />
          Retour
        </Button>

        <Card className="mb-6">
          <CardHeader>
            <CardTitle className="flex items-center gap-3">
              <Swords className="h-6 w-6 text-purple-500" />
              Créer un défi
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="p-4 rounded-lg bg-dark-surface mb-4">
              <p className="text-sm text-muted-foreground">Quiz</p>
              <p className="font-medium">{quizTitle || 'Chargement...'}</p>
            </div>

            <div className="space-y-2 mb-4">
              <label className="text-sm font-medium flex items-center gap-2">
                <Zap className="h-4 w-4 text-yellow-500" />
                Mise XP
              </label>
              <Input
                type="number"
                value={xpBet}
                onChange={(e) => setXpBet(parseInt(e.target.value) || 0)}
                min={10}
                max={10000}
                placeholder="Montant XP à miser"
              />
              <p className="text-xs text-muted-foreground">
                Montant d'XP que vous misez sur ce défi (minimum 10)
              </p>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Users className="h-5 w-5" />
              Sélectionner des amis ({selectedFriends.size} sélectionné(s))
            </CardTitle>
          </CardHeader>
          <CardContent>
            {loadingFriends ? (
              <div className="flex justify-center py-8">
                <Loader2 className="h-6 w-6 animate-spin" />
              </div>
            ) : friends.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">
                <UserPlus className="h-12 w-12 mx-auto mb-4 opacity-50" />
                <p>Aucun ami</p>
                <p className="text-sm">Ajoutez des amis pour les défier</p>
              </div>
            ) : (
              <div className="space-y-2">
                {friends.map((friendship) => {
                  const friend = friendship.friend;
                  const isSelected = selectedFriends.has(friend.id);

                  return (
                    <div
                      key={friendship.id}
                      onClick={() => toggleFriend(friend.id)}
                      className={`flex items-center gap-3 p-3 rounded-lg cursor-pointer transition-colors ${
                        isSelected
                          ? 'border-2 border-purple-500 bg-purple-500/10'
                          : 'border border-muted hover:border-muted-foreground/50'
                      }`}
                    >
                      <Avatar>
                        <AvatarImage src={friend.avatar_url || undefined} />
                        <AvatarFallback>{friend.username[0].toUpperCase()}</AvatarFallback>
                      </Avatar>
                      <div className="flex-1 min-w-0">
                        <p className="font-medium truncate">{friend.nickname || friend.username}</p>
                        <div className="flex items-center gap-2">
                          <Badge variant="secondary" className="text-xs">{friend.rank}</Badge>
                          <span className="text-xs text-muted-foreground">Niv. {friend.level}</span>
                        </div>
                      </div>
                      {isSelected && (
                        <div className="h-6 w-6 rounded-full bg-purple-500 flex items-center justify-center">
                          <svg className="h-4 w-4 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                          </svg>
                        </div>
                      )}
                    </div>
                  );
                })}
              </div>
            )}

            {selectedFriends.size > 0 && (
              <div className="mt-6 pt-4 border-t">
                <Button
                  onClick={handleCreate}
                  disabled={loading || xpBet < 10}
                  className="w-full gap-2"
                  size="lg"
                >
                  {loading ? (
                    <Loader2 className="h-5 w-5 animate-spin" />
                  ) : (
                    <Swords className="h-5 w-5" />
                  )}
                  Créer le défi ({selectedFriends.size} ami(s) - {xpBet} XP)
                </Button>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}