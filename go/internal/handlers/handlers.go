package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

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
	ID               string    `json:"id"`
	Email            string    `json:"email"`
	Username         string    `json:"username"`
	Nickname         *string   `json:"nickname"`
	AvatarURL        *string   `json:"avatar_url"`
	Bio              *string   `json:"bio"`
	Country          *string   `json:"country"`
	FavoriteAnime    *string   `json:"favorite_anime"`
	Phone            *string   `json:"phone"`
	XP               int       `json:"xp"`
	Level            int       `json:"level"`
	Rank             string    `json:"rank"`
	IsAdmin          bool      `json:"is_admin"`
	IsPremium        bool      `json:"is_premium"`
	CanCreateQuiz    bool      `json:"can_create_quiz"`
	CurrentStreak    int       `json:"current_streak"`
	LongestStreak    int       `json:"longest_streak"`
	TotalXP          int       `json:"total_xp"`
	CreatedAt        time.Time `json:"created_at"`
	ChallengesPlayed int       `json:"challenges_played"`
	ChallengesWon    int       `json:"challenges_won"`
}

func renderPage(c *fiber.Ctx, title string, content string) error {
	data := fiber.Map{
		"Title":   title,
		"Content": template.HTML(content),
	}
	if u := c.Locals("user"); u != nil {
		data["User"] = u
		if user, ok := u.(*UserProfile); ok {
			badge := challengeStatsBadge(user.ChallengesPlayed, user.ChallengesWon)
			if badge != "" {
				data["StatsBadge"] = template.HTML(badge)
			}
		}
	}
	return c.Render("layouts/main", data)
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

	body, err := h.db.Select("friendships",
		fmt.Sprintf("or=(requester_id.eq.%s,addressee_id.eq.%s)&status=eq.accepted", userID, userID), true)
	if err != nil {
		return c.JSON(fiber.Map{"friends": []interface{}{}, "error": err.Error()})
	}
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

		pBody, err := h.db.Select("user_profiles",
			fmt.Sprintf("id=eq.%s&select=id,username,nickname,avatar_url,rank,level", friendID), false)
		if err != nil {
			continue
		}
		var profiles []map[string]interface{}
		json.Unmarshal(pBody, &profiles)
		if len(profiles) > 0 {
			friend := profiles[0]
			friend["friendship_id"] = friendshipID
			friends = append(friends, friend)
		}
	}

	return c.JSON(fiber.Map{"friends": friends})
}

func (h *Handler) APIGetFriendRequests(c *fiber.Ctx) error {
	userID := c.Query("user_id")

	body, err := h.db.Select("friendships",
		fmt.Sprintf("addressee_id=eq.%s&status=eq.pending", userID), true)
	if err != nil {
		return c.JSON(fiber.Map{"requests": []interface{}{}, "error": err.Error()})
	}
	var friendships []map[string]interface{}
	json.Unmarshal(body, &friendships)

	requests := []map[string]interface{}{}
	for _, f := range friendships {
		requesterID, _ := f["requester_id"].(string)
		friendshipID, _ := f["id"].(string)

		pBody, err := h.db.Select("user_profiles",
			fmt.Sprintf("id=eq.%s&select=id,username,nickname,avatar_url,rank", requesterID), false)
		if err != nil {
			continue
		}
		var profiles []map[string]interface{}
		json.Unmarshal(pBody, &profiles)
		if len(profiles) > 0 {
			p := profiles[0]
			p["friendship_id"] = friendshipID
			requests = append(requests, p)
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

	checkBody, err := h.db.Select("friendships",
		fmt.Sprintf("or=(and(requester_id.eq.%s,addressee_id.eq.%s),and(requester_id.eq.%s,addressee_id.eq.%s))",
			senderID, req.UserID, req.UserID, senderID), true)
	if err == nil {
		var existing []interface{}
		json.Unmarshal(checkBody, &existing)
		if len(existing) > 0 {
			return c.JSON(fiber.Map{"error": "Demande déjà envoyée"})
		}
	}

	insertData := fmt.Sprintf(`{"requester_id":"%s","addressee_id":"%s","status":"pending"}`, senderID, req.UserID)
	_, err = h.db.Insert("friendships", []byte(insertData), true)
	if err != nil {
		return c.JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) APIAcceptFriendRequest(c *fiber.Ctx) error {
	type Request struct {
		FriendshipID string `json:"friendship_id"`
	}
	var req Request
	c.BodyParser(&req)

	_, err := h.db.Update("friendships", fmt.Sprintf("id=eq.%s", req.FriendshipID),
		[]byte(`{"status":"accepted"}`), true)
	if err != nil {
		return c.JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) APIRejectFriendRequest(c *fiber.Ctx) error {
	type Request struct {
		FriendshipID string `json:"friendship_id"`
	}
	var req Request
	c.BodyParser(&req)

	_, err := h.db.Update("friendships", fmt.Sprintf("id=eq.%s", req.FriendshipID),
		[]byte(`{"status":"rejected"}`), true)
	if err != nil {
		return c.JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) APIRemoveFriend(c *fiber.Ctx) error {
	type Request struct {
		FriendshipID string `json:"friendship_id"`
	}
	var req Request
	c.BodyParser(&req)

	err := h.db.Delete("friendships", fmt.Sprintf("id=eq.%s", req.FriendshipID), true)
	if err != nil {
		return c.JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) APISearchUsers(c *fiber.Ctx) error {
	query := c.Query("q")
	if len(query) < 2 {
		return c.JSON(fiber.Map{"users": []interface{}{}})
	}

	sess, _ := h.store.Get(c)
	currentUserID := ""
	if uid := sess.Get("user_id"); uid != nil {
		currentUserID, _ = uid.(string)
	}

	encodedQuery := strings.ReplaceAll(query, "%", "%25")
	body, err := h.db.Select("user_profiles",
		fmt.Sprintf("or=(username.ilike.*%s*,nickname.ilike.*%s*)&neq.id=%s&order=xp.desc&limit=20", encodedQuery, encodedQuery, currentUserID), false)
	if err != nil {
		return c.JSON(fiber.Map{"users": []interface{}{}, "error": err.Error()})
	}
	var users []map[string]interface{}
	json.Unmarshal(body, &users)

	if currentUserID != "" {
		for i, u := range users {
			uid, _ := u["id"].(string)
			if uid == "" {
				continue
			}
			fBody, err := h.db.Select("friendships",
				fmt.Sprintf("or=(and(requester_id.eq.%s,addressee_id.eq.%s),and(requester_id.eq.%s,addressee_id.eq.%s))&select=status",
					currentUserID, uid, uid, currentUserID), true)
			if err == nil {
				var friendships []map[string]interface{}
				json.Unmarshal(fBody, &friendships)
				if len(friendships) > 0 {
					if s, ok := friendships[0]["status"].(string); ok {
						users[i]["_friendship_status"] = s
					}
				}
			}
		}
	}

	return c.JSON(fiber.Map{"users": users})
}

func (h *Handler) APIGetConversations(c *fiber.Ctx) error {
	userID := c.Query("user_id")

	body, err := h.db.Select("conversations",
		fmt.Sprintf("or=(user1_id.eq.%s,user2_id.eq.%s)&order=last_message_at.desc", userID, userID), true)
	if err != nil {
		return c.JSON(fiber.Map{"conversations": []interface{}{}, "error": err.Error()})
	}
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

		pBody, err := h.db.Select("user_profiles",
			fmt.Sprintf("id=eq.%s&select=id,username,nickname,avatar_url", otherUserID), false)
		if err != nil {
			continue
		}
		var profiles []map[string]interface{}
		json.Unmarshal(pBody, &profiles)
		if len(profiles) > 0 {
			otherUser := profiles[0]
			otherNickname := ""
			if on, ok := otherUser["nickname"].(string); ok {
				otherNickname = on
			}
			otherAvatar := ""
			if oa, ok := otherUser["avatar_url"].(string); ok {
				otherAvatar = oa
			}
			result = append(result, map[string]interface{}{
				"id":               convID,
				"other_user_id":    otherUserID,
				"other_username":   otherUser["username"],
				"other_nickname":   otherNickname,
				"other_avatar_url": otherAvatar,
				"last_message":     conv["last_message"],
				"unread_count":     0,
			})
		}
	}

	return c.JSON(fiber.Map{"conversations": result})
}

func (h *Handler) APIGetMessages(c *fiber.Ctx) error {
	conversationID := c.Query("conversation_id")

	body, err := h.db.Select("messages",
		fmt.Sprintf("conversation_id=eq.%s&order=created_at.asc&limit=100", conversationID), true)
	if err != nil {
		return c.JSON(fiber.Map{"messages": []interface{}{}, "error": err.Error()})
	}
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

	insertData := fmt.Sprintf(`{"conversation_id":"%s","sender_id":"%s","content":"%s"}`, req.ConversationID, senderID, req.Content)
	_, err := h.db.Insert("messages", []byte(insertData), true)
	if err != nil {
		return c.JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true})
}

// Pages protégées
func (h *Handler) Dashboard(c *fiber.Ctx) error {
	user := c.Locals("user").(*UserProfile)

	nickname := ""
	if user.Nickname != nil {
		nickname = *user.Nickname
	}
	displayName := user.Username
	if nickname != "" {
		displayName = nickname
	}

	ranks := h.loadRanks()
	nextXP, nextRank := getXPForNextRank(user.XP, ranks)
	xpPercent := calcXPPercentFromRanks(user.XP, ranks)

	statsBody, err := h.db.Select("user_stats",
		fmt.Sprintf("user_id=eq.%s&select=quizzes_played,best_score_ever,accuracy_rate", user.ID), true)
	quizzesPlayed := 0
	bestScore := 0
	accuracy := 0
	if err == nil {
		var stats []map[string]interface{}
		json.Unmarshal(statsBody, &stats)
		if len(stats) > 0 {
			quizzesPlayed = int(stats[0]["quizzes_played"].(float64))
			bestScore = int(stats[0]["best_score_ever"].(float64))
			acc, _ := stats[0]["accuracy_rate"].(float64)
			accuracy = int(acc)
		}
	}

	streakHTML := ""
	if user.CurrentStreak > 0 {
		streakHTML = fmt.Sprintf(`<div class="ds-card" style="border-color:#f59e0b"><div class="ds-icon">🔥</div><div class="ds-value">%d</div><div class="ds-label">Série (%d max)</div></div>`, user.CurrentStreak, user.LongestStreak)
	}

	return renderPage(c, "Dashboard", fmt.Sprintf(`
<div class="welcome">Bonjour %s%s 👋</div>
<div class="rank">%s • Niveau %d</div>

<div class="d-stats">
    <div class="ds-card ds-brand"><div class="ds-icon">⭐</div><div class="ds-value">%d</div><div class="ds-label">XP Total</div></div>
    <div class="ds-card ds-accent"><div class="ds-icon">🎮</div><div class="ds-value">%d</div><div class="ds-label">Quiz joués</div></div>
    <div class="ds-card ds-green"><div class="ds-icon">🎯</div><div class="ds-value">%d%%</div><div class="ds-label">Précision</div></div>
    <div class="ds-card ds-purple"><div class="ds-icon">🏆</div><div class="ds-value">%d</div><div class="ds-label">Meilleur score</div></div>
    %s
</div>

<div class="card" style="margin-bottom:20px">
    <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:8px">
        <span class="text-sm text-muted">Progression vers %s</span>
        <span class="text-sm">%d XP restants</span>
    </div>
    <div class="progress-bar"><div class="progress-fill" style="width:%d%%"></div></div>
</div>

<a href="/boite-a-idees" style="display:flex;align-items:center;gap:12px;background:#16213e;border:1px solid #2d2d44;border-radius:12px;padding:16px 20px;margin-bottom:20px;text-decoration:none;color:white;transition:border-color .2s" onmouseover="this.style.borderColor='#f59e0b'" onmouseout="this.style.borderColor='#2d2d44'">
    <span style="font-size:1.5rem">💡</span>
    <div><div style="font-weight:700;font-size:1rem">Boîte à Idées</div><div style="color:#94a3b8;font-size:.8rem">Propose et vote pour les améliorations</div></div>
    <span style="margin-left:auto;color:#f59e0b;font-size:1.2rem">→</span>
</a>
	`, displayName, challengeStatsBadge(user.ChallengesPlayed, user.ChallengesWon), user.Rank, user.Level,
		user.XP, quizzesPlayed, accuracy, bestScore, streakHTML,
		nextRank, nextXP, xpPercent))
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

	miniCupCard := `
<div class="quiz-card" style="border:2px solid #f59e0b;background:linear-gradient(135deg,rgba(245,158,11,.08),rgba(234,88,12,.08))">
    <div class="quiz-header">⚽ Mini Cup</div>
    <h3>Mini Cup Word Cup</h3>
    <p class="quiz-meta">Tirs aux buts • Solo vs IA ou 2 joueurs</p>
    <span style="background:rgba(245,158,11,.2);color:#f59e0b;padding:3px 10px;border-radius:12px;font-size:.75rem;font-weight:600">Nouveau</span>
    <div class="quiz-actions">
        <a href="/games/mini-cup" class="btn-primary" style="background:linear-gradient(135deg,#f59e0b,#ea580c)">Jouer ⚽</a>
    </div>
</div>`

	return renderPage(c, "Explorer", fmt.Sprintf(`
<h1 style="margin-bottom: 24px;">Explorer les quiz</h1>
<div class="quiz-grid">%s%s</div>
	`, miniCupCard, quizCards))
}

func (h *Handler) QuizDetail(c *fiber.Ctx) error {
	id := c.Params("id")
	user := c.Locals("user").(*UserProfile)

	quiz, err := h.db.GetQuiz(id)
	if err != nil {
		return renderPage(c, "Quiz introuvable", `<h1>Quiz introuvable</h1><a href="/explore" class="btn-primary">Retour</a>`)
	}

	title, _ := quiz["title"].(string)
	description := database.DBValue(quiz["description"])
	series, _ := quiz["series"].(string)
	category, _ := quiz["category"].(string)
	qCount := database.DBInt(quiz["question_count"])
	pCount := database.DBInt(quiz["play_count"])
	quizType, _ := quiz["quiz_type"].(string)
	likeCount := database.DBInt(quiz["like_count"])
	dislikeCount := database.DBInt(quiz["dislike_count"])

	creatorID := database.DBValue(quiz["creator_id"])
	creatorName := "Anonyme"
	creatorUsername := ""
	cBody, _ := h.db.Select("user_profiles",
		fmt.Sprintf("id=eq.%s&select=username,nickname,avatar_url", creatorID), false)
	var cRows []map[string]interface{}
	json.Unmarshal(cBody, &cRows)
	if len(cRows) > 0 {
		creatorUsername, _ = cRows[0]["username"].(string)
		if nn, ok := cRows[0]["nickname"].(string); ok && nn != "" {
			creatorName = nn
		} else {
			creatorName = creatorUsername
		}
	}

	myVote := ""
	if user != nil {
		vBody, _ := h.db.Select("quiz_votes",
			fmt.Sprintf("quiz_id=eq.%s&user_id=eq.%s&select=vote_type", id, user.ID), true)
		var vRows []map[string]interface{}
		json.Unmarshal(vBody, &vRows)
		if len(vRows) > 0 {
			myVote, _ = vRows[0]["vote_type"].(string)
		}
	}

	badge := ""
	if quizType == "official" {
		badge = `<span class="badge-official">Officiel</span>`
	} else if quizType == "challenge" {
		badge = `<span class="badge-challenge">Défi</span>`
	}

	likeClass := ""
	dislikeClass := ""
	if myVote == "like" {
		likeClass = " vote-active"
	}
	if myVote == "dislike" {
		dislikeClass = " vote-active"
	}

	descHTML := ""
	if description != "" {
		descHTML = `<div class="qd-description">` + description + `</div>`
	}

	creatorAvatarHTML := `<div style="width:32px;height:32px;border-radius:50%;background:#6366f1;display:flex;align-items:center;justify-content:center;font-weight:700;color:white;font-size:.8rem">` + strings.ToUpper(string(creatorName[0])) + `</div>`
	if creatorUsername != "" {
		creatorAvatarHTML = `<a href="/profile/` + creatorUsername + `" style="text-decoration:none"><div style="width:32px;height:32px;border-radius:50%;background:#6366f1;display:flex;align-items:center;justify-content:center;font-weight:700;color:white;font-size:.8rem">` + strings.ToUpper(string(creatorName[0])) + `</div></a>`
	}

	quizDetailHTML := `
<div class="quiz-detail">
    <div class="quiz-detail-header">
        <div>
            <h1>` + title + `</h1>
            ` + badge + `
            <p class="quiz-detail-meta">` + series + ` • ` + category + `</p>
        </div>
    </div>
    <div class="qd-author">
        ` + creatorAvatarHTML + `
        <div><div style="font-size:.8rem;color:#94a3b8">Créé par</div><a href="/profile/` + creatorUsername + `" style="color:#6366f1;font-weight:600;text-decoration:none">` + creatorName + `</a>` + challengeStatsBadgeMap(cRows[0]) + `</div>
    </div>
    ` + descHTML + `
    <div class="quiz-detail-stats">
        <div class="stat"><span class="stat-value">` + fmt.Sprintf("%d", qCount) + `</span><span class="stat-label">Questions</span></div>
        <div class="stat"><span class="stat-value">` + fmt.Sprintf("%d", pCount) + `</span><span class="stat-label">Parties</span></div>
    </div>
    <div class="qd-votes">
        <button class="vote-btn like-btn` + likeClass + `" onclick="voteQuiz('like')">👍 ` + fmt.Sprintf("%d", likeCount) + `</button>
        <button class="vote-btn dislike-btn` + dislikeClass + `" onclick="voteQuiz('dislike')">👎 ` + fmt.Sprintf("%d", dislikeCount) + `</button>
    </div>
    <div class="quiz-detail-actions">
        <a href="/quiz/` + id + `/play" class="btn-primary">🎮 Jouer maintenant</a>
        <a href="/challenges/create/` + id + `" class="btn-challenge">⚔️ Défier un ami</a>
        <a href="/leaderboard/quiz/` + id + `" class="btn-secondary">🏆 Classement</a>
    </div>
</div>

<style>
.quiz-detail{background:#16213e;border:1px solid #2d2d44;border-radius:12px;padding:32px}
.quiz-detail-header{margin-bottom:16px}
.quiz-detail-header h1{font-size:2rem;margin-bottom:8px}
.quiz-detail-meta{color:#94a3b8}
.qd-author{display:flex;align-items:center;gap:10px;margin-bottom:16px;padding:12px;background:#0f172a;border-radius:8px}
.qd-description{color:#94a3b8;margin-bottom:20px;line-height:1.6;padding:12px;background:#0f172a;border-radius:8px}
.quiz-detail-stats{display:flex;gap:32px;margin-bottom:16px;padding:16px;background:#1a1a2e;border-radius:8px}
.stat{display:flex;flex-direction:column;align-items:center}
.stat-value{font-size:1.5rem;font-weight:bold;color:#6366f1}
.stat-label{font-size:.8rem;color:#94a3b8}
.qd-votes{display:flex;gap:12px;margin-bottom:20px}
.vote-btn{padding:8px 20px;border-radius:8px;border:2px solid #2d2d44;background:#0f172a;color:white;font-weight:600;cursor:pointer;transition:all .2s;font-size:.9rem}
.vote-btn:hover{border-color:#6366f1}
.vote-active.like-btn{border-color:#22c55e;background:rgba(34,197,94,.15);color:#22c55e}
.vote-active.dislike-btn{border-color:#ef4444;background:rgba(239,68,68,.15);color:#ef4444}
.quiz-detail-actions{display:flex;gap:12px;flex-wrap:wrap}
.btn-challenge{display:inline-block;padding:10px 20px;background:linear-gradient(135deg,#a855f7,#ec4899);color:white;text-decoration:none;border-radius:8px;font-weight:600;transition:opacity .2s}
.btn-challenge:hover{opacity:.85}
.badge-challenge{display:inline-block;background:linear-gradient(135deg,#a855f7,#ec4899);color:white;padding:2px 8px;border-radius:4px;font-size:.75rem;font-weight:600}
</style>
<script>
function voteQuiz(type){
    fetch('/api/quiz/vote',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({quiz_id:'` + id + `',vote_type:type})}).then(function(r){return r.json()}).then(function(d){
        if(d.success){
            var lb=document.querySelector('.like-btn');
            var db=document.querySelector('.dislike-btn');
            if(lb)lb.innerHTML='👍 '+d.like_count;
            if(db)db.innerHTML='👎 '+d.dislike_count;
            lb.classList.remove('vote-active');
            db.classList.remove('vote-active');
            if(type==='like')lb.classList.add('vote-active');
            if(type==='dislike')db.classList.add('vote-active');
            if(d.action==='removed'){lb.classList.remove('vote-active');db.classList.remove('vote-active');}
        }
    });
}
</script>`

	return renderPage(c, title, quizDetailHTML)
}

func (h *Handler) QuizLeaderboard(c *fiber.Ctx) error {
	id := c.Params("id")
	log.Printf("[QuizLeaderboard] Loading leaderboard for quiz_id=%s", id)

	quizBody, err := h.db.Select("quizzes",
		fmt.Sprintf("id=eq.%s&select=title", id), true)
	quizzes := []map[string]interface{}{}
	if err == nil {
		json.Unmarshal(quizBody, &quizzes)
	} else {
		log.Printf("[QuizLeaderboard] Select quizzes error: %v", err)
	}

	quizTitle := "Quiz"
	if len(quizzes) > 0 {
		quizTitle, _ = quizzes[0]["title"].(string)
	}

	query := fmt.Sprintf("quiz_id=eq.%s&order=xp_earned.desc,time_taken_ms.asc&limit=50&select=*,user:user_id(username,nickname,avatar_url,rank,challenges_played,challenges_won)", id)
	log.Printf("[QuizLeaderboard] Query: %s", query)
	body, err := h.db.Select("quiz_leaderboard", query, true)
	if err != nil {
		log.Printf("[QuizLeaderboard] Select quiz_leaderboard error: %v", err)
	} else {
		log.Printf("[QuizLeaderboard] Raw response (first 500 chars): %s", string(body)[:min(len(body), 500)])
	}
	entries := []map[string]interface{}{}
	if err == nil {
		json.Unmarshal(body, &entries)
	}
	log.Printf("[QuizLeaderboard] Parsed %d entries", len(entries))

	rows := ""
	for i, s := range entries {
		xp, _ := s["xp_earned"].(float64)
		user, _ := s["user"].(map[string]interface{})
		username := "?"
		displayName := "?"
		rank := "F"
		if user != nil {
			username, _ = user["username"].(string)
			rank, _ = user["rank"].(string)
			nickname, _ := user["nickname"].(string)
			displayName = username
			if nickname != "" {
				displayName = nickname
			}
		}

		medal := fmt.Sprintf("#%d", i+1)
		if i == 0 {
			medal = "🥇"
		} else if i == 1 {
			medal = "🥈"
		} else if i == 2 {
			medal = "🥉"
		}

		rows += fmt.Sprintf(`<tr><td class="rank-cell">%s</td><td><span class="username">%s</span>%s<span class="user-rank">%s</span></td><td class="xp-cell">%.0f XP</td></tr>`, medal, displayName, challengeStatsBadgeMap(user), rank, xp)
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
	challengeID := c.Query("challenge", "")

	if challengeID != "" {
		user := c.Locals("user").(*UserProfile)
		scoreBody, _ := h.db.Select("challenge_scores",
			fmt.Sprintf("session_id=eq.%s&user_id=eq.%s&select=id", challengeID, user.ID), true)
		var scoreRows []map[string]interface{}
		json.Unmarshal(scoreBody, &scoreRows)
		if len(scoreRows) > 0 {
			return c.Redirect("/challenges/" + challengeID + "/leaderboard")
		}
	}

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

	// Randomize options for this session
	sessionID := fmt.Sprintf("%s-%d", id, time.Now().UnixNano())
	questions = h.randomizeQuizQuestions(questions, sessionID)

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
            <div id="answers-container" class="answers-container"></div>
            <div id="matching-container" class="matching-container" style="display:none"></div>
            <div id="fillin-container" class="fillin-container" style="display:none"></div>
            <div id="matching-warning" class="matching-warning" style="display:none"></div>
            <div id="matching-submit-row" style="display:none;text-align:center;margin-top:16px">
                <button class="btn-primary" onclick="submitMatching()">Valider</button>
            </div>
            <div id="fillin-submit-row" style="display:none;text-align:center;margin-top:16px">
                <button class="btn-primary" onclick="submitFillIn()">Valider</button>
            </div>
        </div>
    </div>
    <div id="results" style="display:none" class="results">
        <h2>Quiz terminé !</h2>
        <div id="score-display"></div>
        <div id="xp-display"></div>
        <div id="results-actions" style="margin-top:24px"></div>
    </div>
</div>

<script>
(function() {
    var quizId = '%s';
    var challengeSessionId = '%s';
    var questions = %s;
    var currentQuestion = 0;
    var score = 0;
    var totalQuestions = questions.length;
    var answers = [];
    var questionStartTime = Date.now();
    var timerInterval = null;
    var timeLeft = 30;
    var submitting = false;
    var questionAnswered = false;

    // Matching state
    var matchingState = { startX: null, connections: {}, connecting: false };

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
        if (currentQuestion >= totalQuestions) { showResults(); return; }
        questionAnswered = false;
        questionStartTime = Date.now();
        var q = questions[currentQuestion];

        document.getElementById('question-text').textContent = q.question_text;
        document.getElementById('question-text').style.display = '';
        document.getElementById('question-counter').textContent = 'Question ' + (currentQuestion + 1) + '/' + totalQuestions;
        document.getElementById('progress-fill').style.width = ((currentQuestion + 1) / totalQuestions * 100) + '%%';

        // Reset
        document.getElementById('answers-container').style.display = 'none';
        document.getElementById('matching-container').style.display = 'none';
        document.getElementById('fillin-container').style.display = 'none';
        document.getElementById('matching-warning').style.display = 'none';
        document.getElementById('matching-submit-row').style.display = 'none';
        document.getElementById('fillin-submit-row').style.display = 'none';
        document.getElementById('answers-container').innerHTML = '';
        document.getElementById('matching-container').innerHTML = '';
        document.getElementById('fillin-container').innerHTML = '';

        // Timer
        timeLeft = q.time_limit_seconds || 30;
        document.getElementById('timer').textContent = timeLeft + 's';
        clearInterval(timerInterval);
        timerInterval = setInterval(function() {
            timeLeft--;
            document.getElementById('timer').textContent = timeLeft + 's';
            if (timeLeft <= 0 && !questionAnswered) {
                clearInterval(timerInterval);
                questionAnswered = true;
                currentQuestion++;
                loadQuestion();
            }
        }, 1000);

        var qType = q.question_type;

        if (qType === 'matching') {
            renderMatching(q);
        } else if (qType === 'fill_in') {
            renderFillIn(q);
        } else {
            renderQCM(q);
        }
    }

    // =================== QCM ===================
    function renderQCM(q) {
        var container = document.getElementById('answers-container');
        container.style.display = 'flex';
        container.innerHTML = '';
        container.className = 'answers-container';

        var choices = [];
        if (q.presented_options && q.presented_options.length > 0) {
            choices = q.presented_options.map(function(text) {
                var matched = (q.answers || []).find(function(a) { return a.answer_text === text; });
                return { answer_text: text, is_correct: matched ? matched.is_correct : false, id: matched ? matched.id : '' };
            });
        } else {
            choices = (q.answers || []).slice().sort(function(a, b) { return (a.order_index || 0) - (b.order_index || 0); });
        }

        choices.forEach(function(answer) {
            var btn = document.createElement('button');
            btn.className = 'answer-btn';
            btn.textContent = answer.answer_text;
            btn.onclick = function() { selectAnswer(answer, btn); };
            container.appendChild(btn);
        });
    }

    function selectAnswer(answer, btn) {
        if (questionAnswered) return;
        questionAnswered = true;
        clearInterval(timerInterval);
        var q = questions[currentQuestion];
        if (answer.is_correct) {
            btn.classList.add('correct');
            score++;
        } else {
            btn.classList.add('wrong');
        }
        answers.push({question_id: q.id, answer_id: answer.id || '', is_correct: answer.is_correct, time_ms: Date.now() - questionStartTime, type: 'qcm'});
        setTimeout(function() { currentQuestion++; loadQuestion(); }, 500);
    }

    // =================== MATCHING ===================
    function renderMatching(q) {
        var container = document.getElementById('matching-container');
        container.style.display = 'block';
        document.getElementById('matching-submit-row').style.display = 'block';

        matchingState = { connections: {}, pendingXid: null, pendingNumber: 0 };
        var options = q.options;
        if (typeof options === 'string') { try { options = JSON.parse(options); } catch(e) { options = {}; } }
        options = options || {};
        var pairs = options.pairs || [];

        if (pairs.length === 0) {
            container.innerHTML = '<p style="color:#ef4444">Question matching mal formée</p>';
            return;
        }

        var xItems = [];
        for (var i = 0; i < pairs.length; i++) {
            xItems.push({ id: pairs[i].id, text: pairs[i].x });
        }
        var yValues = q.presented_y || [];
        if (yValues.length === 0) {
            for (var i = 0; i < pairs.length; i++) { yValues.push(pairs[i].y); }
        }

        container.innerHTML = '';

        var wrap = document.createElement('div');
        wrap.className = 'matching-wrap';

        var grid = document.createElement('div');
        grid.className = 'matching-grid';
        grid.style.cssText = 'display:grid;grid-template-columns:1fr 1fr;gap:12px 40px;position:relative';

        var hdrL = document.createElement('div');
        hdrL.className = 'matching-hdr';
        hdrL.textContent = 'Éléments';
        var hdrR = document.createElement('div');
        hdrR.className = 'matching-hdr';
        hdrR.textContent = 'Cibles';
        grid.appendChild(hdrL);
        grid.appendChild(hdrR);

        for (var i = 0; i < xItems.length; i++) {
            var xEl = document.createElement('div');
            xEl.className = 'matching-item matching-x';
            xEl.dataset.xid = xItems[i].id;
            xEl.style.position = 'relative';
            xEl.innerHTML = '<span class="matching-num" id="xnum-' + xItems[i].id + '"></span>';
            xEl.appendChild(document.createTextNode(xItems[i].text));
            (function(xid, el) {
                el.onclick = function() { onMatchingXClick(xid, el); };
            })(xItems[i].id, xEl);
            grid.appendChild(xEl);

            var yEl = document.createElement('div');
            yEl.className = 'matching-item matching-y';
            yEl.dataset.yidx = i;
            yEl.dataset.yval = yValues[i];
            yEl.style.position = 'relative';
            yEl.innerHTML = '<span class="matching-num" id="ynum-' + i + '"></span>';
            yEl.appendChild(document.createTextNode(yValues[i]));
            (function(yidx, yval, el) {
                el.onclick = function() { onMatchingYClick(yidx, yval, el); };
            })(i, yValues[i], yEl);
            grid.appendChild(yEl);
        }

        wrap.appendChild(grid);
        container.appendChild(wrap);

        window._matchingPairs = pairs;
        window._matchingXEls = grid.querySelectorAll('.matching-x');
        window._matchingYEls = grid.querySelectorAll('.matching-y');
    }

    function onMatchingXClick(xid, el) {
        if (questionAnswered) return;
        if (matchingState.connections[xid] !== undefined) {
            var oldY = matchingState.connections[xid];
            delete matchingState.connections[xid];
            setNum('xnum-' + xid, '');
            setNum('ynum-' + oldY, '');
            if (matchingState.pendingXid === xid) {
                matchingState.pendingXid = null;
                matchingState.pendingNumber = 0;
            }
            refreshMatchingDisplay();
            return;
        }
        if (matchingState.pendingXid) {
            var prev = matchingState.pendingXid;
            setNum('xnum-' + prev, '');
            matchingState.pendingXid = null;
            matchingState.pendingNumber = 0;
        }
        var num = Object.keys(matchingState.connections).length + 1;
        matchingState.pendingXid = xid;
        matchingState.pendingNumber = num;
        setNum('xnum-' + xid, num);
        refreshMatchingDisplay();
    }

    function onMatchingYClick(yidx, yval, el) {
        if (questionAnswered) return;
        for (var xid in matchingState.connections) {
            if (matchingState.connections[xid] === yidx) {
                delete matchingState.connections[xid];
                setNum('xnum-' + xid, '');
                setNum('ynum-' + yidx, '');
                refreshMatchingDisplay();
                return;
            }
        }
        if (matchingState.pendingXid) {
            var pxid = matchingState.pendingXid;
            var num = matchingState.pendingNumber;
            for (var xid2 in matchingState.connections) {
                if (matchingState.connections[xid2] === yidx) {
                    delete matchingState.connections[xid2];
                    setNum('xnum-' + xid2, '');
                }
            }
            matchingState.connections[pxid] = yidx;
            setNum('xnum-' + pxid, num);
            setNum('ynum-' + yidx, num);
            matchingState.pendingXid = null;
            matchingState.pendingNumber = 0;
            refreshMatchingDisplay();
        }
    }

    function setNum(id, val) {
        var el = document.getElementById(id);
        if (el) el.textContent = val;
    }

    function refreshMatchingDisplay() {
        document.querySelectorAll('.matching-x').forEach(function(el) {
            el.classList.remove('connected', 'selected');
        });
        document.querySelectorAll('.matching-y').forEach(function(el) {
            el.classList.remove('connected');
        });
        for (var xid in matchingState.connections) {
            var yIdx = matchingState.connections[xid];
            var xEl = document.querySelector('.matching-x[data-xid="' + xid + '"]');
            var yEl = document.querySelector('.matching-y[data-yidx="' + yIdx + '"]');
            if (xEl) xEl.classList.add('connected');
            if (yEl) yEl.classList.add('connected');
        }
        if (matchingState.pendingXid) {
            var pxEl = document.querySelector('.matching-x[data-xid="' + matchingState.pendingXid + '"]');
            if (pxEl) pxEl.classList.add('selected');
        }
        var total = (window._matchingPairs || []).length;
        var connected = Object.keys(matchingState.connections).length;
        var warn = document.getElementById('matching-warning');
        if (connected < total) {
            warn.style.display = 'block';
            warn.textContent = (total - connected) + ' connexion(s) manquante(s)';
        } else {
            warn.style.display = 'none';
        }
    }

    window.submitMatching = function() {
        if (questionAnswered) return;
        var pairs = window._matchingPairs || [];
        var total = pairs.length;
        var connected = Object.keys(matchingState.connections).length;
        if (connected < total) {
            document.getElementById('matching-warning').style.display = 'block';
            document.getElementById('matching-warning').textContent = 'Connecte toutes les paires !';
            return;
        }
        questionAnswered = true;
        clearInterval(timerInterval);
        var q = questions[currentQuestion];

        var matches = [];
        var correctCount = 0;
        for (var xid in matchingState.connections) {
            var yIdx = matchingState.connections[xid];
            var yEl = document.querySelector('.matching-y[data-yidx="' + yIdx + '"]');
            var yVal = yEl ? yEl.dataset.yval : '';
            var pair = pairs.find(function(p) { return p.id === xid; });
            var isCorrect = pair && pair.y === yVal;
            matches.push({x_id: xid, y_id: yVal});
            if (isCorrect) correctCount++;
        }

        var isFullyCorrect = correctCount === total;
        if (isFullyCorrect) score++;

        answers.push({
            question_id: q.id,
            answer_id: '',
            is_correct: isFullyCorrect,
            time_ms: Date.now() - questionStartTime,
            type: 'matching',
            matches: matches,
            correct_count: correctCount,
            total_pairs: total
        });

        // Show correct/wrong colors
        for (var xid in matchingState.connections) {
            var yIdx = matchingState.connections[xid];
            var yEl = document.querySelector('.matching-y[data-yidx="' + yIdx + '"]');
            var pair = pairs.find(function(p) { return p.id === xid; });
            var isCorrect = pair && yEl && pair.y === yEl.dataset.yval;
            var xEl = document.querySelector('.matching-x[data-xid="' + xid + '"]');
            if (xEl) { xEl.classList.remove('connected'); xEl.classList.add(isCorrect ? 'correct' : 'wrong'); }
            if (yEl) { yEl.classList.remove('connected'); yEl.classList.add(isCorrect ? 'correct' : 'wrong'); }
        }

        setTimeout(function() { currentQuestion++; loadQuestion(); }, 1200);
    }

    // =================== FILL IN ===================
    function renderFillIn(q) {
        var container = document.getElementById('fillin-container');
        container.style.display = 'block';
        container.innerHTML = '';
        document.getElementById('fillin-submit-row').style.display = 'block';

        // Hide question text heading — fillin renders inline
        document.getElementById('question-text').style.display = 'none';

        var rawOpts = q.options;
        var options = typeof rawOpts === 'string' ? JSON.parse(rawOpts) : (rawOpts || {});
        var template = options.template || q.question_text || '';
        var blanks = options.blanks || [];

        if (!template || blanks.length === 0) {
            container.innerHTML = '<p style="color:#ef4444">Question fill_in mal formée</p>';
            return;
        }

        var el = document.createElement('div');
        el.className = 'fillin-text';
        var parts = template.split(/\{(\d+)\}/);

        var foundBlank = false;
        for (var i = 0; i < parts.length; i++) {
            var part = parts[i];
            var blankIdx = parseInt(part);
            if (!isNaN(blankIdx) && i % 2 === 1) {
                foundBlank = true;
                var blank = null;
                for (var bi = 0; bi < blanks.length; bi++) {
                    if (String(blanks[bi].id) === String(blankIdx)) { blank = blanks[bi]; break; }
                }
                renderFillInBlank(el, blank, blankIdx);
            } else if (part) {
                var span = document.createElement('span');
                span.textContent = part;
                el.appendChild(span);
            }
        }

        // Fallback: if no {N} placeholders found but blanks exist
        if (!foundBlank && blanks.length > 0) {
            for (var bi = 0; bi < blanks.length; bi++) {
                renderFillInBlank(el, blanks[bi], blanks[bi].id);
            }
        }

        container.appendChild(el);

        // Connect all chars across blanks for keyboard navigation
        var allChars = container.querySelectorAll('.fillin-char');
        for (var idx = 0; idx < allChars.length; idx++) {
            (function(i, inp) {
                inp.addEventListener('input', function(e) {
                    if (e.target.value && i + 1 < allChars.length) {
                        allChars[i + 1].focus();
                    }
                });
                inp.addEventListener('keydown', function(e) {
                    if (e.key === 'Backspace' && !e.target.value && i > 0) {
                        e.preventDefault();
                        allChars[i - 1].value = '';
                        allChars[i - 1].focus();
                    }
                    if ((e.key === 'Enter' || e.key === 'Tab') && i + 1 < allChars.length) {
                        e.preventDefault();
                        allChars[i + 1].focus();
                    }
                    if (e.key === 'Tab' && e.shiftKey && i > 0) {
                        e.preventDefault();
                        allChars[i - 1].focus();
                    }
                });
            })(idx, allChars[idx]);
        }

        if (allChars.length > 0) allChars[0].focus();
        window._fillInBlanks = blanks;
    }

    function renderFillInBlank(parentEl, blank, blankIdx) {
        var span = document.createElement('span');
        span.className = 'fillin-blank';
        span.dataset.blankId = blankIdx;
        var charCount = blank ? (blank.char_count || (blank.answer ? blank.answer.length : 4)) : 4;
        for (var c = 0; c < charCount; c++) {
            var inp = document.createElement('input');
            inp.type = 'text';
            inp.className = 'fillin-char';
            inp.maxLength = 1;
            inp.dataset.blank = blankIdx;
            inp.dataset.idx = c;
            inp.autocomplete = 'off';
            inp.autocapitalize = 'off';
            inp.spellcheck = false;
            inp.placeholder = '_';
            span.appendChild(inp);
        }
        parentEl.appendChild(span);
    }

    window.submitFillIn = function() {
        if (questionAnswered) return;
        questionAnswered = true;
        clearInterval(timerInterval);
        var q = questions[currentQuestion];

        var blanks = window._fillInBlanks || [];
        var container = document.getElementById('fillin-container');
        var blankGroups = container.querySelectorAll('.fillin-blank');
        var playerBlanks = [];
        var allCorrect = true;

        blankGroups.forEach(function(group) {
            var blankId = parseInt(group.dataset.blankId);
            var chars = group.querySelectorAll('.fillin-char');
            var value = '';
            chars.forEach(function(c) { value += c.value || ''; });
            var blank = blanks.find(function(b) { return b.id === blankId; });
            var correctVal = blank ? blank.answer : '';
            var acceptNoAccents = blank ? blank.accept_without_accents : false;

            var pv = value.toLowerCase();
            var cv = correctVal.toLowerCase();
            if (acceptNoAccents) { pv = removeAccentsClient(pv); cv = removeAccentsClient(cv); }
            var isCorrect = pv === cv;
            if (!isCorrect) allCorrect = false;

            playerBlanks.push({id: blankId, value: value});

            chars.forEach(function(c) {
                c.style.borderBottomColor = isCorrect ? '#22c55e' : '#ef4444';
                c.style.backgroundColor = isCorrect ? 'rgba(34,197,94,0.1)' : 'rgba(239,68,68,0.1)';
                c.disabled = true;
            });
        });

        if (allCorrect) score++;

        answers.push({
            question_id: q.id,
            answer_id: '',
            is_correct: allCorrect,
            time_ms: Date.now() - questionStartTime,
            type: 'fill_in',
            blanks: playerBlanks
        });

        setTimeout(function() { currentQuestion++; loadQuestion(); }, 1000);
    }

    function removeAccentsClient(s) {
        var map = {'é':'e','è':'e','ê':'e','ë':'e','à':'a','â':'a','ä':'a','ù':'u','û':'u','ü':'u','ô':'o','ö':'o','î':'i','ï':'i','ç':'c','ñ':'n'};
        return s.split('').map(function(c) { return map[c] || c; }).join('');
    }

    // =================== RESULTS ===================
    function showResults() {
        if (submitting) return;
        submitting = true;
        clearInterval(timerInterval);
        document.getElementById('quiz-content').style.display = 'none';
        document.getElementById('results').style.display = 'block';
        document.getElementById('score-display').innerHTML = '<h3>' + score + ' bonnes réponses sur ' + totalQuestions + '</h3>';
        document.getElementById('xp-display').innerHTML = '<p>Calcul XP en cours...</p>';
        document.getElementById('results-actions').innerHTML =
            '<a href="/explore" class="btn-ghost">Retour</a>';

        var totalTimeMs = 0;
        for (var i = 0; i < answers.length; i++) { totalTimeMs += answers[i].time_ms; }
        var body = {quiz_id: quizId, total: totalQuestions, answers: answers, total_time_ms: totalTimeMs};
        if (challengeSessionId) { body.challenge_session_id = challengeSessionId; }

        fetch('/api/quiz/submit', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(body)
        }).then(function(r) {
            return r.json();
        }).then(function(d) {
            document.getElementById('xp-display').innerHTML = '<p>+' + d.xp_earned + ' XP gagnés</p>';
            var lbLink = challengeSessionId ? '/challenges/' + challengeSessionId + '/leaderboard' : '/leaderboard/quiz/' + quizId;
            document.getElementById('results-actions').innerHTML =
                '<a href="' + lbLink + '" class="btn-primary" style="margin-right:8px">🏆 Voir mon classement</a>' +
                '<a href="/explore" class="btn-ghost">Retour</a>';
        }).catch(function(e) {
            document.getElementById('xp-display').innerHTML = '<p style="color:#ef4444">Erreur serveur</p>';
        });
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
.answers-container { display: flex; flex-direction: column; gap: 12px; }
.answer-btn { padding: 16px; background: #1a1a2e; border: 2px solid #2d2d44; border-radius: 8px; color: white; font-size: 1rem; cursor: pointer; transition: all 0.2s; text-align: left; }
.answer-btn:hover { border-color: #6366f1; background: rgba(99,102,241,0.1); }
.answer-btn.correct { border-color: #22c55e; background: rgba(34,197,94,0.1); }
.answer-btn.wrong { border-color: #ef4444; background: rgba(239,68,68,0.1); }
.results { text-align: center; padding: 60px 20px; }
.results h2 { font-size: 2rem; margin-bottom: 24px; }

.matching-wrap { }
.matching-hdr { font-weight: 700; color: #6366f1; font-size: .85rem; text-align: center; padding: 8px; text-transform: uppercase; letter-spacing: 1px; }
.matching-item { padding: 14px 12px; background: #1a1a2e; border: 2px solid #2d2d44; border-radius: 8px; cursor: pointer; text-align: center; font-size: .95rem; transition: all 0.2s; user-select: none; }
.matching-x:hover, .matching-x.selected { border-color: #6366f1; background: rgba(99,102,241,0.15); }
.matching-y:hover { border-color: #a855f7; background: rgba(168,85,247,0.1); }
.matching-x.connected { border-color: #6366f1; background: rgba(99,102,241,0.2); }
.matching-y.connected { border-color: #a855f7; background: rgba(168,85,247,0.15); }
.matching-x.correct, .matching-y.correct { border-color: #22c55e; background: rgba(34,197,94,0.15); }
.matching-x.wrong, .matching-y.wrong { border-color: #ef4444; background: rgba(239,68,68,0.15); }
.matching-num { position: absolute;  top: -8px; right: -8px; width: 26px; height: 26px; background: #6366f1; color: white; font-size: .75rem; font-weight: 700; border-radius: 50%; text-align: center; line-height: 17px; box-shadow: 0 2px 6px rgba(99,102,241,0.5); }
.matching-warning { color: #fbbf24; text-align: center; font-size: .85rem; margin-top: 12px; }

.fillin-text { font-size: 1.2rem; line-height: 3; color: white; padding: 12px 0; }
.fillin-blank { display: inline-flex; gap: 3px; vertical-align: bottom; margin: 0 6px; position: relative; }
.fillin-char { width: 32px; height: 44px; background: transparent; border: none; border-bottom: 3px solid #6366f1; border-radius: 3px 3px 0 0; color: white; font-size: 1.2rem; font-weight: 700; text-align: center; outline: none; padding: 0; transition: all 0.2s; font-family: inherit; caret-color: #a855f7; }
.fillin-char::placeholder { color: #6366f1; font-weight: 400; opacity: 0.7; }
.fillin-char:focus { border-bottom-color: #a855f7; background: rgba(99,102,241,0.1); }
.fillin-char:disabled { opacity: 0.8; }
</style>
	`, len(questions), id, challengeID, string(questionsJSON)))
}

func (h *Handler) QuizSubmit(c *fiber.Ctx) error {
	user := c.Locals("user").(*UserProfile)
	log.Printf("[QuizSubmit] Received from user %s", user.ID)

	type MatchPayload struct {
		XID string `json:"x_id"`
		YID string `json:"y_id"`
	}
	type BlankPayload struct {
		ID    int    `json:"id"`
		Value string `json:"value"`
	}
	type AnswerPayload struct {
		QuestionID   string         `json:"question_id"`
		AnswerID     string         `json:"answer_id"`
		IsCorrect    bool           `json:"is_correct"`
		TimeMs       int            `json:"time_ms"`
		Type         string         `json:"type"`
		Matches      []MatchPayload `json:"matches"`
		Blanks       []BlankPayload `json:"blanks"`
		CorrectCount int            `json:"correct_count"`
		TotalPairs   int            `json:"total_pairs"`
	}
	type Request struct {
		QuizID             string          `json:"quiz_id"`
		Total              int             `json:"total"`
		Answers            []AnswerPayload `json:"answers"`
		ChallengeSessionID string          `json:"challenge_session_id"`
		TotalTimeMs        int             `json:"total_time_ms"`
	}
	var req Request
	if err := c.BodyParser(&req); err != nil {
		log.Printf("[QuizSubmit] BodyParser error: %v", err)
		return c.Status(400).JSON(fiber.Map{"error": "Données invalides"})
	}
	log.Printf("[QuizSubmit] quiz_id=%s total=%d answers=%d", req.QuizID, req.Total, len(req.Answers))

	if req.QuizID == "" {
		log.Printf("[QuizSubmit] Empty quiz_id, aborting")
		return c.Status(400).JSON(fiber.Map{"error": "Quiz ID requis"})
	}

	// Fetch question types and xp_reward for this quiz
	qTypeBody, qTypeErr := h.db.Select("questions",
		fmt.Sprintf("quiz_id=eq.%s&select=id,question_type,xp_reward", req.QuizID), true)
	qTypeMap := map[string]string{}
	xpRewardMap := map[string]int{}
	if qTypeErr == nil {
		var qRows []map[string]interface{}
		json.Unmarshal(qTypeBody, &qRows)
		for _, r := range qRows {
			id, _ := r["id"].(string)
			qt, _ := r["question_type"].(string)
			qTypeMap[id] = qt
			if xpr, ok := r["xp_reward"].(float64); ok && xpr > 0 {
				xpRewardMap[id] = int(xpr)
			}
		}
	}
	log.Printf("[QuizSubmit] Loaded %d question types", len(qTypeMap))

	xpBareme := map[string]int{
		"text": 1, "true_false": 1, "image": 2, "image_shadow": 1,
		"gif": 2, "audio": 3, "character_guess": 5, "impostor": 5,
		"fill_in": 3, "matching": 4,
	}

	correctCount := 0
	totalXP := 0
	now := time.Now().UTC().Format(time.RFC3339)

	// Attempt number for this quiz
	attemptNum := 1
	attBody, attErr := h.db.Select("user_quiz_attempts",
		fmt.Sprintf("user_id=eq.%s&quiz_id=eq.%s&order=attempt_number.desc&limit=1&select=attempt_number", user.ID, req.QuizID), true)
	if attErr == nil {
		var attRows []map[string]interface{}
		json.Unmarshal(attBody, &attRows)
		if len(attRows) > 0 {
			if an, ok := attRows[0]["attempt_number"].(float64); ok {
				attemptNum = int(an) + 1
			}
		}
	}
	log.Printf("[QuizSubmit] attempt_number=%d", attemptNum)

	for _, ans := range req.Answers {
		qType := qTypeMap[ans.QuestionID]
		baseXP := xpRewardMap[ans.QuestionID]
		if baseXP == 0 {
			baseXP = xpBareme[qType]
		}
		if baseXP == 0 {
			baseXP = 1
		}

		// Server-side validation for matching and fill_in
		isCorrect := ans.IsCorrect
		if qType == "matching" && len(ans.Matches) > 0 {
			// Fetch question pairs from DB
			qBody, qErr := h.db.Select("questions",
				fmt.Sprintf("id=eq.%s&select=options", ans.QuestionID), true)
			if qErr == nil {
				var qRows []map[string]interface{}
				json.Unmarshal(qBody, &qRows)
				if len(qRows) > 0 {
					pairs := getMatchingPairs(qRows[0])
					matchesGeneric := make([]map[string]interface{}, len(ans.Matches))
					for mi, m := range ans.Matches {
						matchesGeneric[mi] = map[string]interface{}{"x_id": m.XID, "y_id": m.YID}
					}
					correctCountServer, totalPairs := validateMatchingAnswer(pairs, matchesGeneric)
					isCorrect = correctCountServer == totalPairs && totalPairs > 0
					// Store partial score info for XP
					if !isCorrect && correctCountServer > 0 {
						// Partial XP: fraction of base XP
						baseXP = int(float64(baseXP) * float64(correctCountServer) / float64(totalPairs))
						if baseXP < 1 {
							baseXP = 0
						}
					}
				}
			}
		} else if qType == "fill_in" && len(ans.Blanks) > 0 {
			// Fetch question blanks from DB
			qBody, qErr := h.db.Select("questions",
				fmt.Sprintf("id=eq.%s&select=options", ans.QuestionID), true)
			if qErr == nil {
				var qRows []map[string]interface{}
				json.Unmarshal(qBody, &qRows)
				if len(qRows) > 0 {
					blanks := getFillInBlanks(qRows[0])
					// Convert BlankPayload to generic format
					playerBlanks := make([]map[string]interface{}, len(ans.Blanks))
					for i, b := range ans.Blanks {
						playerBlanks[i] = map[string]interface{}{
							"id":    float64(b.ID),
							"value": b.Value,
						}
					}
					correctBlanks, totalBlanks := validateFillInAnswer(blanks, playerBlanks)
					isCorrect = correctBlanks == totalBlanks && totalBlanks > 0
				}
			}
		}

		if isCorrect {
			correctCount++
		}

		// Check if already answered correctly before
		prevBody, prevErr := h.db.Select("user_question_attempts",
			fmt.Sprintf("user_id=eq.%s&quiz_id=eq.%s&question_id=eq.%s&is_correct=eq.true&select=id&limit=1", user.ID, req.QuizID, ans.QuestionID), true)
		alreadyCorrect := false
		if prevErr == nil {
			var prevRows []map[string]interface{}
			json.Unmarshal(prevBody, &prevRows)
			if len(prevRows) > 0 {
				alreadyCorrect = true
			}
		}

		var questionXP int
		if alreadyCorrect {
			questionXP = 0
			log.Printf("[QuizSubmit] Question %s: already correct, 0 XP", ans.QuestionID[:8])
		} else if !isCorrect {
			questionXP = 0
		} else {
			// Count previous wrong attempts
			wrongBody, wrongErr := h.db.Select("user_question_attempts",
				fmt.Sprintf("user_id=eq.%s&quiz_id=eq.%s&question_id=eq.%s&is_correct=eq.false&select=id", user.ID, req.QuizID, ans.QuestionID), true)
			wrongCount := 0
			if wrongErr == nil {
				var wrongRows []map[string]interface{}
				json.Unmarshal(wrongBody, &wrongRows)
				wrongCount = len(wrongRows)
			}

			divisor := 1
			for i := 0; i < wrongCount; i++ {
				divisor *= 2
			}
			questionXP = baseXP / divisor
			if questionXP < 1 {
				questionXP = 0
			}
			log.Printf("[QuizSubmit] Question %s: baseXP=%d wrongCount=%d divisor=%d -> %d XP", ans.QuestionID[:8], baseXP, wrongCount, divisor, questionXP)
		}

		totalXP += questionXP

		// Record question attempt
		qaData, _ := json.Marshal(map[string]interface{}{
			"user_id":        user.ID,
			"quiz_id":        req.QuizID,
			"question_id":    ans.QuestionID,
			"attempt_number": attemptNum,
			"is_correct":     isCorrect,
			"xp_earned":      questionXP,
		})
		h.db.Insert("user_question_attempts", qaData, true)
	}

	log.Printf("[QuizSubmit] correct=%d/%d totalXP=%d", correctCount, req.Total, totalXP)

	accuracy := 0.0
	if req.Total > 0 {
		accuracy = float64(correctCount) / float64(req.Total) * 100
	}
	isPerfect := correctCount == req.Total && req.Total > 0

	// Insert game_session
	insertData, _ := json.Marshal(map[string]interface{}{
		"quiz_id":         req.QuizID,
		"user_id":         user.ID,
		"score":           correctCount,
		"total_questions": req.Total,
		"correct_count":   correctCount,
		"accuracy_rate":   accuracy,
		"is_perfect":      isPerfect,
		"completed_at":    now,
	})
	log.Printf("[QuizSubmit] Inserting game_session")
	insertBody, insertErr := h.db.Insert("game_sessions", insertData, true)
	if insertErr != nil {
		log.Printf("[QuizSubmit] Insert game_sessions FAILED: %v", insertErr)
		return c.Status(500).JSON(fiber.Map{"error": "Erreur serveur"})
	}
	log.Printf("[QuizSubmit] Insert game_sessions OK: %s", string(insertBody)[:min(len(string(insertBody)), 200)])

	// Record quiz attempt
	h.trackQuizAttempt(user.ID, req.QuizID, correctCount, totalXP)

	// Upsert into quiz_leaderboard
	lbBody, lbErr := h.db.Select("quiz_leaderboard",
		fmt.Sprintf("quiz_id=eq.%s&user_id=eq.%s&select=id,xp_earned", req.QuizID, user.ID), true)
	if lbErr == nil {
		var lbRows []map[string]interface{}
		json.Unmarshal(lbBody, &lbRows)
		if len(lbRows) > 0 {
			existingXP := int(lbRows[0]["xp_earned"].(float64))
			if totalXP > existingXP {
				lbID := lbRows[0]["id"].(string)
				upd, _ := json.Marshal(map[string]interface{}{
					"xp_earned":     totalXP,
					"accuracy_rate": accuracy,
				})
				h.db.Update("quiz_leaderboard", fmt.Sprintf("id=eq.%s", lbID), upd, true)
				log.Printf("[QuizSubmit] quiz_leaderboard updated (%d -> %d)", existingXP, totalXP)
			} else {
				log.Printf("[QuizSubmit] quiz_leaderboard skipped (existing %d >= new %d)", existingXP, totalXP)
			}
		} else {
			lbInsert, _ := json.Marshal(map[string]interface{}{
				"quiz_id":       req.QuizID,
				"user_id":       user.ID,
				"xp_earned":     totalXP,
				"accuracy_rate": accuracy,
			})
			h.db.Insert("quiz_leaderboard", lbInsert, true)
			log.Printf("[QuizSubmit] quiz_leaderboard inserted (xp: %d)", totalXP)
		}
	}

	// XP transaction
	if totalXP > 0 {
		txData, _ := json.Marshal(map[string]interface{}{
			"user_id":   user.ID,
			"source":    "quiz",
			"source_id": req.QuizID,
			"amount":    totalXP,
		})
		h.db.Insert("xp_transactions", txData, true)
	}

	// Update play_count
	quizBytes, qErr := h.db.Select("quizzes", fmt.Sprintf("id=eq.%s&select=play_count", req.QuizID), true)
	if qErr == nil {
		var qRows []map[string]interface{}
		json.Unmarshal(quizBytes, &qRows)
		if len(qRows) > 0 {
			pc := int(qRows[0]["play_count"].(float64))
			pcData, _ := json.Marshal(map[string]interface{}{"play_count": pc + 1})
			h.db.Update("quizzes", fmt.Sprintf("id=eq.%s", req.QuizID), pcData, true)
		}
	}

	// Update user_stats.quizzes_played
	statsBytes, sErr := h.db.Select("user_stats", fmt.Sprintf("user_id=eq.%s&select=quizzes_played", user.ID), true)
	if sErr == nil {
		var sRows []map[string]interface{}
		json.Unmarshal(statsBytes, &sRows)
		if len(sRows) > 0 {
			qp := int(sRows[0]["quizzes_played"].(float64))
			sd, _ := json.Marshal(map[string]interface{}{"quizzes_played": qp + 1})
			h.db.Update("user_stats", fmt.Sprintf("user_id=eq.%s", user.ID), sd, true)
		}
	}

	// Update user XP + rank
	profileBytes, pErr := h.db.Select("user_profiles", fmt.Sprintf("id=eq.%s&select=xp", user.ID), true)
	currentXP := user.XP
	if pErr == nil {
		var pRows []map[string]interface{}
		json.Unmarshal(profileBytes, &pRows)
		if len(pRows) > 0 {
			currentXP = int(pRows[0]["xp"].(float64))
		}
	}
	newXP := currentXP + totalXP
	newLevel := (newXP / 100) + 1
	ranks := h.loadRanks()
	newRank := getRankForXP(newXP, ranks)
	profileData, _ := json.Marshal(map[string]interface{}{
		"xp":    newXP,
		"level": newLevel,
		"rank":  newRank,
	})
	log.Printf("[QuizSubmit] Updating XP: %d -> %d, rank: %s (level %d)", currentXP, newXP, newRank, newLevel)
	h.db.Update("user_profiles", fmt.Sprintf("id=eq.%s", user.ID), profileData, true)

	if req.ChallengeSessionID != "" {
		log.Printf("[QuizSubmit] Saving challenge score: session=%s user=%s correct=%d/%d time=%dms", req.ChallengeSessionID, user.ID, correctCount, req.Total, req.TotalTimeMs)
		csBody, _ := h.db.Select("challenge_scores",
			fmt.Sprintf("session_id=eq.%s&user_id=eq.%s&select=id", req.ChallengeSessionID, user.ID), true)
		var csRows []map[string]interface{}
		json.Unmarshal(csBody, &csRows)
		csData, _ := json.Marshal(map[string]interface{}{
			"session_id":      req.ChallengeSessionID,
			"user_id":         user.ID,
			"correct_count":   correctCount,
			"total_questions": req.Total,
			"time_taken_ms":   req.TotalTimeMs,
		})
		if len(csRows) > 0 {
			h.db.Update("challenge_scores", fmt.Sprintf("id=eq.%s", csRows[0]["id"]), csData, true)
		} else {
			h.db.Insert("challenge_scores", csData, true)
		}
		h.checkAndDistribute(req.ChallengeSessionID)
	}

	return c.JSON(fiber.Map{"success": true, "score": correctCount, "xp_earned": totalXP})
}

func (h *Handler) Friends(c *fiber.Ctx) error {
	user := c.Locals("user").(*UserProfile)

	return renderPage(c, "Amis", fmt.Sprintf(`
<div class="friends-page">
    <div class="fp-header">
        <h1 class="page-title">👥 Amis</h1>
    </div>
    <div class="fp-tabs">
        <button class="tab-btn active" onclick="switchTab('amis',this)">👥 Amis</button>
        <button class="tab-btn" onclick="switchTab('search',this)">🔍 Rechercher</button>
        <button class="tab-btn" onclick="switchTab('requests',this)">🔔 Demandes</button>
        <button class="tab-btn" onclick="switchTab('chat',this)">💬 Chat</button>
    </div>
    <div id="tab-amis" class="tab-content active"><div id="friends-list" class="fr-list"></div></div>
    <div id="tab-search" class="tab-content" style="display:none">
        <div class="search-box"><input type="text" id="search-input" placeholder="Rechercher un utilisateur..." oninput="debounceSearch(this.value)"></div>
        <div id="search-results" class="fr-list"></div>
    </div>
    <div id="tab-requests" class="tab-content" style="display:none">
        <div id="requests-received" class="fr-list"></div>
        <div style="margin-top:16px"><h4 class="fr-section-title">Demandes envoyées</h4></div>
        <div id="requests-sent" class="fr-list"></div>
    </div>
    <div id="tab-chat" class="tab-content" style="display:none">
        <div id="conversations-list" class="fr-list"></div>
    </div>
    <div id="chat-window" style="display:none">
        <div class="chat-header">
            <button class="chat-back-btn" onclick="closeChat()">←</button>
            <div class="ch-avatar" id="chat-avatar">?</div>
            <span class="ch-username" id="chat-username"></span>
            <div class="ch-actions">
                <button class="ch-action-btn" onclick="toggleSelectionMode()" title="Sélectionner">☐</button>
                <button class="ch-action-btn ch-danger" onclick="confirmDeleteConversation()" title="Supprimer">🗑</button>
            </div>
        </div>
        <div id="chat-messages" class="chat-messages"><div class="chat-empty">Chargement...</div></div>
        <div class="chat-input-bar">
            <textarea id="message-input" rows="1" placeholder="Écrire un message..." onkeydown="if(event.key==='Enter'&&!event.shiftKey){event.preventDefault();sendMessage()}"></textarea>
            <button class="btn-primary btn-sm" onclick="sendMessage()">Envoyer</button>
        </div>
    </div>
</div>
<div id="confirm-dialog" class="dialog-overlay" style="display:none" onclick="if(event.target===this)closeDialog()">
    <div class="dialog-box">
        <h3 id="confirm-title">Confirmer</h3>
        <p id="confirm-message" class="text-muted"></p>
        <div class="dialog-actions"><button class="btn-outline btn-sm" onclick="closeDialog()">Annuler</button><button id="confirm-btn" class="btn-danger btn-sm" onclick="">Confirmer</button></div>
    </div>
</div>
<div id="report-dialog" class="dialog-overlay" style="display:none" onclick="if(event.target===this)closeReportDialog()">
    <div class="dialog-box">
        <h3>🚩 Signaler <span id="report-username"></span></h3>
        <p class="text-muted text-sm">Votre signalement sera examiné par les administrateurs.</p>
        <div class="report-reasons">
            <button class="report-reason" data-value="spam" onclick="selectReason(this)">Spam</button>
            <button class="report-reason" data-value="harassment" onclick="selectReason(this)">Harcèlement</button>
            <button class="report-reason" data-value="inappropriate" onclick="selectReason(this)">Contenu inapproprié</button>
            <button class="report-reason" data-value="cheating" onclick="selectReason(this)">Triche</button>
            <button class="report-reason" data-value="other" onclick="selectReason(this)">Autre</button>
        </div>
        <textarea id="report-description" rows="3" placeholder="Décrivez le problème..." class="pe-input"></textarea>
        <div class="dialog-actions"><button class="btn-outline btn-sm" onclick="closeReportDialog()">Annuler</button><button class="btn-danger btn-sm" onclick="submitReport()">Signaler</button></div>
    </div>
</div>
<script>
var currentUserId = '%[1]s';
var currentConversationId = null, currentChatFriendId = null, chatSelectionMode = false, selectedMessageIds = {}, searchTimeout = null;
var selectedReportUserId = null, selectedReportUsername = '', selectedReportReason = '';
function statsBadge(u){
    if(!u)return'';
    var p=parseInt(u.challenges_played)||0,w=parseInt(u.challenges_won)||0;
    if(p===0)return'';
    return'<span style="display:inline-block;padding:1px 6px;border-radius:4px;background:#6366f122;color:#a78bfa;font-size:.7rem;font-weight:600;margin-left:4px;vertical-align:middle">🏅 '+w+'/'+p+'</span>';
}
function switchTab(name,el) {
    console.log('[Friends] switchTab:',name);
    document.querySelectorAll('.tab-content').forEach(function(t){t.style.display='none'});
    document.querySelectorAll('.tab-btn').forEach(function(b){b.classList.remove('active')});
    document.getElementById('tab-'+name).style.display='block';
    if(el)el.classList.add('active');
    if(name==='amis')loadFriends();
    if(name==='requests')loadRequests();

    if(name==='chat')loadConversations();
}
function loadFriends() {
    console.log('[Friends] loadFriends: user_id='+currentUserId);
    var el=document.getElementById('friends-list');el.innerHTML='<div class="fr-loading">Chargement...</div>';
    fetch('/api/friends?user_id='+currentUserId).then(function(r){console.log('[Friends] /api/friends status:',r.status);return r.json()}).then(function(d){
        console.log('[Friends] /api/friends data:',d);
        if(d.friends&&d.friends.length>0){
            var h='';d.friends.forEach(function(f){
                var i=(f.username||'?')[0].toUpperCase(),n=f.nickname||f.username||'Inconnu';
                var avatarHtml='';
                if(f.avatar_url)avatarHtml='<div class="fr-avatar" style="overflow:hidden"><img src="'+f.avatar_url+'" alt="'+n+'" style="width:100%;height:100%;object-fit:cover"></div>';
                else avatarHtml='<div class="fr-avatar">'+i+'</div>';
                var stats=f.challenges_played?Math.max(0,f.challenges_played):0;var won=f.challenges_won?Math.max(0,f.challenges_won):0;var badgeStats=stats>0?'<span style="display:inline-block;padding:0 6px;border-radius:10px;background:rgba(99,102,241,.1);color:#a78bfa;font-size:.65rem;font-weight:600;margin-left:4px;vertical-align:middle">🏅 '+won+'/'+stats+'</span>':'';
                h+='<div class="fr-card"><a href="/profile/'+f.username+'" class="fr-card-main">'+avatarHtml+'<div class="fr-info"><div class="fr-name">'+n+badgeStats+'</div><div class="fr-meta"><span class="rank-badge rank-'+(f.rank||'f').toLowerCase()+'">'+(f.rank||'F')+'</span> Niv. '+(f.level||0)+'</div></div></a><div class="fr-actions"><button class="btn-sm" onclick="openChatFromFriend(\''+f.id+'\',\''+n.replace(/\'/g,"\\\\\'")+'\',\''+(f.nickname||'').replace(/\'/g,"\\\\\'")+'\',\''+(f.avatar_url||'').replace(/\'/g,"\\\\\'")+'\')">💬</button><button class="btn-sm btn-outline" onclick="openReport(\''+f.id+'\',\''+n.replace(/\'/g,"\\\\\'")+'\')">🚩</button><button class="btn-sm btn-outline ch-danger" onclick="confirmRemoveFriend(\''+f.friendship_id+'\')">✕</button></div></div>';
            });
            el.innerHTML=h;
        } else {el.innerHTML='<div class="fr-empty"><div class="fr-empty-icon">👥</div><p>Pas encore d\'amis</p><p class="text-sm text-muted">Recherchez des utilisateurs pour les ajouter</p></div>';}
    }).catch(function(){el.innerHTML='<div class="fr-empty"><p class="text-muted">Erreur de chargement</p></div>';});
}
function confirmRemoveFriend(id){showConfirm('Supprimer cet ami ?','Cette action est irr\u00e9versible.',function(){fetch('/api/friends/remove',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({friendship_id:id})}).then(function(){loadFriends()})});}
function debounceSearch(q){clearTimeout(searchTimeout);searchTimeout=setTimeout(function(){searchUsers(q)},300);}
function searchUsers(query) {
    if(query.length<2){document.getElementById('search-results').innerHTML='';return;}
    var el=document.getElementById('search-results');el.innerHTML='<div class="fr-loading">Recherche...</div>';
    fetch('/api/users/search?q='+encodeURIComponent(query)).then(function(r){return r.json()}).then(function(d){
        if(d.users&&d.users.length>0){
            var h='';d.users.forEach(function(u){var i=(u.username||'?')[0].toUpperCase(),n=u.nickname||u.username;
            var s='<button class="btn-sm" onclick="sendFriendRequest(\''+u.id+'\',this)">Ajouter</button>';
            if(u._friendship_status==='pending')s='<span class="badge-outline">Demande envoy\u00e9e</span>';
            else if(u._friendship_status==='accepted')s='<span class="badge-success">Ami</span>';
            h+='<div class="fr-card"><a href="/profile/'+u.username+'" class="fr-card-main"><div class="fr-avatar">'+(u.avatar_url?'<img src="'+u.avatar_url+'" style="width:100%;height:100%;border-radius:50%;object-fit:cover;display:block">':i)+'</div><div class="fr-info"><div class="fr-name">'+n+statsBadge(u)+'</div><div class="fr-meta"><span class="rank-badge rank-'+(u.rank||'f').toLowerCase()+'">'+(u.rank||'F')+'</span> Niv. '+(u.level||0)+'</div></div></a><div class="fr-actions">'+s+'</div></div>'});


            el.innerHTML=h;
        } else {el.innerHTML='<div class="fr-empty"><p class="text-muted">Aucun utilisateur trouv\u00e9</p></div>';}
    }).catch(function(){el.innerHTML='<div class="fr-empty"><p class="text-muted">Erreur</p></div>';});
}
function sendFriendRequest(userId,btn){
    fetch('/api/friends/request',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({user_id:userId})}).then(function(r){return r.json()}).then(function(d){
        if(d.success){var p=btn.closest('.fr-card');if(p){var a=p.querySelector('.fr-actions');if(a)a.innerHTML='<span class="badge-outline">Demande envoy\u00e9e</span>';}}
        else alert(d.error||'Erreur');
    });
}
function searchUsers(query) {
    if(query.length<2){document.getElementById('search-results').innerHTML='';return;}
    var el=document.getElementById('search-results');el.innerHTML='<div class="fr-loading">Recherche...</div>';
    fetch('/api/users/search?q='+encodeURIComponent(query)).then(function(r){return r.json()}).then(function(d){
        if(d.users&&d.users.length>0){
            var h='';d.users.forEach(function(u){
                var i=(u.username||'?')[0].toUpperCase(),n=u.nickname||u.username;
                var avatarHtml='';
                if(u.avatar_url)avatarHtml='<div class="fr-avatar" style="overflow:hidden"><img src="'+u.avatar_url+'" alt="'+n+'" style="width:100%;height:100%;object-fit:cover"></div>';
                else avatarHtml='<div class="fr-avatar">'+i+'</div>';
                var stats=u.challenges_played?Math.max(0,u.challenges_played):0;var won=u.challenges_won?Math.max(0,u.challenges_won):0;var badgeStats=stats>0?'<span style="display:inline-block;padding:0 6px;border-radius:10px;background:rgba(99,102,241,.1);color:#a78bfa;font-size:.65rem;font-weight:600;margin-left:4px;vertical-align:middle">🏅 '+won+'/'+stats+'</span>':'';
                var s='<button class="btn-sm" onclick="sendFriendRequest(\''+u.id+'\',this)">Ajouter</button>';
                if(u._friendship_status==='pending')s='<span class="badge-outline">Demande envoy\u00e9e</span>';
                else if(u._friendship_status==='accepted')s='<span class="badge-success">Ami</span>';
                h+='<div class="fr-card"><a href="/profile/'+u.username+'" class="fr-card-main">'+avatarHtml+'<div class="fr-info"><div class="fr-name">'+n+badgeStats+'</div><div class="fr-meta"><span class="rank-badge rank-'+(u.rank||'f').toLowerCase()+'">'+(u.rank||'F')+'</span> Niv. '+(u.level||0)+'</div></div></a><div class="fr-actions">'+s+'</div></div>';
            });
            el.innerHTML=h;
        } else {el.innerHTML='<div class="fr-empty"><p class="text-muted">Aucun utilisateur trouv\u00e9</p></div>';}
    }).catch(function(){el.innerHTML='<div class="fr-empty"><p class="text-muted">Erreur</p></div>';});
}
function sendFriendRequest(userId,btn){
    fetch('/api/friends/request',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({user_id:userId})}).then(function(r){return r.json()}).then(function(d){
        if(d.success){var p=btn.closest('.fr-card');if(p){var a=p.querySelector('.fr-actions');if(a)a.innerHTML='<span class="badge-outline">Demande envoy\u00e9e</span>';}}
        else alert(d.error||'Erreur');
    });
}
function loadRequests(){
    console.log('[Friends] loadRequests: user_id='+currentUserId);
    var recvEl=document.getElementById('requests-received');recvEl.innerHTML='<h4 class="fr-section-title">Demandes re\u00e7ues</h4>';
    fetch('/api/friends/requests?user_id='+currentUserId).then(function(r){console.log('[Friends] /api/friends/requests status:',r.status);return r.json()}).then(function(d){
        console.log('[Friends] /api/friends/requests data:',d);
        var h='<h4 class="fr-section-title">Demandes re\u00e7ues</h4>';
        if(d.requests&&d.requests.length>0){d.requests.forEach(function(r){
            var i=(r.username||'?')[0].toUpperCase();
            var avatarHtml='';
            if(r.avatar_url)avatarHtml='<div class="fr-avatar" style="overflow:hidden"><img src="'+r.avatar_url+'" alt="'+(r.nickname||r.username)+'" style="width:100%;height:100%;object-fit:cover"></div>';
            else avatarHtml='<div class="fr-avatar">'+i+'</div>';
            h+='<div class="fr-card"><a href="/profile/'+r.username+'" class="fr-card-main">'+avatarHtml+'<div class="fr-info"><div class="fr-name">'+(r.nickname||r.username)+statsBadge(r)+'</div></div></a><div class="fr-actions"><button class="btn-sm btn-success" onclick="acceptRequest(\''+r.friendship_id+'\',this)">\u2713 Accepter</button><button class="btn-sm btn-outline ch-danger" onclick="rejectRequest(\''+r.friendship_id+'\',this)">\u2715</button></div></div>';
        });}
        else h+='<div class="fr-empty"><p class="text-muted">Aucune demande re\u00e7ue</p></div>';
        recvEl.innerHTML=h;
    });
    var sentEl=document.getElementById('requests-sent');
    fetch('/api/friends/requests/sent?user_id='+currentUserId).then(function(r){return r.json()}).then(function(d){
        var h='<h4 class="fr-section-title">Demandes envoy\u00e9es</h4>';
        if(d.requests&&d.requests.length>0){d.requests.forEach(function(r){
            var i=(r.username||'?')[0].toUpperCase();
            var avatarHtml='';
            if(r.avatar_url)avatarHtml='<div class="fr-avatar" style="overflow:hidden"><img src="'+r.avatar_url+'" alt="'+(r.nickname||r.username)+'" style="width:100%;height:100%;object-fit:cover"></div>';
            else avatarHtml='<div class="fr-avatar">'+i+'</div>';
            h+='<div class="fr-card"><a href="/profile/'+r.username+'" class="fr-card-main">'+avatarHtml+'<div class="fr-info"><div class="fr-name">'+(r.nickname||r.username)+statsBadge(r)+'</div></div></a><div class="fr-actions"><span class="badge-outline">En attente</span></div></div>';
        });}
        else h+='<div class="fr-empty"><p class="text-muted">Aucune demande envoy\u00e9e</p></div>';
        sentEl.innerHTML=h;
    });
}
function acceptRequest(id,btn){fetch('/api/friends/accept',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({friendship_id:id})}).then(function(r){return r.json()}).then(function(d){if(d.success){var c=btn.closest('.fr-card');if(c)c.style.display='none';loadFriends();}});}
function rejectRequest(id,btn){fetch('/api/friends/reject',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({friendship_id:id})}).then(function(r){return r.json()}).then(function(d){if(d.success){var c=btn.closest('.fr-card');if(c)c.style.display='none';}});}
function loadChallengeInvitations(){
    console.log('[Friends] loadChallengeInvitations: user_id='+currentUserId);
    fetch('/api/challenges/invitations?user_id='+currentUserId).then(function(r){return r.json()}).then(function(d){
        var h='<h4 class="fr-section-title">⚔️ Demandes de défi reçues</h4>';
        if(d.invitations&&d.invitations.length>0){d.invitations.forEach(function(inv){var n=inv.inviter_username||inv.inviter_nickname||'Quelqu\'un',i=n[0].toUpperCase();h+='<div class="fr-card challenge-card"><a href="/challenges/'+inv.session_id+'" style="text-decoration:none;flex:1;display:flex"><div class="fr-card-main"><div class="fr-avatar ch-avatar-purple">'+i+'</div><div class="fr-info"><div class="fr-name">'+n+'</div><div class="fr-meta text-sm">⚔️ '+(inv.quiz_title||'Quiz')+'</div></div></div></a><div class="fr-actions"><button class="btn-sm btn-success" onclick="acceptChallenge(\''+inv.id+'\',this)">✓ Accepter</button><button class="btn-sm btn-outline ch-danger" onclick="rejectChallenge(\''+inv.id+'\',this)">✕ Refuser</button></div></div>';});}
        else h+='<div class="fr-empty"><p class="text-muted">Aucune demande de défi en attente</p></div>';
        document.getElementById('challenge-invitations').innerHTML=h;
    }).catch(function(){document.getElementById('challenge-invitations').innerHTML='<div class="fr-empty"><p class="text-muted">Erreur de chargement</p></div>';});
    fetch('/api/challenges/mine?user_id='+currentUserId).then(function(r){return r.json()}).then(function(d){
        var h='<h4 class="fr-section-title">🎮 Mes défis</h4>';
        if(d.challenges&&d.challenges.length>0){d.challenges.forEach(function(ch){
            var sl=ch.status==='waiting'?'⏳ En attente':ch.status==='ready'?'✅ Prêt':ch.status==='playing'?'🎮 En cours':ch.status==='completed'?'🏆 Terminé':ch.status==='cancelled'?'🗑️ Annulé':ch.status;
            var bc=ch.status==='waiting'?'badge-outline':ch.status==='ready'?'badge-primary':ch.status==='completed'?'badge-success':ch.status==='cancelled'?'badge-danger':'badge-outline';
            h+='<a href="/challenges/'+ch.id+'" class="fr-card" style="text-decoration:none"><div class="fr-card-main"><div class="fr-avatar ch-avatar-purple">⚔️</div><div class="fr-info"><div class="fr-name">'+(ch.quiz_title||'Quiz')+'</div><div class="fr-meta text-xs text-muted">'+(ch.participant_count||0)+' participant(s) • '+(ch.total_xp_pool||0)+' XP</div></div></div><div class="fr-actions"><span class="'+bc+'">'+sl+'</span></div></a>';
        });}
        else h+='<div class="fr-empty"><p class="text-muted">Aucun défi</p><p class="text-xs text-muted" style="margin-top:4px">Créez un défi depuis la page d\'un quiz</p></div>';
        document.getElementById('my-challenges').innerHTML=h;
    }).catch(function(){document.getElementById('my-challenges').innerHTML='<div class="fr-empty"><p class="text-muted">Erreur de chargement</p></div>';});
}
function acceptChallenge(id,btn){fetch('/api/challenges/accept',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({invitation_id:id})}).then(function(r){return r.json()}).then(function(d){if(d.success){var c=btn.closest('.fr-card');if(c)c.style.display='none';}else alert(d.error||'Erreur');});}
function rejectChallenge(id,btn){fetch('/api/challenges/refuse',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({invitation_id:id})}).then(function(r){return r.json()}).then(function(d){if(d.success){var c=btn.closest('.fr-card');if(c)c.style.display='none';}});}
function loadConversations(){
    console.log('[Friends] loadConversations: user_id='+currentUserId);
    var el=document.getElementById('conversations-list');el.innerHTML='<div class="fr-loading">Chargement...</div>';
    fetch('/api/conversations?user_id='+currentUserId).then(function(r){console.log('[Friends] /api/conversations status:',r.status);return r.json()}).then(function(d){
        console.log('[Friends] /api/conversations data:',d);
        if(d.conversations&&d.conversations.length>0){
            var h='';d.conversations.forEach(function(conv){
                var n=conv.other_nickname||conv.other_username||'Utilisateur',i=(n[0]||'?').toUpperCase(),a=conv.other_avatar_url||'',lm=conv.last_message||'Aucun message',ub=conv.unread_count>0?'<span class="unread-badge">'+conv.unread_count+'</span>':'';
                var avHtml=a?'<img src="'+a+'" style="width:36px;height:36px;border-radius:50%;object-fit:cover">':'<div class="fr-avatar" style="min-width:36px;width:36px;height:36px;display:flex;align-items:center;justify-content:center">'+i+'</div>';
                h+='<div class="fr-card" onclick="openChat(\''+conv.id+'\',\''+conv.other_user_id+'\',\''+n.replace(/'/g,"\\'")+'\',\''+a+'\')">'+avHtml+'<div class="fr-info"><div class="fr-name">'+n+statsBadge(conv)+'</div><div class="fr-meta text-sm text-muted">'+escapeHtml(lm.substring(0,50))+'</div></div>'+ub+'</div>';
            });
            el.innerHTML=h;
        } else {el.innerHTML='<div class="fr-empty"><div class="fr-empty-icon">\uD83D\uDCAC</div><p>Aucune conversation</p><p class="text-sm text-muted">Commencez \u00e0 discuter avec vos amis</p></div>';}
    });
}
function openChatFromFriend(friendId,username,nickname,avatarUrl){
    fetch('/api/conversations/get-or-create',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({friend_id:friendId})}).then(function(r){return r.json()}).then(function(d){
        if(d.conversation_id){switchTab('chat');openChat(d.conversation_id,friendId,nickname||username,avatarUrl||'');}
    });
}
function openChat(convId,friendId,displayName,avatarUrl){
    currentConversationId=convId;currentChatFriendId=friendId;
    document.getElementById('tab-chat').style.display='none';document.getElementById('chat-window').style.display='flex';
    document.getElementById('chat-username').textContent=displayName;
    var avEl=document.getElementById('chat-avatar');
    if(avatarUrl){avEl.innerHTML='<img src="'+avatarUrl+'" style="width:36px;height:36px;border-radius:50%;object-fit:cover">';}else{avEl.textContent=(displayName[0]||'?').toUpperCase();}
    loadMessages();if(window._chatInterval)clearInterval(window._chatInterval);window._chatInterval=setInterval(function(){if(currentConversationId&&document.getElementById('chat-window').style.display!=='none')loadMessagesSilent()},3000);
}
function closeChat(){document.getElementById('chat-window').style.display='none';document.getElementById('tab-chat').style.display='block';currentConversationId=null;currentChatFriendId=null;chatSelectionMode=false;selectedMessageIds={};loadConversations();}
function loadMessages(){if(!currentConversationId)return;fetch('/api/messages?conversation_id='+currentConversationId).then(function(r){return r.json()}).then(function(d){renderMessages(d.messages||[]);});}
function loadMessagesSilent(){if(!currentConversationId)return;fetch('/api/messages?conversation_id='+currentConversationId).then(function(r){return r.json()}).then(function(d){renderMessages(d.messages||[]);});}
function getDateLabel(d){
    var today=new Date(),yesterday=new Date(today);yesterday.setDate(yesterday.getDate()-1);
    if(d.toDateString()===today.toDateString())return"Aujourd'hui";
    if(d.toDateString()===yesterday.toDateString())return"Hier";
    return d.toLocaleDateString('fr-FR',{weekday:'long',day:'numeric',month:'long',year:'numeric'});
}
function fmtTime(d){var h=d.getHours(),m=d.getMinutes();return(h<10?'0':'')+h+':'+(m<10?'0':'')+m;}
function fmtMsgDate(d){
    var today=new Date(),yesterday=new Date(today);yesterday.setDate(yesterday.getDate()-1);
    var time=fmtTime(d);
    if(d.toDateString()===today.toDateString())return time;
    if(d.toDateString()===yesterday.toDateString())return'Hier '+time;
    return d.toLocaleDateString('fr-FR',{day:'2-digit',month:'2-digit',year:'numeric'})+' '+time;
}
function renderMessages(msgs){
    var container=document.getElementById('chat-messages');
    if(!msgs||msgs.length===0){container.innerHTML='<div class="chat-empty">Aucun message. Commencez la conversation !</div>';return;}
    var h='',lastDate='';
    msgs.forEach(function(msg){
        var isOwn=msg.sender_id===currentUserId,bc=isOwn?'msg-own':'msg-other';
        var msgDate=new Date(msg.created_at);
        var dateKey=msgDate.toLocaleDateString('fr-FR');
        if(dateKey!==lastDate){
            h+='<div class="msg-date-sep"><span>'+getDateLabel(msgDate)+'</span></div>';
            lastDate=dateKey;
        }
        if(chatSelectionMode){var ck=selectedMessageIds[msg.id]?'\u2713':'\u2610',sc=selectedMessageIds[msg.id]?' msg-selected':'';h+='<div class="msg-row '+bc+'"><div class="msg-checkbox">'+ck+'</div><div class="msg-bubble'+sc+'" onclick="toggleSelectMessage(\''+msg.id+'\')"><div class="msg-text">'+escapeHtml(msg.content)+'</div><div class="msg-time">'+fmtMsgDate(msgDate)+'</div></div></div>';}else{h+='<div class="msg-row '+bc+'"><div class="msg-bubble"><div class="msg-text">'+escapeHtml(msg.content)+'</div><div class="msg-time">'+fmtMsgDate(msgDate)+'</div></div></div>';}
    });
    container.innerHTML=h;container.scrollTop=container.scrollHeight;
}
function sendMessage(){var input=document.getElementById('message-input'),content=input.value.trim();if(!content||!currentConversationId)return;input.value='';fetch('/api/messages/send',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({conversation_id:currentConversationId,content:content})}).then(function(r){return r.json()}).then(function(d){if(d.success)loadMessages();});}
function toggleSelectionMode(){chatSelectionMode=!chatSelectionMode;selectedMessageIds={};loadMessages();}
function toggleSelectMessage(id){if(selectedMessageIds[id])delete selectedMessageIds[id];else selectedMessageIds[id]=true;loadMessages();}
function confirmDeleteConversation(){showConfirm('Supprimer la discussion ?','Tous les messages seront supprim\u00e9s.',function(){fetch('/api/conversations/delete',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({conversation_id:currentConversationId})}).then(function(){closeChat()})});}
function showConfirm(title,message,onConfirm){document.getElementById('confirm-title').textContent=title;document.getElementById('confirm-message').textContent=message;document.getElementById('confirm-btn').onclick=function(){closeDialog();onConfirm();};document.getElementById('confirm-dialog').style.display='flex';}
function closeDialog(){document.getElementById('confirm-dialog').style.display='none';}
function openReport(userId,username){selectedReportUserId=userId;selectedReportUsername=username;selectedReportReason='';document.getElementById('report-username').textContent=username;document.getElementById('report-description').value='';document.querySelectorAll('.report-reason').forEach(function(b){b.classList.remove('selected');});document.getElementById('report-dialog').style.display='flex';}
function closeReportDialog(){document.getElementById('report-dialog').style.display='none';}
function selectReason(btn){document.querySelectorAll('.report-reason').forEach(function(b){b.classList.remove('selected');});btn.classList.add('selected');selectedReportReason=btn.getAttribute('data-value');}
function submitReport(){if(!selectedReportReason){alert('Veuillez s\u00e9lectionner une raison');return;}fetch('/api/reports/user',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({user_id:selectedReportUserId,reason:selectedReportReason,description:document.getElementById('report-description').value})}).then(function(r){return r.json()}).then(function(d){if(d.success){closeReportDialog();alert('Signalement envoy\u00e9');}else alert(d.error||'Erreur');});}
function escapeHtml(text){var d=document.createElement('div');d.textContent=text;return d.innerHTML;}
function timeAgo(dateStr){if(!dateStr)return'';var date=new Date(dateStr),now=new Date(),diff=(now-date)/1000;if(diff<60)return"à l'instant";if(diff<3600)return'il y a '+Math.floor(diff/60)+' min';if(diff<86400)return'il y a '+Math.floor(diff/3600)+'h';if(diff<172800)return'hier';return'il y a '+Math.floor(diff/86400)+' jours';}
loadFriends();
</script>
	`, user.ID))
}

func (h *Handler) ChallengesPage(c *fiber.Ctx) error {
	user := c.Locals("user").(*UserProfile)

	return renderPage(c, "Défis", fmt.Sprintf(`
<div class="challenges-page">
    <div class="fp-header">
        <h1 class="page-title">⚔️ Défis</h1>
    </div>
    <div id="challenge-invitations" class="fr-list"></div>
    <div style="margin-top:16px"><h4 class="fr-section-title">Mes défis</h4></div>
    <div id="my-challenges" class="fr-list"></div>
</div>

<style>
.challenges-page{max-width:600px;margin:0 auto;padding:24px}
</style>
<script>
var currentUserId='%[1]s';
function loadChallenges(){
    fetch('/api/challenges/invitations?user_id='+currentUserId).then(function(r){return r.json()}).then(function(d){
        var h='<h4 class="fr-section-title">⚔️ Demandes de défi reçues</h4>';
        if(d.invitations&&d.invitations.length>0){d.invitations.forEach(function(inv){var n=inv.inviter_username||inv.inviter_nickname||'Quelqu\'un',i=n[0].toUpperCase();h+='<div class="fr-card challenge-card"><a href="/challenges/'+inv.session_id+'" style="text-decoration:none;flex:1;display:flex"><div class="fr-card-main"><div class="fr-avatar ch-avatar-purple">'+i+'</div><div class="fr-info"><div class="fr-name">'+n+'</div><div class="fr-meta text-sm">⚔️ '+(inv.quiz_title||'Quiz')+'</div></div></div></a><div class="fr-actions"><button class="btn-sm btn-success" onclick="acceptChallenge(\''+inv.id+'\',this)">✓ Accepter</button><button class="btn-sm btn-outline ch-danger" onclick="rejectChallenge(\''+inv.id+'\',this)">✕ Refuser</button></div></div>';});}
        else h+='<div class="fr-empty"><p class="text-muted">Aucune demande de défi en attente</p></div>';
        document.getElementById('challenge-invitations').innerHTML=h;
    }).catch(function(){document.getElementById('challenge-invitations').innerHTML='<div class="fr-empty"><p class="text-muted">Erreur de chargement</p></div>';});
    fetch('/api/challenges/mine?user_id='+currentUserId).then(function(r){return r.json()}).then(function(d){
        var h='<h4 class="fr-section-title">🎮 Mes défis</h4>';
        if(d.challenges&&d.challenges.length>0){d.challenges.forEach(function(ch){
            var sl=ch.status==='waiting'?'⏳ En attente':ch.status==='ready'?'✅ Prêt':ch.status==='playing'?'🎮 En cours':ch.status==='completed'?'🏆 Terminé':ch.status==='cancelled'?'🗑️ Annulé':ch.status;
            var bc=ch.status==='waiting'?'badge-outline':ch.status==='ready'?'badge-primary':ch.status==='completed'?'badge-success':ch.status==='cancelled'?'badge-danger':'badge-outline';
            h+='<a href="/challenges/'+ch.id+'" class="fr-card" style="text-decoration:none"><div class="fr-card-main"><div class="fr-avatar ch-avatar-purple">⚔️</div><div class="fr-info"><div class="fr-name">'+(ch.quiz_title||'Quiz')+'</div><div class="fr-meta text-xs text-muted">'+(ch.participant_count||0)+' participant(s) • '+(ch.total_xp_pool||0)+' XP</div></div></div><div class="fr-actions"><span class="'+bc+'">'+sl+'</span></div></a>';
        });}
        else h+='<div class="fr-empty"><p class="text-muted">Aucun défi</p><p class="text-xs text-muted" style="margin-top:4px">Créez un défi depuis la page d\'un quiz</p></div>';
        document.getElementById('my-challenges').innerHTML=h;
    }).catch(function(){document.getElementById('my-challenges').innerHTML='<div class="fr-empty"><p class="text-muted">Erreur de chargement</p></div>';});
}
function acceptChallenge(id,btn){fetch('/api/challenges/accept',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({invitation_id:id})}).then(function(r){return r.json()}).then(function(d){if(d.success){var c=btn.closest('.fr-card');if(c)c.style.display='none';}else alert(d.error||'Erreur');});}
function rejectChallenge(id,btn){fetch('/api/challenges/refuse',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({invitation_id:id})}).then(function(r){return r.json()}).then(function(d){if(d.success){var c=btn.closest('.fr-card');if(c)c.style.display='none';}});}
loadChallenges();
</script>
	`, user.ID))
}

func (h *Handler) ChallengeDetail(c *fiber.Ctx) error {
	id := c.Params("id")
	user := c.Locals("user").(*UserProfile)

	challengeHTML := `<div class="challenge-page">
    <a href="/defis" style="color:#6366f1;display:inline-block;margin-bottom:16px;">← Retour aux défis</a>
    <div class="challenge-card">
        <div class="challenge-header">
            <span class="challenge-icon">⚔️</span>
            <h1 id="challenge-title">Défi en cours</h1>
        </div>
        <div id="challenge-content">
            <p style="color:#94a3b8;text-align:center;padding:40px;">Chargement du défi...</p>
        </div>
    </div>
    <div id="raise-panel" style="display:none;margin-top:16px" class="challenge-card">
        <h3 style="margin-bottom:12px">💰 Renchérir la mise</h3>
        <div id="raise-fields">
            <p class="text-muted text-sm" style="margin-bottom:8px">Votre mise actuelle: <strong id="current-my-bet">0</strong> XP</p>
            <p class="text-muted text-sm" style="margin-bottom:8px">Votre solde XP: <strong id="current-my-xp">0</strong> XP</p>
            <label>Ajouter XP</label>
            <input type="number" id="raise-amount" class="pe-input" value="50" min="10" max="10000">
            <button class="btn-primary" style="margin-top:8px;width:100%" onclick="raiseBet()">💰 Renchérir</button>
        </div>
    </div>
    <div id="invite-panel" style="display:none;margin-top:16px" class="challenge-card">
        <h3 style="margin-bottom:12px">👥 Inviter des amis</h3>
        <div id="invite-friends-list"></div>
        <button class="btn-primary" style="margin-top:12px;width:100%" onclick="sendInvites()">Envoyer les invitations</button>
    </div>
</div>

<style>
.challenge-page{max-width:600px;margin:0 auto}
.challenge-card{background:#16213e;border:1px solid #2d2d44;border-radius:12px;padding:24px}
.challenge-header{display:flex;align-items:center;gap:12px;margin-bottom:20px}
.challenge-icon{font-size:2rem}
.ch-status{display:inline-block;padding:4px 12px;border-radius:20px;font-size:.75rem;font-weight:600}
.ch-status.waiting{background:rgba(251,191,36,.15);color:#fbbf24}
.ch-status.ready{background:rgba(34,197,94,.15);color:#22c55e}
.ch-status.playing{background:rgba(99,102,241,.15);color:#6366f1}
.ch-status.completed{background:rgba(148,163,184,.15);color:#94a3b8}
.ch-status.cancelled{background:rgba(239,68,68,.15);color:#ef4444}
.ch-participant{display:flex;align-items:center;gap:12px;padding:10px 0;border-bottom:1px solid #1e293b}
.ch-participant:last-child{border-bottom:none}
.ch-invite{display:flex;align-items:center;justify-content:space-between;padding:8px 0;border-bottom:1px solid #1e293b}
.ch-invite:last-child{border-bottom:none}
.ch-invite-status{font-size:.75rem;padding:2px 8px;border-radius:12px}
.ch-invite-status.pending{background:rgba(251,191,36,.15);color:#fbbf24}
.ch-invite-status.accepted{background:rgba(34,197,94,.15);color:#22c55e}
.ch-invite-status.refused{background:rgba(239,68,68,.15);color:#ef4444}
.ch-invite-checkbox{width:18px;height:18px;accent-color:#6366f1}
.ch-bet-accepted{color:#22c55e;font-size:.8rem}
.ch-bet-pending{color:#fbbf24;font-size:.8rem}
.ch-raise-row{display:flex;gap:8px;align-items:center;margin-top:8px}
</style>
<script>
var challengeId='__CHALLENGE_ID__';
var currentUserId='__USER_ID__';
var challengeCreatorId='';
var challengeStatus='';
var inviteFriendsData=[];
var currentUserParticipant=null;
var currentUserXp=__USER_XP__;
var xpPool=0;

function statsBadge(u){
    var p=parseInt(u.challenges_played)||0,w=parseInt(u.challenges_won)||0;
    if(p===0)return'';
    return'<span style="display:inline-block;padding:1px 6px;border-radius:4px;background:#6366f122;color:#a78bfa;font-size:.7rem;font-weight:600;margin-left:4px;vertical-align:middle">🏅 '+w+'/'+p+'</span>';
}

function loadChallenge(){
    fetch('/api/challenges/'+challengeId).then(function(r){return r.json()}).then(function(d){
        if(d.error){document.getElementById('challenge-content').innerHTML='<p style="color:#ef4444;text-align:center">'+d.error+'</p>';return;}
        var ch=d.session;
        if(!ch){document.getElementById('challenge-content').innerHTML='<p style="color:#ef4444;text-align:center">Défi introuvable</p>';return;}
        var status=ch.status||'waiting';
        challengeStatus=status;
        challengeCreatorId=ch.creator_id||'';
        var isCreator=challengeCreatorId===currentUserId;
        var statusMap={waiting:'En attente',ready:'✅ Prêt',playing:'🎮 En cours',completed:'🏆 Terminé',cancelled:'🗑️ Annulé'};
        var quizTitle=ch.quiz_title||'Quiz';
        var qCount=ch.question_count||'?';
        var rewardMode=ch.reward_mode||'all_for_one';
        var rewardLabel=rewardMode==='all_for_one'?'⚔️ All for one (1er = 100%)':'💪 Pouvoir de l\'amitié (1er=50%, 2e=35%, 3e=15%)';
        xpPool=ch.total_xp_pool||0;
        var h='<div style="margin-bottom:12px"><strong>📝 Quiz:</strong> '+quizTitle+'</div>';
        h+='<div style="margin-bottom:12px"><strong>❓ Questions:</strong> '+qCount+'</div>';
        h+='<div style="margin-bottom:12px"><strong>💰 Mise totale:</strong> '+xpPool+' XP</div>';
        h+='<div style="margin-bottom:12px"><strong>🏆 Récompense:</strong> '+rewardLabel+'</div>';
        h+='<div style="margin-bottom:16px"><strong>📊 Statut:</strong> <span class="ch-status '+status+'">'+(statusMap[status]||status)+'</span></div>';

        var parts=d.participants||[];
        currentUserParticipant=null;
        if(parts.length>0){
            h+='<h4 style="margin-bottom:8px">⚔️ Participants</h4>';
            parts.forEach(function(p){
                var u=p.user||{};
                var name=u.nickname||u.username||'Joueur';
                var rank=u.rank||'F';
                var av=(name[0]||'?').toUpperCase();
                var indBet=p.individual_bet||0;
                var ba=p.bet_accepted?'✅':'⏳';
                if(p.user_id===currentUserId){
                    currentUserParticipant=p;
                }
                h+='<div class="ch-participant"><div style="width:36px;height:36px;border-radius:50%;background:#6366f1;display:flex;align-items:center;justify-content:center;font-weight:700;color:white;font-size:.85rem">'+av+'</div><div style="flex:1"><div style="font-weight:600">'+name+statsBadge(u)+' <span style="font-size:.7rem;color:#94a3b8">('+rank+')</span></div><div style="font-size:.8rem;color:#94a3b8">Mise: '+indBet+' XP <span class="'+(p.bet_accepted?'ch-bet-accepted':'ch-bet-pending')+'">'+ba+'</span></div></div></div>';
            });
        }
        var invs=d.invitations||[];
        var pendingInviteId=null;
        if(invs.length>0){
            h+='<h4 style="margin:16px 0 8px">📧 Invitations</h4>';
            invs.forEach(function(inv){
                var u=inv.invitee||{};
                var name=u.nickname||u.username||'?';
                var invStatus=inv.status||'pending';
                var statusLabel=invStatus==='pending'?'En attente':invStatus==='accepted'?'Accepté':'Refusé';
                h+='<div class="ch-invite"><span>'+name+statsBadge(u)+'</span>';
                h+=' <span class="ch-invite-status '+invStatus+'">'+statusLabel+'</span>';
                if(inv.invitee_id===currentUserId&&invStatus==='pending'){
                    pendingInviteId=inv.id;
                }
                if(isCreator&&invStatus==='refused'){
                    h+=' <button class="btn-sm btn-outline" style="margin-left:8px" onclick="resendInvite(\''+inv.id+'\')">📨 Relancer</button>';
                }
                h+='</div>';
            });
        }

        if(status==='waiting'){
            var myBet=0;
            var canInteract=false;
            if(currentUserParticipant){
                myBet=currentUserParticipant.individual_bet||0;
                document.getElementById('current-my-bet').textContent=myBet;
                if(currentUserParticipant.bet_accepted===false){canInteract=true;}
                h+='<div style="margin-top:12px"><button class="btn-primary btn-sm" onclick="toggleRaisePanel()">💰 Renchérir</button></div>';
            }else if(pendingInviteId){
                myBet=xpPool;
                document.getElementById('current-my-bet').textContent=myBet;
                canInteract=true;
                h+='<div style="margin-top:12px"><button class="btn-primary btn-sm" onclick="toggleRaisePanel()">💰 Renchérir</button></div>';
            }
            if(canInteract){
                if(myBet>currentUserXp){
                    h+='<div style="margin-top:16px;padding:16px;background:rgba(239,68,68,.1);border-radius:8px;border:1px solid rgba(239,68,68,.2)">';
                    h+='<p style="color:#ef4444;font-weight:600">❌ XP insuffisant</p>';
                    h+='<p class="text-sm" style="margin-bottom:12px;color:#94a3b8">Vous avez <strong>'+currentUserXp+' XP</strong> mais votre mise individuelle est de <strong>'+myBet+' XP</strong>. Vous ne pouvez pas accepter ce défi tant que vous n\'aurez pas assez d\'XP.</p>';
                    h+='<div style="display:flex;gap:8px"><button class="btn-sm" disabled style="opacity:0.4;cursor:not-allowed;background:#6366f1;color:white;border:none;padding:8px 16px;border-radius:8px">✅ Accepter (XP insuffisant)</button><button class="btn-sm btn-danger-outline" onclick="respondBet(false)">❌ Refuser</button></div></div>';
                }else{
                    h+='<div style="margin-top:16px;padding:16px;background:rgba(251,191,36,.1);border-radius:8px;border:1px solid rgba(251,191,36,.2)">';
                    h+='<p style="margin-bottom:8px;color:#fbbf24;font-weight:600">⚡ La mise a changé !</p>';
                    h+='<p class="text-sm text-muted" style="margin-bottom:12px">La mise totale est maintenant de <strong>'+xpPool+' XP</strong>. Votre mise individuelle est de <strong>'+myBet+' XP</strong>.</p>';
                    h+='<div style="display:flex;gap:8px"><button class="btn-sm btn-success" onclick="respondBet(true)">✅ Accepter</button><button class="btn-sm btn-danger-outline" onclick="respondBet(false)">❌ Refuser et quitter</button></div></div>';
                }
            }
        }

        if(status==='ready'){
            var played=d.user_score?true:false;
            if(played){
                var sc=d.user_score;
                h+='<div style="margin-top:20px;text-align:center"><p style="margin-bottom:12px;color:#22c55e">✅ Déjà joué — '+sc.correct_count+'/'+sc.total_questions+' bonnes réponses</p><a href="/challenges/'+challengeId+'/leaderboard" class="btn-primary" style="padding:12px 32px">🏆 Voir le classement</a></div>';
            }else{
                h+='<div style="margin-top:20px;text-align:center;display:flex;gap:8px;justify-content:center;flex-wrap:wrap"><a href="/quiz/'+(ch.quiz_id||'')+'/play?challenge='+challengeId+'" class="btn-primary" style="padding:12px 32px">🎮 Jouer le défi</a><a href="/challenges/'+challengeId+'/leaderboard" class="btn-ghost" style="padding:12px 24px">🏆 Classement</a></div>';
            }
        }
        if(isCreator){
            if(status==='waiting'){
                h+='<div style="margin-top:16px;padding:12px;background:#0f172a;border-radius:8px;margin-bottom:8px">';
                h+='<label style="font-size:.85rem;color:#94a3b8;display:block;margin-bottom:6px">🏆 Mode de récompense</label>';
                h+='<select id="reward-mode-select" style="width:100%;padding:8px;background:#16213e;border:1px solid #2d2d44;border-radius:8px;color:white;font-size:.9rem" onchange="setRewardMode(this.value)">';
                var acceptedCount=parts.filter(function(p){return p.status==='accepted'}).length;
                h+='<option value="all_for_one"'+(rewardMode==='all_for_one'?' selected':'')+'>⚔️ All for one — 1er remporte tout (100%)</option>';
                if(acceptedCount>=3)h+='<option value="friendship"'+(rewardMode==='friendship'?' selected':'')+'>💪 Pouvoir de l\'amitié — 1er=50%, 2e=35%, 3e=15%</option>';
                h+='</select></div>';
                h+='<div style="margin-top:16px;display:flex;gap:8px;flex-wrap:wrap">';
                h+='<button class="btn-primary" onclick="showInvitePanel()">👥 Inviter des amis</button>';
                h+='<button class="btn-danger-outline" onclick="deleteChallenge()">🗑️ Supprimer</button>';
                h+='</div>';
            }
        }
        document.getElementById('challenge-title').textContent='⚔️ '+quizTitle;
        document.getElementById('challenge-content').innerHTML=h;
        document.getElementById('raise-panel').style.display='none';
    }).catch(function(e){console.error(e);document.getElementById('challenge-content').innerHTML='<p style="color:#ef4444;text-align:center">Erreur de chargement</p>';});
}

function toggleRaisePanel(){
    var panel=document.getElementById('raise-panel');
    panel.style.display=panel.style.display==='none'?'block':'none';
    if(panel.style.display==='block'){
        document.getElementById('current-my-xp').textContent=currentUserXp;
        var bet=currentUserParticipant?currentUserParticipant.individual_bet||0:xpPool;
        document.getElementById('current-my-bet').textContent=bet;
    }
}

function raiseBet(){
    var amount=parseInt(document.getElementById('raise-amount').value);
    if(!amount||amount<10){alert('Minimum 10 XP');return;}
    var bet=currentUserParticipant?currentUserParticipant.individual_bet||0:xpPool;
    if(bet+amount>currentUserXp){alert('XP insuffisant pour renchérir de '+amount+' XP. Vous avez '+currentUserXp+' XP et votre mise actuelle est de '+bet+' XP.');return;}
    if(!confirm('Renchérir de '+amount+' XP ? Cette action est irréversible si acceptée par tous.'))return;
    fetch('/api/challenges/raise-bet',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({session_id:challengeId,increase_amount:amount})}).then(function(r){return r.json()}).then(function(d){
        if(d.success){alert('Mise augmentée ! En attente de la réponse des autres participants.');document.getElementById('raise-panel').style.display='none';loadChallenge();}
        else alert(d.error||'Erreur');
    });
}

function respondBet(accept){
    fetch('/api/challenges/bet-respond',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({session_id:challengeId,accept:accept})}).then(function(r){return r.json()}).then(function(d){
        if(d.success){
            if(accept)alert('Mise acceptée !');
            loadChallenge();
            if(!accept) window.location.href='/defis';
        }
        else alert(d.error||'Erreur');
    });
}

function showInvitePanel(){
    var panel=document.getElementById('invite-panel');
    panel.style.display=panel.style.display==='none'?'block':'none';
    if(panel.style.display==='none')return;
    var el=document.getElementById('invite-friends-list');
    el.innerHTML='<div style="color:#94a3b8;padding:8px">Chargement...</div>';
    fetch('/api/friends?user_id='+currentUserId).then(function(r){return r.json()}).then(function(d){
        if(!d.friends||d.friends.length===0){el.innerHTML='<p style="color:#94a3b8">Aucun ami</p>';return;}
        inviteFriendsData=d.friends;
        var h='';
        d.friends.forEach(function(f){
            h+='<label style="display:flex;align-items:center;gap:10px;padding:8px 0;cursor:pointer;border-bottom:1px solid #1e293b">';
            h+='<input type="checkbox" class="ch-invite-checkbox" value="'+f.id+'">';
            h+='<div style="width:32px;height:32px;border-radius:50%;background:#6366f1;display:flex;align-items:center;justify-content:center;font-weight:700;color:white;font-size:.75rem">'+((f.username||'?')[0].toUpperCase())+'</div>';
            h+='<div><div style="font-weight:600">'+(f.nickname||f.username)+'</div><div style="font-size:.75rem;color:#94a3b8">'+(f.rank||'F')+'</div></div>';
            h+='</label>';
        });
        el.innerHTML=h;
    });
}

function sendInvites(){
    var checks=document.querySelectorAll('.ch-invite-checkbox:checked');
    var ids=[];
    checks.forEach(function(c){ids.push(c.value);});
    if(ids.length===0){alert('Sélectionnez au moins un ami');return;}
    fetch('/api/challenges/invite',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({session_id:challengeId,friend_ids:ids})}).then(function(r){return r.json()}).then(function(d){
        if(d.success){alert(d.invited+' invitation(s) envoyée(s)');document.getElementById('invite-panel').style.display='none';loadChallenge();}
        else alert(d.error||'Erreur');
    });
}

function resendInvite(invId){
    fetch('/api/challenges/resend',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({invitation_id:invId})}).then(function(r){return r.json()}).then(function(d){
        if(d.success){alert('Invitation renvoyée !');loadChallenge();}
        else alert(d.error||'Erreur');
    });
}

function deleteChallenge(){
    if(!confirm('Supprimer ce défi ? Les XP seront remboursées aux participants.'))return;
    fetch('/api/challenges/delete',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({session_id:challengeId})}).then(function(r){return r.json()}).then(function(d){
        if(d.success){window.location.href='/defis';}
        else alert(d.error||'Erreur');
    });
}

function setRewardMode(mode){
    fetch('/api/challenges/set-reward-mode',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({session_id:challengeId,mode:mode})}).then(function(r){return r.json()}).then(function(d){
        if(d.success){document.getElementById('reward-mode-select').value=mode;loadChallenge();}
        else alert(d.error||'Erreur');
    });
}

loadChallenge();
</script>`

	challengeHTML = strings.ReplaceAll(challengeHTML, "__CHALLENGE_ID__", id)
	challengeHTML = strings.ReplaceAll(challengeHTML, "__USER_ID__", user.ID)
	challengeHTML = strings.ReplaceAll(challengeHTML, "__USER_XP__", fmt.Sprintf("%d", user.XP))

	return renderPage(c, "Défi", challengeHTML)
}

func (h *Handler) ChallengeLeaderboard(c *fiber.Ctx) error {
	id := c.Params("id")

	sBody, err := h.db.Select("challenge_sessions",
		fmt.Sprintf("id=eq.%s&select=*,quiz:quiz_id(title,question_count)", id), true)
	if err != nil {
		return c.Redirect("/defis")
	}
	var sessions []map[string]interface{}
	json.Unmarshal(sBody, &sessions)
	if len(sessions) == 0 {
		return c.Redirect("/defis")
	}
	session := sessions[0]

	quizTitle, _ := session["quiz"].(map[string]interface{})["title"].(string)
	if quizTitle == "" {
		quizTitle = "Quiz"
	}
	questionCount := 0
	if qc, ok := session["quiz"].(map[string]interface{})["question_count"].(float64); ok {
		questionCount = int(qc)
	}

	partBody, _ := h.db.Select("challenge_participants",
		fmt.Sprintf("session_id=eq.%s&status=eq.accepted&select=user_id,individual_bet", id), true)
	var parts []map[string]interface{}
	json.Unmarshal(partBody, &parts)

	var userIDs []string
	for _, p := range parts {
		uid := database.DBValue(p["user_id"])
		userIDs = append(userIDs, uid)
	}

	profiles := map[string]map[string]interface{}{}
	if len(userIDs) > 0 {
		pList, _ := h.db.GetProfiles(userIDs)
		for _, p := range pList {
			if pid, ok := p["id"].(string); ok {
				profiles[pid] = p
			}
		}
	}

	scoreBody, _ := h.db.Select("challenge_scores",
		fmt.Sprintf("session_id=eq.%s&select=user_id,correct_count,total_questions,time_taken_ms", id), true)
	var scores []map[string]interface{}
	json.Unmarshal(scoreBody, &scores)

	scoreMap := map[string]map[string]interface{}{}
	for _, s := range scores {
		uid := database.DBValue(s["user_id"])
		scoreMap[uid] = s
	}

	type entry struct {
		uid         string
		correct     int
		total       int
		timeTakenMs int
		played      bool
	}
	var entries []entry
	for _, p := range parts {
		uid := database.DBValue(p["user_id"])
		prof := profiles[uid]
		if prof == nil {
			continue
		}
		sc := scoreMap[uid]
		e := entry{uid: uid, played: sc != nil}
		if sc != nil {
			e.correct = int(sc["correct_count"].(float64))
			e.total = int(sc["total_questions"].(float64))
			e.timeTakenMs = int(sc["time_taken_ms"].(float64))
		}
		entries = append(entries, e)
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].played && !entries[j].played {
			return true
		}
		if !entries[i].played && entries[j].played {
			return false
		}
		if entries[i].played && entries[j].played {
			if entries[i].correct != entries[j].correct {
				return entries[i].correct > entries[j].correct
			}
			return entries[i].timeTakenMs < entries[j].timeTakenMs
		}
		return false
	})

	totalPool := database.DBInt(session["total_xp_pool"])
	rewardMode, _ := session["reward_mode"].(string)

	var rewardPcts []int
	isFriendship := rewardMode == "friendship" && len(entries) >= 3
	if isFriendship {
		rewardPcts = []int{50, 35, 15}
	} else {
		rewardPcts = []int{100}
	}

	motivationalMsgs := []string{
		"a dompté ce défi comme un véritable héros ! 🌟",
		"sa puissance dépasse l'entendement ! ⚡",
		"est devenu le maître incontesté du quiz ! 👑",
		"a prouvé qu'il mérite le titre de champion ! 🏆",
		"a atteint un nouveau palier de puissance ! 🔥",
		"sa détermination est légendaire ! 💪",
		"a écrasé la concurrence sans pitié ! ⚔️",
		"son savoir est aussi vaste que l'océan ! 🌊",
	}

	getEntryName := func(uid string) string {
		prof := profiles[uid]
		if prof != nil {
			if n, ok := prof["nickname"].(string); ok && n != "" {
				return n
			}
			if u, ok := prof["username"].(string); ok {
				return u
			}
		}
		return "?"
	}

	// --- Podium (friendship only) ---
	var topSection string
	if isFriendship {
		top3 := entries[:min(3, len(entries))]
		podiumIcons := []string{"🥇", "🥈", "🥉"}
		podiumY := []string{"0", "30px", "30px"}
		podiumBorders := []string{"2px solid #fbbf24", "1px solid #94a3b8", "1px solid #b45309"}
		podiumBg := []string{"#1a1a3e", "#16213e", "#16213e"}
		pods := ""
		for pi, pos := range []int{0, 1, 2} {
			if pos < len(top3) && top3[pos].played {
				pName := getEntryName(top3[pos].uid)
				entryReward := 0
				if pos < len(rewardPcts) && totalPool > 0 {
					entryReward = totalPool * rewardPcts[pos] / 100
				}
				pods += fmt.Sprintf(`
<div style="flex:1;text-align:center;padding:16px 8px;background:%s;border:%s;border-radius:12px;margin-top:%s">
    <div style="font-size:2rem">%s</div>
    <div style="font-weight:700;font-size:.9rem;margin:4px 0">%s</div>
    <div style="color:#22c55e;font-size:.85rem;font-weight:600">+%d XP</div>
</div>`, podiumBg[pi], podiumBorders[pi], podiumY[pi], podiumIcons[pi], pName, entryReward)
			}
		}
		topSection = fmt.Sprintf(`
<div style="display:flex;justify-content:center;gap:10px;margin-bottom:16px">%s
</div>`, pods)
	}

	// --- Winner banner ---
	var banner string
	if len(entries) > 0 && entries[0].played {
		winnerName := getEntryName(entries[0].uid)
		winnerAmount := 0
		if totalPool > 0 && len(rewardPcts) > 0 {
			winnerAmount = totalPool * rewardPcts[0] / 100
		}
		randMsg := motivationalMsgs[rand.Intn(len(motivationalMsgs))]
		modeLabel := "⚔️ All for one"
		if isFriendship {
			modeLabel = "💪 Pouvoir de l'amitié"
		}
		banner = fmt.Sprintf(`
<div style="background:linear-gradient(135deg,#1a1a3e,#0f172a);border:2px solid #fbbf24;border-radius:12px;padding:24px;margin-bottom:16px;text-align:center">
    <div style="font-size:3rem;margin-bottom:4px">👑</div>
    <div style="font-size:1.3rem;font-weight:700;color:#fbbf24;margin-bottom:4px">%s</div>
    <div style="color:#94a3b8;font-size:.9rem;margin-bottom:8px">%s</div>
    <div style="font-size:1.6rem;font-weight:700;color:#22c55e">+%d XP</div>
    <div style="color:#94a3b8;font-size:.75rem;margin-top:8px">%s — Cagnotte: %d XP</div>
</div>`, winnerName, randMsg, winnerAmount, modeLabel, totalPool)
		topSection += banner
	}

	// --- Rows ---
	var rows string
	for ri, e := range entries {
		rank := ri + 1
		rankStyle := ""
		rankIcon := ""
		if rank == 1 {
			rankStyle = "color:#fbbf24"
			rankIcon = "🥇"
		} else if rank == 2 {
			rankStyle = "color:#94a3b8"
			rankIcon = "🥈"
		} else if rank == 3 {
			rankStyle = "color:#b45309"
			rankIcon = "🥉"
		}
		name := getEntryName(e.uid)
		var scoreStr string
		var rewardStr string
		var statsStr string
		if e.played {
			seconds := e.timeTakenMs / 1000
			mins := seconds / 60
			secs := seconds % 60
			timeStr := fmt.Sprintf("%d:%02d", mins, secs)
			scoreStr = fmt.Sprintf("%d / %d bonnes réponses — ⏱ %s", e.correct, e.total, timeStr)
			entryReward := 0
			if ri < len(rewardPcts) && totalPool > 0 {
				entryReward = totalPool * rewardPcts[ri] / 100
			}
			rewardStr = fmt.Sprintf(` <span style="color:#22c55e">+%d XP</span>`, entryReward)
		} else {
			scoreStr = "❌ Pas encore joué"
			rewardStr = ""
		}
		prof := profiles[e.uid]
		if prof != nil {
			played := database.DBInt(prof["challenges_played"])
			won := database.DBInt(prof["challenges_won"])
			if played > 0 {
				statsStr = fmt.Sprintf(` <span style="color:#6366f1;font-size:.8rem">🏅 %d/%d</span>`, won, played)
			}
		}
		av := strings.ToUpper(name[:1])
		rows += fmt.Sprintf(`<div class="cl-row"><div class="cl-rank" style="%s">%s%d</div><div style="width:40px;height:40px;border-radius:50%%;background:#6366f1;display:flex;align-items:center;justify-content:center;font-weight:700;color:white">%s</div><div style="flex:1"><div class="cl-name">%s%s</div><div class="cl-score">%s%s</div></div></div>`, rankStyle, rankIcon, rank, av, name, statsStr, scoreStr, rewardStr)
	}

	if rows == "" {
		rows = `<p style="color:#94a3b8;text-align:center;padding:20px">Aucun participant</p>`
	}

	return renderPage(c, "Classement du défi", fmt.Sprintf(`
<style>
.cl-page{max-width:600px;margin:0 auto}
.cl-card{background:#16213e;border:1px solid #2d2d44;border-radius:12px;padding:24px}
.cl-header{display:flex;align-items:center;gap:12px;margin-bottom:16px}
.cl-quiz-info{color:#94a3b8;font-size:.85rem;margin-bottom:16px}
.cl-row{display:flex;align-items:center;gap:12px;padding:12px 0;border-bottom:1px solid #1e293b}
.cl-row:last-child{border-bottom:none}
.cl-rank{font-size:1.1rem;font-weight:700;min-width:36px;text-align:center}
.cl-name{font-weight:600}
.cl-score{font-size:.85rem;color:#94a3b8}
</style>
<div class="cl-page">
    <a href="/challenges/%s" style="color:#6366f1;display:inline-block;margin-bottom:16px;">← Retour au défi</a>
    <div class="cl-card">
        <div class="cl-header">
            <span style="font-size:2rem">🏆</span>
            <h1 style="font-size:1.3rem">Classement du défi</h1>
        </div>
        <div class="cl-quiz-info">%s — %d questions</div>
        %s
        <div>%s</div>
    </div>
</div>`, id, quizTitle, questionCount, topSection, rows))
}

func (h *Handler) Profile(c *fiber.Ctx) error {
	user := c.Locals("user").(*UserProfile)

	displayName := getDisplayName(user)
	bio := deref(user.Bio)
	country := deref(user.Country)
	favAnime := deref(user.FavoriteAnime)
	avatarURL := deref(user.AvatarURL)

	avatarHTML := fmt.Sprintf(`<div class="pv-avatar-initial">%s</div>`, getInitials(displayName))
	if avatarURL != "" {
		avatarHTML = fmt.Sprintf(`<img src="%s" alt="%s" class="pv-avatar-img">`, htmlAttr(avatarURL), htmlAttr(displayName))
	}

	rankColor := getRankColor(user.Rank)

	infoRows := ""
	if bio != "" {
		infoRows += fmt.Sprintf(`<div class="pv-row"><span class="pv-label">Bio</span><span>%s</span></div>`, bio)
	}
	if country != "" {
		infoRows += fmt.Sprintf(`<div class="pv-row"><span class="pv-label">Pays</span><span>%s</span></div>`, country)
	}
	if favAnime != "" {
		infoRows += fmt.Sprintf(`<div class="pv-row"><span class="pv-label">Anime préféré</span><span>%s</span></div>`, favAnime)
	}
	infoRows += fmt.Sprintf(`<div class="pv-row"><span class="pv-label">Inscrit depuis</span><span>%s</span></div>`,
		user.CreatedAt.Format("02/01/2006"))

	ranks := h.loadRanks()
	correctRank := getRankForXP(user.XP, ranks)
	if correctRank != user.Rank {
		rankData, _ := json.Marshal(map[string]interface{}{"rank": correctRank})
		h.db.Update("user_profiles", fmt.Sprintf("id=eq.%s", user.ID), rankData, true)
		user.Rank = correctRank
	}
	xpPct := calcXPPercentFromRanks(user.XP, ranks)
	_, nextRank := getXPForNextRank(user.XP, ranks)
	nextRankXP := 0
	if nextRank != "MAX" {
		for _, r := range ranks {
			if r.Label == nextRank {
				nextRankXP = r.XP
				break
			}
		}
	}

	return renderPage(c, "Profil", fmt.Sprintf(`
<style>
.pv-page{max-width:700px;margin:0 auto;padding:16px}
.pv-card{background:#16213e;border:1px solid #2d2d44;border-radius:12px;padding:24px;margin-bottom:20px}
.pv-top{display:flex;align-items:center;gap:20px;margin-bottom:20px}
.pv-avatar-initial{width:80px;height:80px;border-radius:50%%;background:linear-gradient(135deg,%s,#6366f1);display:flex;align-items:center;justify-content:center;font-size:2rem;font-weight:800;color:white;flex-shrink:0}
.pv-avatar-img{width:80px;height:80px;border-radius:50%%;object-fit:cover;flex-shrink:0}
.pv-name{font-size:1.5rem;font-weight:700;margin:0 0 6px 0}
.pv-meta{display:flex;align-items:center;gap:10px;margin-bottom:4px}
.pv-details{color:#94a3b8;font-size:0.85rem}
.pv-row{display:flex;justify-content:space-between;padding:8px 0;border-bottom:1px solid #1e293b}
.pv-label{color:#94a3b8;font-size:0.85rem}
.pv-xp-bar{height:8px;background:#1e293b;border-radius:4px;overflow:hidden;margin-top:8px}
.pv-xp-fill{height:100%%;background:linear-gradient(90deg,%s,#6366f1);border-radius:4px;transition:width 0.3s}
.pv-xp-text{font-size:0.75rem;color:#94a3b8;margin-top:4px}
.pv-stats{display:grid;grid-template-columns:repeat(4,1fr);gap:12px;margin-top:16px}
.pv-stat{text-align:center;padding:12px;background:#0f172a;border-radius:8px}
.pv-stat-val{font-size:1.3rem;font-weight:700;color:%s}
.pv-stat-label{font-size:0.7rem;color:#94a3b8;margin-top:2px}
</style>
<div class="pv-page">
    <div class="pv-card">
        <div class="pv-top">
            %s
            <div style="flex:1">
                <h1 class="pv-name">%s</h1>
                <div class="pv-meta">
                    <span style="display:inline-block;padding:2px 10px;border-radius:6px;font-size:0.8rem;font-weight:700;background:%s22;color:%s;border:1px solid %s44">%s</span>
                    <span style="color:#94a3b8;font-size:0.85rem">Niveau %d</span>%s
                    %s
                </div>
                <div class="pv-details">⭐ %d XP total</div>
            </div>
            <a href="/profile/edit" class="btn-primary btn-sm">✏️ Modifier</a>
        </div>
        <div class="pv-xp-bar"><div class="pv-xp-fill" style="width:%d%%"></div></div>
        <div class="pv-xp-text">%d / %d XP pour le prochain rang</div>
    </div>

    %s

    <div class="pv-card">
        <h2 style="font-size:1.1rem;margin-bottom:12px">🔥 Série de connexion</h2>
        <div class="pv-stats" style="grid-template-columns:repeat(2,1fr)">
            <div class="pv-stat">
                <div class="pv-stat-val">%d</div>
                <div class="pv-stat-label">Jours consécutifs</div>
            </div>
            <div class="pv-stat">
                <div class="pv-stat-val">%d</div>
                <div class="pv-stat-label">Record</div>
            </div>
        </div>
    </div>

    <div style="text-align:center;margin-top:8px">
        <a href="/badges" class="btn-ghost btn-sm" style="margin-right:8px">🏅 Mes badges</a>
        <a href="/collections" class="btn-ghost btn-sm">📚 Mes collections</a>
    </div>
</div>
	`, rankColor,
		rankColor,
		rankColor,
		avatarHTML, displayName,
		rankColor, rankColor, rankColor, user.Rank, user.Level,
		challengeStatsBadge(user.ChallengesPlayed, user.ChallengesWon),
		func() string {
			if user.IsPremium {
				return `<span style="color:#fbbf24;font-size:0.75rem">⭐ Premium</span>`
			}
			return ""
		}(),
		user.XP,
		xpPct, user.XP, nextRankXP,
		func() string {
			if infoRows != "" {
				return fmt.Sprintf(`<div class="pv-card"><h2 style="font-size:1.1rem;margin-bottom:12px">📋 Informations</h2>%s</div>`, infoRows)
			}
			return ""
		}(),
		user.CurrentStreak, user.LongestStreak))
}

func (h *Handler) ProfileEdit(c *fiber.Ctx) error {
	user := c.Locals("user").(*UserProfile)

	nickname := deref(user.Nickname)
	bio := deref(user.Bio)
	country := deref(user.Country)
	favAnime := deref(user.FavoriteAnime)
	phone := deref(user.Phone)
	avatarURL := deref(user.AvatarURL)

	avatarHTML := fmt.Sprintf(`<div class="pv-avatar-initial" style="width:80px;height:80px;font-size:2rem">%s</div>`, getInitials(getDisplayName(user)))
	if avatarURL != "" {
		avatarHTML = fmt.Sprintf(`<img src="%s" style="width:80px;height:80px;border-radius:50%%;object-fit:cover">`, htmlAttr(avatarURL))
	}

	countries := []string{
		"Algérie", "Angola", "Bénin", "Botswana", "Burkina Faso", "Burundi",
		"Cabo Verde", "Cameroun", "Centrafrique", "Comores", "Congo-Brazzaville", "Congo-RDC",
		"Côte d'Ivoire", "Djibouti", "Égypte", "Érythrée", "Eswatini", "Éthiopie",
		"Gabon", "Gambie", "Ghana", "Guinée", "Guinée-Bissau", "Guinée équatoriale",
		"Kenya", "Lesotho", "Liberia", "Libye", "Madagascar", "Malawi",
		"Mali", "Maroc", "Maurice", "Mauritanie", "Mozambique", "Namibie",
		"Niger", "Nigeria", "Ouganda", "Réunion", "Rwanda", "Sahara Occidental",
		"Sao Tomé-et-Principe", "Sénégal", "Seychelles", "Sierra Leone", "Somalie",
		"Soudan", "Soudan du Sud", "Tanzanie", "Tchad", "Togo", "Tunisie",
		"Zambie", "Zimbabwe", "Autre",
	}
	countryOptions := `<option value="">Sélectionner...</option>`
	for _, c := range countries {
		sel := ""
		if c == country {
			sel = " selected"
		}
		countryOptions += fmt.Sprintf(`<option value="%s"%s>%s</option>`, c, sel, c)
	}

	phoneLock := ""
	if phone != "" {
		phoneLock = ` <span style="color:#94a3b8;font-size:0.75rem">(déjà défini)</span>`
	}

	return renderPage(c, "Modifier le profil", fmt.Sprintf(`
<style>
.edit-form{max-width:650px;margin:0 auto}
.edit-form label{display:block;font-weight:600;margin-bottom:4px;margin-top:14px;font-size:0.9rem}
.edit-form input,.edit-form textarea,.edit-form select{width:100%%;padding:10px;border-radius:8px;border:1px solid #2d2d44;background:#0f172a;color:white;font-size:0.9rem}
.edit-form textarea{resize:vertical}
.edit-form .hint{font-size:0.7rem;color:#94a3b8;margin-top:2px}
.edit-avatar{display:flex;align-items:center;gap:16px;margin-bottom:20px;padding:16px;background:#0f172a;border-radius:10px}
</style>
<h1 style="margin-bottom:24px">✏️ Modifier le profil</h1>
<div class="card edit-form" style="padding:24px">
    <form method="POST" action="/profile/edit">
        <div class="edit-avatar">
            <div id="avatar-preview">%s</div>
            <div>
                <div style="font-weight:600">%s%s</div>
                <input type="file" id="avatar-file" accept="image/*" style="display:none" onchange="uploadAvatar(this)">
                <button type="button" class="btn-sm btn-outline" onclick="document.getElementById('avatar-file').click()">📷 Changer la photo</button>
            </div>
        </div>

        <label>Nom d'utilisateur</label>
        <input type="text" value="%s" disabled style="opacity:0.5">
        <div class="hint">Le nom d'utilisateur ne peut pas être modifié</div>

        <label>Surnom / Nickname</label>
        <input type="text" name="nickname" value="%s" placeholder="Votre surnom" maxlength="30">
        <div class="hint">2-30 caractères. Si défini, sera affiché à la place de votre nom</div>

        <label>Bio</label>
        <textarea name="bio" rows="3" placeholder="Parlez de vous..." maxlength="200">%s</textarea>
        <div class="hint">Maximum 200 caractères</div>

        <label>Pays</label>
        <select name="country">%s</select>

        <label>Anime préféré</label>
        <input type="text" name="favorite_anime" value="%s" placeholder="Ex: Naruto, One Piece...">

        <label>Téléphone %s</label>
        <input type="tel" name="phone" value="%s" placeholder="+226 70 00 00 00" %s>
        <div class="hint">Format international (+XXX suivi des chiffres)</div>

        <div style="display:flex;gap:10px;margin-top:24px;justify-content:flex-end">
            <a href="/profil" class="btn-ghost btn-sm">Annuler</a>
            <button type="submit" class="btn-primary btn-sm">💾 Sauvegarder</button>
        </div>
    </form>
</div>
<script>
function uploadAvatar(input){
    if(!input.files||!input.files[0])return;
    var fd=new FormData();
    fd.append('file',input.files[0]);
    fetch('/api/quiz/upload-image',{method:'POST',body:fd}).then(function(r){return r.json()}).then(function(d){
        if(d.success&&d.url){
            document.getElementById('avatar-preview').innerHTML='<img src="'+d.url+'" style="width:80px;height:80px;border-radius:50%;object-fit:cover">';
        } else alert(d.error||'Erreur upload');
    });
}
</script>
	`, avatarHTML, getDisplayName(user), challengeStatsBadge(user.ChallengesPlayed, user.ChallengesWon),
		user.Username, nickname, bio,
		countryOptions, favAnime,
		phoneLock, phone,
		func() string {
			if phone != "" {
				return "disabled"
			}
			return ""
		}()))
}

func (h *Handler) ProfileUpdate(c *fiber.Ctx) error {
	user := c.Locals("user").(*UserProfile)

	nickname := c.FormValue("nickname")
	bio := c.FormValue("bio")
	country := c.FormValue("country")
	favAnime := c.FormValue("favorite_anime")
	phone := c.FormValue("phone")

	updates := map[string]interface{}{
		"updated_at": time.Now().UTC().Format(time.RFC3339),
	}

	if nickname != "" {
		updates["nickname"] = nickname
	} else {
		updates["nickname"] = nil
	}

	if bio != "" {
		updates["bio"] = bio
	} else {
		updates["bio"] = nil
	}

	if country != "" {
		updates["country"] = country
	} else {
		updates["country"] = nil
	}

	if favAnime != "" {
		updates["favorite_anime"] = favAnime
	} else {
		updates["favorite_anime"] = nil
	}

	if phone != "" {
		updates["phone"] = phone
	}

	jsonData, _ := json.Marshal(updates)
	_, err := h.db.Update("user_profiles", fmt.Sprintf("id=eq.%s", user.ID), jsonData, true)
	if err != nil {
		return c.Redirect("/profile/edit")
	}

	return c.Redirect("/profil")
}

func (h *Handler) Leaderboard(c *fiber.Ctx) error {
	body, err := h.db.Select("user_profiles",
		"order=xp.desc&limit=50&select=id,username,nickname,rank,level,xp", true)
	if err != nil {
		return renderPage(c, "Classement", `<h1 style="margin-bottom:24px">🏆 Classement</h1><p class="text-muted text-center" style="padding:40px">Erreur de chargement</p>`)
	}
	var users []map[string]interface{}
	json.Unmarshal(body, &users)

	rows := ""
	for i, u := range users {
		username, _ := u["username"].(string)
		nickname, _ := u["nickname"].(string)
		rank, _ := u["rank"].(string)
		xp, _ := u["xp"].(float64)
		level, _ := u["level"].(float64)
		uid, _ := u["id"].(string)

		displayName := username
		if nickname != "" {
			displayName = nickname
		}

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
<tr onclick="window.location.href='/profile/%s'" style="cursor:pointer">
    <td class="rank-cell">%s</td>
    <td>
        <div class="user-info">
            <span class="username">%s</span>
            <span class="user-rank">%s</span>
        </div>
    </td>
    <td class="xp-cell">%d XP</td>
    <td class="level-cell">Niv. %d</td>
</tr>`, uid, medal, displayName, rank, int(xp), int(level))
	}

	if rows == "" {
		rows = `<tr><td colspan="4" class="text-muted text-center" style="padding:24px">Aucun joueur classé</td></tr>`
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
    <button class="btn-primary" onclick="document.getElementById('create-quiz-form').style.display=document.getElementById('create-quiz-form').style.display==='none'?'block':'none'">+ Créer</button>
</div>

<div id="create-quiz-form" style="display:none;margin-bottom:24px;">
    <div class="card" style="padding:24px;">
        <h2 style="margin-bottom:16px;">Créer un quiz officiel</h2>
        <form id="quiz-form" style="display:flex;flex-direction:column;gap:12px;">
            <div style="display:grid;grid-template-columns:1fr 1fr;gap:12px;">
                <div>
                    <label style="display:block;margin-bottom:4px;font-weight:600;">Titre *</label>
                    <input type="text" id="q-title" required placeholder="Titre du quiz" style="width:100%%;padding:10px;border-radius:8px;border:1px solid #2d2d44;background:#0f172a;color:white;">
                </div>
                <div>
                    <label style="display:block;margin-bottom:4px;font-weight:600;">Série *</label>
                    <input type="text" id="q-series" required placeholder="Ex: Naruto" style="width:100%%;padding:10px;border-radius:8px;border:1px solid #2d2d44;background:#0f172a;color:white;">
                </div>
            </div>
            <div>
                <label style="display:block;margin-bottom:4px;font-weight:600;">Description</label>
                <textarea id="q-description" rows="2" placeholder="Description..." style="width:100%%;padding:10px;border-radius:8px;border:1px solid #2d2d44;background:#0f172a;color:white;"></textarea>
            </div>
            <div style="display:grid;grid-template-columns:1fr 1fr;gap:12px;">
                <div>
                    <label style="display:block;margin-bottom:4px;font-weight:600;">Catégorie</label>
                    <input type="text" id="q-category" value="Shonen" style="width:100%%;padding:10px;border-radius:8px;border:1px solid #2d2d44;background:#0f172a;color:white;">
                </div>
                <div>
                    <label style="display:block;margin-bottom:4px;font-weight:600;">Sous-catégorie</label>
                    <input type="text" id="q-subcategory" value="Général" style="width:100%%;padding:10px;border-radius:8px;border:1px solid #2d2d44;background:#0f172a;color:white;">
                </div>
            </div>
            <div style="display:grid;grid-template-columns:1fr 1fr;gap:12px;">
                <div>
                    <label style="display:block;margin-bottom:4px;font-weight:600;">Date de début *</label>
                    <input type="datetime-local" id="q-starts" required style="width:100%%;padding:10px;border-radius:8px;border:1px solid #2d2d44;background:#0f172a;color:white;">
                </div>
                <div>
                    <label style="display:block;margin-bottom:4px;font-weight:600;">Date de fin *</label>
                    <input type="datetime-local" id="q-ends" required style="width:100%%;padding:10px;border-radius:8px;border:1px solid #2d2d44;background:#0f172a;color:white;">
                </div>
            </div>
            <div style="display:grid;grid-template-columns:1fr 1fr;gap:12px;">
                <div>
                    <label style="display:block;margin-bottom:4px;font-weight:600;">Durée (secondes)</label>
                    <input type="number" id="q-duration" value="30" min="5" style="width:100%%;padding:10px;border-radius:8px;border:1px solid #2d2d44;background:#0f172a;color:white;">
                </div>
                <div>
                    <label style="display:block;margin-bottom:4px;font-weight:600;">Mode de durée</label>
                    <select id="q-mode" style="width:100%%;padding:10px;border-radius:8px;border:1px solid #2d2d44;background:#0f172a;color:white;">
                        <option value="per_question">Par question</option>
                        <option value="global">Global</option>
                    </select>
                </div>
            </div>
            <div style="display:flex;gap:8px;justify-content:flex-end;">
                <button type="button" class="btn-ghost btn-sm" onclick="document.getElementById('create-quiz-form').style.display='none'">Annuler</button>
                <button type="submit" class="btn-primary btn-sm">Créer le quiz</button>
            </div>
        </form>
    </div>
</div>

<div class="leaderboard">
    <table>
        <thead><tr><th>Titre</th><th>Statut</th><th>Actions</th></tr></thead>
        <tbody>%s</tbody>
    </table>
</div>

<script>
document.getElementById('quiz-form').addEventListener('submit',function(e){
    e.preventDefault();
    var data={
        title:document.getElementById('q-title').value,
        description:document.getElementById('q-description').value,
        series:document.getElementById('q-series').value,
        category:document.getElementById('q-category').value,
        subcategory:document.getElementById('q-subcategory').value,
        starts_at:document.getElementById('q-starts').value,
        ends_at:document.getElementById('q-ends').value,
        duration_seconds:parseInt(document.getElementById('q-duration').value)||30,
        duration_mode:document.getElementById('q-mode').value,
        rewards:[]
    };
    fetch('/api/admin/official-quizzes/create',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(data)})
    .then(function(r){return r.json()}).then(function(d){
        if(d.error){alert(d.error);return;}
        window.location.reload();
    }).catch(function(){alert('Erreur réseau');});
});
</script>
	`, quizRows))
}

func (h *Handler) AdminTickets(c *fiber.Ctx) error {
	convID := c.Params("id")

	if convID != "" {
		return h.adminTicketDetail(c, convID)
	}

	// List view
	body, err := h.db.Select("admin_conversations",
		"order=created_at.desc&select=*", true)
	if err != nil {
		return renderPage(c, "Tickets", `<h1 style="margin-bottom:24px">🎧 Tickets Support</h1><p style="color:#94a3b8;text-align:center;padding:40px">Erreur de chargement</p>`)
	}
	var tickets []map[string]interface{}
	json.Unmarshal(body, &tickets)

	ticketRows := ""
	for _, t := range tickets {
		subject, _ := t["subject"].(string)
		status, _ := t["status"].(string)
		id, _ := t["id"].(string)
		userID, _ := t["user_id"].(string)

		username := "—"
		if userID != "" {
			pBody, err := h.db.Select("user_profiles",
				fmt.Sprintf("id=eq.%s&select=username,nickname", userID), true)
			if err == nil {
				var profiles []map[string]interface{}
				json.Unmarshal(pBody, &profiles)
				if len(profiles) > 0 {
					if nn, ok := profiles[0]["nickname"].(string); ok && nn != "" {
						username = nn
					} else if un, ok := profiles[0]["username"].(string); ok && un != "" {
						username = un
					}
				}
			}
		}

		badgeClass := "badge-open"
		if status == "assigned" {
			badgeClass = "badge-assigned"
		} else if status == "closed" {
			badgeClass = "badge-closed"
		}

		ticketRows += fmt.Sprintf(`
<tr>
    <td><a href="/admin/tickets/%s" style="color:#e2e8f0;text-decoration:none;font-weight:500">%s</a></td>
    <td><a href="/profile/%s" style="color:#6366f1;text-decoration:none">%s</a></td>
    <td><span class="%s">%s</span></td>
    <td style="display:flex;gap:6px">
        <a href="/admin/tickets/%s" class="btn-sm">Voir</a>
        <button class="btn-sm" style="color:#ef4444;background:none;border:none;cursor:pointer" onclick="deleteAdminTicket('%s')">🗑</button>
    </td>
</tr>`, id, subject, userID, username, badgeClass, status, id, id)
	}

	if ticketRows == "" {
		ticketRows = `<tr><td colspan="4" style="text-align:center;color:#94a3b8;padding:24px">Aucun ticket</td></tr>`
	}

	return renderPage(c, "Tickets", fmt.Sprintf(`
<h1 style="margin-bottom:24px">🎧 Tickets Support</h1>
<div class="leaderboard">
    <table>
        <thead><tr><th>Sujet</th><th>Utilisateur</th><th>Statut</th><th>Actions</th></tr></thead>
        <tbody>%s</tbody>
    </table>
</div>
<script>
function deleteAdminTicket(id){
    if(!confirm('Supprimer ce ticket et tous ses messages ?'))return;
    fetch('/api/admin/conversations/delete',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({conversation_id:id})})
    .then(function(r){return r.json()}).then(function(d){
        if(d.error){alert(d.error);return;}
        window.location.reload();
    });
}
</script>
	`, ticketRows))
}

func (h *Handler) adminTicketDetail(c *fiber.Ctx, convID string) error {
	admin := c.Locals("user").(*UserProfile)

	convBody, err := h.db.Select("admin_conversations",
		fmt.Sprintf("id=eq.%s&select=*", convID), true)
	if err != nil {
		return c.Redirect("/admin/tickets")
	}
	var convs []map[string]interface{}
	json.Unmarshal(convBody, &convs)
	if len(convs) == 0 {
		return c.Redirect("/admin/tickets")
	}
	conv := convs[0]
	subject, _ := conv["subject"].(string)
	status, _ := conv["status"].(string)
	userID, _ := conv["user_id"].(string)

	statusLabel := "Ouvert"
	badgeClass := "badge-open"
	if status == "assigned" {
		statusLabel = "Assigné"
		badgeClass = "badge-assigned"
	} else if status == "closed" {
		statusLabel = "Fermé"
		badgeClass = "badge-closed"
	}

	selectedOpen := ""
	selectedAssigned := ""
	selectedClosed := ""
	if status == "open" {
		selectedOpen = " selected"
	} else if status == "assigned" {
		selectedAssigned = " selected"
	} else if status == "closed" {
		selectedClosed = " selected"
	}

	username := "—"
	if userID != "" {
		pBody, pErr := h.db.Select("user_profiles",
			fmt.Sprintf("id=eq.%s&select=username,nickname", userID), true)
		if pErr == nil {
			var profiles []map[string]interface{}
			json.Unmarshal(pBody, &profiles)
			if len(profiles) > 0 {
				if nn, ok := profiles[0]["nickname"].(string); ok && nn != "" {
					username = nn
				} else if un, ok := profiles[0]["username"].(string); ok && un != "" {
					username = un
				}
			}
		}
	}

	return renderPage(c, "Ticket - "+subject, fmt.Sprintf(`
<div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:16px">
    <a href="/admin/tickets" style="color:#6366f1;text-decoration:none">← Retour aux tickets</a>
    <div style="display:flex;align-items:center;gap:10px">
        <span style="color:#94a3b8">Utilisateur : <a href="/profile/%s" style="color:#6366f1;text-decoration:none">%s</a></span>
        <span class="badge-%s">%s</span>
    </div>
</div>
<div class="support-conv" style="max-width:700px;margin:0 auto">
    <div class="conv-header" style="display:flex;justify-content:space-between;align-items:center;margin-bottom:16px;padding:12px 16px;background:#16213e;border-radius:8px;border:1px solid #2d2d44">
        <h1 style="font-size:1.3rem;margin:0">%s</h1>
        <div style="display:flex;align-items:center;gap:10px">
            <select id="status-select" style="padding:6px 10px;border-radius:6px;border:1px solid #2d2d44;background:#0f172a;color:white;font-size:0.85rem">
                <option value="open"%s>Ouvert</option>
                <option value="assigned"%s>Assigné</option>
                <option value="closed"%s>Fermé</option>
            </select>
            <button class="btn-ghost btn-sm" style="color:#ef4444;background:none;border:none;cursor:pointer" onclick="deleteTicket()">🗑 Supprimer</button>
        </div>
    </div>
    <div id="messages-list" class="messages-list" style="display:flex;flex-direction:column;gap:8px;max-height:60vh;overflow-y:auto;padding:12px;background:#0f172a;border:1px solid #2d2d44;border-radius:8px;margin-bottom:12px">
        <div style="text-align:center;color:#94a3b8;padding:20px">Chargement...</div>
    </div>
    <form id="msg-form" style="display:flex;gap:8px">
        <input type="text" id="msg-input" placeholder="Écrire un message..." autocomplete="off" style="flex:1;padding:10px;border-radius:8px;border:1px solid #2d2d44;background:#0f172a;color:white">
        <button type="submit" class="btn-primary btn-sm">Envoyer</button>
    </form>
</div>
<style>
.msg-bubble{max-width:72%%;padding:8px 12px 6px;border-radius:10px;font-size:0.88rem;line-height:1.4;position:relative;margin-bottom:2px}
.msg-me{align-self:flex-end;background:#005c4b;border-bottom-right-radius:4px}
.msg-them{align-self:flex-start;background:#1e2a38;border-bottom-left-radius:4px}
.msg-text{word-wrap:break-word;margin-bottom:2px}
.msg-meta{display:flex;align-items:center;justify-content:flex-end;gap:4px;margin-top:0}
.msg-time{font-size:0.65rem;color:rgba(255,255,255,0.45)}
.msg-check{font-size:0.7rem;color:#53bdeb}
.msg-system{text-align:center;color:#94a3b8;font-size:0.78rem;padding:8px 16px;background:rgba(255,255,255,0.04);border-radius:8px;margin:4px 0;align-self:center}
#messages-list{scroll-behavior:smooth}
</style>
<script>
var convId='%s';
var currentStatus='%s';
function loadMessages(){
    fetch('/api/admin/messages?conversation_id='+convId).then(function(r){return r.json()}).then(function(d){
        var el=document.getElementById('messages-list');
        if(!d.messages||d.messages.length===0){el.innerHTML='<div class="msg-system">Aucun message</div>';return;}
        var h='';d.messages.forEach(function(m){
            var isSystem=m.sender_id==='system';
            if(isSystem){
                h+='<div class="msg-system">'+escapeMsg(m.content)+'</div>';
            } else {
                var isMe=m.sender_id==='%s';
                var cls=isMe?'msg-me':'msg-them';
                var time=m.created_at?fmtTime(new Date(m.created_at)):'';
                var check=isMe?'<span class="msg-check">\u2713\u2713</span>':'';
                h+='<div class="msg-bubble '+cls+'"><div class="msg-text">'+escapeMsg(m.content)+'</div><div class="msg-meta"><span class="msg-time">'+time+'</span>'+check+'</div></div>';
            }
        });
        el.innerHTML=h;
        el.scrollTop=el.scrollHeight;
    });
}
function fmtTime(d){var h=d.getHours();var m=d.getMinutes();return(h<10?'0':'')+h+':'+(m<10?'0':'')+m;}
function escapeMsg(t){if(!t)return'';var d=document.createElement('div');d.appendChild(document.createTextNode(t));return d.innerHTML;}
document.getElementById('msg-form').addEventListener('submit',function(e){
    e.preventDefault();
    var input=document.getElementById('msg-input');
    var content=input.value.trim();
    if(!content)return;
    input.value='';
    fetch('/api/admin/messages/send',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({conversation_id:convId,content:content})})
    .then(function(r){return r.json()}).then(function(){loadMessages();});
});
document.getElementById('status-select').addEventListener('change',function(e){
    currentStatus=e.target.value;
    fetch('/api/admin/conversations/update-status',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({conversation_id:convId,status:currentStatus})})
    .then(function(r){return r.json()}).then(function(d){
        if(d.error){alert(d.error);return;}
    });
});
function deleteTicket(){
    if(!confirm('Supprimer ce ticket et tous ses messages ?'))return;
    fetch('/api/admin/conversations/delete',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({conversation_id:convId})})
    .then(function(r){return r.json()}).then(function(d){
        if(d.error){alert(d.error);return;}
        window.location.href='/admin/tickets';
    });
}
loadMessages();
setInterval(loadMessages,5000);
</script>
	`, userID, username, badgeClass, statusLabel, subject, selectedOpen, selectedAssigned, selectedClosed, convID, status, admin.ID))
}

func (h *Handler) AdminAnnouncements(c *fiber.Ctx) error {
	return renderPage(c, "Annonces", `
<div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px;">
    <h1>📢 Annonces</h1>
    <button class="btn-primary" onclick="var f=document.getElementById('create-form');f.style.display=f.style.display==='none'?'block':'none'">+ Créer</button>
</div>

<div id="create-form" style="display:none;margin-bottom:24px;">
    <div class="card" style="padding:24px;">
        <h2 style="margin-bottom:16px;">Créer une annonce</h2>
        <form id="announcement-form" style="display:flex;flex-direction:column;gap:12px;">
            <div>
                <label style="display:block;margin-bottom:4px;font-weight:600;">Titre *</label>
                <input type="text" id="a-title" required placeholder="Titre de l'annonce" style="width:100%;padding:10px;border-radius:8px;border:1px solid #2d2d44;background:#0f172a;color:white;">
            </div>
            <div>
                <label style="display:block;margin-bottom:4px;font-weight:600;">Description</label>
                <textarea id="a-description" rows="3" placeholder="Description..." style="width:100%;padding:10px;border-radius:8px;border:1px solid #2d2d44;background:#0f172a;color:white;"></textarea>
            </div>
            <div>
                <label style="display:block;margin-bottom:4px;font-weight:600;">URL image</label>
                <input type="text" id="a-image" placeholder="https://..." style="width:100%;padding:10px;border-radius:8px;border:1px solid #2d2d44;background:#0f172a;color:white;">
            </div>
            <div>
                <label style="display:block;margin-bottom:4px;font-weight:600;">Type</label>
                <select id="a-type" style="width:100%;padding:10px;border-radius:8px;border:1px solid #2d2d44;background:#0f172a;color:white;">
                    <option value="quiz">Quiz</option>
                    <option value="event">Événement</option>
                    <option value="news">Actualité</option>
                </select>
            </div>
            <div>
                <label style="display:block;margin-bottom:4px;font-weight:600;">Date de fin</label>
                <input type="datetime-local" id="a-ends" style="width:100%;padding:10px;border-radius:8px;border:1px solid #2d2d44;background:#0f172a;color:white;">
            </div>
            <div style="display:flex;gap:8px;justify-content:flex-end;">
                <button type="button" class="btn-ghost btn-sm" onclick="document.getElementById('create-form').style.display='none'">Annuler</button>
                <button type="submit" class="btn-primary btn-sm">Créer</button>
            </div>
        </form>
    </div>
</div>

<div id="edit-form" style="display:none;margin-bottom:24px;">
    <div class="card" style="padding:24px;">
        <h2 style="margin-bottom:16px;">Modifier l'annonce</h2>
        <form id="edit-announcement-form" style="display:flex;flex-direction:column;gap:12px;">
            <input type="hidden" id="e-id">
            <div>
                <label style="display:block;margin-bottom:4px;font-weight:600;">Titre *</label>
                <input type="text" id="e-title" required style="width:100%;padding:10px;border-radius:8px;border:1px solid #2d2d44;background:#0f172a;color:white;">
            </div>
            <div>
                <label style="display:block;margin-bottom:4px;font-weight:600;">Description</label>
                <textarea id="e-description" rows="3" style="width:100%;padding:10px;border-radius:8px;border:1px solid #2d2d44;background:#0f172a;color:white;"></textarea>
            </div>
            <div>
                <label style="display:block;margin-bottom:4px;font-weight:600;">URL image</label>
                <input type="text" id="e-image" style="width:100%;padding:10px;border-radius:8px;border:1px solid #2d2d44;background:#0f172a;color:white;">
            </div>
            <div>
                <label style="display:block;margin-bottom:4px;font-weight:600;">Type</label>
                <select id="e-type" style="width:100%;padding:10px;border-radius:8px;border:1px solid #2d2d44;background:#0f172a;color:white;">
                    <option value="quiz">Quiz</option>
                    <option value="event">Événement</option>
                    <option value="news">Actualité</option>
                </select>
            </div>
            <div>
                <label style="display:block;margin-bottom:4px;font-weight:600;">Statut</label>
                <select id="e-status" style="width:100%;padding:10px;border-radius:8px;border:1px solid #2d2d44;background:#0f172a;color:white;">
                    <option value="active">Actif</option>
                    <option value="scheduled">Programmé</option>
                    <option value="expired">Expiré</option>
                </select>
            </div>
            <div>
                <label style="display:block;margin-bottom:4px;font-weight:600;">Date de fin</label>
                <input type="datetime-local" id="e-ends" style="width:100%;padding:10px;border-radius:8px;border:1px solid #2d2d44;background:#0f172a;color:white;">
            </div>
            <div style="display:flex;gap:8px;justify-content:flex-end;">
                <button type="button" class="btn-ghost btn-sm" onclick="document.getElementById('edit-form').style.display='none'">Annuler</button>
                <button type="submit" class="btn-primary btn-sm">Enregistrer</button>
            </div>
        </form>
    </div>
</div>

<div id="announcements-list">
    <p style="color: #94a3b8; text-align: center; padding: 40px;">Chargement...</p>
</div>

<script>
var allAnnouncements=[];
function loadAnnouncements(){
    fetch('/api/admin/announcements').then(function(r){
        if(!r.ok) throw new Error('HTTP '+r.status);
        return r.json();
    }).then(function(data){
        var el=document.getElementById('announcements-list');
        if(!data||!Array.isArray(data)||data.length===0){
            el.innerHTML='<div class="card"><p style="color:#94a3b8;text-align:center;padding:40px;">Aucune annonce</p></div>';
            return;
        }
        allAnnouncements=data;
        var h='';
        for(var i=0;i<data.length;i++){
            var a=data[i];
            var status=a.status||'active';
            var type=a.type||'quiz';
            h+='<div class="card" style="padding:16px;margin-bottom:8px;display:flex;justify-content:space-between;align-items:center;">';
            h+='<div><h3 style="margin:0;">'+escapeHtml(a.title)+'</h3>';
            h+='<p style="color:#94a3b8;font-size:0.85rem;margin:4px 0;">'+escapeHtml(a.description||'')+'</p>';
            h+='<span style="font-size:0.75rem;color:#64748b;">'+type+' &middot; '+status+'</span></div>';
            h+='<div style="display:flex;gap:8px;">';
            h+='<button class="btn-ghost btn-sm" data-idx="'+i+'" onclick="editAnnouncement('+i+')">Modifier</button>';
            h+='<button class="btn-ghost btn-sm" style="color:#ef4444;" data-id="'+a.id+'" onclick="deleteAnnouncement(this.getAttribute(\'data-id\'))">Supprimer</button>';
            h+='</div></div>';
        }
        el.innerHTML=h;
    }).catch(function(err){
        console.error('Erreur chargement annonces:',err);
        document.getElementById('announcements-list').innerHTML='<div class="card"><p style="color:#ef4444;text-align:center;padding:40px;">Erreur de chargement</p></div>';
    });
}
function escapeHtml(t){
    if(!t) return '';
    var d=document.createElement('div');
    d.appendChild(document.createTextNode(t));
    return d.innerHTML;
}
function toLocalDatetime(iso){
    if(!iso) return '';
    var d=new Date(iso);
    var pad=function(n){return n<10?'0'+n:n;};
    return d.getFullYear()+'-'+pad(d.getMonth()+1)+'-'+pad(d.getDate())+'T'+pad(d.getHours())+':'+pad(d.getMinutes());
}
document.getElementById('announcement-form').addEventListener('submit',function(e){
    e.preventDefault();
    var payload={
        title:document.getElementById('a-title').value,
        description:document.getElementById('a-description').value,
        image_url:document.getElementById('a-image').value,
        type:document.getElementById('a-type').value,
        ends_at:document.getElementById('a-ends').value
    };
    fetch('/api/admin/announcements/create',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(payload)})
    .then(function(r){
        if(!r.ok) throw new Error('HTTP '+r.status);
        return r.json();
    }).then(function(d){
        if(d.error){alert(d.error);return;}
        document.getElementById('create-form').style.display='none';
        document.getElementById('announcement-form').reset();
        loadAnnouncements();
    }).catch(function(err){
        console.error('Erreur création:',err);
        alert('Erreur lors de la création');
    });
});
function editAnnouncement(idx){
    var a=allAnnouncements[idx];
    if(!a) return;
    document.getElementById('e-id').value=a.id;
    document.getElementById('e-title').value=a.title||'';
    document.getElementById('e-description').value=a.description||'';
    document.getElementById('e-image').value=a.image_url||'';
    document.getElementById('e-type').value=a.type||'quiz';
    document.getElementById('e-status').value=a.status||'active';
    document.getElementById('e-ends').value=toLocalDatetime(a.ends_at);
    document.getElementById('edit-form').style.display='block';
    document.getElementById('create-form').style.display='none';
    window.scrollTo(0,0);
}
document.getElementById('edit-announcement-form').addEventListener('submit',function(e){
    e.preventDefault();
    var payload={
        id:document.getElementById('e-id').value,
        title:document.getElementById('e-title').value,
        description:document.getElementById('e-description').value,
        image_url:document.getElementById('e-image').value,
        type:document.getElementById('e-type').value,
        status:document.getElementById('e-status').value,
        ends_at:document.getElementById('e-ends').value
    };
    fetch('/api/admin/announcements/update',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(payload)})
    .then(function(r){
        if(!r.ok) throw new Error('HTTP '+r.status);
        return r.json();
    }).then(function(d){
        if(d.error){alert(d.error);return;}
        document.getElementById('edit-form').style.display='none';
        loadAnnouncements();
    }).catch(function(err){
        console.error('Erreur modification:',err);
        alert('Erreur lors de la modification');    
    });
});
function deleteAnnouncement(id){
    if(!confirm('Supprimer cette annonce ?'))return;
    fetch('/api/admin/announcements/delete',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({id:id})})
    .then(function(r){
        if(!r.ok) throw new Error('HTTP '+r.status);
        return r.json();
    }).then(function(){loadAnnouncements();})
    .catch(function(err){console.error('Erreur suppression:',err);alert('Erreur suppression');});
}
loadAnnouncements();
</script>
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
