'use client';

import { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Loader2, Plus, Edit, Trash, Archive, Play, Pause, Trophy, Gift, Clock, Calendar } from 'lucide-react';
import { toast } from '@/lib/hooks/useToast';
import { createOfficialQuiz, updateOfficialQuiz, getOfficialQuizzes } from '@/lib/actions/official-quizzes';
import { Quiz, QuizReward } from '@/types';
import { formatDistanceToNow } from 'date-fns';
import { fr } from 'date-fns/locale';

export default function AdminOfficialQuizzesPage() {
  const [quizzes, setQuizzes] = useState<Quiz[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreate, setShowCreate] = useState(false);

  useEffect(() => {
    fetchQuizzes();
  }, []);

  const fetchQuizzes = async () => {
    try {
      setLoading(true);
      const data = await getOfficialQuizzes();
      setQuizzes(data);
    } catch (error) {
      console.error('Erreur chargement quiz:', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="container max-w-6xl mx-auto py-8 px-4">
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-3xl font-bold flex items-center gap-3">
            <Trophy className="h-8 w-8 text-yellow-500" />
            Quiz Officiels
          </h1>
          <p className="text-muted-foreground mt-2">Gérez les quiz officiels de la plateforme</p>
        </div>
        <Button onClick={() => setShowCreate(true)} className="gap-2">
          <Plus className="h-4 w-4" />
          Créer un quiz officiel
        </Button>
      </div>

      {showCreate ? (
        <OfficialQuizForm onClose={() => { setShowCreate(false); fetchQuizzes(); }} />
      ) : (
        <div className="space-y-4">
          {loading ? (
            <div className="flex justify-center py-12">
              <Loader2 className="h-8 w-8 animate-spin" />
            </div>
          ) : quizzes.length === 0 ? (
            <Card>
              <CardContent className="flex flex-col items-center justify-center py-12">
                <Trophy className="h-12 w-12 text-muted-foreground mb-4" />
                <p className="text-muted-foreground">Aucun quiz officiel</p>
              </CardContent>
            </Card>
          ) : (
            quizzes.map(quiz => (
              <OfficialQuizCard key={quiz.id} quiz={quiz} onUpdate={fetchQuizzes} />
            ))
          )}
        </div>
      )}
    </div>
  );
}

function OfficialQuizCard({ quiz, onUpdate }: { quiz: Quiz; onUpdate: () => void }) {
  const getStatusBadge = () => {
    switch (quiz.status) {
      case 'scheduled': return <Badge variant="outline"><Clock className="h-3 w-3 mr-1" /> Programmé</Badge>;
      case 'active': return <Badge variant="default"><Play className="h-3 w-3 mr-1" /> Actif</Badge>;
      case 'archived': return <Badge variant="secondary"><Archive className="h-3 w-3 mr-1" /> Archivé</Badge>;
      default: return <Badge variant="secondary">{quiz.status}</Badge>;
    }
  };

  const handleArchive = async () => {
    try {
      await updateOfficialQuiz(quiz.id, { status: 'archived' });
      toast({ title: 'Quiz archivé' });
      onUpdate();
    } catch (error: any) {
      toast({ title: 'Erreur', description: error.message, variant: 'destructive' });
    }
  };

  return (
    <Card>
      <CardContent className="p-6">
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <div className="flex items-center gap-3 mb-2">
              <h3 className="text-lg font-semibold">{quiz.title}</h3>
              {getStatusBadge()}
            </div>
            {quiz.description && (
              <p className="text-sm text-muted-foreground mb-3">{quiz.description}</p>
            )}
            <div className="flex flex-wrap gap-4 text-sm text-muted-foreground">
              <span className="flex items-center gap-1">
                <Calendar className="h-4 w-4" />
                {quiz.starts_at ? formatDistanceToNow(new Date(quiz.starts_at), { addSuffix: true, locale: fr }) : 'Non programmé'}
              </span>
              <span>{quiz.question_count} questions</span>
              <span>{quiz.play_count} parties</span>
            </div>
          </div>
          <div className="flex gap-2">
            {quiz.status === 'active' && (
              <Button variant="outline" size="sm" onClick={handleArchive}>
                <Archive className="h-4 w-4" />
              </Button>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

function OfficialQuizForm({ onClose }: { onClose: () => void }) {
  const [loading, setLoading] = useState(false);
  const [formData, setFormData] = useState({
    title: '',
    description: '',
    category: 'Shonen',
    subcategory: 'Général',
    series: '',
    starts_at: '',
    ends_at: '',
    duration_seconds: 30,
    duration_mode: 'per_question' as 'global' | 'per_question',
  });
  const [rewards, setRewards] = useState<QuizReward[]>([]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      setLoading(true);
      await createOfficialQuiz({
        ...formData,
        rewards,
      });
      toast({ title: 'Quiz officiel créé' });
      onClose();
    } catch (error: any) {
      toast({ title: 'Erreur', description: error.message, variant: 'destructive' });
    } finally {
      setLoading(false);
    }
  };

  const addReward = () => {
    setRewards([...rewards, {
      title: '',
      description: '',
      url: '',
      rank_from: 1,
      rank_to: 1,
    }]);
  };

  const updateReward = (index: number, field: string, value: any) => {
    const updated = [...rewards];
    (updated[index] as any)[field] = value;
    setRewards(updated);
  };

  const removeReward = (index: number) => {
    setRewards(rewards.filter((_, i) => i !== index));
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Créer un quiz officiel</CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">Titre</label>
              <Input
                value={formData.title}
                onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                required
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Série</label>
              <Input
                value={formData.series}
                onChange={(e) => setFormData({ ...formData, series: e.target.value })}
                required
              />
            </div>
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium">Description</label>
            <Textarea
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              rows={3}
            />
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">Catégorie</label>
              <Input
                value={formData.category}
                onChange={(e) => setFormData({ ...formData, category: e.target.value })}
                required
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Sous-catégorie</label>
              <Input
                value={formData.subcategory}
                onChange={(e) => setFormData({ ...formData, subcategory: e.target.value })}
                required
              />
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">Date de début</label>
              <Input
                type="datetime-local"
                value={formData.starts_at}
                onChange={(e) => setFormData({ ...formData, starts_at: e.target.value })}
                required
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Date de fin</label>
              <Input
                type="datetime-local"
                value={formData.ends_at}
                onChange={(e) => setFormData({ ...formData, ends_at: e.target.value })}
                required
              />
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">Durée (secondes)</label>
              <Input
                type="number"
                value={formData.duration_seconds}
                onChange={(e) => setFormData({ ...formData, duration_seconds: parseInt(e.target.value) })}
                min={5}
                required
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Mode de durée</label>
              <select
                value={formData.duration_mode}
                onChange={(e) => setFormData({ ...formData, duration_mode: e.target.value as any })}
                className="w-full p-2 rounded-md border bg-background"
              >
                <option value="per_question">Par question</option>
                <option value="global">Global (tout le quiz)</option>
              </select>
            </div>
          </div>

          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <label className="text-sm font-medium">Récompenses</label>
              <Button type="button" variant="outline" size="sm" onClick={addReward}>
                <Plus className="h-4 w-4 mr-1" /> Ajouter
              </Button>
            </div>
            {rewards.map((reward, index) => (
              <Card key={index} className="p-4">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <Input
                    placeholder="Titre de la récompense"
                    value={reward.title}
                    onChange={(e) => updateReward(index, 'title', e.target.value)}
                  />
                  <Input
                    placeholder="URL (optionnel)"
                    value={reward.url || ''}
                    onChange={(e) => updateReward(index, 'url', e.target.value)}
                  />
                  <Input
                    placeholder="Description"
                    value={reward.description || ''}
                    onChange={(e) => updateReward(index, 'description', e.target.value)}
                  />
                  <div className="flex gap-2">
                    <Input
                      type="number"
                      placeholder="Rang de"
                      value={reward.rank_from}
                      onChange={(e) => updateReward(index, 'rank_from', parseInt(e.target.value))}
                      min={1}
                    />
                    <Input
                      type="number"
                      placeholder="Rang à"
                      value={reward.rank_to}
                      onChange={(e) => updateReward(index, 'rank_to', parseInt(e.target.value))}
                      min={1}
                    />
                    <Button type="button" variant="destructive" size="sm" onClick={() => removeReward(index)}>
                      <Trash className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              </Card>
            ))}
          </div>

          <div className="flex justify-end gap-3">
            <Button type="button" variant="outline" onClick={onClose}>
              Annuler
            </Button>
            <Button type="submit" disabled={loading}>
              {loading ? <Loader2 className="h-4 w-4 animate-spin mr-2" /> : null}
              Créer le quiz
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}