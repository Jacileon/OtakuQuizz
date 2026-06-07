-- ============================================================
-- SEED BADGES - OTAKU QUIZ AFRICA
-- ============================================================

INSERT INTO badges (slug, name, description, condition_type, condition_value, is_rare) VALUES
-- Badges de jeu
('first_quiz', 'Premier Pas', 'Complète ton premier quiz', 'quizzes_played', 1, false),
('quiz_10', 'Débutant', 'Complète 10 quiz', 'quizzes_played', 10, false),
('quiz_100', 'Vétéran', 'Complète 100 quiz', 'quizzes_played', 100, false),
('quiz_500', 'Légende Vivante', 'Complète 500 quiz', 'quizzes_played', 500, true),
('quiz_1000', 'Immortel', 'Complète 1000 quiz', 'quizzes_played', 1000, true),
('perfect_quiz', 'Perfection', 'Obtiens un score parfait (100%)', 'perfect_quiz', 1, false),
('perfect_5', 'Maître de la Perfection', 'Obtiens 5 scores parfaits', 'perfect_quiz', 5, true),
('accuracy_90', 'Précision Chirurgicale', '90% de précision sur 20+ quiz', 'accuracy_rate', 90, true),

-- Badges de création
('first_creation', 'Créateur', 'Crée ton premier quiz', 'quizzes_created', 1, false),
('creator_10', 'Producteur', 'Crée 10 quiz', 'quizzes_created', 10, false),
('popular_creator', 'Star Montante', 'Un quiz avec 100+ joueurs', 'popular_quiz', 100, true),
('elite_creator', 'Superstar', 'Un quiz avec 1000+ joueurs', 'popular_quiz', 1000, true),

-- Badges de classement
('top10_monthly', 'Top 10', 'Termine dans le top 10 mensuel', 'monthly_top10', 10, true),
('monthly_champion', 'Champion du Mois', 'Termine 1er du classement mensuel', 'monthly_champion', 1, true),

-- Badges spéciaux
('speed_demon', 'Démon de Vitesse', 'Réponds en moins de 2 secondes', 'speed_answer', 2, false),
('streak_master', 'Maître des Séries', 'Série de 10 bonnes réponses', 'streak', 10, true),
('night_owl', 'Chouette Nocturne', 'Joue à 3h du matin', 'night_play', 1, false),
('weekend_warrior', 'Guerrier du Week-end', 'Joue 10 quiz le week-end', 'weekend_play', 10, false);
