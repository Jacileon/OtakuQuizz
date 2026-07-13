"use client";

import { Button } from "@/components/ui/button";
import { Trophy, RotateCcw, Home, Star } from "lucide-react";

interface TeamInfo {
  name: string;
  flag: string;
  primary: string;
}

interface Props {
  message: string;
  scoreA: number;
  scoreB: number;
  teamAInfo: TeamInfo;
  teamBInfo: TeamInfo;
  xpEarned: number;
  onReplay: () => void;
  onQuit: () => void;
}

export function GameOverScreen({
  message,
  scoreA,
  scoreB,
  teamAInfo,
  teamBInfo,
  xpEarned,
  onReplay,
  onQuit,
}: Props) {
  const winner = scoreA > scoreB ? teamAInfo : teamBInfo;
  const isPerfect = scoreA === 5 && scoreB === 0;

  return (
    <div className="absolute inset-0 flex flex-col items-center justify-center bg-black/70 backdrop-blur-md z-30 animate-in zoom-in duration-500">
      <div className="text-center space-y-4 px-6 max-w-sm">
        {/* Trophy */}
        <div className="mx-auto w-20 h-20 bg-yellow-500 rounded-full flex items-center justify-center shadow-xl animate-bounce">
          <Trophy className="w-10 h-10 text-white" />
        </div>

        <h2 className="text-2xl font-black text-white">{message}</h2>

        {/* Score final */}
        <div className="flex items-center justify-center gap-4 bg-white/10 rounded-xl p-4">
          <div className="text-center">
            <div className="text-3xl">{teamAInfo.flag}</div>
            <div className="text-white font-bold text-xl">{scoreA}</div>
          </div>
          <div className="text-white/50 text-lg">-</div>
          <div className="text-center">
            <div className="text-3xl">{teamBInfo.flag}</div>
            <div className="text-white font-bold text-xl">{scoreB}</div>
          </div>
        </div>

        {/* XP */}
        <div className="flex items-center justify-center gap-2 text-yellow-300 font-bold text-lg">
          <Star className="w-5 h-5 fill-yellow-300" />
          +{xpEarned} XP
        </div>

        {isPerfect && (
          <div className="bg-green-500/20 text-green-300 px-3 py-1 rounded-full text-sm font-bold border border-green-500/30">
            🏅 Séance parfaite 5/5 !
          </div>
        )}

        {/* Actions */}
        <div className="grid grid-cols-2 gap-3 pt-2">
          <Button variant="outline" className="gap-2" onClick={onReplay}>
            <RotateCcw className="w-4 h-4" />
            Rejouer
          </Button>
          <Button className="gap-2 bg-green-600 hover:bg-green-700" onClick={onQuit}>
            <Home className="w-4 h-4" />
            Menu
          </Button>
        </div>
      </div>
    </div>
  );
}
