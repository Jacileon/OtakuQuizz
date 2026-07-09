CREATE TABLE IF NOT EXISTS challenge_scores (
  id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
  session_id uuid NOT NULL REFERENCES challenge_sessions(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
  correct_count int NOT NULL DEFAULT 0,
  total_questions int NOT NULL DEFAULT 0,
  time_taken_ms int NOT NULL DEFAULT 0,
  created_at timestamptz DEFAULT now(),
  UNIQUE(session_id, user_id)
);

-- Add column if missing (for tables created before this column was added)
ALTER TABLE challenge_scores ADD COLUMN IF NOT EXISTS time_taken_ms int NOT NULL DEFAULT 0;

ALTER TABLE challenge_scores ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS "challenge_scores_select" ON challenge_scores;
CREATE POLICY "challenge_scores_select"
  ON challenge_scores FOR SELECT
  USING (
    auth.uid() = user_id
    OR EXISTS (
      SELECT 1 FROM challenge_participants
      WHERE session_id = challenge_scores.session_id
      AND user_id = auth.uid()
      AND status = 'accepted'
    )
    OR EXISTS (
      SELECT 1 FROM challenge_invitations
      WHERE session_id = challenge_scores.session_id
      AND invitee_id = auth.uid()
    )
  );

DROP POLICY IF EXISTS "challenge_scores_insert" ON challenge_scores;
CREATE POLICY "challenge_scores_insert"
  ON challenge_scores FOR INSERT
  WITH CHECK (auth.uid() = user_id);
