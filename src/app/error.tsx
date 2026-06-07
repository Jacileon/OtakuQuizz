'use client';
import { useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { AlertTriangle, RotateCcw } from 'lucide-react';

export default function ErrorPage({ error, reset }: { error: Error & { digest?: string }; reset: () => void }) {
  useEffect(() => {
    console.error(error);
  }, [error]);

  return (
    <div className="min-h-screen bg-dark flex items-center justify-center p-4">
      <Card className="border-dark-border bg-dark-card max-w-md w-full">
        <CardContent className="p-8 text-center space-y-4">
          <AlertTriangle className="h-12 w-12 mx-auto text-red-400" />
          <h1 className="font-display text-2xl tracking-wider">OUPS !</h1>
          <p className="text-muted-foreground">Une erreur est survenue. Notre équipe a été notifiée.</p>
          {error.digest && <p className="text-xs text-muted-foreground">Ref: {error.digest}</p>}
          <Button onClick={reset} className="gap-2">
            <RotateCcw className="h-4 w-4" /> Réessayer
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}

