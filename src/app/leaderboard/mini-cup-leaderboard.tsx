// À ajouter dans src/app/leaderboard/page.tsx (ou votre composant d'onglets existant)

"use client";

import { useState, useEffect } from "react";
import { getMiniCupLeaderboard } from "@/lib/actions/mini-cup";
import { Trophy, Target, Flame } from "lucide-react";
import { cn } from "@/lib/utils";

// Types
interface MiniCupLeaderboardEntry {
  id: string;
  user_id: string;
  total_goals: number;
  total_shots: number;
  perfect_sessions: number;
  best_streak: number;
  updated_at: string;
}

// Composant à intégrer comme nouvel onglet dans votre page leaderboard
export function MiniCupLeaderboardTab() {
  const [entries, setEntries] = useState<MiniCupLeaderboardEntry[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getMiniCupLeaderboard(50)
      .then((data) => setEntries(data || []))
      .catch(console.error)
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="w-8 h-8 border-4 border-green-600 border-t-transparent rounded-full animate-spin" />
      </div>
    );
  }

  if (entries.length === 0) {
    return (
      <div className="text-center py-12 text-muted-foreground">
        <Target className="w-12 h-12 mx-auto mb-3 opacity-30" />
        <p>Aucune partie jouée pour le moment.</p>
        <p className="text-sm">Soyez le premier à jouer à Mini Cup !</p>
      </div>
    );
  }

  return (
    <div className="space-y-2">
      {entries.map((entry, i) => {
        const accuracy = entry.total_shots > 0 ? Math.round((entry.total_goals / entry.total_shots) * 100) : 0;
        return (
          <div
            key={entry.id}
            className={cn(
              "flex items-center gap-3 p-3 rounded-lg border transition-colors",
              i === 0 && "bg-yellow-50 border-yellow-200",
              i === 1 && "bg-gray-50 border-gray-200",
              i === 2 && "bg-orange-50 border-orange-200",
              i > 2 && "bg-card hover:bg-muted/50"
            )}
          >
            {/* Rank */}
            <div className="w-8 h-8 flex items-center justify-center font-black text-sm">
              {i === 0 ? (
                <Trophy className="w-5 h-5 text-yellow-500" />
              ) : i === 1 ? (
                <Trophy className="w-5 h-5 text-gray-400" />
              ) : i === 2 ? (
                <Trophy className="w-5 h-5 text-orange-400" />
              ) : (
                <span className="text-muted-foreground">{i + 1}</span>
              )}
            </div>

            {/* User */}
            <div className="flex-1 min-w-0">
              <div className="font-semibold text-sm truncate">
                Joueur {entry.user_id.slice(0, 8)}
              </div>
              <div className="flex items-center gap-2 text-xs text-muted-foreground">
                <span>{entry.total_goals} buts</span>
                <span>•</span>
                <span>{entry.total_shots} tirs</span>
                <span>•</span>
                <span>{accuracy}% précision</span>
              </div>
            </div>

            {/* Stats */}
            <div className="flex items-center gap-3 text-xs">
              {entry.perfect_sessions > 0 && (
                <div className="flex items-center gap-1 text-green-600 font-medium">
                  <Target className="w-3.5 h-3.5" />
                  {entry.perfect_sessions}
                </div>
              )}
              {entry.best_streak > 2 && (
                <div className="flex items-center gap-1 text-orange-500 font-medium">
                  <Flame className="w-3.5 h-3.5" />
                  {entry.best_streak}
                </div>
              )}
            </div>
          </div>
        );
      })}
    </div>
  );
}

// Usage dans votre page leaderboard existante :
// Ajoutez un onglet "Mini Cup" qui rend <MiniCupLeaderboardTab />
