'use client';

import { useState, useRef } from 'react';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Camera, Loader2 } from 'lucide-react';
import { toast } from '@/lib/hooks/useToast';
import { uploadAvatar } from '@/lib/actions/media';
import { updateProfile } from '@/lib/auth/actions';
import { getInitials } from '@/lib/utils';

interface AvatarUploadProps {
  currentAvatarUrl: string | null;
  username: string;
  onUpload: (url: string) => void;
}

export function AvatarUpload({ currentAvatarUrl, username, onUpload }: AvatarUploadProps) {
  const [uploading, setUploading] = useState(false);
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleFileSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    // Vérifier le type de fichier
    if (!file.type.startsWith('image/')) {
      toast({ title: 'Erreur', description: 'Veuillez sélectionner une image', variant: 'destructive' });
      return;
    }

    // Vérifier la taille (max 2MB)
    if (file.size > 2 * 1024 * 1024) {
      toast({ title: 'Erreur', description: 'L\'image ne doit pas dépasser 2 Mo', variant: 'destructive' });
      return;
    }

    try {
      setUploading(true);

      // Créer un aperçu local
      const reader = new FileReader();
      reader.onload = (e) => {
        setPreviewUrl(e.target?.result as string);
      };
      reader.readAsDataURL(file);

      // Upload vers Cloudinary
      const result = await uploadAvatar(file);

      if ('error' in result) {
        throw new Error(result.error);
      }

      // Mettre à jour le profil avec l'URL Cloudinary
      await updateProfile({ avatar_url: result.url } as any);
      
      onUpload(result.url);
      toast({ title: 'Avatar mis à jour' });
    } catch (error: any) {
      console.error('Erreur upload avatar:', error);
      toast({ title: 'Erreur', description: error.message || 'Impossible de mettre à jour l\'avatar', variant: 'destructive' });
      setPreviewUrl(null);
    } finally {
      setUploading(false);
    }
  };

  return (
    <div className="flex flex-col items-center gap-4">
      <div className="relative">
        <Avatar className="h-24 w-24">
          <AvatarImage src={previewUrl || currentAvatarUrl || undefined} />
          <AvatarFallback className="text-2xl bg-dark-surface">
            {getInitials(username)}
          </AvatarFallback>
        </Avatar>
        <button
          type="button"
          onClick={() => fileInputRef.current?.click()}
          disabled={uploading}
          className="absolute bottom-0 right-0 h-8 w-8 rounded-full bg-brand flex items-center justify-center hover:bg-brand/80 transition-colors disabled:opacity-50"
        >
          {uploading ? (
            <Loader2 className="h-4 w-4 text-white animate-spin" />
          ) : (
            <Camera className="h-4 w-4 text-white" />
          )}
        </button>
        <input
          ref={fileInputRef}
          type="file"
          accept="image/*"
          onChange={handleFileSelect}
          className="hidden"
        />
      </div>
      <p className="text-xs text-muted-foreground">
        Cliquez sur l'appareil photo pour changer l'avatar
      </p>
    </div>
  );
}