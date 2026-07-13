"use client";

import { cn } from "@/lib/utils";

interface ShotRecord {
  team: "a" | "b";
  result: "goal" | "saved" | "miss";
}

interface Props {
  shots: ShotRecord[];
}

export function ShotHistory({ shots }: Props) {
  if (shots.length === 0) return null;

  // Regrouper par paires de 2 pour affichage style TV
  const rows: ShotRecord[][] = [];
  for (let i = 0; i < shots.length; i += 2) {
    rows.push(shots.slice(i, i + 2));
  }

  return (
    <div className="flex flex-wrap gap-x-4 gap-y-1">
      {rows.map((row, rowIdx) => (
        <div key={rowIdx} className="flex items-center gap-1 bg-muted/50 rounded-full px-2 py-1">
          {row.map((s, i) => (
            <div
              key={`${rowIdx}-${i}`}
              className={cn(
                "w-3.5 h-3.5 rounded-full border border-white/20 shadow-sm transition-all",
                s.result === "goal" && "bg-green-500",
                s.result === "saved" && "bg-yellow-500",
                s.result === "miss" && "bg-red-500"
              )}
              title={s.result === "goal" ? "But" : s.result === "saved" ? "Arrêt" : "Raté"}
            />
          ))}
        </div>
      ))}
    </div>
  );
}
