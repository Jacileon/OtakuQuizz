-- Fix: Permettre aux participants de voir les sessions de défi

-- Supprimer l'ancienne politique
DROP POLICY IF EXISTS "challenge_sessions_all" ON challenge_sessions;

-- Politiques séparées
CREATE POLICY "challenge_sessions_select" ON challenge_sessions FOR SELECT USING (true);
CREATE POLICY "challenge_sessions_insert" ON challenge_sessions FOR INSERT WITH CHECK (auth.uid() = creator_id);
CREATE POLICY "challenge_sessions_update" ON challenge_sessions FOR UPDATE USING (auth.uid() = creator_id);
CREATE POLICY "challenge_sessions_delete" ON challenge_sessions FOR DELETE USING (auth.uid() = creator_id);