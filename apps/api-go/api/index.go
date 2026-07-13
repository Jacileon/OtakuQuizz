package handler

import (
	"encoding/json"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"strings"

	"otaku-quiz-africa/views"
)

var (
	db   *Supabase
	tmpl *template.Template
)

func init() {
	if os.Getenv("SUPABASE_URL") == "" {
		os.Setenv("SUPABASE_URL", os.Getenv("NEXT_PUBLIC_SUPABASE_URL"))
	}
	if os.Getenv("SUPABASE_ANON_KEY") == "" {
		os.Setenv("SUPABASE_ANON_KEY", os.Getenv("NEXT_PUBLIC_SUPABASE_ANON_KEY"))
	}

	db = Connect()

	subFS, err := fs.Sub(views.TemplatesFS, ".")
	if err == nil {
		tmpl = template.Must(template.ParseFS(subFS, "*.html", "layouts/*.html", "pages/*.html", "partials/*.html"))
	}
}

func Handler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	method := r.Method

	if strings.HasPrefix(path, "/static/") || strings.HasPrefix(path, "/_next/") {
		http.NotFound(w, r)
		return
	}

	user := getUserFromCookie(r)

	switch {
	case path == "/" && method == "GET":
		serveHome(w, r, user)
	case path == "/login" && method == "GET":
		serveLogin(w, r)
	case path == "/register" && method == "GET":
		serveRegister(w, r)
	case path == "/logout":
		logoutHandler(w, r)
	case path == "/dashboard" && method == "GET":
		serveDashboard(w, r, user)
	case path == "/explore" && method == "GET":
		serveExplore(w, r, user)
	case path == "/leaderboard" && method == "GET":
		serveLeaderboard(w, r, user)
	case path == "/friends" && method == "GET":
		serveFriends(w, r, user)
	case path == "/defis" || path == "/d\u00e9fis":
		serveChallenges(w, r, user)
	case path == "/profil" && method == "GET":
		serveProfile(w, r, user)
	case path == "/games/mini-cup" && method == "GET":
		serveMiniCup(w, r, user)
	case path == "/personal-quiz" && method == "GET":
		servePersonalQuizDashboard(w, r, user)
	case path == "/personal-quiz/create":
		if method == "POST" {
			servePersonalQuizCreatePost(w, r, user)
		} else {
			servePersonalQuizCreate(w, r, user)
		}
	case strings.HasPrefix(path, "/personal-quiz/") && strings.HasSuffix(path, "/edit"):
		servePersonalQuizEdit(w, r, user)
	case path == "/anon-box" && method == "GET":
		serveAnonBoxDashboard(w, r, user)
	case path == "/anon-box/create":
		if method == "POST" {
			serveAnonBoxCreatePost(w, r, user)
		} else {
			serveAnonBoxCreate(w, r, user)
		}
	case strings.HasPrefix(path, "/anon/"):
		serveAnonBoxPublic(w, r)
	case path == "/api/quiz/submit" && method == "POST":
		serveQuizSubmitAPI(w, r)
	case strings.HasPrefix(path, "/api/rpc/"):
		serveRPCAPI(w, r)
	case strings.HasPrefix(path, "/api/mini-cup/"):
		serveMiniCupAPI(w, r, user)
	case strings.HasPrefix(path, "/api/personal-quiz/"):
		servePersonalQuizAPI(w, r, user)
	case strings.HasPrefix(path, "/api/anon-box/"):
		serveAnonBoxAPI(w, r, user)
	case strings.HasPrefix(path, "/api/friends"):
		serveFriendsAPI(w, r, user)
	case strings.HasPrefix(path, "/api/challenges"):
		serveChallengesAPI(w, r, user)
	case strings.HasPrefix(path, "/api/forum"):
		serveForumAPI(w, r, user)
	case strings.HasPrefix(path, "/api/quiz"):
		serveQuizAPI(w, r, user)
	case strings.HasPrefix(path, "/api/notifications"):
		serveNotificationsAPI(w, r, user)
	case strings.HasPrefix(path, "/api/admin"):
		serveAdminAPI(w, r, user)
	case path == "/api/auth/session" && method == "POST":
		serveAuthSessionAPI(w, r)
	default:
		http.NotFound(w, r)
	}
}

func getUserFromCookie(r *http.Request) map[string]interface{} {
	cookie, err := r.Cookie("oqa_session")
	if err != nil || cookie.Value == "" {
		return nil
	}

	var sessionData struct {
		UserID    string `json:"user_id"`
		AuthToken string `json:"auth_token"`
	}
	if err := json.Unmarshal([]byte(cookie.Value), &sessionData); err != nil {
		return nil
	}

	if sessionData.UserID == "" || sessionData.AuthToken == "" {
		return nil
	}

	profile, err := db.GetProfile(sessionData.UserID, sessionData.AuthToken)
	if err != nil {
		return nil
	}
	return profile
}

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func renderPage(w http.ResponseWriter, title string, content string, user map[string]interface{}) {
	data := map[string]interface{}{
		"Title":   title,
		"Content": template.HTML(content),
	}
	if user != nil {
		data["User"] = user
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if tmpl != nil {
		tmpl.ExecuteTemplate(w, "layouts/main.html", data)
	} else {
		w.Write([]byte("<h1>Template not loaded</h1>"))
	}
}
