-- Migration 032 : Système de renchérissement XP des défis

ALTER TABLE challenge_participants 
  ADD COLUMN IF NOT EXISTS individual_bet integer DEFAULT 0,
  ADD COLUMN IF NOT EXISTS bet_version integer DEFAULT 0,
  ADD COLUMN IF NOT EXISTS bet_accepted boolean DEFAULT true;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'challenge_participants_status_check') THEN
    ALTER TABLE challenge_participants DROP CONSTRAINT challenge_participants_status_check;
  END IF;
  ALTER TABLE challenge_participants ADD CONSTRAINT challenge_participants_status_check 
    CHECK (status IN ('accepted', 'refused', 'left'));
END $$;

DROP POLICY IF EXISTS "challenge_participants_insert" ON challenge_participants;
DROP POLICY IF EXISTS "challenge_participants_update" ON challenge_participants;
DROP POLICY IF EXISTS "challenge_participants_select" ON challenge_participants;

CREATE POLICY "challenge_participants_select" ON challenge_participants FOR SELECT USING (
  auth.uid() IN (SELECT user_id FROM challenge_participants WHERE session_id = challenge_participants.session_id)
  OR auth.uid() = (SELECT creator_id FROM challenge_sessions WHERE id = challenge_participants.session_id)
);

CREATE POLICY "challenge_participants_insert" ON challenge_participants FOR INSERT WITH CHECK (
  auth.uid() = user_id
  OR EXISTS (SELECT 1 FROM challenge_sessions WHERE id = session_id AND creator_id = auth.uid())
);

CREATE POLICY "challenge_participants_update" ON challenge_participants FOR UPDATE USING (
  auth.uid() = user_id
  OR EXISTS (SELECT 1 FROM challenge_sessions WHERE id = session_id AND creator_id = auth.uid())
);
