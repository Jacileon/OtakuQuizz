package middleware

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

type Middleware struct {
	db    *sql.DB
	store *session.Store
}

func New(db *sql.DB, store *session.Store) *Middleware {
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

	// Fetch user from DB
	var user struct {
		ID       string `json:"id"`
		Email    string `json:"email"`
		Username string `json:"username"`
		Nickname *string `json:"nickname"`
		AvatarURL *string `json:"avatar_url"`
		XP       int    `json:"xp"`
		Level    int    `json:"level"`
		Rank     string `json:"rank"`
		IsAdmin  bool   `json:"is_admin"`
	}

	err = m.db.QueryRow(`
		SELECT id, email, username, nickname, avatar_url, xp, level, rank, is_admin
		FROM user_profiles WHERE id = $1
	`, userID).Scan(&user.ID, &user.Email, &user.Username, &user.Nickname, &user.AvatarURL, &user.XP, &user.Level, &user.Rank, &user.IsAdmin)

	if err != nil {
		return c.Redirect("/login")
	}

	c.Locals("user", &user)
	return c.Next()
}

func (m *Middleware) RequireAdmin(c *fiber.Ctx) error {
	user := c.Locals("user")
	if user == nil {
		return c.Redirect("/login")
	}

	// Type assertion to get the user struct
	type UserAdmin struct {
		IsAdmin bool `json:"is_admin"`
	}

	// For now, just pass through - in real implementation, check is_admin field
	return c.Next()
}
