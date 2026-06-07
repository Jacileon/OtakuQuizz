// ============================================================
// PAGE ÉVÉNEMENTS
// ============================================================

import Link from '../../../../node_modules/next/link';
import { getActiveEvents, getUpcomingEvents, getPastEvents } from '@/lib/queries/events';
import { EventCard } from '@/components/events/EventCard';
import { LiveEventBanner } from '@/components/events/LiveEventBanner';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Zap, Calendar, History } from 'lucide-react';

export default async function EventsPage() {
  const [active, upcoming, past] = await Promise.all([
    getActiveEvents(),
    getUpcomingEvents(),
    getPastEvents(1),
  ]);

  return (
    <div className="p-4 lg:p-8 pb-24 md:pb-8">
      <div className="max-w-4xl mx-auto space-y-8">
        <h1 className="font-display text-3xl tracking-wider flex items-center gap-3">
          <Zap className="h-8 w-8 text-brand" />
          ÉVÉNEMENTS
        </h1>

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
