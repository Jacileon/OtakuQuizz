package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"

	"otaku-quiz-africa/internal/database"
)

type Middleware struct {
	db    *database.Supabase
	store *session.Store
}

func New(db *database.Supabase, store *session.Store) *Middleware {
	return &Middleware{db: db, store: store}
}

func (m *Middleware) RequireAuth(c *fiber.Ctx) error {
	sess, err := m.store.Get(c)
	if err != nil {
		return c.Redirect("/login")
	}

	userID := sess.Get("user_id")
	if userID == nil {
		return c.Redirect("/login")
	}

	// TODO: Fetch user from Supabase
	user := &struct {
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
	}{
		ID:       userID.(string),
		Username: "User",
		Rank:     "F",
	}

	c.Locals("user", user)
	return c.Next()
}

func (m *Middleware) RequireAdmin(c *fiber.Ctx) error {
	user := c.Locals("user")
	if user == nil {
		return c.Redirect("/login")
	}

	// TODO: Check is_admin from Supabase
	return c.Next()
}
