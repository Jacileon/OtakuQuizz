'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { completeProfile } from '@/lib/auth/actions';
import { toast } from '@/lib/hooks/useToast';
import { Loader2, User, ArrowRight } from 'lucide-react';

export default function CompleteProfilePage() {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [nickname, setNickname] = useState('');
  const [country, setCountry] = useState('');

  const AFRICAN_COUNTRIES = [
    'Algérie', 'Angola', 'Bénin', 'Botswana', 'Burkina Faso', 'Burundi',
    'Cabo Verde', 'Cameroun', 'Comores', 'Congo (Brazzaville)', 'Congo (Kinshasa)',
    'Côte d\'Ivoire', 'Djibouti', 'Égypte', 'Érythrée', 'Eswatini', 'Éthiopie',
    'Gabon', 'Gambie', 'Ghana', 'Guinée', 'Guinée-Bissau', 'Guinée équatoriale',
    'Kenya', 'Lesotho', 'Liberia', 'Libye', 'Madagascar', 'Malawi', 'Mali',
    'Maroc', 'Maurice', 'Mauritanie', 'Mozambique', 'Namibie', 'Niger', 'Nigeria',
    'Ouganda', 'République centrafricaine', 'Rwanda', 'Sao Tomé-et-Principe',
    'Sénégal', 'Seychelles', 'Sierra Leone', 'Somalie', 'Soudan', 'Soudan du Sud',
    'Tanzanie', 'Tchad', 'Togo', 'Tunisie', 'Zambie', 'Zimbabwe', 'Autre'
  ];

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!nickname.trim() || nickname.length < 2) {
      toast({ title: 'Erreur', description: 'Le surnom doit faire au moins 2 caractères', variant: 'destructive' });
      return;
    }

    try {
      setLoading(true);
      const result = await completeProfile({ nickname, country });
      if (result.success) {
        toast({ title: 'Profil complété !', description: 'Bienvenue sur Otaku Quiz Africa !' });
        router.push('/dashboard');
      } else {
        toast({ title: 'Erreur', description: result.error, variant: 'destructive' });
      }
    } catch (error: any) {
      toast({ title: 'Erreur', description: error.message, variant: 'destructive' });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-dark p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <div className="mx-auto h-16 w-16 rounded-full bg-brand/10 flex items-center justify-center mb-4">
            <User className="h-8 w-8 text-brand" />
          </div>
          <CardTitle className="text-2xl">Complète ton profil</CardTitle>
          <p className="text-muted-foreground">
            Choisis un surnom pour commencer à jouer
          </p>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="text-sm font-medium mb-1 block">Surnom / Nickname *</label>
              <Input
                value={nickname}
                onChange={(e) => setNickname(e.target.value)}
                placeholder="Ton surnom affiché"
                required
                minLength={2}
                maxLength={30}
                autoFocus
              />
              <p className="text-xs text-muted-foreground mt-1">
                C'est ce nom qui sera affiché partout
              </p>
            </div>

            <div>
              <label className="text-sm font-medium mb-1 block">Pays</label>
              <select
                value={country}
                onChange={(e) => setCountry(e.target.value)}
                className="w-full p-3 rounded-lg bg-dark-surface border border-dark-border text-white focus:border-brand focus:outline-none"
              >
                <option value="">Sélectionner un pays...</option>
                {AFRICAN_COUNTRIES.map((c) => (
                  <option key={c} value={c}>{c}</option>
                ))}
              </select>
            </div>

            <Button type="submit" className="w-full gap-2" disabled={loading || nickname.length < 2}>
              {loading ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <>
                  Commencer <ArrowRight className="h-4 w-4" />
                </>
              )}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}