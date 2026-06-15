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
	return c.Render("pages/home", fiber.Map{
		"Title": "Otaku Quiz Africa",
	})
}

func (h *Handler) LoginPage(c *fiber.Ctx) error {
	return c.Render("pages/login", fiber.Map{
		"Title": "Connexion",
	})
}

func (h *Handler) Login(c *fiber.Ctx) error {
	// TODO: Implement Supabase auth
	return c.Redirect("/dashboard")
}

func (h *Handler) RegisterPage(c *fiber.Ctx) error {
	return c.Render("pages/register", fiber.Map{
		"Title": "Inscription",
	})
}

func (h *Handler) Register(c *fiber.Ctx) error {
	// TODO: Implement Supabase auth
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
	return c.Render("pages/dashboard", fiber.Map{
		"Title": "Dashboard",
		"User":  user,
	})
}

func (h *Handler) Explore(c *fiber.Ctx) error {
	// TODO: Fetch quizzes from Supabase
	return c.Render("pages/explore", fiber.Map{
		"Title":   "Explorer",
		"Quizzes": []interface{}{},
	})
}

func (h *Handler) QuizDetail(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.Render("pages/quiz-detail", fiber.Map{
		"Title": "Quiz",
		"ID":    id,
	})
}

func (h *Handler) QuizPlay(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.Render("pages/quiz-play", fiber.Map{
		"Title": "Jouer",
		"ID":    id,
	})
}

func (h *Handler) QuizSubmit(c *fiber.Ctx) error {
	// TODO: Process quiz submission via Supabase
	return c.JSON(fiber.Map{"success": true, "xpEarned": 0})
}

func (h *Handler) Friends(c *fiber.Ctx) error {
	user := c.Locals("user").(*UserProfile)
	return c.Render("pages/friends", fiber.Map{
		"Title": "Amis",
		"User":  user,
	})
}

func (h *Handler) ChallengeDetail(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.Render("pages/challenge", fiber.Map{
		"Title": "Défi",
		"ID":    id,
	})
}

func (h *Handler) Profile(c *fiber.Ctx) error {
	user := c.Locals("user").(*UserProfile)
	return c.Render("pages/profile", fiber.Map{
		"Title": "Profil",
		"User":  user,
	})
}

func (h *Handler) ProfileEdit(c *fiber.Ctx) error {
	user := c.Locals("user").(*UserProfile)
	return c.Render("pages/profile-edit", fiber.Map{
		"Title": "Modifier le profil",
		"User":  user,
	})
}

func (h *Handler) ProfileUpdate(c *fiber.Ctx) error {
	// TODO: Update profile via Supabase
	return c.Redirect("/profil")
}

func (h *Handler) Leaderboard(c *fiber.Ctx) error {
	return c.Render("pages/leaderboard", fiber.Map{
		"Title":      "Classement",
		"Leaderboard": []interface{}{},
	})
}

func (h *Handler) FAQ(c *fiber.Ctx) error {
	return c.Render("pages/faq", fiber.Map{
		"Title": "FAQ",
	})
}

// Pages admin
func (h *Handler) AdminDashboard(c *fiber.Ctx) error {
	return c.Render("pages/admin/dashboard", fiber.Map{
		"Title": "Admin",
	})
}

func (h *Handler) AdminOfficialQuizzes(c *fiber.Ctx) error {
	return c.Render("pages/admin/official-quizzes", fiber.Map{
		"Title": "Quiz Officiels",
	})
}

func (h *Handler) AdminTickets(c *fiber.Ctx) error {
	return c.Render("pages/admin/tickets", fiber.Map{
		"Title": "Tickets",
	})
}

func (h *Handler) AdminAnnouncements(c *fiber.Ctx) error {
	return c.Render("pages/admin/announcements", fiber.Map{
		"Title": "Annonces",
	})
}

func (h *Handler) AdminSettings(c *fiber.Ctx) error {
	return c.Render("pages/admin/settings", fiber.Map{
		"Title": "Paramètres",
	})
}
