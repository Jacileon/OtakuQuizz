-- FIX: Supprimer TOUTES les politiques challenge et recréer sans récursion

-- challenge_sessions: supprimer toutes les politiques
DROP POLICY IF EXISTS "challenge_sessions_select_own" ON challenge_sessions;
DROP POLICY IF EXISTS "challenge_sessions_select_participant" ON challenge_sessions;
DROP POLICY IF EXISTS "challenge_sessions_insert_own" ON challenge_sessions;
DROP POLICY IF EXISTS "challenge_sessions_update_own" ON challenge_sessions;
DROP POLICY IF EXISTS "challenge_sessions_select_participants" ON challenge_sessions;
DROP POLICY IF EXISTS "challenge_sessions_insert_creator" ON challenge_sessions;
DROP POLICY IF EXISTS "challenge_sessions_update_creator" ON challenge_sessions;

-- Politiques simples sans sous-requête croisée
CREATE POLICY "challenge_sessions_all" ON challenge_sessions FOR ALL USING (
    auth.uid() = creator_id
) WITH CHECK (
    auth.uid() = creator_id
);

-- challenge_participants: supprimer toutes les politiques
DROP POLICY IF EXISTS "challenge_participants_select_own" ON challenge_participants;
DROP POLICY IF EXISTS "challenge_participants_insert_own" ON challenge_participants;
DROP POLICY IF EXISTS "challenge_participants_update_own" ON challenge_participants;
DROP POLICY IF EXISTS "challenge_participants_select_session" ON challenge_participants;
DROP POLICY IF EXISTS "challenge_participants_insert_self" ON challenge_participants;
DROP POLICY IF EXISTS "challenge_participants_update_self" ON challenge_participants;

-- Politiques simples
CREATE POLICY "challenge_participants_select" ON challenge_participants FOR SELECT USING (
    auth.uid() = user_id
);
CREATE POLICY "challenge_participants_insert" ON challenge_participants FOR INSERT WITH CHECK (
    auth.uid() = user_id
);
CREATE POLICY "challenge_participants_update" ON challenge_participants FOR UPDATE USING (
    auth.uid() = user_id
);

-- challenge_invitations: supprimer toutes les politiques
DROP POLICY IF EXISTS "challenge_invitations_select_own" ON challenge_invitations;
DROP POLICY IF EXISTS "challenge_invitations_insert_own" ON challenge_invitations;
DROP POLICY IF EXISTS "challenge_invitations_update_own" ON challenge_invitations;
DROP POLICY IF EXISTS "challenge_invitations_select_involved" ON challenge_invitations;
DROP POLICY IF EXISTS "challenge_invitations_insert_inviter" ON challenge_invitations;
DROP POLICY IF EXISTS "challenge_invitations_update_invitee" ON challenge_invitations;

CREATE POLICY "challenge_invitations_select" ON challenge_invitations FOR SELECT USING (
    auth.uid() = inviter_id OR auth.uid() = invitee_id
);
CREATE POLICY "challenge_invitations_insert" ON challenge_invitations FOR INSERT WITH CHECK (
    auth.uid() = inviter_id
);
CREATE POLICY "challenge_invitations_update" ON challenge_invitations FOR UPDATE USING (
    auth.uid() = invitee_id
);