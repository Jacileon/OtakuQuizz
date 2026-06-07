// ============================================================
// PAGE COLLECTIONS
// ============================================================

import { getCurrentUser } from '@/lib/auth/actions';
import { getUserCollections } from '@/lib/queries/social';
import { Card, CardContent } from '@/components/ui/card';
import { Progress } from '@/components/ui/progress';
import { BookOpen, Trophy } from 'lucide-react';
import { redirect } from '../../../../node_modules/next/navigation';

export default async function CollectionsPage() {
  const user = await getCurrentUser();
  if (!user) redirect('/login');

  const collections = await getUserCollections(user.id);

  return (
    <div className="p-4 lg:p-8 pb-24 md:pb-8">
      <div className="max-w-4xl mx-auto space-y-6">
        <h1 className="font-display text-3xl tracking-wider flex items-center gap-3">
          <BookOpen className="h-8 w-8 text-brand" />
          COLLECTIONS
        </h1>

        <div className="grid gap-4">
          {collections.map((col: any) => (
            <Card key={col.series} className="border-dark-border bg-dark-card">
              <CardContent className="p-6 space-y-4">
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="font-display text-xl tracking-wider">{col.series}</h3>
                    <p className="text-sm text-muted-foreground">
                      {col.completed_quizzes} / {col.total_quizzes} quiz complétés
                    </p>
                  </div>
                  {col.progress_percent === 100 && (
                    <div className="flex items-center gap-1 text-yellow-400">
                      <Trophy className="h-5 w-5" />
                      <span className="text-sm font-medium">100%</span>
                    </div>
                  )}
                </div>
                <Progress value={col.progress_percent} className="h-3" />
                <div className="flex justify-between text-xs text-muted-foreground">
                  <span>{col.progress_percent}% complété</span>
                  {col.best_score && <span>Meilleur score: {col.best_score}</span>}
                </div>
              </CardContent>
            </Card>
          ))}
          {collections.length === 0 && (
            <Card className="border-dark-border bg-dark-card">
              <CardContent className="p-8 text-center">
                <BookOpen className="h-12 w-12 mx-auto mb-4 text-muted-foreground" />
                <p className="text-muted-foreground">Aucune collection commencée</p>
                <p className="text-sm text-muted-foreground mt-1">Joue à des quiz pour commencer tes collections</p>
              </CardContent>
            </Card>
          )}
        </div>
      </div>
    </div>
  );
}

