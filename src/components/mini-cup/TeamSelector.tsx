"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ArrowLeft, Shield } from "lucide-react";
import { cn } from "@/lib/utils";

export const TEAMS = [
  { code: "fr", name: "France", flag: "🇫🇷", primary: "#002395", secondary: "#FFFFFF", accent: "#ED2939" },
  { code: "br", name: "Brésil", flag: "🇧🇷", primary: "#009C3B", secondary: "#FFDF00", accent: "#002776" },
  { code: "ar", name: "Argentine", flag: "🇦🇷", primary: "#75AADB", secondary: "#FFFFFF", accent: "#000000" },
  { code: "de", name: "Allemagne", flag: "🇩🇪", primary: "#000000", secondary: "#DD0000", accent: "#FFCE00" },
  { code: "es", name: "Espagne", flag: "🇪🇸", primary: "#AA151B", secondary: "#F1BF00", accent: "#AA151B" },
  { code: "it", name: "Italie", flag: "🇮🇹", primary: "#0066CC", secondary: "#FFFFFF", accent: "#009246" },
  { code: "pt", name: "Portugal", flag: "🇵🇹", primary: "#006600", secondary: "#FF0000", accent: "#FFFF00" },
  { code: "eng", name: "Angleterre", flag: "🏴󠁧󠁢󠁥󠁮󠁧󠁿", primary: "#FFFFFF", secondary: "#CF081F", accent: "#00247D" },
  { code: "nl", name: "Pays-Bas", flag: "🇳🇱", primary: "#FF4F00", secondary: "#FFFFFF", accent: "#1E4785" },
  { code: "be", name: "Belgique", flag: "🇧🇪", primary: "#000000", secondary: "#FDDA24", accent: "#EF3340" },
  { code: "jp", name: "Japon", flag: "🇯🇵", primary: "#FFFFFF", secondary: "#BC002D", accent: "#BC002D" },
  { code: "us", name: "USA", flag: "🇺🇸", primary: "#3C3B6E", secondary: "#FFFFFF", accent: "#B22234" },
] as const;

export type TeamCode = (typeof TEAMS)[number]["code"];

interface Props {
  mode: string;
  onSelect: (teamA: TeamCode, teamB: TeamCode) => void;
  onBack: () => void;
}

export function TeamSelector({ mode, onSelect, onBack }: Props) {
  const [step, setStep] = useState<"a" | "b">("a");
  const [selectedA, setSelectedA] = useState<TeamCode | null>(null);

  const handleSelect = (code: TeamCode) => {
    if (step === "a") {
      setSelectedA(code);
      setStep("b");
    } else if (selectedA && code !== selectedA) {
      onSelect(selectedA, code);
    }
  };

  const title = step === "a" ? "Choisis ton équipe" : "Choisis l'adversaire";
  const subtitle = step === "a" ? "Joueur 1" : mode === "local2v2" ? "Joueur 2" : "Équipe adverse (IA)";

  return (
    <Card className="border-2">
      <CardHeader className="text-center">
        <CardTitle className="text-2xl">{title}</CardTitle>
        <p className="text-sm text-muted-foreground flex items-center justify-center gap-1">
          <Shield className="w-4 h-4" />
          {subtitle}
        </p>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
          {TEAMS.map((t) => {
            const disabled = step === "b" && t.code === selectedA;
            return (
              <Button
                key={t.code}
                variant="outline"
                className={cn(
                  "h-20 flex-col gap-1 text-sm font-semibold transition-all hover:scale-105",
                  disabled && "opacity-30 cursor-not-allowed"
                )}
                style={{
                  borderColor: t.primary,
                  backgroundColor: `${t.primary}10`,
                }}
                onClick={() => handleSelect(t.code)}
                disabled={disabled}
              >
                <span className="text-2xl">{t.flag}</span>
                <span>{t.name}</span>
              </Button>
            );
          })}
        </div>
        <Button variant="ghost" className="mt-4 w-full" onClick={onBack}>
          <ArrowLeft className="w-4 h-4 mr-1" />
          Retour
        </Button>
      </CardContent>
    </Card>
  );
}
