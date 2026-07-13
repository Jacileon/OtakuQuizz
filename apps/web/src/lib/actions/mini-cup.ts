"use server";

import { createClient } from "@/lib/supabase/server";
import { revalidatePath } from "next/cache";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export async function createMiniCupSession(
  mode: "solo_ai" | "local_2v2",
  teamA: string,
  teamB: string
) {
  const supabase = await createClient();
  const {
    data: { user },
  } = await supabase.auth.getUser();
  if (!user) throw new Error("Unauthorized");

  const res = await fetch(`${API_URL}/api/mini-cup/session`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ mode, team_a: teamA, team_b: teamB }),
  });
  if (!res.ok) {
    const err = await res.text();
    throw new Error(`Failed to create session: ${err}`);
  }
  return res.json();
}

export async function recordShot(
  sessionId: string,
  team: "a" | "b",
  shooterIndex: number,
  result: "goal" | "saved" | "miss",
  shotOrder: number
) {
  const res = await fetch(
    `${API_URL}/api/mini-cup/session/${sessionId}/shot`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        team,
        shooter_index: shooterIndex,
        result,
        shot_order: shotOrder,
      }),
    }
  );
  if (!res.ok) {
    const err = await res.text();
    throw new Error(`Failed to record shot: ${err}`);
  }
  return res.json();
}

export async function finishSession(
  sessionId: string,
  scoreA: number,
  scoreB: number,
  winner: string
) {
  const res = await fetch(
    `${API_URL}/api/mini-cup/session/${sessionId}/finish`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ score_a: scoreA, score_b: scoreB, winner }),
    }
  );
  if (!res.ok) {
    const err = await res.text();
    throw new Error(`Failed to finish session: ${err}`);
  }
  revalidatePath("/leaderboard");
  revalidatePath("/games/mini-cup");
  return res.json();
}

export async function getMiniCupLeaderboard(limit = 50) {
  const res = await fetch(
    `${API_URL}/api/mini-cup/leaderboard?limit=${limit}`,
    { cache: "no-store" }
  );
  if (!res.ok) {
    const err = await res.text();
    throw new Error(`Failed to fetch leaderboard: ${err}`);
  }
  return res.json();
}

export async function getMiniCupSessions() {
  const supabase = await createClient();
  const {
    data: { user },
  } = await supabase.auth.getUser();
  if (!user) throw new Error("Unauthorized");

  const res = await fetch(`${API_URL}/api/mini-cup/sessions`, {
    headers: {
      Authorization: `Bearer ${(await supabase.auth.getSession()).data.session?.access_token}`,
    },
  });
  if (!res.ok) {
    const err = await res.text();
    throw new Error(`Failed to fetch sessions: ${err}`);
  }
  return res.json();
}
