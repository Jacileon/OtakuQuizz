-- Table pour les signalements d'utilisateurs
CREATE TABLE IF NOT EXISTS user_reports (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    reporter_id uuid NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    reported_user_id uuid NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    reason text NOT NULL,
    description text,
    status text DEFAULT 'pending' CHECK (status IN ('pending', 'reviewed', 'resolved', 'dismissed')),
    created_at timestamptz DEFAULT now(),
    UNIQUE(reporter_id, reported_user_id)
);

CREATE INDEX IF NOT EXISTS idx_user_reports_reported ON user_reports(reported_user_id);
CREATE INDEX IF NOT EXISTS idx_user_reports_status ON user_reports(status);

-- RLS
ALTER TABLE user_reports ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS "user_reports_insert_own" ON user_reports;
CREATE POLICY "user_reports_insert_own" ON user_reports FOR INSERT WITH CHECK (
    auth.uid() = reporter_id
);

DROP POLICY IF EXISTS "user_reports_select_own" ON user_reports;
CREATE POLICY "user_reports_select_own" ON user_reports FOR SELECT USING (
    auth.uid() = reporter_id OR 
    EXISTS (SELECT 1 FROM user_profiles WHERE id = auth.uid() AND is_admin = true)
);