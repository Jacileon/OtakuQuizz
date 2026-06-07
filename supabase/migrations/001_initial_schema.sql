-- ============================================================
-- SCHÉMA COMPLET - OTAKU QUIZ AFRICA
-- ============================================================

-- Fonction pour updated_at automatique
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- ============================================================
-- 1. USER_PROFILES
-- ============================================================
CREATE TABLE user_profiles (
    id uuid PRIMARY KEY REFERENCES auth.users(id) ON DELETE CASCADE,
    email text UNIQUE NOT NULL,
    username text UNIQUE NOT NULL,
    avatar_url text,
    bio text CHECK (length(bio) <= 200),
    country text,
    favorite_anime text,
    xp integer DEFAULT 0 NOT NULL,
    level integer DEFAULT 1 NOT NULL,
    rank text DEFAULT 'F' NOT NULL,
    is_premium boolean DEFAULT false,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now(),
    CONSTRAINT username_length CHECK (length(username) >= 3 AND length(username) <= 30),
    CONSTRAINT username_format CHECK (username ~ '^[a-zA-Z0-9_]+$')
);

CREATE TRIGGER update_user_profiles_updated_at
    BEFORE UPDATE ON user_profiles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================
-- 2. USER_STATS
-- ============================================================
CREATE TABLE user_stats (
    user_id uuid PRIMARY KEY REFERENCES user_profiles(id) ON DELETE CASCADE,
    quizzes_played integer DEFAULT 0,
    quizzes_created integer DEFAULT 0,
    total_correct_answers integer DEFAULT 0,
    total_answers integer DEFAULT 0,
    accuracy_rate numeric(5,2) DEFAULT 0,
    best_score_ever integer DEFAULT 0,
    monthly_rank integer,
    global_rank integer,
    updated_at timestamptz DEFAULT now()
);

CREATE TRIGGER update_user_stats_updated_at
    BEFORE UPDATE ON user_stats
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================
-- 3. QUIZZES
-- ============================================================
CREATE TABLE quizzes (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    creator_id uuid NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    title text NOT NULL CHECK (length(title) >= 5 AND length(title) <= 100),
    description text CHECK (length(description) <= 500),
    thumbnail_url text,
    thumbnail_public_id text,
    category text NOT NULL,
    subcategory text NOT NULL,
    series text NOT NULL,
    quiz_type text DEFAULT 'community' CHECK (quiz_type IN ('community', 'official', 'private')),
    status text DEFAULT 'published' CHECK (status IN ('draft', 'published', 'hidden', 'deleted')),
    question_count integer DEFAULT 0,
    play_count integer DEFAULT 0,
    total_reports integer DEFAULT 0,
    is_visible boolean DEFAULT true,
    event_start_at timestamptz,
    event_end_at timestamptz,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE INDEX idx_quizzes_category ON quizzes(category);
CREATE INDEX idx_quizzes_subcategory ON quizzes(subcategory);
CREATE INDEX idx_quizzes_series ON quizzes(series);
CREATE INDEX idx_quizzes_creator ON quizzes(creator_id);
CREATE INDEX idx_quizzes_status ON quizzes(status);
CREATE INDEX idx_quizzes_type ON quizzes(quiz_type);
CREATE INDEX idx_quizzes_visible ON quizzes(is_visible) WHERE is_visible = true;

CREATE TRIGGER update_quizzes_updated_at
    BEFORE UPDATE ON quizzes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================
-- 4. QUESTIONS
-- ============================================================
CREATE TABLE questions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    quiz_id uuid NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    question_text text NOT NULL,
    question_type text NOT NULL CHECK (question_type IN ('text', 'true_false', 'image', 'image_shadow', 'gif', 'audio')),
    media_url text,
    media_public_id text,
    time_limit_seconds integer DEFAULT 30 CHECK (time_limit_seconds >= 3 AND time_limit_seconds <= 60),
    order_index integer NOT NULL,
    created_at timestamptz DEFAULT now()
);

CREATE INDEX idx_questions_quiz ON questions(quiz_id);

-- ============================================================
-- 5. ANSWERS
-- ============================================================
CREATE TABLE answers (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    question_id uuid NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    answer_text text NOT NULL,
    is_correct boolean DEFAULT false NOT NULL,
    order_index integer NOT NULL,
    created_at timestamptz DEFAULT now()
);

CREATE INDEX idx_answers_question ON answers(question_id);

-- ============================================================
-- 6. GAME_SESSIONS
-- ============================================================
CREATE TABLE game_sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    quiz_id uuid NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    started_at timestamptz DEFAULT now(),
    completed_at timestamptz,
    score integer DEFAULT 0,
    correct_count integer DEFAULT 0,
    total_questions integer NOT NULL,
    accuracy_rate numeric(5,2) DEFAULT 0,
    is_perfect boolean DEFAULT false,
    time_taken_ms integer,
    created_at timestamptz DEFAULT now()
);

CREATE INDEX idx_sessions_user ON game_sessions(user_id);
CREATE INDEX idx_sessions_quiz ON game_sessions(quiz_id);
CREATE INDEX idx_sessions_completed ON game_sessions(completed_at) WHERE completed_at IS NOT NULL;

-- ============================================================
-- 7. PLAYER_ANSWERS
-- ============================================================
CREATE TABLE player_answers (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id uuid NOT NULL REFERENCES game_sessions(id) ON DELETE CASCADE,
    question_id uuid NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    answer_id uuid REFERENCES answers(id) ON DELETE SET NULL,
    is_correct boolean NOT NULL,
    time_taken_ms integer NOT NULL,
    points_earned integer DEFAULT 0,
    created_at timestamptz DEFAULT now()
);

CREATE INDEX idx_player_answers_session ON player_answers(session_id);
CREATE INDEX idx_player_answers_question ON player_answers(question_id);

-- ============================================================
-- 8. FRIENDSHIPS
-- ============================================================
CREATE TABLE friendships (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    requester_id uuid NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    addressee_id uuid NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    status text DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'rejected')),
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now(),
    UNIQUE(requester_id, addressee_id),
    CHECK (requester_id != addressee_id)
);

CREATE INDEX idx_friendships_requester ON friendships(requester_id);
CREATE INDEX idx_friendships_addressee ON friendships(addressee_id);
CREATE INDEX idx_friendships_status ON friendships(status);

CREATE TRIGGER update_friendships_updated_at
    BEFORE UPDATE ON friendships
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================
-- 9. BADGES
-- ============================================================
CREATE TABLE badges (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    slug text UNIQUE NOT NULL,
    name text NOT NULL,
    description text NOT NULL,
    icon_url text,
    condition_type text NOT NULL,
    condition_value integer NOT NULL,
    is_rare boolean DEFAULT false,
    created_at timestamptz DEFAULT now()
);

CREATE INDEX idx_badges_slug ON badges(slug);

-- ============================================================
-- 10. USER_BADGES
-- ============================================================
CREATE TABLE user_badges (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    badge_id uuid NOT NULL REFERENCES badges(id) ON DELETE CASCADE,
    earned_at timestamptz DEFAULT now(),
    UNIQUE(user_id, badge_id)
);

CREATE INDEX idx_user_badges_user ON user_badges(user_id);

-- ============================================================
-- 11. REPORTS
-- ============================================================
CREATE TABLE reports (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    reporter_id uuid NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    quiz_id uuid NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    reason text NOT NULL CHECK (reason IN ('wrong_answer', 'incorrect_content', 'spam', 'plagiarism', 'inappropriate')),
    description text,
    status text DEFAULT 'pending' CHECK (status IN ('pending', 'reviewed', 'resolved', 'dismissed')),
    created_at timestamptz DEFAULT now(),
    UNIQUE(reporter_id, quiz_id)
);

CREATE INDEX idx_reports_quiz ON reports(quiz_id);
CREATE INDEX idx_reports_status ON reports(status);

-- ============================================================
-- 12. LEADERBOARD_MONTHLY
-- ============================================================
CREATE TABLE leaderboard_monthly (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    month_year text NOT NULL,
    score integer NOT NULL DEFAULT 0,
    rank_position integer,
    created_at timestamptz DEFAULT now(),
    UNIQUE(user_id, month_year)
);

CREATE INDEX idx_leaderboard_month ON leaderboard_monthly(month_year);
CREATE INDEX idx_leaderboard_month_score ON leaderboard_monthly(month_year, score DESC);

-- ============================================================
-- 13. NOTIFICATIONS
-- ============================================================
CREATE TABLE notifications (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    type text NOT NULL CHECK (type IN ('friend_request', 'badge_unlocked', 'quiz_completed', 'event_starting', 'rank_up')),
    title text NOT NULL,
    message text NOT NULL,
    data jsonb,
    is_read boolean DEFAULT false,
    created_at timestamptz DEFAULT now()
);

CREATE INDEX idx_notifications_user ON notifications(user_id);
CREATE INDEX idx_notifications_unread ON notifications(user_id, is_read) WHERE is_read = false;

-- ============================================================
-- TRIGGERS MÉTIER
-- ============================================================

-- Trigger: Incrémenter play_count
CREATE OR REPLACE FUNCTION increment_play_count()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.completed_at IS NOT NULL AND OLD.completed_at IS NULL THEN
        UPDATE quizzes SET play_count = play_count + 1 WHERE id = NEW.quiz_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_increment_play_count
    AFTER UPDATE ON game_sessions
    FOR EACH ROW EXECUTE FUNCTION increment_play_count();

-- Trigger: Mettre à jour user_stats
CREATE OR REPLACE FUNCTION update_user_stats_after_session()
RETURNS TRIGGER AS $$
DECLARE
    total_correct integer;
    total_answers integer;
    accuracy numeric(5,2);
BEGIN
    IF NEW.completed_at IS NOT NULL THEN
        SELECT 
            COALESCE(SUM(correct_count), 0),
            COALESCE(SUM(total_questions), 0)
        INTO total_correct, total_answers
        FROM game_sessions
        WHERE user_id = NEW.user_id AND completed_at IS NOT NULL;

        accuracy := CASE 
            WHEN total_answers > 0 THEN ROUND((total_correct::numeric / total_answers) * 100, 2)
            ELSE 0 
        END;

        UPDATE user_stats SET
            quizzes_played = (SELECT COUNT(*) FROM game_sessions WHERE user_id = NEW.user_id AND completed_at IS NOT NULL),
            total_correct_answers = total_correct,
            total_answers = total_answers,
            accuracy_rate = accuracy,
            best_score_ever = GREATEST(best_score_ever, NEW.score),
            updated_at = now()
        WHERE user_id = NEW.user_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_user_stats
    AFTER UPDATE ON game_sessions
    FOR EACH ROW EXECUTE FUNCTION update_user_stats_after_session();

-- Trigger: Mettre à jour XP et rang
CREATE OR REPLACE FUNCTION update_user_xp_and_rank()
RETURNS TRIGGER AS $$
DECLARE
    new_xp integer;
    new_rank text;
BEGIN
    IF NEW.completed_at IS NOT NULL AND OLD.completed_at IS NULL THEN
        new_xp := NEW.score;

        UPDATE user_profiles 
        SET xp = xp + new_xp,
            level = GREATEST(1, FLOOR(SQRT(xp + new_xp) / 10) + 1)
        WHERE id = NEW.user_id;

        -- Recalculer le rang
        SELECT CASE
            WHEN xp >= 60000 THEN 'Légende'
            WHEN xp >= 40000 THEN 'SSS'
            WHEN xp >= 25000 THEN 'SS'
            WHEN xp >= 15000 THEN 'S+'
            WHEN xp >= 10000 THEN 'S'
            WHEN xp >= 6000 THEN 'A'
            WHEN xp >= 3000 THEN 'B'
            WHEN xp >= 1500 THEN 'C'
            WHEN xp >= 500 THEN 'D'
            WHEN xp >= 100 THEN 'E'
            ELSE 'F'
        END INTO new_rank
        FROM user_profiles WHERE id = NEW.user_id;

        UPDATE user_profiles SET rank = new_rank WHERE id = NEW.user_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_user_xp
    AFTER UPDATE ON game_sessions
    FOR EACH ROW EXECUTE FUNCTION update_user_xp_and_rank();

-- Trigger: Auto-hide quiz si trop de reports
CREATE OR REPLACE FUNCTION auto_hide_reported_quiz()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE quizzes 
    SET is_visible = false, status = 'hidden'
    WHERE id = NEW.quiz_id AND (
        SELECT COUNT(*) FROM reports WHERE quiz_id = NEW.quiz_id
    ) >= 5;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_auto_hide_quiz
    AFTER INSERT ON reports
    FOR EACH ROW EXECUTE FUNCTION auto_hide_reported_quiz();

-- ============================================================
-- FONCTIONS
-- ============================================================

-- Classement global par XP
CREATE OR REPLACE FUNCTION get_global_leaderboard(limit_count int DEFAULT 100)
RETURNS TABLE (
    rank bigint,
    user_id uuid,
    username text,
    avatar_url text,
    user_rank text,
    xp integer,
    quizzes_played bigint
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        ROW_NUMBER() OVER (ORDER BY p.xp DESC)::bigint as rank,
        p.id as user_id,
        p.username,
        p.avatar_url,
        p.rank as user_rank,
        p.xp,
        COALESCE(s.quizzes_played, 0)::bigint as quizzes_played
    FROM user_profiles p
    LEFT JOIN user_stats s ON p.id = s.user_id
    ORDER BY p.xp DESC
    LIMIT limit_count;
END;
$$ LANGUAGE plpgsql;

-- Classement mensuel
CREATE OR REPLACE FUNCTION get_monthly_leaderboard(year_month text, limit_count int DEFAULT 100)
RETURNS TABLE (
    rank bigint,
    user_id uuid,
    username text,
    avatar_url text,
    user_rank text,
    score integer,
    quiz_count bigint
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        ROW_NUMBER() OVER (ORDER BY l.score DESC)::bigint as rank,
        l.user_id,
        p.username,
        p.avatar_url,
        p.rank as user_rank,
        l.score,
        COUNT(DISTINCT gs.id)::bigint as quiz_count
    FROM leaderboard_monthly l
    JOIN user_profiles p ON l.user_id = p.id
    LEFT JOIN game_sessions gs ON gs.user_id = l.user_id 
        AND gs.completed_at >= (year_month || '-01')::timestamptz
        AND gs.completed_at < ((year_month || '-01')::timestamptz + interval '1 month')
    WHERE l.month_year = year_month
    GROUP BY l.user_id, l.score, p.username, p.avatar_url, p.rank
    ORDER BY l.score DESC
    LIMIT limit_count;
END;
$$ LANGUAGE plpgsql;

-- Classement par quiz
CREATE OR REPLACE FUNCTION get_quiz_leaderboard(quiz_id uuid)
RETURNS TABLE (
    rank bigint,
    user_id uuid,
    username text,
    avatar_url text,
    user_rank text,
    score integer,
    accuracy_rate numeric,
    time_taken_ms integer
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        ROW_NUMBER() OVER (ORDER BY gs.score DESC, gs.time_taken_ms ASC)::bigint as rank,
        gs.user_id,
        p.username,
        p.avatar_url,
        p.rank as user_rank,
        gs.score,
        gs.accuracy_rate,
        gs.time_taken_ms
    FROM game_sessions gs
    JOIN user_profiles p ON gs.user_id = p.id
    WHERE gs.quiz_id = quiz_id AND gs.completed_at IS NOT NULL
    ORDER BY gs.score DESC, gs.time_taken_ms ASC
    LIMIT 100;
END;
$$ LANGUAGE plpgsql;

-- Classement par série
CREATE OR REPLACE FUNCTION get_series_leaderboard(series_name text, limit_count int DEFAULT 100)
RETURNS TABLE (
    rank bigint,
    user_id uuid,
    username text,
    avatar_url text,
    user_rank text,
    total_score integer,
    quiz_count bigint
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        ROW_NUMBER() OVER (ORDER BY SUM(gs.score) DESC)::bigint as rank,
        gs.user_id,
        p.username,
        p.avatar_url,
        p.rank as user_rank,
        SUM(gs.score)::integer as total_score,
        COUNT(DISTINCT gs.quiz_id)::bigint as quiz_count
    FROM game_sessions gs
    JOIN quizzes q ON gs.quiz_id = q.id
    JOIN user_profiles p ON gs.user_id = p.id
    WHERE q.series = series_name AND gs.completed_at IS NOT NULL
    GROUP BY gs.user_id, p.username, p.avatar_url, p.rank
    ORDER BY total_score DESC
    LIMIT limit_count;
END;
$$ LANGUAGE plpgsql;

-- Vérifier et attribuer badges
CREATE OR REPLACE FUNCTION check_and_award_badges(target_user_id uuid)
RETURNS TABLE (badge_id uuid, badge_name text) AS $$
DECLARE
    user_stats_rec RECORD;
    user_xp integer;
    user_quizzes_played integer;
    user_quizzes_created integer;
    user_best_score integer;
    user_perfect_count integer;
    user_accuracy numeric;
BEGIN
    -- Récupérer les stats
    SELECT * INTO user_stats_rec FROM user_stats WHERE user_id = target_user_id;
    SELECT xp INTO user_xp FROM user_profiles WHERE id = target_user_id;

    user_quizzes_played := COALESCE(user_stats_rec.quizzes_played, 0);
    user_quizzes_created := COALESCE(user_stats_rec.quizzes_created, 0);
    user_best_score := COALESCE(user_stats_rec.best_score_ever, 0);
    user_accuracy := COALESCE(user_stats_rec.accuracy_rate, 0);

    SELECT COUNT(*) INTO user_perfect_count 
    FROM game_sessions 
    WHERE user_id = target_user_id AND is_perfect = true;

    RETURN QUERY
    SELECT b.id, b.name
    FROM badges b
    WHERE NOT EXISTS (
        SELECT 1 FROM user_badges ub 
        WHERE ub.badge_id = b.id AND ub.user_id = target_user_id
    )
    AND (
        (b.slug = 'first_quiz' AND user_quizzes_played >= 1)
        OR (b.slug = 'quiz_10' AND user_quizzes_played >= 10)
        OR (b.slug = 'quiz_100' AND user_quizzes_played >= 100)
        OR (b.slug = 'quiz_500' AND user_quizzes_played >= 500)
        OR (b.slug = 'quiz_1000' AND user_quizzes_played >= 1000)
        OR (b.slug = 'perfect_quiz' AND user_perfect_count >= 1)
        OR (b.slug = 'perfect_5' AND user_perfect_count >= 5)
        OR (b.slug = 'accuracy_90' AND user_accuracy >= 90 AND user_quizzes_played >= 20)
        OR (b.slug = 'first_creation' AND user_quizzes_created >= 1)
        OR (b.slug = 'creator_10' AND user_quizzes_created >= 10)
    );
END;
$$ LANGUAGE plpgsql;

-- ============================================================
-- ROW LEVEL SECURITY (RLS)
-- ============================================================

ALTER TABLE user_profiles ENABLE ROW LEVEL SECURITY;
ALTER TABLE quizzes ENABLE ROW LEVEL SECURITY;
ALTER TABLE questions ENABLE ROW LEVEL SECURITY;
ALTER TABLE answers ENABLE ROW LEVEL SECURITY;
ALTER TABLE game_sessions ENABLE ROW LEVEL SECURITY;
ALTER TABLE player_answers ENABLE ROW LEVEL SECURITY;
ALTER TABLE friendships ENABLE ROW LEVEL SECURITY;
ALTER TABLE user_badges ENABLE ROW LEVEL SECURITY;
ALTER TABLE reports ENABLE ROW LEVEL SECURITY;
ALTER TABLE leaderboard_monthly ENABLE ROW LEVEL SECURITY;
ALTER TABLE notifications ENABLE ROW LEVEL SECURITY;

-- user_profiles
CREATE POLICY "user_profiles_select_public" ON user_profiles FOR SELECT USING (true);
CREATE POLICY "user_profiles_update_own" ON user_profiles FOR UPDATE USING (auth.uid() = id);

-- quizzes
CREATE POLICY "quizzes_select_visible" ON quizzes FOR SELECT USING (is_visible = true OR auth.uid() = creator_id);
CREATE POLICY "quizzes_insert_own" ON quizzes FOR INSERT WITH CHECK (auth.uid() = creator_id);
CREATE POLICY "quizzes_update_own" ON quizzes FOR UPDATE USING (auth.uid() = creator_id);
CREATE POLICY "quizzes_delete_own" ON quizzes FOR DELETE USING (auth.uid() = creator_id);

-- questions
CREATE POLICY "questions_select_visible" ON questions FOR SELECT USING (
    EXISTS (SELECT 1 FROM quizzes WHERE quizzes.id = questions.quiz_id AND (quizzes.is_visible = true OR quizzes.creator_id = auth.uid()))
);
CREATE POLICY "questions_insert_own" ON questions FOR INSERT WITH CHECK (
    EXISTS (SELECT 1 FROM quizzes WHERE quizzes.id = questions.quiz_id AND quizzes.creator_id = auth.uid())
);
CREATE POLICY "questions_update_own" ON questions FOR UPDATE USING (
    EXISTS (SELECT 1 FROM quizzes WHERE quizzes.id = questions.quiz_id AND quizzes.creator_id = auth.uid())
);

-- answers
CREATE POLICY "answers_select_visible" ON answers FOR SELECT USING (
    EXISTS (SELECT 1 FROM questions q JOIN quizzes qu ON q.quiz_id = qu.id 
    WHERE q.id = answers.question_id AND (qu.is_visible = true OR qu.creator_id = auth.uid()))
);
CREATE POLICY "answers_insert_own" ON answers FOR INSERT WITH CHECK (
    EXISTS (SELECT 1 FROM questions q JOIN quizzes qu ON q.quiz_id = qu.id 
    WHERE q.id = answers.question_id AND qu.creator_id = auth.uid())
);

-- game_sessions
CREATE POLICY "game_sessions_select_own" ON game_sessions FOR SELECT USING (auth.uid() = user_id);
CREATE POLICY "game_sessions_insert_own" ON game_sessions FOR INSERT WITH CHECK (auth.uid() = user_id);
CREATE POLICY "game_sessions_update_own" ON game_sessions FOR UPDATE USING (auth.uid() = user_id);

-- player_answers
CREATE POLICY "player_answers_select_own" ON player_answers FOR SELECT USING (
    EXISTS (SELECT 1 FROM game_sessions WHERE game_sessions.id = player_answers.session_id AND game_sessions.user_id = auth.uid())
);
CREATE POLICY "player_answers_insert_own" ON player_answers FOR INSERT WITH CHECK (
    EXISTS (SELECT 1 FROM game_sessions WHERE game_sessions.id = player_answers.session_id AND game_sessions.user_id = auth.uid())
);

-- friendships
CREATE POLICY "friendships_select_own" ON friendships FOR SELECT USING (
    auth.uid() = requester_id OR auth.uid() = addressee_id
);
CREATE POLICY "friendships_insert_own" ON friendships FOR INSERT WITH CHECK (
    auth.uid() = requester_id
);
CREATE POLICY "friendships_update_own" ON friendships FOR UPDATE USING (
    auth.uid() = requester_id OR auth.uid() = addressee_id
);

-- user_badges
CREATE POLICY "user_badges_select_public" ON user_badges FOR SELECT USING (true);
CREATE POLICY "user_badges_insert_system" ON user_badges FOR INSERT WITH CHECK (false); -- Insert via fonction uniquement

-- reports
CREATE POLICY "reports_insert_auth" ON reports FOR INSERT WITH CHECK (auth.uid() = reporter_id);
CREATE POLICY "reports_select_own_admin" ON reports FOR SELECT USING (
    auth.uid() = reporter_id OR EXISTS (
        SELECT 1 FROM user_profiles WHERE id = auth.uid() AND rank = 'Légende'
    )
);

-- leaderboard_monthly
CREATE POLICY "leaderboard_select_public" ON leaderboard_monthly FOR SELECT USING (true);

-- notifications
CREATE POLICY "notifications_select_own" ON notifications FOR SELECT USING (auth.uid() = user_id);
CREATE POLICY "notifications_update_own" ON notifications FOR UPDATE USING (auth.uid() = user_id);
