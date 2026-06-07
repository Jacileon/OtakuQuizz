-- ============================================================
-- AJOUT DES NOUVEAUX TYPES DE QUESTIONS
-- ============================================================

-- Ajouter les nouvelles colonnes pour les types spéciaux de questions
ALTER TABLE questions 
ADD COLUMN IF NOT EXISTS character_guess_data jsonb,
ADD COLUMN IF NOT EXISTS character_guess_mode text CHECK (character_guess_mode IN ('image', 'text')),
ADD COLUMN IF NOT EXISTS find_odd_data jsonb;

-- Mettre à jour la contrainte CHECK pour inclure les nouveaux types
ALTER TABLE questions 
DROP CONSTRAINT IF EXISTS questions_question_type_check;

ALTER TABLE questions 
ADD CONSTRAINT questions_question_type_check 
CHECK (question_type IN ('text', 'true_false', 'image', 'image_shadow', 'gif', 'audio', 'character_guess', 'impostor'));

-- Commentaires pour documentation
COMMENT ON COLUMN questions.character_guess_data IS 'Données JSON pour les questions de type character_guess (liste de personnages avec indices)';
COMMENT ON COLUMN questions.character_guess_mode IS 'Mode d''affichage pour character_guess: image ou text';
COMMENT ON COLUMN questions.find_odd_data IS 'Données JSON pour les questions de type impostor (éléments + intrus)';