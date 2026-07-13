package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"otaku-quiz-africa/internal/app"
)

func main() {
	godotenv.Load(".env.local")
	godotenv.Load("../.env.local")
	godotenv.Load("../../.env.local")

	viewsFS := os.DirFS("./views")

	application := app.Setup(viewsFS, "./static")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on http://localhost:%s", port)
	log.Fatal(application.Listen(":" + port))
}
