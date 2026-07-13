// ============================================================
// PAGE SÉRIE
// ============================================================

import { getQuizzesBySeries } from '@/lib/queries/quizzes';
import { getCurrentUser } from '@/lib/auth/actions';
import { QuizCard } from '@/components/dashboard/QuizCard';
import { Card, CardContent } from '@/components/ui/card';
import { Progress } from '@/components/ui/progress';
import { BookOpen, Trophy } from 'lucide-react';
import { getUserCollections } from '@/lib/queries/social';

export default async function SeriesPage({ params }: { params: { series: string } }) {
  const user = await getCurrentUser();
  const result = await getQuizzesBySeries(params.series);

  let collection = null;
  if (user) {
    const collections = await getUserCollections(user.id);
    collection = collections.find((c: any) => c.series === params.series);
  }

  return (
    <div className="p-4 lg:p-8 pb-24 md:pb-8">
      <div className="max-w-6xl mx-auto space-y-8">
        <div className="relative overflow-hidden rounded-lg bg-gradient-to-r from-brand/20 to-accent/20 p-8">
          <BookOpen className="h-16 w-16 text-white/10 absolute right-4 top-4" />
          <h1 className="font-display text-4xl tracking-wider">{params.series}</h1>
          <p className="text-muted-foreground mt-2">{result.count} quiz • {result.data.reduce((sum: number, q: any) => sum + q.play_count, 0)} plays</p>
        </div>

        {collection && (
          <Card className="border-dark-border bg-dark-card">
            <CardContent className="p-6 space-y-3">
              <div className="flex items-center justify-between">
                <span className="font-medium">Ta progression</span>
                <span className="text-sm text-muted-foreground">{collection.completed_quizzes}/{collection.total_quizzes}</span>
              </div>
              <Progress value={collection.progress_percent} className="h-3" />
              {collection.progress_percent === 100 && (
                <div className="flex items-center gap-2 text-yellow-400 text-sm">
                  <Trophy className="h-4 w-4" />
                  <span>Collection complète !</span>
                </div>
              )}
            </CardContent>
          </Card>
        )}

        <div>
          <h2 className="font-display text-xl tracking-wider mb-4">QUIZ</h2>
          <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-4">
            {result.data.map((quiz: any) => (
              <QuizCard key={quiz.id} quiz={quiz} />
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
