package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"strings"

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
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(`<!DOCTYPE html>
<html>
<head>
    <title>Connexion...</title>
</head>
<body>
    <p>Connexion en cours...</p>
    <script>
        var hash = window.location.hash.substring(1);
        var params = new URLSearchParams(hash);
        var accessToken = params.get('access_token');
        var refreshToken = params.get('refresh_token');
        
        if (accessToken) {
            var xhr = new XMLHttpRequest();
            xhr.open('POST', '/auth/session', true);
            xhr.setRequestHeader('Content-Type', 'application/json');
            xhr.onreadystatechange = function() {
                if (xhr.readyState === 4) {
                    window.location.href = '/dashboard';
                }
            };
            xhr.send(JSON.stringify({access_token: accessToken, refresh_token: refreshToken}));
        } else {
            window.location.href = '/login?error=no_token';
        }
    </script>
</body>
</html>`)
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

// API Handlers
func (h *Handler) APIGetFriends(c *fiber.Ctx) error {
	userID := c.Query("user_id")
	supabaseURL := os.Getenv("SUPABASE_URL")
	anonKey := os.Getenv("SUPABASE_ANON_KEY")

	// Récupérer les amitiés acceptées
	url := fmt.Sprintf("%s/rest/v1/friendships?or=(requester_id.eq.%s,addressee_id.eq.%s)&status=eq.accepted", supabaseURL, userID, userID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("apikey", anonKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return c.JSON(fiber.Map{"friends": []interface{}{}})
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var friendships []map[string]interface{}
	json.Unmarshal(body, &friendships)

	friends := []map[string]interface{}{}
	for _, f := range friendships {
		requesterID, _ := f["requester_id"].(string)
		addresseeID, _ := f["addressee_id"].(string)
		friendID := requesterID
		if friendID == userID {
			friendID = addresseeID
		}
		friendshipID, _ := f["id"].(string)

		profileURL := fmt.Sprintf("%s/rest/v1/user_profiles?id=eq.%s&select=id,username,nickname,avatar_url,rank,level", supabaseURL, friendID)
		profileReq, _ := http.NewRequest("GET", profileURL, nil)
		profileReq.Header.Set("apikey", anonKey)
		profileResp, err := http.DefaultClient.Do(profileReq)
		if err == nil {
			defer profileResp.Body.Close()
			profileBody, _ := io.ReadAll(profileResp.Body)
			var profiles []map[string]interface{}
			json.Unmarshal(profileBody, &profiles)
			if len(profiles) > 0 {
				friend := profiles[0]
				friend["friendship_id"] = friendshipID
				friends = append(friends, friend)
			}
		}
	}

	return c.JSON(fiber.Map{"friends": friends})
}

func (h *Handler) APIGetFriendRequests(c *fiber.Ctx) error {
	userID := c.Query("user_id")
	supabaseURL := os.Getenv("SUPABASE_URL")
	anonKey := os.Getenv("SUPABASE_ANON_KEY")

	url := fmt.Sprintf("%s/rest/v1/friendships?addressee_id=eq.%s&status=eq.pending", supabaseURL, userID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("apikey", anonKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return c.JSON(fiber.Map{"requests": []interface{}{}})
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var friendships []map[string]interface{}
	json.Unmarshal(body, &friendships)

	requests := []map[string]interface{}{}
	for _, f := range friendships {
		requesterID, _ := f["requester_id"].(string)
		friendshipID, _ := f["id"].(string)

		profileURL := fmt.Sprintf("%s/rest/v1/user_profiles?id=eq.%s&select=id,username,nickname,avatar_url,rank", supabaseURL, requesterID)
		profileReq, _ := http.NewRequest("GET", profileURL, nil)
		profileReq.Header.Set("apikey", anonKey)
		profileResp, err := http.DefaultClient.Do(profileReq)
		if err == nil {
			defer profileResp.Body.Close()
			profileBody, _ := io.ReadAll(profileResp.Body)
			var profiles []map[string]interface{}
			json.Unmarshal(profileBody, &profiles)
			if len(profiles) > 0 {
				req := profiles[0]
				req["friendship_id"] = friendshipID
				requests = append(requests, req)
			}
		}
	}

	return c.JSON(fiber.Map{"requests": requests})
}

func (h *Handler) APISendFriendRequest(c *fiber.Ctx) error {
	type Request struct {
		UserID string `json:"user_id"`
	}
	var req Request
	c.BodyParser(&req)

	sess, _ := h.store.Get(c)
	senderID := sess.Get("user_id").(string)

	supabaseURL := os.Getenv("SUPABASE_URL")
	anonKey := os.Getenv("SUPABASE_ANON_KEY")
	serviceKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")

	// Vérifier si déjà amis
	checkURL := fmt.Sprintf("%s/rest/v1/friendships?or=(and(requester_id.eq.%s,addressee_id.eq.%s),and(requester_id.eq.%s,addressee_id.eq.%s))", supabaseURL, senderID, req.UserID, req.UserID, senderID)
	checkReq, _ := http.NewRequest("GET", checkURL, nil)
	checkReq.Header.Set("apikey", anonKey)
	checkResp, _ := http.DefaultClient.Do(checkReq)
	if checkResp != nil {
		defer checkResp.Body.Close()
		checkBody, _ := io.ReadAll(checkResp.Body)
		var existing []interface{}
		json.Unmarshal(checkBody, &existing)
		if len(existing) > 0 {
			return c.JSON(fiber.Map{"error": "Demande déjà envoyée"})
		}
	}

	// Envoyer la demande
	insertURL := fmt.Sprintf("%s/rest/v1/friendships", supabaseURL)
	insertData := fmt.Sprintf(`{"requester_id":"%s","addressee_id":"%s","status":"pending"}`, senderID, req.UserID)
	insertReq, _ := http.NewRequest("POST", insertURL, strings.NewReader(insertData))
	insertReq.Header.Set("apikey", serviceKey)
	insertReq.Header.Set("Authorization", "Bearer "+serviceKey)
	insertReq.Header.Set("Content-Type", "application/json")
	insertReq.Header.Set("Prefer", "return=representation")
	http.DefaultClient.Do(insertReq)

	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) APIAcceptFriendRequest(c *fiber.Ctx) error {
	type Request struct {
		FriendshipID string `json:"friendship_id"`
	}
	var req Request
	c.BodyParser(&req)

	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")

	url := fmt.Sprintf("%s/rest/v1/friendships?id=eq.%s", supabaseURL, req.FriendshipID)
	updateReq, _ := http.NewRequest("PATCH", url, strings.NewReader(`{"status":"accepted"}`))
	updateReq.Header.Set("apikey", serviceKey)
	updateReq.Header.Set("Authorization", "Bearer "+serviceKey)
	updateReq.Header.Set("Content-Type", "application/json")
	http.DefaultClient.Do(updateReq)

	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) APIRejectFriendRequest(c *fiber.Ctx) error {
	type Request struct {
		FriendshipID string `json:"friendship_id"`
	}
	var req Request
	c.BodyParser(&req)

	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")

	url := fmt.Sprintf("%s/rest/v1/friendships?id=eq.%s", supabaseURL, req.FriendshipID)
	updateReq, _ := http.NewRequest("PATCH", url, strings.NewReader(`{"status":"rejected"}`))
	updateReq.Header.Set("apikey", serviceKey)
	updateReq.Header.Set("Authorization", "Bearer "+serviceKey)
	updateReq.Header.Set("Content-Type", "application/json")
	http.DefaultClient.Do(updateReq)

	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) APIRemoveFriend(c *fiber.Ctx) error {
	type Request struct {
		FriendshipID string `json:"friendship_id"`
	}
	var req Request
	c.BodyParser(&req)

	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")

	url := fmt.Sprintf("%s/rest/v1/friendships?id=eq.%s", supabaseURL, req.FriendshipID)
	deleteReq, _ := http.NewRequest("DELETE", url, nil)
	deleteReq.Header.Set("apikey", serviceKey)
	deleteReq.Header.Set("Authorization", "Bearer "+serviceKey)
	http.DefaultClient.Do(deleteReq)

	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) APISearchUsers(c *fiber.Ctx) error {
	query := c.Query("q")
	if len(query) < 2 {
		return c.JSON(fiber.Map{"users": []interface{}{}})
	}

	supabaseURL := os.Getenv("SUPABASE_URL")
	anonKey := os.Getenv("SUPABASE_ANON_KEY")

	url := fmt.Sprintf("%s/rest/v1/user_profiles?or=(username.ilike.*%s*,nickname.ilike.*%s*)&limit=20", supabaseURL, query, query)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("apikey", anonKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return c.JSON(fiber.Map{"users": []interface{}{}})
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var users []map[string]interface{}
	json.Unmarshal(body, &users)

	return c.JSON(fiber.Map{"users": users})
}

func (h *Handler) APIGetConversations(c *fiber.Ctx) error {
	userID := c.Query("user_id")
	supabaseURL := os.Getenv("SUPABASE_URL")
	anonKey := os.Getenv("SUPABASE_ANON_KEY")

	url := fmt.Sprintf("%s/rest/v1/conversations?or=(user1_id.eq.%s,user2_id.eq.%s)&order=last_message_at.desc", supabaseURL, userID, userID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("apikey", anonKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return c.JSON(fiber.Map{"conversations": []interface{}{}})
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var conversations []map[string]interface{}
	json.Unmarshal(body, &conversations)

	result := []map[string]interface{}{}
	for _, conv := range conversations {
		convID, _ := conv["id"].(string)
		user1ID, _ := conv["user1_id"].(string)
		user2ID, _ := conv["user2_id"].(string)
		otherUserID := user1ID
		if otherUserID == userID {
			otherUserID = user2ID
		}

		// Get other user profile
		profileURL := fmt.Sprintf("%s/rest/v1/user_profiles?id=eq.%s&select=id,username,nickname,avatar_url", supabaseURL, otherUserID)
		profileReq, _ := http.NewRequest("GET", profileURL, nil)
		profileReq.Header.Set("apikey", anonKey)
		profileResp, err := http.DefaultClient.Do(profileReq)
		if err == nil {
			defer profileResp.Body.Close()
			profileBody, _ := io.ReadAll(profileResp.Body)
			var profiles []map[string]interface{}
			json.Unmarshal(profileBody, &profiles)
			if len(profiles) > 0 {
				otherUser := profiles[0]
				result = append(result, map[string]interface{}{
					"id":              convID,
					"other_user_id":   otherUserID,
					"other_username":  otherUser["username"],
					"other_nickname":  otherUser["nickname"],
					"last_message":    conv["last_message"],
					"unread_count":    0,
				})
			}
		}
	}

	return c.JSON(fiber.Map{"conversations": result})
}

func (h *Handler) APIGetMessages(c *fiber.Ctx) error {
	conversationID := c.Query("conversation_id")
	supabaseURL := os.Getenv("SUPABASE_URL")
	anonKey := os.Getenv("SUPABASE_ANON_KEY")

	url := fmt.Sprintf("%s/rest/v1/messages?conversation_id=eq.%s&order=created_at.asc&limit=100", supabaseURL, conversationID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("apikey", anonKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return c.JSON(fiber.Map{"messages": []interface{}{}})
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var messages []map[string]interface{}
	json.Unmarshal(body, &messages)

	return c.JSON(fiber.Map{"messages": messages})
}

func (h *Handler) APISendMessage(c *fiber.Ctx) error {
	type Request struct {
		ConversationID string `json:"conversation_id"`
		Content        string `json:"content"`
	}
	var req Request
	c.BodyParser(&req)

	sess, _ := h.store.Get(c)
	senderID := sess.Get("user_id").(string)

	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")

	insertURL := fmt.Sprintf("%s/rest/v1/messages", supabaseURL)
	insertData := fmt.Sprintf(`{"conversation_id":"%s","sender_id":"%s","content":"%s"}`, req.ConversationID, senderID, req.Content)
	insertReq, _ := http.NewRequest("POST", insertURL, strings.NewReader(insertData))
	insertReq.Header.Set("apikey", serviceKey)
	insertReq.Header.Set("Authorization", "Bearer "+serviceKey)
	insertReq.Header.Set("Content-Type", "application/json")
	http.DefaultClient.Do(insertReq)

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
	// Récupérer les quiz depuis Supabase
	supabaseURL := os.Getenv("SUPABASE_URL")
	if supabaseURL == "" {
		supabaseURL = os.Getenv("NEXT_PUBLIC_SUPABASE_URL")
	}

	url := fmt.Sprintf("%s/rest/v1/quizzes?is_visible=eq.true&order=play_count.desc&limit=20", supabaseURL)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))

	resp, err := http.DefaultClient.Do(req)
	quizzes := []map[string]interface{}{}
	if err == nil {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		json.Unmarshal(body, &quizzes)
	}

	quizCards := ""
	for _, q := range quizzes {
		title, _ := q["title"].(string)
		series, _ := q["series"].(string)
		questionCount, _ := q["question_count"].(float64)
		playCount, _ := q["play_count"].(float64)
		id, _ := q["id"].(string)
		quizType, _ := q["quiz_type"].(string)

		badge := ""
		if quizType == "official" {
			badge = `<span class="badge-official">Officiel</span>`
		}

		quizCards += fmt.Sprintf(`
<div class="quiz-card">
    <div class="quiz-header">%s</div>
    <h3>%s</h3>
    <p class="quiz-meta">%s • %d questions • %d plays</p>
    %s
    <div class="quiz-actions">
        <a href="/quiz/%s/play" class="btn-primary">Jouer</a>
        <a href="/quiz/%s" class="btn-secondary">Voir</a>
    </div>
</div>`, series, title, series, int(questionCount), int(playCount), badge, id, id)
	}

	if quizCards == "" {
		quizCards = `<p style="color: #94a3b8; text-align: center; padding: 40px;">Aucun quiz disponible</p>`
	}

	return renderPage(c, "Explorer", fmt.Sprintf(`
<h1 style="margin-bottom: 24px;">Explorer les quiz</h1>
<div class="quiz-grid">%s</div>
	`, quizCards))
}

func (h *Handler) QuizDetail(c *fiber.Ctx) error {
	id := c.Params("id")
	
	supabaseURL := os.Getenv("SUPABASE_URL")
	if supabaseURL == "" {
		supabaseURL = os.Getenv("NEXT_PUBLIC_SUPABASE_URL")
	}

	// Récupérer le quiz
	url := fmt.Sprintf("%s/rest/v1/quizzes?id=eq.%s&select=*,creator:creator_id(username,avatar_url)", supabaseURL, id)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))

	resp, err := http.DefaultClient.Do(req)
	quizzes := []map[string]interface{}{}
	if err == nil {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		json.Unmarshal(body, &quizzes)
	}

	if len(quizzes) == 0 {
		return renderPage(c, "Quiz introuvable", `<h1>Quiz introuvable</h1><a href="/explore" class="btn-primary">Retour</a>`)
	}

	quiz := quizzes[0]
	title, _ := quiz["title"].(string)
	description, _ := quiz["description"].(string)
	series, _ := quiz["series"].(string)
	category, _ := quiz["category"].(string)
	questionCount, _ := quiz["question_count"].(float64)
	playCount, _ := quiz["play_count"].(float64)
	quizType, _ := quiz["quiz_type"].(string)

	badge := ""
	if quizType == "official" {
		badge = `<span class="badge-official">Officiel</span>`
	} else if quizType == "challenge" {
		badge = `<span class="badge-challenge">Défi</span>`
	}

	return renderPage(c, title, fmt.Sprintf(`
<a href="/explore" style="color: #6366f1; display: inline-block; margin-bottom: 16px;">← Retour</a>
<div class="quiz-detail">
    <div class="quiz-detail-header">
        <div>
            <h1>%s</h1>
            %s
            <p class="quiz-detail-meta">%s • %s</p>
        </div>
        %s
    </div>
    %s
    <div class="quiz-detail-stats">
        <div class="stat"><span class="stat-value">%d</span><span class="stat-label">Questions</span></div>
        <div class="stat"><span class="stat-value">%.0f</span><span class="stat-label">Parties</span></div>
    </div>
    <div class="quiz-detail-actions">
        <a href="/quiz/%s/play" class="btn-primary">🎮 Jouer maintenant</a>
        <a href="/leaderboard/quiz/%s" class="btn-secondary">🏆 Classement</a>
    </div>
</div>

<style>
.quiz-detail { background: #16213e; border: 1px solid #2d2d44; border-radius: 12px; padding: 32px; }
.quiz-detail-header { display: flex; justify-content: space-between; align-items: start; margin-bottom: 24px; }
.quiz-detail-header h1 { font-size: 2rem; margin-bottom: 8px; }
.quiz-detail-meta { color: #94a3b8; }
.quiz-detail-description { color: #94a3b8; margin-bottom: 24px; line-height: 1.6; }
.quiz-detail-stats { display: flex; gap: 32px; margin-bottom: 24px; padding: 16px; background: #1a1a2e; border-radius: 8px; }
.stat { display: flex; flex-direction: column; align-items: center; }
.stat-value { font-size: 1.5rem; font-weight: bold; color: #6366f1; }
.stat-label { font-size: 0.8rem; color: #94a3b8; }
.quiz-detail-actions { display: flex; gap: 12px; }
.badge-challenge { display: inline-block; background: linear-gradient(135deg, #a855f7, #ec4899); color: white; padding: 2px 8px; border-radius: 4px; font-size: 0.75rem; font-weight: 600; }
</style>
	`, title, series, category, series, badge, 
	func() string { if description != "" { return fmt.Sprintf(`<p class="quiz-detail-description">%s</p>`, description) }; return "" }(),
	int(questionCount), playCount, id, id))
}

func (h *Handler) QuizLeaderboard(c *fiber.Ctx) error {
	id := c.Params("id")

	supabaseURL := os.Getenv("SUPABASE_URL")
	if supabaseURL == "" {
		supabaseURL = os.Getenv("NEXT_PUBLIC_SUPABASE_URL")
	}

	// Récupérer le quiz
	quizURL := fmt.Sprintf("%s/rest/v1/quizzes?id=eq.%s&select=title", supabaseURL, id)
	quizReq, _ := http.NewRequest("GET", quizURL, nil)
	quizReq.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	quizResp, _ := http.DefaultClient.Do(quizReq)
	quizzes := []map[string]interface{}{}
	if quizResp != nil {
		defer quizResp.Body.Close()
		body, _ := io.ReadAll(quizResp.Body)
		json.Unmarshal(body, &quizzes)
	}

	quizTitle := "Quiz"
	if len(quizzes) > 0 {
		quizTitle, _ = quizzes[0]["title"].(string)
	}

	// Récupérer le classement
	url := fmt.Sprintf("%s/rest/v1/game_sessions?quiz_id=eq.%s&completed_at=not.is.null&order=score.desc&limit=50&select=*,user:user_id(username,avatar_url,rank)", supabaseURL, id)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))

	resp, err := http.DefaultClient.Do(req)
	sessions := []map[string]interface{}{}
	if err == nil {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		json.Unmarshal(body, &sessions)
	}

	rows := ""
	for i, s := range sessions {
		score, _ := s["score"].(float64)
		user, _ := s["user"].(map[string]interface{})
		username := "?"
		rank := "F"
		if user != nil {
			username, _ = user["username"].(string)
			rank, _ = user["rank"].(string)
		}

		medal := fmt.Sprintf("#%d", i+1)
		if i == 0 { medal = "🥇" } else if i == 1 { medal = "🥈" } else if i == 2 { medal = "🥉" }

		rows += fmt.Sprintf(`<tr><td class="rank-cell">%s</td><td><span class="username">%s</span><span class="user-rank">%s</span></td><td class="xp-cell">%.0f pts</td></tr>`, medal, username, rank, score)
	}

	if rows == "" {
		rows = `<tr><td colspan="3" style="text-align:center;color:#94a3b8;padding:24px;">Aucun score</td></tr>`
	}

	return renderPage(c, "Classement - "+quizTitle, fmt.Sprintf(`
<a href="/quiz/%s" style="color:#6366f1;display:inline-block;margin-bottom:16px;">← Retour au quiz</a>
<h1 style="margin-bottom:24px;">🏆 %s</h1>
<div class="leaderboard">
    <table><thead><tr><th>Rang</th><th>Joueur</th><th>Score</th></tr></thead>
    <tbody>%s</tbody></table>
</div>
	`, id, quizTitle, rows))
}

func (h *Handler) QuizPlay(c *fiber.Ctx) error {
	id := c.Params("id")

	supabaseURL := os.Getenv("SUPABASE_URL")
	if supabaseURL == "" {
		supabaseURL = os.Getenv("NEXT_PUBLIC_SUPABASE_URL")
	}

	// Récupérer les questions
	url := fmt.Sprintf("%s/rest/v1/questions?quiz_id=eq.%s&order=order_index.asc&select=*,answers(*)", supabaseURL, id)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))

	resp, err := http.DefaultClient.Do(req)
	questions := []map[string]interface{}{}
	if err == nil {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		json.Unmarshal(body, &questions)
	}

	// Convertir en JSON pour le JS
	questionsJSON, _ := json.Marshal(questions)

	return renderPage(c, "Jouer", fmt.Sprintf(`
<div id="quiz-container">
    <div id="countdown" class="countdown">
        <div class="countdown-number">3</div>
    </div>
    <div id="quiz-content" style="display:none">
        <div class="quiz-header">
            <div class="quiz-progress">
                <div class="progress-bar"><div id="progress-fill" class="progress-fill"></div></div>
                <span id="question-counter">Question 1/%d</span>
            </div>
            <div id="timer" class="timer">30s</div>
        </div>
        <div id="question-container" class="question-container">
            <h2 id="question-text"></h2>
            <div id="answers-container" class="answers-grid"></div>
        </div>
    </div>
    <div id="results" style="display:none" class="results">
        <h2>Quiz terminé !</h2>
        <div id="score-display"></div>
        <div id="xp-display"></div>
        <a href="/explore" class="btn-primary">Retour</a>
    </div>
</div>

<script>
(function() {
    var questions = %s;
    var currentQuestion = 0;
    var score = 0;
    var totalQuestions = questions.length;
    var timerInterval = null;
    var timeLeft = 30;
    
    // Countdown
    var count = 3;
    var countdownEl = document.getElementById('countdown');
    var countdownNum = countdownEl.querySelector('.countdown-number');
    
    var countInterval = setInterval(function() {
        count--;
        if (count > 0) {
            countdownNum.textContent = count;
        } else {
            countdownNum.textContent = 'GO!';
            clearInterval(countInterval);
            setTimeout(function() {
                countdownEl.style.display = 'none';
                document.getElementById('quiz-content').style.display = 'block';
                loadQuestion();
            }, 500);
        }
    }, 1000);
    
    function loadQuestion() {
        if (currentQuestion >= totalQuestions) {
            showResults();
            return;
        }
        
        var q = questions[currentQuestion];
        document.getElementById('question-text').textContent = q.question_text;
        document.getElementById('question-counter').textContent = 'Question ' + (currentQuestion + 1) + '/' + totalQuestions;
        document.getElementById('progress-fill').style.width = ((currentQuestion + 1) / totalQuestions * 100) + '%%';
        
        // Timer
        timeLeft = q.time_limit_seconds || 30;
        document.getElementById('timer').textContent = timeLeft + 's';
        clearInterval(timerInterval);
        timerInterval = setInterval(function() {
            timeLeft--;
            document.getElementById('timer').textContent = timeLeft + 's';
            if (timeLeft <= 0) {
                clearInterval(timerInterval);
                currentQuestion++;
                loadQuestion();
            }
        }, 1000);
        
        // Answers
        var container = document.getElementById('answers-container');
        container.innerHTML = '';
        var answers = q.answers || [];
        answers.sort(function(a, b) { return (a.order_index || 0) - (b.order_index || 0); });
        answers.forEach(function(answer) {
            var btn = document.createElement('button');
            btn.className = 'answer-btn';
            btn.textContent = answer.answer_text;
            btn.onclick = function() { selectAnswer(answer, btn); };
            container.appendChild(btn);
        });
    }
    
    function selectAnswer(answer, btn) {
        clearInterval(timerInterval);
        if (answer.is_correct) {
            btn.classList.add('correct');
            score++;
        } else {
            btn.classList.add('wrong');
        }
        
        setTimeout(function() {
            currentQuestion++;
            loadQuestion();
        }, 500);
    }
    
    function showResults() {
        document.getElementById('quiz-content').style.display = 'none';
        document.getElementById('results').style.display = 'block';
        document.getElementById('score-display').innerHTML = '<h3>' + score + ' bonnes réponses sur ' + totalQuestions + '</h3>';
        document.getElementById('xp-display').innerHTML = '<p>+' + (score * 5) + ' XP gagnés</p>';
    }
})();
</script>

<style>
.countdown { display: flex; align-items: center; justify-content: center; min-height: 60vh; }
.countdown-number { font-size: 6rem; font-weight: bold; color: #6366f1; }
.quiz-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.quiz-progress { flex: 1; margin-right: 16px; }
.progress-bar { height: 8px; background: #2d2d44; border-radius: 4px; overflow: hidden; }
.progress-fill { height: 100%%; background: #6366f1; transition: width 0.3s; }
.timer { font-size: 1.5rem; font-weight: bold; color: #6366f1; min-width: 50px; text-align: center; }
.question-container { background: #16213e; border: 1px solid #2d2d44; border-radius: 12px; padding: 32px; }
.question-container h2 { font-size: 1.25rem; margin-bottom: 24px; }
.answers-grid { display: flex; flex-direction: column; gap: 12px; }
.answer-btn { padding: 16px; background: #1a1a2e; border: 2px solid #2d2d44; border-radius: 8px; color: white; font-size: 1rem; cursor: pointer; transition: all 0.2s; text-align: left; }
.answer-btn:hover { border-color: #6366f1; background: rgba(99,102,241,0.1); }
.answer-btn.correct { border-color: #22c55e; background: rgba(34,197,94,0.1); }
.answer-btn.wrong { border-color: #ef4444; background: rgba(239,68,68,0.1); }
.results { text-align: center; padding: 60px 20px; }
.results h2 { font-size: 2rem; margin-bottom: 24px; }
</style>
	`, len(questions), string(questionsJSON)))
}

func (h *Handler) QuizSubmit(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"success": true, "xpEarned": 0})
}

func (h *Handler) Friends(c *fiber.Ctx) error {
	user := c.Locals("user").(*UserProfile)

	return renderPage(c, "Amis", fmt.Sprintf(`
<div class="friends-container">
    <div class="friends-tabs">
        <button class="tab-btn active" onclick="showTab('amis')">👥 Amis</button>
        <button class="tab-btn" onclick="showTab('rechercher')">🔍 Rechercher</button>
        <button class="tab-btn" onclick="showTab('demandes')">🔔 Demandes</button>
        <button class="tab-btn" onclick="showTab('chat')">💬 Chat</button>
    </div>

    <!-- Tab Amis -->
    <div id="tab-amis" class="tab-content active">
        <div id="friends-list" class="friends-list">
            <p style="color: #94a3b8; text-align: center; padding: 40px;">Chargement...</p>
        </div>
    </div>

    <!-- Tab Rechercher -->
    <div id="tab-rechercher" class="tab-content" style="display:none">
        <div class="search-box">
            <input type="text" id="search-input" placeholder="Rechercher un utilisateur..." onkeyup="searchUsers(this.value)">
        </div>
        <div id="search-results" class="friends-list"></div>
    </div>

    <!-- Tab Demandes -->
    <div id="tab-demandes" class="tab-content" style="display:none">
        <div id="requests-list" class="friends-list">
            <p style="color: #94a3b8; text-align: center; padding: 40px;">Chargement...</p>
        </div>
    </div>

    <!-- Tab Chat -->
    <div id="tab-chat" class="tab-content" style="display:none">
        <div id="conversations-list" class="friends-list">
            <p style="color: #94a3b8; text-align: center; padding: 40px;">Chargement...</p>
        </div>
    </div>

    <!-- Chat Window (hidden by default) -->
    <div id="chat-window" style="display:none">
        <div class="chat-header">
            <button onclick="closeChat()" style="background:none;border:none;color:white;cursor:pointer;font-size:1.2rem;">←</button>
            <span id="chat-username" style="font-weight:600;"></span>
        </div>
        <div id="chat-messages" class="chat-messages"></div>
        <div class="chat-input">
            <input type="text" id="message-input" placeholder="Écrire un message..." onkeydown="if(event.key==='Enter')sendMessage()">
            <button onclick="sendMessage()" class="btn-sm">Envoyer</button>
        </div>
    </div>
</div>

<script>
var currentUserId = '%s';

function showTab(name) {
    document.querySelectorAll('.tab-content').forEach(function(t) { t.style.display = 'none'; });
    document.querySelectorAll('.tab-btn').forEach(function(b) { b.classList.remove('active'); });
    document.getElementById('tab-' + name).style.display = 'block';
    event.target.classList.add('active');
    
    if (name === 'amis') loadFriends();
    if (name === 'demandes') loadRequests();
}

function loadFriends() {
    fetch('/api/friends?user_id=' + currentUserId)
        .then(function(r) { return r.json(); })
        .then(function(data) {
            var html = '';
            if (data.friends && data.friends.length > 0) {
                data.friends.forEach(function(f) {
                    html += '<div class="friend-card">';
                    html += '<div class="friend-avatar">' + (f.username ? f.username[0].toUpperCase() : '?') + '</div>';
                    html += '<div class="friend-info"><span class="friend-name">' + (f.nickname || f.username) + '</span>';
                    html += '<span class="friend-rank">' + f.rank + ' • Niv. ' + f.level + '</span></div>';
                    html += '<div class="friend-actions">';
                    html += '<a href="/chat/' + f.id + '" class="btn-sm">💬</a>';
                    html += '<button class="btn-sm btn-danger" onclick="removeFriend(\'' + f.friendship_id + '\')">✕</button>';
                    html += '</div></div>';
                });
            } else {
                html = '<p style="color: #94a3b8; text-align: center; padding: 40px;">Aucun ami. Recherchez des utilisateurs !</p>';
            }
            document.getElementById('friends-list').innerHTML = html;
        });
}

function searchUsers(query) {
    if (query.length < 2) { document.getElementById('search-results').innerHTML = ''; return; }
    fetch('/api/users/search?q=' + encodeURIComponent(query))
        .then(function(r) { return r.json(); })
        .then(function(data) {
            var html = '';
            if (data.users && data.users.length > 0) {
                data.users.forEach(function(u) {
                    html += '<div class="friend-card">';
                    html += '<div class="friend-avatar">' + u.username[0].toUpperCase() + '</div>';
                    html += '<div class="friend-info"><span class="friend-name">' + (u.nickname || u.username) + '</span>';
                    html += '<span class="friend-rank">' + u.rank + ' • Niv. ' + u.level + '</span></div>';
                    html += '<button class="btn-sm" onclick="sendRequest(\'' + u.id + '\')">Ajouter</button>';
                    html += '</div>';
                });
            } else {
                html = '<p style="color: #94a3b8; text-align: center; padding: 20px;">Aucun résultat</p>';
            }
            document.getElementById('search-results').innerHTML = html;
        });
}

function sendRequest(userId) {
    fetch('/api/friends/request', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({user_id: userId})
    }).then(function(r) { return r.json(); }).then(function(data) {
        if (data.success) alert('Demande envoyée !');
        else alert(data.error || 'Erreur');
    });
}

function loadRequests() {
    fetch('/api/friends/requests?user_id=' + currentUserId)
        .then(function(r) { return r.json(); })
        .then(function(data) {
            var html = '';
            if (data.requests && data.requests.length > 0) {
                data.requests.forEach(function(r) {
                    html += '<div class="friend-card">';
                    html += '<div class="friend-avatar">' + r.username[0].toUpperCase() + '</div>';
                    html += '<div class="friend-info"><span class="friend-name">' + (r.nickname || r.username) + '</span></div>';
                    html += '<div class="friend-actions">';
                    html += '<button class="btn-sm btn-success" onclick="acceptRequest(\'' + r.friendship_id + '\')">✓</button>';
                    html += '<button class="btn-sm btn-danger" onclick="rejectRequest(\'' + r.friendship_id + '\')">✕</button>';
                    html += '</div></div>';
                });
            } else {
                html = '<p style="color: #94a3b8; text-align: center; padding: 40px;">Aucune demande</p>';
            }
            document.getElementById('requests-list').innerHTML = html;
        });
}

function acceptRequest(id) {
    fetch('/api/friends/accept', {method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify({friendship_id: id})})
        .then(function() { loadRequests(); loadFriends(); });
}

function rejectRequest(id) {
    fetch('/api/friends/reject', {method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify({friendship_id: id})})
        .then(function() { loadRequests(); });
}

function removeFriend(id) {
    if (!confirm('Supprimer cet ami ?')) return;
    fetch('/api/friends/remove', {method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify({friendship_id: id})})
        .then(function() { loadFriends(); });
}

loadFriends();

// Chat functions
var currentConversationId = null;
var currentChatFriendId = null;

function showTab(name) {
    document.querySelectorAll('.tab-content').forEach(function(t) { t.style.display = 'none'; });
    document.getElementById('chat-window').style.display = 'none';
    document.querySelectorAll('.tab-btn').forEach(function(b) { b.classList.remove('active'); });
    document.getElementById('tab-' + name).style.display = 'block';
    event.target.classList.add('active');
    
    if (name === 'amis') loadFriends();
    if (name === 'demandes') loadRequests();
    if (name === 'chat') loadConversations();
}

function loadConversations() {
    fetch('/api/conversations?user_id=' + currentUserId)
        .then(function(r) { return r.json(); })
        .then(function(data) {
            var html = '';
            if (data.conversations && data.conversations.length > 0) {
                data.conversations.forEach(function(conv) {
                    html += '<div class="friend-card" onclick="openChat(\'' + conv.id + '\', \'' + conv.other_user_id + '\', \'' + (conv.other_username || 'Utilisateur') + '\')">';
                    html += '<div class="friend-avatar">' + (conv.other_username ? conv.other_username[0].toUpperCase() : '?') + '</div>';
                    html += '<div class="friend-info"><span class="friend-name">' + (conv.other_nickname || conv.other_username || 'Utilisateur') + '</span>';
                    html += '<span class="friend-rank">' + (conv.last_message || 'Aucun message') + '</span></div>';
                    if (conv.unread_count > 0) {
                        html += '<span class="badge">' + conv.unread_count + '</span>';
                    }
                    html += '</div>';
                });
            } else {
                html = '<p style="color: #94a3b8; text-align: center; padding: 40px;">Aucune conversation</p>';
            }
            document.getElementById('conversations-list').innerHTML = html;
        });
}

function openChat(convId, friendId, username) {
    currentConversationId = convId;
    currentChatFriendId = friendId;
    document.getElementById('tab-chat').style.display = 'none';
    document.getElementById('chat-window').style.display = 'flex';
    document.getElementById('chat-username').textContent = username;
    loadMessages();
}

function closeChat() {
    document.getElementById('chat-window').style.display = 'none';
    document.getElementById('tab-chat').style.display = 'block';
    currentConversationId = null;
    currentChatFriendId = null;
    loadConversations();
}

function loadMessages() {
    if (!currentConversationId) return;
    fetch('/api/messages?conversation_id=' + currentConversationId)
        .then(function(r) { return r.json(); })
        .then(function(data) {
            var html = '';
            if (data.messages && data.messages.length > 0) {
                data.messages.forEach(function(msg) {
                    var isOwn = msg.sender_id === currentUserId;
                    html += '<div class="message ' + (isOwn ? 'message-own' : 'message-other') + '">';
                    html += '<div class="message-bubble">' + msg.content + '</div>';
                    html += '</div>';
                });
            } else {
                html = '<p style="color: #94a3b8; text-align: center; padding: 40px;">Aucun message</p>';
            }
            document.getElementById('chat-messages').innerHTML = html;
            document.getElementById('chat-messages').scrollTop = document.getElementById('chat-messages').scrollHeight;
        });
}

function sendMessage() {
    var input = document.getElementById('message-input');
    var content = input.value.trim();
    if (!content || !currentConversationId) return;
    
    fetch('/api/messages/send', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({conversation_id: currentConversationId, content: content})
    }).then(function(r) { return r.json(); }).then(function(data) {
        if (data.success) {
            input.value = '';
            loadMessages();
        }
    });
}

// Auto-refresh messages
setInterval(function() {
    if (currentConversationId && document.getElementById('chat-window').style.display !== 'none') {
        loadMessages();
    }
}, 3000);
</script>

<style>
.friends-container { max-width: 600px; margin: 0 auto; }
.friends-tabs { display: flex; gap: 8px; margin-bottom: 24px; flex-wrap: wrap; }
.tab-btn { padding: 8px 16px; background: #1a1a2e; border: 1px solid #2d2d44; border-radius: 8px; color: #94a3b8; cursor: pointer; transition: all 0.2s; }
.tab-btn.active { background: #6366f1; color: white; border-color: #6366f1; }
.friends-list { display: flex; flex-direction: column; gap: 12px; }
.friend-card { background: #16213e; border: 1px solid #2d2d44; border-radius: 12px; padding: 16px; display: flex; align-items: center; gap: 12px; }
.friend-avatar { width: 48px; height: 48px; border-radius: 50%; background: #6366f1; display: flex; align-items: center; justify-content: center; font-weight: bold; font-size: 1.2rem; }
.friend-info { flex: 1; display: flex; flex-direction: column; }
.friend-name { font-weight: 600; }
.friend-rank { font-size: 0.8rem; color: #94a3b8; }
.friend-actions { display: flex; gap: 8px; }
.search-box { margin-bottom: 16px; }
.search-box input { width: 100%%; padding: 12px; background: #1a1a2e; border: 1px solid #2d2d44; border-radius: 8px; color: white; font-size: 1rem; }
.btn-sm { padding: 6px 12px; font-size: 0.8rem; border-radius: 6px; background: #6366f1; color: white; border: none; cursor: pointer; }
.btn-success { background: #22c55e; }
.btn-danger { background: #ef4444; }
.badge { background: #ef4444; color: white; padding: 2px 8px; border-radius: 12px; font-size: 0.75rem; font-weight: 600; }

/* Chat */
#chat-window { display: none; flex-direction: column; height: 500px; background: #16213e; border: 1px solid #2d2d44; border-radius: 12px; overflow: hidden; }
.chat-header { display: flex; align-items: center; gap: 12px; padding: 12px 16px; background: #1a1a2e; border-bottom: 1px solid #2d2d44; }
.chat-messages { flex: 1; overflow-y: auto; padding: 16px; display: flex; flex-direction: column; gap: 8px; }
.chat-input { display: flex; gap: 8px; padding: 12px 16px; border-top: 1px solid #2d2d44; }
.chat-input input { flex: 1; padding: 10px; background: #1a1a2e; border: 1px solid #2d2d44; border-radius: 8px; color: white; }
.message { display: flex; }
.message-own { justify-content: flex-end; }
.message-other { justify-content: flex-start; }
.message-bubble { max-width: 70%; padding: 8px 12px; border-radius: 12px; font-size: 0.9rem; }
.message-own .message-bubble { background: #6366f1; color: white; border-bottom-right-radius: 4px; }
.message-other .message-bubble { background: #1a1a2e; color: #e2e8f0; border-bottom-left-radius: 4px; }
</style>
	`, user.ID))
}

func (h *Handler) ChallengeDetail(c *fiber.Ctx) error {
	id := c.Params("id")
	
	return renderPage(c, "Défi", fmt.Sprintf(`
<div class="challenge-page">
    <a href="/friends" style="color: #6366f1; display: inline-block; margin-bottom: 16px;">← Retour</a>
    <div class="challenge-card">
        <div class="challenge-header">
            <span class="challenge-icon">⚔️</span>
            <h1>Défi en cours</h1>
        </div>
        <div id="challenge-content" data-id="%s">
            <p style="color: #94a3b8; text-align: center; padding: 40px;">Chargement du défi...</p>
        </div>
    </div>
</div>

<style>
.challenge-page { max-width: 600px; margin: 0 auto; }
.challenge-card { background: #16213e; border: 1px solid #2d2d44; border-radius: 12px; padding: 32px; }
.challenge-header { display: flex; align-items: center; gap: 12px; margin-bottom: 24px; }
.challenge-icon { font-size: 2rem; }
</style>
	`, id))
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
	user := c.Locals("user").(*UserProfile)
	
	nickname := ""
	if user.Nickname != nil {
		nickname = *user.Nickname
	}
	
	return renderPage(c, "Modifier le profil", fmt.Sprintf(`
<h1 style="margin-bottom: 24px;">Modifier le profil</h1>
<div class="card" style="max-width: 600px;">
    <form method="POST" action="/profile/edit">
        <label>Nom d'utilisateur</label>
        <input type="text" value="%s" disabled style="opacity: 0.5;">
        <p style="font-size: 0.75rem; color: #94a3b8; margin-bottom: 16px;">Le nom d'utilisateur ne peut pas être modifié</p>
        
        <label>Surnom / Nickname</label>
        <input type="text" name="nickname" value="%s" placeholder="Votre surnom">
        
        <label>Bio</label>
        <textarea name="bio" rows="3" placeholder="Parlez de vous..."></textarea>
        
        <label>Pays</label>
        <select name="country">
            <option value="">Sélectionner...</option>
            <option>Algérie</option><option>Maroc</option><option>Tunisie</option>
            <option>Sénégal</option><option>Côte d'Ivoire</option><option>Cameroun</option>
            <option>RDC</option><option>Madagascar</option><option>Autre</option>
        </select>
        
        <label>Anime préféré</label>
        <input type="text" name="favorite_anime" placeholder="Ex: Naruto, One Piece...">
        
        <button type="submit" class="btn-submit">Sauvegarder</button>
    </form>
</div>
	`, user.Username, nickname))
}

func (h *Handler) ProfileUpdate(c *fiber.Ctx) error {
	return c.Redirect("/profil")
}

func (h *Handler) Leaderboard(c *fiber.Ctx) error {
	// Récupérer le classement depuis Supabase
	supabaseURL := os.Getenv("SUPABASE_URL")
	if supabaseURL == "" {
		supabaseURL = os.Getenv("NEXT_PUBLIC_SUPABASE_URL")
	}

	url := fmt.Sprintf("%s/rest/v1/user_profiles?order=xp.desc&limit=50", supabaseURL)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))

	resp, err := http.DefaultClient.Do(req)
	users := []map[string]interface{}{}
	if err == nil {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		json.Unmarshal(body, &users)
	}

	rows := ""
	for i, u := range users {
		username, _ := u["username"].(string)
		rank, _ := u["rank"].(string)
		xp, _ := u["xp"].(float64)
		level, _ := u["level"].(float64)

		medal := ""
		if i == 0 {
			medal = "🥇"
		} else if i == 1 {
			medal = "🥈"
		} else if i == 2 {
			medal = "🥉"
		} else {
			medal = fmt.Sprintf("#%d", i+1)
		}

		rows += fmt.Sprintf(`
<tr>
    <td class="rank-cell">%s</td>
    <td>
        <div class="user-info">
            <span class="username">%s</span>
            <span class="user-rank">%s</span>
        </div>
    </td>
    <td class="xp-cell">%d XP</td>
    <td class="level-cell">Niv. %d</td>
</tr>`, medal, username, rank, int(xp), int(level))
	}

	return renderPage(c, "Classement", fmt.Sprintf(`
<h1 style="margin-bottom: 24px;">🏆 Classement</h1>
<div class="leaderboard">
    <table>
        <thead>
            <tr>
                <th>Rang</th>
                <th>Joueur</th>
                <th>XP</th>
                <th>Niveau</th>
            </tr>
        </thead>
        <tbody>%s</tbody>
    </table>
</div>
	`, rows))
}

func (h *Handler) FAQ(c *fiber.Ctx) error {
	return renderPage(c, "FAQ", `
<h1 style="margin-bottom: 24px;">❓ FAQ</h1>

<div class="faq-section">
    <h2>🎮 Concept du site</h2>
    <div class="faq-item">
        <strong>Qu'est-ce que Otaku Quiz Africa ?</strong>
        <p>Otaku Quiz Africa est une plateforme de quiz dédiée à la culture anime/manga en Afrique. Testez vos connaissances, défiez vos amis et progressez dans un système de rangs !</p>
    </div>
    <div class="faq-item">
        <strong>Comment ça marche ?</strong>
        <p>Jouez à des quiz, gagnez de l'XP, montez en rang et débloquez de nouvelles fonctionnalités. Vous pouvez aussi créer vos propres quiz et défier vos amis !</p>
    </div>
</div>

<div class="faq-section">
    <h2>📝 Création de quiz</h2>
    <div class="faq-item">
        <strong>Puis-je créer mes propres quiz ?</strong>
        <p>Oui ! Atteignez le rang C ou obtenez une autorisation spéciale d'un admin pour créer vos propres quiz.</p>
    </div>
    <div class="faq-item">
        <strong>Quels types de questions sont disponibles ?</strong>
        <p>QCM, Vrai/Faux, Image, GIF, Audio, Devine le personnage, Trouve l'intrus.</p>
    </div>
</div>

<div class="faq-section">
    <h2>⚔️ Défis</h2>
    <div class="faq-item">
        <strong>Comment défier un ami ?</strong>
        <p>Allez sur la page d'un quiz et cliquez sur "Défier vos amis". Sélectionnez les amis à inviter et définissez votre mise XP.</p>
    </div>
    <div class="faq-item">
        <strong>Comment fonctionne la mise XP ?</strong>
        <p>Vous misez un montant d'XP. Le ou les gagnants remportent la totalité de l'XP misé par tous les participants.</p>
    </div>
</div>

<div class="faq-section">
    <h2>🏆 Système de rangs</h2>
    <div class="faq-item">
        <strong>Quels sont les rangs disponibles ?</strong>
        <p>F → E → D → C → B → A → S → S+ → SS → SSS → Légende</p>
    </div>
    <div class="faq-item">
        <strong>Comment progresser en rang ?</strong>
        <p>Gagnez de l'XP en jouant aux quiz. Plus vous avez de bonnes réponses et plus vous répondez vite, plus vous gagnez d'XP.</p>
    </div>
</div>

<div class="faq-section">
    <h2>🎯 Événements officiels</h2>
    <div class="faq-item">
        <strong>Qu'est-ce qu'un quiz officiel ?</strong>
        <p>Les quiz officiels sont créés par les admins. Ils ont des récompenses spéciales et un classement public permanent.</p>
    </div>
</div>

<style>
.faq-section {
    margin-bottom: 32px;
}
.faq-section h2 {
    font-size: 1.25rem;
    margin-bottom: 16px;
    color: #6366f1;
}
.faq-item {
    background: #16213e;
    border: 1px solid #2d2d44;
    border-radius: 8px;
    padding: 16px;
    margin-bottom: 12px;
}
.faq-item strong {
    display: block;
    margin-bottom: 8px;
}
.faq-item p {
    color: #94a3b8;
    font-size: 0.9rem;
}
</style>
	`)
}

// Pages admin
// Pages admin
func (h *Handler) AdminDashboard(c *fiber.Ctx) error {
	return renderPage(c, "Admin", `
<h1 style="margin-bottom: 24px;">🛡️ Administration</h1>
<div class="stats">
    <a href="/admin/official-quizzes" class="stat-card" style="text-decoration: none;">
        <div class="label">Quiz Officiels</div>
        <div class="value">Gérer</div>
    </a>
    <a href="/admin/tickets" class="stat-card" style="text-decoration: none;">
        <div class="label">Tickets Support</div>
        <div class="value">Voir</div>
    </a>
    <a href="/admin/announcements" class="stat-card" style="text-decoration: none;">
        <div class="label">Annonces</div>
        <div class="value">Gérer</div>
    </a>
    <a href="/admin/settings" class="stat-card" style="text-decoration: none;">
        <div class="label">Paramètres</div>
        <div class="value">Configurer</div>
    </a>
</div>
	`)
}

func (h *Handler) AdminOfficialQuizzes(c *fiber.Ctx) error {
	supabaseURL := os.Getenv("SUPABASE_URL")
	if supabaseURL == "" {
		supabaseURL = os.Getenv("NEXT_PUBLIC_SUPABASE_URL")
	}

	url := fmt.Sprintf("%s/rest/v1/quizzes?quiz_type=eq.official&order=created_at.desc", supabaseURL)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))

	resp, err := http.DefaultClient.Do(req)
	quizzes := []map[string]interface{}{}
	if err == nil {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		json.Unmarshal(body, &quizzes)
	}

	quizRows := ""
	for _, q := range quizzes {
		title, _ := q["title"].(string)
		status, _ := q["status"].(string)
		id, _ := q["id"].(string)
		
		quizRows += fmt.Sprintf(`
<tr>
    <td>%s</td>
    <td><span class="badge-%s">%s</span></td>
    <td><a href="/quiz/%s" class="btn-sm">Voir</a></td>
</tr>`, title, status, status, id)
	}

	return renderPage(c, "Quiz Officiels", fmt.Sprintf(`
<div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px;">
    <h1>🏆 Quiz Officiels</h1>
    <button class="btn-primary" onclick="alert('Fonctionnalité à venir')">+ Créer</button>
</div>
<div class="leaderboard">
    <table>
        <thead><tr><th>Titre</th><th>Statut</th><th>Actions</th></tr></thead>
        <tbody>%s</tbody>
    </table>
</div>
	`, quizRows))
}

func (h *Handler) AdminTickets(c *fiber.Ctx) error {
	supabaseURL := os.Getenv("SUPABASE_URL")
	if supabaseURL == "" {
		supabaseURL = os.Getenv("NEXT_PUBLIC_SUPABASE_URL")
	}

	url := fmt.Sprintf("%s/rest/v1/admin_conversations?order=created_at.desc&select=*,user:user_id(username,email)", supabaseURL)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))

	resp, err := http.DefaultClient.Do(req)
	tickets := []map[string]interface{}{}
	if err == nil {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		json.Unmarshal(body, &tickets)
	}

	ticketRows := ""
	for _, t := range tickets {
		subject, _ := t["subject"].(string)
		status, _ := t["status"].(string)
		id, _ := t["id"].(string)
		
		ticketRows += fmt.Sprintf(`
<tr>
    <td>%s</td>
    <td><span class="badge-%s">%s</span></td>
    <td><a href="/admin/tickets/%s" class="btn-sm">Voir</a></td>
</tr>`, subject, status, status, id)
	}

	return renderPage(c, "Tickets", fmt.Sprintf(`
<h1 style="margin-bottom: 24px;">🎧 Tickets Support</h1>
<div class="leaderboard">
    <table>
        <thead><tr><th>Sujet</th><th>Statut</th><th>Actions</th></tr></thead>
        <tbody>%s</tbody>
    </table>
</div>
	`, ticketRows))
}

func (h *Handler) AdminAnnouncements(c *fiber.Ctx) error {
	return renderPage(c, "Annonces", `
<div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px;">
    <h1>📢 Annonces</h1>
    <button class="btn-primary" onclick="alert('Fonctionnalité à venir')">+ Créer</button>
</div>
<div class="card">
    <p style="color: #94a3b8; text-align: center; padding: 40px;">Aucune annonce pour le moment</p>
</div>
	`)
}

func (h *Handler) AdminSettings(c *fiber.Ctx) error {
	return renderPage(c, "Paramètres", `
<h1 style="margin-bottom: 24px;">⚙️ Paramètres</h1>

<div class="card" style="margin-bottom: 24px;">
    <h2 style="margin-bottom: 16px;">Rangs autorisés à créer des quiz</h2>
    <p style="color: #94a3b8; margin-bottom: 16px;">Sélectionnez les rangs qui peuvent créer des quiz</p>
    <div style="display: flex; gap: 12px; flex-wrap: wrap;">
        <label style="display: flex; align-items: center; gap: 8px; cursor: pointer;">
            <input type="checkbox" checked> F
        </label>
        <label style="display: flex; align-items: center; gap: 8px; cursor: pointer;">
            <input type="checkbox" checked> E
        </label>
        <label style="display: flex; align-items: center; gap: 8px; cursor: pointer;">
            <input type="checkbox" checked> D
        </label>
        <label style="display: flex; align-items: center; gap: 8px; cursor: pointer;">
            <input type="checkbox" checked> C
        </label>
        <label style="display: flex; align-items: center; gap: 8px; cursor: pointer;">
            <input type="checkbox"> B
        </label>
        <label style="display: flex; align-items: center; gap: 8px; cursor: pointer;">
            <input type="checkbox"> A
        </label>
        <label style="display: flex; align-items: center; gap: 8px; cursor: pointer;">
            <input type="checkbox"> S
        </label>
    </div>
    <button class="btn-primary" style="margin-top: 16px;">Sauvegarder</button>
</div>
	`)
}
