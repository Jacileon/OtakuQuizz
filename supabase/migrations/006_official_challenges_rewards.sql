-- ============================================================
-- MIGRATION: Quiz Officiels, Défis, Récompenses
-- ============================================================

-- ============================================================
-- 1. MODIFIER TABLE QUIZZES
-- ============================================================
ALTER TABLE quizzes ADD COLUMN IF NOT EXISTS starts_at timestamptz;
ALTER TABLE quizzes ADD COLUMN IF NOT EXISTS ends_at timestamptz;
ALTER TABLE quizzes ADD COLUMN IF NOT EXISTS duration_seconds integer;
ALTER TABLE quizzes ADD COLUMN IF NOT EXISTS duration_mode text DEFAULT 'per_question' CHECK (duration_mode IN ('global', 'per_question'));
ALTER TABLE quizzes ADD COLUMN IF NOT EXISTS leaderboard_public boolean DEFAULT true;
ALTER TABLE quizzes ADD COLUMN IF NOT EXISTS rewards jsonb DEFAULT '[]'::jsonb;

-- Mettre à jour le type de quiz pour inclure 'challenge'
ALTER TABLE quizzes DROP CONSTRAINT IF EXISTS quizzes_quiz_type_check;
ALTER TABLE quizzes ADD CONSTRAINT quizzes_quiz_type_check CHECK (quiz_type IN ('community', 'official', 'private', 'challenge'));

-- Mettre à jour le statut pour inclure 'scheduled' et 'archived'
ALTER TABLE quizzes DROP CONSTRAINT IF EXISTS quizzes_status_check;
ALTER TABLE quizzes ADD CONSTRAINT quizzes_status_check CHECK (status IN ('draft', 'scheduled', 'active', 'published', 'hidden', 'archived', 'deleted'));

CREATE INDEX idx_quizzes_type ON quizzes(quiz_type);
CREATE INDEX idx_quizzes_dates ON quizzes(starts_at, ends_at) WHERE quiz_type = 'official';

-- ============================================================
-- 2. TABLE CHALLENGE_SESSIONS
-- ============================================================
CREATE TABLE challenge_sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    quiz_id uuid NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    creator_id uuid NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    min_players integer DEFAULT 2 NOT NULL,
    invite_expires_at timestamptz NOT NULL,
    status text DEFAULT 'waiting' CHECK (status IN ('waiting', 'ready', 'playing', 'completed', 'cancelled', 'expired')),
    winner_id uuid REFERENCES user_profiles(id),
    total_xp_pool integer DEFAULT 0,
    started_at timestamptz,
    completed_at timestamptz,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE INDEX idx_challenge_sessions_quiz ON challenge_sessions(quiz_id);
CREATE INDEX idx_challenge_sessions_creator ON challenge_sessions(creator_id);
CREATE INDEX idx_challenge_sessions_status ON challenge_sessions(status);

CREATE TRIGGER update_challenge_sessions_updated_at
    BEFORE UPDATE ON challenge_sessions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================
-- 3. TABLE CHALLENGE_PARTICIPANTS
-- ============================================================
CREATE TABLE challenge_participants (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id uuid NOT NULL REFERENCES challenge_sessions(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    xp_bet integer DEFAULT 0 NOT NULL CHECK (xp_bet >= 0),
    status text DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'refused', 'playing', 'done')),
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

CREATE INDEX idx_challenge_participants_session ON challenge_participants(session_id);
CREATE INDEX idx_challenge_participants_user ON challenge_participants(user_id);

-- ============================================================
-- 4. TABLE CHALLENGE_INVITATIONS
-- ============================================================
CREATE TABLE challenge_invitations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id uuid NOT NULL REFERENCES challenge_sessions(id) ON DELETE CASCADE,
    inviter_id uuid NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    invitee_id uuid NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    status text DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'refused', 'expired')),
    token text UNIQUE NOT NULL DEFAULT gen_random_uuid()::text,
    expires_at timestamptz NOT NULL,
    created_at timestamptz DEFAULT now(),
    UNIQUE(session_id, invitee_id)
);

CREATE INDEX idx_challenge_invitations_session ON challenge_invitations(session_id);
CREATE INDEX idx_challenge_invitations_invitee ON challenge_invitations(invitee_id);
CREATE INDEX idx_challenge_invitations_token ON challenge_invitations(token);

-- ============================================================
-- 5. TABLE QUIZ_REWARDS (pour les récompenses détaillées)
-- ============================================================
CREATE TABLE quiz_rewards (
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

CREATE INDEX idx_quiz_rewards_quiz ON quiz_rewards(quiz_id);

-- ============================================================
-- 6. TABLE XP_LEDGER (pour le gel/restitution XP)
-- ============================================================
CREATE TABLE xp_ledger (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    amount integer NOT NULL,
    type text NOT NULL CHECK (type IN ('frozen', 'unfrozen', 'won', 'lost')),
    reference_type text NOT NULL CHECK (reference_type IN ('challenge', 'reward')),
    reference_id uuid NOT NULL,
    created_at timestamptz DEFAULT now()
);

CREATE INDEX idx_xp_ledger_user ON xp_ledger(user_id);
CREATE INDEX idx_xp_ledger_reference ON xp_ledger(reference_type, reference_id);

-- ============================================================
-- 7. TABLE OFFICIAL_LEADERBOARD (classement permanent)
-- ============================================================
CREATE TABLE official_leaderboard (
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

CREATE INDEX idx_official_leaderboard_quiz ON official_leaderboard(quiz_id, score DESC);
CREATE INDEX idx_official_leaderboard_user ON official_leaderboard(user_id);

-- ============================================================
-- 8. FONCTIONS ET TRIGGERS
-- ============================================================

-- Fonction pour archiver automatiquement les quiz officiels expirés
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

-- Fonction pour vérifier la limite de participation aux défis
CREATE OR REPLACE FUNCTION check_challenge_participation_limit()
RETURNS TRIGGER AS $$
DECLARE
    participation_count integer;
BEGIN
    SELECT COUNT(*) INTO participation_count
    FROM challenge_participants cp
    JOIN challenge_sessions cs ON cp.session_id = cs.id
    WHERE cp.user_id = NEW.user_id
    AND cs.quiz_id = (SELECT quiz_id FROM challenge_sessions WHERE id = NEW.session_id);
    
    IF participation_count >= 3 THEN
        RAISE EXCEPTION 'Limite de 3 participations atteinte pour ce quiz';
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_check_challenge_participation_limit
    BEFORE INSERT ON challenge_participants
    FOR EACH ROW EXECUTE FUNCTION check_challenge_participation_limit();

-- Fonction pour geler l'XP lors d'une mise
CREATE OR REPLACE FUNCTION freeze_xp_on_bet()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.xp_bet > 0 AND NEW.status = 'accepted' AND (OLD IS NULL OR OLD.status != 'accepted') THEN
        -- Vérifier le solde
        IF (SELECT xp FROM user_profiles WHERE id = NEW.user_id) < NEW.xp_bet THEN
            RAISE EXCEPTION 'Solde XP insuffisant';
        END IF;
        
        -- Geler l'XP
        UPDATE user_profiles SET xp = xp - NEW.xp_bet WHERE id = NEW.user_id;
        
        -- Enregistrer dans le ledger
        INSERT INTO xp_ledger (user_id, amount, type, reference_type, reference_id)
        VALUES (NEW.user_id, NEW.xp_bet, 'frozen', 'challenge', NEW.session_id);
        
        -- Mettre à jour le pool total
        UPDATE challenge_sessions 
        SET total_xp_pool = total_xp_pool + NEW.xp_bet 
        WHERE id = NEW.session_id;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_freeze_xp_on_bet
    AFTER INSERT OR UPDATE ON challenge_participants
    FOR EACH ROW EXECUTE FUNCTION freeze_xp_on_bet();

-- Fonction pour restituer l'XP en cas d'annulation
CREATE OR REPLACE FUNCTION refund_xp_on_cancel()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status = 'cancelled' AND OLD.status != 'cancelled' THEN
        -- Restituer l'XP à tous les participants
        UPDATE user_profiles 
        SET xp = xp + cp.xp_bet
        FROM challenge_participants cp
        WHERE cp.session_id = NEW.id
        AND user_profiles.id = cp.user_id
        AND cp.xp_bet > 0;
        
        -- Enregistrer dans le ledger
        INSERT INTO xp_ledger (user_id, amount, type, reference_type, reference_id)
        SELECT cp.user_id, cp.xp_bet, 'unfrozen', 'challenge', NEW.id
        FROM challenge_participants cp
        WHERE cp.session_id = NEW.id
        AND cp.xp_bet > 0;
        
        -- Reset le pool
        UPDATE challenge_sessions SET total_xp_pool = 0 WHERE id = NEW.id;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_refund_xp_on_cancel
    AFTER UPDATE ON challenge_sessions
    FOR EACH ROW EXECUTE FUNCTION refund_xp_on_cancel();

-- ============================================================
-- 9. ROW LEVEL SECURITY
-- ============================================================

ALTER TABLE challenge_sessions ENABLE ROW LEVEL SECURITY;
ALTER TABLE challenge_participants ENABLE ROW LEVEL SECURITY;
ALTER TABLE challenge_invitations ENABLE ROW LEVEL SECURITY;
ALTER TABLE quiz_rewards ENABLE ROW LEVEL SECURITY;
ALTER TABLE xp_ledger ENABLE ROW LEVEL SECURITY;
ALTER TABLE official_leaderboard ENABLE ROW LEVEL SECURITY;

-- challenge_sessions
CREATE POLICY "challenge_sessions_select_participants" ON challenge_sessions FOR SELECT USING (
    auth.uid() = creator_id OR 
    EXISTS (SELECT 1 FROM challenge_participants WHERE session_id = challenge_sessions.id AND user_id = auth.uid())
);
CREATE POLICY "challenge_sessions_insert_creator" ON challenge_sessions FOR INSERT WITH CHECK (auth.uid() = creator_id);
CREATE POLICY "challenge_sessions_update_creator" ON challenge_sessions FOR UPDATE USING (auth.uid() = creator_id);

-- challenge_participants
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

-- challenge_invitations
CREATE POLICY "challenge_invitations_select_involved" ON challenge_invitations FOR SELECT USING (
    auth.uid() = inviter_id OR auth.uid() = invitee_id
);
CREATE POLICY "challenge_invitations_insert_inviter" ON challenge_invitations FOR INSERT WITH CHECK (auth.uid() = inviter_id);
CREATE POLICY "challenge_invitations_update_invitee" ON challenge_invitations FOR UPDATE USING (auth.uid() = invitee_id);

-- quiz_rewards
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

-- xp_ledger
CREATE POLICY "xp_ledger_select_own" ON xp_ledger FOR SELECT USING (auth.uid() = user_id);

-- official_leaderboard
CREATE POLICY "official_leaderboard_select_public" ON official_leaderboard FOR SELECT USING (true);
CREATE POLICY "official_leaderboard_insert_system" ON official_leaderboard FOR INSERT WITH CHECK (true);