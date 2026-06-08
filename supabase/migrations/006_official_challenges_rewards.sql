-- ============================================================
-- MIGRATION: Quiz Officiels, Défis, Récompenses
-- ============================================================

-- ============================================================
-- 1. MODIFIER TABLE QUIZZES
-- ============================================================
ALTER TABLE quizzes ADD COLUMN IF NOT EXISTS starts_at timestamptz;
ALTER TABLE quizzes ADD COLUMN IF NOT EXISTS ends_at timestamptz;
ALTER TABLE quizzes ADD COLUMN IF NOT EXISTS duration_seconds integer;
ALTER TABLE quizzes ADD COLUMN IF NOT EXISTS duration_mode text DEFAULT 'per_question';
ALTER TABLE quizzes ADD COLUMN IF NOT EXISTS leaderboard_public boolean DEFAULT true;
ALTER TABLE quizzes ADD COLUMN IF NOT EXISTS rewards jsonb DEFAULT '[]'::jsonb;

-- Ajouter les contraintes CHECK si elles n'existent pas
DO $$ 
BEGIN
    -- Supprimer et recréer la contrainte quiz_type
    IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'quizzes_quiz_type_check') THEN
        ALTER TABLE quizzes DROP CONSTRAINT quizzes_quiz_type_check;
    END IF;
    ALTER TABLE quizzes ADD CONSTRAINT quizzes_quiz_type_check CHECK (quiz_type IN ('community', 'official', 'private', 'challenge'));
    
    -- Supprimer et recréer la contrainte status
    IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'quizzes_status_check') THEN
        ALTER TABLE quizzes DROP CONSTRAINT quizzes_status_check;
    END IF;
    ALTER TABLE quizzes ADD CONSTRAINT quizzes_status_check CHECK (status IN ('draft', 'scheduled', 'active', 'published', 'hidden', 'archived', 'deleted'));
    
    -- Supprimer et recréer la contrainte duration_mode
    IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'quizzes_duration_mode_check') THEN
        ALTER TABLE quizzes DROP CONSTRAINT quizzes_duration_mode_check;
    END IF;
    ALTER TABLE quizzes ADD CONSTRAINT quizzes_duration_mode_check CHECK (duration_mode IN ('global', 'per_question'));
END $$;

CREATE INDEX IF NOT EXISTS idx_quizzes_type ON quizzes(quiz_type);
CREATE INDEX IF NOT EXISTS idx_quizzes_dates ON quizzes(starts_at, ends_at) WHERE quiz_type = 'official';

-- ============================================================
-- 2. TABLE CHALLENGE_SESSIONS
-- ============================================================
CREATE TABLE IF NOT EXISTS challenge_sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    quiz_id uuid NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    creator_id uuid NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    min_players integer DEFAULT 2 NOT NULL,
    invite_expires_at timestamptz NOT NULL,
    status text DEFAULT 'waiting',
    winner_id uuid REFERENCES user_profiles(id),
    total_xp_pool integer DEFAULT 0,
    started_at timestamptz,
    completed_at timestamptz,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'challenge_sessions_status_check') THEN
        ALTER TABLE challenge_sessions DROP CONSTRAINT challenge_sessions_status_check;
    END IF;
    ALTER TABLE challenge_sessions ADD CONSTRAINT challenge_sessions_status_check CHECK (status IN ('waiting', 'ready', 'playing', 'completed', 'cancelled', 'expired'));
END $$;

CREATE INDEX IF NOT EXISTS idx_challenge_sessions_quiz ON challenge_sessions(quiz_id);
CREATE INDEX IF NOT EXISTS idx_challenge_sessions_creator ON challenge_sessions(creator_id);
CREATE INDEX IF NOT EXISTS idx_challenge_sessions_status ON challenge_sessions(status);

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_challenge_sessions_updated_at') THEN
        CREATE TRIGGER update_challenge_sessions_updated_at
            BEFORE UPDATE ON challenge_sessions
            FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;
END $$;

-- ============================================================
-- 3. TABLE CHALLENGE_PARTICIPANTS
-- ============================================================
CREATE TABLE IF NOT EXISTS challenge_participants (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id uuid NOT NULL REFERENCES challenge_sessions(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    xp_bet integer DEFAULT 0 NOT NULL CHECK (xp_bet >= 0),
    status text DEFAULT 'pending',
    score integer DEFAULT 0,
    correct_count integer DEFAULT 0,
    accuracy_rate numeric(5,2) DEFAULT 0,
    time_taken_ms integer,
    xp_won integer DEFAULT 0,
    xp_lost integer DEFAULT 0,
    joined_at timestamptz DEFAULT now(),
    completed_at timestamptz,
    UNIQUE(session_id, user_id)
);

DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'challenge_participants_status_check') THEN
        ALTER TABLE challenge_participants DROP CONSTRAINT challenge_participants_status_check;
    END IF;
    ALTER TABLE challenge_participants ADD CONSTRAINT challenge_participants_status_check CHECK (status IN ('pending', 'accepted', 'refused', 'playing', 'done'));
END $$;

CREATE INDEX IF NOT EXISTS idx_challenge_participants_session ON challenge_participants(session_id);
CREATE INDEX IF NOT EXISTS idx_challenge_participants_user ON challenge_participants(user_id);

-- ============================================================
-- 4. TABLE CHALLENGE_INVITATIONS
-- ============================================================
CREATE TABLE IF NOT EXISTS challenge_invitations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id uuid NOT NULL REFERENCES challenge_sessions(id) ON DELETE CASCADE,
    inviter_id uuid NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    invitee_id uuid NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    status text DEFAULT 'pending',
    token text UNIQUE NOT NULL DEFAULT gen_random_uuid()::text,
    expires_at timestamptz NOT NULL,
    created_at timestamptz DEFAULT now(),
    UNIQUE(session_id, invitee_id)
);

DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'challenge_invitations_status_check') THEN
        ALTER TABLE challenge_invitations DROP CONSTRAINT challenge_invitations_status_check;
    END IF;
    ALTER TABLE challenge_invitations ADD CONSTRAINT challenge_invitations_status_check CHECK (status IN ('pending', 'accepted', 'refused', 'expired'));
END $$;

CREATE INDEX IF NOT EXISTS idx_challenge_invitations_session ON challenge_invitations(session_id);
CREATE INDEX IF NOT EXISTS idx_challenge_invitations_invitee ON challenge_invitations(invitee_id);
CREATE INDEX IF NOT EXISTS idx_challenge_invitations_token ON challenge_invitations(token);

-- ============================================================
-- 5. TABLE QUIZ_REWARDS
-- ============================================================
CREATE TABLE IF NOT EXISTS quiz_rewards (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    quiz_id uuid NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    title text NOT NULL,
    description text,
    url text,
    url_preview_title text,
    url_preview_image text,
    url_preview_domain text,
    rank_from integer NOT NULL CHECK (rank_from >= 1),
    rank_to integer NOT NULL CHECK (rank_to >= rank_from),
    created_at timestamptz DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_quiz_rewards_quiz ON quiz_rewards(quiz_id);

-- ============================================================
-- 6. TABLE XP_LEDGER
-- ============================================================
CREATE TABLE IF NOT EXISTS xp_ledger (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    amount integer NOT NULL,
    type text NOT NULL,
    reference_type text NOT NULL,
    reference_id uuid NOT NULL,
    created_at timestamptz DEFAULT now()
);

DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'xp_ledger_type_check') THEN
        ALTER TABLE xp_ledger DROP CONSTRAINT xp_ledger_type_check;
    END IF;
    ALTER TABLE xp_ledger ADD CONSTRAINT xp_ledger_type_check CHECK (type IN ('frozen', 'unfrozen', 'won', 'lost'));
    
    IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'xp_ledger_reference_type_check') THEN
        ALTER TABLE xp_ledger DROP CONSTRAINT xp_ledger_reference_type_check;
    END IF;
    ALTER TABLE xp_ledger ADD CONSTRAINT xp_ledger_reference_type_check CHECK (reference_type IN ('challenge', 'reward'));
END $$;

CREATE INDEX IF NOT EXISTS idx_xp_ledger_user ON xp_ledger(user_id);
CREATE INDEX IF NOT EXISTS idx_xp_ledger_reference ON xp_ledger(reference_type, reference_id);

-- ============================================================
-- 7. TABLE OFFICIAL_LEADERBOARD
-- ============================================================
CREATE TABLE IF NOT EXISTS official_leaderboard (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    quiz_id uuid NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    session_id uuid NOT NULL REFERENCES game_sessions(id) ON DELETE CASCADE,
    score integer NOT NULL,
    rank_position integer,
    accuracy_rate numeric(5,2),
    time_taken_ms integer,
    created_at timestamptz DEFAULT now(),
    UNIQUE(quiz_id, user_id, session_id)
);

CREATE INDEX IF NOT EXISTS idx_official_leaderboard_quiz ON official_leaderboard(quiz_id, score DESC);
CREATE INDEX IF NOT EXISTS idx_official_leaderboard_user ON official_leaderboard(user_id);

-- ============================================================
-- 8. FONCTIONS
-- ============================================================

CREATE OR REPLACE FUNCTION auto_archive_expired_official_quizzes()
RETURNS void AS $$
BEGIN
    UPDATE quizzes
    SET status = 'archived'
    WHERE quiz_type = 'official'
    AND status = 'active'
    AND ends_at IS NOT NULL
    AND ends_at < now();
END;
$$ LANGUAGE plpgsql;

-- ============================================================
-- 9. RLS POLICIES
-- ============================================================

ALTER TABLE challenge_sessions ENABLE ROW LEVEL SECURITY;
ALTER TABLE challenge_participants ENABLE ROW LEVEL SECURITY;
ALTER TABLE challenge_invitations ENABLE ROW LEVEL SECURITY;
ALTER TABLE quiz_rewards ENABLE ROW LEVEL SECURITY;
ALTER TABLE xp_ledger ENABLE ROW LEVEL SECURITY;
ALTER TABLE official_leaderboard ENABLE ROW LEVEL SECURITY;

-- Supprimer les politiques existantes si elles existent et les recréer
DO $$ 
BEGIN
    -- challenge_sessions
    DROP POLICY IF EXISTS "challenge_sessions_select_participants" ON challenge_sessions;
    DROP POLICY IF EXISTS "challenge_sessions_insert_creator" ON challenge_sessions;
    DROP POLICY IF EXISTS "challenge_sessions_update_creator" ON challenge_sessions;
    
    -- challenge_participants
    DROP POLICY IF EXISTS "challenge_participants_select_session" ON challenge_participants;
    DROP POLICY IF EXISTS "challenge_participants_insert_self" ON challenge_participants;
    DROP POLICY IF EXISTS "challenge_participants_update_self" ON challenge_participants;
    
    -- challenge_invitations
    DROP POLICY IF EXISTS "challenge_invitations_select_involved" ON challenge_invitations;
    DROP POLICY IF EXISTS "challenge_invitations_insert_inviter" ON challenge_invitations;
    DROP POLICY IF EXISTS "challenge_invitations_update_invitee" ON challenge_invitations;
    
    -- quiz_rewards
    DROP POLICY IF EXISTS "quiz_rewards_select_public" ON quiz_rewards;
    DROP POLICY IF EXISTS "quiz_rewards_insert_admin" ON quiz_rewards;
    DROP POLICY IF EXISTS "quiz_rewards_update_admin" ON quiz_rewards;
    DROP POLICY IF EXISTS "quiz_rewards_delete_admin" ON quiz_rewards;
    
    -- xp_ledger
    DROP POLICY IF EXISTS "xp_ledger_select_own" ON xp_ledger;
    
    -- official_leaderboard
    DROP POLICY IF EXISTS "official_leaderboard_select_public" ON official_leaderboard;
    DROP POLICY IF EXISTS "official_leaderboard_insert_system" ON official_leaderboard;
END $$;

-- Recréer les politiques
CREATE POLICY "challenge_sessions_select_participants" ON challenge_sessions FOR SELECT USING (
    auth.uid() = creator_id OR 
    EXISTS (SELECT 1 FROM challenge_participants WHERE session_id = challenge_sessions.id AND user_id = auth.uid())
);
CREATE POLICY "challenge_sessions_insert_creator" ON challenge_sessions FOR INSERT WITH CHECK (auth.uid() = creator_id);
CREATE POLICY "challenge_sessions_update_creator" ON challenge_sessions FOR UPDATE USING (auth.uid() = creator_id);

CREATE POLICY "challenge_participants_select_session" ON challenge_participants FOR SELECT USING (
    EXISTS (
        SELECT 1 FROM challenge_sessions 
        WHERE id = challenge_participants.session_id 
        AND (creator_id = auth.uid() OR EXISTS (
            SELECT 1 FROM challenge_participants cp2 
            WHERE cp2.session_id = challenge_sessions.id AND cp2.user_id = auth.uid()
        ))
    )
);
CREATE POLICY "challenge_participants_insert_self" ON challenge_participants FOR INSERT WITH CHECK (auth.uid() = user_id);
CREATE POLICY "challenge_participants_update_self" ON challenge_participants FOR UPDATE USING (auth.uid() = user_id);

CREATE POLICY "challenge_invitations_select_involved" ON challenge_invitations FOR SELECT USING (
    auth.uid() = inviter_id OR auth.uid() = invitee_id
);
CREATE POLICY "challenge_invitations_insert_inviter" ON challenge_invitations FOR INSERT WITH CHECK (auth.uid() = inviter_id);
CREATE POLICY "challenge_invitations_update_invitee" ON challenge_invitations FOR UPDATE USING (auth.uid() = invitee_id);

CREATE POLICY "quiz_rewards_select_public" ON quiz_rewards FOR SELECT USING (true);
CREATE POLICY "quiz_rewards_insert_admin" ON quiz_rewards FOR INSERT WITH CHECK (
    EXISTS (SELECT 1 FROM user_profiles WHERE id = auth.uid() AND is_admin = true)
);
CREATE POLICY "quiz_rewards_update_admin" ON quiz_rewards FOR UPDATE USING (
    EXISTS (SELECT 1 FROM user_profiles WHERE id = auth.uid() AND is_admin = true)
);
CREATE POLICY "quiz_rewards_delete_admin" ON quiz_rewards FOR DELETE USING (
    EXISTS (SELECT 1 FROM user_profiles WHERE id = auth.uid() AND is_admin = true)
);

CREATE POLICY "xp_ledger_select_own" ON xp_ledger FOR SELECT USING (auth.uid() = user_id);

CREATE POLICY "official_leaderboard_select_public" ON official_leaderboard FOR SELECT USING (true);
CREATE POLICY "official_leaderboard_insert_system" ON official_leaderboard FOR INSERT WITH CHECK (true);