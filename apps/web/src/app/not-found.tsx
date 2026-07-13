// ============================================================
// PAGE 404
// ============================================================

import Link from '../../node_modules/next/link';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Sword, Home } from 'lucide-react';

export default function NotFoundPage() {
  return (
    <div className="min-h-screen bg-dark flex items-center justify-center p-4">
      <Card className="border-dark-border bg-dark-card max-w-md w-full">
        <CardContent className="p-8 text-center space-y-4">
          <Sword className="h-12 w-12 mx-auto text-brand" />
          <h1 className="font-display text-4xl tracking-wider">404</h1>
          <p className="text-muted-foreground">Ce jutsu n'existe pas...</p>
          <p className="text-sm text-muted-foreground">La page que tu cherches a disparu comme un ninja de Konoha.</p>
          <Link href="/dashboard">
            <Button className="gap-2">
              <Home className="h-4 w-4" /> Retour à l'accueil
            </Button>
          </Link>
        </CardContent>
      </Card>
    </div>
  );
}

