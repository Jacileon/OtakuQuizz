-- Quiz likes/dislikes
CREATE TABLE IF NOT EXISTS quiz_votes (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    quiz_id uuid NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    vote_type text NOT NULL CHECK (vote_type IN ('like', 'dislike')),
    created_at timestamptz DEFAULT now(),
    UNIQUE(quiz_id, user_id)
);

-- Add like/dislike counts to quizzes
ALTER TABLE quizzes ADD COLUMN IF NOT EXISTS like_count integer DEFAULT 0;
ALTER TABLE quizzes ADD COLUMN IF NOT EXISTS dislike_count integer DEFAULT 0;

-- Profile photo URL column (may already exist)
ALTER TABLE user_profiles ADD COLUMN IF NOT EXISTS avatar_url text;

-- RLS
ALTER TABLE quiz_votes ENABLE ROW LEVEL SECURITY;
CREATE POLICY "quiz_votes_select" ON quiz_votes FOR SELECT USING (true);
CREATE POLICY "quiz_votes_insert" ON quiz_votes FOR INSERT WITH CHECK (auth.uid() = user_id);
CREATE POLICY "quiz_votes_delete" ON quiz_votes FOR DELETE USING (auth.uid() = user_id);
CREATE POLICY "quiz_votes_update" ON quiz_votes FOR UPDATE USING (auth.uid() = user_id);
