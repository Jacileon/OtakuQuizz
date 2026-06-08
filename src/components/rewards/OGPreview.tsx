'use client';

import { useState, useEffect } from 'react';
import { Card, CardContent } from '@/components/ui/card';
import { Loader2, ExternalLink, Globe, Image as ImageIcon } from 'lucide-react';

interface OGPreviewProps {
  url: string;
  title?: string;
  image?: string;
  domain?: string;
}

export function OGPreview({ url, title, image, domain }: OGPreviewProps) {
  const [loading, setLoading] = useState(!title);
  const [preview, setPreview] = useState<{ title: string; image: string; domain: string } | null>(
    title ? { title, image: image || '', domain: domain || '' } : null
  );

  useEffect(() => {
    if (title) return;

    const fetchPreview = async () => {
      try {
        setLoading(true);
        const response = await fetch(`/api/og-preview?url=${encodeURIComponent(url)}`);
        const data = await response.json();
        setPreview(data);
      } catch (error) {
        console.error('Erreur prévisualisation:', error);
      } finally {
        setLoading(false);
      }
    };

    if (url) {
      fetchPreview();
    }
  }, [url, title]);

  if (loading) {
    return (
      <Card className="overflow-hidden">
        <CardContent className="p-4 flex items-center gap-3">
          <Loader2 className="h-4 w-4 animate-spin" />
          <span className="text-sm text-muted-foreground">Chargement de la prévisualisation...</span>
        </CardContent>
      </Card>
    );
  }

  if (!preview) {
    return null;
  }

  return (
    <Card className="overflow-hidden">
      <div className="flex">
        {preview.image ? (
          <div className="w-24 h-24 bg-dark-surface flex-shrink-0">
            <img
              src={preview.image}
              alt={preview.title}
              className="w-full h-full object-cover"
              onError={(e) => {
                (e.target as HTMLImageElement).style.display = 'none';
              }}
            />
          </div>
        ) : (
          <div className="w-24 h-24 bg-dark-surface flex-shrink-0 flex items-center justify-center">
            <ImageIcon className="h-8 w-8 text-muted-foreground" />
          </div>
        )}
        <CardContent className="flex-1 p-3 min-w-0">
          <p className="font-medium text-sm truncate">{preview.title || 'Sans titre'}</p>
          <div className="flex items-center gap-1 text-xs text-muted-foreground mt-1">
            <Globe className="h-3 w-3" />
            <span className="truncate">{preview.domain || new URL(url).hostname}</span>
          </div>
          <a
            href={url}
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center gap-1 text-xs text-brand hover:underline mt-2"
          >
            <ExternalLink className="h-3 w-3" />
            Voir le lien
          </a>
        </CardContent>
      </div>
    </Card>
  );
}