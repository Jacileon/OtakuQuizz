-- Forum message edit + report support
ALTER TABLE forum_messages ADD COLUMN IF NOT EXISTS updated_at timestamptz;

CREATE TABLE IF NOT EXISTS forum_message_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id UUID NOT NULL REFERENCES forum_messages(id) ON DELETE CASCADE,
    reporter_id UUID NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    reason text NOT NULL CHECK (length(reason) > 0 AND length(reason) <= 500),
    created_at timestamptz DEFAULT now(),
    UNIQUE(message_id, reporter_id)
);

ALTER TABLE forum_message_reports ENABLE ROW LEVEL SECURITY;

CREATE POLICY "forum_message_reports_select" ON forum_message_reports FOR SELECT USING (true);
CREATE POLICY "forum_message_reports_insert" ON forum_message_reports FOR INSERT WITH CHECK (true);
