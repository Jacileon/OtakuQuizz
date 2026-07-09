-- Forum suggestions (Boîte à Idées)
CREATE TABLE IF NOT EXISTS forum_suggestions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id UUID NOT NULL REFERENCES forum_channels(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    title TEXT NOT NULL CHECK (length(title) > 0 AND length(title) <= 200),
    content TEXT NOT NULL CHECK (length(content) > 0 AND length(content) <= 2000),
    week_label TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Votes on suggestions (pouce haut/bas)
CREATE TABLE IF NOT EXISTS forum_suggestion_votes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    suggestion_id UUID NOT NULL REFERENCES forum_suggestions(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    vote_type TEXT NOT NULL CHECK (vote_type IN ('up', 'down')),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(suggestion_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_suggestions_channel ON forum_suggestions(channel_id);
CREATE INDEX IF NOT EXISTS idx_suggestions_week ON forum_suggestions(week_label);
CREATE INDEX IF NOT EXISTS idx_suggestion_votes_suggestion ON forum_suggestion_votes(suggestion_id);

-- RLS
ALTER TABLE forum_suggestions ENABLE ROW LEVEL SECURITY;
ALTER TABLE forum_suggestion_votes ENABLE ROW LEVEL SECURITY;

CREATE POLICY "forum_suggestions_select" ON forum_suggestions FOR SELECT USING (true);
CREATE POLICY "forum_suggestions_insert" ON forum_suggestions FOR INSERT WITH CHECK (true);
CREATE POLICY "forum_suggestions_delete" ON forum_suggestions FOR DELETE USING (true);

CREATE POLICY "forum_suggestion_votes_select" ON forum_suggestion_votes FOR SELECT USING (true);
CREATE POLICY "forum_suggestion_votes_insert" ON forum_suggestion_votes FOR INSERT WITH CHECK (true);
CREATE POLICY "forum_suggestion_votes_delete" ON forum_suggestion_votes FOR DELETE USING (true);
CREATE POLICY "forum_suggestion_votes_update" ON forum_suggestion_votes FOR UPDATE USING (true);
