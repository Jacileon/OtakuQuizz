-- Ajouter le champ nickname à user_profiles
ALTER TABLE user_profiles ADD COLUMN IF NOT EXISTS nickname text;

-- Ajouter une contrainte de longueur
ALTER TABLE user_profiles ADD CONSTRAINT nickname_length CHECK (length(nickname) >= 2 AND length(nickname) <= 30);