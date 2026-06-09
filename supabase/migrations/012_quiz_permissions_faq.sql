-- ============================================================
-- MODULE 5: Modèle de données
-- ============================================================

-- Ajouter can_create_quiz à user_profiles
ALTER TABLE user_profiles ADD COLUMN IF NOT EXISTS can_create_quiz boolean DEFAULT false;

-- Table AppConfig pour les paramètres globaux
CREATE TABLE IF NOT EXISTS app_config (
    key text PRIMARY KEY,
    value jsonb NOT NULL,
    updated_at timestamptz DEFAULT now(),
    updated_by uuid REFERENCES user_profiles(id)
);

-- Insérer la config par défaut pour les rangs autorisés
INSERT INTO app_config (key, value) 
VALUES ('quiz_creation_allowed_ranks', '["C", "E", "S"]'::jsonb)
ON CONFLICT (key) DO NOTHING;

-- Table FAQ
CREATE TABLE IF NOT EXISTS faq_entries (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    theme text NOT NULL,
    question text NOT NULL,
    answer text NOT NULL,
    order_index integer DEFAULT 0,
    visible boolean DEFAULT true,
    created_at timestamptz DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_faq_theme ON faq_entries(theme, order_index);

-- RLS pour app_config
ALTER TABLE app_config ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS "app_config_select_public" ON app_config;
CREATE POLICY "app_config_select_public" ON app_config FOR SELECT USING (true);

DROP POLICY IF EXISTS "app_config_update_admin" ON app_config;
CREATE POLICY "app_config_update_admin" ON app_config FOR UPDATE USING (
    EXISTS (SELECT 1 FROM user_profiles WHERE id = auth.uid() AND is_admin = true)
);

-- RLS pour faq_entries
ALTER TABLE faq_entries ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS "faq_select_public" ON faq_entries;
CREATE POLICY "faq_select_public" ON faq_entries FOR SELECT USING (visible = true);

DROP POLICY IF EXISTS "faq_insert_admin" ON faq_entries;
CREATE POLICY "faq_insert_admin" ON faq_entries FOR INSERT WITH CHECK (
    EXISTS (SELECT 1 FROM user_profiles WHERE id = auth.uid() AND is_admin = true)
);

DROP POLICY IF EXISTS "faq_update_admin" ON faq_entries;
CREATE POLICY "faq_update_admin" ON faq_entries FOR UPDATE USING (
    EXISTS (SELECT 1 FROM user_profiles WHERE id = auth.uid() AND is_admin = true)
);

DROP POLICY IF EXISTS "faq_delete_admin" ON faq_entries;
CREATE POLICY "faq_delete_admin" ON faq_entries FOR DELETE USING (
    EXISTS (SELECT 1 FROM user_profiles WHERE id = auth.uid() AND is_admin = true)
);