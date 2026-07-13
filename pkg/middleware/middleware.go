package middleware

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"

	"otaku-quiz-africa/pkg/database"
	"otaku-quiz-africa/pkg/handlers"
)

type Middleware struct {
	db    *database.Supabase
	store *session.Store
}

func New(db *database.Supabase, store *session.Store) *Middleware {
	return &Middleware{db: db, store: store}
}

// fetchUserProfile fetches the user profile from Supabase and sets it in c.Locals("user").
// Returns true if the user profile was successfully set.
func (m *Middleware) fetchUserProfile(c *fiber.Ctx, userID, accessToken string) bool {
	supabaseURL := os.Getenv("SUPABASE_URL")
	if supabaseURL == "" {
		supabaseURL = os.Getenv("NEXT_PUBLIC_SUPABASE_URL")
	}

	url := fmt.Sprintf("%s/rest/v1/user_profiles?id=eq.%s&select=*", supabaseURL, userID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return false
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var profiles []handlers.UserProfile
	json.Unmarshal(body, &profiles)

	if len(profiles) == 0 {
		return false
	}

	c.Locals("user", &profiles[0])
	return true
}

// SetUser is a soft-auth middleware that populates c.Locals("user") if a valid
// session exists. Always calls c.Next() — does NOT block unauthenticated requests.
func (m *Middleware) SetUser(c *fiber.Ctx) error {
	sess, err := m.store.Get(c)
	if err != nil {
		return c.Next()
	}

	userID, _ := sess.Get("user_id").(string)
	accessToken, _ := sess.Get("access_token").(string)
	if userID == "" || accessToken == "" {
		return c.Next()
	}

	m.fetchUserProfile(c, userID, accessToken)
	return c.Next()
}

func (m *Middleware) RequireAuth(c *fiber.Ctx) error {
	user := c.Locals("user")
	if user == nil {
		return c.Redirect("/login")
	}

	profile, ok := user.(*handlers.UserProfile)
	if !ok {
		return c.Redirect("/login")
	}

	// Update streak on first request of the day
	go m.updateStreak(profile.ID)

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

func (m *Middleware) updateStreak(userID string) {
	today := time.Now().Format("2006-01-02")

	body, err := m.db.Select("user_profiles",
		fmt.Sprintf("id=eq.%s&select=current_streak,longest_streak,last_login_date", userID), true)
	if err != nil {
		return
	}
	var profiles []map[string]interface{}
	json.Unmarshal(body, &profiles)
	if len(profiles) == 0 {
		return
	}

	lastLogin, _ := profiles[0]["last_login_date"].(string)
	currentStreak, _ := profiles[0]["current_streak"].(float64)
	longestStreak, _ := profiles[0]["longest_streak"].(float64)

	if lastLogin == today {
		return
	}

	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	newStreak := 1
	if lastLogin == yesterday {
		newStreak = int(currentStreak) + 1
	}

	newLongest := int(longestStreak)
	if newStreak > newLongest {
		newLongest = newStreak
	}

	updateData := fmt.Sprintf(`{"current_streak":%d,"longest_streak":%d,"last_login_date":"%s"}`, newStreak, newLongest, today)
	m.db.Update("user_profiles", fmt.Sprintf("id=eq.%s", userID), []byte(updateData), true)
}
