-- Backfill quiz_leaderboard from existing game_sessions + user_quiz_attempts
INSERT INTO quiz_leaderboard (quiz_id, user_id, xp_earned, accuracy_rate, time_taken_ms, attempt_number)
SELECT DISTINCT ON (ua.quiz_id, ua.user_id)
  ua.quiz_id,
  ua.user_id,
  ua.xp_earned,
  gs.accuracy_rate,
  gs.time_taken_ms,
  ua.attempt_number
FROM user_quiz_attempts ua
JOIN game_sessions gs ON gs.user_id = ua.user_id AND gs.quiz_id = ua.quiz_id
WHERE gs.completed_at IS NOT NULL
ORDER BY ua.quiz_id, ua.user_id, ua.xp_earned DESC, gs.time_taken_ms ASC
ON CONFLICT (quiz_id, user_id) DO NOTHING;
