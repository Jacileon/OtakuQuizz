ALTER TABLE challenge_sessions ADD COLUMN IF NOT EXISTS reward_mode text NOT NULL DEFAULT 'all_for_one';
ALTER TABLE challenge_sessions ADD COLUMN IF NOT EXISTS completed_at timestamptz;
