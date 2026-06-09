-- Politiques RLS manquantes pour leaderboard_monthly et user_quiz_attempts

-- leaderboard_monthly: permettre l'upsert
DROP POLICY IF EXISTS "leaderboard_monthly_insert_own" ON leaderboard_monthly;
CREATE POLICY "leaderboard_monthly_insert_own" ON leaderboard_monthly FOR INSERT WITH CHECK (auth.uid() = user_id);

DROP POLICY IF EXISTS "leaderboard_monthly_update_own" ON leaderboard_monthly;
CREATE POLICY "leaderboard_monthly_update_own" ON leaderboard_monthly FOR UPDATE USING (auth.uid() = user_id);

-- user_quiz_attempts: permettre l'insertion
DROP POLICY IF EXISTS "user_quiz_attempts_insert_own" ON user_quiz_attempts;
CREATE POLICY "user_quiz_attempts_insert_own" ON user_quiz_attempts FOR INSERT WITH CHECK (auth.uid() = user_id);

-- user_question_attempts: permettre l'insertion
DROP POLICY IF EXISTS "user_question_attempts_insert_own" ON user_question_attempts;
CREATE POLICY "user_question_attempts_insert_own" ON user_question_attempts FOR INSERT WITH CHECK (auth.uid() = user_id);

-- xp_transactions: permettre l'insertion
DROP POLICY IF EXISTS "xp_transactions_insert_own" ON xp_transactions;
CREATE POLICY "xp_transactions_insert_own" ON xp_transactions FOR INSERT WITH CHECK (auth.uid() = user_id);

-- Fonction RPC pour l'historique des parties
CREATE OR REPLACE FUNCTION get_user_game_history(p_user_id uuid)
RETURNS TABLE (
    id uuid,
    quiz_id uuid,
    score integer,
    correct_count integer,
    total_questions integer,
    accuracy_rate numeric,
    is_perfect boolean,
    time_taken_ms integer,
    started_at timestamptz,
    completed_at timestamptz,
    quiz_title text,
    quiz_thumbnail text,
    quiz_category text,
    quiz_series text
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        gs.id,
        gs.quiz_id,
        gs.score,
        gs.correct_count,
        gs.total_questions,
        gs.accuracy_rate,
        gs.is_perfect,
        gs.time_taken_ms,
        gs.started_at,
        gs.completed_at,
        q.title,
        q.thumbnail_url,
        q.category,
        q.series
    FROM game_sessions gs
    JOIN quizzes q ON q.id = gs.quiz_id
    WHERE gs.user_id = p_user_id
    AND gs.completed_at IS NOT NULL
    ORDER BY gs.completed_at DESC
    LIMIT 50;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;