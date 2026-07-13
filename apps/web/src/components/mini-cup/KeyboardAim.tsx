"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { ArrowUp, ArrowDown, ArrowLeft, ArrowRight, Circle } from "lucide-react";

interface Props {
  onShoot: (dirX: number, dirY: number, power: number) => void;
}

const DIRECTIONS = [
  { label: "↖", x: -0.7, y: -0.7, icon: ArrowUp },
  { label: "↑", x: 0, y: -1, icon: ArrowUp },
  { label: "↗", x: 0.7, y: -0.7, icon: ArrowUp },
  { label: "←", x: -1, y: 0, icon: ArrowLeft },
  { label: "•", x: 0, y: -1, icon: Circle },
  { label: "→", x: 1, y: 0, icon: ArrowRight },
  { label: "↙", x: -0.7, y: 0.7, icon: ArrowDown },
  { label: "↓", x: 0, y: 0.7, icon: ArrowDown },
  { label: "↘", x: 0.7, y: 0.7, icon: ArrowDown },
];

export function KeyboardAim({ onShoot }: Props) {
  const [dir, setDir] = useState({ x: 0, y: -1 });
  const [power, setPower] = useState(0.8);

  const selectedDir = DIRECTIONS.find((d) => d.x === dir.x && d.y === dir.y) || DIRECTIONS[1];

  return (
    <div className="p-4 bg-muted/80 rounded-xl border animate-in slide-in-from-bottom-2">
      <div className="grid grid-cols-3 gap-2 max-w-[180px] mx-auto mb-4">
        {DIRECTIONS.map((d, i) => (
          <Button
            key={i}
            variant={dir.x === d.x && dir.y === d.y ? "default" : "outline"}
            size="sm"
            className={cn("h-12 text-lg font-bold", dir.x === d.x && dir.y === d.y && "bg-green-600 hover:bg-green-700")}
            onClick={() => setDir({ x: d.x, y: d.y })}
          >
            {d.label}
          </Button>
        ))}
      </div>

      <div className="flex flex-col sm:flex-row items-center gap-4 justify-center">
        <div className="flex items-center gap-3 w-full sm:w-auto">
          <span className="text-sm font-medium whitespace-nowrap">Puissance</span>
          <input
            type="range"
            min="0.3"
            max="1.8"
            step="0.1"
            value={power}
            onChange={(e) => setPower(parseFloat(e.target.value))}
            className="w-full sm:w-32 accent-green-600"
          />
          <span className="text-sm font-mono w-8 text-right">{Math.round(power * 100)}%</span>
        </div>
        <Button
          className="w-full sm:w-auto bg-green-600 hover:bg-green-700 text-white font-bold"
          onClick={() => onShoot(dir.x, dir.y, power)}
        >
          TIRER !
        </Button>
      </div>
    </div>
  );
}
