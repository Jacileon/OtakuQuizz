-- Créer la ligne user_stats si elle n'existe pas
INSERT INTO user_stats (user_id)
SELECT id FROM user_profiles
WHERE id NOT IN (SELECT user_id FROM user_stats)
ON CONFLICT DO NOTHING;

-- Trigger pour créer user_stats automatiquement
CREATE OR REPLACE FUNCTION create_user_stats()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO user_stats (user_id) VALUES (NEW.id)
    ON CONFLICT DO NOTHING;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_create_user_stats ON user_profiles;
CREATE TRIGGER trigger_create_user_stats
    AFTER INSERT ON user_profiles
    FOR EACH ROW EXECUTE FUNCTION create_user_stats();