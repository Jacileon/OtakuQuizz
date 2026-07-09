-- Add up_votes and down_votes columns to forum_suggestions
ALTER TABLE forum_suggestions
ADD COLUMN IF NOT EXISTS up_votes INTEGER NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS down_votes INTEGER NOT NULL DEFAULT 0;

-- Create index for sorting by votes
CREATE INDEX IF NOT EXISTS idx_suggestions_votes ON forum_suggestions(up_votes DESC, down_votes ASC);