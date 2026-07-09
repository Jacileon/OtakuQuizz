-- Fix: quiz leaderboard based on XP (not score), one entry per user

-- Create quiz_leaderboard table for per-quiz XP-based rankings
CREATE TABLE IF NOT EXISTS quiz_leaderboard (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    quiz_id uuid NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    xp_earned integer NOT NULL DEFAULT 0,
    accuracy_rate numeric DEFAULT 0,
    time_taken_ms integer DEFAULT 0,
    attempt_number integer NOT NULL DEFAULT 1,
    updated_at timestamptz DEFAULT now(),
    UNIQUE(quiz_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_quiz_leaderboard_quiz ON quiz_leaderboard(quiz_id, xp_earned DESC);

ALTER TABLE quiz_leaderboard ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS "quiz_leaderboard_select" ON quiz_leaderboard;
CREATE POLICY "quiz_leaderboard_select" ON quiz_leaderboard FOR SELECT USING (true);
DROP POLICY IF EXISTS "quiz_leaderboard_insert" ON quiz_leaderboard;
CREATE POLICY "quiz_leaderboard_insert" ON quiz_leaderboard FOR INSERT WITH CHECK (true);
DROP POLICY IF EXISTS "quiz_leaderboard_update" ON quiz_leaderboard;
CREATE POLICY "quiz_leaderboard_update" ON quiz_leaderboard FOR UPDATE USING (true);

-- Drop old version first (signature changed)
DROP FUNCTION IF EXISTS get_quiz_leaderboard(uuid);

-- Updated get_quiz_leaderboard function using XP
CREATE OR REPLACE FUNCTION get_quiz_leaderboard(quiz_id uuid)
RETURNS TABLE (
    rank bigint,
    user_id uuid,
    username text,
    avatar_url text,
    user_rank text,
    xp_earned integer,
    accuracy_rate numeric,
    time_taken_ms integer
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        ROW_NUMBER() OVER (ORDER BY ql.xp_earned DESC, ql.time_taken_ms ASC)::bigint as rank,
        ql.user_id,
        p.username,
        p.avatar_url,
        p.rank as user_rank,
        ql.xp_earned,
        ql.accuracy_rate,
        ql.time_taken_ms
    FROM quiz_leaderboard ql
    JOIN user_profiles p ON ql.user_id = p.id
    WHERE ql.quiz_id = get_quiz_leaderboard.quiz_id
    ORDER BY ql.xp_earned DESC, ql.time_taken_ms ASC
    LIMIT 100;
END;
$$ LANGUAGE plpgsql;