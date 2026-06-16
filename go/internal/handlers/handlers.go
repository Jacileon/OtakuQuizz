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
	return renderPage(c, "Quiz", `<h1>Détail du quiz</h1><p>Page en construction</p>`)
}

func (h *Handler) QuizPlay(c *fiber.Ctx) error {
	return renderPage(c, "Jouer", `<h1>Jouer au quiz</h1><p>Page en construction</p>`)
}

func (h *Handler) QuizSubmit(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"success": true, "xpEarned": 0})
}

func (h *Handler) Friends(c *fiber.Ctx) error {
	user := c.Locals("user").(*UserProfile)
	
	// Récupérer les amis depuis Supabase
	supabaseURL := os.Getenv("SUPABASE_URL")
	if supabaseURL == "" {
		supabaseURL = os.Getenv("NEXT_PUBLIC_SUPABASE_URL")
	}

	accessToken := ""
	// TODO: Get access token from session

	url := fmt.Sprintf("%s/rest/v1/friendships?or=(requester_id.eq.%s,addressee_id.eq.%s)&status=eq.accepted", supabaseURL, user.ID, user.ID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}

	resp, err := http.DefaultClient.Do(req)
	friendships := []map[string]interface{}{}
	if err == nil {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		json.Unmarshal(body, &friendships)
	}

	friendCards := ""
	for _, f := range friendships {
		// Determine friend ID
		requesterID, _ := f["requester_id"].(string)
		friendID := requesterID
		if friendID == user.ID {
			friendID, _ = f["addressee_id"].(string)
		}

		// Get friend profile
		friendURL := fmt.Sprintf("%s/rest/v1/user_profiles?id=eq.%s&select=username,avatar_url,rank,level", supabaseURL, friendID)
		friendReq, _ := http.NewRequest("GET", friendURL, nil)
		friendReq.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
		friendResp, err := http.DefaultClient.Do(friendReq)
		if err == nil {
			defer friendResp.Body.Close()
			friendBody, _ := io.ReadAll(friendResp.Body)
			var friends []map[string]interface{}
			json.Unmarshal(friendBody, &friends)
			if len(friends) > 0 {
				friend := friends[0]
				username, _ := friend["username"].(string)
				rank, _ := friend["rank"].(string)
				level, _ := friend["level"].(float64)

				friendCards += fmt.Sprintf(`
<div class="friend-card">
    <div class="friend-avatar">%s</div>
    <div class="friend-info">
        <span class="friend-name">%s</span>
        <span class="friend-rank">%s • Niv. %d</span>
    </div>
    <div class="friend-actions">
        <a href="/chat/%s" class="btn-sm">💬</a>
    </div>
</div>`, string(username[0]), username, rank, int(level), friendID)
			}
		}
	}

	if friendCards == "" {
		friendCards = `<p style="color: #94a3b8; text-align: center; padding: 40px;">Aucun ami. Recherchez des utilisateurs !</p>`
	}

	return renderPage(c, "Amis", fmt.Sprintf(`
<h1 style="margin-bottom: 24px;">👥 Amis</h1>
<div class="friends-list">%s</div>
	`, friendCards))
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
