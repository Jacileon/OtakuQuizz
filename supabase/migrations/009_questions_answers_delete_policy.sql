-- Politiques DELETE pour questions et answers
DROP POLICY IF EXISTS "questions_delete_own" ON questions;
CREATE POLICY "questions_delete_own" ON questions FOR DELETE USING (
    EXISTS (SELECT 1 FROM quizzes WHERE quizzes.id = questions.quiz_id AND quizzes.creator_id = auth.uid())
);

DROP POLICY IF EXISTS "answers_delete_own" ON answers;
CREATE POLICY "answers_delete_own" ON answers FOR DELETE USING (
    EXISTS (SELECT 1 FROM questions q JOIN quizzes qu ON q.quiz_id = qu.id 
    WHERE q.id = answers.question_id AND qu.creator_id = auth.uid())
);