// ============================================================
// PAGE CLASSEMENT PAR QUIZ
// ============================================================

import { getQuizById } from '@/lib/queries/quizzes';
import { getQuizLeaderboard } from '@/lib/queries/leaderboards';
import { LeaderboardTable } from '@/components/leaderboard/LeaderboardTable';
import { Podium } from '@/components/leaderboard/Podium';
import { Card, CardContent } from '@/components/ui/card';
import { Trophy, ArrowLeft } from 'lucide-react';
import Link from '../../../../../../node_modules/next/link';
import { Button } from '@/components/ui/button';

export default async function QuizLeaderboardPage({ params }: { params: { id: string } }) {
  const [quiz, leaderboard] = await Promise.all([
    getQuizById(params.id),
    getQuizLeaderboard(params.id),
  ]);

  if (!quiz) {
    return (
      <div className="p-4 lg:p-8">
        <p className="text-muted-foreground">Quiz non trouvé</p>
      </div>
    );
  }

  return (
    <div className="p-4 lg:p-8 pb-24 md:pb-8">
      <div className="max-w-4xl mx-auto space-y-8">
        <div className="flex items-center gap-4">
          <Link href={`/quiz/${params.id}/play`}>
            <Button variant="ghost" size="icon">
              <ArrowLeft className="h-5 w-5" />
            </Button>
          </Link>
          <div>
            <h1 className="font-display text-3xl tracking-wider flex items-center gap-3">
              <Trophy className="h-8 w-8 text-brand" />
              CLASSEMENT
            </h1>
            <p className="text-muted-foreground">{quiz.title}</p>
          </div>
        </div>

        <Card className="border-dark-border bg-dark-card">
          <CardContent className="p-6">
            <div className="grid grid-cols-3 gap-4 text-center">
              <div>
                <div className="font-display text-2xl text-brand">{quiz.play_count}</div>
                <div className="text-xs text-muted-foreground">Joueurs</div>
              </div>
              <div>
                <div className="font-display text-2xl text-accent">{quiz.question_count}</div>
                <div className="text-xs text-muted-foreground">Questions</div>
              </div>
              <div>
                <div className="font-display text-2xl text-green-400">{quiz.series}</div>
                <div className="text-xs text-muted-foreground">Série</div>
              </div>
            </div>
          </CardContent>
        </Card>

        {leaderboard.length > 0 && <Podium entries={leaderboard.slice(0, 3)} />}
        <LeaderboardTable entries={leaderboard} type="quiz" />
      </div>
    </div>
  );
}
