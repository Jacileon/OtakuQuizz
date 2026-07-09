ALTER TABLE questions ADD COLUMN IF NOT EXISTS xp_reward INTEGER NOT NULL DEFAULT 10;

UPDATE questions SET xp_reward = 10 WHERE question_type IN ('text', 'true_false', 'image', 'image_shadow', 'gif', 'audio');
UPDATE questions SET xp_reward = 20 WHERE question_type = 'fill_in';
UPDATE questions SET xp_reward = 30 WHERE question_type = 'matching';
UPDATE questions SET xp_reward = 50 WHERE question_type IN ('character_guess', 'impostor');
