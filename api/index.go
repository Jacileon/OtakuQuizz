package handler

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2/middleware/adaptor"

	"otaku-quiz-africa/pkg/app"
)

//go:embed views
var viewsEmbed embed.FS

//go:embed static
var staticEmbed embed.FS

var fiberHandler http.Handler
var staticHandler http.Handler

func init() {
	if os.Getenv("SUPABASE_URL") == "" {
		os.Setenv("SUPABASE_URL", os.Getenv("NEXT_PUBLIC_SUPABASE_URL"))
	}
	if os.Getenv("SUPABASE_ANON_KEY") == "" {
		os.Setenv("SUPABASE_ANON_KEY", os.Getenv("NEXT_PUBLIC_SUPABASE_ANON_KEY"))
	}

	viewsSub, err := fs.Sub(viewsEmbed, "views")
	if err != nil {
		log.Fatal("Failed to sub views embed:", err)
	}
	application := app.Setup(viewsSub, "")
	fiberHandler = adaptor.FiberApp(application)

	staticSub, err := fs.Sub(staticEmbed, "static")
	if err != nil {
		log.Fatal("Failed to sub static embed:", err)
	}
	staticHandler = http.StripPrefix("/static/", http.FileServer(http.FS(staticSub)))

	log.Println("Vercel Go handler initialized (embedded views + static)")
}

func Handler(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/static/") {
		staticHandler.ServeHTTP(w, r)
		return
	}
	fiberHandler.ServeHTTP(w, r)
}
