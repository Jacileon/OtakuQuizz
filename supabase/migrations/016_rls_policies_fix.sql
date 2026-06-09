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