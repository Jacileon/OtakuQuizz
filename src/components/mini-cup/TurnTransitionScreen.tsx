"use client";

import { Target, Shield } from "lucide-react";
import { cn } from "@/lib/utils";

interface Props {
  teamName: string;
  teamFlag: string;
  teamColor: string;
  role: "shoot" | "defend";
  mode: string;
}

export function TurnTransitionScreen({ teamName, teamFlag, teamColor, role, mode }: Props) {
  return (
    <div className="absolute inset-0 flex items-center justify-center bg-black/60 backdrop-blur-sm z-20 animate-in fade-in duration-300">
      <div className="text-center text-white px-6">
        <div
          className="w-20 h-20 rounded-full mx-auto mb-4 flex items-center justify-center text-4xl shadow-xl border-4 border-white/20"
          style={{ backgroundColor: teamColor }}
        >
          {teamFlag}
        </div>
        <div className="text-2xl font-black mb-1">{teamName}</div>
        <div className="flex items-center justify-center gap-2 text-lg text-white/90">
          {role === "shoot" ? (
            <>
              <Target className="w-5 h-5" />
              <span>À toi de tirer !</span>
            </>
          ) : (
            <>
              <Shield className="w-5 h-5" />
              <span>Tu es gardien !</span>
            </>
          )}
        </div>
        {mode === "local_2v2" && (
          <p className="text-xs text-white/50 mt-3">
            Passe l&apos;appareil à l&apos;autre joueur
          </p>
        )}
      </div>
    </div>
  );
}
