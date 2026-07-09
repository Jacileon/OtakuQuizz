"use client";

import { useState } from "react";
import { MiniCupEngine } from "@/components/mini-cup/MiniCupEngine";
import { TeamSelector } from "@/components/mini-cup/TeamSelector";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Trophy, Users, ArrowLeft } from "lucide-react";

type GameMode = "menu" | "solo" | "local2v2";

export default function MiniCupPage() {
  const [mode, setMode] = useState<GameMode>("menu");
  const [teamA, setTeamA] = useState<string>("");
  const [teamB, setTeamB] = useState<string>("");

  if (mode === "menu") {
    return (
      <div className="container mx-auto py-8 max-w-2xl px-4">
        <Card className="border-2 border-green-100">
          <CardHeader className="text-center pb-2">
            <div className="mx-auto w-16 h-16 bg-green-600 rounded-full flex items-center justify-center mb-3">
              <Trophy className="w-8 h-8 text-white" />
            </div>
            <CardTitle className="text-3xl font-black tracking-tight">
              Mini Cup
            </CardTitle>
            <p className="text-muted-foreground mt-1">
              Dessine la trajectoire de ton tir pour marquer !
            </p>
          </CardHeader>
          <CardContent className="space-y-4 pt-2">
            <div className="grid grid-cols-1 gap-3">
              <Button
                size="lg"
                className="h-16 text-lg gap-3 bg-green-600 hover:bg-green-700"
                onClick={() => setMode("solo")}
              >
                <Trophy className="w-5 h-5" />
                Solo vs IA
              </Button>
              <Button
                size="lg"
                variant="outline"
                className="h-16 text-lg gap-3 border-2"
                onClick={() => setMode("local2v2")}
              >
                <Users className="w-5 h-5" />
                2 Joueurs local
              </Button>
            </div>
            <p className="text-xs text-center text-muted-foreground mt-4">
              Accessible clavier : activez le mode clavier dans le jeu
            </p>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (!teamA || !teamB) {
    return (
      <div className="container mx-auto py-8 max-w-2xl px-4">
        <TeamSelector
          mode={mode}
          onSelect={(a, b) => {
            setTeamA(a);
            setTeamB(b);
          }}
          onBack={() => setMode("menu")}
        />
      </div>
    );
  }

  return (
    <div className="container mx-auto py-4 max-w-4xl px-2">
      <Button
        variant="ghost"
        size="sm"
        className="mb-2"
        onClick={() => {
          setMode("menu");
          setTeamA("");
          setTeamB("");
        }}
      >
        <ArrowLeft className="w-4 h-4 mr-1" />
        Retour
      </Button>
      <MiniCupEngine
        mode={mode === "solo" ? "solo_ai" : "local_2v2"}
        teamA={teamA}
        teamB={teamB}
        onQuit={() => {
          setMode("menu");
          setTeamA("");
          setTeamB("");
        }}
      />
    </div>
  );
}
