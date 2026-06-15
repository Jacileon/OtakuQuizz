package handlers

import (
	"fmt"
	"html/template"

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
