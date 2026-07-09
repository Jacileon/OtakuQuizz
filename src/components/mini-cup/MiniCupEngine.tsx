"use client";

import { useState, useCallback, useRef, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { createMiniCupSession, recordShot, finishSession } from "@/lib/actions/mini-cup";
import { ShotHistory } from "./ShotHistory";
import { TurnTransitionScreen } from "./TurnTransitionScreen";
import { KeyboardAim } from "./KeyboardAim";
import { GameOverScreen } from "./GameOverScreen";
import { TEAMS } from "./TeamSelector";
import { cn } from "@/lib/utils";
import { Trophy, RotateCcw, Keyboard, MousePointer } from "lucide-react";

type Phase = "aim" | "shot" | "result" | "transition" | "finished";

interface Props {
  mode: "solo_ai" | "local_2v2";
  teamA: string;
  teamB: string;
  onQuit: () => void;
}

interface ShotRecord {
  team: "a" | "b";
  result: "goal" | "saved" | "miss";
}

// ─── Constantes physiques ───
const GOAL_WIDTH = 220;
const GOAL_HEIGHT = 90;
const BALL_RADIUS = 10;
const KEEPER_WIDTH = 44;
const KEEPER_HEIGHT = 56;
const GRAVITY = 0.18;
const BALL_SPEED_BASE = 7;

export function MiniCupEngine({ mode, teamA, teamB, onQuit }: Props) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  const [phase, setPhase] = useState<Phase>("aim");
  const [shots, setShots] = useState<ShotRecord[]>([]);
  const [scoreA, setScoreA] = useState(0);
  const [scoreB, setScoreB] = useState(0);
  const [currentShooter, setCurrentShooter] = useState(0);
  const [currentTeam, setCurrentTeam] = useState<"a" | "b">("a");
  const [message, setMessage] = useState("");
  const [sessionId, setSessionId] = useState<string>("");
  const [showKeyboard, setShowKeyboard] = useState(false);
  const [xpEarned, setXPEarned] = useState(0);
  const [shake, setShake] = useState(false);

  const shotOrderRef = useRef(0);
  const animFrameRef = useRef<number>(0);
  const pointerActive = useRef(false);
  const trajectoryPoints = useRef<{ x: number; y: number }[]>([]);

  // ─── Ball physics ───
  const ballRef = useRef({
    x: 0, y: 0,
    vx: 0, vy: 0,
    z: 0, vz: 0,
    active: false,
  });

  // ─── Keeper AI ───
  const keeperRef = useRef({
    x: 0, // position relative au centre du but
    targetX: 0,
    delay: 0,
    diving: false,
    diveFrame: 0,
  });

  const teamAInfo = TEAMS.find((t) => t.code === teamA) || TEAMS[0];
  const teamBInfo = TEAMS.find((t) => t.code === teamB) || TEAMS[1];

  // ─── Init session ───
  useEffect(() => {
    createMiniCupSession(mode, teamA, teamB)
      .then((s) => setSessionId(s.id))
      .catch(console.error);
  }, [mode, teamA, teamB]);

  // ─── Resize canvas ───
  useEffect(() => {
    const canvas = canvasRef.current;
    const container = containerRef.current;
    if (!canvas || !container) return;

    const resize = () => {
      const rect = container.getBoundingClientRect();
      const dpr = window.devicePixelRatio || 1;
      canvas.width = rect.width * dpr;
      canvas.height = rect.height * dpr;
      const ctx = canvas.getContext("2d");
      if (ctx) ctx.scale(dpr, dpr);
      canvas.style.width = `${rect.width}px`;
      canvas.style.height = `${rect.height}px`;
    };

    resize();
    window.addEventListener("resize", resize);
    return () => window.removeEventListener("resize", resize);
  }, []);

  // ─── Reset ball ───
  const resetBall = useCallback(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const rect = canvas.getBoundingClientRect();
    ballRef.current = {
      x: rect.width / 2,
      y: rect.height - 70,
      vx: 0, vy: 0,
      z: 0, vz: 0,
      active: false,
    };
    keeperRef.current = {
      x: 0, targetX: 0, delay: 0, diving: false, diveFrame: 0,
    };
    trajectoryPoints.current = [];
  }, []);

  // ─── Get canvas logical dimensions ───
  const getDimensions = useCallback(() => {
    const canvas = canvasRef.current;
    if (!canvas) return { w: 800, h: 600 };
    const rect = canvas.getBoundingClientRect();
    return { w: rect.width, h: rect.height };
  }, []);

  // ─── Draw scene ───
  const draw = useCallback(() => {
    const canvas = canvasRef.current;
    const ctx = canvas?.getContext("2d");
    if (!canvas || !ctx) return;

    const { w, h } = getDimensions();

    // Clear
    ctx.clearRect(0, 0, w, h);

    // ── Terrain ──
    ctx.fillStyle = "#22c55e";
    ctx.fillRect(0, 0, w, h);

    // Rayures d'herbe
    ctx.fillStyle = "#16a34a";
    const stripeW = 60;
    for (let i = 0; i < w; i += stripeW * 2) {
      ctx.fillRect(i, 0, stripeW, h);
    }

    // Ligne de surface
    ctx.strokeStyle = "rgba(255,255,255,0.6)";
    ctx.lineWidth = 3;
    const surfaceY = h * 0.35;
    ctx.beginPath();
    ctx.moveTo(w * 0.15, surfaceY);
    ctx.lineTo(w * 0.85, surfaceY);
    ctx.stroke();

    // Ligne de but
    ctx.lineWidth = 4;
    ctx.strokeStyle = "#ffffff";
    const goalLineY = h * 0.18;
    ctx.beginPath();
    ctx.moveTo(w * 0.2, goalLineY);
    ctx.lineTo(w * 0.8, goalLineY);
    ctx.stroke();

    // Surface de réparation
    ctx.strokeStyle = "rgba(255,255,255,0.4)";
    ctx.lineWidth = 2;
    ctx.strokeRect(w * 0.25, goalLineY, w * 0.5, surfaceY - goalLineY);

    // Point de penalty
    ctx.fillStyle = "#ffffff";
    ctx.beginPath();
    ctx.arc(w / 2, h * 0.55, 4, 0, Math.PI * 2);
    ctx.fill();

    // ── But / Cage ──
    const goalX = (w - GOAL_WIDTH) / 2;
    const goalY = goalLineY - GOAL_HEIGHT;

    // Poteaux
    ctx.fillStyle = "#e5e7eb";
    ctx.fillRect(goalX - 4, goalY, 4, GOAL_HEIGHT);
    ctx.fillRect(goalX + GOAL_WIDTH, goalY, 4, GOAL_HEIGHT);
    ctx.fillRect(goalX - 4, goalY - 4, GOAL_WIDTH + 8, 4);

    // Filet
    ctx.strokeStyle = "rgba(200,200,200,0.5)";
    ctx.lineWidth = 0.8;
    const netSpacing = 12;
    for (let i = 0; i <= GOAL_WIDTH; i += netSpacing) {
      ctx.beginPath();
      ctx.moveTo(goalX + i, goalY);
      ctx.lineTo(goalX + i, goalY + GOAL_HEIGHT);
      ctx.stroke();
    }
    for (let i = 0; i <= GOAL_HEIGHT; i += netSpacing) {
      ctx.beginPath();
      ctx.moveTo(goalX, goalY + i);
      ctx.lineTo(goalX + GOAL_WIDTH, goalY + i);
      ctx.stroke();
    }

    // ── Gardien ──
    const kp = keeperRef.current;
    const keeperScreenX = goalX + GOAL_WIDTH / 2 + kp.x;
    const keeperScreenY = goalY + GOAL_HEIGHT - KEEPER_HEIGHT / 2;

    // Ombre gardien
    ctx.fillStyle = "rgba(0,0,0,0.15)";
    ctx.beginPath();
    ctx.ellipse(keeperScreenX, keeperScreenY + KEEPER_HEIGHT / 2, KEEPER_WIDTH / 2, 6, 0, 0, Math.PI * 2);
    ctx.fill();

    // Corps gardien
    const isAITeam = mode === "solo_ai" && currentTeam === "b";
    const keeperColor = isAITeam ? teamBInfo.primary : teamAInfo.primary;
    const keeperSec = isAITeam ? teamBInfo.secondary : teamAInfo.secondary;

    ctx.fillStyle = keeperColor;
    // Position debout ou plongeon
    if (kp.diving) {
      const diveAngle = (kp.x > 0 ? 1 : -1) * Math.min(Math.abs(kp.x) * 0.02, 0.8);
      ctx.save();
      ctx.translate(keeperScreenX, keeperScreenY);
      ctx.rotate(diveAngle);
      ctx.fillRect(-KEEPER_WIDTH / 2, -KEEPER_HEIGHT / 2, KEEPER_WIDTH, KEEPER_HEIGHT);
      // Gants
      ctx.fillStyle = keeperSec;
      ctx.fillRect(-KEEPER_WIDTH / 2 - 4, -KEEPER_HEIGHT / 2 + 4, 6, 10);
      ctx.fillRect(KEEPER_WIDTH / 2 - 2, -KEEPER_HEIGHT / 2 + 4, 6, 10);
      ctx.restore();
    } else {
      ctx.fillRect(keeperScreenX - KEEPER_WIDTH / 2, keeperScreenY - KEEPER_HEIGHT / 2, KEEPER_WIDTH, KEEPER_HEIGHT);
      // Gants
      ctx.fillStyle = keeperSec;
      ctx.fillRect(keeperScreenX - KEEPER_WIDTH / 2 - 4, keeperScreenY - KEEPER_HEIGHT / 2 + 4, 6, 10);
      ctx.fillRect(keeperScreenX + KEEPER_WIDTH / 2 - 2, keeperScreenY - KEEPER_HEIGHT / 2 + 4, 6, 10);
    }

    // ── Ballon ──
    const bp = ballRef.current;
    const scale = 1 + bp.z * 0.008;
    const r = BALL_RADIUS * scale;

    // Ombre ballon
    ctx.fillStyle = "rgba(0,0,0,0.2)";
    ctx.beginPath();
    ctx.ellipse(bp.x, bp.y + 2, r * 0.8, r * 0.25, 0, 0, Math.PI * 2);
    ctx.fill();

    // Ballon
    ctx.beginPath();
    ctx.arc(bp.x, bp.y - bp.z, r, 0, Math.PI * 2);
    ctx.fillStyle = "#ffffff";
    ctx.fill();
    ctx.strokeStyle = "#1f2937";
    ctx.lineWidth = 1.5;
    ctx.stroke();

    // Motif ballon (pentagones)
    ctx.fillStyle = "#1f2937";
    ctx.beginPath();
    ctx.arc(bp.x, bp.y - bp.z, r * 0.4, 0, Math.PI * 2);
    ctx.fill();

    // ── Trajectoire preview ──
    if (phase === "aim" && trajectoryPoints.current.length > 1) {
      ctx.strokeStyle = "rgba(255,255,255,0.6)";
      ctx.lineWidth = 3;
      ctx.setLineDash([8, 6]);
      ctx.lineCap = "round";
      ctx.beginPath();
      ctx.moveTo(trajectoryPoints.current[0].x, trajectoryPoints.current[0].y);
      for (let i = 1; i < trajectoryPoints.current.length; i++) {
        ctx.lineTo(trajectoryPoints.current[i].x, trajectoryPoints.current[i].y);
      }
      ctx.stroke();
      ctx.setLineDash([]);

      // Flèche de direction
      const last = trajectoryPoints.current[trajectoryPoints.current.length - 1];
      const prev = trajectoryPoints.current[trajectoryPoints.current.length - 2] || trajectoryPoints.current[0];
      const angle = Math.atan2(last.y - prev.y, last.x - prev.x);
      ctx.fillStyle = "rgba(255,255,255,0.8)";
      ctx.beginPath();
      ctx.moveTo(last.x + Math.cos(angle) * 12, last.y + Math.sin(angle) * 12);
      ctx.lineTo(last.x + Math.cos(angle + 2.5) * 8, last.y + Math.sin(angle + 2.5) * 8);
      ctx.lineTo(last.x + Math.cos(angle - 2.5) * 8, last.y + Math.sin(angle - 2.5) * 8);
      ctx.fill();
    }

    // ── Vague d'impact (si but) ──
    if (shake && phase === "result") {
      ctx.strokeStyle = "rgba(255,255,255,0.3)";
      ctx.lineWidth = 2;
      for (let i = 1; i <= 3; i++) {
        ctx.beginPath();
        ctx.arc(goalX + GOAL_WIDTH / 2, goalY + GOAL_HEIGHT / 2, 20 + i * 15, 0, Math.PI * 2);
        ctx.stroke();
      }
    }
  }, [phase, mode, currentTeam, teamAInfo, teamBInfo, shake, getDimensions]);

  // ─── Animation loop ───
  useEffect(() => {
    const loop = () => {
      if (phase === "shot") {
        const bp = ballRef.current;
        const kp = keeperRef.current;
        const { w } = getDimensions();
        const goalX = (w - GOAL_WIDTH) / 2;
        const goalLineY = (canvasRef.current?.getBoundingClientRect().height || 600) * 0.18;
        const goalY = goalLineY - GOAL_HEIGHT;

        // Move ball
        bp.x += bp.vx;
        bp.y += bp.vy;
        bp.z += bp.vz;
        bp.vz -= GRAVITY;
        if (bp.z < 0) {
          bp.z = 0;
          bp.vz = -bp.vz * 0.4; // rebond
        }

        // Friction
        bp.vx *= 0.998;
        bp.vy *= 0.998;

        // Keeper AI reaction
        if (kp.delay > 0) {
          kp.delay--;
        } else {
          kp.diving = true;
          const dx = kp.targetX - kp.x;
          const speed = 3 + Math.min((scoreA + scoreB) * 0.3, 4); // difficulté progressive
          if (Math.abs(dx) > speed) {
            kp.x += Math.sign(dx) * speed;
          } else {
            kp.x = kp.targetX;
          }
        }

        // Collision keeper
        const keeperScreenX = goalX + GOAL_WIDTH / 2 + kp.x;
        const keeperScreenY = goalY + GOAL_HEIGHT - KEEPER_HEIGHT / 2;
        const distToKeeper = Math.sqrt(
          (bp.x - keeperScreenX) ** 2 + (bp.y - bp.z - keeperScreenY) ** 2
        );

        if (
          distToKeeper < KEEPER_WIDTH / 2 + BALL_RADIUS &&
          bp.z < KEEPER_HEIGHT + 10 &&
          bp.y < goalY + GOAL_HEIGHT + 20
        ) {
          endShot("saved");
          return;
        }

        // Goal detection
        if (
          bp.y < goalY + GOAL_HEIGHT + 5 &&
          bp.x > goalX + 5 &&
          bp.x < goalX + GOAL_WIDTH - 5 &&
          bp.z < GOAL_HEIGHT
        ) {
          endShot("goal");
          return;
        }

        // Miss (out of bounds)
        if (bp.y < 0 || bp.x < -20 || bp.x > w + 20 || bp.y > (canvasRef.current?.getBoundingClientRect().height || 600) + 20) {
          endShot("miss");
          return;
        }
      }

      draw();
      animFrameRef.current = requestAnimationFrame(loop);
    };

    animFrameRef.current = requestAnimationFrame(loop);
    return () => cancelAnimationFrame(animFrameRef.current);
  }, [phase, draw, scoreA, scoreB, getDimensions]);

  // ─── End shot logic ───
  const endShot = (result: "goal" | "saved" | "miss") => {
    setPhase("result");

    const msgs = {
      goal: "⚽ BUT !!",
      saved: "🧤 Arrêté !",
      miss: "❌ Raté !",
    };
    setMessage(msgs[result]);

    if (result === "goal") {
      if (currentTeam === "a") setScoreA((s) => s + 1);
      else setScoreB((s) => s + 1);
      setShake(true);
      setTimeout(() => setShake(false), 600);
    }

    setShots((prev) => [...prev, { team: currentTeam, result }]);
    shotOrderRef.current++;

    if (sessionId) {
      recordShot(sessionId, currentTeam, currentShooter, result, shotOrderRef.current).catch(console.error);
    }

    // Délai avant transition
    setTimeout(() => {
      checkEndGame(result);
    }, 1800);
  };

  // ─── Check end of game ───
  const checkEndGame = (lastResult: "goal" | "saved" | "miss") => {
    const totalShots = shots.length + 1;
    const isSolo = mode === "solo_ai";
    const currentScoreA = currentTeam === "a" && lastResult === "goal" ? scoreA + 1 : scoreA;
    const currentScoreB = currentTeam === "b" && lastResult === "goal" ? scoreB + 1 : scoreB;

    if (isSolo) {
      // 5 tirs puis mort subite
      if (totalShots >= 5 && currentScoreA !== currentScoreB) {
        finishGame(currentScoreA, currentScoreB);
        return;
      }
      // Mort subite après 5 tirs
      if (totalShots > 5 && currentScoreA !== currentScoreB) {
        finishGame(currentScoreA, currentScoreB);
        return;
      }
    } else {
      // 2v2 : 10 tirs total (5 chacun), puis mort subite
      if (totalShots >= 10 && currentScoreA !== currentScoreB) {
        finishGame(currentScoreA, currentScoreB);
        return;
      }
      if (totalShots > 10 && currentScoreA !== currentScoreB) {
        finishGame(currentScoreA, currentScoreB);
        return;
      }
    }

    nextTurn();
  };

  // ─── Next turn ───
  const nextTurn = () => {
    if (mode === "solo_ai") {
      setCurrentTeam("a");
      setCurrentShooter((s) => (s + 1) % 5);
    } else {
      setCurrentTeam((t) => (t === "a" ? "b" : "a"));
      if (currentTeam === "b") {
        setCurrentShooter((s) => (s + 1) % 5);
      }
    }
    setPhase("transition");
    setTimeout(() => {
      setPhase("aim");
      resetBall();
      setMessage("");
    }, 2200);
  };

  // ─── Finish game ───
  const finishGame = (finalA: number, finalB: number) => {
    setPhase("finished");
    const winner = finalA > finalB ? (mode === "solo_ai" ? "player" : "team_a") : "team_b";
    const winnerTeam = finalA > finalB ? teamAInfo.name : teamBInfo.name;
    const winnerFlag = finalA > finalB ? teamAInfo.flag : teamBInfo.flag;
    setMessage(`${winnerFlag} ${winnerTeam} remporte la séance !`);

    const xp = calculateXP(finalA, finalB, winner);
    setXPEarned(xp);

    if (sessionId) {
      finishSession(sessionId, finalA, finalB, winner)
        .then(() => console.log("Session finished"))
        .catch(console.error);
    }
  };

  const calculateXP = (a: number, b: number, winner: string) => {
    let xp = (a + b) * 10;
    if (winner === "player" || winner === "team_a") xp += 50;
    if (a === 5 && b === 0) xp += 100;
    return xp;
  };

  // ─── Pointer events (swipe/draw) ───
  const handlePointerDown = (e: React.PointerEvent) => {
    if (phase !== "aim") return;
    e.preventDefault();
    const canvas = canvasRef.current;
    if (!canvas) return;

    const rect = canvas.getBoundingClientRect();
    const startX = e.clientX - rect.left;
    const startY = e.clientY - rect.top;

    pointerActive.current = true;
    trajectoryPoints.current = [{ x: startX, y: startY }];

    const handleMove = (ev: PointerEvent) => {
      if (!pointerActive.current) return;
      const x = ev.clientX - rect.left;
      const y = ev.clientY - rect.top;
      trajectoryPoints.current.push({ x, y });
      draw();
    };

    const handleUp = (ev: PointerEvent) => {
      pointerActive.current = false;
      window.removeEventListener("pointermove", handleMove);
      window.removeEventListener("pointerup", handleUp);
      window.removeEventListener("pointercancel", handleUp);

      const endX = ev.clientX - rect.left;
      const endY = ev.clientY - rect.top;
      const dx = endX - startX;
      const dy = endY - startY;
      const dist = Math.sqrt(dx * dx + dy * dy);

      if (dist < 15) {
        trajectoryPoints.current = [];
        draw();
        return; // trop court
      }

      // Launch
      const power = Math.min(dist / 150, 2.0);
      const angle = Math.atan2(dy, dx);
      const bp = ballRef.current;
      bp.vx = Math.cos(angle) * BALL_SPEED_BASE * power;
      bp.vy = Math.sin(angle) * BALL_SPEED_BASE * power;
      bp.vz = power * 6 + Math.random() * 2;
      bp.active = true;

      // Keeper AI
      const { w } = getDimensions();
      const goalX = (w - GOAL_WIDTH) / 2;
      const difficulty = Math.min(0.25 + (scoreA + scoreB) * 0.07, 0.92);
      const targetInGoal = endX - (goalX + GOAL_WIDTH / 2);
      const error = (Math.random() - 0.5) * (1 - difficulty) * GOAL_WIDTH * 1.2;
      const reactionDelay = Math.random() > difficulty ? 25 : Math.floor(5 + Math.random() * 8);

      keeperRef.current.targetX = Math.max(
        -GOAL_WIDTH / 2 + KEEPER_WIDTH / 2,
        Math.min(GOAL_WIDTH / 2 - KEEPER_WIDTH / 2, targetInGoal + error)
      );
      keeperRef.current.delay = reactionDelay;

      setPhase("shot");
    };

    window.addEventListener("pointermove", handleMove);
    window.addEventListener("pointerup", handleUp);
    window.addEventListener("pointercancel", handleUp);
  };

  // ─── Keyboard shot handler ───
  const handleKeyboardShot = (dirX: number, dirY: number, power: number) => {
    if (phase !== "aim") return;
    const bp = ballRef.current;
    const angle = Math.atan2(dirY, dirX);
    bp.vx = Math.cos(angle) * BALL_SPEED_BASE * power;
    bp.vy = Math.sin(angle) * BALL_SPEED_BASE * power;
    bp.vz = power * 6;
    bp.active = true;

    const { w } = getDimensions();
    const goalX = (w - GOAL_WIDTH) / 2;
    const difficulty = Math.min(0.25 + (scoreA + scoreB) * 0.07, 0.92);
    const targetInGoal = dirX * 80;
    const error = (Math.random() - 0.5) * (1 - difficulty) * GOAL_WIDTH;
    const reactionDelay = Math.random() > difficulty ? 25 : Math.floor(5 + Math.random() * 8);

    keeperRef.current.targetX = Math.max(
      -GOAL_WIDTH / 2 + KEEPER_WIDTH / 2,
      Math.min(GOAL_WIDTH / 2 - KEEPER_WIDTH / 2, targetInGoal + error)
    );
    keeperRef.current.delay = reactionDelay;

    setPhase("shot");
  };

  // ─── Initial reset ───
  useEffect(() => {
    resetBall();
  }, [resetBall]);

  const isPlayerTurn = mode === "solo_ai" ? currentTeam === "a" : true;
  const currentTeamInfo = currentTeam === "a" ? teamAInfo : teamBInfo;
  const totalShotsDone = shots.length;
  const maxRegularShots = mode === "solo_ai" ? 5 : 10;
  const isSuddenDeath = totalShotsDone >= maxRegularShots;

  return (
    <div className="space-y-3">
      {/* Scoreboard */}
      <div className="flex items-center justify-between bg-muted rounded-xl p-3 border">
        <div className="flex items-center gap-2">
          <span className="text-xl">{teamAInfo.flag}</span>
          <span className="font-bold text-sm hidden sm:inline">{teamAInfo.name}</span>
        </div>
        <div className="flex items-center gap-3">
          <span className="text-2xl font-black tabular-nums">{scoreA}</span>
          <span className="text-muted-foreground">-</span>
          <span className="text-2xl font-black tabular-nums">{scoreB}</span>
        </div>
        <div className="flex items-center gap-2">
          <span className="font-bold text-sm hidden sm:inline">{teamBInfo.name}</span>
          <span className="text-xl">{teamBInfo.flag}</span>
        </div>
      </div>

      {/* Info tour */}
      <div className="flex items-center justify-between text-sm">
        <div className="flex items-center gap-2">
          <span
            className="w-3 h-3 rounded-full"
            style={{ backgroundColor: currentTeamInfo.primary }}
          />
          <span className="font-medium">
            {isPlayerTurn ? "À toi de tirer" : `Tour de l'adversaire`}
          </span>
          {isSuddenDeath && (
            <span className="text-red-500 font-bold text-xs bg-red-50 px-2 py-0.5 rounded-full">
              MORT SUBITE
            </span>
          )}
        </div>
        <span className="text-muted-foreground text-xs">
          {mode === "solo_ai"
            ? `Tir ${Math.min(currentShooter + 1, 5)}/5`
            : `Tour ${Math.floor(totalShotsDone / 2) + 1}`}
        </span>
      </div>

      <ShotHistory shots={shots} />

      {/* Canvas Game Area */}
      <div
        ref={containerRef}
        className={cn(
          "relative aspect-[4/3] rounded-xl overflow-hidden border-2 border-green-700 shadow-lg select-none",
          shake && "animate-pulse"
        )}
      >
        <canvas
          ref={canvasRef}
          className="w-full h-full cursor-crosshair touch-none block"
          onPointerDown={handlePointerDown}
        />

        {/* Transition overlay */}
        {phase === "transition" && (
          <TurnTransitionScreen
            teamName={currentTeamInfo.name}
            teamFlag={currentTeamInfo.flag}
            teamColor={currentTeamInfo.primary}
            role={isPlayerTurn ? "shoot" : "defend"}
            mode={mode}
          />
        )}

        {/* Result overlay */}
        {phase === "result" && (
          <div className="absolute inset-0 flex items-center justify-center bg-black/40 backdrop-blur-sm">
            <div
              className={cn(
                "text-4xl sm:text-5xl font-black text-white drop-shadow-lg animate-in zoom-in duration-300",
                message.includes("BUT") && "text-green-300",
                message.includes("Arrêté") && "text-yellow-300",
                message.includes("Raté") && "text-red-300"
              )}
            >
              {message}
            </div>
          </div>
        )}

        {/* Finished overlay */}
        {phase === "finished" && (
          <GameOverScreen
            message={message}
            scoreA={scoreA}
            scoreB={scoreB}
            teamAInfo={teamAInfo}
            teamBInfo={teamBInfo}
            xpEarned={xpEarned}
            onReplay={() => {
              setPhase("aim");
              setShots([]);
              setScoreA(0);
              setScoreB(0);
              setCurrentShooter(0);
              setCurrentTeam("a");
              setMessage("");
              setXPEarned(0);
              shotOrderRef.current = 0;
              resetBall();
              createMiniCupSession(mode, teamA, teamB)
                .then((s) => setSessionId(s.id))
                .catch(console.error);
            }}
            onQuit={onQuit}
          />
        )}
      </div>

      {/* Controls */}
      <div className="flex items-center justify-between gap-2">
        <Button
          variant="outline"
          size="sm"
          className="gap-2"
          onClick={() => setShowKeyboard((v) => !v)}
        >
          {showKeyboard ? <MousePointer className="w-4 h-4" /> : <Keyboard className="w-4 h-4" />}
          {showKeyboard ? "Mode souris" : "Mode clavier"}
        </Button>

        <div className="text-xs text-muted-foreground hidden sm:block">
          {phase === "aim" && !showKeyboard && "Dessine un trait vers le but pour tirer"}
        </div>

        <Button variant="ghost" size="sm" className="text-red-500" onClick={onQuit}>
          Abandonner
        </Button>
      </div>

      {showKeyboard && phase === "aim" && (
        <KeyboardAim onShoot={handleKeyboardShot} />
      )}
    </div>
  );
}
