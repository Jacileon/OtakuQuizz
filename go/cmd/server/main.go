package main

import (
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"

	"otaku-quiz-africa/internal/database"
	"otaku-quiz-africa/internal/handlers"
	"otaku-quiz-africa/internal/middleware"
)

func main() {
	// Charger les variables d'environnement
	godotenv.Load("../../.env.local")
	godotenv.Load("../.env.local")
	godotenv.Load(".env.local")

	// Mapper les variables Supabase
	if os.Getenv("SUPABASE_URL") == "" {
		os.Setenv("SUPABASE_URL", os.Getenv("NEXT_PUBLIC_SUPABASE_URL"))
	}
	if os.Getenv("SUPABASE_ANON_KEY") == "" {
		os.Setenv("SUPABASE_ANON_KEY", os.Getenv("NEXT_PUBLIC_SUPABASE_ANON_KEY"))
	}

	// Connexion à Supabase
	db, err := database.Connect()
	if err != nil {
		log.Fatal("Failed to connect to Supabase:", err)
	}
	log.Println("Connected to Supabase")

	// Template engine
	engine := html.New("./views", ".html")

	// Application Fiber
	app := fiber.New(fiber.Config{
		Views:         engine,
		BodyLimit:     4 * 1024 * 1024, // 4MB
		ReadBufferSize: 16384, // 16KB
	})

	// Middleware
	app.Use(logger.New())
	app.Use(recover.New())

	// Session store avec configuration légère
	store := session.New(session.Config{
		CookieHTTPOnly: true,
		CookieSameSite: "Lax",
		Expiration:     24 * time.Hour,
	})

	// Fichiers statiques
	app.Static("/static", "./static")

	// Handlers
	h := handlers.New(db, store)
	mw := middleware.New(db, store)

	// Routes publiques
	app.Get("/", h.Home)
	app.Get("/login", h.LoginPage)
	app.Post("/login", h.Login)
	app.Get("/register", h.RegisterPage)
	app.Post("/register", h.Register)
	app.Get("/logout", h.Logout)

	// Routes protégées
	protected := app.Group("", mw.RequireAuth)
	protected.Get("/dashboard", h.Dashboard)
	protected.Get("/explore", h.Explore)
	protected.Get("/quiz/:id", h.QuizDetail)
	protected.Get("/quiz/:id/play", h.QuizPlay)
	protected.Post("/api/quiz/submit", h.QuizSubmit)
	protected.Get("/friends", h.Friends)
	protected.Get("/challenges/:id", h.ChallengeDetail)
	protected.Get("/profil", h.Profile)
	protected.Get("/profile/edit", h.ProfileEdit)
	protected.Post("/profile/edit", h.ProfileUpdate)
	protected.Get("/leaderboard", h.Leaderboard)
	protected.Get("/faq", h.FAQ)

	// Routes admin
	admin := protected.Group("/admin", mw.RequireAdmin)
	admin.Get("/", h.AdminDashboard)
	admin.Get("/official-quizzes", h.AdminOfficialQuizzes)
	admin.Get("/tickets", h.AdminTickets)
	admin.Get("/announcements", h.AdminAnnouncements)
	admin.Get("/settings", h.AdminSettings)

	// Port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on http://localhost:%s", port)
	log.Fatal(app.Listen(":" + port))
}
