-- ============================================================
-- MIGRATION COMPLÈTE: XP, Streak, Classement, Téléphone
-- ============================================================

-- Ajouter colonnes à user_profiles
ALTER TABLE user_profiles ADD COLUMN IF NOT EXISTS current_streak integer DEFAULT 0;
ALTER TABLE user_profiles ADD COLUMN IF NOT EXISTS longest_streak integer DEFAULT 0;
ALTER TABLE user_profiles ADD COLUMN IF NOT EXISTS last_login_date date;
ALTER TABLE user_profiles ADD COLUMN IF NOT EXISTS total_xp integer DEFAULT 0;

-- Table UserQuizAttempt
CREATE TABLE IF NOT EXISTS user_quiz_attempts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    quiz_id uuid NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    attempt_number integer NOT NULL,
    score integer DEFAULT 0,
    xp_earned integer DEFAULT 0,
    completed_at timestamptz DEFAULT now(),
    UNIQUE(user_id, quiz_id, attempt_number)
);

CREATE INDEX IF NOT EXISTS idx_user_quiz_attempts_user ON user_quiz_attempts(user_id, quiz_id);

-- Table UserQuestionAttempt
CREATE TABLE IF NOT EXISTS user_question_attempts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    quiz_id uuid NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    question_id uuid NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    attempt_number integer NOT NULL,
    is_correct boolean NOT NULL,
    xp_earned integer DEFAULT 0,
    created_at timestamptz DEFAULT now(),
    UNIQUE(user_id, quiz_id, question_id, attempt_number)
);

CREATE INDEX IF NOT EXISTS idx_user_question_attempts_user ON user_question_attempts(user_id, quiz_id);

-- Table XpTransaction
CREATE TABLE IF NOT EXISTS xp_transactions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    source text NOT NULL CHECK (source IN ('quiz', 'streak', 'challenge', 'event')),
    source_id uuid,
    amount numeric(10,2) NOT NULL,
    created_at timestamptz DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_xp_transactions_user ON xp_transactions(user_id, created_at DESC);

-- Table RankConfig
CREATE TABLE IF NOT EXISTS rank_config (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    rank_label text UNIQUE NOT NULL,
    xp_required integer NOT NULL,
    display_order integer NOT NULL
);

-- Insérer les rangs par défaut
INSERT INTO rank_config (rank_label, xp_required, display_order) VALUES
    ('F', 0, 1),
    ('E', 100, 2),
    ('D', 500, 3),
    ('C', 1500, 4),
    ('B', 3000, 5),
    ('A', 6000, 6),
    ('S', 10000, 7),
    ('S+', 15000, 8),
    ('SS', 25000, 9),
    ('SSS', 40000, 10),
    ('Légende', 60000, 11)
ON CONFLICT (rank_label) DO NOTHING;

-- RLS
ALTER TABLE user_quiz_attempts ENABLE ROW LEVEL SECURITY;
ALTER TABLE user_question_attempts ENABLE ROW LEVEL SECURITY;
ALTER TABLE xp_transactions ENABLE ROW LEVEL SECURITY;
ALTER TABLE rank_config ENABLE ROW LEVEL SECURITY;

-- user_quiz_attempts
DROP POLICY IF EXISTS "user_quiz_attempts_select_own" ON user_quiz_attempts;
CREATE POLICY "user_quiz_attempts_select_own" ON user_quiz_attempts FOR SELECT USING (auth.uid() = user_id);
DROP POLICY IF EXISTS "user_quiz_attempts_insert_own" ON user_quiz_attempts;
CREATE POLICY "user_quiz_attempts_insert_own" ON user_quiz_attempts FOR INSERT WITH CHECK (auth.uid() = user_id);

-- user_question_attempts
DROP POLICY IF EXISTS "user_question_attempts_select_own" ON user_question_attempts;
CREATE POLICY "user_question_attempts_select_own" ON user_question_attempts FOR SELECT USING (auth.uid() = user_id);
DROP POLICY IF EXISTS "user_question_attempts_insert_own" ON user_question_attempts;
CREATE POLICY "user_question_attempts_insert_own" ON user_question_attempts FOR INSERT WITH CHECK (auth.uid() = user_id);

-- xp_transactions
DROP POLICY IF EXISTS "xp_transactions_select_own" ON xp_transactions;
CREATE POLICY "xp_transactions_select_own" ON xp_transactions FOR SELECT USING (auth.uid() = user_id);

-- rank_config
DROP POLICY IF EXISTS "rank_config_select_public" ON rank_config;
CREATE POLICY "rank_config_select_public" ON rank_config FOR SELECT USING (true);