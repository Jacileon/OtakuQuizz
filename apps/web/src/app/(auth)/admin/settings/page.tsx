'use client';

import { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Loader2, Settings, Save, Check } from 'lucide-react';
import { toast } from '@/lib/hooks/useToast';
import { getQuizCreationConfig, updateQuizCreationConfig } from '@/lib/actions/permissions';
import { RANKS } from '@/lib/constants';

export default function AdminSettingsPage() {
  const [allowedRanks, setAllowedRanks] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    const fetchConfig = async () => {
      try {
        const config = await getQuizCreationConfig();
        setAllowedRanks(config.allowedRanks);
      } catch (error) {
        console.error('Erreur chargement config:', error);
      } finally {
        setLoading(false);
      }
    };
    fetchConfig();
  }, []);

  const toggleRank = (rank: string) => {
    setAllowedRanks(prev =>
      prev.includes(rank)
        ? prev.filter(r => r !== rank)
        : [...prev, rank]
    );
  };

  const handleSave = async () => {
    try {
      setSaving(true);
      await updateQuizCreationConfig(allowedRanks);
      toast({ title: 'Configuration sauvegardée' });
    } catch (error: any) {
      toast({ title: 'Erreur', description: error.message, variant: 'destructive' });
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div className="flex justify-center py-12">
        <Loader2 className="h-8 w-8 animate-spin" />
      </div>
    );
  }

  return (
    <div className="container max-w-2xl mx-auto py-8 px-4">
      <h1 className="text-3xl font-bold flex items-center gap-3 mb-8">
        <Settings className="h-8 w-8 text-primary" />
        Paramètres de création de quiz
      </h1>

      <Card>
        <CardHeader>
          <CardTitle>Rangs autorisés à créer des quiz</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-sm text-muted-foreground">
            Sélectionnez les rangs qui peuvent créer des quiz. Les admins et les utilisateurs avec une autorisation individuelle peuvent toujours créer des quiz.
          </p>

          <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
            {RANKS.map((rank) => (
              <button
                key={rank.rank}
                onClick={() => toggleRank(rank.rank)}
                className={`flex items-center gap-2 p-3 rounded-lg border transition-colors ${
                  allowedRanks.includes(rank.rank)
                    ? 'border-primary bg-primary/10'
                    : 'border-muted hover:border-muted-foreground/50'
                }`}
              >
                <div className={`w-5 h-5 rounded flex items-center justify-center ${
                  allowedRanks.includes(rank.rank)
                    ? 'bg-primary text-primary-foreground'
                    : 'border border-muted-foreground'
                }`}>
                  {allowedRanks.includes(rank.rank) && <Check className="h-3 w-3" />}
                </div>
                <span className="font-medium">{rank.rank}</span>
                <span className="text-xs text-muted-foreground">({rank.minXP}+ XP)</span>
              </button>
            ))}
          </div>

          <div className="pt-4 border-t">
            <Button onClick={handleSave} disabled={saving} className="gap-2">
              {saving ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
              Sauvegarder
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}