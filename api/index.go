package handler

import (
	"log"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2/middleware/adaptor"

	"otaku-quiz-africa/internal/app"
)

var handler http.Handler

func init() {
	if os.Getenv("SUPABASE_URL") == "" {
		os.Setenv("SUPABASE_URL", os.Getenv("NEXT_PUBLIC_SUPABASE_URL"))
	}
	if os.Getenv("SUPABASE_ANON_KEY") == "" {
		os.Setenv("SUPABASE_ANON_KEY", os.Getenv("NEXT_PUBLIC_SUPABASE_ANON_KEY"))
	}

	viewsFS := os.DirFS("./views")
	application := app.Setup(viewsFS, "")
	handler = adaptor.FiberApp(application)
	log.Println("Vercel Go handler initialized")
}

func Handler(w http.ResponseWriter, r *http.Request) {
	handler.ServeHTTP(w, r)
}
