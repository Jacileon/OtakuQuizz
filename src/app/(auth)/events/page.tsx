// ============================================================
// PAGE ÉVÉNEMENTS
// ============================================================

import Link from '../../../../node_modules/next/link';
import { getActiveEvents, getUpcomingEvents, getPastEvents } from '@/lib/queries/events';
import { EventCard } from '@/components/events/EventCard';
import { LiveEventBanner } from '@/components/events/LiveEventBanner';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Zap, Calendar, History, Trophy, Shield, ArrowRight } from 'lucide-react';
import { createClient } from '@/lib/supabase/server';
import { getCurrentUser } from '@/lib/auth/actions';

export default async function EventsPage() {
  const [active, upcoming, past, user] = await Promise.all([
    getActiveEvents(),
    getUpcomingEvents(),
    getPastEvents(1),
    getCurrentUser(),
  ]);

  // Vérifier si l'utilisateur est admin
  let isAdmin = false;
  if (user) {
    const supabase = createClient();
    const { data: profile } = await supabase
      .from('user_profiles')
      .select('is_admin')
      .eq('id', user.id)
      .single();
    isAdmin = profile?.is_admin || false;
  }

  // Récupérer les quiz officiels actifs
  const supabase = createClient();
  const { data: officialQuizzes } = await supabase
    .from('quizzes')
    .select('id, title, description, play_count, question_count, starts_at, ends_at')
    .eq('quiz_type', 'official')
    .in('status', ['active', 'scheduled'])
    .eq('is_visible', true)
    .order('starts_at', { ascending: true })
    .limit(6);

  return (
    <div className="p-4 lg:p-8 pb-24 md:pb-8">
      <div className="max-w-4xl mx-auto space-y-8">
        <h1 className="font-display text-3xl tracking-wider flex items-center gap-3">
          <Zap className="h-8 w-8 text-brand" />
          ÉVÉNEMENTS
        </h1>

        {/* Quiz Officiels */}
        <div>
          <div className="flex items-center justify-between mb-4">
            <h2 className="font-display text-xl tracking-wider flex items-center gap-2">
              <Shield className="h-5 w-5 text-yellow-500" />
              QUIZ OFFICIELS
            </h2>
            <Link href="/leaderboard/official">
              <Button variant="ghost" size="sm" className="gap-1">
                Voir tout <ArrowRight className="h-4 w-4" />
              </Button>
            </Link>
          </div>

          {officialQuizzes && officialQuizzes.length > 0 ? (
            <div className="grid md:grid-cols-2 gap-4">
              {officialQuizzes.map((quiz) => (
                <Link key={quiz.id} href={`/quiz/${quiz.id}`}>
                  <Card className="hover:border-yellow-500/30 transition-all cursor-pointer h-full">
                    <CardContent className="p-4">
                      <div className="flex items-start gap-3">
                        <div className="h-10 w-10 rounded-lg bg-gradient-to-br from-yellow-500 to-amber-500 flex items-center justify-center shrink-0">
                          <Trophy className="h-5 w-5 text-white" />
                        </div>
                        <div className="flex-1 min-w-0">
                          <h3 className="font-semibold truncate">{quiz.title}</h3>
                          {quiz.description && (
                            <p className="text-sm text-muted-foreground line-clamp-2 mt-1">
                              {quiz.description}
                            </p>
                          )}
                          <div className="flex items-center gap-3 mt-2 text-xs text-muted-foreground">
                            <span>{quiz.question_count} questions</span>
                            <span>{quiz.play_count} parties</span>
                          </div>
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                </Link>
              ))}
            </div>
          ) : (
            <Card>
              <CardContent className="p-8 text-center">
                <Trophy className="h-12 w-12 text-muted-foreground mx-auto mb-4 opacity-50" />
                <p className="text-muted-foreground">Aucun quiz officiel en cours</p>
                <p className="text-sm text-muted-foreground mt-1">
                  Revenez plus tard pour découvrir les prochains événements
                </p>
              </CardContent>
            </Card>
          )}

          {isAdmin && (
            <Link href="/admin/official-quizzes" className="mt-4 block">
              <Button variant="outline" className="w-full gap-2">
                <Shield className="h-4 w-4" />
                Gérer les quiz officiels (Admin)
              </Button>
            </Link>
          )}
        </div>

        {/* Live Banner */}
        {active.length > 0 && <LiveEventBanner events={active} />}

        {/* Active Events */}
        {active.length > 0 && (
          <div>
            <h2 className="font-display text-xl tracking-wider mb-4 flex items-center gap-2 text-brand">
              <div className="h-2 w-2 rounded-full bg-brand animate-pulse" />
              EN COURS
            </h2>
            <div className="grid md:grid-cols-2 gap-4">
              {active.map((event) => (
                <EventCard key={event.id} event={event} status="live" />
              ))}
            </div>
          </div>
        )}

        {/* Upcoming Events */}
        {upcoming.length > 0 && (
          <div>
            <h2 className="font-display text-xl tracking-wider mb-4 flex items-center gap-2 text-accent">
              <Calendar className="h-5 w-5" />
              À VENIR
            </h2>
            <div className="grid md:grid-cols-2 gap-4">
              {upcoming.map((event) => (
                <EventCard key={event.id} event={event} status="upcoming" />
              ))}
            </div>
          </div>
        )}

        {/* Past Events */}
        {past.data.length > 0 && (
          <div>
            <h2 className="font-display text-xl tracking-wider mb-4 flex items-center gap-2 text-muted-foreground">
              <History className="h-5 w-5" />
              PASSÉS
            </h2>
            <div className="space-y-3">
              {past.data.map((event) => (
                <EventCard key={event.id} event={event} status="past" />
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}