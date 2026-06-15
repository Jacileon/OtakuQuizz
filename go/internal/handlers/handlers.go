package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"

	"otaku-quiz-africa/internal/database"
)

type Handler struct {
	db    *database.Supabase
	store *session.Store
}

func New(db *database.Supabase, store *session.Store) *Handler {
	return &Handler{db: db, store: store}
}

type UserProfile struct {
	ID            string  `json:"id"`
	Email         string  `json:"email"`
	Username      string  `json:"username"`
	Nickname      *string `json:"nickname"`
	AvatarURL     *string `json:"avatar_url"`
	XP            int     `json:"xp"`
	Level         int     `json:"level"`
	Rank          string  `json:"rank"`
	IsAdmin       bool    `json:"is_admin"`
	CanCreateQuiz bool    `json:"can_create_quiz"`
}

// Pages publiques
func (h *Handler) Home(c *fiber.Ctx) error {
	return c.SendString(`
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Otaku Quiz Africa</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #0f0f23; 
            color: #e2e8f0; 
            min-height: 100vh; 
            display: flex;
            flex-direction: column;
        }
        a { color: inherit; text-decoration: none; }
        nav { 
            background: rgba(15, 15, 35, 0.95); 
            border-bottom: 1px solid #2d2d44; 
            padding: 16px;
        }
        nav .inner { 
            max-width: 1200px;
            margin: 0 auto;
            display: flex; 
            align-items: center; 
            justify-content: space-between; 
        }
        nav .logo { 
            display: flex; 
            align-items: center; 
            gap: 8px; 
            font-weight: bold; 
            font-size: 1.25rem; 
        }
        nav .links { 
            display: flex; 
            gap: 24px; 
        }
        nav .links a { 
            color: #94a3b8; 
            transition: color 0.2s; 
        }
        nav .links a:hover { 
            color: white; 
        }
        main { 
            flex: 1;
            max-width: 1200px;
            margin: 0 auto;
            padding: 32px 16px;
            width: 100%;
        }
        .hero {
            text-align: center;
            padding: 80px 20px;
        }
        .hero h1 {
            font-size: 3rem;
            font-weight: bold;
            margin-bottom: 16px;
            background: linear-gradient(135deg, #6366f1, #a855f7);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
        }
        .hero p {
            font-size: 1.25rem;
            color: #94a3b8;
            margin-bottom: 32px;
        }
        .buttons {
            display: flex;
            gap: 16px;
            justify-content: center;
            flex-wrap: wrap;
        }
        .btn-primary {
            background: #6366f1;
            color: white;
            padding: 12px 32px;
            border-radius: 8px;
            font-weight: 500;
            transition: background 0.2s;
        }
        .btn-primary:hover {
            background: #4f46e5;
        }
        .btn-secondary {
            border: 2px solid #6366f1;
            color: #6366f1;
            padding: 12px 32px;
            border-radius: 8px;
            font-weight: 500;
            transition: background 0.2s;
        }
        .btn-secondary:hover {
            background: rgba(99, 102, 241, 0.1);
        }
        footer {
            background: #1a1a2e;
            border-top: 1px solid #2d2d44;
            padding: 24px;
            text-align: center;
            color: #94a3b8;
            font-size: 0.875rem;
        }
    </style>
</head>
<body>
    <nav>
        <div class="inner">
            <a href="/" class="logo">
                <span style="font-size: 1.5rem;">⚔️</span>
                <span>OTAKU QUIZ AFRICA</span>
            </a>
            <div class="links">
                <a href="/dashboard">Accueil</a>
                <a href="/explore">Explorer</a>
                <a href="/friends">Amis</a>
                <a href="/leaderboard">Classement</a>
            </div>
            <div>
                <a href="/login" class="btn-primary">Connexion</a>
            </div>
        </div>
    </nav>
    
    <main>
        <div class="hero">
            <h1>⚔️ OTAKU QUIZ AFRICA</h1>
            <p>Teste tes connaissances anime et manga</p>
            <div class="buttons">
                <a href="/explore" class="btn-primary">Explorer les quiz</a>
                <a href="/register" class="btn-secondary">S'inscrire</a>
            </div>
        </div>
    </main>
    
    <footer>
        <p>&copy; 2024 Otaku Quiz Africa. Tous droits réservés.</p>
    </footer>
</body>
</html>
	`)
}

func (h *Handler) LoginPage(c *fiber.Ctx) error {
	return c.SendString(`
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Connexion | Otaku Quiz Africa</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #0f0f23; 
            color: #e2e8f0; 
            min-height: 100vh; 
            display: flex;
            align-items: center;
            justify-content: center;
        }
        .card {
            background: #16213e;
            border: 1px solid #2d2d44;
            border-radius: 12px;
            padding: 32px;
            width: 100%;
            max-width: 400px;
        }
        h1 {
            font-size: 1.5rem;
            font-weight: bold;
            margin-bottom: 24px;
            text-align: center;
        }
        label {
            display: block;
            font-size: 0.875rem;
            font-weight: 500;
            margin-bottom: 8px;
        }
        input {
            width: 100%;
            padding: 12px;
            background: #1a1a2e;
            border: 1px solid #2d2d44;
            border-radius: 8px;
            color: white;
            font-size: 1rem;
            margin-bottom: 16px;
        }
        input:focus {
            outline: none;
            border-color: #6366f1;
        }
        button {
            width: 100%;
            padding: 12px;
            background: #6366f1;
            color: white;
            border: none;
            border-radius: 8px;
            font-size: 1rem;
            font-weight: 500;
            cursor: pointer;
            transition: background 0.2s;
        }
        button:hover {
            background: #4f46e5;
        }
        .link {
            text-align: center;
            margin-top: 16px;
            color: #94a3b8;
        }
        .link a {
            color: #6366f1;
        }
    </style>
</head>
<body>
    <div class="card">
        <h1>Connexion</h1>
        <form method="POST" action="/login">
            <label>Email</label>
            <input type="email" name="email" required placeholder="votre@email.com">
            
            <label>Mot de passe</label>
            <input type="password" name="password" required placeholder="••••••••">
            
            <button type="submit">Se connecter</button>
        </form>
        <p class="link">
            Pas de compte ? <a href="/register">S'inscrire</a>
        </p>
    </div>
</body>
</html>
	`)
}

func (h *Handler) Login(c *fiber.Ctx) error {
	return c.Redirect("/dashboard")
}

func (h *Handler) RegisterPage(c *fiber.Ctx) error {
	return c.SendString(`
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Inscription | Otaku Quiz Africa</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #0f0f23; 
            color: #e2e8f0; 
            min-height: 100vh; 
            display: flex;
            align-items: center;
            justify-content: center;
        }
        .card {
            background: #16213e;
            border: 1px solid #2d2d44;
            border-radius: 12px;
            padding: 32px;
            width: 100%;
            max-width: 400px;
        }
        h1 {
            font-size: 1.5rem;
            font-weight: bold;
            margin-bottom: 24px;
            text-align: center;
        }
        label {
            display: block;
            font-size: 0.875rem;
            font-weight: 500;
            margin-bottom: 8px;
        }
        input {
            width: 100%;
            padding: 12px;
            background: #1a1a2e;
            border: 1px solid #2d2d44;
            border-radius: 8px;
            color: white;
            font-size: 1rem;
            margin-bottom: 16px;
        }
        input:focus {
            outline: none;
            border-color: #6366f1;
        }
        button {
            width: 100%;
            padding: 12px;
            background: #6366f1;
            color: white;
            border: none;
            border-radius: 8px;
            font-size: 1rem;
            font-weight: 500;
            cursor: pointer;
            transition: background 0.2s;
        }
        button:hover {
            background: #4f46e5;
        }
        .link {
            text-align: center;
            margin-top: 16px;
            color: #94a3b8;
        }
        .link a {
            color: #6366f1;
        }
    </style>
</head>
<body>
    <div class="card">
        <h1>Inscription</h1>
        <form method="POST" action="/register">
            <label>Email</label>
            <input type="email" name="email" required placeholder="votre@email.com">
            
            <label>Mot de passe</label>
            <input type="password" name="password" required placeholder="••••••••">
            
            <label>Confirmer le mot de passe</label>
            <input type="password" name="password_confirm" required placeholder="••••••••">
            
            <button type="submit">S'inscrire</button>
        </form>
        <p class="link">
            Déjà un compte ? <a href="/login">Se connecter</a>
        </p>
    </div>
</body>
</html>
	`)
}

func (h *Handler) Register(c *fiber.Ctx) error {
	return c.Redirect("/complete-profile")
}

func (h *Handler) Logout(c *fiber.Ctx) error {
	sess, err := h.store.Get(c)
	if err == nil {
		sess.Destroy()
	}
	return c.Redirect("/login")
}

// Pages protégées
func (h *Handler) Dashboard(c *fiber.Ctx) error {
	user := c.Locals("user").(*UserProfile)
	return c.SendString(`
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Dashboard | Otaku Quiz Africa</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #0f0f23; color: #e2e8f0; min-height: 100vh; }
        a { color: inherit; text-decoration: none; }
        nav { background: rgba(15, 15, 35, 0.95); border-bottom: 1px solid #2d2d44; padding: 16px; }
        nav .inner { max-width: 1200px; margin: 0 auto; display: flex; align-items: center; justify-content: space-between; }
        nav .logo { display: flex; align-items: center; gap: 8px; font-weight: bold; font-size: 1.25rem; }
        main { max-width: 1200px; margin: 0 auto; padding: 32px 16px; }
        .welcome { font-size: 1.5rem; font-weight: bold; margin-bottom: 8px; }
        .rank { color: #94a3b8; margin-bottom: 24px; }
        .stats { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 16px; margin-bottom: 32px; }
        .stat-card { background: #16213e; border: 1px solid #2d2d44; border-radius: 12px; padding: 20px; }
        .stat-card .label { font-size: 0.875rem; color: #94a3b8; margin-bottom: 4px; }
        .stat-card .value { font-size: 1.5rem; font-weight: bold; }
    </style>
</head>
<body>
    <nav>
        <div class="inner">
            <a href="/" class="logo">⚔️ OTAKU QUIZ AFRICA</a>
            <div style="display: flex; gap: 16px;">
                <a href="/dashboard">Accueil</a>
                <a href="/explore">Explorer</a>
                <a href="/friends">Amis</a>
                <a href="/leaderboard">Classement</a>
                <a href="/profil">Profil</a>
            </div>
        </div>
    </nav>
    <main>
        <div class="welcome">Bonjour ` + user.Username + ` 👋</div>
        <div class="rank">` + user.Rank + ` • Niveau ` + string(rune(user.Level+'0')) + `</div>
        <div class="stats">
            <div class="stat-card">
                <div class="label">XP</div>
                <div class="value">` + string(rune(user.XP+'0')) + `</div>
            </div>
            <div class="stat-card">
                <div class="label">Quiz joués</div>
                <div class="value">0</div>
            </div>
            <div class="stat-card">
                <div class="label">Meilleur score</div>
                <div class="value">0</div>
            </div>
        </div>
    </main>
</body>
</html>
	`)
}

func (h *Handler) Explore(c *fiber.Ctx) error {
	return c.SendString(`<!DOCTYPE html><html><head><title>Explorer</title></head><body style="background:#0f0f23;color:white;font-family:sans-serif;padding:32px"><h1>Explorer les quiz</h1><p>Page en construction</p><a href="/dashboard" style="color:#6366f1">Retour</a></body></html>`)
}

func (h *Handler) QuizDetail(c *fiber.Ctx) error {
	return c.SendString(`<!DOCTYPE html><html><head><title>Quiz</title></head><body style="background:#0f0f23;color:white;font-family:sans-serif;padding:32px"><h1>Détail du quiz</h1><p>Page en construction</p><a href="/explore" style="color:#6366f1">Retour</a></body></html>`)
}

func (h *Handler) QuizPlay(c *fiber.Ctx) error {
	return c.SendString(`<!DOCTYPE html><html><head><title>Jouer</title></head><body style="background:#0f0f23;color:white;font-family:sans-serif;padding:32px"><h1>Jouer au quiz</h1><p>Page en construction</p><a href="/explore" style="color:#6366f1">Retour</a></body></html>`)
}

func (h *Handler) QuizSubmit(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"success": true, "xpEarned": 0})
}

func (h *Handler) Friends(c *fiber.Ctx) error {
	return c.SendString(`<!DOCTYPE html><html><head><title>Amis</title></head><body style="background:#0f0f23;color:white;font-family:sans-serif;padding:32px"><h1>Amis</h1><p>Page en construction</p><a href="/dashboard" style="color:#6366f1">Retour</a></body></html>`)
}

func (h *Handler) ChallengeDetail(c *fiber.Ctx) error {
	return c.SendString(`<!DOCTYPE html><html><head><title>Défi</title></head><body style="background:#0f0f23;color:white;font-family:sans-serif;padding:32px"><h1>Défi</h1><p>Page en construction</p><a href="/friends" style="color:#6366f1">Retour</a></body></html>`)
}

func (h *Handler) Profile(c *fiber.Ctx) error {
	user := c.Locals("user").(*UserProfile)
	return c.SendString(`
<!DOCTYPE html>
<html><head><title>Profil</title></head>
<body style="background:#0f0f23;color:white;font-family:sans-serif;padding:32px">
<h1>Profil de ` + user.Username + `</h1>
<p>Rang: ` + user.Rank + `</p>
<p>XP: ` + string(rune(user.XP+'0')) + `</p>
<a href="/profile/edit" style="color:#6366f1">Modifier</a> | <a href="/dashboard" style="color:#6366f1">Retour</a>
</body></html>
	`)
}

func (h *Handler) ProfileEdit(c *fiber.Ctx) error {
	return c.SendString(`<!DOCTYPE html><html><head><title>Modifier profil</title></head><body style="background:#0f0f23;color:white;font-family:sans-serif;padding:32px"><h1>Modifier le profil</h1><p>Page en construction</p><a href="/profil" style="color:#6366f1">Retour</a></body></html>`)
}

func (h *Handler) ProfileUpdate(c *fiber.Ctx) error {
	return c.Redirect("/profil")
}

func (h *Handler) Leaderboard(c *fiber.Ctx) error {
	return c.SendString(`<!DOCTYPE html><html><head><title>Classement</title></head><body style="background:#0f0f23;color:white;font-family:sans-serif;padding:32px"><h1>Classement</h1><p>Page en construction</p><a href="/dashboard" style="color:#6366f1">Retour</a></body></html>`)
}

func (h *Handler) FAQ(c *fiber.Ctx) error {
	return c.SendString(`<!DOCTYPE html><html><head><title>FAQ</title></head><body style="background:#0f0f23;color:white;font-family:sans-serif;padding:32px"><h1>FAQ</h1><p>Page en construction</p><a href="/dashboard" style="color:#6366f1">Retour</a></body></html>`)
}

// Pages admin
func (h *Handler) AdminDashboard(c *fiber.Ctx) error {
	return c.SendString(`<!DOCTYPE html><html><head><title>Admin</title></head><body style="background:#0f0f23;color:white;font-family:sans-serif;padding:32px"><h1>Administration</h1><p>Page en construction</p><a href="/dashboard" style="color:#6366f1">Retour</a></body></html>`)
}

func (h *Handler) AdminOfficialQuizzes(c *fiber.Ctx) error {
	return c.SendString(`<!DOCTYPE html><html><head><title>Quiz Officiels</title></head><body style="background:#0f0f23;color:white;font-family:sans-serif;padding:32px"><h1>Quiz Officiels</h1><p>Page en construction</p><a href="/admin" style="color:#6366f1">Retour</a></body></html>`)
}

func (h *Handler) AdminTickets(c *fiber.Ctx) error {
	return c.SendString(`<!DOCTYPE html><html><head><title>Tickets</title></head><body style="background:#0f0f23;color:white;font-family:sans-serif;padding:32px"><h1>Tickets</h1><p>Page en construction</p><a href="/admin" style="color:#6366f1">Retour</a></body></html>`)
}

func (h *Handler) AdminAnnouncements(c *fiber.Ctx) error {
	return c.SendString(`<!DOCTYPE html><html><head><title>Annonces</title></head><body style="background:#0f0f23;color:white;font-family:sans-serif;padding:32px"><h1>Annonces</h1><p>Page en construction</p><a href="/admin" style="color:#6366f1">Retour</a></body></html>`)
}

func (h *Handler) AdminSettings(c *fiber.Ctx) error {
	return c.SendString(`<!DOCTYPE html><html><head><title>Paramètres</title></head><body style="background:#0f0f23;color:white;font-family:sans-serif;padding:32px"><h1>Paramètres</h1><p>Page en construction</p><a href="/admin" style="color:#6366f1">Retour</a></body></html>`)
}
