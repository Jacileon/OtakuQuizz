'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { updateProfile } from '@/lib/auth/actions';
import { UserProfile } from '@/types';
import { toast } from '@/lib/hooks/useToast';
import { getInitials } from '@/lib/utils';
import { Camera, User, Save, Loader2, Lock } from 'lucide-react';
import { AvatarUpload } from './AvatarUpload';

const AFRICAN_COUNTRIES = [
  'Algérie',
  'Angola',
  'Bénin',
  'Botswana',
  'Burkina Faso',
  'Burundi',
  'Cabo Verde',
  'Cameroun',
  'Comores',
  'Congo (Brazzaville)',
  'Congo (Kinshasa)',
  'Côte d\'Ivoire',
  'Djibouti',
  'Égypte',
  'Érythrée',
  'Eswatini',
  'Éthiopie',
  'Gabon',
  'Gambie',
  'Ghana',
  'Guinée',
  'Guinée-Bissau',
  'Guinée équatoriale',
  'Kenya',
  'Lesotho',
  'Liberia',
  'Libye',
  'Madagascar',
  'Malawi',
  'Mali',
  'Maroc',
  'Maurice',
  'Mauritanie',
  'Mozambique',
  'Namibie',
  'Niger',
  'Nigeria',
  'Ouganda',
  'République centrafricaine',
  'Rwanda',
  'Sao Tomé-et-Principe',
  'Sénégal',
  'Seychelles',
  'Sierra Leone',
  'Somalie',
  'Soudan',
  'Soudan du Sud',
  'Tanzanie',
  'Tchad',
  'Togo',
  'Tunisie',
  'Zambie',
  'Zimbabwe',
  'Autre'
];

interface ProfileEditFormProps {
  profile: UserProfile;
}

export function ProfileEditForm({ profile }: ProfileEditFormProps) {
  const router = useRouter();
  const [isSubmitting, setIsSubmitting] = useState(false);
  
  const [username, setUsername] = useState(profile.username);
  const [nickname, setNickname] = useState(profile.nickname || '');
  const [bio, setBio] = useState(profile.bio || '');
  const [country, setCountry] = useState(profile.country || '');
  const [phone, setPhone] = useState(profile.phone || '');
  const [favoriteAnime, setFavoriteAnime] = useState(profile.favorite_anime || '');
  const [avatarUrl, setAvatarUrl] = useState(profile.avatar_url);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);

    const result = await updateProfile({
      nickname,
      bio,
      country,
      phone,
      favorite_anime: favoriteAnime,
    });

    if (result.success) {
      toast({
        title: 'Succès',
        description: 'Profil mis à jour avec succès !',
        variant: 'default',
      });
      router.refresh();
    } else {
      toast({
        title: 'Erreur',
        description: result.error || 'Erreur lors de la mise à jour',
        variant: 'destructive',
      });
    }

    setIsSubmitting(false);
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <Card className="border-dark-border bg-dark-card">
        <CardContent className="p-6 space-y-6">
          {/* Avatar */}
          <AvatarUpload
            currentAvatarUrl={avatarUrl}
            username={profile.username}
            onUpload={(url) => setAvatarUrl(url)}
          />

          {/* Champs du formulaire */}
          <div className="space-y-4">
            {/* Username (non modifiable) */}
            <div>
              <label className="text-sm font-medium mb-1 block">Username</label>
              <Input
                value={username}
                className="bg-dark-surface opacity-60 cursor-not-allowed"
                disabled
                readOnly
              />
              <p className="text-xs text-muted-foreground mt-1">
                Le username ne peut pas être modifié
              </p>
            </div>

            {/* Nickname */}
            <div>
              <label className="text-sm font-medium mb-1 block">Surnom / Nickname *</label>
              <Input
                value={nickname}
                onChange={(e) => setNickname(e.target.value)}
                placeholder="Ton surnom affiché"
                className="bg-dark-surface"
                required
                minLength={2}
                maxLength={30}
              />
              <p className="text-xs text-muted-foreground mt-1">
                C'est ce nom qui sera affiché publiquement
              </p>
            </div>

            {/* Bio */}
            <div>
              <label className="text-sm font-medium mb-1 block">Bio</label>
              <textarea
                value={bio}
                onChange={(e) => setBio(e.target.value)}
                placeholder="Parle-nous de toi..."
                className="w-full p-3 rounded-lg bg-dark-surface border border-dark-border text-white placeholder:text-muted-foreground focus:border-brand focus:outline-none resize-none"
                rows={3}
                maxLength={200}
              />
              <p className="text-xs text-muted-foreground mt-1">{bio.length}/200</p>
            </div>

            {/* Pays */}
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

            {/* Téléphone */}
            <div>
              <label className="text-sm font-medium mb-1 block">Numéro de téléphone</label>
              {profile.phone ? (
                <div className="relative">
                  <Input
                    type="tel"
                    value={profile.phone}
                    className="bg-dark-surface opacity-60 cursor-not-allowed"
                    disabled
                    readOnly
                  />
                  <div className="absolute right-3 top-1/2 -translate-y-1/2">
                    <Lock className="h-4 w-4 text-muted-foreground" />
                  </div>
                </div>
              ) : (
                <Input
                  type="tel"
                  value={phone}
                  onChange={(e) => setPhone(e.target.value)}
                  placeholder="Ex: +225 07 07 07 07"
                  className="bg-dark-surface"
                />
              )}
              {profile.phone ? (
                <p className="text-xs text-orange-500 mt-1 flex items-center gap-1">
                  <Lock className="h-3 w-3" />
                  Pour modifier votre numéro, contactez le support.
                </p>
              ) : (
                <p className="text-xs text-muted-foreground mt-1">Avec l'indicatif pays</p>
              )}
            </div>

            {/* Anime préféré */}
            <div>
              <label className="text-sm font-medium mb-1 block">Anime préféré</label>
              <Input
                value={favoriteAnime}
                onChange={(e) => setFavoriteAnime(e.target.value)}
                placeholder="Ex: Naruto, One Piece..."
                className="bg-dark-surface"
              />
            </div>
          </div>

          {/* Bouton de sauvegarde */}
          <Button type="submit" disabled={isSubmitting} className="w-full gap-2">
            {isSubmitting ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              <Save className="h-4 w-4" />
            )}
            {isSubmitting ? 'Sauvegarde...' : 'Sauvegarder les modifications'}
          </Button>
        </CardContent>
      </Card>
    </form>
  );
}