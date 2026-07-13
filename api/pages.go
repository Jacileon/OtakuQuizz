package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

func serveHome(w http.ResponseWriter, r *http.Request, user map[string]interface{}) {
	content := `<div class="hero"><h1>⚔️ OTAKU QUIZ AFRICA</h1><p>Teste tes connaissances anime et manga</p><div class="buttons"><a href="/explore" class="btn-primary">Explorer les quiz</a><a href="/register" class="btn-secondary">S'inscrire</a></div></div>`
	renderPage(w, "Accueil", content, user)
}

func serveLogin(w http.ResponseWriter, r *http.Request) {
	renderPage(w, "Connexion", `<div class="auth-page"><h1>Connexion</h1><div id="supabase-auth"></div></div>`, nil)
}

func serveRegister(w http.ResponseWriter, r *http.Request) {
	renderPage(w, "Inscription", `<div class="auth-page"><h1>Inscription</h1><div id="supabase-auth"></div></div>`, nil)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "oqa_session", Value: "", Path: "/", MaxAge: -1, HttpOnly: true})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func serveDashboard(w http.ResponseWriter, r *http.Request, user map[string]interface{}) {
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	userID := getMapStr(user, "id")
	statsBody, _ := db.Select("user_stats", fmt.Sprintf("user_id=eq.%s&select=*", userID), true)
	var stats []map[string]interface{}
	json.Unmarshal(statsBody, &stats)
	quizzesBody, _ := db.Select("quizzes", "is_visible=eq.true&order=play_count.desc&limit=6&select=*", false)
	var quizzes []map[string]interface{}
	json.Unmarshal(quizzesBody, &quizzes)
	content := fmt.Sprintf(`<div class="dashboard"><h1>Bienvenue, %s !</h1><div class="stats-grid"><div class="stat-card"><span class="stat-value">%v XP</span><span class="stat-label">Total XP</span></div><div class="stat-card"><span class="stat-value">#%v</span><span class="stat-label">Rang</span></div></div><h2>Quiz populaires</h2><div class="quiz-grid">%s</div></div>`, getDisplayName(user), getMapInt(user, "total_xp"), user["rank"], renderQuizCards(quizzes))
	renderPage(w, "Dashboard", content, user)
}

func serveExplore(w http.ResponseWriter, r *http.Request, user map[string]interface{}) {
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	quizzesBody, _ := db.Select("quizzes", "is_visible=eq.true&order=play_count.desc&limit=20&select=*", false)
	var quizzes []map[string]interface{}
	json.Unmarshal(quizzesBody, &quizzes)
	categories := []string{"Shonen", "Seinen", "Isekai", "Action", "Fantasy", "Comédie", "Drame", "Horreur"}
	catTags := ""
	for _, cat := range categories {
		catTags += fmt.Sprintf(`<a href="/explore?category=%s" class="tag">%s</a>`, cat, cat)
	}
	content := fmt.Sprintf(`<div class="explore"><h1>🔍 Explorer</h1><div class="tags">%s</div><div class="quiz-grid">%s</div></div>`, catTags, renderQuizCards(quizzes))
	renderPage(w, "Explorer", content, user)
}

func serveLeaderboard(w http.ResponseWriter, r *http.Request, user map[string]interface{}) {
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	body, _ := db.RPC("get_global_leaderboard", map[string]interface{}{"limit_count": 20})
	var entries []map[string]interface{}
	json.Unmarshal(body, &entries)
	rows := ""
	for i, e := range entries {
		medal := fmt.Sprintf("#%d", i+1)
		if i == 0 { medal = "🥇" } else if i == 1 { medal = "🥈" } else if i == 2 { medal = "🥉" }
		rows += fmt.Sprintf(`<div class="lb-row"><span class="lb-rank">%s</span><span class="lb-name">%s</span><span class="lb-xp">%v XP</span></div>`, medal, getMapStr(e, "username"), getMapInt(e, "total_xp"))
	}
	content := fmt.Sprintf(`<div class="leaderboard"><h1>🏆 Classement</h1>%s</div>`, rows)
	renderPage(w, "Classement", content, user)
}

func serveFriends(w http.ResponseWriter, r *http.Request, user map[string]interface{}) {
	if user == nil { http.Redirect(w, r, "/login", http.StatusSeeOther); return }
	renderPage(w, "Amis", `<div class="friends"><h1>👥 Amis</h1><p>Chargement...</p></div>`, user)
}

func serveChallenges(w http.ResponseWriter, r *http.Request, user map[string]interface{}) {
	if user == nil { http.Redirect(w, r, "/login", http.StatusSeeOther); return }
	renderPage(w, "Défis", `<div class="challenges"><h1>⚔️ Défis</h1><p>Chargement...</p></div>`, user)
}

func serveProfile(w http.ResponseWriter, r *http.Request, user map[string]interface{}) {
	if user == nil { http.Redirect(w, r, "/login", http.StatusSeeOther); return }
	nickname := getDisplayName(user)
	bio := getMapStr(user, "bio")
	xp := getMapInt(user, "total_xp")
	rank := getMapStr(user, "rank")
	content := fmt.Sprintf(`<div class="profile"><h1>👤 %s</h1><p class="bio">%s</p><div class="stats"><span>%d XP</span><span class="rank rank-%s">%s</span></div></div>`, nickname, bio, xp, strings.ToLower(rank), rank)
	renderPage(w, "Profil", content, user)
}

func serveMiniCup(w http.ResponseWriter, r *http.Request, user map[string]interface{}) {
	if user == nil { http.Redirect(w, r, "/login", http.StatusSeeOther); return }
	renderPage(w, "Mini Cup", `<div><h1>⚽ Mini Cup</h1><p>Le jeu de tir aux buts !</p></div>`, user)
}

func servePersonalQuizDashboard(w http.ResponseWriter, r *http.Request, user map[string]interface{}) {
	if user == nil { http.Redirect(w, r, "/login", http.StatusSeeOther); return }
	userID := getMapStr(user, "id")
	body, _ := db.Select("personal_quizzes", fmt.Sprintf("creator_id=eq.%s&select=*", userID), true)
	var quizzes []map[string]interface{}
	json.Unmarshal(body, &quizzes)
	rows := ""
	for _, q := range quizzes {
		rows += fmt.Sprintf(`<div class="quiz-card"><h3>%s</h3><p>%s</p><a href="/personal-quiz/%s/edit" class="btn-primary">Modifier</a></div>`, getMapStr(q, "title"), getMapStr(q, "status"), getMapStr(q, "id"))
	}
	if rows == "" { rows = `<p>Aucun quiz. <a href="/personal-quiz/create">Créer un quiz</a></p>` }
	renderPage(w, "Mon Quiz Personnel", fmt.Sprintf(`<div><h1>🎯 Mon Quiz Personnel</h1>%s</div>`, rows), user)
}

func servePersonalQuizCreate(w http.ResponseWriter, r *http.Request, user map[string]interface{}) {
	if user == nil { http.Redirect(w, r, "/login", http.StatusSeeOther); return }
	renderPage(w, "Créer mon quiz", fmt.Sprintf(`<div><h1>Crée ton quiz personnel</h1><p>Salut %s !</p></div>`, template.HTMLEscapeString(getDisplayName(user))), user)
}

func servePersonalQuizCreatePost(w http.ResponseWriter, r *http.Request, user map[string]interface{}) {
	http.Redirect(w, r, "/personal-quiz", http.StatusSeeOther)
}

func servePersonalQuizEdit(w http.ResponseWriter, r *http.Request, user map[string]interface{}) {
	if user == nil { http.Redirect(w, r, "/login", http.StatusSeeOther); return }
	renderPage(w, "Modifier Quiz", `<div><h1>Modifier Quiz</h1></div>`, user)
}

func serveAnonBoxDashboard(w http.ResponseWriter, r *http.Request, user map[string]interface{}) {
	if user == nil { http.Redirect(w, r, "/login", http.StatusSeeOther); return }
	renderPage(w, "Messages Anonymes", `<div><h1>💌 Messages Anonymes</h1></div>`, user)
}

func serveAnonBoxCreate(w http.ResponseWriter, r *http.Request, user map[string]interface{}) {
	if user == nil { http.Redirect(w, r, "/login", http.StatusSeeOther); return }
	renderPage(w, "Créer Anon Box", `<div><h1>Créer ma boîte</h1></div>`, user)
}

func serveAnonBoxCreatePost(w http.ResponseWriter, r *http.Request, user map[string]interface{}) {
	http.Redirect(w, r, "/anon-box", http.StatusSeeOther)
}

func serveAnonBoxPublic(w http.ResponseWriter, r *http.Request) {
	renderPage(w, "Message Anonyme", `<div><h1>💌 Envoie un message anonyme</h1></div>`, nil)
}

// API handlers
func serveQuizSubmitAPI(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, 200, map[string]interface{}{"status": "ok"})
}

func serveRPCAPI(w http.ResponseWriter, r *http.Request) {
	rpcName := strings.TrimPrefix(r.URL.Path, "/api/rpc/")
	params := map[string]interface{}{}
	json.NewDecoder(r.Body).Decode(&params)
	body, err := db.RPC(rpcName, params)
	if err != nil { jsonResponse(w, 500, map[string]string{"error": err.Error()}); return }
	var result interface{}
	json.Unmarshal(body, &result)
	jsonResponse(w, 200, result)
}

func serveMiniCupAPI(w http.ResponseWriter, r *http.Request, user map[string]interface{}) {
	jsonResponse(w, 200, map[string]string{"status": "ok"})
}

func servePersonalQuizAPI(w http.ResponseWriter, r *http.Request, user map[string]interface{}) {
	jsonResponse(w, 200, map[string]string{"status": "ok"})
}

func serveAnonBoxAPI(w http.ResponseWriter, r *http.Request, user map[string]interface{}) {
	jsonResponse(w, 200, map[string]string{"status": "ok"})
}

func serveFriendsAPI(w http.ResponseWriter, r *http.Request, user map[string]interface{}) {
	jsonResponse(w, 200, []interface{}{})
}

func serveChallengesAPI(w http.ResponseWriter, r *http.Request, user map[string]interface{}) {
	jsonResponse(w, 200, []interface{}{})
}

func serveForumAPI(w http.ResponseWriter, r *http.Request, user map[string]interface{}) {
	jsonResponse(w, 200, []interface{}{})
}

func serveQuizAPI(w http.ResponseWriter, r *http.Request, user map[string]interface{}) {
	jsonResponse(w, 200, map[string]string{"status": "ok"})
}

func serveNotificationsAPI(w http.ResponseWriter, r *http.Request, user map[string]interface{}) {
	jsonResponse(w, 200, map[string]interface{}{"notifications": []interface{}{}, "unread_count": 0})
}

func serveAdminAPI(w http.ResponseWriter, r *http.Request, user map[string]interface{}) {
	jsonResponse(w, 200, map[string]string{"status": "ok"})
}

// Helpers
func renderQuizCards(quizzes []map[string]interface{}) string {
	html := ""
	for _, q := range quizzes {
		html += fmt.Sprintf(`<div class="quiz-card"><div class="quiz-header">%s</div><h3>%s</h3><p>%d questions • %d plays</p><a href="/quiz/%s/play" class="btn-primary">Jouer</a></div>`, getMapStr(q, "series"), getMapStr(q, "title"), getMapInt(q, "question_count"), getMapInt(q, "play_count"), getMapStr(q, "id"))
	}
	return html
}

func getDisplayName(user map[string]interface{}) string {
	if user == nil { return "Joueur" }
	if nn, ok := user["nickname"].(string); ok && nn != "" { return nn }
	if un, ok := user["username"].(string); ok && un != "" { return un }
	return "Joueur"
}

func getMapStr(m map[string]interface{}, key string) string {
	if m == nil { return "" }
	if v, ok := m[key].(string); ok { return v }
	return ""
}

func getMapInt(m map[string]interface{}, key string) int {
	if m == nil { return 0 }
	if v, ok := m[key].(float64); ok { return int(v) }
	return 0
}
