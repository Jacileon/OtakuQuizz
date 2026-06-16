package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"

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

func renderPage(c *fiber.Ctx, title string, content string) error {
	return c.Render("layouts/main", fiber.Map{
		"Title":   title,
		"Content": template.HTML(content),
	})
}

// Pages publiques
func (h *Handler) Home(c *fiber.Ctx) error {
	return renderPage(c, "Accueil", `
<div class="hero">
    <h1>⚔️ OTAKU QUIZ AFRICA</h1>
    <p>Teste tes connaissances anime et manga</p>
    <div class="buttons">
        <a href="/explore" class="btn-primary">Explorer les quiz</a>
        <a href="/register" class="btn-secondary">S'inscrire</a>
    </div>
</div>
	`)
}

func (h *Handler) LoginPage(c *fiber.Ctx) error {
	return c.Render("pages/login", fiber.Map{"Title": "Connexion"})
}

func (h *Handler) Login(c *fiber.Ctx) error {
	return c.Redirect("/dashboard")
}

func (h *Handler) RegisterPage(c *fiber.Ctx) error {
	return c.Render("pages/register", fiber.Map{"Title": "Inscription"})
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

func (h *Handler) GoogleAuth(c *fiber.Ctx) error {
	supabaseURL := os.Getenv("SUPABASE_URL")
	if supabaseURL == "" {
		supabaseURL = os.Getenv("NEXT_PUBLIC_SUPABASE_URL")
	}
	
	redirectURL := fmt.Sprintf("%s/auth/v1/authorize?provider=google&redirect_to=http://localhost:8080/auth/callback", supabaseURL)
	return c.Redirect(redirectURL)
}

func (h *Handler) GoogleCallback(c *fiber.Ctx) error {
	// Supabase redirige avec des tokens dans l'URL fragment
	// On utilise JavaScript pour récupérer les tokens côté client
	return c.SendString(`
<!DOCTYPE html>
<html>
<head>
    <title>Connexion...</title>
    <script>
        // Récupérer les tokens depuis l'URL fragment
        const hash = window.location.hash.substring(1);
        const params = new URLSearchParams(hash);
        const accessToken = params.get('access_token');
        const refreshToken = params.get('refresh_token');
        
        if (accessToken) {
            // Envoyer les tokens au serveur
            fetch('/auth/session', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ access_token: accessToken, refresh_token: refreshToken })
            }).then(() => {
                window.location.href = '/dashboard';
            });
        } else {
            window.location.href = '/login?error=no_token';
        }
    </script>
</head>
<body>
    <p>Connexion en cours...</p>
</body>
</html>
	`)
}

func (h *Handler) CreateSession(c *fiber.Ctx) error {
	type SessionRequest struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}

	var req SessionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Récupérer l'utilisateur depuis Supabase avec le token
	supabaseURL := os.Getenv("SUPABASE_URL")
	if supabaseURL == "" {
		supabaseURL = os.Getenv("NEXT_PUBLIC_SUPABASE_URL")
	}

	// Appeler Supabase pour obtenir l'utilisateur
	url := fmt.Sprintf("%s/auth/v1/user", supabaseURL)
	req2, _ := http.NewRequest("GET", url, nil)
	req2.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req2.Header.Set("Authorization", "Bearer "+req.AccessToken)

	resp, err := http.DefaultClient.Do(req2)
	if err != nil || resp.StatusCode != 200 {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	defer resp.Body.Close()

	// Lire la réponse
	body, _ := io.ReadAll(resp.Body)
	
	// Parser l'utilisateur
	type SupabaseUser struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	}
	var user SupabaseUser
	json.Unmarshal(body, &user)

	// Créer la session
	sess, err := h.store.Get(c)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Session error"})
	}

	sess.Set("user_id", user.ID)
	sess.Set("access_token", req.AccessToken)
	sess.Set("refresh_token", req.RefreshToken)
	sess.Save()

	return c.JSON(fiber.Map{"success": true})
}

// Pages protégées
func (h *Handler) Dashboard(c *fiber.Ctx) error {
	user := c.Locals("user").(*UserProfile)
	return renderPage(c, "Dashboard", fmt.Sprintf(`
<div class="welcome">Bonjour %s 👋</div>
<div class="rank">%s • Niveau %d</div>
<div class="stats">
    <div class="stat-card">
        <div class="label">XP</div>
        <div class="value">%d</div>
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
	`, user.Username, user.Rank, user.Level, user.XP))
}

func (h *Handler) Explore(c *fiber.Ctx) error {
	return renderPage(c, "Explorer", `<h1>Explorer les quiz</h1><p>Page en construction</p>`)
}

func (h *Handler) QuizDetail(c *fiber.Ctx) error {
	return renderPage(c, "Quiz", `<h1>Détail du quiz</h1><p>Page en construction</p>`)
}

func (h *Handler) QuizPlay(c *fiber.Ctx) error {
	return renderPage(c, "Jouer", `<h1>Jouer au quiz</h1><p>Page en construction</p>`)
}

func (h *Handler) QuizSubmit(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"success": true, "xpEarned": 0})
}

func (h *Handler) Friends(c *fiber.Ctx) error {
	return renderPage(c, "Amis", `<h1>Amis</h1><p>Page en construction</p>`)
}

func (h *Handler) ChallengeDetail(c *fiber.Ctx) error {
	return renderPage(c, "Défi", `<h1>Défi</h1><p>Page en construction</p>`)
}

func (h *Handler) Profile(c *fiber.Ctx) error {
	user := c.Locals("user").(*UserProfile)
	return renderPage(c, "Profil", fmt.Sprintf(`
<h1>Profil de %s</h1>
<p>Rang: %s | XP: %d</p>
<a href="/profile/edit" class="btn-primary" style="display:inline-block;margin-top:16px;">Modifier</a>
	`, user.Username, user.Rank, user.XP))
}

func (h *Handler) ProfileEdit(c *fiber.Ctx) error {
	return renderPage(c, "Modifier le profil", `<h1>Modifier le profil</h1><p>Page en construction</p>`)
}

func (h *Handler) ProfileUpdate(c *fiber.Ctx) error {
	return c.Redirect("/profil")
}

func (h *Handler) Leaderboard(c *fiber.Ctx) error {
	return renderPage(c, "Classement", `<h1>Classement</h1><p>Page en construction</p>`)
}

func (h *Handler) FAQ(c *fiber.Ctx) error {
	return renderPage(c, "FAQ", `<h1>FAQ</h1><p>Page en construction</p>`)
}

// Pages admin
func (h *Handler) AdminDashboard(c *fiber.Ctx) error {
	return renderPage(c, "Admin", `<h1>Administration</h1><p>Page en construction</p>`)
}

func (h *Handler) AdminOfficialQuizzes(c *fiber.Ctx) error {
	return renderPage(c, "Quiz Officiels", `<h1>Quiz Officiels</h1><p>Page en construction</p>`)
}

func (h *Handler) AdminTickets(c *fiber.Ctx) error {
	return renderPage(c, "Tickets", `<h1>Tickets</h1><p>Page en construction</p>`)
}

func (h *Handler) AdminAnnouncements(c *fiber.Ctx) error {
	return renderPage(c, "Annonces", `<h1>Annonces</h1><p>Page en construction</p>`)
}

func (h *Handler) AdminSettings(c *fiber.Ctx) error {
	return renderPage(c, "Paramètres", `<h1>Paramètres</h1><p>Page en construction</p>`)
}
