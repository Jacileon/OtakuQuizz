'use client';

import { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Badge } from '@/components/ui/badge';
import { Loader2, Plus, Trash, Megaphone, Calendar, ExternalLink } from 'lucide-react';
import { toast } from '@/lib/hooks/useToast';
import { createAnnouncement, getAllAnnouncements, deleteAnnouncement } from '@/lib/actions/announcements';
import { Announcement } from '@/types';
import { formatDistanceToNow } from 'date-fns';
import { fr } from 'date-fns/locale';

export default function AdminAnnouncementsPage() {
  const [announcements, setAnnouncements] = useState<Announcement[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreate, setShowCreate] = useState(false);

  useEffect(() => {
    fetchAnnouncements();
  }, []);

  const fetchAnnouncements = async () => {
    try {
      setLoading(true);
      const data = await getAllAnnouncements();
      setAnnouncements(data);
    } catch (error) {
      console.error('Erreur chargement annonces:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Supprimer cette annonce ?')) return;
    try {
      await deleteAnnouncement(id);
      toast({ title: 'Annonce supprimée' });
      fetchAnnouncements();
    } catch (error: any) {
      toast({ title: 'Erreur', description: error.message, variant: 'destructive' });
    }
  };

  return (
    <div className="container max-w-4xl mx-auto py-8 px-4">
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-3xl font-bold flex items-center gap-3">
            <Megaphone className="h-8 w-8 text-brand" />
            Annonces
          </h1>
          <p className="text-muted-foreground mt-2">Gérez les annonces visibles sur la page événements</p>
        </div>
        <Button onClick={() => setShowCreate(true)} className="gap-2">
          <Plus className="h-4 w-4" />
          Créer une annonce
        </Button>
      </div>

      {showCreate ? (
        <CreateAnnouncementForm
          onClose={() => { setShowCreate(false); fetchAnnouncements(); }}
        />
      ) : (
        <div className="space-y-4">
          {loading ? (
            <div className="flex justify-center py-12">
              <Loader2 className="h-8 w-8 animate-spin" />
            </div>
          ) : announcements.length === 0 ? (
            <Card>
              <CardContent className="flex flex-col items-center justify-center py-12">
                <Megaphone className="h-12 w-12 text-muted-foreground mb-4" />
                <p className="text-muted-foreground">Aucune annonce</p>
              </CardContent>
            </Card>
          ) : (
            announcements.map((announcement) => (
              <AnnouncementCard
                key={announcement.id}
                announcement={announcement}
                onDelete={handleDelete}
              />
            ))
          )}
        </div>
      )}
    </div>
  );
}

function AnnouncementCard({ announcement, onDelete }: { announcement: Announcement; onDelete: (id: string) => void }) {
  return (
    <Card>
      <CardContent className="p-4">
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <div className="flex items-center gap-2 mb-2">
              <h3 className="font-semibold">{announcement.title}</h3>
              <Badge variant={announcement.status === 'active' ? 'default' : 'secondary'}>
                {announcement.status === 'active' ? 'Active' : announcement.status}
              </Badge>
              <Badge variant="outline">{announcement.type}</Badge>
            </div>
            {announcement.description && (
              <p className="text-sm text-muted-foreground mb-2">{announcement.description}</p>
            )}
            <div className="flex items-center gap-4 text-xs text-muted-foreground">
              <span className="flex items-center gap-1">
                <Calendar className="h-3 w-3" />
                Créée {formatDistanceToNow(new Date(announcement.created_at), { addSuffix: true, locale: fr })}
              </span>
              {announcement.quiz && (
                <span className="flex items-center gap-1">
                  <ExternalLink className="h-3 w-3" />
                  Quiz: {announcement.quiz.title}
                </span>
              )}
            </div>
          </div>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onDelete(announcement.id)}
            className="text-destructive hover:text-destructive"
          >
            <Trash className="h-4 w-4" />
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

function CreateAnnouncementForm({ onClose }: { onClose: () => void }) {
  const [loading, setLoading] = useState(false);
  const [formData, setFormData] = useState({
    title: '',
    description: '',
    image_url: '',
    quiz_id: '',
    type: 'quiz' as 'quiz' | 'event' | 'news',
    ends_at: '',
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!formData.title.trim()) return;

    try {
      setLoading(true);
      await createAnnouncement({
        title: formData.title,
        description: formData.description || undefined,
        image_url: formData.image_url || undefined,
        quiz_id: formData.quiz_id || undefined,
        type: formData.type,
        ends_at: formData.ends_at || undefined,
      });
      toast({ title: 'Annonce créée' });
      onClose();
    } catch (error: any) {
      toast({ title: 'Erreur', description: error.message, variant: 'destructive' });
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Créer une annonce</CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="text-sm font-medium mb-1 block">Titre *</label>
            <Input
              value={formData.title}
              onChange={(e) => setFormData({ ...formData, title: e.target.value })}
              placeholder="Titre de l'annonce"
              required
            />
          </div>

          <div>
            <label className="text-sm font-medium mb-1 block">Description</label>
            <Textarea
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              placeholder="Description de l'annonce"
              rows={3}
            />
          </div>

          <div>
            <label className="text-sm font-medium mb-1 block">URL de l'image</label>
            <Input
              value={formData.image_url}
              onChange={(e) => setFormData({ ...formData, image_url: e.target.value })}
              placeholder="https://..."
            />
          </div>

          <div>
            <label className="text-sm font-medium mb-1 block">ID du quiz (optionnel)</label>
            <Input
              value={formData.quiz_id}
              onChange={(e) => setFormData({ ...formData, quiz_id: e.target.value })}
              placeholder="UUID du quiz"
            />
          </div>

          <div>
            <label className="text-sm font-medium mb-1 block">Type</label>
            <select
              value={formData.type}
              onChange={(e) => setFormData({ ...formData, type: e.target.value as any })}
              className="w-full p-3 rounded-lg bg-dark-surface border border-dark-border text-white focus:border-brand focus:outline-none"
            >
              <option value="quiz">Quiz</option>
              <option value="event">Événement</option>
              <option value="news">Actualité</option>
            </select>
          </div>

          <div>
            <label className="text-sm font-medium mb-1 block">Date de fin (optionnel)</label>
            <Input
              type="datetime-local"
              value={formData.ends_at}
              onChange={(e) => setFormData({ ...formData, ends_at: e.target.value })}
            />
          </div>

          <div className="flex justify-end gap-3">
            <Button type="button" variant="outline" onClick={onClose}>
              Annuler
            </Button>
            <Button type="submit" disabled={loading || !formData.title.trim()}>
              {loading ? <Loader2 className="h-4 w-4 animate-spin mr-2" /> : null}
              Créer l'annonce
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}