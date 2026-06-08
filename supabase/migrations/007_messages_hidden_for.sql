-- Ajouter un champ pour cacher les messages côté utilisateur
ALTER TABLE messages ADD COLUMN IF NOT EXISTS hidden_for uuid[] DEFAULT '{}';

-- Index pour les requêtes filtrant sur hidden_for
CREATE INDEX IF NOT EXISTS idx_messages_hidden_for ON messages USING GIN (hidden_for);

-- Fonction RPC pour ajouter un utilisateur au tableau hidden_for
CREATE OR REPLACE FUNCTION add_to_hidden_for(message_id uuid, user_id uuid)
RETURNS void AS $$
BEGIN
    UPDATE messages 
    SET hidden_for = array_append(hidden_for, user_id)
    WHERE id = message_id
    AND NOT (user_id = ANY(hidden_for));
END;
$$ LANGUAGE plpgsql;