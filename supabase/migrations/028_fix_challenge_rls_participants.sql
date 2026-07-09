-- FIX: Allow participants to view challenge sessions they're part of

-- ============================================================
-- challenge_sessions
-- ============================================================
DROP POLICY IF EXISTS "challenge_sessions_all" ON challenge_sessions;

CREATE POLICY "challenge_sessions_select" ON challenge_sessions FOR SELECT USING (
    auth.uid() = creator_id
    OR EXISTS (
        SELECT 1 FROM challenge_participants cp
        WHERE cp.session_id = challenge_sessions.id
        AND cp.user_id = auth.uid()
    )
);

CREATE POLICY "challenge_sessions_insert" ON challenge_sessions FOR INSERT WITH CHECK (
    auth.uid() = creator_id
);

CREATE POLICY "challenge_sessions_update" ON challenge_sessions FOR UPDATE USING (
    auth.uid() = creator_id
) WITH CHECK (
    auth.uid() = creator_id
);

-- ============================================================
-- challenge_participants — allow seeing all participants of a session you belong to
-- ============================================================
DROP POLICY IF EXISTS "challenge_participants_select" ON challenge_participants;

CREATE POLICY "challenge_participants_select" ON challenge_participants FOR SELECT USING (
    auth.uid() = user_id
    OR EXISTS (
        SELECT 1 FROM challenge_sessions cs
        WHERE cs.id = challenge_participants.session_id
        AND (
            cs.creator_id = auth.uid()
            OR EXISTS (
                SELECT 1 FROM challenge_participants cp2
                WHERE cp2.session_id = cs.id
                AND cp2.user_id = auth.uid()
            )
        )
    )
);