-- Add 'challenge_invitation' to notifications type CHECK constraint
ALTER TABLE notifications DROP CONSTRAINT IF EXISTS notifications_type_check;
ALTER TABLE notifications ADD CONSTRAINT notifications_type_check
    CHECK (type IN ('friend_request', 'badge_unlocked', 'quiz_completed', 'event_starting', 'rank_up', 'challenge_invitation'));
