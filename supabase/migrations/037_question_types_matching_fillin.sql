-- ============================================================
-- NOUVEAUX TYPES DE QUESTIONS : matching + fill_in
-- + colonnes options/distractors + table quiz_session_questions
-- ============================================================

-- Étendre la contrainte CHECK pour inclure matching et fill_in
ALTER TABLE questions
DROP CONSTRAINT IF EXISTS questions_question_type_check;

ALTER TABLE questions
ADD CONSTRAINT questions_question_type_check
CHECK (question_type IN ('text', 'true_false', 'image', 'image_shadow', 'gif', 'audio', 'character_guess', 'impostor', 'matching', 'fill_in'));

-- Ajouter colonnes options et distractors (JSONB)
ALTER TABLE questions ADD COLUMN IF NOT EXISTS options jsonb;
ALTER TABLE questions ADD COLUMN IF NOT EXISTS distractors jsonb;

COMMENT ON COLUMN questions.options IS 'Structure JSON spécifique au type (pairs pour matching, blanks/template pour fill_in)';
COMMENT ON COLUMN questions.distractors IS 'Pool de distracteurs pour QCM (array de string)';

-- Table quiz_session_questions : stocke les options randomisées par session
CREATE TABLE IF NOT EXISTS quiz_session_questions (
  id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
  session_id text NOT NULL,
  question_id uuid NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
  presented_options jsonb,
  session_seed int NOT NULL,
  created_at timestamptz DEFAULT now(),
  UNIQUE(session_id, question_id)
);

-- Index pour recherche rapide par session
CREATE INDEX IF NOT EXISTS idx_quiz_session_questions_session ON quiz_session_questions(session_id);

-- RLS policies
ALTER TABLE quiz_session_questions ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Users can read own session questions"
  ON quiz_session_questions FOR SELECT
  USING (true);

CREATE POLICY "Service can insert session questions"
  ON quiz_session_questions FOR INSERT
  WITH CHECK (true);
