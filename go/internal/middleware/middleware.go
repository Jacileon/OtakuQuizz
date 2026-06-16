package middleware

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"

	"otaku-quiz-africa/internal/database"
	"otaku-quiz-africa/internal/handlers"
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
	accessToken := sess.Get("access_token")
	if userID == nil || accessToken == nil {
		return c.Redirect("/login")
	}

	// Récupérer le profil depuis Supabase
	supabaseURL := os.Getenv("SUPABASE_URL")
	if supabaseURL == "" {
		supabaseURL = os.Getenv("NEXT_PUBLIC_SUPABASE_URL")
	}

	url := fmt.Sprintf("%s/rest/v1/user_profiles?id=eq.%s&select=*", supabaseURL, userID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Authorization", "Bearer "+accessToken.(string))

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return c.Redirect("/login")
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var profiles []handlers.UserProfile
	json.Unmarshal(body, &profiles)

	if len(profiles) == 0 {
		return c.Redirect("/login")
	}

	c.Locals("user", &profiles[0])
	return c.Next()
}

func (m *Middleware) RequireAdmin(c *fiber.Ctx) error {
	user := c.Locals("user")
	if user == nil {
		return c.Redirect("/login")
	}

	profile, ok := user.(*handlers.UserProfile)
	if !ok || !profile.IsAdmin {
		return c.Redirect("/dashboard")
	}

	return c.Next()
}
