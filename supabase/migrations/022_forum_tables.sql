-- Forum community tables
CREATE TABLE IF NOT EXISTS forum_channels (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL,
    description text,
    channel_type text DEFAULT 'community' CHECK (channel_type IN ('community', 'ideas')),
    created_at timestamptz DEFAULT now()
);

CREATE TABLE IF NOT EXISTS forum_messages (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id uuid NOT NULL REFERENCES forum_channels(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    content text NOT NULL CHECK (length(content) > 0 AND length(content) <= 5000),
    created_at timestamptz DEFAULT now()
);

-- Seed default channels
INSERT INTO forum_channels (name, description, channel_type) VALUES
('💬 Discussion Générale', 'Parlez de tout et de rien autour des anime et manga', 'community'),
('💡 Boîte à Idées', 'Proposez vos idées pour améliorer Otaku Quiz Africa', 'community');

-- RLS policies
ALTER TABLE forum_channels ENABLE ROW LEVEL SECURITY;
ALTER TABLE forum_messages ENABLE ROW LEVEL SECURITY;

CREATE POLICY "forum_channels_select" ON forum_channels FOR SELECT USING (true);
CREATE POLICY "forum_messages_select" ON forum_messages FOR SELECT USING (true);
CREATE POLICY "forum_messages_insert" ON forum_messages FOR INSERT WITH CHECK (auth.uid() = user_id);
