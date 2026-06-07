-- ============================================================
-- AJOUT DE LA COLONNE PHONE AUX USER_PROFILES
-- ============================================================

-- Ajouter la colonne phone
ALTER TABLE user_profiles 
ADD COLUMN IF NOT EXISTS phone text;

-- Mettre à jour le type de la vue UserProfile si nécessaire
COMMENT ON COLUMN user_profiles.phone IS 'Numéro de téléphone de l\'utilisateur avec indicatif pays';