import Link from 'next/link';
import { createClient } from '@/lib/supabase/server';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Trophy, Shield, Calendar, Clock, Users, ArrowRight } from 'lucide-react';

export default async function OfficialLeaderboardPage() {
  const supabase = createClient();

  const { data: quizzes } = await supabase
    .from('quizzes')
    .select('id, title, description, play_count, question_count, starts_at, ends_at, status')
    .eq('quiz_type', 'official')
    .neq('status', 'deleted')
    .order('starts_at', { ascending: false });

  return (
    <div className="p-4 lg:p-8 pb-24 md:pb-8">
      <div className="max-w-4xl mx-auto">
        <div className="mb-8">
          <h1 className="font-display text-3xl tracking-wider flex items-center gap-3">
            <Trophy className="h-8 w-8 text-yellow-500" />
            CLASSEMENTS OFFICIELS
          </h1>
          <p className="text-muted-foreground mt-2">
            Consultez les classements des quiz officiels
          </p>
        </div>

        {quizzes && quizzes.length > 0 ? (
          <div className="space-y-4">
            {quizzes.map((quiz) => (
              <Link key={quiz.id} href={`/leaderboard/official/${quiz.id}`}>
                <Card className="hover:border-yellow-500/30 transition-all cursor-pointer">
                  <CardContent className="p-4">
                    <div className="flex items-center gap-4">
                      <div className="h-12 w-12 rounded-lg bg-gradient-to-br from-yellow-500 to-amber-500 flex items-center justify-center shrink-0">
                        <Trophy className="h-6 w-6 text-white" />
                      </div>
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2 mb-1">
                          <h3 className="font-semibold truncate">{quiz.title}</h3>
                          <Badge variant={quiz.status === 'active' ? 'default' : 'secondary'}>
                            {quiz.status === 'active' ? 'Actif' : quiz.status === 'archived' ? 'Archivé' : quiz.status}
                          </Badge>
                        </div>
                        {quiz.description && (
                          <p className="text-sm text-muted-foreground line-clamp-1">{quiz.description}</p>
                        )}
                        <div className="flex items-center gap-4 mt-2 text-xs text-muted-foreground">
                          <span className="flex items-center gap-1">
                            <Users className="h-3 w-3" />
                            {quiz.play_count} parties
                          </span>
                          <span>{quiz.question_count} questions</span>
                          {quiz.starts_at && (
                            <span className="flex items-center gap-1">
                              <Calendar className="h-3 w-3" />
                              {new Date(quiz.starts_at).toLocaleDateString('fr-FR')}
                            </span>
                          )}
                        </div>
                      </div>
                      <ArrowRight className="h-5 w-5 text-muted-foreground shrink-0" />
                    </div>
                  </CardContent>
                </Card>
              </Link>
            ))}
          </div>
        ) : (
          <Card>
            <CardContent className="flex flex-col items-center justify-center py-16">
              <Trophy className="h-16 w-16 text-muted-foreground mb-4 opacity-50" />
              <h3 className="text-lg font-semibold mb-2">Aucun quiz officiel</h3>
              <p className="text-muted-foreground text-center">
                Les classements officiels apparaîtront ici quand des quiz seront créés
              </p>
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  );
}