'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { completeProfile } from '@/lib/auth/actions';
import { toast } from '@/lib/hooks/useToast';
import { Loader2, User, ArrowRight, Phone } from 'lucide-react';

export default function CompleteProfilePage() {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [nickname, setNickname] = useState('');
  const [phone, setPhone] = useState('');
  const [country, setCountry] = useState('');
  const [phoneError, setPhoneError] = useState('');

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

  const validatePhone = (value: string): boolean => {
    const cleaned = value.replace(/[\s\-]/g, '');
    const regex = /^\+[1-9]\d{7,14}$/;
    if (!regex.test(cleaned)) {
      setPhoneError('Numéro invalide. Format attendu : +XXX suivi des chiffres (ex : +22670000000)');
      return false;
    }
    setPhoneError('');
    return true;
  };

  const handlePhoneChange = (value: string) => {
    setPhone(value);
    if (value.length > 3) {
      validatePhone(value);
    } else {
      setPhoneError('');
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!nickname.trim() || nickname.length < 2) {
      toast({ title: 'Erreur', description: 'Le surnom doit faire au moins 2 caractères', variant: 'destructive' });
      return;
    }

    if (!validatePhone(phone)) {
      toast({ title: 'Erreur', description: phoneError, variant: 'destructive' });
      return;
    }

    try {
      setLoading(true);
      const cleanedPhone = phone.replace(/[\s\-]/g, '');
      const result = await completeProfile({ nickname, phone: cleanedPhone, country });
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
            Remplis ces informations pour commencer à jouer
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
              <label className="text-sm font-medium mb-1 block flex items-center gap-1">
                <Phone className="h-3 w-3" />
                Numéro de téléphone *
              </label>
              <Input
                type="tel"
                value={phone}
                onChange={(e) => handlePhoneChange(e.target.value)}
                placeholder="+22670000000"
                required
                className={phoneError ? 'border-destructive' : ''}
              />
              {phoneError ? (
                <p className="text-xs text-destructive mt-1">{phoneError}</p>
              ) : (
                <p className="text-xs text-muted-foreground mt-1">
                  Format : +XXX suivi des chiffres (ex : +22670000000)
                </p>
              )}
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

            <Button 
              type="submit" 
              className="w-full gap-2" 
              disabled={loading || nickname.length < 2 || !phone || !!phoneError}
            >
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