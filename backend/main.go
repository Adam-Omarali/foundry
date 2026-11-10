// main.go
package main

import (
	"fmt"
	"foundry/backend/api"
	"foundry/backend/middleware"
	"foundry/backend/utils"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// In production (the container) compose injects the environment. Locally,
	// `go run .` from backend/ reads the single repo-root .env.
	if os.Getenv("APP_ENV") != "production" {
		if err := godotenv.Load("../.env"); err != nil {
			log.Println("Warning: could not load ../.env:", err)
		} else {
			log.Println("Loaded ../.env for local development")
		}
	}

	setupLogger()

	// Initialize JWT
	if err := utils.InitializeJWT(); err != nil {
		log.Fatal("Error initializing JWT:", err)
	}
	if err := middleware.InitGoogleAuth(); err != nil {
		log.Fatal("Error initializing Google auth:", err)
	}
	middleware.InitCORS()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Go 1.22+ patterns: "METHOD /path/{var}". The mux returns 405 for a wrong
	// method, and r.PathValue("id") reads {id}.
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/auth/google", api.GoogleAuthHandler)
	mux.HandleFunc("POST /api/documents", api.NewDocumentHandler)
	mux.HandleFunc("GET /api/verifytoken", api.VerifyJWTHandler)
	mux.HandleFunc("GET /api/documents", api.GetDocumentsHandler)
	mux.HandleFunc("PATCH /api/documents/{id}", api.UpdateDocumentHandler)

	// Middleware runs outside-in: LogRequests → EnableCORS → mux → handler.
	// Logging is outermost so it also sees preflights and CORS rejections.
	// CORS wraps the mux so OPTIONS preflights are answered before routing.
	var handler http.Handler = mux
	handler = middleware.EnableCORS(handler)
	handler = middleware.LogRequests(handler)

	slog.Info("server listening", "port", port)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, handler))
}

// setupLogger: JSON lines in production (easy to grep/ship to a log service),
// human-readable text locally.
//
// Level: LOG_LEVEL=debug|info|warn|error; defaults to info in production and
// debug locally. SetDefault also routes the std `log` package through slog.
func setupLogger() {
	prod := os.Getenv("APP_ENV") == "production"

	level := slog.LevelDebug
	if prod {
		level = slog.LevelInfo
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		if err := level.UnmarshalText([]byte(v)); err != nil {
			fmt.Fprintf(os.Stderr, "invalid LOG_LEVEL %q, using %s\n", v, level)
		}
	}

	opts := &slog.HandlerOptions{Level: level}
	var h slog.Handler = slog.NewTextHandler(os.Stdout, opts)
	if prod {
		h = slog.NewJSONHandler(os.Stdout, opts)
	}
	slog.SetDefault(slog.New(middleware.ContextHandler{Handler: h}))
}
