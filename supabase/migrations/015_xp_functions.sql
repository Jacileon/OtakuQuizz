-- Fonction pour incrémenter l'XP d'un utilisateur
CREATE OR REPLACE FUNCTION increment_user_xp(user_id uuid, amount numeric)
RETURNS void AS $$
BEGIN
    UPDATE user_profiles 
    SET xp = xp + amount::integer,
        total_xp = total_xp + amount::integer,
        level = GREATEST(1, FLOOR(SQRT((xp + amount::integer) / 10)) + 1)
    WHERE id = user_id;
    
    -- Mettre à jour le rang
    PERFORM update_user_rank(user_id);
END;
$$ LANGUAGE plpgsql;

-- Fonction pour mettre à jour le rang d'un utilisateur
CREATE OR REPLACE FUNCTION update_user_rank(target_user_id uuid)
RETURNS void AS $$
DECLARE
    user_xp integer;
    new_rank text;
BEGIN
    SELECT xp INTO user_xp FROM user_profiles WHERE id = target_user_id;
    
    SELECT rank_label INTO new_rank 
    FROM rank_config 
    WHERE xp_required <= user_xp 
    ORDER BY display_order DESC 
    LIMIT 1;
    
    IF new_rank IS NOT NULL THEN
        UPDATE user_profiles SET rank = new_rank WHERE id = target_user_id;
    END IF;
END;
$$ LANGUAGE plpgsql;