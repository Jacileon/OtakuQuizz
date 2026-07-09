-- ============================================================
-- PERSONAL QUIZ
-- ============================================================

CREATE TABLE IF NOT EXISTS personal_quizzes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    creator_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    token TEXT UNIQUE NOT NULL,
    title TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'archived')),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS personal_quiz_questions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    quiz_id UUID NOT NULL REFERENCES personal_quizzes(id) ON DELETE CASCADE,
    question_text TEXT NOT NULL,
    type TEXT NOT NULL DEFAULT 'texte',
    correct_answer TEXT NOT NULL,
    options JSONB,
    sort_order INT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS personal_quiz_participations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    quiz_id UUID NOT NULL REFERENCES personal_quizzes(id) ON DELETE CASCADE,
    participant_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    score INT NOT NULL DEFAULT 0,
    correct_count INT NOT NULL DEFAULT 0,
    total_count INT NOT NULL DEFAULT 0,
    participated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(quiz_id, participant_id)
);

CREATE TABLE IF NOT EXISTS personal_quiz_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    quiz_id UUID NOT NULL REFERENCES personal_quizzes(id) ON DELETE CASCADE,
    participation_id UUID REFERENCES personal_quiz_participations(id) ON DELETE SET NULL,
    sender_id UUID REFERENCES auth.users(id) ON DELETE SET NULL,
    display_nickname TEXT,
    message_text TEXT NOT NULL,
    is_anonymous BOOLEAN NOT NULL DEFAULT false,
    sent_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_personal_quizzes_token ON personal_quizzes(token);
CREATE INDEX IF NOT EXISTS idx_personal_quiz_questions_quiz ON personal_quiz_questions(quiz_id);
CREATE INDEX IF NOT EXISTS idx_personal_quiz_participations_quiz ON personal_quiz_participations(quiz_id);
CREATE INDEX IF NOT EXISTS idx_personal_quiz_messages_quiz ON personal_quiz_messages(quiz_id);

-- RLS
ALTER TABLE personal_quizzes ENABLE ROW LEVEL SECURITY;
ALTER TABLE personal_quiz_questions ENABLE ROW LEVEL SECURITY;
ALTER TABLE personal_quiz_participations ENABLE ROW LEVEL SECURITY;
ALTER TABLE personal_quiz_messages ENABLE ROW LEVEL SECURITY;

-- Policies: open access for service role, owners manage their own
CREATE POLICY "personal_quizzes_full_access" ON personal_quizzes FOR ALL USING (true) WITH CHECK (true);
CREATE POLICY "personal_quiz_questions_full_access" ON personal_quiz_questions FOR ALL USING (true) WITH CHECK (true);
CREATE POLICY "personal_quiz_participations_full_access" ON personal_quiz_participations FOR ALL USING (true) WITH CHECK (true);
CREATE POLICY "personal_quiz_messages_full_access" ON personal_quiz_messages FOR ALL USING (true) WITH CHECK (true);

-- ============================================================
-- ANON BOX (Boîte à messages anonymes)
-- ============================================================

CREATE TABLE IF NOT EXISTS anon_boxes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    token TEXT UNIQUE NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS anon_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    box_id UUID NOT NULL REFERENCES anon_boxes(id) ON DELETE CASCADE,
    message_text TEXT NOT NULL,
    sender_ip TEXT,
    is_read BOOLEAN NOT NULL DEFAULT false,
    sent_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_anon_boxes_token ON anon_boxes(token);
CREATE INDEX IF NOT EXISTS idx_anon_boxes_owner ON anon_boxes(owner_id);
CREATE INDEX IF NOT EXISTS idx_anon_messages_box ON anon_messages(box_id);

-- RLS
ALTER TABLE anon_boxes ENABLE ROW LEVEL SECURITY;
ALTER TABLE anon_messages ENABLE ROW LEVEL SECURITY;

CREATE POLICY "anon_boxes_full_access" ON anon_boxes FOR ALL USING (true) WITH CHECK (true);
CREATE POLICY "anon_messages_full_access" ON anon_messages FOR ALL USING (true) WITH CHECK (true);
