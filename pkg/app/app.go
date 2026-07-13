package app

import (
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/template/html/v2"

	"otaku-quiz-africa/pkg/database"
	"otaku-quiz-africa/pkg/handlers"
	"otaku-quiz-africa/pkg/middleware"
	"otaku-quiz-africa/pkg/sessionstore"
)

func Setup(viewsFS fs.FS, staticPath string) *fiber.App {
	if os.Getenv("SUPABASE_URL") == "" {
		os.Setenv("SUPABASE_URL", os.Getenv("NEXT_PUBLIC_SUPABASE_URL"))
	}
	if os.Getenv("SUPABASE_ANON_KEY") == "" {
		os.Setenv("SUPABASE_ANON_KEY", os.Getenv("NEXT_PUBLIC_SUPABASE_ANON_KEY"))
	}

	db, err := database.Connect()
	if err != nil {
		log.Println("WARNING: Failed to connect to Supabase:", err)
	}

	engine := html.NewFileSystem(http.FS(viewsFS), ".html")
	engine.AddFunc("upperFirst", func(s string) string {
		if len(s) == 0 {
			return "?"
		}
		return strings.ToUpper(string(s[0]))
	})

	app := fiber.New(fiber.Config{
		Views:          engine,
		BodyLimit:      4 * 1024 * 1024,
		ReadBufferSize: 16384,
	})

	var store *session.Store
	isVercel := os.Getenv("VERCEL") != ""
	if isVercel {
		memStore := sessionstore.NewMemory()
		store = session.New(session.Config{
			Storage:         memStore,
			CookieHTTPOnly:  true,
			CookieSameSite:  "Lax",
			Expiration:      72 * time.Hour,
		})
		log.Println("Using in-memory session store (Vercel serverless)")
	} else {
		fileStore := sessionstore.New("./data/sessions")
		store = session.New(session.Config{
			Storage:         fileStore,
			CookieHTTPOnly:  true,
			CookieSameSite:  "Lax",
			Expiration:      72 * time.Hour,
		})
	}

	h := handlers.New(db, store)
	mw := middleware.New(db, store)

	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(mw.SetUser)

	if staticPath != "" {
		app.Static("/static", staticPath)
	}

	// Routes publiques
	app.Get("/", h.Home)
	app.Get("/login", h.LoginPage)
	app.Get("/register", h.RegisterPage)
	app.Get("/logout", h.Logout)
	app.Get("/auth/google", h.GoogleAuth)
	app.Get("/auth/callback", h.GoogleCallback)
	app.Post("/auth/session", h.CreateSession)

	// API routes (protected with auth)
	api := app.Group("/api", mw.RequireAuth)
	api.Get("/rpc/get_global_leaderboard", h.APIGetGlobalLeaderboard)
	api.Post("/rpc/get_global_leaderboard", h.APIGetGlobalLeaderboard)
	api.Get("/rpc/get_monthly_leaderboard", h.APIGetMonthlyLeaderboard)
	api.Get("/rpc/get_quiz_leaderboard", h.APIGetQuizLeaderboard)
	api.Get("/rpc/get_series_leaderboard", h.APIGetSeriesLeaderboard)
	api.Get("/friends", h.APIGetFriends)
	api.Get("/friends/requests", h.APIGetFriendRequests)
	api.Post("/friends/request", h.APISendFriendRequest)
	api.Post("/friends/accept", h.APIAcceptFriendRequest)
	api.Post("/friends/reject", h.APIRejectFriendRequest)
	api.Post("/friends/remove", h.APIRemoveFriend)
	api.Get("/users/search", h.APISearchUsers)
	api.Get("/friends/requests/sent", h.APIGetSentRequests)
	api.Get("/friends/status", h.APIGetFriendshipStatus)
	api.Get("/conversations", h.APIGetConversations)
	api.Post("/conversations/get-or-create", h.APIGetOrCreateConversation)
	api.Post("/conversations/delete", h.APIDeleteConversation)
	api.Get("/messages", h.APIGetMessages)
	api.Post("/messages/send", h.APISendMessage)
	api.Get("/challenges/invitations", h.APIGetChallengeInvitations)
	api.Post("/challenges/accept", h.APIAcceptChallengeInvitation)
	api.Post("/challenges/refuse", h.APIRefuseChallengeInvitation)
	api.Get("/challenges/mine", h.APIGetMyChallenges)
	api.Post("/challenges/delete", h.APIDeleteChallenge)
	api.Post("/challenges/invite", h.APIInviteToChallenge)
	api.Post("/challenges/resend", h.APIResendChallengeInvitation)
	api.Post("/challenges/raise-bet", h.APIRaiseBet)
	api.Post("/challenges/bet-respond", h.APIBetRespond)
	api.Post("/challenges/set-reward-mode", h.APISetRewardMode)
	api.Get("/challenges/:id", h.APIGetChallenge)
	api.Post("/challenges/:id/start", h.APIStartChallenge)
	api.Post("/reports/user", h.APIReportUser)
	api.Post("/quiz/vote", h.APIQuizVote)
	api.Post("/quiz/upload-image", h.APIQuizUploadImage)
	api.Post("/quiz/submit", h.QuizSubmit)
	api.Post("/quiz/:id/questions", h.APIQuizAddQuestion)
	api.Delete("/quiz/:id/questions/:qid", h.APIQuizDeleteQuestion)
	api.Post("/quiz/:id/questions/:qid/xp", h.APIQuizUpdateQuestionXp)
	api.Post("/quiz/:id/questions/:qid", h.APIQuizUpdateQuestion)
	api.Post("/quiz/:id/meta", h.APIQuizUpdateMeta)
	api.Post("/quiz/:id/publish", h.APIQuizPublish)
	api.Post("/quiz/:id/generate", h.APIQuizGenerateQuestions)
	api.Get("/notifications", h.APIGetNotifications)
	api.Post("/notifications/read", h.APIMarkNotificationsRead)
	api.Get("/xp/history", h.APIGetXPHistory)
	api.Get("/admin/conversations", h.APIGetAdminConversations)
	api.Post("/admin/conversations/create", h.APICreateAdminConversation)
	api.Get("/admin/messages", h.APIGetAdminMessages)
	api.Post("/admin/messages/send", h.APISendAdminMessage)
	api.Post("/admin/conversations/delete", h.APIDeleteAdminConversation)
	api.Post("/admin/conversations/update-status", h.APIUpdateAdminConversationStatus)
	api.Get("/admin/announcements", h.APIGetAllAnnouncements)
	api.Post("/admin/announcements/create", h.APICreateAnnouncement)
	api.Post("/admin/announcements/update", h.APIUpdateAnnouncement)
	api.Post("/admin/announcements/delete", h.APIDeleteAnnouncement)
	api.Get("/admin/official-quizzes", h.APIGetAllOfficialQuizzes)
	api.Post("/admin/official-quizzes/create", h.APICreateOfficialQuiz)
	api.Get("/forum/channels", h.APIGetForumChannels)
	api.Get("/forum/messages", h.APIGetForumMessages)
	api.Post("/forum/messages/send", h.APISendForumMessage)
	api.Post("/forum/messages/update", h.APIUpdateForumMessage)
	api.Post("/forum/messages/delete", h.APIDeleteForumMessage)
	api.Post("/forum/messages/report", h.APIReportForumMessage)
	api.Get("/forum/suggestions", h.APIGetForumSuggestions)
	api.Get("/forum/suggestions/all", h.APIGetAllSuggestions)
	api.Post("/forum/suggestions/create", h.APICreateForumSuggestion)
	api.Post("/forum/suggestions/vote", h.APISuggestionVote)
	api.Post("/forum/suggestions/delete", h.APIDeleteForumSuggestion)
	api.Post("/forum/suggestions/update", h.APISuggestionUpdate)

	api.Post("/mini-cup/session", h.CreateMiniCupSession)
	api.Get("/mini-cup/sessions", h.GetMiniCupSessions)
	api.Post("/mini-cup/session/:id/shot", h.RecordMiniCupShot)
	api.Post("/mini-cup/session/:id/finish", h.FinishMiniCupSession)
	api.Get("/mini-cup/leaderboard", h.GetMiniCupLeaderboard)

	api.Post("/personal-quiz/:id/questions", h.APIPersonalQuizAddQuestion)
	api.Patch("/personal-quiz/:id/questions/:qid", h.APIPersonalQuizUpdateQuestion)
	api.Delete("/personal-quiz/:id/questions/:qid", h.APIPersonalQuizDeleteQuestion)
	api.Post("/personal-quiz/:id/archive", h.APIPersonalQuizArchive)
	api.Patch("/personal-quiz/:id", h.APIPersonalQuizUpdate)
	api.Post("/personal-quiz/:token/submit", h.PersonalQuizSubmit)
	api.Post("/personal-quiz/:token/message", h.PersonalQuizSendMessage)

	api.Post("/anon-box/create", h.AnonBoxCreatePost)
	api.Post("/anon-box/reset", h.APIAnonBoxReset)
	api.Patch("/anon-box", h.APIAnonBoxUpdate)
	app.Post("/api/anon-box/:id/send", h.APIAnonBoxSend)
	api.Post("/anon-box/messages/:mid/read", h.APIAnonBoxMarkRead)
	api.Post("/anon-box/messages/:mid/delete", h.APIAnonBoxDeleteMsg)

	// Routes publiques
	app.Get("/anon/:token", h.AnonBoxPublic)
	app.Get("/quiz/personal/:token", h.PersonalQuizView)
	app.Get("/quiz/personal/:token/play", h.PersonalQuizPlay)
	app.Get("/quiz/personal/:token/results", h.PersonalQuizResults)
	app.Post("/quiz/personal/:token/message", h.PersonalQuizSendMessage)

	// Routes protégées
	protected := app.Group("", mw.RequireAuth)
	protected.Get("/dashboard", h.Dashboard)
	protected.Get("/explore", h.Explore)
	protected.Get("/friends", h.Friends)
	protected.Get("/defis", h.ChallengesPage)
	protected.Get("/défis", h.ChallengesPage)
	protected.Get("/profil", h.Profile)
	protected.Get("/profile/edit", h.ProfileEdit)
	protected.Post("/profile/edit", h.ProfileUpdate)
	protected.Get("/leaderboard", h.Leaderboard)
	protected.Get("/leaderboard/official", h.OfficialLeaderboard)
	protected.Get("/faq", h.FAQ)
	protected.Get("/faq/all", h.FAQFromDB)
	protected.Get("/events", h.Events)
	protected.Get("/announcements", h.Announcements)
	protected.Get("/badges", h.Badges)
	protected.Get("/collections", h.Collections)
	protected.Get("/notifications/page", h.NotificationsPage)
	protected.Get("/quiz/create", h.QuizCreate)
	protected.Post("/quiz/create", h.QuizCreatePost)
	protected.Get("/challenges/create/:id", h.ChallengeCreate)
	protected.Post("/challenges/create/:id", h.ChallengeCreatePost)
	protected.Get("/forum", h.ForumPage)
	protected.Get("/boite-a-idees", h.IdeasBoxPage)
	protected.Get("/quiz/mine", h.MyQuizzes)
	protected.Get("/quiz/:id", h.QuizDetail)
	protected.Get("/quiz/:id/play", h.QuizPlay)
	protected.Get("/quiz/:id/results/:session", h.QuizResults)
	protected.Get("/quiz/:id/edit", h.QuizEdit)
	protected.Post("/quiz/:id/edit", h.QuizEditPost)
	protected.Post("/quiz/:id/archive", h.APIQuizArchive)
	protected.Post("/quiz/:id/delete", h.APIQuizDelete)
	protected.Get("/challenges/:id", h.ChallengeDetail)
	protected.Get("/challenges/:id/leaderboard", h.ChallengeLeaderboard)
	protected.Get("/profile/:username", h.ProfileView)
	protected.Get("/leaderboard/quiz/:id", h.QuizLeaderboard)
	protected.Get("/leaderboard/official/:id", h.OfficialLeaderboardDetail)
	protected.Get("/explore/:category", h.ExploreCategory)
	protected.Get("/series/:series", h.Series)
	protected.Get("/complete-profile", h.CompleteProfile)
	protected.Post("/complete-profile", h.CompleteProfilePost)
	protected.Get("/notifications", h.NotificationsPage)
	protected.Get("/xp-history", h.XPHistoryPage)
	protected.Get("/support", h.SupportPage)
	protected.Get("/support/:id", h.SupportConversation)

	protected.Get("/personal-quiz/create", h.PersonalQuizCreate)
	protected.Post("/personal-quiz/create", h.PersonalQuizCreatePost)
	protected.Get("/personal-quiz", h.PersonalQuizDashboard)
	protected.Get("/personal-quiz/:id/edit", h.PersonalQuizEdit)

	protected.Get("/games/mini-cup", h.MiniCupPage)

	protected.Get("/anon-box/create", h.AnonBoxCreate)
	protected.Post("/anon-box/create", h.AnonBoxCreatePost)
	protected.Get("/anon-box", h.AnonBoxDashboard)

	// Routes admin
	admin := protected.Group("/admin", mw.RequireAdmin)
	admin.Get("/", h.AdminDashboard)
	admin.Get("/official-quizzes", h.AdminOfficialQuizzes)
	admin.Get("/tickets", h.AdminTickets)
	admin.Get("/tickets/:id", h.AdminTickets)
	admin.Get("/announcements", h.AdminAnnouncements)
	admin.Get("/settings", h.AdminSettings)

	return app
}
